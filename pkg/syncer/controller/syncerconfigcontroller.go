/*
Copyright 2022 The KubeStellar Authors.

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

package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	edgev1alpha1typed "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/typed/edge/v1alpha1"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
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
