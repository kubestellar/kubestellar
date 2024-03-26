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

package status

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kubestellar/kubestellar/pkg/util"
)

const controllerName = "Status"

// Status controller watches workstatues and checks whether the corresponding
// workload object asks for the singleton status returning. If yes,
// the full status will be copied to the workload object in WDS.
type Controller struct {
	logger             logr.Logger
	wdsName            string
	wdsDynClient       dynamic.Interface
	wdsKubeClient      kubernetes.Interface
	imbsDynClient      dynamic.Interface
	imbsKubeClient     kubernetes.Interface
	workStatusInformer cache.SharedIndexInformer
	workStatusLister   cache.GenericLister
	workqueue          workqueue.RateLimitingInterface
	// all wds listers are used to retrieve objects and update status
	// without having to re-create new caches for this controller
	listers util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister]
}

// Create a new  status controller
func NewController(wdsRestConfig *rest.Config, imbsRestConfig *rest.Config, wdsName string) (*Controller, error) {
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	wdsDynClient, err := dynamic.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	wdsKubeClient, err := kubernetes.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	imbsDynClient, err := dynamic.NewForConfig(imbsRestConfig)
	if err != nil {
		return nil, err
	}

	imbsKubeClient, err := kubernetes.NewForConfig(imbsRestConfig)
	if err != nil {
		return nil, err
	}

	zapLogger := zap.New(zap.UseDevMode(true))
	log.SetLogger(zapLogger)

	controller := &Controller{
		wdsName:        wdsName,
		logger:         log.Log.WithName(controllerName),
		wdsDynClient:   wdsDynClient,
		wdsKubeClient:  wdsKubeClient,
		imbsDynClient:  imbsDynClient,
		imbsKubeClient: imbsKubeClient,
		workqueue:      workqueue.NewRateLimitingQueue(ratelimiter),
	}

	return controller, nil
}

// Start the status controller
func (c *Controller) Start(parentCtx context.Context, workers int, cListers chan interface{}) error {
	logger := klog.FromContext(parentCtx).WithName(controllerName)
	ctx := klog.NewContext(parentCtx, logger)
	errChan := make(chan error, 1)
	go func() {
		errChan <- c.run(ctx, workers, cListers)
	}()

	// check for errors at startup, after all started we let it continue
	// so we can start the controller-runtime manager
	select {
	case err := <-errChan:
		return err
	case <-time.After(3 * time.Second):
		return nil
	}
}

// Invoked by Start() to run the translator
func (c *Controller) run(ctx context.Context, workers int, cListers chan interface{}) error {
	defer c.workqueue.ShutDown()

	go c.runWorkStatusInformer(ctx)

	c.listers = (<-cListers).(util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister])
	c.logger.Info("Received listers")

	c.logger.Info("Starting workers", "count", workers)
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}
	c.logger.Info("Started workers")

	<-ctx.Done()
	c.logger.Info("Shutting down workers")

	return nil
}

func (c *Controller) runWorkStatusInformer(ctx context.Context) {
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(c.imbsDynClient, 0*time.Minute)

	gvr := schema.GroupVersionResource{Group: util.WorkStatusGroup,
		Version:  util.WorkStatusVersion,
		Resource: util.WorkStatusResource}

	c.workStatusInformer = informerFactory.ForResource(gvr).Informer()
	c.workStatusLister = cache.NewGenericLister(c.workStatusInformer.GetIndexer(), gvr.GroupResource())

	// add the event handler functions
	c.workStatusInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleObject,
		UpdateFunc: func(old, new interface{}) {
			if shouldSkipUpdate(old, new) {
				return
			}
			c.handleObject(new)
		},
		DeleteFunc: func(obj interface{}) {
			if shouldSkipDelete(obj) {
				return
			}
			c.handleObject(obj)
		},
	})

	informerFactory.Start(ctx.Done())

	c.logger.Info("waiting for workstatus cache to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.workStatusInformer.HasSynced); !ok {
		c.logger.Info("failed to wait for workstatus caches to sync")
	}
	c.logger.Info("workstatus cache synced")

	<-ctx.Done()
}

func shouldSkipUpdate(old, new interface{}) bool {
	oldMObj := old.(metav1.Object)
	newMObj := new.(metav1.Object)
	// do not enqueue update events for objects that have not changed
	return newMObj.GetResourceVersion() == oldMObj.GetResourceVersion()
}

