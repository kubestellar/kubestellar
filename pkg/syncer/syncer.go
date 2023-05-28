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

package syncer

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/version"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgeclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
	edgeinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
	"github.com/kcp-dev/edge-mc/pkg/syncer/clientfactory"
	"github.com/kcp-dev/edge-mc/pkg/syncer/controller"
	"github.com/kcp-dev/edge-mc/pkg/syncer/syncers"
)

type SyncerConfig struct {
	UpstreamConfig   *rest.Config
	DownstreamConfig *rest.Config
	SyncTargetPath   logicalcluster.Path
	SyncTargetName   string
	SyncTargetUID    string
	Interval         time.Duration
}

const (
	resyncPeriod    = 10 * time.Hour
	defaultInterval = time.Second * 15
	minimumInterval = time.Second * 1
)

func RunSyncer(ctx context.Context, cfg *SyncerConfig, numSyncerThreads int) error {
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("syncTargetName", cfg.SyncTargetName)
	logger.V(2).Info("starting edge-mc syncer")
	kcpVersion := version.Get().GitVersion

	bootstrapConfig := rest.CopyConfig(cfg.UpstreamConfig)
	rest.AddUserAgent(bootstrapConfig, "edge-mc#syncer/"+kcpVersion)

	// For edgeSyncConfig
	syncConfigClientSet, err := edgeclientset.NewForConfig(bootstrapConfig)
	if err != nil {
		return err
	}
	syncConfigClient := syncConfigClientSet.EdgeV1alpha1().EdgeSyncConfigs()
	// syncConfigInformerFactory to watch a certain syncConfig on upstream
	syncConfigInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(syncConfigClientSet, resyncPeriod)
	syncConfigAccess := syncConfigInformerFactory.Edge().V1alpha1().EdgeSyncConfigs()

	syncConfigAccess.Lister().List(labels.Everything()) // TODO: Remove (for now, need to invoke List at once)

	syncConfigInformerFactory.Start(ctx.Done())
	syncConfigInformerFactory.WaitForCacheSync(ctx.Done())

	// For syncerConfig
	syncerConfigClientSet, err := edgeclientset.NewForConfig(bootstrapConfig)
	if err != nil {
		return err
	}
	syncerConfigClient := syncerConfigClientSet.EdgeV1alpha1().SyncerConfigs()
	// syncerConfigInformerFactory to watch a certain syncConfig on upstream
	syncerConfigInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(syncerConfigClientSet, resyncPeriod)
	syncerConfigAccess := syncerConfigInformerFactory.Edge().V1alpha1().SyncerConfigs()

	syncerConfigAccess.Lister().List(labels.Everything()) // TODO: Remove (for now, need to invoke List at once)

	syncerConfigInformerFactory.Start(ctx.Done())
	syncerConfigInformerFactory.WaitForCacheSync(ctx.Done())

	upstreamConfig := rest.CopyConfig(cfg.UpstreamConfig)
	rest.AddUserAgent(upstreamConfig, "edge-mc#syncer/"+kcpVersion)
	upstreamDynamicClient, err := dynamic.NewForConfig(upstreamConfig)
	if err != nil {
		return err
	}
	upstreamDiscoveryClient := discovery.NewDiscoveryClientForConfigOrDie(upstreamConfig)
	upstreamClientFactory, err := clientfactory.NewClientFactory(logger, upstreamDynamicClient, upstreamDiscoveryClient)
	if err != nil {
		return err
	}

	downstreamConfig := rest.CopyConfig(cfg.DownstreamConfig)
	rest.AddUserAgent(downstreamConfig, "edge-mc#syncer/"+kcpVersion)
	downstreamDynamicClient, err := dynamic.NewForConfig(downstreamConfig)
	if err != nil {
		return err
	}
	downstreamDiscoveryClient := discovery.NewDiscoveryClientForConfigOrDie(downstreamConfig)
	downstreamClientFactory, err := clientfactory.NewClientFactory(logger, downstreamDynamicClient, downstreamDiscoveryClient)
	if err != nil {
		return err
	}

	upSyncer, err := syncers.NewUpSyncer(logger, upstreamClientFactory, downstreamClientFactory, []edgev1alpha1.EdgeSyncConfigResource{}, []edgev1alpha1.EdgeSynConversion{})
	if err != nil {
		return err
	}
	downSyncer, err := syncers.NewDownSyncer(logger, upstreamClientFactory, downstreamClientFactory, []edgev1alpha1.EdgeSyncConfigResource{}, []edgev1alpha1.EdgeSynConversion{})
	if err != nil {
		return err
	}

	syncConfigManager := controller.NewSyncConfigManager(logger)
	syncConfigController, err := controller.NewEdgeSyncConfigController(logger, syncConfigClient, syncConfigAccess, syncConfigManager, upSyncer, downSyncer, 5*time.Second)
	if err != nil {
		return err
	}

	syncerConfigManager := controller.NewSyncerConfigManager(logger, syncConfigManager, upstreamClientFactory, downstreamClientFactory)
	syncerConfigController, err := controller.NewSyncerConfigController(logger, syncerConfigClient, syncerConfigAccess, syncerConfigManager, 5*time.Second)
	if err != nil {
		return err
	}

	go syncConfigController.Run(ctx, numSyncerThreads)
	go syncerConfigController.Run(ctx, numSyncerThreads)
	runSync(ctx, cfg, syncConfigManager, syncerConfigManager, upSyncer, downSyncer)
	return nil
}

