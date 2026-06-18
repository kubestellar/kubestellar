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
	"sync"
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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/abstract"
	"github.com/kubestellar/kubestellar/pkg/binding"
	ksclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	ksinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions"
	controllisters "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
	ksmetrics "github.com/kubestellar/kubestellar/pkg/metrics"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const (
	ControllerName      = "Status"
	defaultResyncPeriod = time.Duration(0)
	queueingDelay       = 5 * time.Second
	originWdsLabelKey   = "transport.kubestellar.io/originWdsName"
)

// Controller watches workstatues and checks whether the corresponding
// workload object asks for the singleton status returning. If yes,
// the full status will be copied to the workload object in WDS.
type Controller struct {
	wdsName               string
	wdsDynClient          dynamic.Interface
	wdsKsClient           ksclient.Interface
	bindingPolicyClient   ksmetrics.ClientModNamespace[*v1alpha1.BindingPolicy, *v1alpha1.BindingPolicyList]
	bindingClient         ksmetrics.ClientModNamespace[*v1alpha1.Binding, *v1alpha1.BindingList]
	statusCollectorClient ksmetrics.ClientModNamespace[*v1alpha1.StatusCollector, *v1alpha1.StatusCollectorList]
	combinedStatusClient  ksmetrics.BasicNamespacedClient[*v1alpha1.CombinedStatus, *v1alpha1.CombinedStatusList]
	itsDynClient          dynamic.Interface

	bindingLister           controllisters.BindingLister
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

	celEvaluator           *celEvaluator
	bindingPolicyResolver  binding.BindingPolicyResolver
	combinedStatusResolver CombinedStatusResolver

	// workStatusToObject maps the namespace/name of WorkStatus to the ID of its workload object.
	// This map has entries for WorkStatus objects that exist.
	// This map is safe for concurrent access, but
	// the Set values revealed by workStatusToObject.ReadInverse can not be retained outside of
	// the funcs that are given them.
	workStatusToObject abstract.MutableMapToComparable[cache.ObjectName, util.ObjectIdentifier]

	mutex sync.RWMutex // used in workStatusToObject
}

type workloadObjectRef struct{ util.ObjectIdentifier }

// bindingRef is a workqueue item that references a Binding
type bindingRef string

type workStatusON cache.ObjectName

// workStatusRef is a workqueue item that references a WorkStatus
type workStatusRef struct {
	// Name is the Name of the WorkStatus object
	Name string
	// WECName is the WorkStatus namespace
	WECName string
	// SourceObjectIdentifier is the identifier of the source object
	SourceObjectIdentifier util.ObjectIdentifier
}

func (wsr workStatusRef) ObjectName() cache.ObjectName {
	return cache.ObjectName{Namespace: wsr.WECName, Name: wsr.Name}
}

// combinedStatusRef is a workqueue item that references a CombinedStatus
type combinedStatusRef string

// statusCollectorRef is a workqueue item that references a StatusCollector
type statusCollectorRef string

