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
	"k8s.io/klog/v2"
)

func RelayWhatResolver() RelayMap[ExternalName, WorkloadParts] {
	return NewRelayMap[ExternalName, WorkloadParts](true)
}

func RelayWhereResolver() RelayMap[ExternalName, ResolvedWhere] {
	return NewRelayMap[ExternalName, ResolvedWhere](true)
}

type dummyClient[Producer any] struct{}

var _ Client[float64] = dummyClient[float64]{}

func (dummyClient[Producer]) SetProvider(prod Producer) {}

func NewLoggingWorkloadProjector(logger klog.Logger) WorkloadProjector {
	return loggingWorkloadProjector{logger}
}

type loggingWorkloadProjector struct {
	logger klog.Logger
}

func (lwp loggingWorkloadProjector) Transact(xn func(WorkloadProjectionSections)) {
	xn(WorkloadProjectionSections{
		NamespaceDistributions:          NewLoggingSetChangeReceiver[NamespaceDistributionTuple]("NamespaceDistributionTuple", lwp.logger),
		NamespacedResourceDistributions: NewLoggingSetChangeReceiver[NamespacedResourceDistributionTuple]("NamespacedResourceDistributionTuple", lwp.logger),
		NamespacedModes:                 NewLoggingMappingReceiver[ProjectionModeKey, ProjectionModeVal]("NamespacedModes", lwp.logger),
		NonNamespacedDistributions:      NewLoggingSetChangeReceiver[NonNamespacedDistributionTuple]("NonNamespacedDistributions", lwp.logger),
		NonNamespacedModes:              NewLoggingMappingReceiver[ProjectionModeKey, ProjectionModeVal]("NonNamespacedModes", lwp.logger),
	})
}

type TrivialTransactor[OpsType any] struct{ Ops OpsType }

func (tt TrivialTransactor[OpsType]) Transact(xn func(OpsType)) { xn(tt.Ops) }
