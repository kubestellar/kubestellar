/*
Copyright 2022 The KCP Authors.

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
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	edgev1alpha1typed "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/typed/edge/v1alpha1"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/syncer/syncers"
)

// NewEdgeSyncConfigController returns a edgeSyncConfigController watching a [edge.kcp.io.EdgeSyncConfig] and update down/up syncer,
func NewEdgeSyncConfigController(
	logger klog.Logger,
	syncConfigClient edgev1alpha1typed.EdgeSyncConfigInterface,
	syncConfigInformer edgev1alpha1informers.EdgeSyncConfigInformer,
	syncConfigManager *SyncConfigManager,
	upSyncer syncers.SyncerInterface,
	downSyncer syncers.SyncerInterface,
	reconcileInterval time.Duration,
) (*edgeSyncConfigController, error) {
	rateLimitter := workqueue.NewItemFastSlowRateLimiter(reconcileInterval, reconcileInterval*5, 1000)
	c := &edgeSyncConfigController{
		syncConfigLister:  syncConfigInformer.Lister(),
		syncConfigClient:  syncConfigClient,
		syncConfigManager: syncConfigManager,
		upSyncer:          upSyncer,
		downSyncer:        downSyncer,
	}
	controllerName := "kcp-edge-syncconfig-controller"
	controllerBase := &controllerBase{
		name:              controllerName,
		target:            "SyncConfig",
		queue:             workqueue.NewNamedRateLimitingQueue(rateLimitter, controllerName),
		informerHasSynced: syncConfigInformer.Informer().HasSynced,
		process:           c.process,
	}
	c.controllerBase = controllerBase
	syncConfigInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
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

// edgeSyncConfigController is a control loop that watches EdgeSyncConfig.
type edgeSyncConfigController struct {
	*controllerBase
	syncConfigLister  edgev1alpha1listers.EdgeSyncConfigLister
	syncConfigClient  edgev1alpha1typed.EdgeSyncConfigInterface
	syncConfigManager *SyncConfigManager
	upSyncer          syncers.SyncerInterface
	downSyncer        syncers.SyncerInterface
}

func (c *edgeSyncConfigController) process(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Error(err, "failed to split key, dropping")
		return nil
	}

	refresh := func() error {
		downsyncerError := c.downSyncer.ReInitializeClients(c.syncConfigManager.GetDownSyncedResources(), c.syncConfigManager.GetConversions())
		upSyncerError := c.upSyncer.ReInitializeClients(c.syncConfigManager.GetUpSyncedResources(), c.syncConfigManager.GetConversions())

		if downsyncerError != nil && upSyncerError != nil {
			return fmt.Errorf("failed to reinitialize downsyncer (%w) and upsyncer (%w)", downsyncerError, upSyncerError)
		} else if downsyncerError != nil && upSyncerError == nil {
			return fmt.Errorf("failed to reinitialize downsyncer (%w)", downsyncerError)
		} else if downsyncerError == nil && upSyncerError != nil {
			return fmt.Errorf("failed to reinitialize upsyncer (%w)", upSyncerError)
		}
		return nil
	}

	syncConfig, err := c.syncConfigLister.Get(name)
	if err != nil {
		if errors.IsNotFound(err) { // object is deleted
			c.syncConfigManager.delete(name)
			return refresh()
		}
		return err
	}

	c.syncConfigManager.upsert(*syncConfig)

	return refresh()
}
