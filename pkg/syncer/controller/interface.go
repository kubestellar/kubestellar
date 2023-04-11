package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/kcp-dev/edge-mc/pkg/syncer/shared"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

type controllerBase struct {
	name              string
	target            string
	process           func(ctx context.Context, key string) error
	queue             workqueue.RateLimitingInterface
	informerHasSynced cache.InformerSynced
}

// Ready returns true if the controller is ready to return the GVRs to sync.
// It implements [informer.GVRSource.Ready].
func (c *controllerBase) Ready() bool {
	return c.informerHasSynced()
}

func (c *controllerBase) enqueue(obj interface{}, logger klog.Logger) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	shared.WithQueueKey(logger, key).V(2).Info("queueing " + c.target)
	c.queue.Add(key)
}

// Run the controller workers.
func (c *controllerBase) Run(ctx context.Context, numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	logger := shared.WithReconciler(klog.FromContext(ctx), c.name)
	ctx = klog.NewContext(ctx, logger)
	logger.Info("Starting controller")
	defer logger.Info("Shutting down controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	<-ctx.Done()
}

func (c *controllerBase) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *controllerBase) processNextWorkItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	k, quit := c.queue.Get()
	if quit {
		return false
	}
	key := k.(string)

	logger := klog.FromContext(ctx).WithValues("key", key)
	ctx = klog.NewContext(ctx, logger)
	logger.V(1).Info("processing key")

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(key)

	if err := c.process(ctx, key); err != nil {
		runtime.HandleError(fmt.Errorf("%q controller failed to sync %q, err: %w", c.name, key, err))
		c.queue.AddRateLimited(key)
		return true
	}

	c.queue.Forget(key)
	return true
}
