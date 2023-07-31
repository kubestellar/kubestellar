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
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/klog/v2"

	ksclient "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	ksinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	spacemanager "github.com/kubestellar/kubestellar/pkg/space-manager"
)

var (
	resyncPeriod = 30 * time.Second
	numThreads   = 2
)

const managerSpaceContext string = "kind-mgt"

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
		&clientcmd.ConfigOverrides{CurrentContext: managerSpaceContext}).ClientConfig()
	if err != nil {
		logger.Error(err, "Error")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	clientset, err := ksclient.NewForConfig(config)
	if err != nil {
		logger.Error(err, "failed to create clientset for controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	k8sClientset, err := kubeclient.NewForConfig(config)
	if err != nil {
		logger.Error(err, "failed to create k8s clientset for controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	informerFactory := ksinformers.NewSharedScopedInformerFactory(clientset, resyncPeriod, metav1.NamespaceAll)
	informerSpace := informerFactory.Space().V1alpha1().Spaces().Informer()
	informerProvider := informerFactory.Space().V1alpha1().SpaceProviderDescs().Informer()

	controller := spacemanager.NewController(
		ctx,
		clientset,
		k8sClientset,
		informerSpace,
		informerProvider,
	)

	doneCh := ctx.Done()
	informerFactory.Start(doneCh)
	informerFactory.WaitForCacheSync(doneCh)

	controller.Run(numThreads)
}
