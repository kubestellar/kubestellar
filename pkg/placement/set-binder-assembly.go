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

	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

// A setBinder works as follows.
// For each EdgePlacement, it maintains:
// - a map differencer for the downsync part of the resolved "what",
// - a slice differencer for the upsync part of the resolved "what",
// - a set difference for the resolved "where".
// Those differencers feed into the downsyncJoinLeftInput and downsyncJoinRightInput, respectively.
// These drive an equijoin on the ExternalName of the EdgePlacement.
// The change stream of that equijoin feeds the SingleBinder.
type setBinder struct {
	logger klog.Logger
	sync.Mutex
	downsyncPartsDifferencerConstructor DownsyncsDifferencerConstructor
	upsyncsDifferenceConstructor        UpsyncsDifferenceConstructor
	resolvedWhereDifferencerConstructor ResolvedWhereDifferencerConstructor
	perCluster                          map[logicalcluster.Name]*setBindingForCluster
	singleBinder                        SingleBinder
	downsyncJoinLeftInput               MappingReceiver[Pair[ExternalName, WorkloadPartID], WorkloadPartDetails]
	bothJoinRightInput                  SetWriter[Pair[ExternalName, SinglePlacement]]
	upsyncJoinLeftInput                 SetWriter[Pair[ExternalName, edgeapi.UpsyncSet]]
	singleBindingOps                    SingleBindingOps
	upsyncOps                           UpsyncOps
}

type setBindingForCluster struct {
	*setBinder
	cluster      logicalcluster.Name
	perPlacement map[string]*setBindingForPlacement
}

type setBindingForPlacement struct {
	*setBindingForCluster
	downsyncPartsReceiver Receiver[WorkloadParts]
	upsyncReceiver        Receiver[[]edgeapi.UpsyncSet]
	resolvedWhereReceiver Receiver[ResolvedWhere]
}

var _ SetBinderConstructor = NewSetBinder

func NewSetBinder(
	logger klog.Logger,
	downsyncPartsDifferencerConstructor DownsyncsDifferencerConstructor,
	upsyncsDifferenceConstructor UpsyncsDifferenceConstructor,
	resolvedWhereDifferencerConstructor ResolvedWhereDifferencerConstructor,
	bindingOrganizer BindingOrganizer,
	discovery APIMapProvider,
	resourceModes ResourceModes,
	eventHandler EventHandler,
) SetBinder {
	return func(workloadProjector WorkloadProjector) (
		whatReceiver MappingReceiver[ExternalName, ResolvedWhat],
		whereReceiver MappingReceiver[ExternalName, ResolvedWhere],
	) {
		singleBinder := bindingOrganizer(discovery, resourceModes, eventHandler, workloadProjector)
		sb := &setBinder{
			logger:                              logger,
			downsyncPartsDifferencerConstructor: downsyncPartsDifferencerConstructor,
			upsyncsDifferenceConstructor:        upsyncsDifferenceConstructor,
			resolvedWhereDifferencerConstructor: resolvedWhereDifferencerConstructor,
			perCluster:                          map[logicalcluster.Name]*setBindingForCluster{},
			singleBinder:                        singleBinder,
		}

		var downsyncJoinRightInput, upsyncJoinRightInput SetWriter[Pair[ExternalName, SinglePlacement]]
		sb.downsyncJoinLeftInput, downsyncJoinRightInput = NewDynamicFullJoin12VWith13[ExternalName, WorkloadPartID, SinglePlacement, WorkloadPartDetails](sb.logger,
			NewMappingReceiverFuncs(
				func(tup Triple[ExternalName, WorkloadPartID, SinglePlacement], workloadPartDetails WorkloadPartDetails) {
					sb.logger.V(4).Info("Adding singleBinding mapping", "epRef", tup.First, "partID", tup.Second, "where", tup.Third, "details", workloadPartDetails)
					sb.singleBindingOps.Put(tup, workloadPartDetails)
				},
				func(tup Triple[ExternalName, WorkloadPartID, SinglePlacement]) {
					sb.logger.V(4).Info("Removing singleBinding mapping", "epRef", tup.First, "partID", tup.Second, "where", tup.Third)
					sb.singleBindingOps.Delete(tup)
				},
			))
		sb.upsyncJoinLeftInput, upsyncJoinRightInput = NewDynamicFullJoin12with13Parametric[ExternalName, edgeapi.UpsyncSet, SinglePlacement](sb.logger,
			HashExternalName,
			HashUpsyncSet{},
			HashSinglePlacement{},
			NewSetWriterFuncs(
				func(tup Triple[ExternalName, edgeapi.UpsyncSet, SinglePlacement]) bool {
					sb.logger.V(4).Info("Adding upsync tuple", "epRef", tup.First, "upsyncSet", tup.Second, "where", tup.Third)
					sb.upsyncOps(true, tup)
					return true
				},
				func(tup Triple[ExternalName, edgeapi.UpsyncSet, SinglePlacement]) bool {
					sb.logger.V(4).Info("Removing upsync tuple", "epRef", tup.First, "upsyncSet", tup.Second, "where", tup.Third)
					sb.upsyncOps(false, tup)
					return true
				},
			))

		sb.bothJoinRightInput = SetWriterFork(true, downsyncJoinRightInput, upsyncJoinRightInput)

		return sbAsResolvedWhatReceiver{sb}, sbAsResolvedWhereReceiver{sb}
	}
}

