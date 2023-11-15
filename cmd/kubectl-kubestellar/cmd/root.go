/*
Copyright 2023 The KubeStellar Authors.

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

// This is the root for kubectl-kubestellar.

package cmd

import (
	"fmt"
	"os"
	"errors"
	"flag"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/klog/v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kubestellar/kubestellar/cmd/kubectl-kubestellar/cmd/ensure"
	"github.com/kubestellar/kubestellar/cmd/kubectl-kubestellar/cmd/remove"
)

// Root kubestellar command
// HACK the usage section of help will not show "kubectl" prefixing the
// command string if the "Use" key is set to just "kubestellar". If the "Use"
// key is set to "kubectl kubestellar", then "kubestellar" gets dropped from
// the usage command string. This is an open issue in Cobra for kubectl plugins.
// The workaround is to add "kubectl" along with a non-breaking space to the
// "Use" string, but this will break the autocompletion script generator.
// Luckily, the help printed to the screen does contain a regular space that
// can be copied and pasted without issue.
// Having accurate help is more important than autocompletion, so we will ensure
// help prints the full command string, and we will disable the "completion"
// script generator command with a dummy function.
// See https://github.com/spf13/cobra/issues/2017
var rootCmd = &cobra.Command{
	Use:   "kubectl\u00A0kubestellar",
	Short: "KubeStellar plugin for kubectl",
	Long:  `KubeStellar is a flexible solution for challenges associated with multi-cluster 
configuration management for edge, multi-cloud, and hybrid cloud.
This command provides the kubestellar sub-command for kubectl.`,
	Args:  cobra.ExactArgs(1),
	// If an invalid sub-command is sent, the function in RunE will execute.
	// Use this to inform of invalid arguments, and return an error.
	// For details on inconsistent error behavior, see
	// https://github.com/spf13/cobra/issues/706
	RunE:  func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			// NOTE this only runs if "Args:  cobra.ExactArgs(1)" is set; if not
			// set the error message is brief and does not print the help.
			return errors.New(fmt.Sprintf("Invalid kubestellar sub-command: %s\n", args[0]))
		} else {
			// NOTE this does not run if "Args:  cobra.ExactArgs(1)" is set,
			// although the error message printed is acceptable.
			return errors.New("Missing kubestellar sub-command\n")
		}
	},
}

// Dummy function to disable auto-completion script generation, since this
// feature is broken (see comments above rootCmd).
var completionCmd = &cobra.Command{
	Use:	"completion",
	Short:	"Generate the autocompletion script for the specified shell",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("Not implemented\n")
	},
}

func init() {
	// Get config flags with default values.
	// Passing "true" will "use persistent client config, rest mapper,
	// discovery client, and propagate them to the places that need them,
	// rather than instantiating them multiple times."
	cliOpts := genericclioptions.NewConfigFlags(true)
	// Make a new flag set named en
	fs := pflag.NewFlagSet("root", pflag.ExitOnError)
	// Add cliOpts flags to fs (flow from syntax is confusing, goes -->)
	cliOpts.AddFlags(fs)

	// Allow logging to pick up flags
	klog.InitFlags(flag.CommandLine)
	// Add logging flags to fs
	fs.AddGoFlagSet(flag.CommandLine)
	// Add flags to our command; make these persistent (available to this
	// command and all sub-commands)
	rootCmd.PersistentFlags().AddFlagSet(fs)

	// Add sub-commands
	rootCmd.AddCommand(ensure.EnsureCmd)
	rootCmd.AddCommand(remove.RemoveCmd)
	rootCmd.AddCommand(completionCmd)
}

// Function for executing root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
