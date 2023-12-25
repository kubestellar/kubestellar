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
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"

	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	clientopts "github.com/kubestellar/kubestellar/pkg/client-options"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
	"github.com/kubestellar/kubestellar/pkg/transport"
	spaceclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
	spacemanager "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager"
)

const (
	transportControllerName         = "transport-controller"
	defaultResyncPeriod             = time.Duration(0)
	defaultConcurrency              = 4
	defaultMetricsServerBindAddress = ":10472"
	defaultKubestellarCoreSpaceName = "espw"
	defaultTransportSpaceName       = "transport"
	defaultSpaceProviderObjectName  = "default"
	defaultExternalAccess           = false
)

func parseArguments(fs *pflag.FlagSet) (concurrency int, metricsServerBindAddress string, kcsName string, transportSpaceName string,
	externalAccess bool, spaceProvider string, spaceMgtOpts *clientopts.ClientOpts) {
	metricsServerBindAddress = defaultMetricsServerBindAddress
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	fs.IntVar(&concurrency, "concurrency", defaultConcurrency, "number of concurrent workers to run in parallel")
	fs.Var(&utilflag.IPPortVar{Val: &metricsServerBindAddress}, "metrics-server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")
	fs.StringVar(&kcsName, "core-space", defaultKubestellarCoreSpaceName, "the name of the KubeStellar core space")
	fs.StringVar(&transportSpaceName, "transport-space", defaultTransportSpaceName, "the name of the transport space")
	fs.BoolVar(&externalAccess, "external-access", defaultExternalAccess, "the access to the spaces. True when the space-provider is hosted in a space while the controller is running outside of that space")
	fs.StringVar(&spaceProvider, "space-provider", defaultSpaceProviderObjectName, "the name of the KubeStellar space provider")
	spaceMgtOpts = clientopts.NewClientOpts("space-mgt", "access to the space reference space")
	spaceMgtOpts.AddFlags(fs)

	fs.Parse(os.Args[1:])

	return
}

func startMetricsServer(serverBindAddress string, logger logr.Logger) {
	mymux := mux.NewPathRecorderMux(transportControllerName)
	mymux.Handle("/metrics", legacyregistry.Handler())
	routes.Profiling{}.Install(mymux)
	go func() {
		err := http.ListenAndServe(serverBindAddress, mymux)
		if err != nil {
			logger.Error(err, "Failure in web serving")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
	}()
}

func Run(transportImplementation transport.Transport) {
	logger := klog.Background().WithName(transportControllerName)
	ctx := klog.NewContext(context.Background(), logger)

	fs := pflag.NewFlagSet(transportControllerName, pflag.ExitOnError)
	concurrency, metricsServerBindAddress, kcsName, transportSpaceName, externalAccess, spaceProvider, spaceMgtOpts := parseArguments(fs)
	fs.VisitAll(func(flg *pflag.Flag) {
		logger.Info("Command line flag", flg.Name, flg.Value) // log all arguments
	})

	startMetricsServer(metricsServerBindAddress, logger)

	// create space-aware client
	spaceManagementConfig, err := spaceMgtOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to create space management API client config from flags")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	spaceclient, err := spaceclient.NewMultiSpace(ctx, spaceManagementConfig, externalAccess)
	if err != nil {
		logger.Error(err, "Failed to create space-aware client")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	spaceProviderNs := spacemanager.ProviderNS(spaceProvider)

	kubestellarCoreSpaceConfig, err := spaceclient.ConfigForSpace(kcsName, spaceProviderNs)
	if err != nil {
		logger.Error(err, "Failed to construct kubestellar core space config", "spacename", kcsName)
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	kubestellarCoreSpaceConfig.UserAgent = transportControllerName

	transportSpaceConfig, err := spaceclient.ConfigForSpace(transportSpaceName, spaceProviderNs)
	if err != nil {
		logger.Error(err, "Failed to construct transport space config", "spacename", transportSpaceName)
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	edgeClientset, err := edgeclientset.NewForConfig(kubestellarCoreSpaceConfig)
	if err != nil {
		logger.Error(err, "Failed to create edge clientset for KubeStellar Core Space")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	edgeSharedInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(edgeClientset, defaultResyncPeriod)
	edgePlacementDecisionInformer := edgeSharedInformerFactory.Edge().V2alpha1().EdgePlacementDecisions()

	kubeClient, err := kubernetes.NewForConfig(kubestellarCoreSpaceConfig)
	if err != nil {
		logger.Error(err, "Failed to create k8s clientset for KubeStellar Core Space")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	kbSpaceRelation := kbuser.NewKubeBindSpaceRelation(ctx, kubeClient)
	cache.WaitForCacheSync(ctx.Done(), kbSpaceRelation.InformerSynced)

	transportController, err := transport.NewTransportController(ctx, edgePlacementDecisionInformer, transportImplementation, kbSpaceRelation, spaceclient, spaceProviderNs, transportSpaceConfig)

	// notice that there is no need to run Start method in a separate goroutine.
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	edgeSharedInformerFactory.Start(ctx.Done())

	if err := transportController.Run(ctx, concurrency); err != nil {
		logger.Error(err, "failed to run transport controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	logger.Info("transport controller stopped")
}
