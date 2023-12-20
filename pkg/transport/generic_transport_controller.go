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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgev2alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev2alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
	msclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
)

const (
	controllerName           = "transport-controller"
	transportFinalizer       = "kubestellar.io/transport-object-cleanup"
	notFoundErrorSuffix      = "not found"
	alreadyExistsErrorSuffix = "already exists"
)

// NewTransportController returns a new transport controller
func NewTransportController(ctx context.Context, edgePlacementDecisionInformer edgev2alpha1informers.EdgePlacementDecisionInformer, transport Transport,
	kubeBindSpaceRelation kbuser.KubeBindSpaceRelation, spaceManagementClient msclient.KubestellarSpaceInterface, spaceProviderNs string,
	transportSpaceClient client.Client) *genericTransportController {

	transportController := &genericTransportController{
		logger:                              klog.FromContext(ctx).WithName(controllerName),
		edgePlacementDecisionLister:         edgePlacementDecisionInformer.Lister(),
		edgePlacementDecisionInformerSynced: edgePlacementDecisionInformer.Informer().HasSynced,
		workqueue:                           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		transport:                           transport,
		kubeBindSpaceRelation:               kubeBindSpaceRelation,
		spaceManagementClient:               spaceManagementClient,
		spaceProviderNs:                     spaceProviderNs,
		transportSpaceClient:                transportSpaceClient,
	}

	transportController.logger.Info("Setting up event handlers")
	// Set up an event handler for when EdgePlacementDecision resources change
	edgePlacementDecisionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: transportController.enqueueEdgePlacementDecision,
		UpdateFunc: func(_, new interface{}) {
			transportController.enqueueEdgePlacementDecision(new)
		},
		DeleteFunc: transportController.enqueueEdgePlacementDecision,
	})

	// Set up an event handler for when WrappedObject resources change. This handler will lookup the origin EdgePlacementDecision
	// of the given WrappedObject and enqueue that EdgePlacementDecision resource for processing.
	// This way, we don't need to implement custom logic for handling WrappedObjects resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md

	// TODO transport.SetupEventHandlers()
	// (we should have another entry point for the transport specific)

	return transportController
}

type genericTransportController struct {
	logger logr.Logger

	edgePlacementDecisionLister         edgev2alpha1listers.EdgePlacementDecisionLister
	edgePlacementDecisionInformerSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be processed instead of performing it as soon as a change happens.
	// This means we can ensure we only process a fixed amount of resources at a time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	//transport is a specific implementation for the transport interface.
	transport Transport

	kubeBindSpaceRelation kbuser.KubeBindSpaceRelation
	spaceManagementClient msclient.KubestellarSpaceInterface
	spaceProviderNs       string
	transportSpaceClient  client.Client
}

