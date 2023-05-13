package main

import (
	"context"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	log "k8s.io/klog/v2"
	kind "sigs.k8s.io/kind/pkg/cluster"

	"github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	ksclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
	"github.com/kcp-dev/edge-mc/pkg/mcclient"
)

// This is an example for cluster-aware client
// To run this example you need to create 2 regular kind clusters:
// kind create cluster -n cluster1
// kind create cluster -n management-cluster
// kubectl create -f config/crds/edge.kcp.io_logicalclusters.yaml
// go run main.go
func main() {
	ctx := context.Background()

	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	managementClusterConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		log.Fatalf("failed to build management cluster config: %v", err)
	}
	managementksClientset, err := ksclientset.NewForConfig(managementClusterConfig)
	if err != nil {
		log.Fatalf("failed to build management cluster clientset: %v", err)
	}

	createClusterObjectKS(ctx, managementksClientset, "cluster1")

	mcclient, err := mcclient.NewMultiCluster(ctx, managementClusterConfig)
	if err != nil {
		log.Fatalf("get client failed: %v", err)
	}

	log.Info("----------  List EdgePlacements in cluster1")
	list, _ := mcclient.Cluster("cluster1").KS().EdgeV1alpha1().EdgePlacements().List(ctx, metav1.ListOptions{})
	for _, ep := range list.Items {
		log.Infof("Cluster: cluster1 ; ObjName: %s", ep.Name)
	}
	log.Info("----------  List configmaps in cluster1")
	listm, _ := mcclient.Cluster("cluster1").Kube().CoreV1().ConfigMaps(metav1.NamespaceDefault).List(ctx, metav1.ListOptions{})
	for _, ep := range listm.Items {
		log.Infof("Cluster: cluster1 ; ObjName: %s", ep.Name)
	}

	updateClusterObjectKS(ctx, managementksClientset, "cluster1")
	deleteClusterObjectKS(ctx, managementksClientset, "cluster1")
}

func createClusterObjectKS(ctx context.Context, clientset ksclientset.Interface, cluster string) {
	provider := kind.NewProvider()
	kubeconfig, err := provider.KubeConfig(cluster, false)
	if err != nil {
		log.Error("failed to get config from kind provider")
		return
	}

	cm := v1alpha1.LogicalCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "LogicalCluster",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cluster,
		},
		Status: v1alpha1.LogicalClusterStatus{
			Phase:         v1alpha1.LogicalClusterPhaseReady,
			ClusterConfig: kubeconfig,
		},
	}

	_, err = clientset.EdgeV1alpha1().LogicalClusters().Create(ctx, &cm, metav1.CreateOptions{})
	if err != nil {
		log.Errorf("failed to create LogicalCluster: %v", err)
	}
	log.Infof("------ created: %s", cluster)
}

func deleteClusterObjectKS(ctx context.Context, clientset ksclientset.Interface, cluster string) {
	err := clientset.EdgeV1alpha1().LogicalClusters().Delete(ctx, cluster, metav1.DeleteOptions{})
	if err != nil {
		log.Errorf("failed to delete LogicalCluster: %v", err)
	}
	log.Infof("------ deleted: %s", cluster)
}

func updateClusterObjectKS(ctx context.Context, clientset ksclientset.Interface, cluster string) {
	clusterConfig, err := clientset.EdgeV1alpha1().LogicalClusters().Get(ctx, cluster, metav1.GetOptions{})
	if err != nil {
		log.Errorf("failed to get LogicalCluster: %v", err)
	}

	clusterConfig.Status.Phase = v1alpha1.LogicalClusterPhaseNotReady
	_, err = clientset.EdgeV1alpha1().LogicalClusters().Update(ctx, clusterConfig, metav1.UpdateOptions{})
	if err != nil {
		log.Errorf("failed to update LogicalCluster: %v", err)
	}
	log.Infof("------ updated: %s", cluster)
}
