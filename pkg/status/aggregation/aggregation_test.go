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
)

func TestGetMin(t *testing.T) {
	cases := []struct {
		name     string
		statuses []map[string]interface{}
		field    string
		want     int64
	}{
		{
			name:     "empty statuses returns 0",
			statuses: nil,
			field:    "replicas",
			want:     0,
		},
		{
			name:     "field absent in all statuses returns 0",
			statuses: []map[string]interface{}{{"other": int64(5)}},
			field:    "replicas",
			want:     0,
		},
		{
			name:     "single value",
			statuses: []map[string]interface{}{{"replicas": int64(3)}},
			field:    "replicas",
			want:     3,
		},
		{
			name: "picks minimum",
			statuses: []map[string]interface{}{
				{"replicas": int64(5)},
				{"replicas": int64(2)},
				{"replicas": int64(8)},
			},
			field: "replicas",
			want:  2,
		},
		{
			name: "ignores entries where field has wrong type",
			statuses: []map[string]interface{}{
				{"replicas": "not-an-int"},
				{"replicas": int64(4)},
			},
			field: "replicas",
			want:  4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetMin(tc.statuses, tc.field)
			if got != tc.want {
				t.Errorf("GetMin() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestGetMax(t *testing.T) {
	cases := []struct {
		name     string
		statuses []map[string]interface{}
		field    string
		want     int64
	}{
		{
			name:     "empty statuses returns 0",
			statuses: nil,
			field:    "unavailableReplicas",
			want:     0,
		},
		{
			name:     "field absent returns 0",
			statuses: []map[string]interface{}{{"other": int64(1)}},
			field:    "unavailableReplicas",
			want:     0,
		},
		{
			name: "picks maximum",
			statuses: []map[string]interface{}{
				{"unavailableReplicas": int64(1)},
				{"unavailableReplicas": int64(5)},
				{"unavailableReplicas": int64(3)},
			},
			field: "unavailableReplicas",
			want:  5,
		},
		{
			name: "ignores entries where field has wrong type",
			statuses: []map[string]interface{}{
				{"unavailableReplicas": "bad"},
				{"unavailableReplicas": int64(7)},
			},
			field: "unavailableReplicas",
			want:  7,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := GetMax(tc.statuses, tc.field)
			if got != tc.want {
				t.Errorf("GetMax() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestAggregateDeploymentStatus(t *testing.T) {
	t.Run("aggregates numeric fields correctly", func(t *testing.T) {
		statuses := []map[string]any{
			{
				"replicas":            int64(5),
				"availableReplicas":   int64(4),
				"readyReplicas":       int64(3),
				"updatedReplicas":     int64(5),
				"unavailableReplicas": int64(1),
				"observedGeneration":  int64(2),
			},
			{
				"replicas":            int64(3),
				"availableReplicas":   int64(2),
				"readyReplicas":       int64(2),
				"updatedReplicas":     int64(3),
				"unavailableReplicas": int64(3),
				"observedGeneration":  int64(1),
			},
		}
		got, err := AggregateDeploymentStatus(statuses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertInt64Field(t, got, "replicas", 3)
		assertInt64Field(t, got, "availableReplicas", 2)
		assertInt64Field(t, got, "readyReplicas", 2)
		assertInt64Field(t, got, "updatedReplicas", 3)
		assertInt64Field(t, got, "unavailableReplicas", 3)
		assertInt64Field(t, got, "observedGeneration", 1)
	})

	t.Run("returns ProgressDeadlineExceeded condition when present", func(t *testing.T) {
		deadlineCondition := map[string]any{
			"type":   "Progressing",
			"reason": "ProgressDeadlineExceeded",
			"status": "False",
		}
		statuses := []map[string]any{
			{"conditions": []any{deadlineCondition}},
			{"conditions": []any{map[string]any{"type": "Available", "status": "True"}}},
		}
		got, err := AggregateDeploymentStatus(statuses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		conditions, ok := got["conditions"].([]any)
		if !ok {
			t.Fatal("conditions field missing or wrong type")
		}
		if len(conditions) != 1 {
			t.Fatalf("expected 1 condition, got %d", len(conditions))
		}
		cond, ok := conditions[0].(map[string]any)
		if !ok {
			t.Fatal("condition is not a map")
		}
		if cond["reason"] != "ProgressDeadlineExceeded" {
			t.Errorf("expected ProgressDeadlineExceeded, got %v", cond["reason"])
		}
	})

	t.Run("returns empty conditions when no ProgressDeadlineExceeded", func(t *testing.T) {
		statuses := []map[string]any{
			{"conditions": []any{map[string]any{"type": "Available", "status": "True"}}},
		}
		got, err := AggregateDeploymentStatus(statuses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		conditions, ok := got["conditions"].([]any)
		if !ok {
			t.Fatal("conditions missing")
		}
		if len(conditions) != 0 {
			t.Errorf("expected empty conditions, got %d", len(conditions))
		}
	})
}

func TestAggregateDaemonSetStatus(t *testing.T) {
	statuses := []map[string]any{
		{
			"currentNumberScheduled": int64(10),
			"numberMisscheduled":     int64(2),
			"desiredNumberScheduled": int64(10),
			"numberReady":            int64(8),
			"observedGeneration":     int64(3),
			"updatedNumberScheduled": int64(10),
			"numberAvailable":        int64(8),
		},
		{
			"currentNumberScheduled": int64(5),
			"numberMisscheduled":     int64(5),
			"desiredNumberScheduled": int64(5),
			"numberReady":            int64(3),
			"observedGeneration":     int64(2),
			"updatedNumberScheduled": int64(4),
			"numberAvailable":        int64(3),
		},
	}
	got, err := AggregateDaemonSetStatus(statuses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertInt64Field(t, got, "currentNumberScheduled", 5)
	assertInt64Field(t, got, "numberMisscheduled", 5)
	assertInt64Field(t, got, "desiredNumberScheduled", 5)
	assertInt64Field(t, got, "numberReady", 3)
	assertInt64Field(t, got, "observedGeneration", 2)
	assertInt64Field(t, got, "updatedNumberScheduled", 4)
	assertInt64Field(t, got, "numberAvailable", 3)
}

func TestAggregateReplicaSetStatus(t *testing.T) {
	t.Run("aggregates numeric fields", func(t *testing.T) {
		statuses := []map[string]any{
			{
				"replicas":             int64(6),
				"fullyLabeledReplicas": int64(6),
				"readyReplicas":        int64(5),
				"availableReplicas":    int64(5),
				"observedGeneration":   int64(4),
			},
			{
				"replicas":             int64(4),
				"fullyLabeledReplicas": int64(3),
				"readyReplicas":        int64(2),
				"availableReplicas":    int64(2),
				"observedGeneration":   int64(3),
			},
		}
		got, err := AggregateReplicaSetStatus(statuses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertInt64Field(t, got, "replicas", 4)
		assertInt64Field(t, got, "fullyLabeledReplicas", 3)
		assertInt64Field(t, got, "readyReplicas", 2)
		assertInt64Field(t, got, "availableReplicas", 2)
		assertInt64Field(t, got, "observedGeneration", 3)
	})

	t.Run("surfaces ReplicaFailure condition", func(t *testing.T) {
		failureCondition := map[string]any{
			"type":   "ReplicaFailure",
			"status": "True",
			"reason": "FailedCreate",
		}
		statuses := []map[string]any{
			{"conditions": []any{failureCondition}},
			{"conditions": []any{map[string]any{"type": "Available", "status": "True"}}},
		}
		got, err := AggregateReplicaSetStatus(statuses)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		conditions, ok := got["conditions"].([]any)
		if !ok || len(conditions) != 1 {
			t.Fatalf("expected 1 condition, got %v", got["conditions"])
		}
		cond := conditions[0].(map[string]any)
		if cond["type"] != "ReplicaFailure" {
			t.Errorf("expected ReplicaFailure, got %v", cond["type"])
		}
	})

	t.Run("empty conditions when no ReplicaFailure", func(t *testing.T) {
		statuses := []map[string]any{
			{"conditions": []any{map[string]any{"type": "Available", "status": "True"}}},
		}
		got, _ := AggregateReplicaSetStatus(statuses)
		conditions := got["conditions"].([]any)
		if len(conditions) != 0 {
			t.Errorf("expected 0 conditions, got %d", len(conditions))
		}
	})
}

func assertInt64Field(t *testing.T, m map[string]any, field string, want int64) {
	t.Helper()
	v, ok := m[field]
	if !ok {
		t.Errorf("field %q missing from result", field)
		return
	}
	got, ok := v.(int64)
	if !ok {
		t.Errorf("field %q: expected int64, got %T", field, v)
		return
	}
	if got != want {
		t.Errorf("field %q: got %d, want %d", field, got, want)
	}
}
