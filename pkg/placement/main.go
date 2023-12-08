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

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgev1a1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev1a1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
	spacev1alpha1 "github.com/kubestellar/kubestellar/space-framework/pkg/client/informers/externalversions/space/v1alpha1"
	spacev1a1listers "github.com/kubestellar/kubestellar/space-framework/pkg/client/listers/space/v1alpha1"
	msclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
)

type placementTranslator struct {
	context        context.Context
	apiProvider    APIWatchMapProvider
	spsInformer    k8scache.SharedIndexInformer
	syncfgInformer k8scache.SharedIndexInformer
	syncfgLister   edgev1a1listers.SyncerConfigLister
	spaceInformer  k8scache.SharedIndexInformer
	spaceLister    spacev1a1listers.SpaceLister

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
	// pre-informer on SinglePlacementSlice objects
	epPreInformer edgev1a1informers.EdgePlacementInformer,
	// pre-informer on SinglePlacementSlice objects
	spsPreInformer edgev1a1informers.SinglePlacementSliceInformer,
	// pre-informer on syncer config objects, should be in mailbox spaces
	syncfgPreInformer edgev1a1informers.SyncerConfigInformer,

	spaceclient msclient.KubestellarSpaceInterface,
	spaceProviderNs string,
	spacePreInformer spacev1alpha1.SpaceInformer,
	kbSpaceRelation kbuser.KubeBindSpaceRelation,
) *placementTranslator {
	amp := NewAPIWatchMapProvider(ctx, numThreads, spaceclient, spaceProviderNs)
	pt := &placementTranslator{
		context:        ctx,
		apiProvider:    amp,
		spsInformer:    spsPreInformer.Informer(),
		syncfgInformer: syncfgPreInformer.Informer(),
		syncfgLister:   syncfgPreInformer.Lister(),
		spaceInformer:  spacePreInformer.Informer(),
		spaceLister:    spacePreInformer.Lister(),

		whatResolver:  NewWhatResolver(ctx, epPreInformer, spaceclient, spaceProviderNs, kbSpaceRelation, numThreads),
		whereResolver: NewWhereResolver(ctx, spsPreInformer, kbSpaceRelation, numThreads),
	}
	pt.workloadProjector = NewWorkloadProjector(ctx, numThreads, DefaultResourceModes,
		pt.spaceInformer, pt.spaceLister, pt.syncfgInformer,
		spaceclient, spaceProviderNs, kbSpaceRelation)

	return pt
}

func (pt *placementTranslator) Run() {
	ctx := pt.context
	logger := klog.FromContext(ctx)

	doneCh := ctx.Done()
	if !(k8scache.WaitForNamedCacheSync("placement-translator(sps)", doneCh, pt.spsInformer.HasSynced) &&
		k8scache.WaitForNamedCacheSync("placement-translator(space)", doneCh, pt.spaceInformer.HasSynced) &&
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
