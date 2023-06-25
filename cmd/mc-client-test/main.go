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

package main

import (
	"context"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"
	kind "sigs.k8s.io/kind/pkg/cluster"

	lcv1 "github.com/kubestellar/kubestellar/pkg/apis/logicalcluster/v1alpha1"
	ksclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	"github.com/kubestellar/kubestellar/pkg/mcclient"
)

const defaultProviderNs = "lcprovider-default"

// This is an example for cluster-aware client
// To run this example you need to create 2 regular kind clusters:
// kind create cluster -n cluster1
// kind create cluster -n management-cluster
// kubectl create -f config/crds/logicalcluster.kubestellar.io_logicalclusters.yaml
// kubectl create ns lcprovider-default
// go run main.go
func main() {
	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	managementClusterConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		logger.Error(err, "failed to build management cluster config")
		panic(err)
	}
	managementksClientset, err := ksclientset.NewForConfig(managementClusterConfig)
	if err != nil {
		logger.Error(err, "failed to build management cluster clientset")
		panic(err)
	}

	createClusterObjectKS(ctx, managementksClientset, "cluster1")

	mcclient, err := mcclient.NewMultiCluster(ctx, managementClusterConfig)
	if err != nil {
		logger.Error(err, "get client failed")
		panic(err)
	}

	logger.Info("List EdgePlacements in cluster1")
	list, _ := mcclient.Cluster("cluster1").KS().EdgeV1alpha1().EdgePlacements().List(ctx, metav1.ListOptions{})
	for _, ep := range list.Items {
		logger.Info("Cluster: cluster1", "edgePlacement", ep.Name)
	}
	logger.Info("List configmaps in cluster1")
	listm, _ := mcclient.Cluster("cluster1").Kube().CoreV1().ConfigMaps(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{})
	for _, cm := range listm.Items {
		logger.Info("Cluster: cluster1", "configMap", cm.Name)
	}

	updateClusterObjectKS(ctx, managementksClientset, "cluster1")
	deleteClusterObjectKS(ctx, managementksClientset, "cluster1")
}

func createClusterObjectKS(ctx context.Context, clientset ksclientset.Interface, cluster string) {
	logger := klog.FromContext(ctx)
	provider := kind.NewProvider()
	kubeconfig, err := provider.KubeConfig(cluster, false)
	if err != nil {
		logger.Error(err, "failed to get config from kind provider")
		return
	}

	cm := lcv1.LogicalCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LogicalCluster",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster,
		},
		Status: lcv1.LogicalClusterStatus{
			Phase:         lcv1.LogicalClusterPhaseReady,
			ClusterConfig: kubeconfig,
		},
	}

	_, err = clientset.LogicalclusterV1alpha1().LogicalClusters(defaultProviderNs).Create(ctx, &cm, metav1.CreateOptions{})
	if err != nil {
		logger.Error(err, "failed to create LogicalCluster")
		return
	}
	logger.Info("LogicalCluster created", "cluster", cluster)
}

func deleteClusterObjectKS(ctx context.Context, clientset ksclientset.Interface, cluster string) {
	logger := klog.FromContext(ctx)
	err := clientset.LogicalclusterV1alpha1().LogicalClusters(defaultProviderNs).Delete(ctx, cluster, metav1.DeleteOptions{})
	if err != nil {
		logger.Error(err, "failed to delete LogicalCluster")
		return
	}
	logger.Info("LogicalCluster deleted", "cluster", cluster)
}

func updateClusterObjectKS(ctx context.Context, clientset ksclientset.Interface, cluster string) {
	logger := klog.FromContext(ctx)
	clusterConfig, err := clientset.LogicalclusterV1alpha1().LogicalClusters(defaultProviderNs).Get(ctx, cluster, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "failed to get LogicalCluster")
		return
	}

	clusterConfig.Status.Phase = lcv1.LogicalClusterPhaseNotReady
	_, err = clientset.LogicalclusterV1alpha1().LogicalClusters(defaultProviderNs).Update(ctx, clusterConfig, metav1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to update LogicalCluster")
		return
	}
	logger.Info("LogicalCluster updated", "cluster", cluster)
}
