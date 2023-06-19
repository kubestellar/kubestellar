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

package kindprovider

import (
	"context"
	"strings"

	kind "sigs.k8s.io/kind/pkg/cluster"

	clusterprovider "github.com/kcp-dev/edge-mc/cluster-provider-client/cluster"
)

// KindClusterProvider is a cluster provider that works with a local Kind instance.
type KindClusterProvider struct {
	kindProvider *kind.Provider
	providerName string
}

// New creates a new KindClusterProvider
func New(providerName string) KindClusterProvider {
	kindProvider := kind.NewProvider()
	return KindClusterProvider{
		kindProvider: kindProvider,
		providerName: providerName,
	}
}

func (k KindClusterProvider) Create(ctx context.Context,
	name string,
	opts clusterprovider.Options) (clusterprovider.LogicalClusterInfo, error) {
	var resCluster clusterprovider.LogicalClusterInfo

	err := k.kindProvider.Create(string(name), kind.CreateWithKubeconfigPath(opts.KubeconfigPath))
	if err != nil {
		if strings.HasPrefix(err.Error(), "node(s) already exist for a cluster with the name") {
			// TODO: check whether it's the same cluster and return success if true
		} else {
			return resCluster, err
		}
	}

	cfg, err := k.kindProvider.KubeConfig(string(name), false)
	if err != nil {
		return resCluster, err
	}
	resCluster = *clusterprovider.New(cfg, opts)
	return resCluster, err
}

func (k KindClusterProvider) Delete(ctx context.Context,
	name string,
	opts clusterprovider.Options) error {

	return k.kindProvider.Delete(string(name), opts.KubeconfigPath)
}

func (k KindClusterProvider) List() ([]string, error) {
	list, err := k.kindProvider.List()
	if err != nil {
		return nil, err
	}
	logicalNameList := make([]string, 0, len(list))
	for _, cluster := range list {
		logicalNameList = append(logicalNameList, cluster)
	}
	return logicalNameList, err
}

func (k KindClusterProvider) Watch() (mywatch clusterprovider.Watcher, err error) {
	// TODO
	return nil, nil
}
