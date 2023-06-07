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
	"reflect"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	log "k8s.io/klog/v2"

	"github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	ksclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
	ksinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
	mcclientset "github.com/kcp-dev/edge-mc/pkg/mcclient/clientset"
	"github.com/kcp-dev/edge-mc/pkg/mcclient/listwatch"
)

type KubestellarClusterInterface interface {
	// Cluster returns clientset for given cluster.
	Cluster(name string) (client mcclientset.Interface)
	// ConfigForCluster returns rest config for given cluster.
	ConfigForCluster(name string) (*rest.Config, error)
	// CrossClusterListWatch returns cross-cluster ListWatch
	CrossClusterListWatch(gv schema.GroupVersion, resource string, namespace string, fieldSelector fields.Selector) *listwatch.CrossClusterListerWatcher
}

type multiClusterClient struct {
	configs    map[string]*rest.Config
	clientsets map[string]*mcclientset.Clientset
	lock       sync.Mutex
}

func (mcc *multiClusterClient) Cluster(name string) mcclientset.Interface {
	mcc.lock.Lock()
	defer mcc.lock.Unlock()
	clientset, ok := mcc.clientsets[name]
	if !ok {
		//TODO change to ClusterOrDie and panic? return an error?
		panic("invalid cluster name")
	}
	return clientset
}

func (mcc *multiClusterClient) ConfigForCluster(name string) (*rest.Config, error) {
	mcc.lock.Lock()
	defer mcc.lock.Unlock()
	if _, ok := mcc.configs[name]; !ok {
		//TODO get from server
		return nil, errors.New("failed to get config for cluster")
	}
	return mcc.configs[name], nil
}

// CrossClusterListWatch NOT implemented. WIP
func (mcc *multiClusterClient) CrossClusterListWatch(gv schema.GroupVersion, resource string, namespace string, fieldSelector fields.Selector) *listwatch.CrossClusterListerWatcher {
	optionsModifier := func(options *metav1.ListOptions) {
		options.FieldSelector = fieldSelector.String()
	}
	clusterLW := make(map[string]*cache.ListWatch)
	for cluster, config := range mcc.configs {
		clusterLW[cluster] = listwatch.ClusterListWatch(config, gv, resource, namespace, optionsModifier)
	}
	return listwatch.NewCrossClusterListerWatcher(clusterLW)
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

	client = &multiClusterClient{
		configs:    make(map[string]*rest.Config),
		clientsets: make(map[string]*mcclientset.Clientset),
		lock:       sync.Mutex{},
	}
	managerClientset, err := ksclientset.NewForConfig(managerConfig)
	if err != nil {
		return client, err
	}

	client.startClusterCollection(ctx, managerClientset)
	return client, nil
}

func (mcc *multiClusterClient) startClusterCollection(ctx context.Context, managerClientset *ksclientset.Clientset) {
	clusterInformerFactory := ksinformers.NewSharedScopedInformerFactory(managerClientset, 0, metav1.NamespaceAll)
	clusterInformer := clusterInformerFactory.Edge().V1alpha1().LogicalClusters().Informer()

	clusterInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			clusterInfo, ok := obj.(*v1alpha1.LogicalCluster)
			if !ok {
				log.Error("unexpected object type. expected LogicalCluster")
				return
			}
			go mcc.handleAdd(clusterInfo)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldInfo := oldObj.(*v1alpha1.LogicalCluster)
			newInfo := newObj.(*v1alpha1.LogicalCluster)
			if reflect.DeepEqual(oldInfo.Status, newInfo.Status) {
				return
			}
			go mcc.handleAdd(newInfo)
		},
		DeleteFunc: func(obj interface{}) {
			clusterInfo := obj.(*v1alpha1.LogicalCluster)
			go mcc.handleDelete(clusterInfo)
		},
	})

	clusterInformerFactory.Start(ctx.Done())
	cache.WaitForNamedCacheSync("management", ctx.Done(), clusterInformer.HasSynced)
}

func (mcc *multiClusterClient) handleAdd(clusterInfo *v1alpha1.LogicalCluster) {
	if clusterInfo.Status.Phase != v1alpha1.LogicalClusterPhaseReady {
		mcc.handleDelete(clusterInfo)
		return
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(clusterInfo.Status.ClusterConfig))
	if err != nil {
		log.Errorf("failed to get cluster config: %v", err)
		return
	}
	mcc.lock.Lock()
	defer mcc.lock.Unlock()
	mcc.configs[clusterInfo.Name] = config

	cs, err := mcclientset.NewForConfig(config)
	if err != nil {
		//ES: do we want to try again ? how ?
		// should we delete  the config as well ?
		log.Errorf("failed to create clientset for cluster config: %v", err)
		return
	}
	mcc.clientsets[clusterInfo.Name] = cs
}

func (mcc *multiClusterClient) handleDelete(clusterInfo *v1alpha1.LogicalCluster) {
	mcc.lock.Lock()
	defer mcc.lock.Unlock()
	if _, ok := mcc.configs[clusterInfo.Name]; !ok {
		return
	}
	delete(mcc.configs, clusterInfo.Name)
	delete(mcc.clientsets, clusterInfo.Name)
}
