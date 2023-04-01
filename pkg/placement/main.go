/*
Copyright 2023 The KCP Authors.

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

package placement

import (
	"context"
	"os"

	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	clusterdiscovery "github.com/kcp-dev/client-go/discovery"
	clusterdynamic "github.com/kcp-dev/client-go/dynamic"
	kcpinformers "github.com/kcp-dev/client-go/informers"
	kcpclusterclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	tenancyv1a1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev1a1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
)

type placementTranslator struct {
	context                context.Context
	logger                 klog.Logger
	spsClusterInformer     kcpcache.ScopeableSharedIndexInformer
	mbwsInformer           k8scache.SharedIndexInformer
	kcpClusterClientset    kcpclusterclientset.ClusterInterface
	discoveryClusterClient clusterdiscovery.DiscoveryClusterInterface
	crdClusterInformer     kcpcache.ScopeableSharedIndexInformer
	bindingClusterInformer kcpcache.ScopeableSharedIndexInformer
	dynamicClusterClient   clusterdynamic.ClusterInterface

	whatResolverLauncher  func() *whatResolver
	whereResolverLauncher func() *whereResolver
}

func NewPlacementTranslator(
	numThreads int,
	ctx context.Context,
	// pre-informer on all SinglePlacementSlice objects, cross-workspace
	epClusterPreInformer edgev1a1informers.EdgePlacementClusterInformer,
	// pre-informer on all SinglePlacementSlice objects, cross-workspace
	spsClusterPreInformer edgev1a1informers.SinglePlacementSliceClusterInformer,
	// pre-informer on Workspaces objects in the ESPW
	mbwsPreInformer tenancyv1a1informers.WorkspaceInformer,
	// all-cluster clientset for kcp APIs,
	// needed to manipulate TMC objects in mailbox workspaces
	kcpClusterClientset kcpclusterclientset.ClusterInterface,
	// needed for enumerating resources in workload mgmt workspaces
	discoveryClusterClient clusterdiscovery.DiscoveryClusterInterface,
	// needed to watch for new resources appearing
	crdClusterPreInformer kcpinformers.GenericClusterInformer,
	// needed to watch for new resources appearing
	bindingClusterPreInformer kcpinformers.GenericClusterInformer,
	// needed to read and write arbitrary objects
	dynamicClusterClient clusterdynamic.ClusterInterface,
	// TODO: add the other needed arguments
) *placementTranslator {
	pt := &placementTranslator{
		context:                ctx,
		logger:                 klog.FromContext(ctx),
		spsClusterInformer:     spsClusterPreInformer.Informer(),
		mbwsInformer:           mbwsPreInformer.Informer(),
		kcpClusterClientset:    kcpClusterClientset,
		discoveryClusterClient: discoveryClusterClient,
		crdClusterInformer:     crdClusterPreInformer.Informer(),
		bindingClusterInformer: bindingClusterPreInformer.Informer(),
		dynamicClusterClient:   dynamicClusterClient,
		whatResolverLauncher: NewWhatResolverLauncher(ctx, epClusterPreInformer, discoveryClusterClient,
			crdClusterPreInformer, bindingClusterPreInformer, dynamicClusterClient, numThreads),
		whereResolverLauncher: NewWhereResolverLauncher(ctx, spsClusterPreInformer, numThreads),
	}
	return pt
}

func (pt *placementTranslator) Run() {
	ctx := pt.context
	doneCh := ctx.Done()
	if !k8scache.WaitForNamedCacheSync("placement-translator", doneCh,
		pt.spsClusterInformer.HasSynced, pt.mbwsInformer.HasSynced,
		pt.crdClusterInformer.HasSynced, pt.crdClusterInformer.HasSynced,
	) {
		pt.logger.Error(nil, "Informer syncs not achieved")
		os.Exit(100)
	}

	// TODO: make all the other needed infrastructure

	// TODO: replace all these dummies
	whatResolver := pt.whatResolverLauncher()
	whatResolver.AddReceiver(LoggingMappingReceiver[ExternalName, WorkloadParts]{pt.logger}, true) // debugging
	whereResolver := pt.whereResolverLauncher()
	whereResolver.AddReceiver(LoggingMappingReceiver[ExternalName, ResolvedWhere]{pt.logger}, true) // debugging
	dummyBaseAPIProvider := NewRelayMap[logicalcluster.Name, ScopedAPIProvider](false)
	setBinder := NewSetBinder(pt.logger, NewResolvedWhatDifferencer, NewResolvedWhereDifferencer,
		SimpleBindingOrganizer(pt.logger),
		NewTestAPIMapProvider(dummyBaseAPIProvider),
		DefaultResourceModes,
		nil)
	workloadProjector := NewLoggingWorkloadProjector(pt.logger)
	placementProjector := NewDummyWorkloadProjector()
	AssemplePlacementTranslator(whatResolver, whereResolver, setBinder, workloadProjector, placementProjector)
	<-doneCh
}

type LoggingMappingReceiver[Key comparable, Val any] struct {
	logger klog.Logger
}

var _ MappingReceiver[string, []any] = &LoggingMappingReceiver[string, []any]{}

func (lmr LoggingMappingReceiver[Key, Val]) Put(key Key, val Val) {
	lmr.logger.Info("Put", "key", key, "val", val)
}

func (lmr LoggingMappingReceiver[Key, Val]) Delete(key Key) {
	lmr.logger.Info("Delete", "key", key)
}
