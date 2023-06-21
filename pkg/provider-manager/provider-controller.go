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

package providermanager

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

	clusterproviderclient "github.com/kcp-dev/edge-mc/cluster-provider-client"
	lcv1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
	edgeclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
)

const (
	ControllerName = "provider-manager"
)

type controller struct {
	ctx              context.Context
	logger           logr.Logger
	clientset        edgeclient.Interface
	queue            workqueue.RateLimitingInterface
	providerInformer cache.SharedIndexInformer
	listProviders    map[string]clusterproviderclient.ProviderClient
	lock             sync.Mutex
}

func NewController(
	ctx context.Context,
	clientset edgeclient.Interface,
	providerInformer cache.SharedIndexInformer,
) *controller {
	ctx = klog.NewContext(ctx, klog.FromContext(ctx).WithValues("controller", ControllerName))

	c := &controller{
		ctx:              ctx,
		logger:           klog.FromContext(ctx),
		clientset:        clientset,
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ControllerName),
		providerInformer: providerInformer,
		listProviders:    make(map[string]clusterproviderclient.ProviderClient),
	}

	providerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldInfo := oldObj.(*lcv1alpha1.ClusterProviderDesc)
			newInfo := newObj.(*lcv1alpha1.ClusterProviderDesc)
			if !reflect.DeepEqual(oldInfo.Status, newInfo.Status) {
				c.enqueue(newObj)
			}
		},
		DeleteFunc: c.enqueue,
	})

	return c
}

func (c *controller) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.queue.Add(key)
}

// Run starts the controller, which stops when c.ctx.Done() is closed.
func (c *controller) Run(numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("starting provider controller")
	defer c.logger.Info("shutting down provider controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(c.ctx, c.runWorker, time.Second)
	}

	<-c.ctx.Done()
}

func (c *controller) runWorker(ctx context.Context) {
	c.logger.Info("runWorker")
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	keyRaw := key.(string)

	// Done with this key, unblock other workers.
	defer c.queue.Done(key)

	if err := c.process(keyRaw); err != nil {
		runtime.HandleError(err)
		c.queue.AddRateLimited(key)
		return true
	}
	c.queue.Forget(key)
	return true
}

func (c *controller) process(key string) error {
	cluster, exists, err := c.providerInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		c.handleDelete(key)
	} else {
		c.handleAdd(cluster)
	}
	return nil
}
