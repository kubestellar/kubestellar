/*
Copyright 2025 The KubeStellar Authors.

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
// Priority 1: Direct kubeconfig file path (--wds-kubeconfig)
// Priority 2: Cluster-based discovery using WDS name (--wds-name) - same as kubestellar controller
func resolveWDSKubeconfig(options *TransportOptions, logger klog.Logger) (*rest.Config, string, error) {
	// Priority 1: Direct kubeconfig file path (highest priority)
	if options.WdsKubeconfigPath != "" {
		logger.Info("Using WDS kubeconfig from file", "path", options.WdsKubeconfigPath)
		config, err := loadKubeconfigFromPath(options.WdsKubeconfigPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to load WDS kubeconfig from path %s: %w", options.WdsKubeconfigPath, err)
		}
		return config, "", nil
	}

	// Priority 2: Cluster-based discovery using WDS name (same as kubestellar controller-manager)
	if options.WdsName != "" {
		logger.Info("Getting WDS kubeconfig from cluster", "wdsName", options.WdsName)
		// Use the SAME function that kubestellar controller-manager uses
		return ctrlutil.GetWDSKubeconfig(logger, options.WdsName)
	}

	return nil, "", fmt.Errorf("no WDS configuration provided: specify either --wds-kubeconfig or --wds-name")
}

// loadKubeconfigFromPath loads kubeconfig from file path
func loadKubeconfigFromPath(kubeconfigPath string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig file: %w", err)
	}
	return config, nil
}
