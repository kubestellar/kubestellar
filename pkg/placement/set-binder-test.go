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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachtypes "k8s.io/apimachinery/pkg/types"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

// exerciseSetBinder tests a given SetBinder.
// For a particular SetBinder implementation Foo,
// invoke this from TestFoo (or whatever).
// At the current state of development, the test is utterly simple.
// TODO: make this test harder.
func exerciseSetBinder(t *testing.T, binder SetBinder) {
	gr1 := metav1.GroupResource{
		Group:    "group1.test",
		Resource: "customresourcedefinitions"}
	workloadPartDetails1 := WorkloadPartDetails{APIVersion: "v1"}
	gvr1 := metav1.GroupVersionResource{
		Group:    gr1.Group,
		Version:  workloadPartDetails1.APIVersion,
		Resource: gr1.Resource}
	workloadPartID1 := WorkloadPartID{
		APIGroup: gvr1.Group,
		Resource: gvr1.Resource,
		Name:     "crd1",
	}
	what1 := WorkloadParts{workloadPartID1: workloadPartDetails1}
	sc1 := logicalcluster.Name("wm1")
	ep1Ref := ExternalName{Cluster: sc1, Name: "ep1"}
	sp1 := edgeapi.SinglePlacement{
		Cluster:        "inv1",
		LocationName:   "loc1",
		SyncTargetName: "st1",
		SyncTargetUID:  apimachtypes.UID("uid1"),
	}
	sps1 := &edgeapi.SinglePlacementSlice{
		Destinations: []edgeapi.SinglePlacement{sp1},
	}
	where1 := ResolvedWhere{sps1}
	NamespacedDistributions := NewMapSet[NamespacedDistributionTuple]()
	NamespacedModes := NewMapMap[ProjectionModeKey, ProjectionModeVal](nil)
	NonNamespacedDistributions := NewMapSet[NonNamespacedDistributionTuple]()
	NonNamespacedModes := NewMapMap[ProjectionModeKey, ProjectionModeVal](nil)
	projectionTracker := WorkloadProjectionSections{
		NamespacedDistributions:    NamespacedDistributions,
		NamespacedModes:            NamespacedModes,
		NonNamespacedDistributions: NonNamespacedDistributions,
		NonNamespacedModes:         NonNamespacedModes,
	}
	whatReceiver, whereReceiver := binder(TrivialTransactor[WorkloadProjectionSections]{projectionTracker})
	whatReceiver.Put(ep1Ref, what1)
	whereReceiver.Put(ep1Ref, where1)
	pmk1 := ProjectionModeKey{
		GroupResource: gr1,
		Destination:   sp1,
	}
	pmv1 := ProjectionModeVal{workloadPartDetails1.APIVersion}
	objRef1 := ExternalName{Cluster: sc1, Name: workloadPartID1.Name}
	expectedNonNamespacedDistributions := NewMapSet[NonNamespacedDistributionTuple](
		NonNamespacedDistributionTuple{pmk1, objRef1},
	)
	if !SetEqual[NonNamespacedDistributionTuple](expectedNonNamespacedDistributions, NonNamespacedDistributions) {
		t.Fatalf("Wrong NonNamespacedDistributions; expected=%v, got=%v", expectedNonNamespacedDistributions, NonNamespacedDistributions)
	}
	expectedNonNamespacedModes := NewMapMap[ProjectionModeKey, ProjectionModeVal](nil)
	expectedNonNamespacedModes.Put(pmk1, pmv1)
	MapEnumerateDifferences[ProjectionModeKey, ProjectionModeVal](expectedNonNamespacedModes, NonNamespacedModes,
		MapChangeReceiverFuncs[ProjectionModeKey, ProjectionModeVal]{
			OnCreate: func(key ProjectionModeKey, val ProjectionModeVal) {
				t.Fatalf("Missing entry in NonNamespacedModes; key=%v, val=%v", key, val)
			},
			OnUpdate: func(key ProjectionModeKey, goodVal, badVal ProjectionModeVal) {
				t.Fatalf("Wrong entry in NonNamespacedModes; key=%v, expected=%v, got=%v", key, goodVal, badVal)
			},
			OnDelete: func(key ProjectionModeKey, val ProjectionModeVal) {
				t.Fatalf("Extra entry in NonNamespacedModes; key=%v, val=%v", key, val)
			},
		})
}
