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

package placement

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
)

// SimpleBindingOrganizer constructs a BindingOrganizer.
// It is not so simple any more.
// See the comment on the implementation for the queries and the query plan that implement this thing.
func SimpleBindingOrganizer(logger klog.Logger) BindingOrganizer {
	return func(discovery APIMapProvider, resourceModes ResourceModes, eventHandler EventHandler, workloadProjector WorkloadProjector) SingleBinder {
		sbo := &simpleBindingOrganizer{
			logger:            logger,
			discovery:         discovery,
			resourceModes:     resourceModes,
			eventHandler:      eventHandler,
			workloadProjector: workloadProjector,
			perSourceCluster:  NewMapMap[string, *simpleBindingPerCluster](nil),
		}
		namespacedModesReceiver := MappingReceiverFuncs[ProjectionModeKey, ProjectionModeVal]{
			OnPut: func(mk ProjectionModeKey, val ProjectionModeVal) {
				logger.V(4).Info("NamespacedModes.Put", "key", mk, "val", val)
				sbo.workloadProjectionSections.NamespacedModes.Put(mk, val)
			},
			OnDelete: func(mk ProjectionModeKey) {
				logger.V(4).Info("NamespacedModes.Delete", "key", mk)
				sbo.workloadProjectionSections.NamespacedModes.Delete(mk)
			},
		}

		sbo.resourceDiscoveryReceiver = NewMappingReceiverFuncs(
			func(key Pair[string, metav1.GroupResource], val ProjectionModeVal) {
				// TODO: implement version agreement
			},
			func(key Pair[string, metav1.GroupResource]) {
				// TODO: implement version agreement
			})

		namespacedObjectDistributions := NewMappingReceiverFuncs(
			func(key NamespacedDistributionTuple, val WorkloadPartDetails) {
				logger.V(4).Info("NamespacedDistributionTuple added", "key", key, "returnSingleton", val.ReturnSingletonState)
				sbo.workloadProjectionSections.NamespacedObjectDistributions.Put(key, val.distributionBits())
			},
			func(key NamespacedDistributionTuple) {
				logger.V(4).Info("NamespacedDistributionTuple removed", "key", key)
				sbo.workloadProjectionSections.NamespacedObjectDistributions.Delete(key)
			})
		aggregateForNamespaced := NewFactoredMapMapAggregator[NamespacedDistributionTuple, ProjectionModeKey, ExternalNamespacedName /*of downsynced object*/, ProjectionModeVal, ProjectionModeVal](
			PairFactorer[ProjectionModeKey, ExternalNamespacedName /*of downsynced object*/](),
			nil,
			nil,
			pickThe1[ProjectionModeKey, ExternalNamespacedName /*of downsynced object*/](sbo.logger, "Not implemented yet: handling version conflicts for cluster-scoped resources"),
			namespacedModesReceiver,
		)

		namespacedChangeReceiver := MappingReceiverFuncs[NamespacedDistributionTuple, Map[ObjectName /*epName*/, WorkloadPartDetails]]{
			OnPut: func(nndt NamespacedDistributionTuple, versions Map[ObjectName /*epName*/, WorkloadPartDetails]) {
				version, ok := detailsSansEP(versions)
				if !ok {
					sbo.logger.Error(nil, "Impossible: addition of empty version map", "nndt", nndt)
				}
				namespacedObjectDistributions.Put(nndt, version)
				aggregateForNamespaced.Put(nndt, ProjectionModeVal{APIVersion: version.APIVersion})
			},
			OnDelete: func(nndt NamespacedDistributionTuple) {
				namespacedObjectDistributions.Delete(nndt)
				aggregateForNamespaced.Delete(nndt)
			},
		}

		sbo.namespacedWhatWhereFull = NewFactoredMapMap[NamespacedWhatWhereFullKey, NamespacedDistributionTuple, ObjectName /* ep name */, WorkloadPartDetails](
			factorNamespacedWhatWhereFullKey,
			nil,
			nil,
			namespacedChangeReceiver,
		)

		clusterObjectDistributionsReceiver := NewMappingReceiverFuncs(
			func(key NonNamespacedDistributionTuple, val WorkloadPartDetails) {
				sbo.workloadProjectionSections.NonNamespacedObjectDistributions.Put(key, val.distributionBits())
			},
			func(key NonNamespacedDistributionTuple) {
				sbo.workloadProjectionSections.NonNamespacedObjectDistributions.Delete(key)
			},
		)
		clusterModesReceiver := MappingReceiverFuncs[ProjectionModeKey, ProjectionModeVal]{
			OnPut: func(mk ProjectionModeKey, val ProjectionModeVal) {
				sbo.workloadProjectionSections.NonNamespacedModes.Put(mk, val)
			},
			OnDelete: func(mk ProjectionModeKey) {
				sbo.workloadProjectionSections.NonNamespacedModes.Delete(mk)
			},
		}
		// aggregateForCluster is driven by the change stream of the map with epName and returnSingleton projected out,
		// and does a GROUP BY ProjectionModeKey and then aggregates over ExternalName (of downsynced object) solving the version conflicts (if any).
		aggregateForCluster := NewFactoredMapMapAggregator[NonNamespacedDistributionTuple, ProjectionModeKey, ExternalName /*of downsynced object*/, ProjectionModeVal, ProjectionModeVal](
			PairFactorer[ProjectionModeKey, ExternalName /*of downsynced object*/](),
			nil,
			nil,
			pickThe1[ProjectionModeKey, ExternalName /*of downsynced object*/](sbo.logger, "Not implemented yet: handling version conflicts for cluster-scoped resources"),
			clusterModesReceiver,
		)
		// clusterChangeReceiver receives the change stream of the full map and projects out the EdgePlacement
		// object name to feed to sansEPName.
		// This is relatively simple because the API version does not vary for a given resource and source cluster.
		clusterChangeReceiver := MappingReceiverFuncs[NonNamespacedDistributionTuple, Map[ObjectName /*epName*/, WorkloadPartDetails]]{
			OnPut: func(nndt NonNamespacedDistributionTuple, versions Map[ObjectName /*epName*/, WorkloadPartDetails]) {
				version, ok := detailsSansEP(versions)
				if !ok {
					sbo.logger.Error(nil, "Impossible: addition of empty version map", "nndt", nndt)
				}
				clusterObjectDistributionsReceiver.Put(nndt, version)
				aggregateForCluster.Put(nndt, ProjectionModeVal{version.APIVersion})
			},
			OnDelete: func(nndt NonNamespacedDistributionTuple) {
				clusterObjectDistributionsReceiver.Delete(nndt)
				aggregateForCluster.Delete(nndt)
			},
		}
		// clusterWhatWhereFull is a map from ClusterWhatWhereFullKey to API version (no group),
		// factored into a map from NonNamespacedDistributionTuple to epName to WorkloadPartDetails.
		sbo.clusterWhatWhereFull = NewFactoredMapMap[ClusterWhatWhereFullKey, NonNamespacedDistributionTuple, ObjectName /* ep name */, WorkloadPartDetails](
			factorClusterWhatWhereFullKey,
			nil,
			nil,
			clusterChangeReceiver,
		)

		upsyncsRelay := NewSetWriterFuncs(
			func(tup Pair[SinglePlacement, edgeapi.UpsyncSet]) bool {
				logger.V(4).Info("Upsyncs added", "tuple", tup)
				return sbo.workloadProjectionSections.Upsyncs.Add(tup)
			},
			func(tup Pair[SinglePlacement, edgeapi.UpsyncSet]) bool {
				logger.V(4).Info("Upsyncs removed", "tuple", tup)
				return sbo.workloadProjectionSections.Upsyncs.Remove(tup)
			})
		sbo.upsyncsFull = NewSetChangeProjectorByHashMap(
			factorUpsyncTuple,
			upsyncsRelay,
			PairHashDomain[SinglePlacement, edgeapi.UpsyncSet](HashSinglePlacement{}, HashUpsyncSet{}),
			HashExternalName)
		return sbo
	}
}