func shouldSkipDelete(_ interface{}) bool {
	return false
}

// Event handler: enqueues the objects to be processed
// At this time it is very simple, more complex processing might be required here
func (c *Controller) handleObject(obj any) {
	rObj := obj.(runtime.Object)
	c.logger.V(2).Info("Got object event", "obj", util.RefToRuntimeObj(rObj))
	c.enqueueObject(obj)
}

// enqueueObject generates key and put it onto the work queue.
func (c *Controller) enqueueObject(obj interface{}) {
	ref, err := cache.ObjectToName(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(ref)
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem reads a single work item off the workqueue and
// attempt to process it by calling the reconcile.
func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)

		// We expect a cache.ObjectName to come off the workqueue. We do this as the delayed
		// nature of the workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		ref, ok := obj.(cache.ObjectName)
		if !ok {
			// if the item in the workqueue is invalid, we call
			// Forget here to avoid process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected a string key in the workqueue but got %#v", obj))
			return nil
		}
		// Run the reconciler, passing it the full key or the metav1 Object
		if err := c.reconcile(ctx, ref); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(obj)
			return fmt.Errorf("error syncing key '%s': %s, requeuing", obj, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		c.logger.V(2).Info("Successfully synced", "object", obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) reconcile(ctx context.Context, ref cache.ObjectName) error {
	obj, err := getObject(c.workStatusLister, ref.Namespace, ref.Name)
	if err != nil {
		// The resource no longer exist, which means it has been deleted.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("object %#v in work queue no longer exists", ref))
			return nil
		}
		return err
	}

	// only process workstatues with the label for single reported status
	statusLabelVal, ok := obj.(metav1.Object).GetLabels()[util.BindingPolicyLabelSingletonStatusKey]
	if !ok {
		return nil
	}

	sourceRef, err := util.GetWorkStatusSourceRef(obj)
	if err != nil {
		return err
	}

	// remove the status if singleton status label value is unset
	if statusLabelVal == util.BindingPolicyLabelSingletonStatusValueUnset {
		emptyStatus := make(map[string]interface{})
		return updateObjectStatus(ctx, sourceRef, emptyStatus, c.listers, c.wdsDynClient)
	}

	status, err := util.GetWorkStatusStatus(obj)
	if err != nil {
		// status gets updated after workstatus is created, it's ok to requeue
		return err
	}

	c.logger.Info("updating singleton status", "kind", sourceRef.Kind, "name", sourceRef.Name, "namespace", sourceRef.Namespace)
	return updateObjectStatus(ctx, sourceRef, status, c.listers, c.wdsDynClient)
}

func updateObjectStatus(ctx context.Context, objRef *util.SourceRef, status map[string]interface{},
	listers util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister], wdsDynClient dynamic.Interface) error {

	gvr := schema.GroupVersionResource{Group: objRef.Group, Version: objRef.Version, Resource: objRef.Resource}
	lister, found := listers.Get(gvr)
	if !found {
		return fmt.Errorf("could not find lister for gvr %s", gvr)
	}

	obj, err := getObject(lister, objRef.Namespace, objRef.Name)
	if err != nil {
		return err
	}

	unstrObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("object cannot be cast to *unstructured.Unstructured: object: %s", util.RefToRuntimeObj(obj))
	}

	// set the status and update the object
	unstrObj.Object["status"] = status

	if objRef.Namespace == "" {
		_, err = wdsDynClient.Resource(gvr).UpdateStatus(ctx, unstrObj, metav1.UpdateOptions{})
	} else {
		_, err = wdsDynClient.Resource(gvr).Namespace(objRef.Namespace).UpdateStatus(ctx, unstrObj, metav1.UpdateOptions{})
	}
	if err != nil {
		// if resource not found it may mean no status subresource - try to patch the status
		if errors.IsNotFound(err) {
			return util.PatchStatus(ctx, unstrObj, status, objRef.Namespace, gvr, wdsDynClient)
		}
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// TODO - move these to a common lib
func getObject(lister cache.GenericLister, namespace, name string) (runtime.Object, error) {
	if namespace != "" {
		return lister.ByNamespace(namespace).Get(name)
	}
	return lister.Get(name)
}
