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

	apiextinformers "k8s.io/apiextensions-apiserver/pkg/client/kcp/informers/externalversions/apiextensions/v1"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	clusterdiscovery "github.com/kcp-dev/client-go/discovery"
	clusterdynamic "github.com/kcp-dev/client-go/dynamic"
	kcpkubecorev1informers "github.com/kcp-dev/client-go/informers/core/v1"
	kcpkubecorev1client "github.com/kcp-dev/client-go/kubernetes/typed/core/v1"
	kcpclusterclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	bindinginformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/apis/v1alpha1"
	tenancyv1a1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
	tenancyv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgeclusterclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
	edgev1a1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev1a1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
)

type placementTranslator struct {
	context                context.Context
	apiProvider            APIWatchMapProvider
	spsInformer            k8scache.SharedIndexInformer
	syncfgInformer         k8scache.SharedIndexInformer
	syncfgLister           edgev1a1listers.SyncerConfigLister
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
	locationPreInformer edgev1a1informers.LocationInformer,
	// pre-informer on all SinglePlacementSlice objects, cross-workspace
	epPreInformer edgev1a1informers.EdgePlacementInformer,
	// pre-informer on all SinglePlacementSlice objects, cross-workspace
	spsPreInformer edgev1a1informers.SinglePlacementSliceInformer,
	// pre-informer on syncer config objects, should be in mailbox workspaces
	syncfgPreInformer edgev1a1informers.SyncerConfigInformer,

	customizerPreInformer edgev1a1informers.CustomizerInformer,
	// pre-informer on Workspaces objects
	mbwsPreInformer tenancyv1a1informers.WorkspaceInformer,
	// all-cluster clientset for kcp APIs,
	// needed to manipulate TMC objects in mailbox workspaces
	kcpClusterClientset kcpclusterclientset.ClusterInterface,
	// needed for enumerating resources in workload mgmt workspaces
	discoveryClusterClient clusterdiscovery.DiscoveryClusterInterface,
	// needed to watch for new resources appearing
	crdClusterPreInformer apiextinformers.CustomResourceDefinitionClusterInformer,
	// needed to watch for new resources appearing
	bindingClusterPreInformer bindinginformers.APIBindingClusterInformer,
	// needed to read and write arbitrary objects
	dynamicClusterClient clusterdynamic.ClusterInterface,
	// to read and write syncer config objects
	edgeClusterClientset edgeclusterclientset.ClusterInterface,
	// for monitoring namespaces in mailbox workspaces
	nsClusterPreInformer kcpkubecorev1informers.NamespaceClusterInformer,
	// for creating namespaces in mailbox workspaces
	nsClusterClient kcpkubecorev1client.NamespaceClusterInterface,
	kbSpaceRelation kbuser.KubeBindSpaceRelation,
) *placementTranslator {
	amp := NewAPIWatchMapProvider(ctx, numThreads, discoveryClusterClient, crdClusterPreInformer, bindingClusterPreInformer)
	mbwsPreInformer.Lister()
	pt := &placementTranslator{
		context:                ctx,
		apiProvider:            amp,
		spsInformer:            spsPreInformer.Informer(),
		syncfgInformer:         syncfgPreInformer.Informer(),
		syncfgLister:           syncfgPreInformer.Lister(),
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
		whatResolver: NewWhatResolver(ctx, epPreInformer, discoveryClusterClient,
			crdClusterPreInformer, bindingClusterPreInformer, dynamicClusterClient, kbSpaceRelation, numThreads),
		whereResolver: NewWhereResolver(ctx, spsPreInformer, kbSpaceRelation, numThreads),
	}
	pt.workloadProjector = NewWorkloadProjector(ctx, numThreads, DefaultResourceModes, pt.mbwsInformer, pt.mbwsLister,
		pt.syncfgInformer,
		edgeClusterClientset, dynamicClusterClient,
		nsClusterPreInformer, nsClusterClient)

	return pt
}

func (pt *placementTranslator) Run() {
	ctx := pt.context
	logger := klog.FromContext(ctx)

	doneCh := ctx.Done()
	if !(k8scache.WaitForNamedCacheSync("placement-translator(sps)", doneCh, pt.spsInformer.HasSynced) &&
		k8scache.WaitForNamedCacheSync("placement-translator(mbws)", doneCh, pt.mbwsInformer.HasSynced) &&
		k8scache.WaitForNamedCacheSync("placement-translator(crds)", doneCh, pt.crdClusterInformer.HasSynced) &&
		k8scache.WaitForNamedCacheSync("placement-translator(bind)", doneCh, pt.bindingClusterInformer.HasSynced) &&
		k8scache.WaitForNamedCacheSync("placement-translator(sync)", doneCh, pt.syncfgInformer.HasSynced) &&
		true) {
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