func detailsSansEP(versions Map[ObjectName /*epName*/, WorkloadPartDetails]) (WorkloadPartDetails, bool) {
	var ans WorkloadPartDetails
	if versions.IsEmpty() {
		return ans, false
	}
	versions.Visit(func(pair Pair[ObjectName /*epName*/, WorkloadPartDetails]) error {
		ans.APIVersion = pair.Second.APIVersion // should be the same for every pair in versions
		ans.ReturnSingletonState = ans.ReturnSingletonState || pair.Second.ReturnSingletonState
		ans.CreateOnly = ans.CreateOnly || pair.Second.CreateOnly
		return nil
	})
	return ans, true
}

var factorUpsyncTuple = NewFactorer(
	func(whole Triple[ExternalName /* of EdgePlacement object */, edgeapi.UpsyncSet, SinglePlacement]) Pair[Pair[SinglePlacement, edgeapi.UpsyncSet], ExternalName /* of EdgePlacement object */] {
		return NewPair(NewPair(whole.Third, whole.Second), whole.First)
	},
	func(parts Pair[Pair[SinglePlacement, edgeapi.UpsyncSet], ExternalName /* of EdgePlacement object */]) Triple[ExternalName /* of EdgePlacement object */, edgeapi.UpsyncSet, SinglePlacement] {
		return NewTriple(parts.Second, parts.First.Second, parts.First.First) // cdr, cdar, caar
	})

