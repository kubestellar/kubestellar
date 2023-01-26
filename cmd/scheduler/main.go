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

	emcclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	emcinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
)

func main() {
	resyncPeriod := time.Duration(0)
	fs := pflag.NewFlagSet("mailbox-controller", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}

	fs.StringVar(&loadingRules.ExplicitPath, "kubeconfig", loadingRules.ExplicitPath, "pathname of kubeconfig file")
	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.FromContext(ctx)

	// create config for accessing service provider workspace
	providerClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	providerConfig, err := providerClientConfig.ClientConfig()
	if err != nil {
		logger.Error(err, "failed to make config, if running out of cluster, make sure $KUBECONFIG points to kcp server")
		os.Exit(2)
	}

	providerConfig.UserAgent = "mailbox-controller"

	edgeViewConfig, err := configForViewOfExport(ctx, providerConfig, "edge.kcp.io")
	if err != nil {
		logger.Error(err, "Failed to create config for view of edge exports")
		os.Exit(4)
	}

	emcClusterClientset, err := emcclientset.NewForConfig(edgeViewConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for view of edge exports")
		os.Exit(6)
	}

	edgeInformerFactory := emcinformers.NewSharedInformerFactoryWithOptions(emcClusterClientset, resyncPeriod)
	placementsInformer := edgeInformerFactory.Edge().V1alpha1().EdgePlacements().Informer()

	onAdd := func(obj any) {
		logger.Info("Observed add", "obj", obj)
	}
	onUpdate := func(oldObj, newObj any) {
		logger.Info("Observed update", "oldObj", oldObj, "newObj", newObj)
	}
	onDelete := func(obj any) {
		logger.Info("Observed delete", "obj", obj)
	}
	placementsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	edgeInformerFactory.Start(ctx.Done())
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
