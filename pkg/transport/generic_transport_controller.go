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
	ControllerName                  = "transport-controller"
	transportFinalizer              = "transport.kubestellar.io/object-cleanup"
	originOwnerReferenceLabel       = "transport.kubestellar.io/originOwnerReferencePlacementDecisionKey"
	originWdsLabel                  = "transport.kubestellar.io/originWdsName"
	originOwnerGenerationAnnotation = "transport.kubestellar.io/originOwnerReferencePlacementDecisionGeneration"
)

// NewTransportController returns a new transport controller
func NewTransportController(ctx context.Context, placementDecisionInformer edgev1alpha1informers.PlacementDecisionInformer, transport Transport,
	wdsClientset edgeclientset.Interface, wdsDynamicClient dynamic.Interface, transportClientset kubernetes.Interface,
	transportDynamicClient dynamic.Interface, wdsName string) (*genericTransportController, error) {
	emptyWrappedObject := transport.WrapObjects(make([]*unstructured.Unstructured, 0)) // empty wrapped object to get GVR from it.
	wrappedObjectGVR, err := getGvrFromWrappedObject(transportClientset, emptyWrappedObject)
	if err != nil {
		return nil, fmt.Errorf("failed to get wrapped object GVR - %w", err)
	}

	dynamicInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(transportDynamicClient, 0)
	wrappedObjectGenericInformer := dynamicInformerFactory.ForResource(wrappedObjectGVR)

	transportController := &genericTransportController{
		logger:                          klog.FromContext(ctx),
		placementDecisionLister:         placementDecisionInformer.Lister(),
		placementDecisionInformerSynced: placementDecisionInformer.Informer().HasSynced,
		wrappedObjectInformerSynced:     wrappedObjectGenericInformer.Informer().HasSynced,
		workqueue:                       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ControllerName),
		transport:                       transport,
		transportClient:                 transportDynamicClient,
		wrappedObjectGVR:                wrappedObjectGVR,
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
	dynamicInformerFactory.Start(ctx.Done())

	return transportController, nil
}

func convertObjectToUnstructured(object runtime.Object) (*unstructured.Unstructured, error) {
	unstructuredObject, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil, fmt.Errorf("failed to convert given object to unstructured - %w", err)
	}
	return &unstructured.Unstructured{Object: unstructuredObject}, nil
}

func getGvrFromWrappedObject(clientset kubernetes.Interface, wrappedObject runtime.Object) (schema.GroupVersionResource, error) {
	unstructuredWrappedObject, err := convertObjectToUnstructured(wrappedObject)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to convert wrapped object to unstructured - %w", err)
	}

	gvk := unstructuredWrappedObject.GroupVersionKind()
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(cacheddiscovery.NewMemCacheClient(clientset.Discovery()))

	restMapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to discover GroupVersionResource from given GroupVersionKind - %w", err)
	}

	return restMapping.Resource, nil
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

	transport        Transport         //transport is a specific implementation for the transport interface.
	transportClient  dynamic.Interface // dynamic client to transport wrapped object. since object kind is unknown during complilation, we use dynamic
	wrappedObjectGVR schema.GroupVersionResource

	wdsClientset     edgeclientset.Interface
	wdsDynamicClient dynamic.Interface
	wdsName          string
}

