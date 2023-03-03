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

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func RelayWhatResolver() RelayMap[edgeapi.ExternalName, WorkloadParts] {
	return NewRelayMap[edgeapi.ExternalName, WorkloadParts](true)
}

func RelayWhereResolver() RelayMap[edgeapi.ExternalName, ResolvedWhere] {
	return NewRelayMap[edgeapi.ExternalName, ResolvedWhere](true)
}

func NewDummySetBinder() SetBinder {
	return dummySetBinder{NewRelayMap[ProjectionKey, *ProjectionPerCluster](true)}
}

type dummySetBinder struct {
	DynamicMapProducer[ProjectionKey, *ProjectionPerCluster]
}

func (dummySetBinder) AsWhatConsumer() DynamicMapConsumer[edgeapi.ExternalName, WorkloadParts] {
	return RelayWhatResolver()
}

func (dummySetBinder) AsWhereConsumer() DynamicMapConsumer[edgeapi.ExternalName, ResolvedWhere] {
	return RelayWhereResolver()
}

type dummyClient[Producer any] struct{}

var _ Client[float64] = dummyClient[float64]{}

func (dummyClient[Producer]) SetProvider(prod Producer) {}

func NewDummyWorkloadProjector(mailboxPathToName DynamicMapProducer[string, logicalcluster.Name]) WorkloadProjector {
	return dummyClient[ProjectionMapProducer]{}
}
