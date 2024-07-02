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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/binding"
	ksclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	ksinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions"
	controllisters "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const (
	controllerName      = "Status"
	defaultResyncPeriod = time.Duration(0)
	queueingDelay       = 2 * time.Second
	originWdsLabelKey   = "transport.kubestellar.io/originWdsName"
)

// Controller watches workstatues and checks whether the corresponding
// workload object asks for the singleton status returning. If yes,
// the full status will be copied to the workload object in WDS.
type Controller struct {
	logger       logr.Logger
	wdsName      string
	wdsDynClient dynamic.Interface
	wdsKsClient  ksclient.Interface
	itsDynClient dynamic.Interface

	statusCollectorInformer cache.SharedIndexInformer
	statusCollectorLister   controllisters.StatusCollectorLister
	combinedStatusInformer  cache.SharedIndexInformer
	combinedStatusLister    controllisters.CombinedStatusLister
	workStatusInformer      cache.SharedIndexInformer
	workStatusLister        cache.GenericLister
	workStatusIndexer       cache.Indexer
	workqueue               workqueue.RateLimitingInterface
	// all wds listers are used to retrieve objects and update status
	// without having to re-create new caches for this controller
	listers util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister]

	bindingResolutionBroker binding.ResolutionBroker
	combinedStatusResolver  CombinedStatusResolver
}

// bindingRef is a workqueue item that references a Binding
type bindingRef string

// workStatusRef is a workqueue item that references a WorkStatus
type workStatusRef struct {
	// name is the name of the WorkStatus object
	name string
	// wecName is the WorkStatus namespace
	wecName string
	// sourceObjectIdentifier is the identifier of the source object
	sourceObjectIdentifier util.ObjectIdentifier
}

// combinedStatusRef is a workqueue item that references a CombinedStatus
type combinedStatusRef string

// statusCollectorRef is a workqueue item that references a StatusCollector
type statusCollectorRef string

// Create a new  status controller
func NewController(wdsRestConfig *rest.Config, itsRestConfig *rest.Config, wdsName string,
	bindingResolutionBroker binding.ResolutionBroker) (*Controller, error) {
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	wdsDynClient, err := dynamic.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	itsDynClient, err := dynamic.NewForConfig(itsRestConfig)
	if err != nil {
		return nil, err
	}

	wdsKsClient, err := ksclient.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	zapLogger := zap.New(zap.UseDevMode(true))
	log.SetLogger(zapLogger)

	controller := &Controller{
		wdsName:                 wdsName,
		logger:                  log.Log.WithName(controllerName),
		wdsDynClient:            wdsDynClient,
		wdsKsClient:             wdsKsClient,
		itsDynClient:            itsDynClient,
		workqueue:               workqueue.NewRateLimitingQueue(ratelimiter),
		bindingResolutionBroker: bindingResolutionBroker,
	}

	celEvaluator, err := newCELEvaluator()
	if err != nil {
		return controller, err
	}
	resolver := NewCombinedStatusResolver(celEvaluator)

	controller.combinedStatusResolver = resolver
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

	c.bindingResolutionBroker.RegisterCallback(func(bindingPolicyKey string) {
		// add binding to workqueue
		c.workqueue.Add(bindingRef(bindingPolicyKey))
	}) // this will have the broker call the callback for all existing resolutions

	go c.runWorkStatusInformer(ctx)

	ksInformerFactory := ksinformers.NewSharedInformerFactory(c.wdsKsClient, defaultResyncPeriod)
	if err := c.setupStatusCollectorInformer(ctx, ksInformerFactory); err != nil {
		return err
	}
	if err := c.setupCombinedStatusInformer(ctx, ksInformerFactory); err != nil {
		return err
	}
	ksInformerFactory.Start(ctx.Done())
	if ok := cache.WaitForCacheSync(ctx.Done(), c.statusCollectorInformer.HasSynced, c.combinedStatusInformer.HasSynced); !ok {
		return fmt.Errorf("failed to wait for KubeStellar informers to sync")
	}

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

func (c *Controller) setupStatusCollectorInformer(ctx context.Context, ksInformerFactory ksinformers.SharedInformerFactory) error {
	logger := klog.FromContext(ctx)
	c.statusCollectorInformer = ksInformerFactory.Control().V1alpha1().StatusCollectors().Informer()
	c.statusCollectorLister = ksInformerFactory.Control().V1alpha1().StatusCollectors().Lister()
	_, err := c.statusCollectorInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			sc := obj.(*v1alpha1.StatusCollector)
			logger.V(5).Info("Enqueuing reference to StatusCollector because of informer add event", "name", sc.Name, "resourceVersion", sc.ResourceVersion)
			c.workqueue.Add(statusCollectorRef(sc.Name))
		},
		UpdateFunc: func(old, new interface{}) {
			oldSc := old.(*v1alpha1.StatusCollector)
			newSc := new.(*v1alpha1.StatusCollector)
			if oldSc.Generation != newSc.Generation {
				logger.V(5).Info("Enqueuing reference to StatusCollector because of informer update event", "name", newSc.Name, "resourceVersion", newSc.ResourceVersion)
				c.workqueue.Add(statusCollectorRef(newSc.Name))
			}
		},
		DeleteFunc: func(obj interface{}) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			sc := obj.(*v1alpha1.StatusCollector)
			logger.V(5).Info("Enqueuing reference to StatusCollector because of informer delete event", "name", sc.Name)
			c.workqueue.Add(statusCollectorRef(sc.Name))
		},
	})
	if err != nil {
		c.logger.Error(err, "failed to add statuscollectors informer event handler")
		return err
	}
	return nil
}

