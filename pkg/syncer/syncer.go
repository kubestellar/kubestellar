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

package syncer

import (
	"context"
	"fmt"
	"time"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"
	kcpclientset "github.com/kcp-dev/edge-mc/pkg/syncer/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/edge-mc/pkg/syncer/client/informers/externalversions"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/syncer/client/listers/edge/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/kcp-dev/edge-mc/pkg/syncer/syncers"
	"k8s.io/apimachinery/pkg/fields"

	"github.com/kcp-dev/edge-mc/pkg/syncer/controller"
	"github.com/kcp-dev/logicalcluster/v3"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/pkg/version"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

type SyncerConfig struct {
	UpstreamConfig   *rest.Config
	DownstreamConfig *rest.Config
	SyncTargetPath   logicalcluster.Path
	SyncTargetName   string
	SyncTargetUID    string
}

const (
	resyncPeriod = 10 * time.Hour
)

func StartSyncer(ctx context.Context, cfg *SyncerConfig, numSyncerThreads int) error {
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("syncTarget.name", cfg.SyncTargetName)
	logger.V(2).Info("starting edge-mc syncer")
	kcpVersion := version.Get().GitVersion

	bootstrapConfig := rest.CopyConfig(cfg.UpstreamConfig)
	rest.AddUserAgent(bootstrapConfig, "edge-mc#syncer/"+kcpVersion)

	syncConfigClientSet, err := kcpclientset.NewForConfig(bootstrapConfig)
	if err != nil {
		return err
	}
	syncConfigClient := syncConfigClientSet.EdgeV1alpha1().EdgeSyncConfigs()
	// syncConfigInformerFactory to watch a certain syncConfig on upstream
	syncConfigInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(syncConfigClientSet, resyncPeriod, kcpinformers.WithTweakListOptions(
		func(listOptions *metav1.ListOptions) {
			listOptions.FieldSelector = fields.OneTermEqualSelector("metadata.name", cfg.SyncTargetName).String()
		},
	))
	syncConfigInformer := syncConfigInformerFactory.Edge().V1alpha1().EdgeSyncConfigs()

	syncConfigInformer.Lister().List(labels.Everything()) // TODO: Remove (for now, need to invoke List at once)

	syncConfigInformerFactory.Start(ctx.Done())
	syncConfigInformerFactory.WaitForCacheSync(ctx.Done())

	upstreamConfig := rest.CopyConfig(cfg.UpstreamConfig)
	rest.AddUserAgent(upstreamConfig, "edge-mc#syncer/"+kcpVersion)
	upstreamDyClient, err := dynamic.NewForConfig(upstreamConfig)
	if err != nil {
		return err
	}
	upstreamDiscoveryClient := discovery.NewDiscoveryClientForConfigOrDie(upstreamConfig)
	upstreamClientFactory, err := syncers.NewClientFactory(logger, upstreamDyClient, upstreamDiscoveryClient)
	if err != nil {
		return err
	}

	downstreamConfig := rest.CopyConfig(cfg.DownstreamConfig)
	rest.AddUserAgent(downstreamConfig, "edge-mc#syncer/"+kcpVersion)
	downstreamDyClient, err := dynamic.NewForConfig(downstreamConfig)
	if err != nil {
		return err
	}
	downstreamDiscoveryClient := discovery.NewDiscoveryClientForConfigOrDie(downstreamConfig)
	downstreamClientFactory, err := syncers.NewClientFactory(logger, downstreamDyClient, downstreamDiscoveryClient)
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

	controller, err := controller.NewSyncConfigController(logger, syncConfigClient, syncConfigInformer, cfg.SyncTargetName, upSyncer, downSyncer, 5*time.Second)
	if err != nil {
		return err
	}

	go controller.Run(ctx, numSyncerThreads)
	startSync(ctx, logger, cfg, syncConfigInformer.Lister(), upSyncer, downSyncer)
	return nil
}

func startSync(ctx context.Context, logger klog.Logger, cfg *SyncerConfig, syncConfigLister edgev1alpha1listers.EdgeSyncConfigLister, upSyncer *syncers.UpSyncer, downSyncer *syncers.DownSyncer) {
	logger.V(2).Info("Start sync")
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(time.Second * 15):
			logger.V(2).Info("Sync ")
			syncConfig, err := syncConfigLister.Get(cfg.SyncTargetName)
			if err != nil {
				logger.Error(err, "failed to get syncConfig")
			} else {
				for _, resource := range syncConfig.Spec.DownSyncedResources {
					if err := downSyncer.SyncOne(resource, syncConfig.Spec.Conversions); err != nil {
						logger.V(1).Info(fmt.Sprintf("failed to downsync %s.%s/%s (ns=%s)", resource.Kind, resource.Group, resource.Name, resource.Namespace))
					}
					if err := downSyncer.BackStatusOne(resource, syncConfig.Spec.Conversions); err != nil {
						logger.V(1).Info(fmt.Sprintf("failed to status upsync %s.%s/%s (ns=%s)", resource.Kind, resource.Group, resource.Name, resource.Namespace))
					}
				}
				for _, resource := range syncConfig.Spec.UpSyncedResources {
					if err := upSyncer.SyncOne(resource, syncConfig.Spec.Conversions); err != nil {
						logger.V(1).Info(fmt.Sprintf("failed to upsync %s.%s/%s (ns=%s)", resource.Kind, resource.Group, resource.Name, resource.Namespace))
					}
				}
			}
		}
	}
}
