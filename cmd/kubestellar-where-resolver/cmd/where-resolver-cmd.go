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
	"time"

	"github.com/spf13/cobra"

	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/component-base/featuregate"
	"k8s.io/component-base/logs"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	resolveroptions "github.com/kubestellar/kubestellar/cmd/kubestellar-where-resolver/options"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
	wheresolver "github.com/kubestellar/kubestellar/pkg/where-resolver"
	spaceclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
	spacemanager "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager"
)

func NewResolverCommand() *cobra.Command {
	options := resolveroptions.NewOptions()
	resolverCommand := &cobra.Command{
		Use:   "where-resolver",
		Short: "Maintains SinglePlacementSlice API objects for EdgePlacements",
		RunE: func(cmd *cobra.Command, args []string) error {
			featureGate := utilfeature.DefaultMutableFeatureGate
			if err := featureGate.Add(map[featuregate.Feature]featuregate.FeatureSpec{
				logs.ContextualLogging: {Default: true, PreRelease: featuregate.Alpha},
			}); err != nil {
				return err
			}

			if err := options.Logs.ValidateAndApply(featureGate); err != nil {
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

	spaceManagementConfig, err := options.SpaceMgtClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to create space management API client config from flags")
		return err
	}
	spaceClient, err := spaceclient.NewMultiSpace(ctx, spaceManagementConfig)
	if err != nil {
		logger.Error(err, "Failed to create space-aware client")
		return err
	}
	spaceProviderNs := spacemanager.ProviderNS(options.SpaceProvider)

	kcsRestConfig, err := spaceClient.ConfigForSpace(options.KcsName, spaceProviderNs)
	if err != nil {
		logger.Error(err, "Failed to construct space config", "spacename", options.KcsName)
		return err
	}

	edgeClientset, err := edgeclientset.NewForConfig(kcsRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create edge clientset for KubeStellar Core Space")
		return err
	}
	edgeSharedInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(edgeClientset, resyncPeriod)

	// create where-resolver
	kubeClient, err := kubernetes.NewForConfig(kcsRestConfig)
	if err != nil {
		logger.Error(err, "failed to create k8s clientset for KubeStellar Core Space")
		return err
	}
	kbSpaceRelation := kbuser.NewKubeBindSpaceRelation(ctx, kubeClient)

	es, err := wheresolver.NewController(
		ctx,
		spaceClient,
		spaceProviderNs,
		edgeSharedInformerFactory.Edge().V2alpha1().EdgePlacements(),
		edgeSharedInformerFactory.Edge().V2alpha1().SinglePlacementSlices(),
		edgeSharedInformerFactory.Edge().V2alpha1().Locations(),
		edgeSharedInformerFactory.Edge().V2alpha1().SyncTargets(),
		kbSpaceRelation,
	)
	if err != nil {
		logger.Error(err, "failed to create controller", "name", wheresolver.ControllerName)
		return err
	}

	// run where-resolver
	doneCh := ctx.Done()

	edgeSharedInformerFactory.Start(doneCh)
	cache.WaitForCacheSync(doneCh, kbSpaceRelation.InformerSynced)
	edgeSharedInformerFactory.WaitForCacheSync(doneCh)

	es.Run(numThreads)

	return nil
}