func (c *Controller) setupCombinedStatusInformer(ctx context.Context, ksInformerFactory ksinformers.SharedInformerFactory) error {
	logger := klog.FromContext(ctx)
	c.combinedStatusInformer = ksInformerFactory.Control().V1alpha1().CombinedStatuses().Informer()
	c.combinedStatusLister = ksInformerFactory.Control().V1alpha1().CombinedStatuses().Lister()
	_, err := c.combinedStatusInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			cs := obj.(*v1alpha1.CombinedStatus)
			logger.V(5).Info("Enqueuing reference to CombinedStatus because of informer add event",
				"name", cs.Name, "resourceVersion", cs.ResourceVersion)
			c.enqueueCombinedStatus(cs)
		},
		UpdateFunc: func(old, new interface{}) {
			oldCs := old.(*v1alpha1.CombinedStatus)
			newCs := new.(*v1alpha1.CombinedStatus)
			if oldCs.Generation != newCs.Generation {
				logger.V(5).Info("Enqueuing reference to CombinedStatus because of informer update event",
					"name", newCs.Name, "resourceVersion", newCs.ResourceVersion)
				c.enqueueCombinedStatus(newCs)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			cs := obj.(*v1alpha1.CombinedStatus)
			logger.V(5).Info("Enqueuing reference to CombinedStatus because of informer delete event",
				"name", cs.Name)
			c.enqueueCombinedStatus(cs)
		},
	})
	if err != nil {
		c.logger.Error(err, "failed to add combinedstatuses informer event handler")
		return err
	}
	return nil
}

func (c *Controller) enqueueCombinedStatus(obj metav1.Object) {
	key := cache.MetaObjectToName(obj).String()
	c.workqueue.AddAfter(combinedStatusRef(key), queueingDelay)
}

func (c *Controller) runWorkStatusInformer(ctx context.Context) {
	logger := klog.FromContext(ctx)

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(c.itsDynClient, 0*time.Minute)

	gvr := schema.GroupVersionResource{Group: util.WorkStatusGroup,
		Version:  util.WorkStatusVersion,
		Resource: util.WorkStatusResource}

	c.workStatusInformer = informerFactory.ForResource(gvr).Informer()
	// add indexer on key from (wecName, sourceRef) for workstatus fetching efficiency
	c.workStatusInformer.AddIndexers(cache.Indexers{
		workStatusIdentificationIndexKey: func(obj interface{}) ([]string, error) {
			wecName := obj.(metav1.Object).GetNamespace()
			sourceRef, err := util.GetWorkStatusSourceRef(obj.(runtime.Object))
			if err != nil {
				logger.Error(err, "Failed to get source ref",
					"object", util.RefToRuntimeObj(obj.(runtime.Object)))

				return nil, nil
			}

			return []string{util.KeyFromSourceRefAndWecName(sourceRef, wecName)}, nil
		}})
	c.workStatusIndexer = c.workStatusInformer.GetIndexer()
	c.workStatusLister = cache.NewGenericLister(c.workStatusInformer.GetIndexer(), gvr.GroupResource())

	// add the event handler functions
	c.workStatusInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.handleWorkStatus,
		UpdateFunc: func(old, new interface{}) {
			if shouldSkipUpdate(old, new) {
				return
			}
			c.handleWorkStatus(new)
		},
		DeleteFunc: func(obj interface{}) {
			if shouldSkipDelete(obj) {
				return
			}
			c.handleWorkStatus(obj)
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

// Event handler: enqueues the workstatus objects to be processed
// At this time it is very simple, more complex processing might be required here
func (c *Controller) handleWorkStatus(obj any) {
	wsRef, err := runtimeObjectToWorkStatusRef(obj.(runtime.Object))
	if err != nil {
		utilruntime.HandleError(err)
		return
	}

	c.logger.V(5).Info("Enqueuing reference to WorkStatus because of informer event",
		"sourceObjectName", wsRef.sourceObjectIdentifier.ObjectName,
		"sourceObjectGVK", wsRef.sourceObjectIdentifier.GVK, "wecName", wsRef.wecName)

	c.workqueue.Add(*wsRef)
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
	item, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(item interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(item)

		// Run the reconciler, passing it the full key of the metav1 Object
		if err := c.reconcile(ctx, item); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(item)
			return fmt.Errorf("error syncing key '%s': %s, requeuing", item, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(item)
		c.logger.V(2).Info("Successfully synced", "object", item)
		return nil
	}(item)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) reconcile(ctx context.Context, item any) error {
	logger := klog.FromContext(ctx)

	switch ref := item.(type) {
	case workStatusRef:
		return c.syncWorkStatus(ctx, ref)
	case bindingRef:
		return c.syncBinding(ctx, string(ref))
	case statusCollectorRef:
		return c.syncStatusCollector(ctx, string(ref))
	case combinedStatusRef:
		return c.syncCombinedStatus(ctx, string(ref))
	}
	logger.Error(nil, "Impossible workqueue entry", "type", fmt.Sprintf("%T", item), "value", item)
	return nil
}
