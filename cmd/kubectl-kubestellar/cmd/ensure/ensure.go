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

// This is the "ensure" sub-command for kubestellar.

package ensure

import (
	"errors"
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)



// Create Cobra sub-command for 'kubectl kubestellar ensure'
var EnsureCmd = &cobra.Command{
	Use:	"ensure",
	Short:  "Ensure a KubeStellar object is correctly configured",
//	Args:  cobra.ExactArgs(1),
	// If an invalid sub-command is sent, the function in RunE will execute.
	// Use this to inform of invalid arguments, and return an error.
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New(fmt.Sprintf("Invalid sub-command for 'ensure': %s\n", args[0]))
		} else {
			return errors.New(fmt.Sprintf("Missing sub-command for 'ensure'\n"))
		}
	},
}

func init() {
	// Get config flags with default values.
	// Passing "true" will "use persistent client config, rest mapper,
	// discovery client, and propagate them to the places that need them,
	// rather than instantiating them multiple times."
	cliOpts := genericclioptions.NewConfigFlags(true)
	// Make a new flag set named en
	fs := pflag.NewFlagSet("en", pflag.ExitOnError)
	// Add cliOpts flags to fs (flow from syntax is confusing, goes -->)
	cliOpts.AddFlags(fs)

	// Add logging flags to fs
	fs.AddGoFlagSet(flag.CommandLine)
	// Add flags to our command; make these persistent (available to this
	// command and all sub-commands)
	EnsureCmd.PersistentFlags().AddFlagSet(fs)

	// Add location sub-command
	EnsureCmd.AddCommand(newCmdEnsureLocation())
	// Add wds sub-command
	EnsureCmd.AddCommand(newCmdEnsureWds())
}