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

package edgeplacement

import (
	"context"
	"fmt"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	schedulingv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/scheduling/v1alpha1"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	schedulingv1alpha1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/scheduling/v1alpha1"
	schedulingv1alpha1listers "github.com/kcp-dev/kcp/pkg/client/listers/scheduling/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/logging"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgeclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/indexers"
)

const (
	ControllerName = "edge-scheduler"
)

// NewController returns a new controller placing namespaces onto locations by create
// a placement annotation..
func NewController(
	kcpClusterClient kcpclientset.ClusterInterface,
	edgeClusterClient edgeclientset.ClusterInterface,
	locationInformer schedulingv1alpha1informers.LocationClusterInformer,
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

		locationLister:  locationInformer.Lister(),
		locationIndexer: locationInformer.Informer().GetIndexer(),

		edgePlacementLister:  edgePlacementInformer.Lister(),
		edgePlacementIndexer: edgePlacementInformer.Informer().GetIndexer(),
	}

	indexers.AddIfNotPresentOrDie(locationInformer.Informer().GetIndexer(), cache.Indexers{
		indexers.ByLogicalClusterPath: indexers.IndexByLogicalClusterPath,
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

	locationLister  schedulingv1alpha1listers.LocationClusterLister
	locationIndexer cache.Indexer

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

func (c *controller) enqueueLocation(obj interface{}) {
	// logger := logging.WithReconciler(klog.Background(), ControllerName)
	// key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(obj)
	// if err != nil {
	// 	runtime.HandleError(err)
	// 	return
	// }
	// clusterName, _, _, err := kcpcache.SplitMetaClusterNamespaceKey(key)
	// if err != nil {
	// 	runtime.HandleError(err)
	// 	return
	// }

	// TODO: We need to enqueue edgeplacements because of changes from locations.
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
			logger.V(3).Info("object deleted before handled")
			return nil
		}
		return err
	}
	obj = obj.DeepCopy()

	logger = logging.WithObject(logger, obj)
	ctx = klog.NewContext(ctx, logger)

	reconcileErr := c.reconcile(ctx, obj)

	// TODO: find a better place for this logic
	_, err = c.edgeClusterClient.EdgeV1alpha1().SinglePlacementSlices().Cluster(clusterName.Path()).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(2).Info("creating SinglePlacementSlice", "workspace", clusterName.String(), "name", name)
			sps := &edgev1alpha1.SinglePlacementSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name: obj.Name,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: edgev1alpha1.SchemeGroupVersion.String(),
							Kind:       "EdgePlacement",
							Name:       obj.Name,
							UID:        obj.UID,
						},
					},
				},
				Destinations: []edgev1alpha1.SinglePlacement{},
			}
			_, err = c.edgeClusterClient.Cluster(clusterName.Path()).EdgeV1alpha1().SinglePlacementSlices().Create(ctx, sps, metav1.CreateOptions{})
			if err != nil {
				logger.Error(err, "failed creating singlePlacementSlice", "workspace", clusterName.String(), "name", sps.Name)
			} else {
				logger.V(2).Info("created SinglePlacementSlice", "workspace", clusterName.String(), "name", sps.Name)
			}
		} else {
			logger.Error(err, "failed getting SinglePlacementSlice for EdgePlacement", "workspace", clusterName.String(), "name", name)
		}
	}

	// TODO: If the object being reconciled changed as a result, update it.
	return reconcileErr
}
