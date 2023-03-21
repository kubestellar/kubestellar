/*
Copyright 2023 The KCP Authors.

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

package placement

import (
	"github.com/kcp-dev/logicalcluster/v3"
)

// TestAPIMapProvider is a simple implementation of APIMapProvider.
// It relies on a base map provider and caches the mappings,
// and deletes unneeded entries from the base map provider.
// In the locking order:
// - callers of the TestAPIMapProvider methods precede the baseProducer
// - callers of NewTestAPIMapProvider precede the baseProducer
// - the baseProducer precedes this TestAPIMapProvider
// - this TestAPIMapProvider precedes each of its Clients
type TestAPIMapProvider struct {
	baseProducer BaseAPIMapProvider

	// No mutex needed here because of expected exclusivity of callbacks from baseProducer

	clusters map[logicalcluster.Name]*ClientTracker[ScopedAPIProvider]
}

// BaseAPIMapProvider is a source of API information.
// It is expected to hold a mutex while calling into this client.
type BaseAPIMapProvider DynamicMapProviderWithRelease[logicalcluster.Name, ScopedAPIProvider]

var _ APIMapProvider = &TestAPIMapProvider{}

func NewTestAPIMapProvider(baseProducer BaseAPIMapProvider) *TestAPIMapProvider {
	ans := &TestAPIMapProvider{
		baseProducer: baseProducer,
		clusters:     map[logicalcluster.Name]*ClientTracker[ScopedAPIProvider]{},
	}
	baseProducer.AddReceiver(TestAPIMapProviderAsreceiver{ans}, false)
	return ans
}

type TestAPIMapProviderAsreceiver struct{ *TestAPIMapProvider }

func (tamp TestAPIMapProviderAsreceiver) Receive(cluster logicalcluster.Name, producer ScopedAPIProvider) {
	clusterData, found := tamp.clusters[cluster]
	if !found {
		return
	}
	clusterData.SetProvider(producer)
}

func (tamp *TestAPIMapProvider) AddClient(cluster logicalcluster.Name, client Client[ScopedAPIProvider]) {
	tamp.baseProducer.Get(cluster, func(producer ScopedAPIProvider) {
		clusterData, found := tamp.clusters[cluster]
		if !found {
			clusterData = NewClientTracker[ScopedAPIProvider]()
			clusterData.SetProvider(producer)
		}
		clusterData.AddClient(client)
	})
}

func (tamp *TestAPIMapProvider) RemoveClient(cluster logicalcluster.Name, client Client[ScopedAPIProvider]) {
	tamp.baseProducer.MaybeRelease(cluster, func(ScopedAPIProvider) bool {
		clusterData, found := tamp.clusters[cluster]
		if !found {
			return true
		}
		release := clusterData.RemoveClient(client)
		if release {
			delete(tamp.clusters, cluster)
		}
		return release
	})
}