// enqueuePlacementDecision takes an PlacementDecision resource and
// converts it into a namespace/name string which is put onto the workqueue.
// This func *shouldn't* handle any resource other than PlacementDecision.
func (c *genericTransportController) enqueuePlacementDecision(obj interface{}) {
	var key string
	var err error
	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
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
	ownerPlacementDecisionKey, found := wrappedObject.GetLabels()[originOwnerReferenceLabel] // safe if GetLabels() returns nil
	if !found {
		c.logger.Info("failed to extract placementdecision key from wrapped object", "Name", wrappedObject.GetName(),
			"Namespace", wrappedObject.GetNamespace())
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

	if errors.IsNotFound(err) { // the object was deleted and it had no finalizer on it. this means transport controller
		// finished cleanup of wrapped objects from mailbox namespaces. no need to do anything in this state.
		return nil
	}
	if err != nil { // in case of a different error, log it and retry.
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
	for _, destination := range placementDecision.Spec.Destinations {
		if err := c.deleteWrappedObject(ctx, destination.ClusterId, fmt.Sprintf("%s-%s", placementDecision.GetName(), c.wdsName)); err != nil {
			// wrapped object name is in the format (PlacementDecision.GetName()-WdsName). see updateWrappedObject func for explanation.
			return fmt.Errorf("failed to delete wrapped object from all destinations' - %w", err)
		}
	}

	if err := c.removeFinalizerFromPlacementDecision(ctx, placementDecision); err != nil {
		return fmt.Errorf("failed to remove finalizer from PlacementDecision object '%s' - %w", placementDecision.GetName(), err)
	}

	return nil
}

func (c *genericTransportController) deleteWrappedObject(ctx context.Context, namespace string, objectName string) error {
	err := c.transportClient.Resource(c.wrappedObjectGVR).Namespace(namespace).Delete(ctx, objectName, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) { // if object is already not there, we do not report an error cause desired state was achieved.
		return fmt.Errorf("failed to delete wrapped object '%s' from destination WEC mailbox namespace '%s' - %w", objectName, namespace, err)
	}
	return nil
}

func (c *genericTransportController) removeFinalizerFromPlacementDecision(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) error {
	return c.updatePlacementDecision(ctx, placementDecision, func(placementDecision *v1alpha1.PlacementDecision) (*v1alpha1.PlacementDecision, bool) {
		return removeFinalizer(placementDecision, transportFinalizer)
	})
}

func (c *genericTransportController) addFinalizerToPlacementDecision(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) error {
	return c.updatePlacementDecision(ctx, placementDecision, func(placementDecision *v1alpha1.PlacementDecision) (*v1alpha1.PlacementDecision, bool) {
		return addFinalizer(placementDecision, transportFinalizer)
	})
}

func (c *genericTransportController) updateWrappedObjectsAndFinalizer(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) error {
	if err := c.addFinalizerToPlacementDecision(ctx, placementDecision); err != nil {
		return fmt.Errorf("failed to add finalizer to PlacementDecision object '%s' - %w", placementDecision.GetName(), err)
	}
	// get current state
	currentWrappedObjectList, err := c.transportClient.Resource(c.wrappedObjectGVR).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s,%s=%s", originOwnerReferenceLabel, placementDecision.GetName(), originWdsLabel, c.wdsName),
	})
	if err != nil {
		return fmt.Errorf("failed to get current wrapped objects that are owned by PlacementDecision '%s' - %w", placementDecision.GetName(), err)
	}
	// calculate desired state
	desiredWrappedObject, err := c.initializeWrappedObject(ctx, placementDecision)
	if err != nil {
		return fmt.Errorf("failed to build wrapped object from PlacementDecision '%s' - %w", placementDecision.GetName(), err)
	}
	// converge actual state to the desired state
	for _, destination := range placementDecision.Spec.Destinations {
		currentWrappedObject := c.popWrappedObjectByNamespace(currentWrappedObjectList, destination.ClusterId)
		if currentWrappedObject != nil && currentWrappedObject.GetAnnotations() != nil &&
			currentWrappedObject.GetAnnotations()[originOwnerGenerationAnnotation] == desiredWrappedObject.GetAnnotations()[originOwnerGenerationAnnotation] {
			continue // current wrapped object is already in the desired state
		}
		// othereise, need to create or update the wrapped object
		if err := c.createOrUpdateWrappedObject(ctx, destination.ClusterId, desiredWrappedObject); err != nil {
			return fmt.Errorf("failed to propagate wrapped object '%s' to all required WECs - %w", desiredWrappedObject.GetName(), err)
		}
	}
	// all objects that appear in the desired state were handled. need to remove wrapped objects that are not part of the desired state
	for _, wrappedObject := range currentWrappedObjectList.Items { // objects left in currentWrappedObjectList.Items have to be deleted
		if err := c.deleteWrappedObject(ctx, wrappedObject.GetNamespace(), wrappedObject.GetName()); err != nil {
			return fmt.Errorf("failed to delete wrapped object from destinations that were removed from desired state - %w", err)
		}
	}

	return nil
}

