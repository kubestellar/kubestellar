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

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpclusterclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
)

func main() {
	resyncPeriod := time.Duration(0)
	fs := pflag.NewFlagSet("mailbox-controller", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)

	inventoryLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	inventoryConfigOverrides := &clientcmd.ConfigOverrides{}

	fs.StringVar(&inventoryLoadingRules.ExplicitPath, "inventory-kubeconfig", inventoryLoadingRules.ExplicitPath, "pathname of kubeconfig file for inventory service provider workspace")
	fs.StringVar(&inventoryConfigOverrides.CurrentContext, "inventory-context", "root", "current-context override for inventory-kubeconfig")

	workloadLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	workloadConfigOverrides := &clientcmd.ConfigOverrides{}

	fs.StringVar(&workloadLoadingRules.ExplicitPath, "workload-kubeconfig", workloadLoadingRules.ExplicitPath, "pathname of kubeconfig file for edge workload service provider workspace")
	fs.StringVar(&workloadConfigOverrides.CurrentContext, "workload-context", workloadConfigOverrides.CurrentContext, "current-context override for workload-kubeconfig")

	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.FromContext(ctx)

	// create config for accessing TMC service provider workspace
	inventoryClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(inventoryLoadingRules, inventoryConfigOverrides)
	inventoryConfig, err := inventoryClientConfig.ClientConfig()
	if err != nil {
		logger.Error(err, "failed to make inventory config")
		os.Exit(2)
	}

	inventoryConfig.UserAgent = "mailbox-controller"

	// Get client config for view of SyncTarget objects
	syncTargetViewConfig, err := configForViewOfExport(ctx, inventoryConfig, "workload.kcp.io")
	if err != nil {
		logger.Error(err, "Failed to create client config for view of SyncTarget exports")
		os.Exit(4)
	}

	tmcClusterClientset, err := kcpclusterclientset.NewForConfig(syncTargetViewConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for view of SyncTarget exports")
		os.Exit(6)
	}

	tmcInformerFactory := kcpinformers.NewSharedInformerFactoryWithOptions(tmcClusterClientset, resyncPeriod)
	syncTargetsInformer := tmcInformerFactory.Workload().V1alpha1().SyncTargets().Informer()

	// create config for accessing TMC service provider workspace
	workspacesClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(workloadLoadingRules, workloadConfigOverrides)
	workspacesConfig, err := workspacesClientConfig.ClientConfig()
	if err != nil {
		logger.Error(err, "failed to make workspaces config")
		os.Exit(8)
	}

	workspacesConfig.UserAgent = "mailbox-controller"

	workspacesClient, err := kcpclientset.NewForConfig(workspacesConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for workspaces")
	}

	workspacesInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(workspacesClient, resyncPeriod)
	workspacesInformer := workspacesInformerFactory.Tenancy().V1alpha1().Workspaces().Informer()

	onAdd := func(obj any) {
		logger.Info("Observed add", "obj", obj)
	}
	onUpdate := func(oldObj, newObj any) {
		logger.Info("Observed update", "oldObj", oldObj, "newObj", newObj)
	}
	onDelete := func(obj any) {
		logger.Info("Observed delete", "obj", obj)
	}
	syncTargetsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	workspacesInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	tmcInformerFactory.Start(ctx.Done())
	workspacesInformerFactory.Start(ctx.Done())
	<-ctx.Done()
	logger.Info("Time to stop")
}

func configForViewOfExport(ctx context.Context, providerConfig *rest.Config, exportName string) (*rest.Config, error) {
	providerClient, err := kcpclientset.NewForConfig(providerConfig)
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
				logger.Info("Pause because APIExport not found", "exportName", exportName)
				time.Sleep(time.Second * 15)
				continue
			}
			return nil, fmt.Errorf("error reading APIExport %s: %w", exportName, err)
		}
		if isAPIExportReady(logger, apiExport) {
			break
		}
		logger.Info("Pause because APIExport not ready", "exportName", exportName)
		time.Sleep(time.Second * 15)
	}
	viewConfig := rest.CopyConfig(providerConfig)
	serverURL := apiExport.Status.VirtualWorkspaces[0].URL
	logger.Info("Found APIExport view", "exportName", exportName, "serverURL", serverURL)
	viewConfig.Host = serverURL
	return viewConfig, nil
}

func isAPIExportReady(logger klog.Logger, apiExport *apisv1alpha1.APIExport) bool {
	if !conditions.IsTrue(apiExport, apisv1alpha1.APIExportVirtualWorkspaceURLsReady) {
		logger.Info("APIExport virtual workspace URLs are not ready", "APIExport", apiExport.Name)
		return false
	}
	if len(apiExport.Status.VirtualWorkspaces) == 0 {
		logger.Info("APIExport does not have any virtual workspace URLs", "APIExport", apiExport.Name)
		return false
	}
	return true
}