// simpleBindingOrganizer is the top-level data structure of the organizer.
// In the locking order it precedes its discovery and its projectionMapProvider,
// which in turn precedes each projectionPerClusterImpl.
//
// This thing is implemented in relational aglebra style.
// It works with change streams of relations, mainly in the passive voice.
// That is, the pipefitting is in terms of receivers of changes.
//
// The namespaced and non-namespaced (AKA cluster-scoped) resources are handled
// separately.
//
// The designed version agreement algorithm belongs in here, but is not
// implemented yet; instead there is a hack.
// TODO: implement the designed version agreement.
//
// For the namespaced resources, as a SingleBinder this organizer
// is given the stream of changes to the following map:
// - WhatWheres: map of ((epCluster,epName),(GroupResource,NSName,ObjName),destination) -> (APIVersion, DistrBits)
// and produces the change streams to the following two maps:
// - map of NamespacedObjectDistributions ((GroupResource,destination), (epCluster,NSName,ObjName)) -> DistrBits
// - map of ProjectionModeKey (GroupResource,destination) -> ProjectionModeVal (APIVersion).
//
// As a relational algebra expression, the desired computation is as follows.
// common = WhatWheres.GroupBy(all but epName).Aggregate(OR[DistrBits])
// (that's a map (epCluster,(GroupResource,NSName,ObjName),destination) -> (APIVersion, DistrBits))
// NamespacedObjectDistributions = common.ProjectOut(APIVersion)
// ProjectionModes = common.GroupBy(GroupResource,destination).Aggregate(PickVersion)
//
// The query plan is as follows.
// nsModesReceiver <- nsSansEPName.GroupBy(GroupResource,destination).Aggregate(PickVersion)
// nsObjectDistributions <- nsSansEPName.ProjectOut(APIVersion)
// nsSansEPName <- WhatWheres.ProjectOut(epName)
//
// For the cluster-scoped resources, as a SingleBinder this organizer
// is given the stream of changes to the following map:
// - WhatWheres: map of ((epCluster,epName),GroupResource,ObjName,destination) -> (APIVersion, DistrBits)
// and produces the change streams to the following two maps:
// - map of NonNamespacedObjectDistribution (epCluster,GroupResource,ObjName,destination) -> DistrBits
// - map of ProjectionModeKey (GroupResource,destination) -> ProjectionModeVal (APIVersion).
//
// As a relational algebra expression, the desired computation is as follows.
// common = WhatWheres.GroupBy(all but epName).Aggregate(OR[DistrBits])
// (that's a map (epCluster,GroupResource,ObjName,destination) -> (APIVersion, DistrBits))
// // NonNamespacedDistributionTuples = common.Keys()
// NonNamespacedObjectDistributions = common.ProjectOut(APIVersion)
// ProjectionModes = common.GroupBy(GroupResource,destination).Aggregate(PickVersion)
//
// The query plan is as follows.
// clusterModesReceiver <- ctSansEPName.GroupBy(GroupResource,destination).Aggregate(PickVersion)
// clusterObjectDistributions <- ctSansEPName.ProjectOut(APIVersion)
// ctSansEPName <- WhatWheres.GroupBy(all but epName).Aggregate(OR[DistrBits])
//
// For the upsyncs, as a SingleBinder this organizer is given the stream of changes
// to the following relation:
// - WhatWheres: set of ((epCluster,epName),UpsyncSet,destination)
// and produces the change stream to the following relation:
// - upsyncs: set of (UpsyncSet,destination)
//
// In relational algebra, the desired computation is as follows.
// upsyncs = WhatWheres.ProjectOut((epCluster,epName))
//
// The query plan is as follows.
// upsyncsRelay <- WhatWheres.ProjectOut((epCluster,epName))
type simpleBindingOrganizer struct {
	logger        klog.Logger
	discovery     APIMapProvider
	resourceModes ResourceModes
	eventHandler  EventHandler

	workloadProjector WorkloadProjector

	sync.Mutex

	perSourceCluster MutableMap[string, *simpleBindingPerCluster]

	workloadProjectionSections WorkloadProjectionSections // non-zero only during a transaction!

	// The following fields hold the same value throughout the lifetime of this object,
	// but those values use workloadProjectionSections --- synchronously --- and so can
	// only be invoked during a transaction.

	clusterWhatWhereFull      MappingReceiver[ClusterWhatWhereFullKey, WorkloadPartDetails]
	namespacedWhatWhereFull   MappingReceiver[NamespacedWhatWhereFullKey, WorkloadPartDetails]
	upsyncsFull               SetWriter[Triple[ExternalName /* of EdgePlacement object */, edgeapi.UpsyncSet, SinglePlacement]]
	resourceDiscoveryReceiver MappingReceiver[ResourceDiscoveryKey, ProjectionModeVal]
}

