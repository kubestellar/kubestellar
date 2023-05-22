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

package placement

import (
	"context"
	"os"

	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	clusterdiscovery "github.com/kcp-dev/client-go/discovery"
	clusterdynamic "github.com/kcp-dev/client-go/dynamic"
	kcpkubeinformers "github.com/kcp-dev/client-go/informers"
	kcpkubecorev1informers "github.com/kcp-dev/client-go/informers/core/v1"
	kcpkubecorev1client "github.com/kcp-dev/client-go/kubernetes/typed/core/v1"
	kcpclusterclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	tenancyv1a1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
	tenancyv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgeclusterclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgev1a1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1a1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
)

type placementTranslator struct {
	context                context.Context
	apiProvider            APIWatchMapProvider
	spsClusterInformer     kcpcache.ScopeableSharedIndexInformer
	syncfgClusterInformer  kcpcache.ScopeableSharedIndexInformer
	syncfgClusterLister    edgev1a1listers.SyncerConfigClusterLister
	mbwsInformer           k8scache.SharedIndexInformer
	mbwsLister             tenancyv1a1listers.WorkspaceLister
	kcpClusterClientset    kcpclusterclientset.ClusterInterface
	discoveryClusterClient clusterdiscovery.DiscoveryClusterInterface
	crdClusterInformer     kcpcache.ScopeableSharedIndexInformer
	bindingClusterInformer kcpcache.ScopeableSharedIndexInformer
	dynamicClusterClient   clusterdynamic.ClusterInterface
	edgeClusterClientset   edgeclusterclientset.ClusterInterface
	nsClusterPreInformer   kcpkubecorev1informers.NamespaceClusterInformer
	nsClusterClient        kcpkubecorev1client.NamespaceClusterInterface

	workloadProjector interface {
		WorkloadProjector
		Runnable
	}

	whatResolver  WhatResolver
	whereResolver WhereResolver
}

func NewPlacementTranslator(
	numThreads int,
	ctx context.Context,
	// pre-informer on all SinglePlacementSlice objects, cross-workspace
	epClusterPreInformer edgev1a1informers.EdgePlacementClusterInformer,
	// pre-informer on all SinglePlacementSlice objects, cross-workspace
	spsClusterPreInformer edgev1a1informers.SinglePlacementSliceClusterInformer,
	// pre-informer on syncer config objects, should be in mailbox workspaces
	syncfgClusterPreInformer edgev1a1informers.SyncerConfigClusterInformer,
	// pre-informer on Workspaces objects in the ESPW
	mbwsPreInformer tenancyv1a1informers.WorkspaceInformer,
	// all-cluster clientset for kcp APIs,
	// needed to manipulate TMC objects in mailbox workspaces
	kcpClusterClientset kcpclusterclientset.ClusterInterface,
	// needed for enumerating resources in workload mgmt workspaces
	discoveryClusterClient clusterdiscovery.DiscoveryClusterInterface,
	// needed to watch for new resources appearing
	crdClusterPreInformer kcpkubeinformers.GenericClusterInformer,
	// needed to watch for new resources appearing
	bindingClusterPreInformer kcpkubeinformers.GenericClusterInformer,
	// needed to read and write arbitrary objects
	dynamicClusterClient clusterdynamic.ClusterInterface,
	// to read and write syncer config objects
	edgeClusterClientset edgeclusterclientset.ClusterInterface,
	// for monitoring namespaces in mailbox workspaces
	nsClusterPreInformer kcpkubecorev1informers.NamespaceClusterInformer,
	// for creating namespaces in mailbox workspaces
	nsClusterClient kcpkubecorev1client.NamespaceClusterInterface,
) *placementTranslator {
	amp := NewAPIWatchMapProvider(ctx, numThreads, discoveryClusterClient, crdClusterPreInformer, bindingClusterPreInformer)
	mbwsPreInformer.Lister()
	pt := &placementTranslator{
		context:                ctx,
		apiProvider:            amp,
		spsClusterInformer:     spsClusterPreInformer.Informer(),
		syncfgClusterInformer:  syncfgClusterPreInformer.Informer(),
		syncfgClusterLister:    syncfgClusterPreInformer.Lister(),
		mbwsInformer:           mbwsPreInformer.Informer(),
		mbwsLister:             mbwsPreInformer.Lister(),
		kcpClusterClientset:    kcpClusterClientset,
		discoveryClusterClient: discoveryClusterClient,
		crdClusterInformer:     crdClusterPreInformer.Informer(),
		bindingClusterInformer: bindingClusterPreInformer.Informer(),
		dynamicClusterClient:   dynamicClusterClient,
		edgeClusterClientset:   edgeClusterClientset,
		nsClusterPreInformer:   nsClusterPreInformer,
		nsClusterClient:        nsClusterClient,
		whatResolver: NewWhatResolver(ctx, epClusterPreInformer, discoveryClusterClient,
			crdClusterPreInformer, bindingClusterPreInformer, dynamicClusterClient, numThreads),
		whereResolver: NewWhereResolver(ctx, spsClusterPreInformer, numThreads),
	}
	pt.workloadProjector = NewWorkloadProjector(ctx, numThreads, DefaultResourceModes, pt.mbwsInformer, pt.mbwsLister,
		pt.syncfgClusterInformer, pt.syncfgClusterLister, edgeClusterClientset, dynamicClusterClient,
		nsClusterPreInformer, nsClusterClient)

	return pt
}

func (pt *placementTranslator) Run() {
	ctx := pt.context
	logger := klog.FromContext(ctx)

	doneCh := ctx.Done()
	if !k8scache.WaitForNamedCacheSync("placement-translator", doneCh,
		pt.spsClusterInformer.HasSynced, pt.mbwsInformer.HasSynced,
		pt.crdClusterInformer.HasSynced, pt.bindingClusterInformer.HasSynced,
		pt.syncfgClusterInformer.HasSynced,
	) {
		logger.Error(nil, "Informer syncs not achieved")
		os.Exit(100)
	}

	whatResolver := func(mr MappingReceiver[ExternalName, ResolvedWhat]) Runnable {
		fork := MappingReceiverFork[ExternalName, ResolvedWhat]{NewLoggingMappingReceiver[ExternalName, ResolvedWhat]("what", logger), mr}
		return pt.whatResolver(fork)
	}
	whereResolver := func(mr MappingReceiver[ExternalName, ResolvedWhere]) Runnable {
		fork := MappingReceiverFork[ExternalName, ResolvedWhere]{NewLoggingMappingReceiver[ExternalName, ResolvedWhere]("where", logger), mr}
		return pt.whereResolver(fork)
	}
	setBinder := NewSetBinder(logger, NewWorkloadPartsDifferencer, NewUpsyncDifferencer, NewResolvedWhereDifferencer,
		SimpleBindingOrganizer(logger),
		pt.apiProvider,
		DefaultResourceModes, // TODO: replace with configurable
		nil,                  // TODO: get this right
	)
	// workloadProjector := NewLoggingWorkloadProjector(logger)
	runner := AssemplePlacementTranslator(whatResolver, whereResolver, setBinder, pt.workloadProjector)
	// TODO: move all that stuff up before Run
	go pt.apiProvider.Run(ctx)       // TODO: also wait for this to finish
	go pt.workloadProjector.Run(ctx) // TODO: also wait for this to finish
	runner.Run(ctx)
}

func NewUpsyncDifferencer(eltReceiver SetChangeReceiver[edgeapi.UpsyncSet]) Receiver[ /*immutable*/ []edgeapi.UpsyncSet] {
	return NewSliceDifferencerParametric(UpsyncSetEqual, eltReceiver, nil)
}
