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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
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
		pickVersionForEP := func(versions MutableMap[string /*epName*/, ProjectionModeVal]) (ProjectionModeVal, bool) {
			var version ProjectionModeVal
			if versions.Visit(func(pair Pair[string /*epName*/, ProjectionModeVal]) error {
				version = pair.Second
				return errStop
			}) == nil {
				return version, false
			}
			return version, true
		}
		namespacedModesReceiver := MappingReceiverFuncs[ProjectionModeKey, ProjectionModeVal]{
			OnPut: func(mk ProjectionModeKey, val ProjectionModeVal) {
				sbo.workloadProjectionSections.NamespacedModes.Put(mk, val)
			},
			OnDelete: func(mk ProjectionModeKey) {
				sbo.workloadProjectionSections.NamespacedModes.Delete(mk)
			},
		}
		ntSansNS := NewFactoredMapMapAggregator[NamespacedJoinKeyLessnS, ProjectionModeKey, logicalcluster.Name, ProjectionModeVal, ProjectionModeVal](
			factorNamespacedJoinKeyLessNS,
			nil,
			nil,
			pickThe1[ProjectionModeKey, logicalcluster.Name](sbo, "Not implemented yet: handling version conflicts for namespaced resources"),
			namespacedModesReceiver,
		)
		nsProjector := NewFactoredMapMapAggregator[NamespacedJoinKey, NamespacedJoinKeyLessnS, NamespaceName, ProjectionModeVal, ProjectionModeVal](
			factorNamespacedJoinKey,
			nil,
			nil,
			pickThe1[NamespacedJoinKeyLessnS, NamespaceName](sbo, "Should not happen: version difference between namespaces"),
			ntSansNS,
		)
		namespacedDistributionRelay := MappingReceiverFuncs[NamespacedJoinKey, ProjectionModeVal]{
			OnPut: func(key NamespacedJoinKey, val ProjectionModeVal) {
				dist := NamespacedJoinKeyToDistribution(key)
				sbo.workloadProjectionSections.NamespacedDistributions.Add(dist)
			},
			OnDelete: func(key NamespacedJoinKey) {
				dist := NamespacedJoinKeyToDistribution(key)
				sbo.workloadProjectionSections.NamespacedDistributions.Remove(dist)
			},
		}
		nsCommon := MappingReceiverFork[NamespacedJoinKey, ProjectionModeVal]{namespacedDistributionRelay, nsProjector}

		rscDisco, ntSansEPName := NewDynamicFullJoin12VWith13[logicalcluster.Name, metav1.GroupResource, NamespaceAndDestination, ProjectionModeVal](
			logger, nsCommon)
		sbo.resourceDiscoveryReceiver = rscDisco

		sbo.namespacedWhatWhereFull = NewSetChangeProjector[NamespacedWhatWhereFullKey, Pair[logicalcluster.Name, NamespaceAndDestination], string /*epName*/](
			factorNamespacedWhatWhereFullKey, ntSansEPName)

		clusterDistributionsReceiver := SetChangeReceiverFuncs[NonNamespacedDistributionTuple]{
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
			pickThe1[ProjectionModeKey, ExternalName /*of downsynced object*/](sbo, "Not implemented yet: handling version conflicts for cluster-scoped resources"),
			clusterModesReceiver,
		)
		// ctSansEPName receives the change stream of clusterWhatWhereFull with epName projected out,
		// and passes it along to clusterDistributionsReceiver and aggregateForCluster.
		var ctSansEPName MapChangeReceiver[NonNamespacedDistributionTuple, ProjectionModeVal] = MapChangeReceiverFork[NonNamespacedDistributionTuple, ProjectionModeVal]{
			MapKeySetReceiver[NonNamespacedDistributionTuple, ProjectionModeVal](clusterDistributionsReceiver),
			MappingReceiverDiscardsPrevious[NonNamespacedDistributionTuple, ProjectionModeVal](aggregateForCluster),
		}
		// clusterChangeReceiver receives the change stream of the full map and projects out the EdgePlacement
		// object name to feed to sansEPName.
		// This is relatively simple because the API version does not vary for a given resource and source cluster.
		clusterChangeReceiver := MapChangeReceiverFuncs[NonNamespacedDistributionTuple, MutableMap[string /*epName*/, ProjectionModeVal]]{
			OnCreate: func(nndt NonNamespacedDistributionTuple, versions MutableMap[string /*epName*/, ProjectionModeVal]) {
				version, ok := pickVersionForEP(versions)
				if !ok {
					sbo.logger.Error(nil, "Impossible: addition of empty version map", "nndt", nndt)
				}
				ctSansEPName.Create(nndt, version)
			},
			OnDelete: func(nndt NonNamespacedDistributionTuple, versions MutableMap[string /*epName*/, ProjectionModeVal]) {
				version, ok := pickVersionForEP(versions)
				if !ok {
					sbo.logger.Error(nil, "Impossible: removal of empty version map", "nndt", nndt)
				}
				ctSansEPName.DeleteWithFinal(nndt, version)
			},
		}
		// clusterWhatWhereFull is a map from ClusterWhatWhereFullKey to API version (no group),
		// factored into a map from NonNamespacedDistributionTuple to epName to API version.
		sbo.clusterWhatWhereFull = NewFactoredMapMap[ClusterWhatWhereFullKey, NonNamespacedDistributionTuple, string /* ep name */, ProjectionModeVal](
			factorClusterWhatWhereFullKey,
			nil,
			clusterChangeReceiver,
		)
		return sbo
	}
}

