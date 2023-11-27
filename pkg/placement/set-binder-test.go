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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
)

// exerciseSetBinder tests a given SetBinder.
// For a particular SetBinder implementation Foo,
// invoke this from TestFoo (or whatever).
// At the current state of development, the test is utterly simple.
// TODO: make this test harder.
func exerciseSetBinder(t *testing.T, logger klog.Logger, resourceDiscoveryReceiver MappingReceiver[Pair[logicalcluster.Name, metav1.GroupResource], ResourceDetails], binder SetBinder) {
	gr1 := metav1.GroupResource{
		Group:    "apiextensions.k8s.io",
		Resource: "customresourcedefinitions"}
	workloadPartDetails1 := WorkloadPartDetails{APIVersion: "v1", ReturnSingletonState: true}
	workloadPartID1 := WorkloadPartID{
		First:  gr1,
		Second: metav1.NamespaceNone,
		Third:  ObjectName("crd1"),
	}
	parts1 := WorkloadParts{workloadPartID1: workloadPartDetails1}
	ups1 := []edgeapi.UpsyncSet{
		{APIGroup: "group1.test", Resources: []string{"sprockets", "flanges"}, Names: []string{"George", "Cosmo"}},
		{APIGroup: "group2.test", Resources: []string{"cogs"}, Names: []string{"William"}}}
	gr2 := metav1.GroupResource{
		Group:    "",
		Resource: "namespaces"}
	workloadPartDetails2 := WorkloadPartDetails{APIVersion: "v1", CreateOnly: true}
	workloadPartID2 := WorkloadPartID{
		First:  gr2,
		Second: metav1.NamespaceNone,
		Third:  ObjectName("ns-a"),
	}
	parts2 := WorkloadParts{workloadPartID2: workloadPartDetails2}
	ups2 := []edgeapi.UpsyncSet{
		ups1[0],
		{APIGroup: "group3.test", Resources: []string{"widgets"}, Names: []string{"*"}}}
	sc1 := logicalcluster.Name("wm1")
	ep1Ref := ExternalName{Cluster: sc1, Name: "ep1"}
	spA := SinglePlacement{
		Cluster:        "inv1",
		LocationName:   "loc1",
		SyncTargetName: "st1",
		SyncTargetUID:  apimachtypes.UID("uid1"),
	}
	spsA := &edgeapi.SinglePlacementSlice{
		Destinations: []SinglePlacement{spA},
	}
	where1 := ResolvedWhere{spsA}
	NamespacedObjectDistributions := NewMapMap[NamespacedDistributionTuple, DistributionBits](nil)
	NamespacedModes := NewMapMap[ProjectionModeKey, ProjectionModeVal](nil)
	NonNamespacedObjectDistributions := NewMapMap[NonNamespacedDistributionTuple, DistributionBits](nil)
	NonNamespacedModes := NewMapMap[ProjectionModeKey, ProjectionModeVal](nil)
	Upsyncs := NewHashSet(PairHashDomain[SinglePlacement, edgeapi.UpsyncSet](HashSinglePlacement{}, HashUpsyncSet{}))
	projectionTracker := WorkloadProjectionSections{
		NamespacedObjectDistributions:    NamespacedObjectDistributions,
		NamespacedModes:                  NamespacedModes,
		NonNamespacedObjectDistributions: NonNamespacedObjectDistributions,
		NonNamespacedModes:               NonNamespacedModes,
		Upsyncs:                          Upsyncs,
	}
	whatReceiver, whereReceiver := binder(TrivialTransactor[WorkloadProjectionSections]{projectionTracker})
	rw1 := ResolvedWhat{parts1, ups1}
	t.Logf("Setting epRef=%v, ResolvedWhat=%v", ep1Ref, rw1)
	logger.Info("Setting ResolvedWhat", "epRef", ep1Ref, "resolvedWhat", rw1)
	whatReceiver.Put(ep1Ref, rw1)
	whereReceiver.Put(ep1Ref, where1)
	pmk1 := ProjectionModeKey{
		GroupResource: gr1,
		Destination:   spA,
	}
	pmv1 := ProjectionModeVal{workloadPartDetails1.APIVersion}
	pmv2 := ProjectionModeVal{workloadPartDetails2.APIVersion}
	objRef1 := ExternalName{Cluster: sc1, Name: workloadPartID1.Third}
	objRef2 := ExternalName{Cluster: sc1, Name: workloadPartID2.Third}
	expectedNonNamespacedObjectDistributions := NewMapMap[NonNamespacedDistributionTuple, DistributionBits](nil)
	expectedNonNamespacedObjectDistributions.Put(NewPair(pmk1, objRef1), DistributionBits{ReturnSingletonState: true})
	if !MapEqual[NonNamespacedDistributionTuple, DistributionBits](expectedNonNamespacedObjectDistributions, NonNamespacedObjectDistributions) {
		t.Fatalf("Wrong NonNamespacedDistributions; expected=%v, got=%v", expectedNonNamespacedObjectDistributions, NonNamespacedObjectDistributions)
	}
	expectedNonNamespacedModes := NewMapMap[ProjectionModeKey, ProjectionModeVal](nil)
	expectedNonNamespacedModes.Put(pmk1, pmv1)
	MapEnumerateDifferences[ProjectionModeKey, ProjectionModeVal](expectedNonNamespacedModes, NonNamespacedModes,
		MapChangeReceiverFuncs[ProjectionModeKey, ProjectionModeVal]{
			OnCreate: func(key ProjectionModeKey, val ProjectionModeVal) {
				t.Fatalf("Extra entry in NonNamespacedModes; key=%v, val=%v", key, val)
			},
			OnUpdate: func(key ProjectionModeKey, goodVal, badVal ProjectionModeVal) {
				t.Fatalf("Wrong entry in NonNamespacedModes; key=%v, expected=%v, got=%v", key, goodVal, badVal)
			},
			OnDelete: func(key ProjectionModeKey, val ProjectionModeVal) {
				t.Fatalf("Missing entry in NonNamespacedModes; key=%v, val=%v", key, val)
			},
		})

	expectedUpsyncs := NewHashSet(
		PairHashDomain[SinglePlacement, edgeapi.UpsyncSet](HashSinglePlacement{}, HashUpsyncSet{}),
		NewPair(spA, ups1[0]),
		NewPair(spA, ups1[1]))
	if !SetEqual[Pair[SinglePlacement, edgeapi.UpsyncSet]](expectedUpsyncs, Upsyncs) {
		t.Fatalf("Wrong Upsyncs: expected %v, got %v",
			VisitableToSlice[Pair[SinglePlacement, edgeapi.UpsyncSet]](expectedUpsyncs),
			VisitableToSlice[Pair[SinglePlacement, edgeapi.UpsyncSet]](Upsyncs))
	}
	expectedNamespacedDistributions := NewMapMap[NamespacedDistributionTuple, DistributionBits](nil)
	expectedNamespacedModes := NewMapMap[ProjectionModeKey, ProjectionModeVal](nil)

	rd2 := ResourceDetails{Namespaced: true, SupportsInformers: true, PreferredVersion: workloadPartDetails2.APIVersion}
	rw2 := ResolvedWhat{parts2, ups2}
	t.Logf("Setting epRef=%v, ResolvedWhat=%v", ep1Ref, rw2)
	logger.Info("Setting ResolvedWhat", "epRef", ep1Ref, "resolvedWhat", rw2)
	whatReceiver.Put(ep1Ref, rw2)
	t.Logf("Adding resource discovery key=%v, val=%v", NewPair(sc1, gr2), rd2)
	logger.Info("Adding resource discovery mapping", "key", NewPair(sc1, gr2), "val", rd2)
	resourceDiscoveryReceiver.Put(NewPair(sc1, gr2), rd2)
	pmk2 := ProjectionModeKey{
		GroupResource: gr2,
		Destination:   spA,
	}
	expectedNonNamespacedObjectDistributions.Delete(NewPair(pmk1, objRef1))
	expectedNonNamespacedModes.Delete(pmk1)
	expectedNonNamespacedObjectDistributions.Put(NewPair(pmk2, objRef2), DistributionBits{CreateOnly: true})
	expectedNonNamespacedModes.Put(pmk2, pmv2)

	expectedUpsyncs.Add(NewPair(spA, ups2[1]))
	expectedUpsyncs.Remove(NewPair(spA, ups1[1]))
	if !MapEqual[NonNamespacedDistributionTuple, DistributionBits](expectedNonNamespacedObjectDistributions, NonNamespacedObjectDistributions) {
		t.Errorf("Wrong NonNamespacedDistributions; expected=%v, got=%v", expectedNonNamespacedObjectDistributions, NonNamespacedObjectDistributions)
	}
	if !MapEqual[NamespacedDistributionTuple, DistributionBits](expectedNamespacedDistributions, NamespacedObjectDistributions) {
		t.Errorf("Wrong NamespacedDistributions; expected=%v, got=%v", expectedNamespacedDistributions, NamespacedObjectDistributions)
	}
	MapEnumerateDifferences[ProjectionModeKey, ProjectionModeVal](expectedNamespacedModes, NamespacedModes,
		MapChangeReceiverFuncs[ProjectionModeKey, ProjectionModeVal]{
			OnCreate: func(key ProjectionModeKey, val ProjectionModeVal) {
				t.Errorf("Extra entry in NamespacedModes; key=%v, val=%v", key, val)
			},
			OnUpdate: func(key ProjectionModeKey, goodVal, badVal ProjectionModeVal) {
				t.Errorf("Wrong entry in NamespacedModes; key=%v, expected=%v, got=%v", key, goodVal, badVal)
			},
			OnDelete: func(key ProjectionModeKey, val ProjectionModeVal) {
				t.Errorf("Missing entry in NamespacedModes; key=%v, val=%v", key, val)
			},
		})
	expectedUpsyncs.Add(NewPair(spA, ups2[1]))
	if !SetEqual[Pair[SinglePlacement, edgeapi.UpsyncSet]](expectedUpsyncs, Upsyncs) {
		t.Errorf("Wrong Upsyncs: expected %v, got %v",
			VisitableToSlice[Pair[SinglePlacement, edgeapi.UpsyncSet]](expectedUpsyncs),
			VisitableToSlice[Pair[SinglePlacement, edgeapi.UpsyncSet]](Upsyncs))
	}
}