var factorNamespacedJoinKeyLessNS = Factorer[NamespacedJoinKeyLessnS, ProjectionModeKey, string]{
	First: func(whole NamespacedJoinKeyLessnS) Pair[ProjectionModeKey, string] {
		return Pair[ProjectionModeKey, string]{
			First:  ProjectionModeKey{whole.Second, whole.Third},
			Second: whole.First}
	},
	Second: func(parts Pair[ProjectionModeKey, string]) NamespacedJoinKeyLessnS {
		return NamespacedJoinKeyLessnS{
			First:  parts.Second,
			Second: parts.First.GroupResource,
			Third:  parts.First.Destination}
	},
}

func pickThe1[KeyPartA, KeyPartB comparable](logger klog.Logger, errmsg string) func(keyPartA KeyPartA, problem Map[KeyPartB, ProjectionModeVal]) ProjectionModeVal {
	return func(keyPartA KeyPartA, problem Map[KeyPartB, ProjectionModeVal]) ProjectionModeVal {
		versions := NewMapSet[ProjectionModeVal]()
		var solution ProjectionModeVal
		problem.Visit(func(pair Pair[KeyPartB, ProjectionModeVal]) error {
			versions.Add(pair.Second)
			solution = pair.Second
			return nil
		})
		if versions.Len() != 1 {
			logger.Error(nil, errmsg, "keyPartA", keyPartA, "problem", problem, "chosen", solution)
		}
		return solution
	}
}

var factorNamespacedWhatWhereFullKey = Factorer[NamespacedWhatWhereFullKey, NamespacedDistributionTuple, ObjectName /*epName*/]{
	First: func(nfk NamespacedWhatWhereFullKey) Pair[NamespacedDistributionTuple, ObjectName /*epName*/] {
		return Pair[NamespacedDistributionTuple, ObjectName /*epName*/]{
			First: NewPair(
				ProjectionModeKey{nfk.Second.First, nfk.Third},
				NewTriple(nfk.First.Cluster, nfk.Second.Second, nfk.Second.Third)),
			Second: nfk.First.Name}
	},
	Second: func(parts Pair[NamespacedDistributionTuple, ObjectName /*epName*/]) NamespacedWhatWhereFullKey {
		return NamespacedWhatWhereFullKey{
			First:  ExternalName{parts.First.Second.First, parts.Second},
			Second: NewTriple(parts.First.First.GroupResource, parts.First.Second.Second, parts.First.Second.Third),
			Third:  parts.First.First.Destination}
	},
}

