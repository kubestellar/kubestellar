package controller

import (
	"context"
	"fmt"
	"time"

	edgev1alpha1typed "github.com/kcp-dev/edge-mc/pkg/syncer/client/clientset/versioned/typed/edge/v1alpha1"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/syncer/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/syncer/client/listers/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/syncer/shared"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"k8s.io/klog/v2"
)

const (
	controllerName = "kcp-edge-syncconfig-controller"
)

// NewSyncConfigController returns a controller watching a [edge.kcp.io.EdgeSyncConfig] and update down/up syncer,
func NewSyncConfigController(
	logger klog.Logger,
	syncConfigClient edgev1alpha1typed.EdgeSyncConfigInterface,
	syncConfigInformer edgev1alpha1informers.EdgeSyncConfigInformer,
	syncConfigUID types.UID,
	syncConfigName string,
	downstreamDynamicClient dynamic.Interface,
	downstreamSyncerDiscoveryClient discovery.DiscoveryInterface,
) (*controller, error) {
	c := &controller{
		queue:                           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		syncConfigUID:                   syncConfigUID,
		syncConfigLister:                syncConfigInformer.Lister(),
		syncConfigInformerHasSynced:     syncConfigInformer.Informer().HasSynced,
		syncConfigClient:                syncConfigClient,
		downstreamDynamicClient:         downstreamDynamicClient,
		downstreamSyncerDiscoveryClient: downstreamSyncerDiscoveryClient,
	}
	syncConfigInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				return false
			}
			_, name, err := cache.SplitMetaNamespaceKey(key)
			if err != nil {
				return false
			}
			return name == syncConfigName
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { c.enqueueSyncConfig(obj, logger) },
			UpdateFunc: func(old, obj interface{}) { c.enqueueSyncConfig(obj, logger) },
			DeleteFunc: func(obj interface{}) { c.enqueueSyncConfig(obj, logger) },
		},
	})

	return c, nil
}

// controller is a control loop that watches EdgeSyncConfig.
type controller struct {
	queue                           workqueue.RateLimitingInterface
	syncConfigUID                   types.UID
	syncConfigLister                edgev1alpha1listers.EdgeSyncConfigLister
	syncConfigInformerHasSynced     cache.InformerSynced
	syncConfigClient                edgev1alpha1typed.EdgeSyncConfigInterface
	downstreamDynamicClient         dynamic.Interface
	downstreamSyncerDiscoveryClient discovery.DiscoveryInterface
}

// Ready returns true if the controller is ready to return the GVRs to sync.
// It implements [informer.GVRSource.Ready].
func (c *controller) Ready() bool {
	return c.syncConfigInformerHasSynced()
}

func (c *controller) enqueueSyncConfig(obj interface{}, logger klog.Logger) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	shared.WithQueueKey(logger, key).V(2).Info("queueing SyncConfig")
	c.queue.Add(key)
}

// Start starts the controller workers.
func (c *controller) Start(ctx context.Context, numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	logger := shared.WithReconciler(klog.FromContext(ctx), controllerName)
	ctx = klog.NewContext(ctx, logger)
	logger.Info("Starting controller")
	defer logger.Info("Shutting down controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(ctx, c.startWorker, time.Second)
	}

	<-ctx.Done()
}

func (c *controller) startWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *controller) processNextWorkItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	k, quit := c.queue.Get()
	if quit {
		return false
	}
	key := k.(string)

	logger := klog.FromContext(ctx).WithValues("key", key)
	ctx = klog.NewContext(ctx, logger)
	logger.V(1).Info("processing key")

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(key)

	if err := c.process(ctx, key); err != nil {
		runtime.HandleError(fmt.Errorf("%q controller failed to sync %q, err: %w", controllerName, key, err))
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)
	return true
}

func (c *controller) process(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Error(err, "failed to split key, dropping")
		return nil
	}

	syncConfig, err := c.syncConfigLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil // object deleted before we handled it
		}
		return err
	}

	if syncConfig.GetUID() != c.syncConfigUID {
		return nil
	}

	return nil
}
