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
// For each EdgePlacement, it maintains a map differencer for the resolved "what"
// and a set difference for the resolved "where".
// Those differencers feed into the join12v and join13, respectively.
// These drive an equijoin on the ExternalName of the EdgePlacement.
// The change stream of that equijoin feeds the SingleBinder.
type setBinder struct {
	logger klog.Logger
	sync.Mutex
	resolvedWhatDifferencerConstructor  ResolvedWhatDifferencerConstructor
	resolvedWhereDifferencerConstructor ResolvedWhereDifferencerConstructor
	perCluster                          map[logicalcluster.Name]*setBindingForCluster
	singleBinder                        SingleBinder
	join12v                             MappingReceiver[Pair[ExternalName, WorkloadPartID], WorkloadPartDetails]
	join13                              SetChangeReceiver[Pair[ExternalName, edgeapi.SinglePlacement]]
	singleBindingOps                    SingleBindingOps
}

type setBindingForCluster struct {
	*setBinder
	cluster      logicalcluster.Name
	perPlacement map[string]*setBindingForPlacement
}

type setBindingForPlacement struct {
	*setBindingForCluster
	resolvedWhatReceiver  Receiver[WorkloadParts]
	resolvedWhereReceiver Receiver[ResolvedWhere]
}

var _ SetBinderConstructor = NewSetBinder

func NewSetBinder(
	logger klog.Logger,
	resolvedWhatDifferencerConstructor ResolvedWhatDifferencerConstructor,
	resolvedWhereDifferencerConstructor ResolvedWhereDifferencerConstructor,
	bindingOrganizer BindingOrganizer,
	discovery APIMapProvider,
	resourceModes ResourceModes,
	eventHandler EventHandler,
) SetBinder {
	return func(workloadProjector WorkloadProjector) (
		whatReceiver MappingReceiver[ExternalName, WorkloadParts],
		whereReceiver MappingReceiver[ExternalName, ResolvedWhere],
	) {
		singleBinder := bindingOrganizer(discovery, resourceModes, eventHandler, workloadProjector)
		sb := &setBinder{
			logger:                              logger,
			resolvedWhatDifferencerConstructor:  resolvedWhatDifferencerConstructor,
			resolvedWhereDifferencerConstructor: resolvedWhereDifferencerConstructor,
			perCluster:                          map[logicalcluster.Name]*setBindingForCluster{},
			singleBinder:                        singleBinder,
		}

		sb.join12v, sb.join13 = NewDynamicFullJoin12VWith13[ExternalName, WorkloadPartID, edgeapi.SinglePlacement, WorkloadPartDetails](sb.logger, sb)

		return sbAsResolvedWhatReceiver{sb}, sbAsResolvedWhereReceiver{sb}
	}
}

type sbAsResolvedWhatReceiver struct{ *setBinder }

func (sb sbAsResolvedWhatReceiver) Put(epName ExternalName, resolvedWhat WorkloadParts) {
	sb.Lock()
	defer sb.Unlock()
	sbc := sb.getCluster(epName.Cluster, true)
	sbc.singleBinder.Transact(func(sbo SingleBindingOps) {
		sbp := sbc.ensurePlacement(epName.Name)
		sbc.singleBindingOps = sbo
		sbp.resolvedWhatReceiver.Receive(resolvedWhat)
		sbc.singleBindingOps = nil
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
	sbc.singleBinder.Transact(func(sbo SingleBindingOps) {
		sbp := sbc.ensurePlacement(epName.Name)
		sbc.singleBindingOps = sbo
		sbp.resolvedWhatReceiver.Receive(resolvedWhat)
		sbc.singleBindingOps = nil
	})
}

type sbAsResolvedWhereReceiver struct{ *setBinder }

func (sb sbAsResolvedWhereReceiver) Put(epName ExternalName, resolvedWhere ResolvedWhere) {
	sb.Lock()
	defer sb.Unlock()
	sbc := sb.getCluster(epName.Cluster, true)
	sbc.singleBinder.Transact(func(sbo SingleBindingOps) {
		sbp := sbc.ensurePlacement(epName.Name)
		sb.singleBindingOps = sbo
		sbp.resolvedWhereReceiver.Receive(resolvedWhere)
		sb.singleBindingOps = nil
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
	sbc.singleBinder.Transact(func(sbo SingleBindingOps) {
		sbp := sbc.ensurePlacement(epName.Name)
		sb.singleBindingOps = sbo
		sbp.resolvedWhereReceiver.Receive(resolvedWhere)
		sb.singleBindingOps = nil
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

func (sb *setBinder) Put(tup Triple[ExternalName, WorkloadPartID, edgeapi.SinglePlacement], workloadPartDetails WorkloadPartDetails) {
	sb.logger.V(4).Info("Adding joined mapping", "epRef", tup.First, "partID", tup.Second, "where", tup.Third, "details", workloadPartDetails)
	sb.singleBindingOps.Put(tup, workloadPartDetails)
}

func (sb *setBinder) Delete(tup Triple[ExternalName, WorkloadPartID, edgeapi.SinglePlacement]) {
	sb.logger.V(4).Info("Removing joined mapping", "epRef", tup.First, "partID", tup.Second, "where", tup.Third)
	sb.singleBindingOps.Delete(tup)
}

func (sbc *setBindingForCluster) ensurePlacement(epName string) *setBindingForPlacement {
	sbp := sbc.perPlacement[epName]
	if sbp == nil {
		epID := ExternalName{sbc.cluster, epName}
		sbp = &setBindingForPlacement{
			setBindingForCluster: sbc,
		}
		sbp.resolvedWhatReceiver = sbc.resolvedWhatDifferencerConstructor(MappingReceiverFuncs[WorkloadPartID, WorkloadPartDetails]{
			OnPut: func(partID WorkloadPartID, partDetails WorkloadPartDetails) {
				sbc.join12v.Put(NewPair(epID, partID), partDetails)
			},
			OnDelete: func(partID WorkloadPartID) {
				sbc.join12v.Delete(NewPair(epID, partID))
			}})
		sbp.resolvedWhereReceiver = sbc.resolvedWhereDifferencerConstructor(TransformSetChangeReceiver(
			NewPair1Then2[ExternalName, edgeapi.SinglePlacement](epID),
			sbc.join13,
		))
		sbc.perPlacement[epName] = sbp
	}
	return sbp
}