// factorClusterWhatWhereFullKey factors a ClusterWhatWhereFullKey into
// a (NonNamespacedDistributionTuple, string/*ep name*/) pair.
var factorClusterWhatWhereFullKey = Factorer[ClusterWhatWhereFullKey, NonNamespacedDistributionTuple, ObjectName /* ep name */]{
	First: func(cfk ClusterWhatWhereFullKey) Pair[NonNamespacedDistributionTuple, ObjectName /* ep name */] {
		return Pair[NonNamespacedDistributionTuple, ObjectName /* ep name */]{
			First: NonNamespacedDistributionTuple{
				First: ProjectionModeKey{
					GroupResource: cfk.Second.First,
					Destination:   cfk.Third,
				},
				Second: ExternalName{
					Cluster: cfk.First.Cluster,
					Name:    cfk.Second.Second,
				},
			},
			Second: cfk.First.Name}
	},
	Second: func(decomp Pair[NonNamespacedDistributionTuple, ObjectName /*ep name*/]) ClusterWhatWhereFullKey {
		return ClusterWhatWhereFullKey{
			First:  ExternalName{decomp.First.Second.Cluster, decomp.Second},
			Second: Pair[metav1.GroupResource, ObjectName]{decomp.First.First.GroupResource, decomp.First.Second.Name},
			Third:  decomp.First.First.Destination,
		}
	},
}

type ResourceDiscoveryKey = Pair[string /*wmw*/, metav1.GroupResource]

type NamespacedWhatWhereFullKey = Triple[ExternalName, WorkloadPartID, SinglePlacement]

// ClusterWhatWhereFullKey is (EdgePlacement id, (resource, object name), destination)
type ClusterWhatWhereFullKey = Triple[ExternalName, Pair[metav1.GroupResource, ObjectName], SinglePlacement]

func (sbo *simpleBindingOrganizer) Transact(xn func(SingleBindingOps, UpsyncOps)) {
	sbo.Lock()
	defer sbo.Unlock()
	sbo.logger.V(3).Info("Begin transaction")
	sbo.workloadProjector.Transact(func(wps WorkloadProjectionSections) {
		sbo.workloadProjectionSections = wps
		xn(sboXnOps{sbo}, sbo.receiveUpsyncChange)
		sbo.workloadProjectionSections = WorkloadProjectionSections{}
	})
	sbo.logger.V(3).Info("End transaction")
}

func (sbo *simpleBindingOrganizer) receiveUpsyncChange(add bool, tup Triple[ExternalName /* of EdgePlacement object */, edgeapi.UpsyncSet, SinglePlacement]) {
	if add {
		sbo.upsyncsFull.Add(tup)
	} else {
		sbo.upsyncsFull.Remove(tup)
	}
}

// sboXnOps exposes the SingleBindingOps behavior, in the locked context of a transaction
type sboXnOps struct {
	sbo *simpleBindingOrganizer
}

func (sxo sboXnOps) Put(tup Triple[ExternalName, WorkloadPartID, SinglePlacement], val WorkloadPartDetails) {
	sbo := sxo.sbo
	rscMode := sbo.resourceModes(tup.Second.First)
	if !rscMode.GoesToMailbox() {
		sbo.logger.V(4).Info("Ignoring WhatWhere tuple because it does not go to the mailbox workspaces", "tup", tup, "rscMode", rscMode)
		return
	}
	sbo.getSourceCluster(tup.First.Cluster, true)
	gr := tup.Second.First
	namespaced := len(tup.Second.Second) > 0
	if namespaced {
		sbo.namespacedWhatWhereFull.Put(tup, val)
	} else {
		key := ClusterWhatWhereFullKey{tup.First, Pair[metav1.GroupResource, ObjectName]{gr, tup.Second.Third}, tup.Third}
		sbo.clusterWhatWhereFull.Put(key, val)
	}
}

