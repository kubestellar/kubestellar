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

package cmd

import (
	"context"
	"flag"
	"fmt"
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

	transportoptions "github.com/kubestellar/kubestellar/cmd/transport/options"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
	"github.com/kubestellar/kubestellar/pkg/transport"
	spaceclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
	spacemanager "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager"
)

const (
	transportControllerName = "transport-controller"
	defaultResyncPeriod     = time.Duration(0)
)

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

	options := transportoptions.NewOptions()
	fs := pflag.NewFlagSet(transportControllerName, pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	options.AddFlags(fs)
	fs.Parse(os.Args[1:])

	fs.VisitAll(func(flg *pflag.Flag) {
		logger.Info(fmt.Sprintf("Command line flag '%s' - value '%s'", flg.Name, flg.Value)) // log all arguments
	})

	startMetricsServer(options.MetricsServerBindAddress, logger)

	// create space-aware client
	spaceManagementConfig, err := options.SpaceMgtClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to create space management API client config from flags")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	spaceclient, err := spaceclient.NewMultiSpace(ctx, spaceManagementConfig, options.ExternalAccess)
	if err != nil {
		logger.Error(err, "Failed to create space-aware client")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	spaceProviderNs := spacemanager.ProviderNS(options.SpaceProviderObjectName)

	kubestellarCoreSpaceConfig, err := spaceclient.ConfigForSpace(options.KubestellarCoreSpaceName, spaceProviderNs)
	if err != nil {
		logger.Error(err, "Failed to construct kubestellar core space config", "spacename", options.KubestellarCoreSpaceName)
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	kubestellarCoreSpaceConfig.UserAgent = transportControllerName

	transportSpaceConfig, err := spaceclient.ConfigForSpace(options.TransportSpaceName, spaceProviderNs)
	if err != nil {
		logger.Error(err, "Failed to construct transport space config", "spacename", options.TransportSpaceName)
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

	if err := transportController.Run(ctx, options.Concurrency); err != nil {
		logger.Error(err, "failed to run transport controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	logger.Info("transport controller stopped")
}
