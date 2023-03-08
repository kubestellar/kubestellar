/*
Copyright 2022 The KCP Authors.

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

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	kcpscopedclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"

	edgeclusterclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgeinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
	"github.com/kcp-dev/edge-mc/pkg/placement"
)

type ClientOpts struct {
	loadingRules *clientcmd.ClientConfigLoadingRules
	overrides    clientcmd.ConfigOverrides
}

func NewClientOpts() *ClientOpts {
	return &ClientOpts{
		loadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
		overrides:    clientcmd.ConfigOverrides{},
	}
}

func (opts *ClientOpts) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.loadingRules.ExplicitPath, "kubeconfig", opts.loadingRules.ExplicitPath, "Path to the kubeconfig file to use for CLI requests")
	flags.StringVar(&opts.overrides.CurrentContext, "context", opts.overrides.CurrentContext, "The name of the kubeconfig context to use")
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
	cliOpts := NewClientOpts()
	cliOpts.AddFlags(fs)
	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	fs.VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info("Command line flag", flg.Name, flg.Value)
	})

	inventoryLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	inventoryConfigOverrides := &clientcmd.ConfigOverrides{}

	fs.StringVar(&inventoryLoadingRules.ExplicitPath, "inventory-kubeconfig", inventoryLoadingRules.ExplicitPath, "pathname of kubeconfig file for inventory service provider workspace")
	fs.StringVar(&inventoryConfigOverrides.CurrentContext, "inventory-context", "root", "current-context override for inventory-kubeconfig")

	workloadLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	workloadConfigOverrides := &clientcmd.ConfigOverrides{}

	fs.StringVar(&workloadLoadingRules.ExplicitPath, "workload-kubeconfig", workloadLoadingRules.ExplicitPath, "pathname of kubeconfig file for edge workload service provider workspace")
	fs.StringVar(&workloadConfigOverrides.CurrentContext, "workload-context", workloadConfigOverrides.CurrentContext, "current-context override for workload-kubeconfig")

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

	restConfig, err := cliOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to build config from flags")
		os.Exit(10)
	}

	// create config for accessing TMC service provider workspace
	inventoryClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(inventoryLoadingRules, inventoryConfigOverrides)
	inventoryConfig, err := inventoryClientConfig.ClientConfig()
	if err != nil {
		logger.Error(err, "failed to make inventory config")
		os.Exit(2)
	}

	inventoryConfig.UserAgent = "mailbox-controller"

	// Get client config for view of SyncTarget objects
	edgeViewConfig, err := configForViewOfExport(ctx, restConfig, "edge.kcp.io")
	if err != nil {
		logger.Error(err, "Failed to create client config for view of edge APIExport")
		os.Exit(4)
	}

	edgeClusterClientset, err := edgeclusterclientset.NewForConfig(edgeViewConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for view of SyncTarget exports")
		os.Exit(6)
	}

	edgeInformerFactory := edgeinformers.NewSharedInformerFactoryWithOptions(edgeClusterClientset, resyncPeriod)
	spsPreInformer := edgeInformerFactory.Edge().V1alpha1().SinglePlacementSlices()
	spsClusterInformer := spsPreInformer.Informer()

	workspacesScopedClientset, err := kcpscopedclientset.NewForConfig(restConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for workspaces")
	}

	workspacesScopedInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(workspacesScopedClientset, resyncPeriod)
	workspacesScopedPreInformer := workspacesScopedInformerFactory.Tenancy().V1alpha1().Workspaces()
	workspacesScopedInformer := workspacesScopedPreInformer.Informer()

	doneCh := ctx.Done()
	edgeInformerFactory.Start(doneCh)
	workspacesScopedInformerFactory.Start(doneCh)
	if !cache.WaitForNamedCacheSync("mailbox-controller", doneCh, spsClusterInformer.HasSynced, workspacesScopedInformer.HasSynced) {
		logger.Error(nil, "Informer syncs not achieved")
		os.Exit(100)
	}
	// TODO: more
	placement.NewPlacementTranslator(ctx, workspacesScopedInformer)
	<-doneCh
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
