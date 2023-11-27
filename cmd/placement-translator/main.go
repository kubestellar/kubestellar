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
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"

	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/kcp/clientset/versioned"
	apiextinfactory "k8s.io/apiextensions-apiserver/pkg/client/kcp/informers/externalversions"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	clusterdiscovery "github.com/kcp-dev/client-go/discovery"
	clusterdynamic "github.com/kcp-dev/client-go/dynamic"
	clusterdynamicinformer "github.com/kcp-dev/client-go/dynamic/dynamicinformer"
	kcpkubeinformers "github.com/kcp-dev/client-go/informers"
	kcpkubeclient "github.com/kcp-dev/client-go/kubernetes"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	kcpscopedclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpclusterclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	tenancyv1a1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"

	emcclusterclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
	emcinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	edgev1a1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/placement"
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
	fs := pflag.NewFlagSet("placement-translator", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	fs.Var(&utilflag.IPPortVar{Val: &serverBindAddress}, "server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")
	fs.IntVar(&concurrency, "concurrency", concurrency, "number of syncs to run in parallel")
	espwClientOpts := NewClientOpts("espw", "access to the edge service provider workspace")
	espwClientOpts.AddFlags(fs)
	rootClientOpts := NewClientOpts("root", "access to root workspace")
	rootClientOpts.overrides.CurrentContext = "root"
	rootClientOpts.AddFlags(fs)
	baseClientOpts := NewClientOpts("allclusters", "access to all clusters")
	baseClientOpts.overrides.CurrentContext = "system:admin"
	baseClientOpts.AddFlags(fs)
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

	espwRestConfig, err := espwClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to build config from flags", "which", espwClientOpts.which)
		os.Exit(10)
	}

	rootRestConfig, err := rootClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to build config from flags", "which", rootClientOpts.which)
		os.Exit(11)
	}

	baseRestConfig, err := baseClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to build config from flags", "which", baseClientOpts.which)
		os.Exit(15)
	}

	edgeClusterClientset, err := emcclusterclientset.NewForConfig(baseRestConfig)
	if err != nil {
		logger.Error(err, "Failed to build all-cluster edge clientset")
		os.Exit(25)
	}

	// Get client config for view of APIExport of edge API
	edgeViewConfig, err := configForViewOfExport(ctx, espwRestConfig, "edge.kubestellar.io")
	if err != nil {
		logger.Error(err, "Failed to create client config for view of edge APIExport")
		os.Exit(30)
	}

	edgeViewClusterClientset, err := emcclusterclientset.NewForConfig(edgeViewConfig)
	if err != nil {
		logger.Error(err, "Failed to create cluster clientset for view of edge APIExport")
		os.Exit(40)
	}

	apiextClusterClient, err := apiextclient.NewForConfig(baseRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create cluster clientset for CustomResourceDefinitions")
		os.Exit(43)
	}

	apiextFactory := apiextinfactory.NewSharedInformerFactory(apiextClusterClient, 0)
	crdClusterPreInformer := apiextFactory.Apiextensions().V1().CustomResourceDefinitions()

	kubeClusterClient, err := kcpkubeclient.NewForConfig(baseRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create kcp-kube all-cluster client")
		os.Exit(44)
	}
	nsClusterClient := kubeClusterClient.CoreV1().Namespaces()
	kubeClusterInformerFactory := kcpkubeinformers.NewSharedInformerFactory(kubeClusterClient, 0)
	nsClusterPreInformer := kubeClusterInformerFactory.Core().V1().Namespaces()

	edgeInformerFactory := emcinformers.NewSharedInformerFactoryWithOptions(edgeViewClusterClientset, resyncPeriod)
	epClusterPreInformer := edgeInformerFactory.Edge().V2alpha1().EdgePlacements()
	spsClusterPreInformer := edgeInformerFactory.Edge().V2alpha1().SinglePlacementSlices()
	syncfgClusterPreInformer := edgeInformerFactory.Edge().V2alpha1().SyncerConfigs()
	customizerClusterPreInformer := edgeInformerFactory.Edge().V2alpha1().Customizers()
	var _ edgev1a1informers.SinglePlacementSliceClusterInformer = spsClusterPreInformer

	rootClientset, err := kcpscopedclientset.NewForConfig(rootRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for root workspace")
		os.Exit(50)
	}

	rootInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(rootClientset, resyncPeriod)
	mbwsPreInformer := rootInformerFactory.Tenancy().V1alpha1().Workspaces()
	var _ tenancyv1a1informers.WorkspaceInformer = mbwsPreInformer

	locationClusterPreInformer := edgeInformerFactory.Edge().V2alpha1().Locations()
	var _ edgev1a1informers.LocationClusterInformer = locationClusterPreInformer

	kcpClusterClientset, err := kcpclusterclientset.NewForConfig(baseRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create all-cluster clientset for kcp APIs")
		os.Exit(60)
	}

	discoveryClusterClient, err := clusterdiscovery.NewForConfig(baseRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create all-cluster discovery client")
		os.Exit(70)
	}

	dynamicClusterClient, err := clusterdynamic.NewForConfig(baseRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create all-cluster dynamic client")
		os.Exit(80)
	}

	dynamicClusterInformerFactory := clusterdynamicinformer.NewDynamicSharedInformerFactory(dynamicClusterClient, 0)

	kcpClusterInformerFactory := kcpinformers.NewSharedInformerFactory(kcpClusterClientset, 0)
	bindingClusterPreInformer := kcpClusterInformerFactory.Apis().V1alpha1().APIBindings()

	doneCh := ctx.Done()
	// TODO: more
	pt := placement.NewPlacementTranslator(concurrency, ctx, locationClusterPreInformer, epClusterPreInformer, spsClusterPreInformer, syncfgClusterPreInformer, customizerClusterPreInformer,
		mbwsPreInformer, kcpClusterClientset, discoveryClusterClient, crdClusterPreInformer, bindingClusterPreInformer,
		dynamicClusterClient, edgeClusterClientset, nsClusterPreInformer, nsClusterClient)

	apiextFactory.Start(doneCh)
	edgeInformerFactory.Start(doneCh)
	rootInformerFactory.Start(doneCh)
	dynamicClusterInformerFactory.Start(doneCh)
	kubeClusterInformerFactory.Start(doneCh)
	kcpClusterInformerFactory.Start(doneCh)
	pt.Run()
	logger.Info("Time to stop")
}

