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

package where_resolver

import (
	"context"
	"fmt"
	"time"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
	edgev2alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev2alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
)

const (
	ControllerName = "where-resolver"
)

type triggeringKind string

const (
	triggeringKindEdgePlacement triggeringKind = "EdgePlacement"
	triggeringKindLocation      triggeringKind = "Location"
	triggeringKindSyncTarget    triggeringKind = "SyncTarget"
)

type queueItem struct {
	triggeringKind triggeringKind
	key            string
}

type controller struct {
	context context.Context
	queue   workqueue.RateLimitingInterface

	edgeClusterClient edgeclientset.ClusterInterface

	singlePlacementSliceLister  edgev2alpha1listers.SinglePlacementSliceLister
	singlePlacementSliceIndexer cache.Indexer

	edgePlacementLister  edgev2alpha1listers.EdgePlacementLister
	edgePlacementIndexer cache.Indexer

	locationLister  edgev2alpha1listers.LocationLister
	locationIndexer cache.Indexer

	synctargetLister  edgev2alpha1listers.SyncTargetLister
	synctargetIndexer cache.Indexer
}

func NewController(
	context context.Context,
	edgeClusterClient edgeclientset.ClusterInterface,
	edgePlacementAccess edgev2alpha1informers.EdgePlacementInformer,
	singlePlacementSliceAccess edgev2alpha1informers.SinglePlacementSliceInformer,
	locationAccess edgev2alpha1informers.LocationInformer,
	syncTargetAccess edgev2alpha1informers.SyncTargetInformer,
) (*controller, error) {
	context = klog.NewContext(context, klog.FromContext(context).WithValues("controller", ControllerName))
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ControllerName)

	c := &controller{
		context: context,
		queue:   queue,

		edgeClusterClient: edgeClusterClient,

		edgePlacementLister:  edgePlacementAccess.Lister(),
		edgePlacementIndexer: edgePlacementAccess.Informer().GetIndexer(),

		singlePlacementSliceLister:  singlePlacementSliceAccess.Lister(),
		singlePlacementSliceIndexer: singlePlacementSliceAccess.Informer().GetIndexer(),

		locationLister:  locationAccess.Lister(),
		locationIndexer: locationAccess.Informer().GetIndexer(),

		synctargetLister:  syncTargetAccess.Lister(),
		synctargetIndexer: syncTargetAccess.Informer().GetIndexer(),
	}

	edgePlacementAccess.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueueEdgePlacement,
		UpdateFunc: func(_, newObj interface{}) { c.enqueueEdgePlacement(newObj) },
		DeleteFunc: c.enqueueEdgePlacement,
	})

	locationAccess.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueLocation,
		UpdateFunc: func(old, obj interface{}) {
			oldLoc := old.(*edgev2alpha1.Location)
			newLoc := obj.(*edgev2alpha1.Location)
			if !apiequality.Semantic.DeepEqual(oldLoc.Spec, newLoc.Spec) || !apiequality.Semantic.DeepEqual(oldLoc.Labels, newLoc.Labels) {
				c.enqueueLocation(obj)
			}
		},
		DeleteFunc: c.enqueueLocation,
	})

	syncTargetAccess.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueSyncTarget,
		UpdateFunc: func(old, obj interface{}) {
			oldST := old.(*edgev2alpha1.SyncTarget)
			newST := obj.(*edgev2alpha1.SyncTarget)
			if !apiequality.Semantic.DeepEqual(oldST.Spec, newST.Spec) || !apiequality.Semantic.DeepEqual(oldST.Labels, newST.Labels) {
				c.enqueueSyncTarget((obj))
			}
		},
		DeleteFunc: c.enqueueSyncTarget,
	})

	return c, nil
}

func (c *controller) enqueueEdgePlacement(obj interface{}) {
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	klog.FromContext(c.context).V(2).Info("queueing EdgePlacement", "key", key)
	c.queue.Add(
		queueItem{
			triggeringKind: triggeringKindEdgePlacement,
			key:            key,
		},
	)
}

func (c *controller) enqueueLocation(obj interface{}) {
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	klog.FromContext(c.context).V(2).Info("queueing Location", "key", key)
	c.queue.Add(
		queueItem{
			triggeringKind: triggeringKindLocation,
			key:            key,
		},
	)
}

func (c *controller) enqueueSyncTarget(obj interface{}) {
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	klog.FromContext(c.context).V(2).Info("queueing SyncTarget", "key", key)
	c.queue.Add(
		queueItem{
			triggeringKind: triggeringKindSyncTarget,
			key:            key,
		},
	)
}

// Run starts the controller, which stops when c.context.Done() is closed.
func (c *controller) Run(numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	logger := klog.FromContext(c.context)
	logger.Info("starting controller")
	defer logger.Info("shutting down controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(c.context, c.runWorker, time.Second)
	}

	<-c.context.Done()
}

func (c *controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *controller) processNextWorkItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	i, quit := c.queue.Get()
	if quit {
		return false
	}
	item := i.(queueItem)
	key := item.key

	ctx = klog.NewContext(ctx, klog.FromContext(ctx).WithValues("triggeringKind", item.triggeringKind, "key", key))
	klog.FromContext(ctx).V(2).Info("processing queueItem")

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(i)

	if err := c.process(ctx, item); err != nil {
		runtime.HandleError(fmt.Errorf("%q controller didn't sync %q, err: %w", ControllerName, key, err))
		c.queue.AddRateLimited(i)
		return true
	}
	c.queue.Forget(i)
	return true
}

func (c *controller) process(ctx context.Context, item queueItem) error {
	tk, key := item.triggeringKind, item.key
	var err error
	switch tk {
	case triggeringKindEdgePlacement:
		err = c.reconcileOnEdgePlacement(ctx, key)
	case triggeringKindLocation:
		err = c.reconcileOnLocation(ctx, key)
	case triggeringKindSyncTarget:
		err = c.reconcileOnSyncTarget(ctx, key)
	}
	return err
}
