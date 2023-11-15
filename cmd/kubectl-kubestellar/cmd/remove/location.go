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

// Sub-command for removing a location.
// The IMW is provided by the required --imw flag, and the location name is
// given as a command line argument. The SyncTarget and Location object within
// the IMW will be deleted.

package remove

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"k8s.io/klog/v2"

	clientopts "github.com/kubestellar/kubestellar/pkg/client-options"
	clientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	plugin "github.com/kubestellar/kubestellar/pkg/cliplugins/kubestellar/remove"
)

var imw string // IMW workspace path

// Create the Cobra sub-command for 'kubectl kubestellar remove location'
func newCmdRemoveLocation() *cobra.Command {
	// Make location command
	cmdLocation := &cobra.Command{
		Use:     "location --imw <IMW_NAME> <LOCATION_NAME>",
		Aliases: []string{"loc"},
		Short:   "Delete an inventory entry for a given WEC",
		Args:    cobra.ExactArgs(1),
		RunE:    func(cmd *cobra.Command, args []string) error {
			// At this point set silence usage to true, so that any errors
			// following do not result in the help being printed. We only
			// want the help to be displayed when the error is due to an
			// invalid command.
			cmd.SilenceUsage = true
			err := removeLocation(cmd, args)
			return err
		},
	}

	// Add flag for IMW workspace
	cmdLocation.Flags().StringVar(&imw, "imw", "", "IMW workspace")
	cmdLocation.MarkFlagRequired("imw")
	return cmdLocation
}

// Delete SyncTarget and Location from IMW.
// The IMW name is provided by the --imw flag (stored in the "imw" string
// variable), and the location name is a command line argument.
func removeLocation(cmdLocation *cobra.Command, args []string) error {
	locationName := args[0]
	ctx := context.Background()
	logger := klog.FromContext(ctx)

	// Print all flags and their values if verbosity level is at least 1
	cmdLocation.Flags().VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info(fmt.Sprintf("Command line flag %s=%s", flg.Name, flg.Value))
	})

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
	client, err := clientset.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed create client-go instance")
		return err
	}

	// Delete the SyncTarget object
	removed, err := plugin.DeleteSyncTarget(client, ctx, locationName)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Problem removing SyncTarget %s from workspace root:%s", locationName, imw))
		return err
	}
	if removed {
		logger.Info(fmt.Sprintf("Removed SyncTarget %s from workspace root:%s", locationName, imw))
	} else {
		logger.Info(fmt.Sprintf("Verified no SyncTarget %s in workspace root:%s", locationName, imw))
	}

	// Delete the Location object
	removed, err = plugin.DeleteLocation(client, ctx, locationName)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Problem removing Location %s from workspace root:%s", locationName, imw))
		return err
	}
	if removed {
		logger.Info(fmt.Sprintf("Removed Location %s from workspace root:%s", locationName, imw))
	} else {
		logger.Info(fmt.Sprintf("Verified no Location %s in workspace root:%s", locationName, imw))
	}

	return nil
}