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
	"fmt"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kcp-dev/edge-mc/cmd/kubectl-kcpforedgesyncer/cliplugins/workload/plugin"
)

var (
	edgeSyncExample = `
	# Ensure a syncer is running on the specified sync target.
	%[1]s workload edge-sync <sync-target-name> --syncer-image <kcp-syncer-image> -o edge-syncer.yaml
	KUBECONFIG=<pcluster-config> kubectl apply -f edge-syncer.yaml

	# Directly apply the manifest
	%[1]s workload sync <sync-target-name> --syncer-image <kcp-syncer-image> -o - | KUBECONFIG=<pcluster-config> kubectl apply -f -
`
)

// New provides a cobra command for workload operations.
func New(streams genericclioptions.IOStreams) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Aliases:          []string{"workloads"},
		Use:              "workload",
		Short:            "Manages KCP sync targets",
		SilenceUsage:     true,
		TraverseChildren: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// EdgeSync command
	edgeSyncOptions := plugin.NewEdgeSyncOptions(streams)

	enableEdgeSyncerCmd := &cobra.Command{
		Use:          "edge-sync <sync-target-name> --syncer-image <kcp-syncer-image> [--resources=<resource1>,<resource2>..] -o <output-file>",
		Short:        "Create a synctarget for Edge MC in kcp with service account and RBAC permissions. Output a manifest to deploy a syncer for the given sync target in a physical cluster.",
		Example:      fmt.Sprintf(edgeSyncExample, "kubectl kcp"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) != 1 {
				return c.Help()
			}

			if err := edgeSyncOptions.Complete(args); err != nil {
				return err
			}

			if err := edgeSyncOptions.Validate(); err != nil {
				return err
			}

			return edgeSyncOptions.Run(c.Context())
		},
	}

	edgeSyncOptions.BindFlags(enableEdgeSyncerCmd)
	cmd.AddCommand(enableEdgeSyncerCmd)

	return cmd, nil
}