func configForViewOfExport(ctx context.Context, providerConfig *rest.Config, exportName string) (*rest.Config, error) {
	providerScopedClient, err := kcpscopedclientset.NewForConfig(providerConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating client for service provider workspace: %w", err)
	}
	apiExportClient := providerScopedClient.ApisV1alpha1().APIExports()
	logger := klog.FromContext(ctx)
	var apiExport *apisv1alpha1.APIExport
	for {
		apiExport, err = apiExportClient.Get(ctx, exportName, metav1.GetOptions{})
		if err != nil {
			if k8sapierrors.IsNotFound(err) {
				logger.V(2).Info("Pause because APIExport not found", "exportName", exportName)
				time.Sleep(time.Second * 15)
				continue
			}
			return nil, fmt.Errorf("error reading APIExport %s: %w", exportName, err)
		}
		if isAPIExportReady(logger, apiExport) {
			break
		}
		logger.V(2).Info("Pause because APIExport not ready", "exportName", exportName)
		time.Sleep(time.Second * 15)
	}
	viewConfig := rest.CopyConfig(providerConfig)
	serverURL := apiExport.Status.VirtualWorkspaces[0].URL
	logger.V(2).Info("Found APIExport view", "exportName", exportName, "serverURL", serverURL)
	viewConfig.Host = serverURL
	return viewConfig, nil
}

func isAPIExportReady(logger klog.Logger, apiExport *apisv1alpha1.APIExport) bool {
	if !conditions.IsTrue(apiExport, apisv1alpha1.APIExportVirtualWorkspaceURLsReady) {
		logger.V(2).Info("APIExport virtual workspace URLs are not ready", "APIExport", apiExport.Name)
		return false
	}
	if len(apiExport.Status.VirtualWorkspaces) == 0 {
		logger.V(2).Info("APIExport does not have any virtual workspace URLs", "APIExport", apiExport.Name)
		return false
	}
	return true
}
