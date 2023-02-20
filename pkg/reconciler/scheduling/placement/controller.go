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

package placement

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	kcpcorev1informers "github.com/kcp-dev/client-go/informers/core/v1"
	corev1listers "github.com/kcp-dev/client-go/listers/core/v1"
	schedulingv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/scheduling/v1alpha1"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	schedulingv1alpha1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/scheduling/v1alpha1"
	schedulingv1alpha1listers "github.com/kcp-dev/kcp/pkg/client/listers/scheduling/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/logging"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgeclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/indexers"
)

const (
	ControllerName      = "edge-scheduler"
	byLocationWorkspace = ControllerName + "-byLocationWorkspace"
)

// NewController returns a new controller placing namespaces onto locations by create
// a placement annotation..
func NewController(
	kcpClusterClient kcpclientset.ClusterInterface,
	edgeClusterClient edgeclientset.ClusterInterface,
	namespaceInformer kcpcorev1informers.NamespaceClusterInformer,
	locationInformer schedulingv1alpha1informers.LocationClusterInformer,
	placementInformer schedulingv1alpha1informers.PlacementClusterInformer,
	edgePlacementInformer edgev1alpha1informers.EdgePlacementClusterInformer,
) (*controller, error) {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ControllerName)

	c := &controller{
		queue: queue,
		enqueueAfter: func(ns *corev1.Namespace, duration time.Duration) {
			key, err := kcpcache.MetaClusterNamespaceKeyFunc(ns)
			if err != nil {
				runtime.HandleError(err)
				return
			}
			queue.AddAfter(key, duration)
		},
		kcpClusterClient:  kcpClusterClient,
		edgeClusterClient: edgeClusterClient,

		namespaceLister: namespaceInformer.Lister(),

		locationLister:  locationInformer.Lister(),
		locationIndexer: locationInformer.Informer().GetIndexer(),

		placementLister:  placementInformer.Lister(),
		placementIndexer: placementInformer.Informer().GetIndexer(),

		edgePlacementLister:  edgePlacementInformer.Lister(),
		edgePlacementIndexer: edgePlacementInformer.Informer().GetIndexer(),
	}

	if err := placementInformer.Informer().AddIndexers(cache.Indexers{
		byLocationWorkspace: indexByLocationWorkspace,
	}); err != nil {
		return nil, err
	}

	indexers.AddIfNotPresentOrDie(locationInformer.Informer().GetIndexer(), cache.Indexers{
		indexers.ByLogicalClusterPath: indexers.IndexByLogicalClusterPath,
	})

	// namespaceBlocklist holds a set of namespaces that should never be synced from kcp to physical clusters.
	var namespaceBlocklist = sets.NewString("kube-system", "kube-public", "kube-node-lease")
	namespaceInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			switch ns := obj.(type) {
			case *corev1.Namespace:
				return !namespaceBlocklist.Has(ns.Name)
			case cache.DeletedFinalStateUnknown:
				return true
			default:
				return false
			}
		},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: c.enqueueNamespace,
			UpdateFunc: func(old, obj interface{}) {
				oldNs := old.(*corev1.Namespace)
				newNs := obj.(*corev1.Namespace)

				if !reflect.DeepEqual(oldNs.Annotations, newNs.Annotations) {
					c.enqueueNamespace(obj)
				}
			},
			DeleteFunc: c.enqueueNamespace,
		},
	})

	locationInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.enqueueLocation,
			UpdateFunc: func(old, obj interface{}) {
				oldLoc := old.(*schedulingv1alpha1.Location)
				newLoc := obj.(*schedulingv1alpha1.Location)
				if !reflect.DeepEqual(oldLoc.Spec, newLoc.Spec) || !reflect.DeepEqual(oldLoc.Labels, newLoc.Labels) {
					c.enqueueLocation(obj)
				}
			},
			DeleteFunc: c.enqueueLocation,
		},
	)

	placementInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueuePlacement,
		UpdateFunc: func(_, obj interface{}) { c.enqueuePlacement(obj) },
		DeleteFunc: c.enqueuePlacement,
	})

	edgePlacementInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueuePlacement,
		UpdateFunc: func(_, newObj interface{}) { c.enqueuePlacement(newObj) },
		DeleteFunc: c.enqueuePlacement,
	})

	return c, nil
}

// controller.
type controller struct {
	queue        workqueue.RateLimitingInterface
	enqueueAfter func(*corev1.Namespace, time.Duration)

	kcpClusterClient  kcpclientset.ClusterInterface
	edgeClusterClient edgeclientset.ClusterInterface

	namespaceLister corev1listers.NamespaceClusterLister

	locationLister  schedulingv1alpha1listers.LocationClusterLister
	locationIndexer cache.Indexer

	placementLister  schedulingv1alpha1listers.PlacementClusterLister
	placementIndexer cache.Indexer

	edgePlacementLister  edgev1alpha1listers.EdgePlacementClusterLister
	edgePlacementIndexer cache.Indexer
}

