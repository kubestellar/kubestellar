package main

import (
	"context"
	"flag"
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	edgeclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
	edgeinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
	clustermanager "github.com/kcp-dev/edge-mc/pkg/logical-cluster-manager"
)

var (
	resyncPeriod = 4 * time.Second
	numThreads   = 2
)

const nameLogicalClusterManagerCluster string = "kind-fleet-test1"

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := klog.FromContext(ctx)

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: *kubeconfig},
		&clientcmd.ConfigOverrides{CurrentContext: nameLogicalClusterManagerCluster}).ClientConfig()
	if err != nil {
		logger.Error(err, "Error")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	clusterClientset, err := edgeclient.NewForConfig(config)
	if err != nil {
		logger.Error(err, "failed to create edge clientset for controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	clusterInformerFactory := edgeinformers.NewSharedScopedInformerFactory(clusterClientset, resyncPeriod, metav1.NamespaceAll)
	clusterInformer := clusterInformerFactory.Edge().V1alpha1().LogicalClusters()

	doneCh := ctx.Done()
	clusterController := clustermanager.NewController(
		kubeconfig,
		ctx,
		clusterClientset,
		clusterInformer,
	)
	if err != nil {
		logger.Error(err, "failed to create controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	clusterInformerFactory.Start(doneCh)
	clusterController.Run(numThreads)
	logger.Info("Time to stop")
}
