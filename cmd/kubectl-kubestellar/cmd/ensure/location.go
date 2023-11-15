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

// Sub-command for ensuring the existence and configuration a location in a WEC.
// The IMW is provided by the required --imw flag.
// The location name is provided as a required command-line argument.
// Labels in key=value pairs are provided as command-line arguments, for which
// we will ensure that these exist as labels in the Location and SyncTarget.

package ensure

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/klog/v2"

	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"


	clientopts "github.com/kubestellar/kubestellar/pkg/client-options"
	clientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	plugin "github.com/kubestellar/kubestellar/pkg/cliplugins/kubestellar/ensure"
)

var imw string // IMW workspace path

// Create the Cobra sub-command for 'kubectl kubestellar ensure location'
func newCmdEnsureLocation() *cobra.Command {
	// Make location command
	cmdLocation := &cobra.Command{
		Use:     "location --imw <IMW_NAME> <LOCATION_NAME> <\"KEY=VALUE\" ...>",
		Aliases: []string{"loc"},
		Short:   "Ensure existence and configuration of an inventory listing for a WEC",
		// We actually require at least 2 arguments (location name and a label),
		// but more descriptive error messages will be provided by leaving this
		// set to 1.
		Args:    cobra.MinimumNArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error {
			// At this point set silence usage to true, so that any errors
			// following do not result in the help being printed. We only
			// want the help to be displayed when the error is due to an
			// invalid command.
			cmd.SilenceUsage = true
			err := ensureLocation(cmd, args)
			return err
		},
	}

	// Add flag for IMW workspace
	cmdLocation.Flags().StringVar(&imw, "imw", "", "IMW workspace")
	cmdLocation.MarkFlagRequired("imw")
	return cmdLocation
}

// The IMW name is provided by the --imw flag (stored in the "imw" string
// variable), and the location name is a command line argument.
// Labels to check are provided as additional arguments in key=value pairs.
// In this function we will:
// - Work in the provided IMW workspace
// - Check if APIBinding "edge.kubestellar.io" exists in IMW, create if not
// - Check for SyncTarget of provided name in IMW, create if not
// - Check that SyncTarget has an "id" label matching the Location name
// - Ensure that SyncTarget has the labels provided by the user
// - Check for Location of provided name in IMW, create if not
// - Ensure that Location has the labels provided by the user
// - If Location "default" exists, delete it
func ensureLocation(cmdLocation *cobra.Command, args []string) error {
	locationName := args[0]
	labels := args[1:]
	ctx := context.Background()
	logger := klog.FromContext(ctx)

	// Print all flags and their values if verbosity level is at least 1
	cmdLocation.Flags().VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info(fmt.Sprintf("Command line flag %s=%s", flg.Name, flg.Value))
	})

	// Make sure user provided location name is valid
	err := plugin.CheckLocationName(locationName)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Problem with location name %s", locationName))
		return err
	}

	// Make sure user provided labels are valid
	err = plugin.CheckLabelArgs(labels)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Problem with label arguments %s", labels))
		return err
	}

	// Options for IMW workspace
	imwClientOpts := clientopts.NewClientOpts("imw", "Access to the IMW workspace")
	// Set default context to "root"; we will need to append the IMW name to the root server
	imwClientOpts.SetDefaultCurrentContext("root")

	// Get client config from flags
	config, err := imwClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to get config from flags")
		return err
	}

	// Update host to work on objects within IMW workspace
	config.Host += ":" + imw
	logger.V(1).Info(fmt.Sprintf("Set host to %s", config.Host))

	// Create client-go instance from config
	kcpClient, err := kcpclientset.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed create client-go instance")
		return err
	}

	// Check that APIBinding exists, create if not
	// This function prints its own log messages, so no need to add any here.
	err = plugin.VerifyOrCreateAPIBinding(kcpClient, ctx, "edge.kubestellar.io", "edge.kubestellar.io", "root:espw")
	if err != nil {
		return err
	}

	// Create client-go instance from config
	client, err := clientset.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed create client-go instance")
		return err
	}

	// Check that SyncTarget exists and is configured, create/update if not
	// This function prints its own log messages, so no need to add any here.
	err = plugin.VerifyOrCreateSyncTarget(client, ctx, imw, locationName, labels)
	if err != nil {
		return err
	}

	// Check if Location exists and is configured, create/update if not
	// This function prints its own log messages, so no need to add any here.
	err = plugin.VerifyOrCreateLocation(client, ctx, imw, locationName, labels)
	if err != nil {
		return err
	}

	// Check if "default" Location exists, and delete it if so
	// This function prints its own log messages, so no need to add any here.
	err = plugin.VerifyNoDefaultLocation(client, ctx, imw)
	if err != nil {
		return err
	}

	return nil
}
