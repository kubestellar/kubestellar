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

package scheduler

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
	schedulingv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/scheduling/v1alpha1"
	workloadv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/workload/v1alpha1"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	schedulingv1alpha1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/scheduling/v1alpha1"
	workloadv1alpha1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/workload/v1alpha1"
	schedulingv1alpha1listers "github.com/kcp-dev/kcp/pkg/client/listers/scheduling/v1alpha1"
	workloadv1alpha1listers "github.com/kcp-dev/kcp/pkg/client/listers/workload/v1alpha1"

	edgeclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/indexers"
)

const (
	ControllerName = "edge-scheduler"
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

	kcpClusterClient  kcpclientset.ClusterInterface
	edgeClusterClient edgeclientset.ClusterInterface

	edgePlacementLister  edgev1alpha1listers.EdgePlacementClusterLister
	edgePlacementIndexer cache.Indexer

	locationLister  schedulingv1alpha1listers.LocationClusterLister
	locationIndexer cache.Indexer

	synctargetLister  workloadv1alpha1listers.SyncTargetClusterLister
	synctargetIndexer cache.Indexer
}

func NewController(
	context context.Context,
	kcpClusterClient kcpclientset.ClusterInterface,
	edgeClusterClient edgeclientset.ClusterInterface,
	edgePlacementAccess edgev1alpha1informers.EdgePlacementClusterInformer,
	locationAccess schedulingv1alpha1informers.LocationClusterInformer,
	syncTargetAccess workloadv1alpha1informers.SyncTargetClusterInformer,
) (*controller, error) {
	context = klog.NewContext(context, klog.FromContext(context).WithValues("controller", ControllerName))
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ControllerName)

	c := &controller{
		context: context,
		queue:   queue,

		kcpClusterClient:  kcpClusterClient,
		edgeClusterClient: edgeClusterClient,

		edgePlacementLister:  edgePlacementAccess.Lister(),
		edgePlacementIndexer: edgePlacementAccess.Informer().GetIndexer(),

		locationLister:  locationAccess.Lister(),
		locationIndexer: locationAccess.Informer().GetIndexer(),

		synctargetLister:  syncTargetAccess.Lister(),
		synctargetIndexer: syncTargetAccess.Informer().GetIndexer(),
	}

	indexers.AddIfNotPresentOrDie(locationAccess.Informer().GetIndexer(), cache.Indexers{
		indexers.ByLogicalClusterPath: indexers.IndexByLogicalClusterPath,
	})

	edgePlacementAccess.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueueEdgePlacement,
		UpdateFunc: func(_, newObj interface{}) { c.enqueueEdgePlacement(newObj) },
		DeleteFunc: c.enqueueEdgePlacement,
	})

	locationAccess.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueLocation,
		UpdateFunc: func(old, obj interface{}) {
			oldLoc := old.(*schedulingv1alpha1.Location)
			newLoc := obj.(*schedulingv1alpha1.Location)
			if !apiequality.Semantic.DeepEqual(oldLoc.Spec, newLoc.Spec) || !apiequality.Semantic.DeepEqual(oldLoc.Labels, newLoc.Labels) {
				c.enqueueLocation(obj)
			}
		},
		DeleteFunc: c.enqueueLocation,
	})

	syncTargetAccess.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueSyncTarget,
		UpdateFunc: func(old, obj interface{}) {
			oldST := old.(*workloadv1alpha1.SyncTarget)
			newST := obj.(*workloadv1alpha1.SyncTarget)
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
	klog.FromContext(ctx).V(1).Info("processing queueItem")

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(i)

	if err := c.process(ctx, item); err != nil {
		runtime.HandleError(fmt.Errorf("%q controller didn't sync %q, err: %w", ControllerName, key, err))
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
