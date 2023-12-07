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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	// edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgev2alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev2alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
)

const (
	controllerName            = "transport-controller"
	transportCleanupFinalizer = "kubestellar.io/transport-object-cleanup"
)

// NewTransportController returns a new transport controller
func NewTransportController(ctx context.Context, edgePlacementDecisionInformer edgev2alpha1informers.EdgePlacementDecisionInformer, transport Transport) *genericTransportController {
	logger := klog.FromContext(ctx).WithName(controllerName)

	transportController := &genericTransportController{
		edgePlacementDecisionLister:         edgePlacementDecisionInformer.Lister(),
		edgePlacementDecisionInformerSynced: edgePlacementDecisionInformer.Informer().HasSynced,
		workqueue:                           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		transport:                           transport,
	}

	logger.Info("Setting up event handlers")
	// Set up an event handler for when EdgePlacementDecision resources change
	edgePlacementDecisionInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: transportController.enqueueEdgePlacementDecision,
		UpdateFunc: func(old, new interface{}) {
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
	edgePlacementDecisionLister         edgev2alpha1listers.EdgePlacementDecisionLister
	edgePlacementDecisionInformerSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be processed instead of performing it as soon as a change happens.
	// This means we can ensure we only process a fixed amount of resources at a time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	//transport is a specific implementation for the transport interface.
	transport Transport
}

// enqueueEdgePlacementDecision takes an EdgePlacementDecision resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be passed resources of any type other than EdgePlacementDecision.
func (ctrl *genericTransportController) enqueueEdgePlacementDecision(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
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
	logger := klog.FromContext(ctx)

	logger.Info("Starting transport controller")

	// Wait for the caches to be synced before starting workers
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), ctrl.transport.InformerSynced, ctrl.edgePlacementDecisionInformerSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workersCount)
	// Launch workers to process EdgePlacementDecision
	for i := 0; i < workersCount; i++ {
		go wait.UntilWithContext(ctx, ctrl.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (ctrl *genericTransportController) runWorker(ctx context.Context) {
	for ctrl.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (ctrl *genericTransportController) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := ctrl.workqueue.Get()
	if shutdown {
		return false
	}

	if err := ctrl.process(ctx, obj); err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (ctrl *genericTransportController) process(ctx context.Context, obj interface{}) error {
	// We call Done here so the workqueue knows we have finished
	// processing this item. We also must remember to call Forget if we
	// do not want this work item being re-queued. For example, we do
	// not call Forget if a transient error occurs, instead the item is
	// put back on the workqueue and attempted again after a back-off
	// period.
	defer ctrl.workqueue.Done(obj)
	var key string
	var ok bool
	// We expect strings to come off the workqueue. These are of the
	// form namespace/name. We do this as the delayed nature of the
	// workqueue means the items in the informer cache may actually be
	// more up to date that when the item was initially put onto the
	// workqueue.
	if key, ok = obj.(string); !ok {
		// As the item in the workqueue is actually invalid, we call
		// Forget here else we'd go into a loop of attempting to
		// process a work item that is invalid.
		ctrl.workqueue.Forget(obj)
		utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		return nil
	}
	// Run the syncHandler, passing it the namespace/name string of the EdgePlacementDecision resource to be synced.
	if err := ctrl.syncHandler(ctx, key); err != nil {
		// Put the item back on the workqueue to handle any transient errors.
		ctrl.workqueue.AddRateLimited(key)
		return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
	}
	// Finally, if no error occurs we Forget this item so it does not
	// get queued again until another change happens.
	ctrl.workqueue.Forget(obj)

	return nil
}

// syncHandler compares the actual state with the desired, and attempts to converge the two.
func (ctrl *genericTransportController) syncHandler(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	logger := klog.LoggerWithValues(klog.FromContext(ctx), "objectName", key)

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// TODO implement business logic

	// TODO Create/Update logic

	// Get the EdgePlacementDecision object with this name
	edgePlacementDecision, err := ctrl.edgePlacementDecisionLister.Get(name)
	if err != nil {
		// The EdgePlacementDecision object may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("EdgePlacementDecision '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	// TODO Put finalizer on the EdgePlacementDecision in case it doesn't have already (we need to make sure it's not deleted before we clean up the WrappedObj)
	if addFinalizer(edgePlacementDecision, transportCleanupFinalizer) {
		// ...
	}
	// TODO Fetch all workload objects from WDS
	// TODO Wrap the objects in wrapper object - ctrl.transport.WrapObjects(...
	// TODO parse WECs destinations from EdgePlacementDecision Object
	// TODO put a copy of the wrapped object in every namespace of the destination WEC.

	// TODO Delete logic

	// TODO delete all wrapped objects from list of destinations
	// TODO remove finalizer

	logger.Info("Successfully synced")
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

// helper function to check if object is being deleted - we then should clean wrapped objects from WECs namespaces and only then remove finalizer.
func isObjectBeingDeleted(object metav1.Object) bool {
	return !object.GetDeletionTimestamp().IsZero()
}

// helper function to clean object before adding it to a wrapped object. these fields shouldn't be propagated to WEC.
func cleanObject(object metav1.Object) {
	object.SetManagedFields(nil)
	object.SetFinalizers(nil)
	object.SetGeneration(0)
	object.SetOwnerReferences(nil)
	object.SetSelfLink("")
	object.SetResourceVersion("")
	object.SetUID("")
}
