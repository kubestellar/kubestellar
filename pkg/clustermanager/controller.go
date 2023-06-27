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

package clustermanager

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	lcv1alpha1 "github.com/kubestellar/kubestellar/pkg/apis/logicalcluster/v1alpha1"
	edgeclient "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	providerclient "github.com/kubestellar/kubestellar/pkg/clustermanager/provider-client-interface"
	clusterprovider "github.com/kubestellar/kubestellar/pkg/clustermanager/provider-client-interface/cluster"
)

const ()

type triggeringKind string

const (
	triggeringKindLogicalCluster      triggeringKind = "LogicalCluster"
	triggeringKindClusterProviderDesc triggeringKind = "ClusterProviderDesc"
	controllerName                                   = "logicalcluster-manager"
)

type queueItem struct {
	triggeringKind triggeringKind
	key            string
}

type providerInfo struct {
	providerClient  providerclient.ProviderClient
	providerWatcher clusterprovider.Watcher
}

type controller struct {
	context                 context.Context
	clientset               edgeclient.Interface
	logger                  logr.Logger
	queue                   workqueue.RateLimitingInterface
	logicalClusterInformer  cache.SharedIndexInformer
	clusterProviderInformer cache.SharedIndexInformer
	providers               map[string]providerInfo
	lock                    sync.Mutex
}

// NewController returns logicalcluster-manager controller
func NewController(
	context context.Context,
	clientset edgeclient.Interface,
	logicalClusterInformer cache.SharedIndexInformer,
	providerInformer cache.SharedIndexInformer,
) *controller {
	context = klog.NewContext(context, klog.FromContext(context).WithValues("controller", controllerName))

	c := &controller{
		context:                 context,
		clientset:               clientset,
		logger:                  klog.FromContext(context),
		queue:                   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		logicalClusterInformer:  logicalClusterInformer,
		clusterProviderInformer: providerInformer,
		providers:               make(map[string]providerInfo),
	}

	logicalClusterInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueLogicalCluster,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*lcv1alpha1.LogicalCluster)
			new := newObj.(*lcv1alpha1.LogicalCluster)
			if !reflect.DeepEqual(old.Status, new.Status) {
				c.enqueueLogicalCluster(newObj)
			}
		},
		DeleteFunc: c.enqueueLogicalCluster,
	})

	providerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueueClusterProviderDesc,
		DeleteFunc: c.enqueueClusterProviderDesc,
	})

	return c
}

func (c *controller) enqueueLogicalCluster(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	c.logger.V(2).Info("queueing LogicalCluster", "key", key)
	c.queue.Add(
		queueItem{
			triggeringKind: triggeringKindLogicalCluster,
			key:            key,
		},
	)
}

func (c *controller) enqueueClusterProviderDesc(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	c.logger.V(2).Info("queueing ClusterProviderDesc", "key", key)
	c.queue.Add(
		queueItem{
			triggeringKind: triggeringKindClusterProviderDesc,
			key:            key,
		},
	)
}

// Run starts the controller, which stops when c.context.Done() is closed.
func (c *controller) Run(numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("starting manager logicalcluster controller")
	defer c.logger.Info("shutting down manager logicalcluster controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(c.context, c.runWorker, time.Second)
	}

	<-c.context.Done()
}

func (c *controller) runWorker(ctx context.Context) {
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	i, quit := c.queue.Get()
	if quit {
		return false
	}
	item := i.(queueItem)
	key := item.key

	// Done with this key, unblock other workers.
	defer c.queue.Done(key)

	if err := c.process(item); err != nil {
		runtime.HandleError(err)
		c.queue.AddRateLimited(key)
		return true
	}
	c.queue.Forget(key)
	return true
}

func (c *controller) process(item queueItem) error {
	trigger, key := item.triggeringKind, item.key
	var err error
	switch trigger {
	case triggeringKindLogicalCluster:
		err = c.reconcileLogicalCluster(key)
	case triggeringKindClusterProviderDesc:
		err = c.reconcileClusterProviderDesc(key)
	}
	return err
}
