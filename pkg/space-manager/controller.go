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

package spacemanager

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	spacev1alpha1 "github.com/kubestellar/kubestellar/pkg/apis/space/v1alpha1"
	edgeclient "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
)

const ()

type triggeringKind string

const (
	triggeringKindSpace             triggeringKind = "Space"
	triggeringKindSpaceProviderDesc triggeringKind = "SpaceProviderDesc"
	controllerName                                 = "space-manager"
)

type queueItem struct {
	triggeringKind triggeringKind
	key            string
}

type controller struct {
	ctx                   context.Context
	clientset             edgeclient.Interface
	k8sClientset          *kubeclient.Clientset
	logger                logr.Logger
	queue                 workqueue.RateLimitingInterface
	spaceInformer         cache.SharedIndexInformer
	spaceProviderInformer cache.SharedIndexInformer
	providers             map[string]*provider
	lock                  sync.Mutex
}

// NewController returns space-manager controller
func NewController(
	ctx context.Context,
	clientset edgeclient.Interface,
	k8sClientset *kubeclient.Clientset,
	spaceInformer cache.SharedIndexInformer,
	providerInformer cache.SharedIndexInformer,
) *controller {
	ctx = klog.NewContext(ctx, klog.FromContext(ctx).WithValues("controller", controllerName))

	c := &controller{
		ctx:                   ctx,
		clientset:             clientset,
		k8sClientset:          k8sClientset,
		logger:                klog.FromContext(ctx),
		queue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		spaceInformer:         spaceInformer,
		spaceProviderInformer: providerInformer,
		providers:             make(map[string]*provider),
	}

	spaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueueSpace,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*spacev1alpha1.Space)
			new := newObj.(*spacev1alpha1.Space)
			if !reflect.DeepEqual(old, new) {
				c.enqueueSpace(newObj)
			}
		},
		DeleteFunc: c.enqueueSpace,
	})

	providerInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.enqueueSpaceProviderDesc,
		DeleteFunc: c.enqueueSpaceProviderDesc,
	})

	return c
}

func (c *controller) enqueueSpace(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	c.logger.V(2).Info("queueing Space", "key", key)
	c.queue.Add(
		queueItem{
			triggeringKind: triggeringKindSpace,
			key:            key,
		},
	)
}

func (c *controller) enqueueSpaceProviderDesc(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	c.logger.V(2).Info("queueing SpaceProviderDesc", "key", key)
	c.queue.Add(
		queueItem{
			triggeringKind: triggeringKindSpaceProviderDesc,
			key:            key,
		},
	)
}

// Run starts the controller, which stops when c.context.Done() is closed.
func (c *controller) Run(numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("starting manager space controller")
	defer c.logger.Info("shutting down manager space controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(c.ctx, c.runWorker, time.Second)
	}

	<-c.ctx.Done()
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

	// Done with this key, unblock other workers.
	defer c.queue.Done(i)

	if err := c.process(item); err != nil {
		c.queue.AddRateLimited(i)
		return true
	}
	c.queue.Forget(i)
	return true
}

func (c *controller) process(item queueItem) error {
	trigger, key := item.triggeringKind, item.key
	var err error
	switch trigger {
	case triggeringKindSpace:
		err = c.reconcileSpace(key)
	case triggeringKindSpaceProviderDesc:
		err = c.reconcileSpaceProviderDesc(key)
	}
	return err
}
