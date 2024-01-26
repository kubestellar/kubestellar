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
	ctrlm "sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kubestellar/kubestellar/api/edge/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// Status controller watches workstatues and checks associated placements for singleton status. If
// a placement that cuase an object to be delivered to a cluster has singleton statsus specified
// the full status will be copied to the object.
type Controller struct {
	ctx                context.Context
	logger             logr.Logger
	wdsName            string
	wdsDynClient       *dynamic.DynamicClient
	wdsKubeClient      *kubernetes.Clientset
	imbsDynClient      *dynamic.DynamicClient
	imbsKubeClient     *kubernetes.Clientset
	workStatusInformer cache.SharedIndexInformer
	workStatusLister   cache.GenericLister
	placementInformer  cache.SharedIndexInformer
	placementLister    cache.GenericLister
	workqueue          workqueue.RateLimitingInterface
	// all wds listers/informers are required to retrieve objects and update status
	// without having to re-create new caches for this coontroller
	listers   map[string]*cache.GenericLister
	informers map[string]*cache.SharedIndexInformer
}

// Create a new  status controller
func NewController(mgr ctrlm.Manager, wdsRestConfig *rest.Config, imbsRestConfig *rest.Config,
	wdsName string, listers map[string]*cache.GenericLister,
	informers map[string]*cache.SharedIndexInformer) (*Controller, error) {
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

	controller := &Controller{
		wdsName:        wdsName,
		logger:         mgr.GetLogger(),
		wdsDynClient:   wdsDynClient,
		wdsKubeClient:  wdsKubeClient,
		imbsDynClient:  imbsDynClient,
		imbsKubeClient: imbsKubeClient,
		workqueue:      workqueue.NewRateLimitingQueue(ratelimiter),
		listers:        listers,
		informers:      informers,
	}

	return controller, nil
}

// Start the status controller
func (c *Controller) Start(workers int) error {
	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- c.run(workers)
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
func (c *Controller) run(workers int) error {
	defer c.workqueue.ShutDown()

	// start informers
	go c.startPlacementInformer()
	go c.startWorkStatusInformer()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// wait for all informers caches to be synced
	c.logger.Info("waiting for caches to sync")

	// we can't use cache.WaitForCacheSync here because informers are started by the placement controllers
	// and may not be started yet, thus WaitForCacheSync throws an exception. PollUntilContextCancel is
	// a safer approach here.
	for _, informer := range c.informers {
		wait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
			if informer != nil && (*informer).HasSynced() {
				return true, nil
			}
			return false, nil
		})
	}

	if ok := cache.WaitForCacheSync(ctx.Done(), (c.placementInformer).HasSynced); !ok {
		return fmt.Errorf("failed to wait for placement caches to sync")
	}

	if ok := cache.WaitForCacheSync(ctx.Done(), (c.workStatusInformer).HasSynced); !ok {
		return fmt.Errorf("failed to wait for workstatus caches to sync")
	}

	c.logger.Info("All caches synced")

	c.logger.Info("Starting workers", "count", workers)
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}
	c.logger.Info("Started workers")

	<-ctx.Done()
	c.logger.Info("Shutting down workers")

	return nil
}

func (c *Controller) startPlacementInformer() {
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(c.wdsDynClient, 0*time.Minute)

	gvr := schema.GroupVersionResource{Group: v1alpha1.GroupVersion.Group,
		Version:  v1alpha1.GroupVersion.Version,
		Resource: util.PlacementResource}

	c.placementInformer = informerFactory.ForResource(gvr).Informer()
	c.placementLister = cache.NewGenericLister(c.placementInformer.GetIndexer(), gvr.GroupResource())

	stopper := make(chan struct{})
	defer close(stopper)
	informerFactory.Start(stopper)

	<-stopper
}

func (c *Controller) startWorkStatusInformer() {
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

	stopper := make(chan struct{})
	defer close(stopper)
	informerFactory.Start(stopper)

	<-stopper
}

func shouldSkipUpdate(old, new interface{}) bool {
	oldMObj := old.(metav1.Object)
	newMObj := new.(metav1.Object)
	// do not enqueue update events for objects that have not changed
	return newMObj.GetResourceVersion() == oldMObj.GetResourceVersion()
}

func shouldSkipDelete(obj interface{}) bool {
	return false
}

// Event handler: enqueues the objects to be processed
// At this time it is very simple, more complex processing might be required here
func (c *Controller) handleObject(obj any) {
	mObj := obj.(metav1.Object)
	rObj := obj.(runtime.Object)
	ok := rObj.GetObjectKind()
	gvk := ok.GroupVersionKind()
	c.logger.V(2).Info("Got object event", gvk.GroupVersion().String(), gvk.Kind, mObj.GetNamespace(), mObj.GetName())
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
	if _, ok := obj.(metav1.Object).GetLabels()[util.PlacementLabelSingletonStatus]; !ok {
		return nil
	}

	sourceRef, err := util.GetWorkStatusSourceRef(obj)
	if err != nil {
		return err
	}

	status, err := util.GetWorkStatusStatus(obj)
	if err != nil {
		// status gets updated after workstatus is created, it's ok to requeue
		return err
	}

	c.logger.Info("updating singleton status", "kind", sourceRef.Kind, "name", sourceRef.Name, "namespace", sourceRef.Namespace)
	return updateObjectStatus(sourceRef, status, c.listers, c.wdsDynClient)
}

func updateObjectStatus(objRef *util.SourceRef, status map[string]interface{},
	listers map[string]*cache.GenericLister, wdsDynClient *dynamic.DynamicClient) error {

	key := util.KeyForGroupVersionKind(objRef.Group, objRef.Version, objRef.Kind)

	lister, ok := listers[key]
	if !ok {
		return fmt.Errorf("could not find lister for GVK key %s", key)
	}

	obj, err := getObject(*lister, objRef.Namespace, objRef.Name)
	if err != nil {
		return err
	}

	unstrObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("object cannot be cast to *unstructured.Unstructured: object: %s", util.GenerateObjectInfoString(obj))
	}

	// set the status and update the object
	unstrObj.Object["status"] = status

	gvr := schema.GroupVersionResource{Group: objRef.Group, Version: objRef.Version, Resource: objRef.Resource}
	if objRef.Namespace == "" {
		_, err = wdsDynClient.Resource(gvr).UpdateStatus(context.Background(), unstrObj, metav1.UpdateOptions{})
	} else {
		_, err = wdsDynClient.Resource(gvr).Namespace(objRef.Namespace).UpdateStatus(context.Background(), unstrObj, metav1.UpdateOptions{})
	}
	if err != nil {
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