// type SingleBindingOps MappingReceiver[Triple[ExternalName /* of EdgePlacement object */, WorkloadPartID, edgeapi.SinglePlacement], WorkloadPartDetails]

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
// and produces change streams for the following two relations:
// - set of NamespacedDistributionTuple (epCluster,GroupResource,(namespace,destionation))
// - map of ProjectionModeKey (GroupResource,destionation) -> ProjectionModeVal (APIVersion).
//
// As a relational expression, the desired computation is as follows.
// common = DiscoR.Equijoin(WhatWhere.ProjectOut(epName))
// (that's a map (epCluster,GroupResource,(namespace,destination)) -> APIVesion)
// NamespacedDistributionTuples = common.Keys()
// ProjectionsModes = common.ProjectOut(namespace)
// .GroupBy(GroupResource,destination)
// .Aggregate(PickVersion)
//
// The query plan is as follows.
// namespacedModesReceiver <- ntSansNS.GroupBy(GroupResource,destination).Aggregate(PickVersion)
// ntSansNS <- nsCommon.ProjectOut(namespace)
// NamespacedDistributionTuples <- nsCommon.Keys()
// nsCommon <- Equijoin(DiscoR, ntSansEPName)
// ntSansEPName <- WhatWheres.ProjectOut(epName)
//
// For the cluster-scoped resources, as a SingleBinder this organizer
// is given the stream of change to following relations:
// - WhatWheres: map of ((epCluster,epName),GroupResource,ObjName,destination) -> APIVersion
// and produces the change streams to the following two relations:
// - set of NonNamespacedDistributionTuple (epCluster,GroupResource,ObjName,destination)
// - map of ProjectionModeKey (GroupResource,destionation) -> ProjectionModeVal (APIVersion).
//
// As a relational algebra expression, the desired computation is as follows.
// common = WhatWheres.ProjectOut(epName)
// (that's a map (epCluster,GroupResource,ObjName,destination) -> APIVersion)
// NonNamespacedDistributionTuples = common.Keys()
// ProjectionModes = common.GroupBy(GroupResource,destionation).Aggregate(PickVersion)
//
// The query plan is as follows.
// clusterModesReceiver <- ctSansEPName.GroupBy(GroupResource,destionation).Aggregate(PickVersion)
// clusterDistributionsReceiver <- ctSansEPName.Keys()
// ctSansEPName <- WhatWheres.ProjectOut(epName)
//
// Not implemented yet: namespaced objects.
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
	namespacedWhatWhereFull   SetChangeReceiver[NamespacedWhatWhereFullKey]
	resourceDiscoveryReceiver MappingReceiver[ResourceDiscoveryKey, ProjectionModeVal]
}

type NamespaceName string

type NamespaceAndDestination struct {
	NamespaceName
	Destination edgeapi.SinglePlacement
}

type NamespacedJoinKey = Triple[logicalcluster.Name, metav1.GroupResource, NamespaceAndDestination]

func NamespacedJoinKeyToDistribution(njk NamespacedJoinKey) NamespacedDistributionTuple {
	return NamespacedDistributionTuple{
		SourceCluster: njk.First,
		ProjectionModeKey: ProjectionModeKey{
			GroupResource: njk.Second,
			Destination:   njk.Third.Destination},
		NamespaceName: njk.Third.NamespaceName}
}

type NamespacedJoinKeyLessnS = Triple[logicalcluster.Name, metav1.GroupResource, edgeapi.SinglePlacement]

var factorNamespacedJoinKey = Factorer[NamespacedJoinKey, NamespacedJoinKeyLessnS, NamespaceName]{
	First: func(whole NamespacedJoinKey) Pair[NamespacedJoinKeyLessnS, NamespaceName] {
		return Pair[NamespacedJoinKeyLessnS, NamespaceName]{
			First: NamespacedJoinKeyLessnS{
				First:  whole.First,
				Second: whole.Second,
				Third:  whole.Third.Destination},
			Second: whole.Third.NamespaceName}
	},
	Second: func(parts Pair[NamespacedJoinKeyLessnS, NamespaceName]) NamespacedJoinKey {
		return NamespacedJoinKey{
			First:  parts.First.First,
			Second: parts.First.Second,
			Third:  NamespaceAndDestination{parts.Second, parts.First.Third}}
	},
}

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

