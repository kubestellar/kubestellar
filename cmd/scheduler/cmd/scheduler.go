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

	scheduleroptions "github.com/kcp-dev/edge-mc/cmd/scheduler/options"
	edgeclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgeinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
	"github.com/kcp-dev/edge-mc/pkg/scheduler"
)

func NewSchedulerCommand() *cobra.Command {
	options := scheduleroptions.NewOptions()
	schedulerCommand := &cobra.Command{
		Use:   "scheduler",
		Short: "Reconciles SinglePlacementSlice API objects",
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

	options.AddFlags(schedulerCommand.Flags())

	if v := version.Get().String(); len(v) == 0 {
		schedulerCommand.Version = "<unknown>"
	} else {
		schedulerCommand.Version = v
	}

	return schedulerCommand
}

func Run(ctx context.Context, options *scheduleroptions.Options) error {
	const resyncPeriod = 10 * time.Hour

	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

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

	// create edgeSharedInformerFactory
	edgeConfig, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	edgeViewConfig, err := configForViewOfExport(ctx, edgeConfig, "edge.kcp.io")
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

	// create schedulingSharedInformerFactory
	schedulingConfig, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{
			Context: clientcmdapi.Context{
				Cluster:  "root",
				AuthInfo: "kcp-admin",
			},
		},
	).ClientConfig()
	schedulingViewConfig, err := configForViewOfExport(ctx, schedulingConfig, "scheduling.kcp.io")
	if err != nil {
		logger.Error(err, "failed to create config for view of scheduling exports")
		return err
	}
	schedulingViewClusterClientset, err := kcpclientset.NewForConfig(schedulingViewConfig)
	if err != nil {
		logger.Error(err, "failed to create clientset for view of scheduling exports")
		return err
	}
	schedulingSharedInformerFactory := kcpinformers.NewSharedInformerFactoryWithOptions(schedulingViewClusterClientset, resyncPeriod)

	// create workloadSharedInformerFactory
	workloadConfig, _ := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{
			Context: clientcmdapi.Context{
				Cluster:  "root",
				AuthInfo: "kcp-admin",
			},
		},
	).ClientConfig()
	workloadViewConfig, err := configForViewOfExport(ctx, workloadConfig, "workload.kcp.io")
	if err != nil {
		logger.Error(err, "failed to create config for view of workload exports")
		return err
	}
	workloadViewClusterClientset, err := kcpclientset.NewForConfig(workloadViewConfig)
	if err != nil {
		logger.Error(err, "failed to create clientset for view of workload exports")
		return err
	}
	workloadSharedInformerFactory := kcpinformers.NewSharedInformerFactoryWithOptions(workloadViewClusterClientset, resyncPeriod)

	// create edge-scheduler
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
	es, err := scheduler.NewController(
		ctx,
		kcpClusterClientset,
		edgeClusterClientset,
		edgeSharedInformerFactory.Edge().V1alpha1().EdgePlacements(),
		schedulingSharedInformerFactory.Scheduling().V1alpha1().Locations(),
		workloadSharedInformerFactory.Workload().V1alpha1().SyncTargets(),
	)
	if err != nil {
		logger.Error(err, "failed to create controller", "name", scheduler.ControllerName)
		return err
	}

	// run edge-scheduler
	doneCh := ctx.Done()

	kubeSharedInformerFactory.Start(doneCh)
	edgeSharedInformerFactory.Start(doneCh)
	schedulingSharedInformerFactory.Start(doneCh)
	workloadSharedInformerFactory.Start(doneCh)

	kubeSharedInformerFactory.WaitForCacheSync(doneCh)
	edgeSharedInformerFactory.WaitForCacheSync(doneCh)
	schedulingSharedInformerFactory.WaitForCacheSync(doneCh)
	workloadSharedInformerFactory.WaitForCacheSync(doneCh)
	es.Run(1)

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
