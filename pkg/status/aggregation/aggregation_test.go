/*
Copyright 2025 The KubeStellar Authors.

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

package aggregation

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// toStatusMap converts a typed status into the unstructured form that the
// status controller reads from a WorkStatus. Fields tagged omitempty that hold
// their zero value are dropped, as they are on the wire.
func toStatusMap(t *testing.T, status any) map[string]any {
	t.Helper()
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(status)
	if err != nil {
		t.Fatalf("convert status: %v", err)
	}
	return obj
}

func expectInt(t *testing.T, got map[string]any, field string, want int64) {
	t.Helper()
	if v, ok := got[field].(int64); !ok || v != want {
		t.Errorf("%s: got %#v, want %d", field, got[field], want)
	}
}

func TestGetMinMaxTreatMissingAsZero(t *testing.T) {
	statuses := []map[string]any{
		{"readyReplicas": int64(3)},
		{},
	}
	if got := GetMin(statuses, "readyReplicas"); got != 0 {
		t.Errorf("GetMin: got %d, want 0", got)
	}
	if got := GetMax(statuses, "readyReplicas"); got != 3 {
		t.Errorf("GetMax: got %d, want 3", got)
	}
	if got := GetMin(nil, "readyReplicas"); got != 0 {
		t.Errorf("GetMin with no statuses: got %d, want 0", got)
	}
	if got := GetMax(nil, "readyReplicas"); got != 0 {
		t.Errorf("GetMax with no statuses: got %d, want 0", got)
	}
}

func TestAggregateDeploymentStatusWithZeroReadyWEC(t *testing.T) {
	healthy := toStatusMap(t, &appsv1.DeploymentStatus{
		ObservedGeneration: 2, Replicas: 3, UpdatedReplicas: 3, ReadyReplicas: 3, AvailableReplicas: 3,
	})
	notReady := toStatusMap(t, &appsv1.DeploymentStatus{
		ObservedGeneration: 2, Replicas: 3, UpdatedReplicas: 3, UnavailableReplicas: 3,
	})
	if _, present := notReady["readyReplicas"]; present {
		t.Fatalf("test setup: expected readyReplicas to be omitted when 0")
	}

	got, err := AggregateDeploymentStatus([]map[string]any{healthy, notReady})
	if err != nil {
		t.Fatal(err)
	}
	expectInt(t, got, "replicas", 3)
	expectInt(t, got, "updatedReplicas", 3)
	expectInt(t, got, "readyReplicas", 0)
	expectInt(t, got, "availableReplicas", 0)
	expectInt(t, got, "unavailableReplicas", 3)
	expectInt(t, got, "observedGeneration", 2)
}

func TestAggregateReplicaSetStatusWithZeroReadyWEC(t *testing.T) {
	healthy := toStatusMap(t, &appsv1.ReplicaSetStatus{
		ObservedGeneration: 1, Replicas: 2, FullyLabeledReplicas: 2, ReadyReplicas: 2, AvailableReplicas: 2,
	})
	notReady := toStatusMap(t, &appsv1.ReplicaSetStatus{
		ObservedGeneration: 1, Replicas: 2, FullyLabeledReplicas: 2,
	})

	got, err := AggregateReplicaSetStatus([]map[string]any{healthy, notReady})
	if err != nil {
		t.Fatal(err)
	}
	expectInt(t, got, "replicas", 2)
	expectInt(t, got, "fullyLabeledReplicas", 2)
	expectInt(t, got, "readyReplicas", 0)
	expectInt(t, got, "availableReplicas", 0)
}

func TestAggregateDaemonSetStatusWithZeroAvailableWEC(t *testing.T) {
	healthy := toStatusMap(t, &appsv1.DaemonSetStatus{
		ObservedGeneration: 1, CurrentNumberScheduled: 2, DesiredNumberScheduled: 2,
		NumberReady: 2, UpdatedNumberScheduled: 2, NumberAvailable: 2,
	})
	notReady := toStatusMap(t, &appsv1.DaemonSetStatus{
		ObservedGeneration: 1, CurrentNumberScheduled: 2, DesiredNumberScheduled: 2,
		NumberUnavailable: 2,
	})

	got, err := AggregateDaemonSetStatus([]map[string]any{healthy, notReady})
	if err != nil {
		t.Fatal(err)
	}
	expectInt(t, got, "numberReady", 0)
	expectInt(t, got, "updatedNumberScheduled", 0)
	expectInt(t, got, "numberAvailable", 0)
	expectInt(t, got, "desiredNumberScheduled", 2)
}
