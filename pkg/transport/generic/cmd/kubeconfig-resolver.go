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
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/ctrlutil"
)

// resolveWDSKubeconfig resolves WDS kubeconfig using priority order
// Priority 1: WdsClientOptions (handles --wds-kubeconfig flag)
// Priority 2: Cluster-based discovery using WDS name (--wds-name)
func resolveWDSKubeconfig(options *TransportOptions, logger klog.Logger) (*rest.Config, string, error) {
	// Priority 1: Try WdsClientOptions first (this handles --wds-kubeconfig)
	if wdsConfig, err := options.WdsClientOptions.ToRESTConfig(); err == nil {
		logger.Info("Using WDS kubeconfig from client options")
		return wdsConfig, "", nil
	}

	// Priority 2: Cluster-based discovery using WDS name
	if options.WdsName != "" {
		logger.Info("Getting WDS kubeconfig from cluster", "wdsName", options.WdsName)
		return ctrlutil.GetWDSKubeconfig(logger, options.WdsName)
	}

	// Priority 3: Fail if neither option provided
	return nil, "", fmt.Errorf("no WDS configuration provided: specify either --wds-kubeconfig or --wds-name")
}

// resolveTransportKubeconfig resolves ITS/Transport kubeconfig using priority order
// Priority 1: TransportClientOptions (handles --transport-kubeconfig flag)
// Priority 2: Cluster-based discovery using transport name (--transport-name)
// Priority 3: Auto-discovery of single ITS
func resolveTransportKubeconfig(options *TransportOptions, logger klog.Logger) (*rest.Config, string, error) {
	// Priority 1: Try TransportClientOptions first (handles --transport-kubeconfig)
	if transportConfig, err := options.TransportClientOptions.ToRESTConfig(); err == nil {
		logger.Info("Using transport kubeconfig from client options")
		return transportConfig, "", nil
	}

	// Priority 2: Cluster-based discovery using transport name
	if options.TransportName != "" {
		logger.Info("Getting transport kubeconfig from cluster", "transportName", options.TransportName)
		return ctrlutil.GetITSKubeconfig(logger, options.TransportName)
	}

	// Priority 3: Auto-discovery of single ITS
	logger.Info("Attempting to auto-discover single ITS control plane")
	config, name, err := ctrlutil.GetITSKubeconfig(logger, "")
	if err == nil {
		logger.Info("Auto-discovered ITS control plane", "name", name)
		return config, name, nil
	}

	// Priority 4: Fail
	return nil, "", fmt.Errorf("no transport configuration provided: specify --transport-kubeconfig, --transport-name, or ensure exactly one ITS exists")
}

// loadKubeconfigFromPath loads kubeconfig from file path
func loadKubeconfigFromPath(kubeconfigPath string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig file: %w", err)
	}
	return config, nil
}
