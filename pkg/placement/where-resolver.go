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
	"sync"
	"time"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgev2alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev2alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
)

type whereResolver struct {
	ctx        context.Context
	logger     klog.Logger
	numThreads int
	queue      workqueue.RateLimitingInterface

	spsInformer     upstreamcache.SharedIndexInformer
	spsLister       edgev2alpha1listers.SinglePlacementSliceLister
	kbSpaceRelation kbuser.KubeBindSpaceRelation

	// resolutions maps EdgePlacement name to its ResolvedWhere
	resolutions RelayMap[ExternalName, ResolvedWhere]
}

type queueItem struct {
	GK      schema.GroupKind
	Cluster string
	Name    string
}

func (qi queueItem) toExternalName() ExternalName {
	return ExternalName{Cluster: qi.Cluster, Name: ObjectName(qi.Name)}
}

// NewWhereResolver returns a WhereResolver.
func NewWhereResolver(
	ctx context.Context,
	spsPreInformer edgev2alpha1informers.SinglePlacementSliceInformer,
	kbSpaceRelation kbuser.KubeBindSpaceRelation,
	numThreads int,
) WhereResolver {
	return func(receiver MappingReceiver[ExternalName, ResolvedWhere]) Runnable {
		controllerName := "where-resolver"
		logger := klog.FromContext(ctx).WithValues("part", controllerName)
		ctx = klog.NewContext(ctx, logger)
		wr := &whereResolver{
			ctx:             ctx,
			logger:          logger,
			numThreads:      numThreads,
			queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
			spsInformer:     spsPreInformer.Informer(),
			spsLister:       spsPreInformer.Lister(),
			kbSpaceRelation: kbSpaceRelation,
			resolutions:     NewRelayMap[ExternalName, ResolvedWhere](false),
		}
		wr.resolutions.AddReceiver(receiver, false)
		wr.spsInformer.AddEventHandler(WhereResolverClusterHandler{wr, mkgk(edgeapi.SchemeGroupVersion.Group, "SinglePlacementSlice")})
		if !upstreamcache.WaitForNamedCacheSync(controllerName, ctx.Done(), wr.spsInformer.HasSynced) {
			logger.Info("Failed to sync SinglePlacementSlices in time")
		}
		return wr
	}
}

func (wr *whereResolver) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(wr.numThreads)
	for i := 0; i < wr.numThreads; i++ {
		go func() {
			wait.Until(wr.runWorker, time.Second, ctx.Done())
			wg.Done()
		}()
	}
	wg.Wait()
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
	key, err := upstreamcache.DeletionHandlingMetaNamespaceKeyFunc(objAny)
	if err != nil {
		wr.logger.Error(err, "Failed to extract object reference", "object", objAny)
		return
	}
	_, name, err := upstreamcache.SplitMetaNamespaceKey(key)
	if err != nil {
		wr.logger.Error(err, "Impossible! SplitMetaClusterNamespaceKey failed", "key", key)
	}
	// enqueue with empty cluster. set it later
	item := queueItem{GK: gk, Cluster: "", Name: name}
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

	logger := klog.FromContext(wr.ctx).WithValues("group", item.GK.Group, "kind", item.GK.Kind, "cluster", item.Cluster, "name", item.Name)
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
	epName := item.Name
	sps, err := wr.spsLister.Get(epName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Failed to fetch SinglePlacementSlice from local cache", "epName", epName)
		return true // I think these errors are not transient
	}
	//change to consumer SpaceID
	_, spsOriginalName, kbSpaceID, err := kbuser.AnalyzeObjectID(sps)
	if err != nil {
		logger.Error(err, "Object does not appear to be a provider's copy of a consumer's object")
		return true
	}
	spaceID := wr.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if spaceID == "" {
		logger.Error(nil, "Failed to get consumer space ID from a provider's copy")
		return false
	}
	item.Name = spsOriginalName
	item.Cluster = spaceID
	objName := item.toExternalName()

	if err == nil {
		wr.resolutions.Put(objName, []*edgeapi.SinglePlacementSlice{sps})
	} else {
		wr.resolutions.Delete(objName)
	}
	return true
}
