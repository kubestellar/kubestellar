/*
Copyright 2023 The KubeStellar Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package transport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/edge/v1alpha1"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	edgev1alpha1informers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kubestellar/kubestellar/pkg/generated/listers/edge/v1alpha1"
)

const (
	ControllerName                 = "transport-controller"
	transportFinalizer             = "transport.kubestellar.io/object-cleanup"
	notFoundErrorSuffix            = "not found"
	alreadyExistsErrorSuffix       = "already exists"
	originOwnerReferenceAnnotation = "transport.kubestellar.io/originOwnerReferencePlacementDecisionKey"
	originWdsAnnotation            = "transport.kubestellar.io/originWdsName"
)

// NewTransportController returns a new transport controller
func NewTransportController(ctx context.Context, placementDecisionInformer edgev1alpha1informers.PlacementDecisionInformer, transport Transport,
	wdsClientset *edgeclientset.Clientset, wdsDynamicClient *dynamic.DynamicClient, transportRestConfig *rest.Config, wdsName string) (*genericTransportController, error) {
	transportDynamicClient, err := dynamic.NewForConfig(transportRestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic k8s clientset for transport space - %w", err)
	}

	emptyWrappedObject := transport.WrapObjects(make([]*unstructured.Unstructured, 0)) // empty wrapped object to get GVR from it.
	wrappedObjectGVR, err := getGvrFromWrappedObject(transportRestConfig, emptyWrappedObject)
	if err != nil {
		return nil, fmt.Errorf("failed to get transport wrapped object GVR - %w", err)
	}

	dynamicInformer := dynamicinformer.NewDynamicSharedInformerFactory(transportDynamicClient, 0)
	wrappedObjectGenericInformer := dynamicInformer.ForResource(*wrappedObjectGVR)

	transportController := &genericTransportController{
		logger:                          klog.FromContext(ctx).WithName(ControllerName),
		placementDecisionLister:         placementDecisionInformer.Lister(),
		placementDecisionInformerSynced: placementDecisionInformer.Informer().HasSynced,
		wrappedObjectInformerSynced:     wrappedObjectGenericInformer.Informer().HasSynced,
		workqueue:                       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ControllerName),
		transport:                       transport,
		transportClient:                 transportDynamicClient,
		wrappedObjectGVR:                *wrappedObjectGVR,
		wdsClientset:                    wdsClientset,
		wdsDynamicClient:                wdsDynamicClient,
		wdsName:                         wdsName,
	}

	transportController.logger.Info("Setting up event handlers")
	// Set up an event handler for when PlacementDecision resources change
	placementDecisionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    transportController.enqueuePlacementDecision,
		UpdateFunc: func(_, new interface{}) { transportController.enqueuePlacementDecision(new) },
		DeleteFunc: transportController.enqueuePlacementDecision,
	})
	// Set up event handlers for when WrappedObject resources change. The handlers will lookup the origin PlacementDecision
	// of the given WrappedObject and enqueue that PlacementDecision object for processing.
	// This way, we don't need to implement custom logic for handling WrappedObject resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	wrappedObjectGenericInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    transportController.handleWrappedObject,
		UpdateFunc: func(_, new interface{}) { transportController.handleWrappedObject(new) },
		DeleteFunc: transportController.handleWrappedObject,
	})
	dynamicInformer.Start(ctx.Done())

	return transportController, nil
}

func convertObjectToUnstructured(object runtime.Object) (*unstructured.Unstructured, error) {
	unstructuredObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil, fmt.Errorf("failed to convert given object to unstructured - %w", err)
	}
	return &unstructured.Unstructured{Object: unstructuredObject}, nil
}

func getGvrFromWrappedObject(restConfig *rest.Config, wrappedObject runtime.Object) (*schema.GroupVersionResource, error) {
	unstructuredWrappedObject, err := convertObjectToUnstructured(wrappedObject)
	if err != nil {
		return nil, fmt.Errorf("failed to convert wrapped object to unstructured - %w", err)
	}

	gvk := unstructuredWrappedObject.GroupVersionKind()
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset for given config - %w", err)
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cacheddiscovery.NewMemCacheClient(clientset.Discovery()))

	restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to discover GroupVersionResource from given GroupVersionKind - %w", err)
	}

	return &(restMapping.Resource), nil
}

type genericTransportController struct {
	logger logr.Logger

	placementDecisionLister         edgev1alpha1listers.PlacementDecisionLister
	placementDecisionInformerSynced cache.InformerSynced
	wrappedObjectInformerSynced     cache.InformerSynced

	// workqueue is a rate limited work queue.
	// This is used to queue work to be processed instead of performing it as soon as a change happens.
	// This means we can ensure we only process a fixed amount of resources at a time, and makes it
	// easy to ensure we are never processing the same item simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	//transport is a specific implementation for the transport interface.
	transport        Transport
	transportClient  dynamic.Interface // dynamic client to transport wrapped object. since object kind is unknown during complilation, we use dynamic
	wrappedObjectGVR schema.GroupVersionResource

	wdsClientset     *edgeclientset.Clientset
	wdsDynamicClient *dynamic.DynamicClient
	wdsName          string
}

// enqueuePlacementDecision takes an PlacementDecision resource and
// converts it into a namespace/name string which is put onto the workqueue.
// This func *shouldn't* handle any resource other than PlacementDecision.
func (c *genericTransportController) enqueuePlacementDecision(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		c.logger.Error(err, "failed to enqueue PlacementDecision object")
		return
	}

	c.workqueue.Add(key)
}

// handleWrappedObject takes transport-specific wrapped object resource,
// extracts the origin PlacementDecision of the given wrapped object and
// enqueue that PlacementDecision object for processing. This way, we
// don't need to implement custom logic for handling WrappedObject resources.
// More info on this pattern here:
// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
func (c *genericTransportController) handleWrappedObject(obj interface{}) {
	wrappedObject := obj.(metav1.Object)
	ownerPlacementDecisionKey, found := wrappedObject.GetAnnotations()[originOwnerReferenceAnnotation] // safe if GetAnnotations() returns nil
	if !found {
		c.logger.Info("failed to extract placementdecision key from transport wrapped object", "WrappedObjectName", wrappedObject.GetName())
		return
	}

	// enqueue PlacementDecision key to trigger reconciliation.
	// if wrapped object was created not as a result of PlacementDecision,
	// the required annotation won't be found and nothing will happen.
	c.workqueue.Add(ownerPlacementDecisionKey)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until context
// is cancelled, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *genericTransportController) Run(ctx context.Context, workersCount int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.logger.Info("starting transport controller")

	// Wait for the caches to be synced before starting workers
	c.logger.Info("waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.placementDecisionInformerSynced, c.wrappedObjectInformerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	c.logger.Info("starting workers", "count", workersCount)
	// Launch workers to process PlacementDecision
	for i := 1; i <= workersCount; i++ {
		go wait.UntilWithContext(ctx, func(ctx context.Context) { c.runWorker(ctx, i) }, time.Second)
	}

	c.logger.Info("started workers")
	<-ctx.Done()
	c.logger.Info("shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *genericTransportController) runWorker(ctx context.Context, workerId int) {
	for c.processNextWorkItem(ctx, workerId) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *genericTransportController) processNextWorkItem(ctx context.Context, workerID int) bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	if err := c.process(ctx, obj); err != nil {
		c.logger.Error(err, "failed to process PlacementDecision object", "objectKey", obj, "workerID", workerID)
	} else {
		c.logger.Info("synced PlacementDecision object successfully.", "objectKey", obj, "workerID", workerID)
	}

	return true
}

func (c *genericTransportController) process(ctx context.Context, obj interface{}) error {
	// We call Done here so the workqueue knows we have finished processing this item.
	// We also must remember to call Forget if we do not want this work item being re-queued.
	// For example, we do not call Forget if a transient error occurs, instead the item is
	// put back on the workqueue and attempted again after a back-off period.
	defer c.workqueue.Done(obj)
	var key string
	var ok bool
	// We expect strings to come off the workqueue. These are of the form namespace/name.
	// We do this as the delayed nature of the workqueue means the items in the informer cache
	// may actually be more up to date that when the item was initially put onto the workqueue.
	if key, ok = obj.(string); !ok {
		// As the item in the workqueue is actually invalid, we call Forget here else we'd go
		// into a loop of attempting to process a work item that is invalid.
		c.workqueue.Forget(obj)
		return fmt.Errorf("expected key from type string in workqueue but got %#v", obj)
	}

	_, objectName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		// As the item in the workqueue is actually invalid, we call Forget here else we'd go
		// into a loop of attempting to process a work item that is invalid.
		c.workqueue.Forget(obj)
		return fmt.Errorf("invalid object key '%s' - %w", key, err)
	}

	// Run the syncHandler, passing it the PlacementDecision object name to be synced.
	if err := c.syncHandler(ctx, objectName); err != nil {
		// Put the item back on the workqueue to handle any transient errors.
		c.workqueue.AddRateLimited(key)
		return fmt.Errorf("failed to process object with key '%s' - %w", key, err)
	}
	// Finally, if no error occurs we Forget this item so it does not
	// get queued again until another change happens.
	c.workqueue.Forget(obj)

	return nil
}

// syncHandler compares the actual state with the desired, and attempts to converge actual state to the desired state.
// returning an error from this function will result in a requeue of the given object key.
// therefore, if object shouldn't be requeued, don't return error.
func (c *genericTransportController) syncHandler(ctx context.Context, objectName string) error {
	// Get the PlacementDecision object with this name from WDS
	placementDecision, err := c.placementDecisionLister.Get(objectName)
	if err != nil {
		if errors.IsNotFound(err) { // the object was deleted and it had no finalizer on it. this means transport controller
			// finished cleanup of wrapped objects from mailbox namespaces. no need to do anything in this state.
			return nil
		}
		// in case of a different error, log it and retry.
		return fmt.Errorf("failed to get PlacementDecision object '%s' - %w", objectName, err)
	}

	if isObjectBeingDeleted(placementDecision) {
		return c.deleteWrappedObjectsAndFinalizer(ctx, placementDecision)
	}
	// otherwise, object was not deleted and no error occurered while reading the object.
	return c.updateWrappedObjectsAndFinalizer(ctx, placementDecision)
}

// isObjectBeingDeleted is a helper function to check if object is being deleted.
func isObjectBeingDeleted(object metav1.Object) bool {
	return !object.GetDeletionTimestamp().IsZero()
}

func (c *genericTransportController) deleteWrappedObjectsAndFinalizer(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) error {
	for _, destination := range placementDecision.Spec.Destinations { // TODO need to revisit the destination struct and see how to use it properly
		if err := c.transportClient.Resource(c.wrappedObjectGVR).Namespace(destination.ClusterId).Delete(ctx, fmt.Sprintf("%s-%s", placementDecision.GetName(), c.wdsName),
			metav1.DeleteOptions{}); err != nil { // wrapped object name is in the format (PlacementDecision.GetName()-WdsName). see updateWrappedObject func for explanation.
			if !strings.HasSuffix(err.Error(), notFoundErrorSuffix) { // if object is already not there, we do not report an error cause desired state was achieved.
				return fmt.Errorf("failed to delete wrapped object '%s' in destination WEC with namespace '%s' - %w", fmt.Sprintf("%s-%s", placementDecision.GetName(),
					c.wdsName), destination.ClusterId, err)
			}
		}
	}

	if err := c.removeFinalizerFromPlacementDecision(ctx, placementDecision); err != nil {
		return fmt.Errorf("failed to remove finalizer from PlacementDecision object '%s' - %w", placementDecision.GetName(), err)
	}

	return nil
}

func (c *genericTransportController) removeFinalizerFromPlacementDecision(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) error {
	return c.updatePlacementDecision(ctx, placementDecision, func(placementDecision *v1alpha1.PlacementDecision) bool {
		return removeFinalizer(placementDecision, transportFinalizer)
	})
}

func (c *genericTransportController) addFinalizerToPlacementDecision(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) error {
	return c.updatePlacementDecision(ctx, placementDecision, func(placementDecision *v1alpha1.PlacementDecision) bool {
		return addFinalizer(placementDecision, transportFinalizer)
	})
}

func (c *genericTransportController) updateWrappedObjectsAndFinalizer(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) error {
	if err := c.addFinalizerToPlacementDecision(ctx, placementDecision); err != nil {
		return fmt.Errorf("failed to add finalizer to PlacementDecision object '%s' - %w", placementDecision.GetName(), err)
	}

	objectsToPropagate, err := c.getObjectsFromWDS(ctx, placementDecision)
	if err != nil {
		return fmt.Errorf("failed to get objects to propagate to WECs from PlacementDecision object '%s' - %w", placementDecision.GetName(), err)
	}

	wrappedObject, err := convertObjectToUnstructured(c.transport.WrapObjects(objectsToPropagate))
	if err != nil {
		return fmt.Errorf("failed to wrap objects to a single wrapped object - %w", err)
	}
	// wrapped object name is (PlacementDecision.GetName()-WdsName).
	// pay attention - we cannot use the PlacementDecision object name, cause we might have duplicate names coming from different WDS spaces.
	// we add WdsName to the object name to assure name uniqueness,
	// in order to easily get the origin PlacementDecision object name and wds, we add it as an annotations.
	wrappedObject.SetName(fmt.Sprintf("%s-%s", placementDecision.GetName(), c.wdsName))
	setAnnotation(wrappedObject, originOwnerReferenceAnnotation, placementDecision.GetName())
	setAnnotation(wrappedObject, originWdsAnnotation, c.wdsName)

	for _, destination := range placementDecision.Spec.Destinations { // TODO need to revisit the destination struct and see how to use it properly
		if err := c.createOrUpdateWrappedObject(ctx, destination.ClusterId, wrappedObject); err != nil {
			return fmt.Errorf("failed to propagate wrapped object '%s' to all required WECs - %w", wrappedObject.GetName(), err)
		}
	}

	return nil
}

func (c *genericTransportController) getObjectsFromWDS(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) ([]*unstructured.Unstructured, error) {
	objectsToPropagate := make([]*unstructured.Unstructured, 0)
	// add cluster-scoped objects to the 'objectsToPropagate' slice
	for _, clusterScopedObject := range placementDecision.Spec.Workload.ClusterScope {
		if clusterScopedObject.Objects == nil {
			continue // no objects from this gvr, skip
		}
		gvr := schema.GroupVersionResource{Group: clusterScopedObject.Group, Version: clusterScopedObject.Version, Resource: clusterScopedObject.Resource}
		for _, objectName := range clusterScopedObject.Objects {
			object, err := c.wdsDynamicClient.Resource(gvr).Get(ctx, objectName, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to get required cluster-scoped object '%s' with gvr %s from WDS - %w", objectName, gvr, err)
			}
			objectsToPropagate = append(objectsToPropagate, cleanObject(object))
		}
	}
	// add namespace-scoped objects to the 'objectsToPropagate' slice
	for _, namespaceScopedObject := range placementDecision.Spec.Workload.NamespaceScope {
		gvr := schema.GroupVersionResource{Group: namespaceScopedObject.Group, Version: namespaceScopedObject.APIVersion, Resource: namespaceScopedObject.Resource}
		for _, objectsByNamespace := range namespaceScopedObject.ObjectsByNamespace {
			if objectsByNamespace.Names == nil {
				continue // no objects from this namespace, skip
			}
			for _, objectName := range objectsByNamespace.Names {
				object, err := c.wdsDynamicClient.Resource(gvr).Namespace(objectsByNamespace.Namespace).Get(ctx, objectName, metav1.GetOptions{})
				if err != nil {
					return nil, fmt.Errorf("failed to get required namespace-scoped object '%s' with gvr '%s' from WDS - %w", objectName, gvr, err)
				}
				objectsToPropagate = append(objectsToPropagate, cleanObject(object))
			}
		}
	}

	return objectsToPropagate, nil
}

func (c *genericTransportController) createOrUpdateWrappedObject(ctx context.Context, namespace string, wrappedObject *unstructured.Unstructured) error {
	_, err := c.transportClient.Resource(c.wrappedObjectGVR).Namespace(namespace).Create(ctx, wrappedObject, metav1.CreateOptions{})
	if err != nil {
		if !strings.HasSuffix(err.Error(), alreadyExistsErrorSuffix) { // if object is already there, we need to use update. otherwise report an error.
			return fmt.Errorf("failed to create wrapped object '%s' in destination WEC with namespace '%s' - %w", wrappedObject.GetName(), namespace, err)
		}
		// if we reached here, create had an error that object already exists. try update object instead.
		_, err = c.transportClient.Resource(c.wrappedObjectGVR).Namespace(namespace).Update(ctx, wrappedObject, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update wrapped object '%s' in destination WEC with namespace '%s' - %w", wrappedObject.GetName(), namespace, err)
		}
	}

	return nil
}

// updateObjectFunc is a function that updates the given object. returns true if object was updated, otherwise false.
type updateObjectFunc func(*v1alpha1.PlacementDecision) bool

func (c *genericTransportController) updatePlacementDecision(ctx context.Context, placementDecision *v1alpha1.PlacementDecision, updateObjectFunc updateObjectFunc) error {
	if !updateObjectFunc(placementDecision) { // returns an indication if object was updated or not.
		return nil // if object was not updated, no need to update in API server, return.
	}

	_, err := c.wdsClientset.EdgeV1alpha1().PlacementDecisions().Update(ctx, placementDecision, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update PlacementDecision object in WDS - %w", err)
	}

	return nil
}

// addFinalizer accepts an object and adds the provided finalizer if not present.
// It returns an indication of whether it updated the object's list of finalizers.
func addFinalizer(object metav1.Object, finalizer string) (finalizersUpdated bool) {
	finalizers := object.GetFinalizers()
	for _, item := range finalizers {
		if item == finalizer {
			return false
		}
	}
	object.SetFinalizers(append(finalizers, finalizer))
	return true
}

// removeFinalizer accepts an object and removes the provided finalizer if present.
// It returns an indication of whether it updated the object's list of finalizers.
func removeFinalizer(object metav1.Object, finalizer string) (finalizersUpdated bool) {
	finalizersList := object.GetFinalizers()
	length := len(finalizersList)

	index := 0
	for i := 0; i < length; i++ {
		if finalizersList[i] == finalizer {
			continue
		}
		finalizersList[index] = finalizersList[i]
		index++
	}
	object.SetFinalizers(finalizersList[:index])
	return length != index
}

// setAnnotation sets metadata annotation on the given object.
func setAnnotation(object metav1.Object, key string, value string) {
	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[key] = value

	object.SetAnnotations(annotations)
}

// cleanObject is a function to clean object before adding it to a wrapped object. these fields shouldn't be propagated to WEC.
func cleanObject(object *unstructured.Unstructured) *unstructured.Unstructured {
	object.SetManagedFields(nil)
	object.SetFinalizers(nil)
	object.SetGeneration(0)
	object.SetOwnerReferences(nil)
	object.SetSelfLink("")
	object.SetResourceVersion("")
	object.SetUID("")
	delete(object.GetAnnotations(), "kubectl.kubernetes.io/last-applied-configuration")

	return object

}