func (c *genericTransportController) getObjectsFromWDS(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) ([]*unstructured.Unstructured, error) {
	objectsToPropagate := make([]*unstructured.Unstructured, 0)
	// add cluster-scoped objects to the 'objectsToPropagate' slice
	for _, clusterScopedObject := range placementDecision.Spec.Workload.ClusterScope {
		if clusterScopedObject.ObjectNames == nil {
			continue // no objects from this gvr, skip
		}
		gvr := schema.GroupVersionResource{Group: clusterScopedObject.Group, Version: clusterScopedObject.Version, Resource: clusterScopedObject.Resource}
		gvrDynamicClient := c.wdsDynamicClient.Resource(gvr)
		for _, objectName := range clusterScopedObject.ObjectNames {
			object, err := gvrDynamicClient.Get(ctx, objectName, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to get required cluster-scoped object '%s' with gvr %s from WDS - %w", objectName, gvr, err)
			}
			objectsToPropagate = append(objectsToPropagate, cleanObject(object))
		}
	}
	// add namespace-scoped objects to the 'objectsToPropagate' slice
	for _, namespaceScopedObject := range placementDecision.Spec.Workload.NamespaceScope {
		gvr := schema.GroupVersionResource{Group: namespaceScopedObject.Group, Version: namespaceScopedObject.Version, Resource: namespaceScopedObject.Resource}
		gvrDynamicClient := c.wdsDynamicClient.Resource(gvr)
		for _, objectsByNamespace := range namespaceScopedObject.ObjectsByNamespace {
			if objectsByNamespace.Names == nil {
				continue // no objects from this namespace, skip
			}
			for _, objectName := range objectsByNamespace.Names {
				object, err := gvrDynamicClient.Namespace(objectsByNamespace.Namespace).Get(ctx, objectName, metav1.GetOptions{})
				if err != nil {
					return nil, fmt.Errorf("failed to get required namespace-scoped object '%s' in namespace '%s' with gvr '%s' from WDS - %w", objectName,
						objectsByNamespace.Namespace, gvr, err)
				}
				objectsToPropagate = append(objectsToPropagate, cleanObject(object))
			}
		}
	}

	return objectsToPropagate, nil
}

func (c *genericTransportController) initializeWrappedObject(ctx context.Context, placementDecision *v1alpha1.PlacementDecision) (*unstructured.Unstructured, error) {
	objectsToPropagate, err := c.getObjectsFromWDS(ctx, placementDecision)
	if err != nil {
		return nil, fmt.Errorf("failed to get objects to propagate to WECs from PlacementDecision object '%s' - %w", placementDecision.GetName(), err)
	}

	wrappedObject, err := convertObjectToUnstructured(c.transport.WrapObjects(objectsToPropagate))
	if err != nil {
		return nil, fmt.Errorf("failed to convert wrapped object to unstructured - %w", err)
	}
	// wrapped object name is (PlacementDecision.GetName()-WdsName).
	// pay attention - we cannot use the PlacementDecision object name, cause we might have duplicate names coming from different WDS spaces.
	// we add WdsName to the object name to assure name uniqueness,
	// in order to easily get the origin PlacementDecision object name and wds, we add it as an annotations.
	wrappedObject.SetName(fmt.Sprintf("%s-%s", placementDecision.GetName(), c.wdsName))
	setLabel(wrappedObject, originOwnerReferenceLabel, placementDecision.GetName())
	setLabel(wrappedObject, originWdsLabel, c.wdsName)
	setAnnotation(wrappedObject, originOwnerGenerationAnnotation, placementDecision.GetGeneration())

	return wrappedObject, nil
}

// pops wrapped object by namespace from the list and returns the requested wrapped object.
// if the object is not found, list remains the same and nil is returned.
// since the order of items in the list is not important, the implementation is efficient and was done as follows:
// the functions goes over the list, if the requested object is found, it's replaced with the last object in the list.
// then the function removes the last object in the list and returns the object.
// in worst case where object is not found, it will go over all items in the list.
func (c *genericTransportController) popWrappedObjectByNamespace(list *unstructured.UnstructuredList, namespace string) *unstructured.Unstructured {
	length := len(list.Items)
	for i := 0; i < length; i++ {
		if list.Items[i].GetNamespace() == namespace {
			requiredObject := list.Items[i]
			list.Items[i] = list.Items[length-1]
			list.Items = list.Items[:length-1]
			return &requiredObject
		}
	}

	return nil
}

func (c *genericTransportController) createOrUpdateWrappedObject(ctx context.Context, namespace string, wrappedObject *unstructured.Unstructured) error {
	existingWrappedObject, err := c.transportClient.Resource(c.wrappedObjectGVR).Namespace(namespace).Get(ctx, wrappedObject.GetName(), metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) { // if object is not there, we need to create it. otherwise report an error.
			return fmt.Errorf("failed to create wrapped object '%s' in destination WEC with namespace '%s' - %w", wrappedObject.GetName(), namespace, err)
		}
		// object not found when using get, create it
		_, err = c.transportClient.Resource(c.wrappedObjectGVR).Namespace(namespace).Create(ctx, wrappedObject, metav1.CreateOptions{
			FieldManager: ControllerName,
		})
		if err != nil {
			return fmt.Errorf("failed to create wrapped object '%s' in destination WEC mailbox namespace '%s' - %w", wrappedObject.GetName(), namespace, err)
		}
		return nil
	}
	// // if we reached here object already exists, try update object
	wrappedObject.SetResourceVersion(existingWrappedObject.GetResourceVersion())
	_, err = c.transportClient.Resource(c.wrappedObjectGVR).Namespace(namespace).Update(ctx, wrappedObject, metav1.UpdateOptions{
		FieldManager: ControllerName,
	})
	if err != nil {
		return fmt.Errorf("failed to update wrapped object '%s' in destination WEC mailbox namespace '%s' - %w", wrappedObject.GetName(), namespace, err)
	}

	return nil
}

