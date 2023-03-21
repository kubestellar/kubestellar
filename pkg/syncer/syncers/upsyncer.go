package syncers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/kcp-dev/edge-mc/pkg/syncer/shared"
	ddsif "github.com/kcp-dev/kcp/pkg/informer"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	controllerName = "kcp-edge-syncer-upsyncer"
)

type Controller struct {
	queue               workqueue.RateLimitingInterface
	downstreamClient    dynamic.Interface
	getDownstreamLister func(gvr schema.GroupVersionResource) (cache.GenericLister, error)
}

func NewUpsyncer(
	logger klog.Logger,
	downstreamClient dynamic.Interface,
	ddsifForDownstream *ddsif.GenericDiscoveringDynamicSharedInformerFactory[cache.SharedIndexInformer, cache.GenericLister, informers.GenericInformer],
) (*Controller, error) {
	c := &Controller{
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		downstreamClient: downstreamClient,
		getDownstreamLister: func(gvr schema.GroupVersionResource) (cache.GenericLister, error) {
			informers, notSynced := ddsifForDownstream.Informers()
			informer, ok := informers[gvr]
			if !ok {
				if ContainsGVR(notSynced, gvr) {
					return nil, fmt.Errorf("informer for gvr %v not synced in the downstream informer factory - should retry", gvr)
				}
				return nil, fmt.Errorf("gvr %v should be known in the downstream informer factory", gvr)
			}
			return informer.Lister(), nil
		},
	}
	ddsifForDownstream.AddEventHandler(
		ddsif.GVREventHandlerFuncs{
			AddFunc: func(gvr schema.GroupVersionResource, obj interface{}) {
				unstrObj, ok := obj.(*unstructured.Unstructured)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("resource should be a *unstructured.Unstructured, but was %T", unstrObj))
					return
				}
				c.AddToQueue(gvr, obj, logger)
			},
			UpdateFunc: func(gvr schema.GroupVersionResource, oldObj, newObj interface{}) {
				oldUnstrob := oldObj.(*unstructured.Unstructured)
				newUnstrob := newObj.(*unstructured.Unstructured)
				if !deepEqualFinalizersAndStatus(oldUnstrob, newUnstrob) {
					c.AddToQueue(gvr, newUnstrob, logger)
				}
			},
			DeleteFunc: func(gvr schema.GroupVersionResource, obj interface{}) {
				if d, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					obj = d.Obj
				}
				unstrObj, ok := obj.(*unstructured.Unstructured)
				if !ok {
					utilruntime.HandleError(fmt.Errorf("resource should be a *unstructured.Unstructured, but was %T", unstrObj))
					return
				}
				c.AddToQueue(gvr, obj, logger)
			},
		})
	return c, nil
}

type queueKey struct {
	gvr schema.GroupVersionResource
	key string // meta namespace key
}

func (c *Controller) AddToQueue(gvr schema.GroupVersionResource, obj interface{}, logger logr.Logger) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	shared.WithQueueKey(logger, key).V(2).Info("queueing GVR", "gvr", gvr.String())
	c.queue.Add(
		queueKey{
			gvr: gvr,
			key: key,
		},
	)
}

// Start starts N worker processes processing work items.
func (c *Controller) Start(ctx context.Context, numThreads int) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	logger := shared.WithReconciler(klog.FromContext(ctx), controllerName)
	ctx = klog.NewContext(ctx, logger)
	logger.Info("Starting syncer workers")
	defer logger.Info("Stopping syncer workers")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(ctx, c.startWorker, time.Second)
	}

	<-ctx.Done()
}

// startWorker processes work items until stopCh is closed.
func (c *Controller) startWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	qk := key.(queueKey)

	logger := shared.WithQueueKey(klog.FromContext(ctx), qk.key).WithValues("gvr", qk.gvr.String())
	ctx = klog.NewContext(ctx, logger)
	logger.V(1).Info("processing key")

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(key)

	if err := c.process(ctx, qk.gvr, qk.key); err != nil {
		utilruntime.HandleError(fmt.Errorf("%s failed to sync %q, err: %w", controllerName, key, err))
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)

	return true
}

func (c *Controller) process(ctx context.Context, gvr schema.GroupVersionResource, key string) error {
	logger := klog.FromContext(ctx)

	// from downstream
	downstreamNamespace, downstreamName, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Error(err, "Invalid key")
		return nil
	}

	logger = logger.WithValues("DownstreamNamespace", downstreamNamespace, "DownstreamName", downstreamName)

	downstreamLister, err := c.getDownstreamLister(gvr)
	if err != nil {
		return err
	}

	var resourceExists bool
	var obj runtime.Object
	if downstreamNamespace != "" {
		obj, err = downstreamLister.ByNamespace(downstreamNamespace).Get(downstreamName)
	} else {
		obj, err = downstreamLister.Get(downstreamName)
	}
	if err == nil {
		resourceExists = true
	} else if !apierrors.IsNotFound(err) {
		return err
	}

	logger = logger.WithValues("UpstreamNamespace", "ns", "UpstreamName", "name")
	ctx = klog.NewContext(ctx, logger)

	_ = ctx
	_ = obj

	if resourceExists {
		// transfer obj from downstream to upstream
	}
	return nil
}

func ContainsGVR(gvrs []schema.GroupVersionResource, gvr schema.GroupVersionResource) bool {
	for _, item := range gvrs {
		if gvr == item {
			return true
		}
	}
	return false
}

func deepEqualFinalizersAndStatus(oldUnstrob, newUnstrob *unstructured.Unstructured) bool {
	newFinalizers := newUnstrob.GetFinalizers()
	oldFinalizers := oldUnstrob.GetFinalizers()

	newStatus := newUnstrob.UnstructuredContent()["status"]
	oldStatus := oldUnstrob.UnstructuredContent()["status"]

	return equality.Semantic.DeepEqual(oldFinalizers, newFinalizers) && equality.Semantic.DeepEqual(oldStatus, newStatus)
}
