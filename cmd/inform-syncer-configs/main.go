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

// Import of k8s.io/client-go/plugin/pkg/client/auth ensures
// that all in-tree Kubernetes client auth plugins
// (e.g. Azure, GCP, OIDC, etc.)  are available.

import (
	"context"
	"flag"
	"os"
	"sync"

	"github.com/spf13/pflag"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	upstreamcache "k8s.io/client-go/tools/cache"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"

	kcpscopedclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	clientopts "github.com/kcp-dev/edge-mc/pkg/client-options"
	emcclusterclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgescopedclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/typed/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/mailboxwatch"
)

func main() {
	fs := pflag.NewFlagSet("inform-syncer-configs", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)

	espwClientOpts := clientopts.NewClientOpts("espw", "access to the edge service provider workspace")
	espwClientOpts.AddFlags(fs)

	allClientOpts := clientopts.NewClientOpts("all", "access to the SyncerConfig objects in all clusters")
	allClientOpts.SetDefaultCurrentContext("system:admin")
	allClientOpts.AddFlags(fs)

	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	espwClientConfig, err := espwClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make ESPW client config")
		os.Exit(2)
	}
	espwClientConfig.UserAgent = "inform-syncer-configs"

	espwClient := kcpscopedclient.NewForConfigOrDie(espwClientConfig)

	allClientConfig, err := allClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make all-cluster client config")
		os.Exit(2)
	}
	allClientConfig.UserAgent = "inform-syncer-configs"

	clusterClientset := emcclusterclientset.NewForConfigOrDie(allClientConfig)

	clusterEdge := clusterClientset.EdgeV1alpha1()
	clusterEdgeSyncfgs := clusterEdge.SyncerConfigs()

	scGV := edgeapi.SchemeGroupVersion
	kind := "SyncerConfig"
	// scGVK := scGV.WithKind(kind)
	sclGVK := scGV.WithKind(kind + "List")

	espwInformerFactory := kcpinformers.NewSharedScopedInformerFactory(espwClient, 0, "")
	mbPreInformer := espwInformerFactory.Tenancy().V1alpha1().Workspaces()

	informer := mailboxwatch.NewSharedInformer[edgescopedclient.SyncerConfigInterface, *edgeapi.SyncerConfigList](ctx, sclGVK, mbPreInformer, clusterEdgeSyncfgs, &edgeapi.SyncerConfig{}, 0, upstreamcache.Indexers{})

	informer.AddEventHandler(upstreamcache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { log(logger, "add", obj) },
		UpdateFunc: func(oldObj, newObj any) { log(logger, "update", newObj) },
		DeleteFunc: func(obj any) { log(logger, "delete", obj) },
	})
	espwInformerFactory.Start(ctx.Done())
	upstreamcache.WaitForCacheSync(ctx.Done(), mbPreInformer.Informer().HasSynced)
	go informer.Run(ctx.Done())
	logger.Info("Running")
	<-ctx.Done()
}

var logmu sync.Mutex

func log(logger klog.Logger, action string, obj any) {
	logmu.Lock()
	defer logmu.Unlock()
	logger.Info("Notified", "action", action, "object", obj)
}
