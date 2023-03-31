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
	"k8s.io/klog/v2"
)

func RelayWhatResolver() RelayMap[ExternalName, WorkloadParts] {
	return NewRelayMap[ExternalName, WorkloadParts](true)
}

func RelayWhereResolver() RelayMap[ExternalName, ResolvedWhere] {
	return NewRelayMap[ExternalName, ResolvedWhere](true)
}

func NewDummySetBinder() SetBinder {
	return dummySetBinder{NewRelayMap[ProjectionKey, *ProjectionPerCluster](true)}
}

type dummySetBinder struct {
	DynamicMapProvider[ProjectionKey, *ProjectionPerCluster]
}

func (dummySetBinder) AsWhatReceiver() MappingReceiver[ExternalName, WorkloadParts] {
	return RelayWhatResolver()
}

func (dummySetBinder) AsWhereReceiver() MappingReceiver[ExternalName, ResolvedWhere] {
	return RelayWhereResolver()
}

type dummyClient[Producer any] struct{}

var _ Client[float64] = dummyClient[float64]{}

func (dummyClient[Producer]) SetProvider(prod Producer) {}

func NewDummyWorkloadProjector(mailboxPathToName DynamicMapProvider[string, logicalcluster.Name]) WorkloadProjector {
	return dummyClient[ProjectionMapProvider]{}
}

func NewLoggingWorkloadProjector(logger klog.Logger) WorkloadProjector {
	return loggingWorkloadProjector{logger}
}

type loggingWorkloadProjector struct {
	logger klog.Logger
}

func (lwp loggingWorkloadProjector) SetProvider(pmp ProjectionMapProvider) {
	pmp.AddReceiver(lwp, true)
}

func (lwp loggingWorkloadProjector) Receive(pk ProjectionKey, ppc *ProjectionPerCluster) {
	lwp.logger.Info("Received outer projection info", "pk", pk, "ppc", ppc)
	if ppc != nil {
		ppc.PerSourceCluster.AddReceiver(loggingWorkloadProjectorMid{lwp, pk}, true)
	}
}

type loggingWorkloadProjectorMid struct {
	loggingWorkloadProjector
	pk ProjectionKey
}

func (lwpm loggingWorkloadProjectorMid) Receive(cluster logicalcluster.Name, pd ProjectionDetails) {
	lwpm.logger.Info("Received inner projection info", "pk", lwpm.pk, "cluster", cluster, "pd", pd)
}
