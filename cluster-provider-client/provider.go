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

package clusterproviderclient

import (
	"context"
	"errors"

	clusterprovider "github.com/kcp-dev/edge-mc/cluster-provider-client/cluster"
	kindprovider "github.com/kcp-dev/edge-mc/cluster-provider-client/kind"
	v1alpha1apis "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
)

// Each provider gets its own namespace named prefixNamespace+providerName
const prefixNamespace = "lcprovider-"

func GetNamespace(providerName string) string {
	return prefixNamespace + providerName
}

// TODO: this is termporary for stage 1. For stage 2 we expect to have a uniform interface for all informers.
func NewProvider(ctx context.Context,
	providerName string,
	providerType v1alpha1apis.ClusterProviderType) (ProviderClient, error) {
	var newProvider ProviderClient = nil
	switch providerType {
	case v1alpha1apis.KindProviderType:
		newProvider = kindprovider.New(providerName)
	default:
		err := errors.New("unknown provider type")
		return nil, err
	}
	return newProvider, nil
}

// Provider defines methods to retrieve, list, and watch fleet of clusters.
// The provider is responsible for discovering and managing the lifecycle of each
// cluster.
//
// Example:
// ES: add example, remove list() add Get()
type ProviderClient interface {
	Create(ctx context.Context, name string, opts clusterprovider.Options) (clusterprovider.LogicalClusterInfo, error)
	Delete(ctx context.Context, name string, opts clusterprovider.Options) error

	// List returns a list of logical clusters.
	// This method is used to discover the initial set of logical clusters
	// and to refresh the list of logical clusters periodically.
	List() ([]string, error)

	// Watch returns a Watcher that watches for changes to a list of logical clusters
	// and react to potential changes.
	// TODO: I have yet to implement it for the kind provider type, so commenting it out for now
	Watch() (clusterprovider.Watcher, error)
}