// updateObjectFunc is a function that updates the given object.
// returns the updated object (if it was updated) or the object as is if it wasn't, and true if object was updated, or false otherwise.
type updateObjectFunc func(*v1alpha1.PlacementDecision) (*v1alpha1.PlacementDecision, bool)

func (c *genericTransportController) updatePlacementDecision(ctx context.Context, placementDecision *v1alpha1.PlacementDecision, updateObjectFunc updateObjectFunc) error {
	updatedPlacementDecision, isUpdated := updateObjectFunc(placementDecision) // returns an indication if object was updated or not.
	if !isUpdated {
		return nil // if object was not updated, no need to update in API server, return.
	}

	_, err := c.wdsClientset.EdgeV1alpha1().PlacementDecisions().Update(ctx, updatedPlacementDecision, metav1.UpdateOptions{
		FieldManager: ControllerName,
	})
	if err != nil {
		return fmt.Errorf("failed to update PlacementDecision object '%s' in WDS - %w", placementDecision.GetName(), err)
	}

	return nil
}

// addFinalizer accepts PlacementDecision object and adds the provided finalizer if not present.
// It returns the updated (or not) PlacementDecision and an indication of whether it updated the object's list of finalizers.
func addFinalizer(placementDecision *v1alpha1.PlacementDecision, finalizer string) (*v1alpha1.PlacementDecision, bool) {
	finalizers := placementDecision.GetFinalizers()
	for _, item := range finalizers {
		if item == finalizer { // finalizer already exists, no need to add
			return placementDecision, false
		}
	}
	// if we reached here, finalizer has to be added to the placement decision object.
	// objects returned from a PlacementDecisionLister must be treated as read-only.
	// Therefore, create a deep copy before updating the object.
	updatedPlacementDecision := placementDecision.DeepCopy()
	updatedPlacementDecision.SetFinalizers(append(finalizers, finalizer))
	return updatedPlacementDecision, true
}

// removeFinalizer accepts PlacementDecision object and removes the provided finalizer if present.
// It returns the updated (or not) PlacementDecision and an indication of whether it updated the object's list of finalizers.
func removeFinalizer(placementDecision *v1alpha1.PlacementDecision, finalizer string) (*v1alpha1.PlacementDecision, bool) {
	finalizersList := placementDecision.GetFinalizers()
	length := len(finalizersList)

	index := 0
	for i := 0; i < length; i++ {
		if finalizersList[i] == finalizer {
			continue
		}
		finalizersList[index] = finalizersList[i]
		index++
	}
	if length == index { // finalizer wasn't found, no need to remove
		return placementDecision, false
	}
	// otherwise, finalizer was found and has to be removed.
	// objects returned from a PlacementDecisionLister must be treated as read-only.
	// Therefore, create a deep copy before updating the object.
	updatedPlacementDecision := placementDecision.DeepCopy()
	updatedPlacementDecision.SetFinalizers(finalizersList[:index])
	return updatedPlacementDecision, true
}

// setAnnotation sets metadata annotation on the given object.
func setAnnotation(object metav1.Object, key string, value any) {
	annotations := object.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[key] = fmt.Sprint(value)

	object.SetAnnotations(annotations)
}

// setLabel sets metadata label on the given object.
func setLabel(object metav1.Object, key string, value any) {
	labels := object.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	labels[key] = fmt.Sprint(value)

	object.SetLabels(labels)
}

// cleanObject is a function to clean object before adding it to a wrapped object. these fields shouldn't be propagated to WEC.
func cleanObject(object *unstructured.Unstructured) *unstructured.Unstructured {
	objectCopy := object.DeepCopy() // don't modify object directly. create a copy before zeroing fields
	objectCopy.SetManagedFields(nil)
	objectCopy.SetFinalizers(nil)
	objectCopy.SetGeneration(0)
	objectCopy.SetOwnerReferences(nil)
	objectCopy.SetSelfLink("")
	objectCopy.SetResourceVersion("")
	objectCopy.SetUID("")

	annotations := objectCopy.GetAnnotations()
	delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
	objectCopy.SetAnnotations(annotations)

	return objectCopy

}