// Create a new  status controller
func NewController(logger logr.Logger,
	wdsClientMetrics, itsClientMetrics ksmetrics.ClientMetrics,
	wdsRestConfig *rest.Config, itsRestConfig *rest.Config, wdsName string,
	bindingPolicyResolver binding.BindingPolicyResolver) (*Controller, error) {
	logger = logger.WithName(ControllerName)
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	wdsDynClientBase, err := dynamic.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}
	wdsDynClient := ksmetrics.NewWrappedDynamicClient(wdsClientMetrics, wdsDynClientBase)

	itsDynClientBase, err := dynamic.NewForConfig(itsRestConfig)
	if err != nil {
		return nil, err
	}
	itsDynClient := ksmetrics.NewWrappedDynamicClient(itsClientMetrics, itsDynClientBase)

	wdsKsClient, err := ksclient.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	controller := &Controller{
		wdsName:               wdsName,
		wdsDynClient:          wdsDynClient,
		wdsKsClient:           wdsKsClient,
		itsDynClient:          itsDynClient,
		bindingClient:         ksmetrics.NewWrappedClusterScopedClient(wdsClientMetrics, util.GetBindingGVR(), wdsKsClient.ControlV1alpha1().Bindings()),
		bindingPolicyClient:   ksmetrics.NewWrappedClusterScopedClient(wdsClientMetrics, util.GetBindingPolicyGVR(), wdsKsClient.ControlV1alpha1().BindingPolicies()),
		statusCollectorClient: ksmetrics.NewWrappedClusterScopedClient(wdsClientMetrics, v1alpha1.GroupVersion.WithResource("statuscollectors"), wdsKsClient.ControlV1alpha1().StatusCollectors()),
		combinedStatusClient: ksmetrics.NewWrappedBasicNamespacedClient(wdsClientMetrics, v1alpha1.GroupVersion.WithResource("combinedstatuses"), func(ns string) ksmetrics.BasicClientModNamespace[*v1alpha1.CombinedStatus, *v1alpha1.CombinedStatusList] {
			return wdsKsClient.ControlV1alpha1().CombinedStatuses(ns)
		}),
		workqueue:             workqueue.NewRateLimitingQueueWithConfig(ratelimiter, workqueue.RateLimitingQueueConfig{Name: ControllerName + "-" + wdsName}),
		bindingPolicyResolver: bindingPolicyResolver,
	}
	controller.workStatusToObject = abstract.NewLockedMapToComparable(&controller.mutex,
		abstract.NewPrimitiveMapToComparable[cache.ObjectName, util.ObjectIdentifier]())

	broker := controller.bindingPolicyResolver.Broker()
	klog.Infof("Registering callbacks with broker=%p=%v", broker, broker)
	err = broker.RegisterCallbacks(binding.ResolutionCallbacks{
		BindingPolicyChanged: func(bindingPolicyKey string) {
			// add binding to workqueue
			logger.V(5).Info("Enqueuing reference to Binding due to notification from BindingResolutionBroker", "name", bindingPolicyKey)
			controller.workqueue.Add(bindingRef(bindingPolicyKey))
		},
		ReportedStateRequestChanged: func(bindingPolicyKey string, objId util.ObjectIdentifier) {
			logger.V(5).Info("Enqueuing reference to workload object due to change in reported state request", "binding", bindingPolicyKey, "objId", objId)
			controller.workqueue.Add(workloadObjectRef{objId})
		}})
	if err != nil {
		return controller, err
	}
	logger.Info("Registered binding resolution broker callback")

	return controller, nil
}

func (c *Controller) HandleWorkloadObjectEvent(gvr schema.GroupVersionResource, oldObj, obj util.MRObject, eventType binding.WorkloadEventType, wasDeletedFinalStateUnknown bool) {
	objId := util.IdentifierForObject(obj, gvr.Resource)
	labels := obj.GetLabels()
	if _, hasLabel := labels[util.BindingPolicyLabelSingletonStatusKey]; hasLabel {
		c.workqueue.Add(workloadObjectRef{objId})
	}
}

