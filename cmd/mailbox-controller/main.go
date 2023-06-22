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

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	kcpscopedclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpclusterclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	"github.com/kcp-dev/logicalcluster/v3"

	clientopts "github.com/kubestellar/kubestellar/pkg/client-options"
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

	inventoryClientOpts := clientopts.NewClientOpts("inventory", "access to APIExport view of SyncTarget objects")
	inventoryClientOpts.SetDefaultCurrentContext("root")

	workloadClientOpts := clientopts.NewClientOpts("workload", "access to edge service provider workspace")
	mbwsClientOpts := clientopts.NewClientOpts("mbws", "access to mailbox workspaces (really all clusters)")
	mbwsClientOpts.SetDefaultCurrentContext("base")

	inventoryClientOpts.AddFlags(fs)
	workloadClientOpts.AddFlags(fs)
	mbwsClientOpts.AddFlags(fs)

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

	// create config for accessing TMC service provider workspace
	inventoryClientConfig, err := inventoryClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make inventory config")
		os.Exit(2)
	}

	inventoryClientConfig.UserAgent = "mailbox-controller"

	// Get client config for view of SyncTarget objects
	syncTargetViewConfig, err := configForViewOfExport(ctx, inventoryClientConfig, "workload.kcp.io")
	if err != nil {
		logger.Error(err, "Failed to create client config for view of SyncTarget exports")
		os.Exit(4)
	}

	stViewClusterClientset, err := kcpclusterclientset.NewForConfig(syncTargetViewConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for view of SyncTarget exports")
		os.Exit(6)
	}

	stViewInformerFactory := kcpinformers.NewSharedInformerFactoryWithOptions(stViewClusterClientset, resyncPeriod)
	syncTargetClusterPreInformer := stViewInformerFactory.Workload().V1alpha1().SyncTargets()

	// create config for accessing edge service provider workspace
	workspaceClientConfig, err := workloadClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make workspaces config")
		os.Exit(8)
	}

	workspaceClientConfig.UserAgent = "mailbox-controller"

	workspaceScopedClientset, err := kcpscopedclientset.NewForConfig(workspaceClientConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for workspaces")
	}

	workspaceScopedInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(workspaceScopedClientset, resyncPeriod)
	workspaceScopedPreInformer := workspaceScopedInformerFactory.Tenancy().V1alpha1().Workspaces()

	mbwsClientConfig, err := mbwsClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make all-cluster config")
		os.Exit(20)
	}
	mbwsClientConfig.UserAgent = "mailbox-controller"
	mbwsClientset, err := kcpclusterclientset.NewForConfig(mbwsClientConfig)
	if err != nil {
		logger.Error(err, "Failed to create all-cluster clientset")
		os.Exit(24)
	}

	ctl := newMailboxController(ctx, espwPath, syncTargetClusterPreInformer, workspaceScopedPreInformer,
		workspaceScopedClientset.TenancyV1alpha1().Workspaces(),
		mbwsClientset.ApisV1alpha1().APIBindings(),
	)

	doneCh := ctx.Done()
	stViewInformerFactory.Start(doneCh)
	workspaceScopedInformerFactory.Start(doneCh)

	ctl.Run(concurrency)

	logger.Info("Time to stop")
}

func configForViewOfExport(ctx context.Context, providerConfig *rest.Config, exportName string) (*rest.Config, error) {
	providerClient, err := kcpscopedclientset.NewForConfig(providerConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating client for service provider workspace: %w", err)
	}
	apiExportClient := providerClient.ApisV1alpha1().APIExports()
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
