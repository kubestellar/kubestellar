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

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	kcpclientsetnoncluster "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	kcpfeatures "github.com/kcp-dev/kcp/pkg/features"

	resolveroptions "github.com/kubestellar/kubestellar/cmd/kubestellar-where-resolver/options"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	wheresolver "github.com/kubestellar/kubestellar/pkg/where-resolver"
)

func NewResolverCommand() *cobra.Command {
	options := resolveroptions.NewOptions()
	resolverCommand := &cobra.Command{
		Use:   "where-resolver",
		Short: "Maintains SinglePlacementSlice API objects for EdgePlacements",
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

	options.AddFlags(resolverCommand.Flags())

	if v := version.Get().String(); len(v) == 0 {
		resolverCommand.Version = "<unknown>"
	} else {
		resolverCommand.Version = v
	}

	return resolverCommand
}

func Run(ctx context.Context, options *resolveroptions.Options) error {
	const resyncPeriod = 10 * time.Hour
	const numThreads = 2

	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	// create edgeSharedInformerFactory
	espwRestConfig, err := options.EspwClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to create config from flags")
		return err
	}
	edgeViewConfig, err := configForViewOfExport(ctx, espwRestConfig, "edge.kcp.io")
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
	rootRestConfig, err := options.RootClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to create config from flags")
		return err
	}
	schedulingViewConfig, err := configForViewOfExport(ctx, rootRestConfig, "scheduling.kcp.io")
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
	workloadViewConfig, err := configForViewOfExport(ctx, rootRestConfig, "workload.kcp.io")
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

	// create where-resolver
	sysAdmRestConfig, err := options.SysAdmClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to create config from flags")
		return err
	}
	kcpClusterClientset, err := kcpclientset.NewForConfig(sysAdmRestConfig)
	if err != nil {
		logger.Error(err, "failed to create kcp clientset for controller")
		return err
	}
	edgeClusterClientset, err := edgeclientset.NewForConfig(sysAdmRestConfig)
	if err != nil {
		logger.Error(err, "failed to create edge clientset for controller")
		return err
	}
	es, err := wheresolver.NewController(
		ctx,
		kcpClusterClientset,
		edgeClusterClientset,
		edgeSharedInformerFactory.Edge().V1alpha1().EdgePlacements(),
		edgeSharedInformerFactory.Edge().V1alpha1().SinglePlacementSlices(),
		schedulingSharedInformerFactory.Scheduling().V1alpha1().Locations(),
		workloadSharedInformerFactory.Workload().V1alpha1().SyncTargets(),
	)
	if err != nil {
		logger.Error(err, "failed to create controller", "name", wheresolver.ControllerName)
		return err
	}

	// run where-resolver
	doneCh := ctx.Done()

	edgeSharedInformerFactory.Start(doneCh)
	schedulingSharedInformerFactory.Start(doneCh)
	workloadSharedInformerFactory.Start(doneCh)

	edgeSharedInformerFactory.WaitForCacheSync(doneCh)
	schedulingSharedInformerFactory.WaitForCacheSync(doneCh)
	workloadSharedInformerFactory.WaitForCacheSync(doneCh)

	es.Run(numThreads)

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