func (c *controller) enqueuePlacement(obj interface{}) {
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	logger := logging.WithQueueKey(logging.WithReconciler(klog.Background(), ControllerName), key)
	logger.V(2).Info("queueing EdgePlacement")
	c.queue.Add(key)
}

// enqueueNamespace enqueues all placements for the namespace.
func (c *controller) enqueueNamespace(obj interface{}) {
	logger := logging.WithReconciler(klog.Background(), ControllerName)
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	clusterName, _, _, err := kcpcache.SplitMetaClusterNamespaceKey(key)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	placements, err := c.placementLister.Cluster(clusterName).List(labels.Everything())
	if err != nil {
		runtime.HandleError(err)
		return
	}

	for _, placement := range placements {
		namespaceKey := key
		key, err := kcpcache.MetaClusterNamespaceKeyFunc(placement)
		if err != nil {
			runtime.HandleError(err)
			continue
		}
		logging.WithQueueKey(logger, key).V(2).Info("queueing Placement because Namespace changed", "namespace", namespaceKey)
		c.queue.Add(key)
	}
}

func (c *controller) enqueueLocation(obj interface{}) {
	logger := logging.WithReconciler(klog.Background(), ControllerName)
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	clusterName, _, _, err := kcpcache.SplitMetaClusterNamespaceKey(key)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	placements, err := c.placementIndexer.ByIndex(byLocationWorkspace, clusterName.String())
	if err != nil {
		runtime.HandleError(err)
		return
	}

	for _, obj := range placements {
		placement := obj.(*schedulingv1alpha1.Placement)
		locationKey := key
		key, err := kcpcache.MetaClusterNamespaceKeyFunc(placement)
		if err != nil {
			runtime.HandleError(err)
			continue
		}
		logging.WithQueueKey(logger, key).V(2).Info("queueing Placement because Location changed", "location", locationKey)
		c.queue.Add(key)
	}
}

// Start starts the controller, which stops when ctx.Done() is closed.
func (c *controller) Start(ctx context.Context, numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	logger := logging.WithReconciler(klog.FromContext(ctx), ControllerName)
	ctx = klog.NewContext(ctx, logger)
	logger.Info("Starting controller")
	defer logger.Info("Shutting down controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(ctx, c.startWorker, time.Second)
	}

	<-ctx.Done()
}

func (c *controller) startWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *controller) processNextWorkItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	k, quit := c.queue.Get()
	if quit {
		return false
	}
	key := k.(string)

	logger := logging.WithQueueKey(klog.FromContext(ctx), key)
	ctx = klog.NewContext(ctx, logger)
	logger.V(1).Info("processing key")

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(key)

	if err := c.process(ctx, key); err != nil {
		runtime.HandleError(fmt.Errorf("%q controller failed to sync %q, err: %w", ControllerName, key, err))
		c.queue.AddRateLimited(key)
		return true
	}
	c.queue.Forget(key)
	return true
}

func (c *controller) process(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)
	clusterName, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(key)
	if err != nil {
		logger.Error(err, "invalid key")
		return nil
	}

	obj, err := c.edgePlacementLister.Cluster(clusterName).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("object deleted before handled")
			return nil
		}
		return err
	}
	obj = obj.DeepCopy()

	logger = logging.WithObject(logger, obj)
	ctx = klog.NewContext(ctx, logger)

	reconcileErr := c.reconcile(ctx, obj)

	ws := logicalcluster.From(obj)

	sps := &edgev1alpha1.SinglePlacementSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hardcoded",
		},
		Destinations: []edgev1alpha1.SinglePlacement{},
	}
	_, err = c.edgeClusterClient.Cluster(ws.Path()).EdgeV1alpha1().SinglePlacementSlices().Create(ctx, sps, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "failed creating singlePlacementSlice")
	} else {
		logger.Info("created SinglePlacementSlice", "name", sps.Name, "workspace", ws.String())
	}

	ep := &edgev1alpha1.EdgePlacement{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hardcoded",
		},
		Spec: edgev1alpha1.EdgePlacementSpec{
			LocationSelectors: []metav1.LabelSelector{},
			NamespaceSelector: metav1.LabelSelector{},
		},
	}
	_, err = c.edgeClusterClient.Cluster(ws.Path()).EdgeV1alpha1().EdgePlacements().Create(ctx, ep, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "failed creating EdgePlacement")
	} else {
		logger.Info("created EdgePlacement", "name", sps.Name, "workspace", ws.String())
	}

	// TODO: If the object being reconciled changed as a result, update it.
	return reconcileErr
}
