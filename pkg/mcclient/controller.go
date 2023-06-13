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

package mcclient

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	lcv1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
	mcclientset "github.com/kcp-dev/edge-mc/pkg/mcclient/clientset"
)

const (
	ControllerName = "logicalclusters-client"
)

type controller struct {
	context                context.Context
	logger                 logr.Logger
	queue                  workqueue.RateLimitingInterface
	logicalClusterInformer cache.SharedIndexInformer
	multiClusterClient     *multiClusterClient
}

func newController(
	context context.Context,
	logicalClusterInformer cache.SharedIndexInformer,
	multiClusterClient *multiClusterClient,
) *controller {
	context = klog.NewContext(context, klog.FromContext(context).WithValues("controller", ControllerName))

	c := &controller{
		context:                context,
		logger:                 klog.FromContext(context),
		queue:                  workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), ControllerName),
		logicalClusterInformer: logicalClusterInformer,
		multiClusterClient:     multiClusterClient,
	}

	logicalClusterInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldInfo := oldObj.(*lcv1alpha1.LogicalCluster)
			newInfo := newObj.(*lcv1alpha1.LogicalCluster)
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

	c.logger.Info("starting client logicalcluster controller")
	defer c.logger.Info("shutting down client logicalcluster controller")

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
	cluster, exists, err := c.logicalClusterInformer.GetIndexer().GetByKey(key)
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

func (c *controller) handleAdd(cluster interface{}) {
	clusterInfo, ok := cluster.(*lcv1alpha1.LogicalCluster)
	if !ok {
		runtime.HandleError(errors.New("unexpected object type. expected LogicalCluster"))
		return
	}
	// Only clusters in ready state are cached.
	// We will get another event when the cluster is Ready and then we cache it.
	if clusterInfo.Status.Phase != lcv1alpha1.LogicalClusterPhaseReady {
		c.handleDelete(clusterInfo.Name)
		return
	}

	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(clusterInfo.Status.ClusterConfig))
	if err != nil {
		runtime.HandleError(err)
		return
	}
	cs, err := mcclientset.NewForConfig(config)
	if err != nil {
		runtime.HandleError(err)
		return
	}

	c.multiClusterClient.lock.Lock()
	defer c.multiClusterClient.lock.Unlock()
	c.multiClusterClient.configs[clusterInfo.Name] = config
	c.multiClusterClient.clientsets[clusterInfo.Name] = cs
}

// handleDelete deletes cluster from the cache maps
func (c *controller) handleDelete(clusterName string) {
	c.multiClusterClient.lock.Lock()
	defer c.multiClusterClient.lock.Unlock()
	if _, ok := c.multiClusterClient.configs[clusterName]; !ok {
		return
	}
	delete(c.multiClusterClient.configs, clusterName)
	delete(c.multiClusterClient.clientsets, clusterName)
}
