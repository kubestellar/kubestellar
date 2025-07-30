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

package cmd

import (
	"fmt"

	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/ctrlutil"
)

// resolveWDSKubeconfig resolves WDS kubeconfig using priority order
// Priority 1: WdsClientOptions (handles --wds-kubeconfig flag)
// Priority 2: Cluster-based discovery using WDS name (--wds-name)
func resolveWDSKubeconfig(options *TransportOptions, logger klog.Logger) (*rest.Config, string, error) {
	// Priority 1: Try WdsClientOptions first (this handles --wds-kubeconfig)
	// This will succeed if user provided --wds-kubeconfig flag
	if wdsConfig, err := options.WdsClientOptions.ToRESTConfig(); err == nil {
		logger.Info("Using WDS kubeconfig from client options")
		return wdsConfig, "", nil
	}

	// Priority 2: Cluster-based discovery using WDS name (same as kubestellar controller-manager)
	if options.WdsName != "" {
		logger.Info("Getting WDS kubeconfig from cluster", "wdsName", options.WdsName)
		// Use the SAME function that kubestellar controller-manager uses
		return ctrlutil.GetWDSKubeconfig(logger, options.WdsName)
	}

	// Priority 3: Fail if neither option provided
	return nil, "", fmt.Errorf("no WDS configuration provided: specify either --wds-kubeconfig or --wds-name")
}
