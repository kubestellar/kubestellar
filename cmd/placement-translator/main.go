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
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	ksclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	emcinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
	"github.com/kubestellar/kubestellar/pkg/placement"
	spaceclientset "github.com/kubestellar/kubestellar/space-framework/pkg/client/clientset/versioned"
	spaceinformers "github.com/kubestellar/kubestellar/space-framework/pkg/client/informers/externalversions"
	spaceclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
	spacemanager "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager"
)

type ClientOpts struct {
	which        string
	description  string
	loadingRules *clientcmd.ClientConfigLoadingRules
	overrides    clientcmd.ConfigOverrides
}

func NewClientOpts(which, description string) *ClientOpts {
	return &ClientOpts{
		which:        which,
		description:  description,
		loadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
		overrides:    clientcmd.ConfigOverrides{},
	}
}

func (opts *ClientOpts) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.loadingRules.ExplicitPath, opts.which+"-kubeconfig", opts.loadingRules.ExplicitPath, "Path to the kubeconfig file to use for "+opts.description)
	flags.StringVar(&opts.overrides.CurrentContext, opts.which+"-context", opts.overrides.CurrentContext, "The name of the kubeconfig context to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.AuthInfo, opts.which+"-user", opts.overrides.Context.AuthInfo, "The name of the kubeconfig user to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.Cluster, opts.which+"-cluster", opts.overrides.Context.Cluster, "The name of the kubeconfig cluster to use for "+opts.description)

}

func (opts *ClientOpts) ToRESTConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(opts.loadingRules, &opts.overrides)
	return clientConfig.ClientConfig()
}

func main() {
	resyncPeriod := time.Duration(0)
	var concurrency int = 4
	serverBindAddress := ":10204"
	coreSpace := "espw"
	spaceProvider := "default"
	fs := pflag.NewFlagSet("placement-translator", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	fs.Var(&utilflag.IPPortVar{Val: &serverBindAddress}, "server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")
	fs.IntVar(&concurrency, "concurrency", concurrency, "number of syncs to run in parallel")
	fs.StringVar(&coreSpace, "core-space", coreSpace, "the name of the KubeStellar core space")
	fs.StringVar(&spaceProvider, "space-provider", spaceProvider, "the name of the KubeStellar space provider")

	spaceMgtOpts := NewClientOpts("space-mgt", "access to space management")
	spaceMgtOpts.AddFlags(fs)
	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	fs.VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info("Command line flag", flg.Name, flg.Value)
	})

	mymux := mux.NewPathRecorderMux("placement-translator")
	mymux.Handle("/metrics", legacyregistry.Handler())
	routes.Profiling{}.Install(mymux)
	go func() {
		err := http.ListenAndServe(serverBindAddress, mymux)
		if err != nil {
			logger.Error(err, "Failure in web serving")
			panic(err)
		}
	}()

	spaceManagementConfig, err := spaceMgtOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to create space management config from flags")
		os.Exit(3)
	}
	spaceclient, err := spaceclient.NewMultiSpace(ctx, spaceManagementConfig)
	if err != nil {
		logger.Error(err, "Failed to create space-aware client")
		os.Exit(4)
	}
	spaceProviderNs := spacemanager.ProviderNS(spaceProvider)

	coreRestConfig, err := spaceclient.ConfigForSpace(coreSpace, spaceProviderNs)
	if err != nil {
		logger.Error(err, "Failed to fetch space config", "spacename", coreSpace)
		os.Exit(5)
	}

	edgeClientset, err := ksclientset.NewForConfig(coreRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create provider clientset from config")
		os.Exit(30)
	}

	edgeInformerFactory := emcinformers.NewSharedScopedInformerFactoryWithOptions(edgeClientset, resyncPeriod)
	epPreInformer := edgeInformerFactory.Edge().V2alpha1().EdgePlacements()
	spsPreInformer := edgeInformerFactory.Edge().V2alpha1().SinglePlacementSlices()
	syncfgPreInformer := edgeInformerFactory.Edge().V2alpha1().SyncerConfigs()
	locationPreInformer := edgeInformerFactory.Edge().V2alpha1().Locations()

	managementClientset, err := spaceclientset.NewForConfig(spaceManagementConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for space management")
	}
	spaceInformerFactory := spaceinformers.NewSharedInformerFactory(managementClientset, resyncPeriod)
	spacePreInformer := spaceInformerFactory.Space().V1alpha1().Spaces()

	kubeClient, err := kubernetes.NewForConfig(coreRestConfig)
	if err != nil {
		logger.Error(err, "failed to create k8s clientset for service provider space")
		os.Exit(90)
	}
	kbSpaceRelation := kbuser.NewKubeBindSpaceRelation(ctx, kubeClient)

	doneCh := ctx.Done()

	pt := placement.NewPlacementTranslator(concurrency, ctx,
		locationPreInformer, epPreInformer, spsPreInformer, syncfgPreInformer,
		spaceclient, spaceProviderNs, spacePreInformer, kbSpaceRelation)

	cache.WaitForCacheSync(doneCh, kbSpaceRelation.InformerSynced)
	edgeInformerFactory.Start(doneCh)
	spaceInformerFactory.Start(doneCh)
	pt.Run()
	logger.Info("Time to stop")
}