// Start the status controller
func (c *Controller) Start(parentCtx context.Context, workers int, cListers chan interface{}) error {
	logger := klog.FromContext(parentCtx).WithName(ControllerName)
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
	logger := klog.FromContext(ctx)

	if err := c.ensureNamespaceExists(ctx, util.ClusterScopedObjectsCombinedStatusNamespace); err != nil {
		return fmt.Errorf("failed to ensure namespace (%s) for combinedstatuses associated with "+
			"cluster-scoped objects: %w", util.ClusterScopedObjectsCombinedStatusNamespace, err)
	}

	go c.runWorkStatusInformer(ctx)

	ksInformerFactory := ksinformers.NewSharedInformerFactory(c.wdsKsClient, defaultResyncPeriod)
	c.bindingLister = ksInformerFactory.Control().V1alpha1().Bindings().Lister()
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
	logger.Info("Received listers")

	celEvaluator, err := newCELEvaluator()
	if err != nil {
		return err
	}

	c.celEvaluator = celEvaluator
	c.combinedStatusResolver = NewCombinedStatusResolver(celEvaluator, c.listers)

	logger.Info("Starting workers", "count", workers)
	for i := 0; i < workers; i++ {
		workerCtx := klog.NewContext(ctx, logger.WithName(fmt.Sprintf("worker-%d", i)))
		go wait.UntilWithContext(workerCtx, c.runWorker, time.Second)
	}
	logger.Info("Started workers")

	<-ctx.Done()
	logger.Info("Shutting down workers")

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
		logger.Error(err, "failed to add statuscollectors informer event handler")
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
		logger.Error(err, "failed to add combinedstatuses informer event handler")
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
		AddFunc: func(obj interface{}) {
			if objNotInThisWDS(obj, c.wdsName) {
				return
			}
			c.handleWorkStatus(ctx, "add", obj)
		},
		UpdateFunc: func(old, new interface{}) {
			if objNotInThisWDS(new, c.wdsName) || shouldSkipUpdate(old, new) {
				return
			}
			c.handleWorkStatus(ctx, "update", new)
		},
		DeleteFunc: func(obj interface{}) {
			if objNotInThisWDS(obj, c.wdsName) {
				return
			}
			c.handleWorkStatus(ctx, "delete", obj)
		},
	})
	informerFactory.Start(ctx.Done())

	logger.Info("waiting for workstatus cache to sync")
	if ok := cache.WaitForCacheSync(ctx.Done(), c.workStatusInformer.HasSynced); !ok {
		logger.Info("failed to wait for workstatus caches to sync")
	}
	logger.Info("workstatus cache synced")

	<-ctx.Done()
}

func shouldSkipUpdate(old, new interface{}) bool {
	oldMObj := old.(metav1.Object)
	newMObj := new.(metav1.Object)
	// do not enqueue update events for objects that have not changed
	return newMObj.GetResourceVersion() == oldMObj.GetResourceVersion()
}

func objNotInThisWDS(obj interface{}, thisWDS string) bool {
	if objWDS, ok := obj.(metav1.Object).GetLabels()[originWdsLabelKey]; ok {
		if objWDS != thisWDS {
			return true
		}
	}
	return false
}

// Informer event handler: enqueues the workstatus objects to be processed
// At this time it is very simple, more complex processing might be required here
func (c *Controller) handleWorkStatus(ctx context.Context, eventType string, obj any) {
	logger := klog.FromContext(ctx)
	wsRef, err := runtimeObjectToWorkStatusRef(obj.(runtime.Object))
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	logger.V(5).Info("Enqueuing reference to WorkStatus because of informer event", "eventType", eventType,
		"sourceObjectName", wsRef.SourceObjectIdentifier.ObjectName,
		"sourceObjectGVK", wsRef.SourceObjectIdentifier.GVK, "wecName", wsRef.WECName)
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
		logger := klog.FromContext(ctx)
		logger.V(4).Info("Successfully synced", "object", item, "type", fmt.Sprintf("%T", item))
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
	case workloadObjectRef:
		return c.syncWorkloadObject(ctx, ref.ObjectIdentifier)
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

func (c *Controller) ensureNamespaceExists(ctx context.Context, ns string) error {
	namespaceGVR := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}

	_, err := c.wdsDynClient.Resource(namespaceGVR).Namespace(ns).Get(ctx, ns, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			namespaceObj := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Namespace",
					"metadata": map[string]interface{}{
						"name": ns,
					},
				},
			}
			_, err = c.wdsDynClient.Resource(namespaceGVR).Create(ctx, namespaceObj,
				metav1.CreateOptions{FieldManager: ControllerName})
			if err != nil && !errors.IsAlreadyExists(err) {
				return fmt.Errorf("failed to create namespace: %w", err)
			}
		} else {
			return fmt.Errorf("failed to get namespace: %w", err)
		}
	}

	return nil
}