type sbAsResolvedWhatReceiver struct{ *setBinder }

func (sb sbAsResolvedWhatReceiver) Put(epName ExternalName, resolvedWhat ResolvedWhat) {
	sb.Lock()
	defer sb.Unlock()
	sbc := sb.getCluster(epName.Cluster, true)
	sbc.singleBinder.Transact(func(downsyncOps SingleBindingOps, upsyncOps UpsyncOps) {
		sbp := sbc.ensurePlacement(epName.Name)
		sbc.singleBindingOps = downsyncOps
		sbc.upsyncOps = upsyncOps
		sbp.downsyncPartsReceiver.Receive(resolvedWhat.Downsync)
		sbp.upsyncReceiver.Receive(resolvedWhat.Upsync)
		sbc.singleBindingOps = nil
		sbc.upsyncOps = nil
	})
}

func (sb sbAsResolvedWhatReceiver) Delete(epName ExternalName) {
	sb.Lock()
	defer sb.Unlock()
	sbc := sb.getCluster(epName.Cluster, false)
	if sbc == nil {
		return
	}
	var resolvedWhat WorkloadParts
	sbc.singleBinder.Transact(func(sbo SingleBindingOps, upsyncOps UpsyncOps) {
		sbp := sbc.ensurePlacement(epName.Name)
		sbc.singleBindingOps = sbo
		sbc.upsyncOps = upsyncOps
		sbp.downsyncPartsReceiver.Receive(resolvedWhat)
		sbp.upsyncReceiver.Receive([]edgeapi.UpsyncSet{})
		sbc.singleBindingOps = nil
		sbc.upsyncOps = nil
	})
}

type sbAsResolvedWhereReceiver struct{ *setBinder }

func (sb sbAsResolvedWhereReceiver) Put(epName ExternalName, resolvedWhere ResolvedWhere) {
	sb.Lock()
	defer sb.Unlock()
	sbc := sb.getCluster(epName.Cluster, true)
	sbc.singleBinder.Transact(func(sbo SingleBindingOps, uso UpsyncOps) {
		sbp := sbc.ensurePlacement(epName.Name)
		sb.singleBindingOps = sbo
		sb.upsyncOps = uso
		sbp.resolvedWhereReceiver.Receive(resolvedWhere)
		sb.singleBindingOps = nil
		sb.upsyncOps = nil
	})
}

func (sb sbAsResolvedWhereReceiver) Delete(epName ExternalName) {
	sb.Lock()
	defer sb.Unlock()
	sbc := sb.getCluster(epName.Cluster, false)
	if sbc == nil {
		return
	}
	var resolvedWhere ResolvedWhere
	sbc.singleBinder.Transact(func(sbo SingleBindingOps, uso UpsyncOps) {
		sbp := sbc.ensurePlacement(epName.Name)
		sb.singleBindingOps = sbo
		sb.upsyncOps = uso
		sbp.resolvedWhereReceiver.Receive(resolvedWhere)
		sb.singleBindingOps = nil
		sb.upsyncOps = nil
	})
}

func (sb *setBinder) getCluster(cluster logicalcluster.Name, want bool) *setBindingForCluster {
	sbc := sb.perCluster[cluster]
	if sbc == nil && want {
		sbc = &setBindingForCluster{
			setBinder:    sb,
			cluster:      cluster,
			perPlacement: map[string]*setBindingForPlacement{},
		}
		sb.perCluster[cluster] = sbc
	}
	return sbc
}

func (sbc *setBindingForCluster) ensurePlacement(epName string) *setBindingForPlacement {
	sbp := sbc.perPlacement[epName]
	if sbp == nil {
		epID := ExternalName{sbc.cluster, epName}
		sbp = &setBindingForPlacement{
			setBindingForCluster: sbc,
		}
		sbp.downsyncPartsReceiver = sbc.downsyncPartsDifferencerConstructor(MappingReceiverFuncs[WorkloadPartID, WorkloadPartDetails]{
			OnPut: func(partID WorkloadPartID, partDetails WorkloadPartDetails) {
				sbc.downsyncJoinLeftInput.Put(NewPair(epID, partID), partDetails)
			},
			OnDelete: func(partID WorkloadPartID) {
				sbc.downsyncJoinLeftInput.Delete(NewPair(epID, partID))
			}})
		sbp.upsyncReceiver = sbc.upsyncsDifferenceConstructor(func(add bool, upTerm edgeapi.UpsyncSet) {
			if add {
				sbc.upsyncJoinLeftInput.Add(NewPair(epID, upTerm))
			} else {
				sbc.upsyncJoinLeftInput.Remove(NewPair(epID, upTerm))
			}
		})
		sbp.resolvedWhereReceiver = sbc.resolvedWhereDifferencerConstructor(TransformSetWriter(
			NewPair1Then2[ExternalName, SinglePlacement](epID),
			sbc.bothJoinRightInput,
		))
		sbc.perPlacement[epName] = sbp
	}
	return sbp
}
