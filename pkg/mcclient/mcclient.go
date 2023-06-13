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

// kubestellar cluster-aware client impl

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	lcv1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
	ksclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
	ksinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
	mcclientset "github.com/kcp-dev/edge-mc/pkg/mcclient/clientset"
)

type KubestellarClusterInterface interface {
	// Cluster returns clientset for given cluster.
	Cluster(name string) (client mcclientset.Interface)
	// ConfigForCluster returns rest config for given cluster.
	ConfigForCluster(name string) (*rest.Config, error)
}

type multiClusterClient struct {
	ctx              context.Context
	configs          map[string]*rest.Config
	clientsets       map[string]*mcclientset.Clientset
	managerClientset *ksclientset.Clientset
	lock             sync.Mutex
}

func (mcc *multiClusterClient) Cluster(name string) mcclientset.Interface {
	var err error
	mcc.lock.Lock()
	defer mcc.lock.Unlock()
	clientset, ok := mcc.clientsets[name]
	if !ok {
		// Try to get LogicalCluster from API server.
		if clientset, err = mcc.getFromServer(name); err != nil {
			panic(fmt.Sprintf("invalid cluster name: %s. error: %v", name, err))
		}
	}
	return clientset
}

func (mcc *multiClusterClient) ConfigForCluster(name string) (*rest.Config, error) {
	mcc.lock.Lock()
	defer mcc.lock.Unlock()
	if _, ok := mcc.configs[name]; !ok {
		return nil, fmt.Errorf("failed to get config for cluster: %s", name)
	}
	return mcc.configs[name], nil
}

var client *multiClusterClient
var clientLock = &sync.Mutex{}

// NewMultiCluster creates new multi-cluster client and starts collecting cluster configs
func NewMultiCluster(ctx context.Context, managerConfig *rest.Config) (KubestellarClusterInterface, error) {
	clientLock.Lock()
	defer clientLock.Unlock()

	if client != nil {
		return client, nil
	}

	managerClientset, err := ksclientset.NewForConfig(managerConfig)
	if err != nil {
		return client, err
	}

	client = &multiClusterClient{
		ctx:              ctx,
		configs:          make(map[string]*rest.Config),
		clientsets:       make(map[string]*mcclientset.Clientset),
		managerClientset: managerClientset,
		lock:             sync.Mutex{},
	}

	client.startClusterCollection(ctx, managerClientset)
	return client, nil
}

func (mcc *multiClusterClient) startClusterCollection(ctx context.Context, managerClientset *ksclientset.Clientset) {
	numThreads := 2
	resyncPeriod := time.Duration(0)

	clusterInformerFactory := ksinformers.NewSharedScopedInformerFactory(managerClientset, resyncPeriod, metav1.NamespaceAll)
	clusterInformer := clusterInformerFactory.Logicalcluster().V1alpha1().LogicalClusters().Informer()

	clusterInformerFactory.Start(ctx.Done())
	cache.WaitForNamedCacheSync("logicalclusters-management", ctx.Done(), clusterInformer.HasSynced)

	logicalClusterController := newController(ctx, clusterInformer, mcc)
	go logicalClusterController.Run(numThreads)
}

// getFromServer will query API server for specific LogicalCluster and cache it if it exists and ready.
// getFromServer caller function should acquire the mcc lock.
func (mcc *multiClusterClient) getFromServer(name string) (*mcclientset.Clientset, error) {
	cluster, err := mcc.managerClientset.LogicalclusterV1alpha1().LogicalClusters().Get(mcc.ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if cluster.Status.Phase != lcv1alpha1.LogicalClusterPhaseReady {
		return nil, errors.New("cluster is not ready")
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(cluster.Status.ClusterConfig))
	if err != nil {
		return nil, err
	}
	cs, err := mcclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	// Calling function should acquire the mcc lock
	mcc.configs[cluster.Name] = config
	mcc.clientsets[cluster.Name] = cs

	return cs, nil
}
