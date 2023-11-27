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

package msclientlib

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	spacev1alpha1 "github.com/kubestellar/kubestellar/space-framework/pkg/apis/space/v1alpha1"
)

const (
	ControllerName = "multi-space-client"
)

type controller struct {
	context       context.Context
	logger        logr.Logger
	msClient      *multiSpaceClient
	queue         workqueue.RateLimitingInterface
	spaceInformer cache.SharedIndexInformer
}

func newController(
	context context.Context,
	spaceInformer cache.SharedIndexInformer,
	msClient *multiSpaceClient,
) *controller {
	context = klog.NewContext(context, klog.FromContext(context).WithValues("controller", ControllerName))

	c := &controller{
		context:       context,
		logger:        klog.FromContext(context),
		queue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ControllerName),
		spaceInformer: spaceInformer,
		msClient:      msClient,
	}

	spaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldInfo := oldObj.(*spacev1alpha1.Space)
			newInfo := newObj.(*spacev1alpha1.Space)
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

// Run starts the controller, which stops when c.context.Done() is closed.
func (c *controller) Run(numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("starting client space controller")
	defer c.logger.Info("shutting down client space controller")

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
	space, exists, err := c.spaceInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		c.handleDelete(key)
	} else {
		c.handleAdd(space, key)
	}
	return nil
}

func (c *controller) handleAdd(space interface{}, spaceKey string) {
	nSpace, ok := space.(*spacev1alpha1.Space)
	if !ok {
		runtime.HandleError(errors.New("unexpected object type. expected Space"))
		return
	}
	// Only spaces in ready state are cached.
	// We will get another event when the space is Ready and then we cache it.
	if nSpace.Status.Phase != spacev1alpha1.SpacePhaseReady {
		c.handleDelete(spaceKey)
		return
	}
	// add space reference to cache
	c.cacheSpaceReference(nSpace, spaceKey)

	// TODO We assume the space is in a pod in the same cluster where the controllers are deployed
	// Need to add flags to indicate it
	config, err := c.msClient.getRestConfigFromSecret(true, nSpace)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	c.logger.Info("New space detected", "space", nSpace.Name)
	c.msClient.lock.Lock()
	defer c.msClient.lock.Unlock()
	c.msClient.configs[spaceKey] = config
}

// handleDelete deletes space from the cache maps
func (c *controller) handleDelete(spaceKey string) {
	c.msClient.lock.Lock()
	defer c.msClient.lock.Unlock()
	for ref, space := range c.msClient.refToSpace {
		if space == spaceKey {
			delete(c.msClient.refToSpace, ref)
			break
		}
	}
	delete(c.msClient.configs, spaceKey)
}

func (c *controller) cacheSpaceReference(space *spacev1alpha1.Space, spaceKey string) {
	annotations := space.GetAnnotations()
	if annotations == nil {
		return
	}
	c.msClient.lock.Lock()
	defer c.msClient.lock.Unlock()
	if ref, ok := annotations[refAnnotationKey]; ok {
		c.msClient.refToSpace[ref] = spaceKey
	}
}
