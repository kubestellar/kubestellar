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
	"fmt"
	"reflect"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	clusterproviderclient "github.com/kcp-dev/edge-mc/cluster-provider-client"
	cluster "github.com/kcp-dev/edge-mc/cluster-provider-client/cluster"
	lcv1alpha1apis "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
	edgeclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
	lcclient "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/logicalcluster/v1alpha1"
)

const controllerName = "cluster-manager"

type op string

const (
	opAdd    op = "Add"
	opUpdate op = "Update"
	opDelete op = "Delete"
)

type queueItem struct {
	op  op
	key any
}

type Controller struct {
	kubeconfig       *string
	ctx              context.Context
	clusterclientset edgeclient.Interface
	clusterInformer  lcclient.LogicalClusterInformer
	queue            workqueue.RateLimitingInterface
}

func NewController(
	kubeconfig *string,
	ctx context.Context,
	clusterclientset edgeclient.Interface,
	clusterInformer lcclient.LogicalClusterInformer) *Controller {
	logger := klog.FromContext(ctx)

	// TODO: We are keeping a hash table of logical cluster provider clients
	// this might change once we create the cluster provider client resource.
	clusterproviderclient.ProviderList = make(map[string]clusterproviderclient.ProviderClient)

	controller := &Controller{
		kubeconfig:       kubeconfig,
		ctx:              ctx,
		clusterclientset: clusterclientset,
		clusterInformer:  clusterInformer,
		queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(),
			"cluster-controller"),
	}

	logger.Info("Setting up event handlers")
	clusterInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logger.Info("Add cluster")
			controller.queue.Add(
				queueItem{
					op:  opAdd,
					key: obj,
				},
			)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			// Most fields in the logical cluster are immutable with the
			// exception of the URL and fields in the status.
			oldInfo := oldObj.(*lcv1alpha1apis.LogicalCluster)
			newInfo := newObj.(*lcv1alpha1apis.LogicalCluster)
			if !reflect.DeepEqual(oldInfo.Status, newInfo.Status) {
				controller.queue.Add(
					queueItem{
						op:  opUpdate,
						key: newObj,
					},
				)
			}
		},
		DeleteFunc: func(delObj interface{}) {
			logger.Info("Delete cluster")
			controller.queue.Add(
				queueItem{
					op:  opDelete,
					key: delObj,
				},
			)
		},
	})

	logger.Info("New %s controller", controllerName)
	return controller
}

// Run starts the controller, which stops when c.context.Done() is closed.
func (c *Controller) Run(numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	logger := klog.FromContext(c.ctx)
	logger.Info("starting controller")
	defer logger.Info("shutting down controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(c.ctx, c.runWorker, time.Second)
	}

	<-c.ctx.Done()
}

func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	// Wait until there is a new item in the working queue
	i, quit := c.queue.Get()
	if quit {
		return false
	}

	// No matter what, tell the queue we're done with this key, to unblock
	// other workers.
	defer c.queue.Done(i)

	//ES: not sure we should treat all errors in the same way (queue back)
	if err := c.process(ctx, i.(queueItem)); err != nil {
		runtime.HandleError(fmt.Errorf("%q controller didn't sync, err: %w", controllerName, err))
		c.queue.AddRateLimited(i)
		return true
	}
	c.queue.Forget(i)
	return true
}

func (c *Controller) processAdd(ctx context.Context, key any) error {
	logger := klog.FromContext(ctx)
	var err error

	newClusterConfig := key.(*lcv1alpha1apis.LogicalCluster)
	clusterName := newClusterConfig.Spec.ClusterName

	providerInfo, err := c.clusterclientset.LogicalclusterV1alpha1().ClusterProviderDescs().Get(ctx, newClusterConfig.Spec.ClusterProviderDesc, v1.GetOptions{})
	if err != nil {
		logger.Error(err, "failed to get the provider resource")
		return err
	}

	// TODO: we have yet to finalize the fields in the provider,
	// so likely this one liner will expend in the future.
	provider := clusterproviderclient.GetProviderClient(providerInfo.Spec.ProviderType, newClusterConfig.Spec.ClusterProviderDesc)

	// Update status to NotReady
	newClusterConfig.Status.Phase = lcv1alpha1apis.LogicalClusterPhaseNotReady
	_, err = c.clusterclientset.
		LogicalclusterV1alpha1().
		LogicalClusters(newClusterConfig.Spec.ClusterProviderDesc).
		Update(ctx, newClusterConfig, v1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to update cluster status.")
		return err
	}

	// Create cluster
	var opts cluster.Options
	//ES: what exactly is this kubeconfig
	opts.KubeconfigPath = *c.kubeconfig
	newCluster, err := provider.Create(ctx, clusterName, opts)
	if err != nil {
		logger.Error(err, "failed to create cluster")
		return err
	}
	logger.Info("Done creating cluster", clusterName)

	// Update the new cluster's status - specifically the config string and the phase
	newClusterConfig.Status.ClusterConfig = newCluster.Config
	newClusterConfig.Status.Phase = lcv1alpha1apis.LogicalClusterPhaseReady
	_, err = c.clusterclientset.
		LogicalclusterV1alpha1().
		LogicalClusters(newClusterConfig.Spec.ClusterProviderDesc).
		Update(ctx, newClusterConfig, v1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to update cluster status.")
		return err
	}

	return err
}

func (c *Controller) processUpdate(ctx context.Context, key any) error {
	logger := klog.FromContext(ctx)
	var err error
	clusterConfig := key.(*lcv1alpha1apis.LogicalCluster)
	_, err = c.clusterclientset.
		LogicalclusterV1alpha1().
		LogicalClusters(clusterConfig.Spec.ClusterProviderDesc).
		Update(ctx, clusterConfig, v1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to update cluster status.")
		return err
	}
	return err
}

func (c *Controller) processDelete(ctx context.Context, key any) error {
	logger := klog.FromContext(ctx)
	var err error

	var opts cluster.Options
	opts.KubeconfigPath = *c.kubeconfig
	delClusterConfig := key.(*lcv1alpha1apis.LogicalCluster)
	clusterName := delClusterConfig.Spec.ClusterName

	providerInfo, err := c.clusterclientset.LogicalclusterV1alpha1().ClusterProviderDescs().Get(ctx, delClusterConfig.Spec.ClusterProviderDesc, v1.GetOptions{})
	if err != nil {
		logger.Error(err, "failed to get provider resource.")
		return err
	}

	provider := clusterproviderclient.GetProviderClient(providerInfo.Spec.ProviderType, delClusterConfig.Spec.ClusterProviderDesc)
	err = provider.Delete(ctx, clusterName, opts)
	if err != nil {
		logger.Error(err, "failed to delete cluster")
		return err
	}

	return err
}

func (c *Controller) process(ctx context.Context, item queueItem) error {
	// Process the object
	logger := klog.FromContext(ctx)
	var err error

	op, key := item.op, item.key
	switch op {
	case opAdd:
		err = c.processAdd(ctx, key)
		if err != nil {
			logger.Info("Error:", err)
		}
	case opUpdate:
		err = c.processUpdate(ctx, key)
		if err != nil {
			logger.Info("Error:", err)
		}
	case opDelete:
		err = c.processDelete(ctx, key)
		if err != nil {
			logger.Info("Error:", err)
		}
	}
	return err
}
