package main

import (
	"context"
	"os"
	"time"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	kcpkubernetesinformers "github.com/kcp-dev/client-go/informers"
	kcpkubernetesclientset "github.com/kcp-dev/client-go/kubernetes"
	kcpclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	"github.com/kcp-dev/logicalcluster/v2"

	edgeindexers "github.com/kcp-dev/edge-mc/pkg/indexers"
	edgeplacement "github.com/kcp-dev/edge-mc/pkg/reconciler/scheduling/placement"
)

func main() {
	const resyncPeriod = 10 * time.Hour

	ctx := context.Background()
	logger := klog.FromContext(ctx)

	cfg, err := config.GetConfigWithContext("system:admin")
	if err != nil {
		logger.Error(err, "failed to get config, is KUBECONFIG pointing to kcp server if running out of cluster?")
		os.Exit(1)
	}

	// create kubeSharedInformerFactory
	kubernetesConfig := rest.CopyConfig(cfg)
	kubeClientset, err := kcpkubernetesclientset.NewForConfig(kubernetesConfig)
	if err != nil {
		logger.Error(err, "failed to create kube clientset")
		os.Exit(1)
	}
	kubeSharedInformerFactory := kcpkubernetesinformers.NewSharedInformerFactory(kubeClientset, 10*time.Minute)

	// create kcpSharedInformerFactory
	kcpConfig := rest.CopyConfig(cfg)
	kcpClusterClient, err := kcpclient.NewClusterForConfig(kcpConfig)
	if err != nil {
		logger.Error(err, "failed to create kcp cluster client")
		os.Exit(1)
	}
	kcpSharedInformerFactory := kcpinformers.NewSharedInformerFactoryWithOptions(
		kcpClusterClient.Cluster(logicalcluster.Wildcard),
		resyncPeriod,
		kcpinformers.WithExtraClusterScopedIndexers(edgeindexers.ClusterScoped()),
		kcpinformers.WithExtraNamespaceScopedIndexers(edgeindexers.NamespaceScoped()),
	)

	// create edge placement controller
	controllerConfig := rest.CopyConfig(cfg)
	kcpClientset, err := kcpclient.NewForConfig(controllerConfig)
	if err != nil {
		logger.Error(err, "failed to create kcp clientset")
		os.Exit(1)
	}
	c, err := edgeplacement.NewController(
		kcpClientset,
		kubeSharedInformerFactory.Core().V1().Namespaces(),
		kcpSharedInformerFactory.Scheduling().V1alpha1().Locations(),
		kcpSharedInformerFactory.Scheduling().V1alpha1().Placements(),
	)
	if err != nil {
		logger.Error(err, "Failed to create controller")
		os.Exit(1)
	}

	// run edge placement controller
	kubeSharedInformerFactory.Start(ctx.Done())
	kubeSharedInformerFactory.WaitForCacheSync(ctx.Done())
	kcpSharedInformerFactory.Start(ctx.Done())
	kcpSharedInformerFactory.WaitForCacheSync(ctx.Done())
	c.Start(ctx, 1)
}
