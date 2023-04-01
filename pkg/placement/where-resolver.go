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
	"time"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
)

type whereResolver struct {
	ctx    context.Context
	logger klog.Logger
	queue  workqueue.RateLimitingInterface

	spsInformer kcpcache.ScopeableSharedIndexInformer
	spsLister   edgev1alpha1listers.SinglePlacementSliceClusterLister

	// resolutions maps EdgePlacement name to its ResolvedWhere
	resolutions RelayMap[ExternalName, ResolvedWhere]
}

var _ WhereResolver = &whereResolver{}

// NewWhereResolverLauncher returns a function that returns a WhereResolver.
func NewWhereResolverLauncher(
	ctx context.Context,
	spsPreInformer edgev1alpha1informers.SinglePlacementSliceClusterInformer,
	numThreads int,
) func() *whereResolver {
	controllerName := "where-resolver"
	logger := klog.FromContext(ctx).WithValues("part", controllerName)
	ctx = klog.NewContext(ctx, logger)
	wr := &whereResolver{
		ctx:         ctx,
		logger:      logger,
		queue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		spsInformer: spsPreInformer.Informer(),
		spsLister:   spsPreInformer.Lister(),
		resolutions: NewRelayMap[ExternalName, ResolvedWhere](true),
	}
	wr.spsInformer.AddEventHandler(WhereResolverClusterHandler{wr, mkgk(edgeapi.SchemeGroupVersion.Group, "SinglePlacementSlice")})
	return func() *whereResolver {
		if upstreamcache.WaitForNamedCacheSync(controllerName, ctx.Done(), wr.spsInformer.HasSynced) {
			for i := 0; i < numThreads; i++ {
				go wait.Until(wr.runWorker, time.Second, ctx.Done())
			}
		} else {
			logger.Info("Failed to sync SinglePlacementSlices in time")
		}
		return wr
	}
}

type WhereResolverClusterHandler struct {
	*whereResolver
	gk schema.GroupKind
}

func (wrh WhereResolverClusterHandler) OnAdd(obj any) {
	wrh.enqueue(wrh.gk, obj)
}

func (wrh WhereResolverClusterHandler) OnUpdate(oldObj, newObj any) {
	wrh.enqueue(wrh.gk, newObj)
}

func (wrh WhereResolverClusterHandler) OnDelete(obj any) {
	wrh.enqueue(wrh.gk, obj)
}

func (wr *whereResolver) enqueue(gk schema.GroupKind, objAny any) {
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(objAny)
	if err != nil {
		wr.logger.Error(err, "Failed to extract object reference", "object", objAny)
		return
	}
	cluster, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(key)
	if err != nil {
		wr.logger.Error(err, "Impossible! SplitMetaClusterNamespaceKey failed", "key", key)
	}
	item := queueItem{gk: gk, cluster: cluster, name: name}
	wr.logger.V(4).Info("Enqueuing", "item", item)
	wr.queue.Add(item)
}

func (wr *whereResolver) AddReceiver(receiver MappingReceiver[ExternalName, ResolvedWhere], notifyCurrent bool) {
	wr.resolutions.AddReceiver(receiver, notifyCurrent)
}

func (wr *whereResolver) Get(placement ExternalName, kont func(ResolvedWhere)) {
	wr.resolutions.Get(placement, kont)
}

func (wr *whereResolver) runWorker() {
	for wr.processNextWorkItem() {
	}
}

func (wr *whereResolver) processNextWorkItem() bool {
	// Wait until there is a new item in the working queue
	itemAny, quit := wr.queue.Get()
	if quit {
		return false
	}
	defer wr.queue.Done(itemAny)
	item := itemAny.(queueItem)

	logger := klog.FromContext(wr.ctx).WithValues("group", item.gk.Group, "kind", item.gk.Kind, "cluster", item.cluster, "name", item.name)
	ctx := klog.NewContext(wr.ctx, logger)
	logger.V(4).Info("processing queueItem")

	if wr.process(ctx, item) {
		wr.queue.Forget(itemAny)
	} else {
		wr.queue.AddRateLimited(itemAny)
	}
	return true
}

// process returns true on success or unrecoverable error, false to retry
func (wr *whereResolver) process(ctx context.Context, item queueItem) bool {
	logger := klog.FromContext(ctx)
	cluster := item.cluster
	epName := item.name
	objName := item.toExternalName()
	sps, err := wr.spsLister.Cluster(cluster).Get(epName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Failed to fetch SinglePlacementSlice from local cache", "cluster", cluster, "epName", epName)
		return true // I think these errors are not transient
	}
	if err == nil {
		wr.resolutions.Put(objName, []*edgeapi.SinglePlacementSlice{sps})
	} else {
		wr.resolutions.Delete(objName)
	}
	return true
}