func (sxo sboXnOps) Delete(tup Triple[ExternalName, WorkloadPartID, SinglePlacement]) {
	sbo := sxo.sbo
	rscMode := sbo.resourceModes(tup.Second.First)
	if !rscMode.GoesToMailbox() {
		sbo.logger.V(4).Info("Ignoring WhatWhere tuple because it does not go to the mailbox workspaces", "tup", tup, "rscMode", rscMode)
		return
	}
	sbc := sbo.getSourceCluster(tup.First.Cluster, false)
	if sbc == nil {
		return
	}
	gr := tup.Second.First
	namespaced := len(tup.Second.Second) > 0
	if namespaced /* && !val.IncludeNamespaceObject */ {
		sbo.namespacedWhatWhereFull.Delete(tup)
	}
	key := ClusterWhatWhereFullKey{tup.First, Pair[metav1.GroupResource, ObjectName]{gr, tup.Second.Third}, tup.Third}
	sbo.clusterWhatWhereFull.Delete(key)
	if false {
		// TODO: make this happen iff there is no remaining data for the cluster
		sbo.logger.V(4).Info("Removing discovery receivers", "cluster", sbc.cluster)
		sbo.discovery.RemoveReceivers(sbc.cluster, &sbc.groupReceiver, &sbc.resourceReceiver)
	}
}

func (sbo *simpleBindingOrganizer) getSourceCluster(cluster string, want bool) *simpleBindingPerCluster {
	sbc, have := sbo.perSourceCluster.Get(cluster)
	if want && !have {
		sbc = &simpleBindingPerCluster{
			simpleBindingOrganizer: sbo,
			cluster:                cluster,
		}
		sbo.perSourceCluster.Put(cluster, sbc)
		sbc.groupReceiver.MappingReceiver = sbcGroupReceiver{sbc}
		sbc.resourceReceiver.MappingReceiver = sbcResourceReceiver{sbc}
		sbo.logger.V(4).Info("Adding discovery receivers", "cluster", sbc.cluster)
		sbo.discovery.AddReceivers(cluster, &sbc.groupReceiver, &sbc.resourceReceiver)
	}
	return sbc
}

type simpleBindingPerCluster struct {
	*simpleBindingOrganizer
	cluster          string
	groupReceiver    MappingReceiverHolder[string /*group name*/, APIGroupInfo]
	resourceReceiver MappingReceiverHolder[metav1.GroupResource, ResourceDetails]
}

type sbcGroupReceiver struct {
	sbc *simpleBindingPerCluster
}

func (sgr sbcGroupReceiver) Put(group string, info APIGroupInfo) {
	// TODO: something once apiwatch supplies this info
}

func (sgr sbcGroupReceiver) Delete(group string) {
	// TODO: something once apiwatch supplies this info
}

type sbcResourceReceiver struct {
	sbc *simpleBindingPerCluster
}

func (srr sbcResourceReceiver) Put(gr metav1.GroupResource, details ResourceDetails) {
	sbc := srr.sbc
	good := details.Namespaced && details.SupportsInformers
	key := NewPair(sbc.cluster, gr)
	sbc.Lock()
	defer sbc.Unlock()
	sbc.workloadProjector.Transact(func(ops WorkloadProjectionSections) {
		sbc.workloadProjectionSections = ops
		sbc.logger.V(4).Info("sbcResourceReceiver.Put", "cluster", sbc.cluster, "gr", gr, "good", good, "details", details)
		if good {
			sbc.resourceDiscoveryReceiver.Put(key, ProjectionModeVal{details.PreferredVersion})
		} else {
			sbc.resourceDiscoveryReceiver.Delete(key)
		}
		sbc.workloadProjectionSections = WorkloadProjectionSections{}
	})
}

func (srr sbcResourceReceiver) Delete(gr metav1.GroupResource) {
	sbc := srr.sbc
	key := NewPair(sbc.cluster, gr)
	sbc.Lock()
	defer sbc.Unlock()
	sbc.workloadProjector.Transact(func(ops WorkloadProjectionSections) {
		sbc.workloadProjectionSections = ops
		sbc.logger.V(4).Info("sbcResourceReceiver.Delete", "cluster", sbc.cluster, "gr", gr)
		sbc.resourceDiscoveryReceiver.Delete(key)
		sbc.workloadProjectionSections = WorkloadProjectionSections{}
	})
}

func mgrIsNamespace(gr metav1.GroupResource) bool {
	return gr.Group == "" && gr.Resource == "namespaces"
}