func runSync(ctx context.Context, cfg *SyncerConfig, syncConfigManager *controller.SyncConfigManager, syncerConfigManager *controller.SyncerConfigManager, upSyncer *syncers.UpSyncer, downSyncer *syncers.DownSyncer) {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Start sync")
	interval := cfg.Interval
	if interval < minimumInterval {
		interval = defaultInterval
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(interval):
			logger.V(2).Info(fmt.Sprintf("Sync with interval: %v", interval))
			syncerConfigManager.Refresh()
			downSyncedResources := syncConfigManager.GetDownSyncedResources()
			downUnsyncedResources := syncConfigManager.GetDownUnsyncedResources()
			upSyncedReousrces := syncConfigManager.GetUpSyncedResources()
			upUnsyncedReousrces := syncConfigManager.GetUpUnsyncedResources()
			conversions := syncConfigManager.GetConversions()
			_ = downSyncer.ReInitializeClients(downSyncedResources, conversions)
			_ = upSyncer.ReInitializeClients(upSyncedReousrces, conversions)
			sync(logger.WithValues("actor", "DownSyncer:Sync"), downSyncer, downSyncedResources, conversions)
			sync(logger.WithValues("actor", "DownSyncer:Unsync"), downSyncer, downUnsyncedResources, conversions)
			syncStatus(logger.WithValues("actor", "DownSyncer:Sync"), downSyncer, downSyncedResources, conversions)
			sync(logger.WithValues("actor", "UpSyncer:Sync"), upSyncer, upSyncedReousrces, conversions)
			sync(logger.WithValues("actor", "UpSyncer:Unsync"), upSyncer, upUnsyncedReousrces, conversions)
		}
	}
}

func sync(logger klog.Logger, syncer syncers.SyncerInterface, resources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) {
	for _, resource := range resources {
		if resource.Name == "*" || resource.Namespace == "*" {
			if err := syncer.SyncMany(resource, conversions); err != nil {
				logger.V(1).Info(fmt.Sprintf("failed to sync-many %s.%s/%s (ns=%s)", resource.Kind, resource.Group, resource.Name, resource.Namespace))
			}
		} else {
			if err := syncer.SyncOne(resource, conversions); err != nil {
				logger.V(1).Info(fmt.Sprintf("failed to sync %s.%s/%s (ns=%s)", resource.Kind, resource.Group, resource.Name, resource.Namespace))
			}
		}
	}
}

func syncStatus(logger klog.Logger, downSyncer *syncers.DownSyncer, resources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) {
	for _, resource := range resources {
		if resource.Name == "*" || resource.Namespace == "*" {
			if err := downSyncer.BackStatusMany(resource, conversions); err != nil {
				logger.V(1).Info(fmt.Sprintf("failed to status sync-many %s.%s/%s (ns=%s)", resource.Kind, resource.Group, resource.Name, resource.Namespace))
			}
		} else {
			if err := downSyncer.BackStatusOne(resource, conversions); err != nil {
				logger.V(1).Info(fmt.Sprintf("failed to status sync %s.%s/%s (ns=%s)", resource.Kind, resource.Group, resource.Name, resource.Namespace))
			}
		}
	}
}
