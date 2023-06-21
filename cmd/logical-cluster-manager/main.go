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
	providermanager "github.com/kcp-dev/edge-mc/pkg/provider-manager"
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

	doneCh := ctx.Done()

	clusterInformerFactory := edgeinformers.NewSharedScopedInformerFactory(clusterClientset, resyncPeriod, metav1.NamespaceAll)
	clusterInformer := clusterInformerFactory.Logicalcluster().V1alpha1().LogicalClusters().Informer()
	clusterInformerFactory.Start(doneCh)

	providerInformerFactory := edgeinformers.NewSharedScopedInformerFactory(clusterClientset, resyncPeriod, metav1.NamespaceAll)
	providerInformer := providerInformerFactory.Logicalcluster().V1alpha1().ClusterProviderDescs().Informer()
	providerInformerFactory.Start(doneCh)

	// TODO: clusterInformer is not the same as providerInformer
	clusterController := clustermanager.NewController(
		kubeconfig,
		ctx,
		clusterClientset,
		clusterInformer,
	)
	if err != nil {
		logger.Error(err, "failed to create logical cluster controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	logger.Info("about to create provider controller")
	providerController := providermanager.NewController(
		ctx,
		clusterClientset,
		providerInformer,
	)
	if err != nil {
		logger.Error(err, "failed to create provider controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	logger.Info("done creating provider controller")

	clusterController.Run(numThreads)
	go providerController.Run(numThreads)
	logger.Info("Time to stop")
}