func pickThe1[KeyPartA, KeyPartB comparable](sbo *simpleBindingOrganizer, errmsg string) func(keyPartA KeyPartA, problem Map[KeyPartB, ProjectionModeVal]) ProjectionModeVal {
	return func(keyPartA KeyPartA, problem Map[KeyPartB, ProjectionModeVal]) ProjectionModeVal {
		versions := NewMapSet[ProjectionModeVal]()
		var solution ProjectionModeVal
		problem.Visit(func(pair Pair[KeyPartB, ProjectionModeVal]) error {
			versions.Add(pair.Second)
			solution = pair.Second
			return nil
		})
		if versions.Len() != 1 {
			sbo.logger.Error(nil, errmsg, "keyPartA", keyPartA, "problem", problem, "chosen", solution)
		}
		return solution
	}
}

var factorNamespacedWhatWhereFullKey = Factorer[NamespacedWhatWhereFullKey, Pair[logicalcluster.Name, NamespaceAndDestination], string /*epName*/]{
	First: func(nfk NamespacedWhatWhereFullKey) Pair[Pair[logicalcluster.Name, NamespaceAndDestination], string /*epName*/] {
		return Pair[Pair[logicalcluster.Name, NamespaceAndDestination], string /*epName*/]{
			First: Pair[logicalcluster.Name, NamespaceAndDestination]{
				First: nfk.First.Cluster,
				Second: NamespaceAndDestination{
					NamespaceName: nfk.Second,
					Destination:   nfk.Third}},
			Second: nfk.First.Name}
	},
	Second: func(parts Pair[Pair[logicalcluster.Name, NamespaceAndDestination], string /*epName*/]) NamespacedWhatWhereFullKey {
		return NamespacedWhatWhereFullKey{
			First:  ExternalName{parts.First.First, parts.Second},
			Second: parts.First.Second.NamespaceName,
			Third:  parts.First.Second.Destination}
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

type NamespacedWhatWhereFullKey Triple[ExternalName, NamespaceName, edgeapi.SinglePlacement]

// ClusterWhatWhereFullKey is (EdgePlacement id, (resource, object name), destination)
type ClusterWhatWhereFullKey Triple[ExternalName, Pair[metav1.GroupResource, string], edgeapi.SinglePlacement]

func (sbo *simpleBindingOrganizer) Transact(xn func(SingleBindingOps)) {
	sbo.Lock()
	defer sbo.Unlock()
	sbo.logger.V(3).Info("Begin transaction")
	sbo.workloadProjector.Transact(func(wps WorkloadProjectionSections) {
		sbo.workloadProjectionSections = wps
		xn(sboXnOps{sbo})
		sbo.workloadProjectionSections = WorkloadProjectionSections{}
	})
	sbo.logger.V(3).Info("End transaction")
}

// sboXnOps exposes the SingleBindingOps behavior, in the locked context of a transaction
type sboXnOps struct {
	sbo *simpleBindingOrganizer
}

func (sxo sboXnOps) Put(tup Triple[ExternalName, WorkloadPartID, edgeapi.SinglePlacement], val WorkloadPartDetails) {
	sbo := sxo.sbo
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

func (sxo sboXnOps) Delete(tup Triple[ExternalName, WorkloadPartID, edgeapi.SinglePlacement]) {
	sbo := sxo.sbo
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
	sbo.discovery.RemoveReceivers(sbc.cluster, sbcGroupReceiver{sbc}, sbcResourceReceiver{sbc})
}

func (sbo *simpleBindingOrganizer) getSourceCluster(cluster logicalcluster.Name, want bool) *simpleBindingPerCluster {
	sbc, have := sbo.perSourceCluster.Get(cluster)
	if want && !have {
		sbc = &simpleBindingPerCluster{
			simpleBindingOrganizer: sbo,
			cluster:                cluster,
		}
		sbo.perSourceCluster.Put(cluster, sbc)
		// sbo.discovery.AddClient(cluster, sbc)
		sbo.discovery.AddReceivers(cluster, sbcGroupReceiver{sbc}, sbcResourceReceiver{sbc})
	}
	return sbc
}

type simpleBindingPerCluster struct {
	*simpleBindingOrganizer
	cluster logicalcluster.Name
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
	key := Pair[logicalcluster.Name, metav1.GroupResource]{sbc.cluster, gr}
	sbc.Lock()
	defer sbc.Unlock()
	sbc.workloadProjector.Transact(func(ops WorkloadProjectionSections) {
		sbc.workloadProjectionSections = ops
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
	key := Pair[logicalcluster.Name, metav1.GroupResource]{sbc.cluster, gr}
	sbc.Lock()
	defer sbc.Unlock()
	sbc.workloadProjector.Transact(func(ops WorkloadProjectionSections) {
		sbc.workloadProjectionSections = ops
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