// enqueueEdgePlacementDecision takes an EdgePlacementDecision resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be passed resources of any type other than EdgePlacementDecision.
func (ctrl *genericTransportController) enqueueEdgePlacementDecision(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		ctrl.logger.Error(err, "failed to enqueue EdgePlacementDecision object")
		return
	}
	ctrl.workqueue.Add(key)
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (ctrl *genericTransportController) Run(ctx context.Context, workersCount int) error {
	defer utilruntime.HandleCrash()
	defer ctrl.workqueue.ShutDown()

	ctrl.logger.Info("starting transport controller")

	// Wait for the caches to be synced before starting workers
	ctrl.logger.Info("waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), ctrl.transport.InformerSynced, ctrl.edgePlacementDecisionInformerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	ctrl.logger.Info("starting workers", "count", workersCount)
	// Launch workers to process EdgePlacementDecision
	for i := 1; i <= workersCount; i++ {
		go wait.UntilWithContext(ctx, func(ctx context.Context) { ctrl.runWorker(ctx, i) }, time.Second)
	}

	ctrl.logger.Info("started workers")
	<-ctx.Done()
	ctrl.logger.Info("shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (ctrl *genericTransportController) runWorker(ctx context.Context, workerId int) {
	for ctrl.processNextWorkItem(ctx, workerId) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (ctrl *genericTransportController) processNextWorkItem(ctx context.Context, workerID int) bool {
	obj, shutdown := ctrl.workqueue.Get()
	if shutdown {
		return false
	}

	if err := ctrl.process(ctx, obj); err != nil {
		ctrl.logger.Error(err, "failed to process EdgePlacementDecision object", "objectKey", obj, "workerID", workerID)
	} else {
		ctrl.logger.Info("synced EdgePlacementDecision object successfully.", "objectKey", obj, "workerID", workerID)
	}

	return true
}

func (ctrl *genericTransportController) process(ctx context.Context, obj interface{}) error {
	// We call Done here so the workqueue knows we have finished processing this item.
	// We also must remember to call Forget if we do not want this work item being re-queued.
	// For example, we do not call Forget if a transient error occurs, instead the item is
	// put back on the workqueue and attempted again after a back-off period.
	defer ctrl.workqueue.Done(obj)
	var key string
	var ok bool
	// We expect strings to come off the workqueue. These are of the form namespace/name.
	// We do this as the delayed nature of the workqueue means the items in the informer cache
	// may actually be more up to date that when the item was initially put onto the workqueue.
	if key, ok = obj.(string); !ok {
		// As the item in the workqueue is actually invalid, we call Forget here else we'd go
		// into a loop of attempting to process a work item that is invalid.
		ctrl.workqueue.Forget(obj)
		return fmt.Errorf("expected key from type string in workqueue but got %#v", obj)
	}

	_, objectName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		// As the item in the workqueue is actually invalid, we call Forget here else we'd go
		// into a loop of attempting to process a work item that is invalid.
		ctrl.workqueue.Forget(obj)
		return fmt.Errorf("invalid object key '%s' - %w", key, err)
	}

	// Run the syncHandler, passing it the object name of the EdgePlacementDecision resource to be synced.
	if err := ctrl.syncHandler(ctx, objectName); err != nil {
		// Put the item back on the workqueue to handle any transient errors.
		ctrl.workqueue.AddRateLimited(key)
		return fmt.Errorf("failed to process object with key '%s' - %w", key, err)
	}
	// Finally, if no error occurs we Forget this item so it does not
	// get queued again until another change happens.
	ctrl.workqueue.Forget(obj)

	return nil
}

// syncHandler compares the actual state with the desired, and attempts to converge the two.
// returning an error from this function will result in a requeue of the given object key.
// therefore, if object shouldn't be requeued, don't return error.
func (ctrl *genericTransportController) syncHandler(ctx context.Context, objectName string) error {
	// Get the EdgePlacementDecision object with this name (from KubeStellar Core Space)
	edgePlacementDecision, err := ctrl.edgePlacementDecisionLister.Get(objectName)
	if err != nil {
		if errors.IsNotFound(err) { // the object was deleted and it had no finalizer on it. this means transport controller
			// finished cleanup of wrapped objects from mailbox namespaces. no need to do anything in this state.
			return nil
		}
		// in case of a different error, log it and retry.
		return fmt.Errorf("failed to get provider's copy of EdgePlacementDecision object '%s' - %w", objectName, err)
	}

	if isObjectBeingDeleted(edgePlacementDecision) {
		return ctrl.deleteWrappedObjectsAndFinalizer(ctx, edgePlacementDecision)
	}
	// otherwise, object was not deleted and no error occurered while reading provider's copy of the object.
	return ctrl.updateWrappedObjectsAndFinalizer(ctx, edgePlacementDecision)
}

// isObjectBeingDeleted is a helper function to check if object is being deleted.
func isObjectBeingDeleted(object metav1.Object) bool {
	return !object.GetDeletionTimestamp().IsZero()
}

func (ctrl *genericTransportController) deleteWrappedObjectsAndFinalizer(ctx context.Context, edgePlacementDecision *v2alpha1.EdgePlacementDecision) error {
	_, originEdgePlacementDecisionName, kbSpaceID, err := kbuser.AnalyzeObjectID(edgePlacementDecision)
	if err != nil {
		return fmt.Errorf("object does not appear to be a provider's copy of a consumer's object - %w", err)
	}

	emptyWrappedObject := ctrl.transport.WrapObjects(make([]*unstructured.Unstructured, 0))
	// wrapped object name is identical to the kube-bind copy of EdgePlacementDecision object name, see explanation in the updateWrappedObject function
	emptyWrappedObject.SetName(edgePlacementDecision.GetName())

	for _, destination := range edgePlacementDecision.Spec.Destinations {
		emptyWrappedObject.SetNamespace(destination.Namespace) // TODO need to revisit the destination struct and see how to get namespace from it properly
		if err := ctrl.transportSpaceClient.Delete(ctx, emptyWrappedObject); err != nil {
			if !strings.HasSuffix(err.Error(), notFoundErrorSuffix) { // if object is already not there, we do not report an error cause desired state was achieved.
				return fmt.Errorf("failed to delete wrapped object '%s' in destination WEC with namespace '%s' - %w", edgePlacementDecision.GetName(), destination.Namespace, err)
			}
		}
	}

	if err := ctrl.removeFinalizerFromOriginEdgePlacementDecision(ctx, originEdgePlacementDecisionName, kbSpaceID); err != nil {
		return fmt.Errorf("failed to remove finalizer from EdgePlacementDecision object '%s' - %w", originEdgePlacementDecisionName, err)
	}

	return nil
}

func (ctrl *genericTransportController) removeFinalizerFromOriginEdgePlacementDecision(ctx context.Context, originEdgePlacementDecisionName string, kbSpaceID string) error {
	return ctrl.updateOriginEdgePlacementDecision(ctx, originEdgePlacementDecisionName, kbSpaceID, func(edgePlacementDecision *v2alpha1.EdgePlacementDecision) bool {
		return removeFinalizer(edgePlacementDecision, transportFinalizer)
	})
}

func (ctrl *genericTransportController) addFinalizerToOriginEdgePlacementDecision(ctx context.Context, originEdgePlacementDecisionName string, kbSpaceID string) error {
	return ctrl.updateOriginEdgePlacementDecision(ctx, originEdgePlacementDecisionName, kbSpaceID, func(edgePlacementDecision *v2alpha1.EdgePlacementDecision) bool {
		return addFinalizer(edgePlacementDecision, transportFinalizer)
	})
}

func (ctrl *genericTransportController) updateWrappedObjectsAndFinalizer(ctx context.Context, edgePlacementDecision *v2alpha1.EdgePlacementDecision) error {
	_, originEdgePlacementDecisionName, kbSpaceID, err := kbuser.AnalyzeObjectID(edgePlacementDecision)
	if err != nil {
		return fmt.Errorf("object does not appear to be a provider's copy of a consumer's object - %w", err)
	}

	if err := ctrl.addFinalizerToOriginEdgePlacementDecision(ctx, originEdgePlacementDecisionName, kbSpaceID); err != nil {
		return fmt.Errorf("failed to add finalizer to EdgePlacementDecision object '%s' - %w", originEdgePlacementDecisionName, err)
	}

	objectsToPropagate, err := ctrl.getObjectsFromOriginWDS(ctx, kbSpaceID, edgePlacementDecision)
	if err != nil {
		return fmt.Errorf("failed to get objects to propagate to WECs from EdgePlacementDecision object '%s' - %w", edgePlacementDecision.GetName(), err)
	}

	wrappedObject := ctrl.transport.WrapObjects(objectsToPropagate)
	// wrapped object name is identical to the kube-bind EdgePlacementDecision object name
	// need to pay attention - we cannot use the origin EdgePlacementDecision object name, cause we might have duplicate names coming from different WDS spaces.
	// we use the kube-bind object name to assure name uniqueness,
	// in order to get the originEdgePlacementDecision object name and kbSpaceID from the wrapped object, one can use the above `kbuser.analyzeObjectID` func.
	wrappedObject.SetName(edgePlacementDecision.GetName())

	for _, destination := range edgePlacementDecision.Spec.Destinations {
		wrappedObject.SetNamespace(destination.Namespace) // TODO need to revisit the destination struct and see how to use it properly
		objectBytes, err := wrappedObject.MarshalJSON()
		if err != nil {
			return fmt.Errorf("failed to marshal wrapped object '%s' - %w", edgePlacementDecision.GetName(), err)
		}
		forceChanges := true
		if err := ctrl.transportSpaceClient.Patch(ctx, wrappedObject, client.RawPatch(types.ApplyPatchType, objectBytes), &client.PatchOptions{
			FieldManager: controllerName,
			Force:        &forceChanges,
		}); err != nil {
			return fmt.Errorf("failed to update wrapped object '%s' in destination WEC with namespace '%s' - %w", edgePlacementDecision.GetName(), destination.Namespace, err)
		}
	}

	return nil
}

func (ctrl *genericTransportController) getObjectsFromOriginWDS(ctx context.Context, kbSpaceID string,
	edgePlacementDecision *v2alpha1.EdgePlacementDecision) ([]*unstructured.Unstructured, error) {
	spaceConfig, err := ctrl.getOriginSpaceConfig(kbSpaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get space config of WDS from kube-bind space id '%s' - %w", kbSpaceID, err)
	}

	dynamicClient, err := dynamic.NewForConfig(spaceConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic k8s clientset for WDS - %w", err)
	}

	objectsToPropagate := make([]*unstructured.Unstructured, 0)
	// add cluster-scoped objects to the 'objectsToPropagate' slice
	for _, clusterScopedObject := range edgePlacementDecision.Spec.Workload.ClusterScope {
		if clusterScopedObject.Objects == nil {
			continue // no objects from this gvr, skip
		}
		// TODO we should use k8s runtime gvr directly in clusterScoped Object struct! do not inherit from existing code this unless there is good reason for that.
		gvr := schema.GroupVersionResource{Group: clusterScopedObject.Group, Version: clusterScopedObject.APIVersion, Resource: clusterScopedObject.Resource}
		for _, objectName := range clusterScopedObject.Objects {
			object, err := dynamicClient.Resource(gvr).Get(ctx, objectName, metav1.GetOptions{})
			if err != nil {
				return nil, fmt.Errorf("failed to get required cluster-scoped object '%s' with gvr %s from WDS - %w", objectName, gvr, err)
			}
			objectsToPropagate = append(objectsToPropagate, cleanObject(object))
		}
	}
	// add namespace-scoped objects to the 'objectsToPropagate' slice
	for _, namespaceScopedObject := range edgePlacementDecision.Spec.Workload.NamespaceScope {
		// TODO we should use k8s runtime gvr directly in namespaceScoped Object struct! do not inherit from existing code this unless there is good reason for that.
		gvr := schema.GroupVersionResource{Group: namespaceScopedObject.Group, Version: namespaceScopedObject.APIVersion, Resource: namespaceScopedObject.Resource}
		for _, objectsByNamespace := range namespaceScopedObject.ObjectsByNamespace {
			if objectsByNamespace.Names == nil {
				continue // no objects from this namespace, skip
			}
			for _, objectName := range objectsByNamespace.Names {
				object, err := dynamicClient.Resource(gvr).Namespace(objectsByNamespace.Namespace).Get(ctx, objectName, metav1.GetOptions{})
				if err != nil {
					return nil, fmt.Errorf("failed to get required namespace-scoped object '%s' with gvr '%s' from WDS - %w", objectName, gvr, err)
				}
				objectsToPropagate = append(objectsToPropagate, cleanObject(object))
			}
		}
	}

	return objectsToPropagate, nil
}

// updateObjectFunc is a function that updates the given object. returns true if object was updated, otherwise false.
type updateObjectFunc func(*v2alpha1.EdgePlacementDecision) bool

// pay attention that given edgePlacementDecision is the kube-bind provider's copy of the original object from WDS.
func (ctrl *genericTransportController) updateOriginEdgePlacementDecision(ctx context.Context, originEdgePlacementDecisionName string, kbSpaceID string,
	updateObjectFunc updateObjectFunc) error {
	spaceConfig, err := ctrl.getOriginSpaceConfig(kbSpaceID)
	if err != nil {
		return fmt.Errorf("failed to get config for space from consumer space ID - %w", err)
	}

	edgeClientset, err := edgeclientset.NewForConfig(spaceConfig)
	if err != nil {
		return fmt.Errorf("failed to create clientset for consumer space - %w", err)
	}

	originEdgePlacementDecision, err := edgeClientset.EdgeV2alpha1().EdgePlacementDecisions().Get(ctx, originEdgePlacementDecisionName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get consumer's object - %w", err)
	}

	if !updateObjectFunc(originEdgePlacementDecision) { // returns an indication if object was updated or not.
		return nil // if object was not updated, no need to update in API server, return.
	}

	_, err = edgeClientset.EdgeV2alpha1().EdgePlacementDecisions().Update(ctx, originEdgePlacementDecision, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update consumer's object - %w", err)
	}

	return nil
}

func (ctrl *genericTransportController) getOriginSpaceConfig(kbSpaceID string) (*rest.Config, error) {
	spaceID := ctrl.kubeBindSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if spaceID == "" {
		return nil, fmt.Errorf("failed to get consumer space ID from a provider's copy")
	}

	spaceConfig, err := ctrl.spaceManagementClient.ConfigForSpace(spaceID, ctrl.spaceProviderNs)
	if err != nil {
		return nil, fmt.Errorf("failed to get config for space from consumer space ID - %w", err)
	}

	return spaceConfig, nil
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
