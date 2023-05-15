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
	goflags "flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	plugin "github.com/kcp-dev/edge-mc/pkg/cliplugins/kcp-edge/syncer-gen"
)

var (
	syncerGenExample = `
	# Setup workspace for syncer to interact and then install syncer on a physical cluster
	%[1]s syncer-gen <name> --syncer-image <edge-syncer-image> -o edge-syncer.yaml
	KUBECONFIG=<a-physical-cluster-kubeconfig> kubectl apply -f edge-syncer.yaml

	# Directly apply the manifest
	%[1]s syncer-gen <name> --syncer-image <edge-syncer-image> -o - | KUBECONFIG=<a-physical-cluster-kubeconfig> kubectl apply -f -
`
)

func syncerGenCommand() *cobra.Command {

	// syncer-gen command
	options := plugin.NewEdgeSyncOptions(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})

	cmd := &cobra.Command{
		Use:          "syncer-gen <name> --syncer-image <edge-syncer-image> -o <output-file>",
		Short:        "Create service account and RBAC permissions in the workspace in kcp for Edge MC. Output a manifest to deploy a syncer in a physical cluster.",
		Example:      fmt.Sprintf(syncerGenExample, "kubectl kubestellar"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return c.Help()
			}

			if err := options.Complete(args); err != nil {
				return err
			}

			if err := options.Validate(); err != nil {
				return err
			}

			return options.Run(c.Context())
		},
	}

	options.BindFlags(cmd)

	// setup klog
	fs := goflags.NewFlagSet("klog", goflags.PanicOnError)
	klog.InitFlags(fs)
	cmd.PersistentFlags().AddGoFlagSet(fs)

	if v := version.Get().String(); len(v) == 0 {
		cmd.Version = "<unknown>"
	} else {
		cmd.Version = v
	}

	return cmd
}

func main() {
	cmd := syncerGenCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
