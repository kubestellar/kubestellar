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

	clientopts "github.com/kubestellar/kubestellar/pkg/client-options"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	spaceclientset "github.com/kubestellar/kubestellar/space-framework/pkg/client/clientset/versioned"
	spaceinformers "github.com/kubestellar/kubestellar/space-framework/pkg/client/informers/externalversions"
	spaceclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
	spacemanager "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager"
)

func main() {
	resyncPeriod := time.Duration(0)
	var concurrency int = 4
	espwSpace := "espw"
	spaceProvider := "default"
	serverBindAddress := ":10203"
	fs := pflag.NewFlagSet("mailbox-controller", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	fs.Var(&utilflag.IPPortVar{Val: &serverBindAddress}, "server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")

	fs.IntVar(&concurrency, "concurrency", concurrency, "number of syncs to run in parallel")
	fs.StringVar(&espwSpace, "espw-space", espwSpace, "the name of the KubeStellar service provider space")
	fs.StringVar(&spaceProvider, "space-provider", spaceProvider, "the name of the KubeStellar space provider")

	spaceMgtOpts := clientopts.NewClientOpts("space-mgt", "access to space management workspace")
	spaceMgtOpts.AddFlags(fs)

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
	spaceManagementConfig, err := spaceMgtOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to create space management config from flags")
		os.Exit(3)
	}
	// create space-aware client
	spaceclient, err := spaceclient.NewMultiSpace(ctx, spaceManagementConfig)
	if err != nil {
		logger.Error(err, "failed to create space-aware client")
		os.Exit(4)
	}
	spaceProviderNs := spacemanager.ProviderNS(spaceProvider)

	espwRestConfig, err := spaceclient.ConfigForSpace(espwSpace, spaceProviderNs)
	if err != nil {
		logger.Error(err, "failed to fetch space config", "spacename", espwSpace)
		os.Exit(5)
	}

	edgeClientset, err := edgeclientset.NewForConfig(espwRestConfig)
	if err != nil {
		logger.Error(err, "failed to create clientset for service provider space")
		os.Exit(6)
	}
	edgeSharedInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(edgeClientset, resyncPeriod)
	syncTargetPreInformer := edgeSharedInformerFactory.Edge().V2alpha1().SyncTargets()

	managementClientset, err := spaceclientset.NewForConfig(spaceManagementConfig)
	if err != nil {
		logger.Error(err, "failed to create clientset for space management")
	}

	spaceInformerFactory := spaceinformers.NewSharedInformerFactory(managementClientset, resyncPeriod)
	spacePreInformer := spaceInformerFactory.Space().V1alpha1().Spaces()

	ctl := newMailboxController(ctx, syncTargetPreInformer, spacePreInformer,
		managementClientset, spaceProvider, spaceProviderNs,
	)

	doneCh := ctx.Done()
	edgeSharedInformerFactory.Start(doneCh)

	spaceInformerFactory.Start(doneCh)

	ctl.Run(concurrency)

	logger.Info("Time to stop")
}
