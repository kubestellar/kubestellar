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

	edgeclient "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	clustermanager "github.com/kubestellar/kubestellar/pkg/logical-cluster-manager"
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

	clusterController.Run(numThreads)
	logger.Info("Time to stop")
}
