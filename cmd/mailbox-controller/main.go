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

// Import of k8s.io/client-go/plugin/pkg/client/auth ensures
// that all in-tree Kubernetes client auth plugins
// (e.g. Azure, GCP, OIDC, etc.)  are available.
//
// Import of k8s.io/component-base/metrics/prometheus/clientgo
// makes the k8s client library produce Prometheus metrics.

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"

	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	kcpscopedclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	"github.com/kcp-dev/logicalcluster/v3"

	clientopts "github.com/kubestellar/kubestellar/pkg/client-options"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
)

func main() {
	resyncPeriod := time.Duration(0)
	var concurrency int = 4
	serverBindAddress := ":10203"
	espwPath := logicalcluster.Name("root").Path().Join("espw").String()
	fs := pflag.NewFlagSet("mailbox-controller", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	fs.Var(&utilflag.IPPortVar{Val: &serverBindAddress}, "server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")

	fs.IntVar(&concurrency, "concurrency", concurrency, "number of syncs to run in parallel")
	fs.StringVar(&espwPath, "espw-path", espwPath, "the pathname of the edge service provider workspace")

	rootClientOpts := clientopts.NewClientOpts("root", "access to the root workspace")
	rootClientOpts.SetDefaultCurrentContext("root")
	rootClientOpts.AddFlags(fs)
	espwClientOpts := clientopts.NewClientOpts("espw", "access to the edge service provider workspace")
	espwClientOpts.AddFlags(fs)

	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	fs.VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info("Command line flag", flg.Name, flg.Value)
	})

	mymux := mux.NewPathRecorderMux("mailbox-controller")
	mymux.Handle("/metrics", legacyregistry.Handler())
	routes.Profiling{}.Install(mymux)
	go func() {
		err := http.ListenAndServe(serverBindAddress, mymux)
		if err != nil {
			logger.Error(err, "Failure in web serving")
			panic(err)
		}
	}()

	// create edgeSharedInformerFactory
	espwRestConfig, err := espwClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to create config from flags")
		os.Exit(3)
	}

	edgeClientset, err := edgeclientset.NewForConfig(espwRestConfig)
	if err != nil {
		logger.Error(err, "failed to create clientset for view of edge exports")
		os.Exit(6)
	}

	edgeSharedInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(edgeClientset, resyncPeriod)
	syncTargetClusterPreInformer := edgeSharedInformerFactory.Edge().V2alpha1().SyncTargets()

	rootRestConfig, err := rootClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make root config")
		os.Exit(8)
	}
	rootRestConfig.UserAgent = "mailbox-controller"

	workspaceScopedClientset, err := kcpscopedclientset.NewForConfig(rootRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for workspaces")
	}

	workspaceScopedInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(workspaceScopedClientset, resyncPeriod)
	workspaceScopedPreInformer := workspaceScopedInformerFactory.Tenancy().V1alpha1().Workspaces()

	ctl := newMailboxController(ctx, espwPath, syncTargetClusterPreInformer, workspaceScopedPreInformer,
		workspaceScopedClientset.TenancyV1alpha1().Workspaces(),
	)

	doneCh := ctx.Done()
	edgeSharedInformerFactory.Start(doneCh)

	workspaceScopedInformerFactory.Start(doneCh)

	ctl.Run(concurrency)

	logger.Info("Time to stop")
}
