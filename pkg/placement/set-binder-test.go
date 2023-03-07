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
	k8ssets "k8s.io/apimachinery/pkg/util/sets"

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
	gvr1 := metav1.GroupVersionResource{
		Group:    gr1.Group,
		Version:  "v1",
		Resource: gr1.Resource}
	workloadPart1 := WorkloadPart{
		WorkloadPartID{
			APIGroup: gvr1.Group,
			Resource: gvr1.Resource,
			Name:     "crd1",
		},
		WorkloadPartDetails{APIVersion: "v1"},
	}
	what1 := WorkloadParts{workloadPart1.WorkloadPartID: workloadPart1.WorkloadPartDetails}
	whatProvider := NewRelayMap[ExternalName, WorkloadParts](false)
	sc1 := logicalcluster.Name("wm1")
	ep1Ref := ExternalName{Cluster: sc1, Name: "ep1"}
	whatProvider.Set(ep1Ref, what1)
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
	whereProvider := NewRelayMap[ExternalName, ResolvedWhere](false)
	whereProvider.Set(ep1Ref, where1)
	whatProvider.AddConsumer(binder.AsWhatConsumer(), true)
	whereProvider.AddConsumer(binder.AsWhereConsumer(), true)
	projectionTracker := NewRelayMap[ProjectionKey, *ProjectionPerCluster](false)
	binder.AddConsumer(projectionTracker, true)
	if projectionTracker.Len() != 1 {
		t.Errorf("Wrong amount of stuff in projectionTracker.theMap: %#+v", projectionTracker)
	}
	pk1 := ProjectionKey{gr1, sp1}
	ppc1 := DynamicMapProducerGet[ProjectionKey, *ProjectionPerCluster](projectionTracker, pk1)
	if ppc1 == nil {
		t.Errorf("Missing ProjectionPerCluster")
		return // to stop linter from complaining about possible nil pointer below
	}
	if ppc1.APIVersion != gvr1.Version {
		t.Errorf("Wrong API version: got %q, expected %q", ppc1.APIVersion, gvr1.Version)
	}
	pcTracker := NewRelayMap[logicalcluster.Name, ProjectionDetails](false)
	ppc1.PerSourceCluster.AddConsumer(pcTracker, true)
	if pcTracker.Len() != 1 {
		t.Errorf("Wrong amount of stuff in pcTracker.theMap: %#+v", pcTracker)
	}
	pd1 := DynamicMapProducerGet[logicalcluster.Name, ProjectionDetails](pcTracker, sc1)
	if pd1.Namespaces != nil {
		t.Errorf("Expected no namespaces but got %#+v", *pd1.Namespaces)
	}
	if pd1.Names == nil {
		t.Error("Expected one name but got none")
	} else {
		expectedNames := k8ssets.NewString(workloadPart1.Name)
		if !expectedNames.Equal(*pd1.Names) {
			t.Errorf("Got wrong set of names: expected %v, got %v", expectedNames, *pd1.Names)
		}
	}
}
