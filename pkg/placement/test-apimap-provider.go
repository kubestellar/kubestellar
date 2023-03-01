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
	"sync"

	"github.com/kcp-dev/logicalcluster/v3"
)

// TestAPIMapProducer is a simple implementation of APIMapProvider.
// It relies on a base map provider and caches the mappings,
// and deletes unneeded entries from the base map provider.
// In the locking order:
// - callers of the TestAPIMapProducer methods precede the baseProducer
// - callers of NewTestAPIMapProducer precede the baseProducer
// - the baseProducer precedes this TestAPIMapProducer
// - this TestAPIMapProducer precedes each of its Clients
type TestAPIMapProducer struct {
	// baseProducer precedes this TestAPIMapProducer in the locking order
	baseProducer BaseAPIMapProducer

	sync.Mutex

	clusters map[logicalcluster.Name]ClientTracker[ScopedAPIProducer]
}

type BaseAPIMapProducer DynamicMapProducerWithDelete[logicalcluster.Name, ScopedAPIProducer]

var _ APIMapProvider = &TestAPIMapProducer{}

func (tamp *TestAPIMapProducer) noteProducer(cluster logicalcluster.Name, producer ScopedAPIProducer) {
	tamp.Lock()
	defer tamp.Unlock()
	clusterData, found := tamp.clusters[cluster]
	if !found {
		return
	}
	clusterData.SetProvider(producer)
	tamp.clusters[cluster] = clusterData
}

func NewTestAPIMapProducer(baseProducer BaseAPIMapProducer) *TestAPIMapProducer {
	ans := &TestAPIMapProducer{
		baseProducer: baseProducer,
		clusters:     map[logicalcluster.Name]ClientTracker[ScopedAPIProducer]{},
	}
	baseProducer.AddConsumer(ans.noteProducer)
	return ans
}

func (tamp *TestAPIMapProducer) AddClient(cluster logicalcluster.Name, client Client[ScopedAPIProducer]) {
	tamp.baseProducer.Get(cluster, func(producer ScopedAPIProducer) {
		tamp.Lock()
		defer tamp.Unlock()
		clusterData, found := tamp.clusters[cluster]
		if !found {
			clusterData = NewClientTracker[ScopedAPIProducer]()
		}
		clusterData.AddClient(client)
		client.SetProvider(producer)
		tamp.clusters[cluster] = clusterData
	})
}

func (tamp *TestAPIMapProducer) RemoveClient(cluster logicalcluster.Name, client Client[ScopedAPIProducer]) {
	tamp.baseProducer.MaybeDelete(cluster, func(ScopedAPIProducer) bool {
		tamp.Lock()
		defer tamp.Unlock()
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
