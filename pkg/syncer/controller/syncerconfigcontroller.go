package controller

import (
	"context"
	"time"

	edgev1alpha1typed "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/typed/edge/v1alpha1"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

func NewSyncerConfigController(
	logger klog.Logger,
	syncerConfigClient edgev1alpha1typed.SyncerConfigInterface,
	syncerConfigInformer edgev1alpha1informers.SyncerConfigInformer,
	syncerConfigManager *SyncerConfigManager,
	reconcileInterval time.Duration,
) (*syncerConfigController, error) {
	rateLimitter := workqueue.NewItemFastSlowRateLimiter(reconcileInterval, reconcileInterval*5, 1000)
	c := &syncerConfigController{
		syncerConfigLister:  syncerConfigInformer.Lister(),
		syncerConfigClient:  syncerConfigClient,
		syncerConfigManager: syncerConfigManager,
	}
	controllerName := "kcp-edge-syncerconfig-controller"
	controllerBase := &controllerBase{
		name:              controllerName,
		target:            "SyncerConfig",
		queue:             workqueue.NewNamedRateLimitingQueue(rateLimitter, controllerName),
		informerHasSynced: syncerConfigInformer.Informer().HasSynced,
		process:           c.process,
	}
	c.controllerBase = controllerBase
	syncerConfigInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				return false
			}
			_, _, err = cache.SplitMetaNamespaceKey(key)
			return err == nil
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { c.enqueue(obj, logger) },
			UpdateFunc: func(old, obj interface{}) { c.enqueue(obj, logger) },
			DeleteFunc: func(obj interface{}) { c.enqueue(obj, logger) },
		},
	})

	return c, nil
}

// controller is a control loop that watches EdgeSyncConfig.
type syncerConfigController struct {
	*controllerBase
	syncerConfigLister  edgev1alpha1listers.SyncerConfigLister
	syncerConfigClient  edgev1alpha1typed.SyncerConfigInterface
	syncerConfigManager *SyncerConfigManager
}

func (c *syncerConfigController) process(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Error(err, "failed to split key, dropping")
		return nil
	}

	syncerConfig, err := c.syncerConfigLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) { // object is deleted
			c.syncerConfigManager.delete(name)
			return nil
		}
		return err
	}
	c.syncerConfigManager.upsert(*syncerConfig)

	return nil
}
