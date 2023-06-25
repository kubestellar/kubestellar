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

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v1alpha1"
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
			perSourceCluster:  NewMapMap[logicalcluster.Name, *simpleBindingPerCluster](nil),
		}
		namespaceDistributionsRelay := SetWriterFuncs[NamespaceDistributionTuple]{
			OnAdd: func(tup NamespaceDistributionTuple) bool {
				logger.V(4).Info("NamespaceDistributionTuple added", "tuple", tup)
				return sbo.workloadProjectionSections.NamespaceDistributions.Add(tup)
			},
			OnRemove: func(tup NamespaceDistributionTuple) bool {
				logger.V(4).Info("NamespaceDistributionTuple removed", "tuple", tup)
				return sbo.workloadProjectionSections.NamespaceDistributions.Remove(tup)
			}}
		namespacedResourceDistributionRelay := SetWriterFuncs[Triple[logicalcluster.Name, metav1.GroupResource, SinglePlacement]]{
			OnAdd: func(tup Triple[logicalcluster.Name, metav1.GroupResource, SinglePlacement]) bool {
				dist := NamespacedResourceDistributionTuple{tup.First, ProjectionModeKey{tup.Second, tup.Third}}
				logger.V(4).Info("NamespacedResourceDistributionTuple added", "tuple", tup)
				return sbo.workloadProjectionSections.NamespacedResourceDistributions.Add(dist)
			},
			OnRemove: func(tup Triple[logicalcluster.Name, metav1.GroupResource, SinglePlacement]) bool {
				dist := NamespacedResourceDistributionTuple{tup.First, ProjectionModeKey{tup.Second, tup.Third}}
				logger.V(4).Info("NamespacedResourceDistributionTuple removed", "tuple", tup)
				return sbo.workloadProjectionSections.NamespacedResourceDistributions.Remove(dist)
			},
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

		nsToAggregate := NewFactoredMapMapAggregator[Triple[logicalcluster.Name, metav1.GroupResource, SinglePlacement],
			ProjectionModeKey, logicalcluster.Name, ProjectionModeVal, ProjectionModeVal](
			factorNamespacedJoinKeyLessNS,
			nil,
			nil,
			pickThe1[ProjectionModeKey, logicalcluster.Name](sbo.logger, "Should not happen: version difference between namespaces"),
			namespacedModesReceiver,
		)

		nsCommon := MappingReceiverFork[Triple[logicalcluster.Name, metav1.GroupResource, SinglePlacement], ProjectionModeVal]{
			NewLoggingMappingReceiver[Triple[logicalcluster.Name, metav1.GroupResource, SinglePlacement], ProjectionModeVal]("nsCommon", logger.V(4)),
			MapKeySetReceiverLossy[Triple[logicalcluster.Name, metav1.GroupResource, SinglePlacement], ProjectionModeVal](namespacedResourceDistributionRelay),
			nsToAggregate}

		rscDisco, nsSrcAndDest := NewDynamicFullJoin12VWith13[logicalcluster.Name, metav1.GroupResource, SinglePlacement, ProjectionModeVal](
			logger, nsCommon)
		// sbo.resourceDiscoveryReceiver = rscDisco
		sbo.resourceDiscoveryReceiver = NewMappingReceiverFuncs(
			func(key Pair[logicalcluster.Name, metav1.GroupResource], val ProjectionModeVal) {
				rscMode := resourceModes(key.Second)
				if rscMode.GoesToMailbox() {
					logger.V(4).Info("Binder got namespaced resource", "key", key, "val", val)
					rscDisco.Put(key, val)
				} else {
					logger.V(4).Info("Ignoring namespaced resource because it does not propagate", "key", key, "val", val)
				}
			},
			func(key Pair[logicalcluster.Name, metav1.GroupResource]) {
				rscMode := resourceModes(key.Second)
				if rscMode.GoesToMailbox() {
					logger.V(4).Info("Binder told there is no such namespaced resource", "key", key)
					rscDisco.Delete(key)
				} else {
					logger.V(4).Info("Ignoring resource because it does not propagate", "key", key)
				}
			})
		nsSrcAndDestAndLog := NewSetWriterFuncs( // logging and the ability to set breakpoints
			func(elt Pair[logicalcluster.Name, SinglePlacement]) bool {
				news := nsSrcAndDest.Add(elt)
				logger.V(4).Info("Add to nsSrcAndDest", "tuple", elt, "news", news)
				return news
			},
			func(elt Pair[logicalcluster.Name, SinglePlacement]) bool {
				news := nsSrcAndDest.Remove(elt)
				logger.V(4).Info("Remove from nsSrcAndDest", "tuple", elt, "news", news)
				return news
			},
		)

		nsToGoLoseNamespace := NewSetChangeProjectorByMapMap(
			TripleFactorerTo13and2[logicalcluster.Name, NamespaceName, SinglePlacement](), nsSrcAndDestAndLog)

		nsToGo := SetWriterFork[NamespaceDistributionTuple](false, namespaceDistributionsRelay, nsToGoLoseNamespace)

		sbo.namespacedWhatWhereFull = NewSetChangeProjectorByMapMap(
			factorNamespacedWhatWhereFullKey, nsToGo)

		clusterDistributionsReceiver := SetWriterFuncs[NonNamespacedDistributionTuple]{
			OnAdd: func(nnd NonNamespacedDistributionTuple) bool {
				return sbo.workloadProjectionSections.NonNamespacedDistributions.Add(nnd)
			},
			OnRemove: func(nnd NonNamespacedDistributionTuple) bool {
				return sbo.workloadProjectionSections.NonNamespacedDistributions.Remove(nnd)
			},
		}
		clusterModesReceiver := MappingReceiverFuncs[ProjectionModeKey, ProjectionModeVal]{
			OnPut: func(mk ProjectionModeKey, val ProjectionModeVal) {
				sbo.workloadProjectionSections.NonNamespacedModes.Put(mk, val)
			},
			OnDelete: func(mk ProjectionModeKey) {
				sbo.workloadProjectionSections.NonNamespacedModes.Delete(mk)
			},
		}
		// aggregateForCluster is driven by the change stream of the map with epName projected out,
		// and does a GROUP BY ProjectionModeKey and then aggregates over ExternalName (of downsynced object) solving the version conflicts (if any).
		aggregateForCluster := NewFactoredMapMapAggregator[NonNamespacedDistributionTuple, ProjectionModeKey, ExternalName /*of downsynced object*/, ProjectionModeVal, ProjectionModeVal](
			PairFactorer[ProjectionModeKey, ExternalName /*of downsynced object*/](),
			nil,
			nil,
			pickThe1[ProjectionModeKey, ExternalName /*of downsynced object*/](sbo.logger, "Not implemented yet: handling version conflicts for cluster-scoped resources"),
			clusterModesReceiver,
		)
		// ctSansEPName receives the change stream of clusterWhatWhereFull with epName projected out,
		// and passes it along to clusterDistributionsReceiver and aggregateForCluster.
		var ctSansEPName MappingReceiver[NonNamespacedDistributionTuple, ProjectionModeVal] = MappingReceiverFork[NonNamespacedDistributionTuple, ProjectionModeVal]{
			MapKeySetReceiverLossy[NonNamespacedDistributionTuple, ProjectionModeVal](clusterDistributionsReceiver),
			aggregateForCluster,
		}
		pickVersionForEP := func(versions Map[string /*epName*/, ProjectionModeVal]) (ProjectionModeVal, bool) {
			var version ProjectionModeVal
			if versions.Visit(func(pair Pair[string /*epName*/, ProjectionModeVal]) error {
				version = pair.Second
				return errStop
			}) == nil {
				return version, false
			}
			return version, true
		}
		// clusterChangeReceiver receives the change stream of the full map and projects out the EdgePlacement
		// object name to feed to sansEPName.
		// This is relatively simple because the API version does not vary for a given resource and source cluster.
		clusterChangeReceiver := MappingReceiverFuncs[NonNamespacedDistributionTuple, Map[string /*epName*/, ProjectionModeVal]]{
			OnPut: func(nndt NonNamespacedDistributionTuple, versions Map[string /*epName*/, ProjectionModeVal]) {
				version, ok := pickVersionForEP(versions)
				if !ok {
					sbo.logger.Error(nil, "Impossible: addition of empty version map", "nndt", nndt)
				}
				ctSansEPName.Put(nndt, version)
			},
			OnDelete: func(nndt NonNamespacedDistributionTuple) {
				ctSansEPName.Delete(nndt)
			},
		}
		// clusterWhatWhereFull is a map from ClusterWhatWhereFullKey to API version (no group),
		// factored into a map from NonNamespacedDistributionTuple to epName to API version.
		sbo.clusterWhatWhereFull = NewFactoredMapMap[ClusterWhatWhereFullKey, NonNamespacedDistributionTuple, string /* ep name */, ProjectionModeVal](
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
// For the namespaced resources, as a SingleBinder this organizer
// is given the stream of changes to following three relations:
// - WhatWheres: set of ((epCluster,epName),(namespace,destination))
// - DiscoG: map of (epCluster,APIGroup) -> APIVersionInfo
// - DiscoR: map of (epCluster,GroupResource) -> APIVersion (the preferred one)
// and produces change streams for the following three relations:
// - nsToGo: set of NamespaceDistributionTuple (epCluster,namespace,destination)
// - set of NamespacedResourceDistributionTuple (epCluster,GroupResource,destination)
// - map of ProjectionModeKey (GroupResource,destination) -> ProjectionModeVal (APIVersion).
//
// As a relational expression, the desired computation is as follows.
// nsToGo = WhatWhere.ProjectOut(epName)
// nsCommon = DiscoR.Equijoin(nsToGo.ProjectOut(namespace))
// (that's a map (epCluster,GroupResource,destination) -> APIVesion)
// NamespacedResourceDistributionTuples = nsCommon.Keys()
// ProjectionsModes = nsCommon.GroupBy(GroupResource,destination).Aggregate(PickVersion)
//
// The query plan is as follows.
// nsModesReceiver <- nsCommon.GroupBy(GroupResource,destination).Aggregate(PickVersion)
// NamespacedResourceDistributionTuples <- nsCommon.Keys()
// nsCommon <- Equijoin(DiscoR, nsSrcAndDest)
// nsSrcAndDest <- nsToGo.ProjectOut(namespace)
// nsToGo <- WhatWheres.ProjectOut(epName)
//
// For the cluster-scoped resources, as a SingleBinder this organizer
// is given the stream of change to following relations:
// - WhatWheres: map of ((epCluster,epName),GroupResource,ObjName,destination) -> APIVersion
// and produces the change streams to the following two relations:
// - set of NonNamespacedDistributionTuple (epCluster,GroupResource,ObjName,destination)
// - map of ProjectionModeKey (GroupResource,destination) -> ProjectionModeVal (APIVersion).
//
// As a relational algebra expression, the desired computation is as follows.
// common = WhatWheres.ProjectOut(epName)
// (that's a map (epCluster,GroupResource,ObjName,destination) -> APIVersion)
// NonNamespacedDistributionTuples = common.Keys()
// ProjectionModes = common.GroupBy(GroupResource,destination).Aggregate(PickVersion)
//
// The query plan is as follows.
// clusterModesReceiver <- ctSansEPName.GroupBy(GroupResource,destination).Aggregate(PickVersion)
// clusterDistributionsReceiver <- ctSansEPName.Keys()
// ctSansEPName <- WhatWheres.ProjectOut(epName)
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

	perSourceCluster MutableMap[logicalcluster.Name, *simpleBindingPerCluster]

	workloadProjectionSections WorkloadProjectionSections // non-zero only during a transaction!

	// The following fields hold the same value throughout the lifetime of this object,
	// but those values use workloadProjectionSections --- synchronously --- and so can
	// only be invoked during a transaction.

	clusterWhatWhereFull      MappingReceiver[ClusterWhatWhereFullKey, ProjectionModeVal]
	namespacedWhatWhereFull   SetWriter[NamespacedWhatWhereFullKey]
	upsyncsFull               SetWriter[Triple[ExternalName /* of EdgePlacement object */, edgeapi.UpsyncSet, SinglePlacement]]
	resourceDiscoveryReceiver MappingReceiver[ResourceDiscoveryKey, ProjectionModeVal]
}

type NamespaceName string

type NamespaceAndDestination = Pair[NamespaceName, SinglePlacement]

type SourceAndDestination = Pair[logicalcluster.Name, SinglePlacement]

type NamespacedJoinKeyLessnS = Triple[logicalcluster.Name, metav1.GroupResource, SinglePlacement]

var factorNamespacedJoinKeyLessNS = Factorer[NamespacedJoinKeyLessnS, ProjectionModeKey, logicalcluster.Name]{
	First: func(whole NamespacedJoinKeyLessnS) Pair[ProjectionModeKey, logicalcluster.Name] {
		return Pair[ProjectionModeKey, logicalcluster.Name]{
			First:  ProjectionModeKey{whole.Second, whole.Third},
			Second: whole.First}
	},
	Second: func(parts Pair[ProjectionModeKey, logicalcluster.Name]) NamespacedJoinKeyLessnS {
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

var factorNamespacedWhatWhereFullKey = Factorer[NamespacedWhatWhereFullKey, NamespaceDistributionTuple, string /*epName*/]{
	First: func(nfk NamespacedWhatWhereFullKey) Pair[NamespaceDistributionTuple, string /*epName*/] {
		return Pair[NamespaceDistributionTuple, string /*epName*/]{
			First: Triple[logicalcluster.Name, NamespaceName, SinglePlacement]{
				nfk.First.Cluster,
				nfk.Second,
				nfk.Third},
			Second: nfk.First.Name}
	},
	Second: func(parts Pair[NamespaceDistributionTuple, string /*epName*/]) NamespacedWhatWhereFullKey {
		return NamespacedWhatWhereFullKey{
			First:  ExternalName{parts.First.First, parts.Second},
			Second: parts.First.Second,
			Third:  parts.First.Third}
	},
}

// factorClusterWhatWhereFullKey factors a ClusterWhatWhereFullKey into
// a (NonNamespacedDistributionTuple, string/*ep name*/) pair.
var factorClusterWhatWhereFullKey = Factorer[ClusterWhatWhereFullKey, NonNamespacedDistributionTuple, string /* ep name */]{
	First: func(cfk ClusterWhatWhereFullKey) Pair[NonNamespacedDistributionTuple, string /* ep name */] {
		return Pair[NonNamespacedDistributionTuple, string /* ep name */]{
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
	Second: func(decomp Pair[NonNamespacedDistributionTuple, string /*ep name*/]) ClusterWhatWhereFullKey {
		return ClusterWhatWhereFullKey{
			First:  ExternalName{decomp.First.Second.Cluster, decomp.Second},
			Second: Pair[metav1.GroupResource, string]{decomp.First.First.GroupResource, decomp.First.Second.Name},
			Third:  decomp.First.First.Destination,
		}
	},
}

type ResourceDiscoveryKey = Pair[logicalcluster.Name /*wmw*/, metav1.GroupResource]

type NamespacedWhatWhereFullKey = Triple[ExternalName, NamespaceName, SinglePlacement]

// ClusterWhatWhereFullKey is (EdgePlacement id, (resource, object name), destination)
type ClusterWhatWhereFullKey = Triple[ExternalName, Pair[metav1.GroupResource, string], SinglePlacement]

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
	rscMode := sbo.resourceModes(metav1.GroupResource{Group: tup.Second.APIGroup, Resource: tup.Second.Resource})
	if !rscMode.GoesToMailbox() {
		sbo.logger.V(4).Info("Ignoring WhatWhere tuple because it does not go to the mailbox workspaces", "tup", tup, "rscMode", rscMode)
		return
	}
	sbo.getSourceCluster(tup.First.Cluster, true)
	gr := tup.Second.GroupResource()
	if mgrIsNamespace(gr) {
		wwTup := NamespacedWhatWhereFullKey{tup.First, NamespaceName(tup.Second.Name), tup.Third}
		sbo.namespacedWhatWhereFull.Add(wwTup)
		if !val.IncludeNamespaceObject {
			return
		}
	}
	key := ClusterWhatWhereFullKey{tup.First, Pair[metav1.GroupResource, string]{gr, tup.Second.Name}, tup.Third}
	sbo.clusterWhatWhereFull.Put(key, ProjectionModeVal{APIVersion: val.APIVersion})
}

func (sxo sboXnOps) Delete(tup Triple[ExternalName, WorkloadPartID, SinglePlacement]) {
	sbo := sxo.sbo
	rscMode := sbo.resourceModes(metav1.GroupResource{Group: tup.Second.APIGroup, Resource: tup.Second.Resource})
	if !rscMode.GoesToMailbox() {
		sbo.logger.V(4).Info("Ignoring WhatWhere tuple because it does not go to the mailbox workspaces", "tup", tup, "rscMode", rscMode)
		return
	}
	sbc := sbo.getSourceCluster(tup.First.Cluster, false)
	if sbc == nil {
		return
	}
	gr := tup.Second.GroupResource()
	if mgrIsNamespace(gr) /* && !val.IncludeNamespaceObject */ {
		wwTup := NamespacedWhatWhereFullKey{tup.First, NamespaceName(tup.Second.Name), tup.Third}
		sbo.namespacedWhatWhereFull.Remove(wwTup)
	}
	key := ClusterWhatWhereFullKey{tup.First, Pair[metav1.GroupResource, string]{gr, tup.Second.Name}, tup.Third}
	sbo.clusterWhatWhereFull.Delete(key)
	if false {
		// TODO: make this happen iff there is no remaining data for the cluster
		sbo.logger.V(4).Info("Removing discovery receivers", "cluster", sbc.cluster)
		sbo.discovery.RemoveReceivers(sbc.cluster, &sbc.groupReceiver, &sbc.resourceReceiver)
	}
}

func (sbo *simpleBindingOrganizer) getSourceCluster(cluster logicalcluster.Name, want bool) *simpleBindingPerCluster {
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
	cluster          logicalcluster.Name
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

func (partID WorkloadPartID) GroupResource() metav1.GroupResource {
	return metav1.GroupResource{
		Group:    partID.APIGroup,
		Resource: partID.Resource,
	}
}
