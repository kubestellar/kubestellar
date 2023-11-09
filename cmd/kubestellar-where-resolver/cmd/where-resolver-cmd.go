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

	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	kcpfeatures "github.com/kcp-dev/kcp/pkg/features"

	resolveroptions "github.com/kubestellar/kubestellar/cmd/kubestellar-where-resolver/options"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeclusterclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
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
	espwClientset, err := edgeclientset.NewForConfig(espwRestConfig)
	if err != nil {
		logger.Error(err, "failed to create clientset for service provider space")
		return err
	}
	edgeSharedInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(espwClientset, resyncPeriod)

	// create where-resolver
	baseRestConfig, err := options.BaseClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to create config from flags")
		return err
	}
	edgeClusterClientset, err := edgeclusterclientset.NewForConfig(baseRestConfig)
	if err != nil {
		logger.Error(err, "failed to create edge clientset for controller")
		return err
	}
	es, err := wheresolver.NewController(
		ctx,
		edgeClusterClientset,
		edgeSharedInformerFactory.Edge().V2alpha1().EdgePlacements(),
		edgeSharedInformerFactory.Edge().V2alpha1().SinglePlacementSlices(),
		edgeSharedInformerFactory.Edge().V2alpha1().Locations(),
		edgeSharedInformerFactory.Edge().V2alpha1().SyncTargets(),
	)
	if err != nil {
		logger.Error(err, "failed to create controller", "name", wheresolver.ControllerName)
		return err
	}

	// run where-resolver
	doneCh := ctx.Done()

	edgeSharedInformerFactory.Start(doneCh)

	edgeSharedInformerFactory.WaitForCacheSync(doneCh)

	es.Run(numThreads)

	return nil
}
