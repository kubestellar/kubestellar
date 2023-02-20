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

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	kcpkubernetesinformers "github.com/kcp-dev/client-go/informers"
	kcpkubernetesclientset "github.com/kcp-dev/client-go/kubernetes"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	kcpclientsetnoncluster "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	kcpfeatures "github.com/kcp-dev/kcp/pkg/features"

	placementoptions "github.com/kcp-dev/edge-mc/cmd/placement/options"
	edgeclient "github.com/kcp-dev/edge-mc/pkg/client"
	edgeclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgeinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
	edgeplacement "github.com/kcp-dev/edge-mc/pkg/reconciler/scheduling/placement"
)

func NewPlacementCommand() *cobra.Command {
	options := placementoptions.NewOptions()
	placementCommand := &cobra.Command{
		Use:   "placement",
		Short: "Reconciles edge placement API objects",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := options.Logs.ValidateAndApply(kcpfeatures.DefaultFeatureGate); err != nil {
				return err
			}
			if err := options.Complete(); err != nil {
				return err
			}

			if err := options.Validate(); err != nil {
				return err
			}

			ctx := context.Background()
			if err := Run(ctx, options); err != nil {
				return err
			}

			<-ctx.Done()

			return nil
		},
	}

	options.AddFlags(placementCommand.Flags())

	if v := version.Get().String(); len(v) == 0 {
		placementCommand.Version = "<unknown>"
	} else {
		placementCommand.Version = v
	}

	return placementCommand
}

func Run(ctx context.Context, options *placementoptions.Options) error {
	const resyncPeriod = 10 * time.Hour

	logger := klog.FromContext(ctx)

	// create cfg
	loadingRules := clientcmd.ClientConfigLoadingRules{ExplicitPath: options.KcpKubeconfig}
	configOverrides := &clientcmd.ConfigOverrides{
		Context: clientcmdapi.Context{
			Cluster:  "base",
			AuthInfo: "shard-admin",
		},
	}
	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(&loadingRules, configOverrides).ClientConfig()
	if err != nil {
		logger.Error(err, "failed to make config, is kcp-kubeconfig correct?")
		return err
	}

	// create kubeSharedInformerFactory
	kubernetesConfig := rest.CopyConfig(cfg)
	clientgoExternalClient, err := kcpkubernetesclientset.NewForConfig(kubernetesConfig)
	if err != nil {
		logger.Error(err, "failed to create kube cluter client")
		return err
	}
	kubeSharedInformerFactory := kcpkubernetesinformers.NewSharedInformerFactory(
		clientgoExternalClient,
		resyncPeriod,
	)

	// create kcpSharedInformerFactory
	kcpConfig := rest.CopyConfig(cfg)
	edgeclient.ConfigForScheduling(kcpConfig)
	kcpClusterClient, err := kcpclientset.NewForConfig(kcpConfig)
	if err != nil {
		logger.Error(err, "failed to create kcp cluster client")
		return err
	}
	kcpSharedInformerFactory := kcpinformers.NewSharedInformerFactoryWithOptions(
		kcpClusterClient,
		resyncPeriod,
	)

	// create edgeSharedInformerFactory
	providerConfig, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	edgeViewConfig, err := configForViewOfExport(ctx, providerConfig, "edge.kcp.io")
	if err != nil {
		logger.Error(err, "failed to create config for view of edge exports")
		return err
	}
	edgeViewClusterClientset, err := edgeclientset.NewForConfig(edgeViewConfig)
	if err != nil {
		logger.Error(err, "failed to create clientset for view of edge exports")
		return err
	}
	edgeSharedInformerFactory := edgeinformers.NewSharedInformerFactoryWithOptions(edgeViewClusterClientset, resyncPeriod)

	// create the edge-scheduler
	controllerConfig := rest.CopyConfig(cfg)
	kcpClusterClientset, err := kcpclientset.NewForConfig(controllerConfig)
	if err != nil {
		logger.Error(err, "failed to create kcp clientset for controller")
		return err
	}
	edgeClusterClientset, err := edgeclientset.NewForConfig(controllerConfig)
	if err != nil {
		logger.Error(err, "failed to create edge clientset for controller")
		return err
	}
	c, err := edgeplacement.NewController(
		kcpClusterClientset,
		edgeClusterClientset,
		kubeSharedInformerFactory.Core().V1().Namespaces(),
		kcpSharedInformerFactory.Scheduling().V1alpha1().Locations(),
		kcpSharedInformerFactory.Scheduling().V1alpha1().Placements(),
		edgeSharedInformerFactory.Edge().V1alpha1().EdgePlacements(),
	)
	if err != nil {
		logger.Error(err, "Failed to create controller")
		return err
	}

	// run the edge-scheduler
	kubeSharedInformerFactory.Start(ctx.Done())
	kcpSharedInformerFactory.Start(ctx.Done())
	edgeSharedInformerFactory.Start(ctx.Done())
	kubeSharedInformerFactory.WaitForCacheSync(ctx.Done())
	kcpSharedInformerFactory.WaitForCacheSync(ctx.Done())
	edgeSharedInformerFactory.WaitForCacheSync(ctx.Done())
	c.Start(ctx, 1)

	return nil
}

func configForViewOfExport(ctx context.Context, providerConfig *rest.Config, exportName string) (*rest.Config, error) {
	providerClient, err := kcpclientsetnoncluster.NewForConfig(providerConfig)
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
