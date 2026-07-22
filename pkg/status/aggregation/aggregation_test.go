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
	"reflect"
	"testing"
)

func TestGetMinAndMax(t *testing.T) {
	statuses := []map[string]interface{}{
		{"replicas": int64(3), "missed": int64(1)},
		{"replicas": int64(1), "missed": int64(5)},
		{"replicas": int64(5), "missed": int64(2)},
		{"not_int": "value"},
	}

	// Test GetMin
	if got := GetMin(statuses, "replicas"); got != 1 {
		t.Errorf("GetMin(replicas) expected 1, got %d", got)
	}
	if got := GetMin(statuses, "nonexistent"); got != 0 {
		t.Errorf("GetMin(nonexistent) expected 0, got %d", got)
	}

	// Test GetMax
	if got := GetMax(statuses, "missed"); got != 5 {
		t.Errorf("GetMax(missed) expected 5, got %d", got)
	}
	if got := GetMax(statuses, "nonexistent"); got != 0 {
		t.Errorf("GetMax(nonexistent) expected 0, got %d", got)
	}
}

func TestAggregateDeploymentStatus(t *testing.T) {
	testCases := []struct {
		name     string
		statuses []map[string]any
		expected map[string]any
	}{
		{
			name: "basic aggregation",
			statuses: []map[string]any{
				{
					"replicas":            int64(3),
					"updatedReplicas":     int64(3),
					"availableReplicas":   int64(2),
					"readyReplicas":       int64(3),
					"unavailableReplicas": int64(1),
					"observedGeneration":  int64(5),
					"conditions": []any{
						map[string]any{"type": "Available", "status": "True"},
					},
				},
				{
					"replicas":            int64(2),
					"updatedReplicas":     int64(1),
					"availableReplicas":   int64(1),
					"readyReplicas":       int64(1),
					"unavailableReplicas": int64(2),
					"observedGeneration":  int64(4),
					"conditions": []any{
						map[string]any{"type": "Progressing", "status": "True"},
					},
				},
			},
			expected: map[string]any{
				"replicas":            int64(2),
				"updatedReplicas":     int64(1),
				"availableReplicas":   int64(1),
				"readyReplicas":       int64(1),
				"unavailableReplicas": int64(2),
				"observedGeneration":  int64(4),
				"conditions":          []any{},
			},
		},
		{
			name: "aggregation with progress deadline exceeded",
			statuses: []map[string]any{
				{
					"replicas":            int64(3),
					"updatedReplicas":     int64(3),
					"availableReplicas":   int64(3),
					"readyReplicas":       int64(3),
					"unavailableReplicas": int64(0),
					"observedGeneration":  int64(2),
					"conditions": []any{
						map[string]any{"type": "Progressing", "reason": "ProgressDeadlineExceeded"},
					},
				},
				{
					"replicas":            int64(3),
					"updatedReplicas":     int64(3),
					"availableReplicas":   int64(3),
					"readyReplicas":       int64(3),
					"unavailableReplicas": int64(0),
					"observedGeneration":  int64(2),
					"conditions": []any{
						map[string]any{"type": "Available", "status": "True"},
					},
				},
			},
			expected: map[string]any{
				"replicas":            int64(3),
				"updatedReplicas":     int64(3),
				"availableReplicas":   int64(3),
				"readyReplicas":       int64(3),
				"unavailableReplicas": int64(0),
				"observedGeneration":  int64(2),
				"conditions": []any{
					map[string]any{"type": "Progressing", "reason": "ProgressDeadlineExceeded"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AggregateDeploymentStatus(tc.statuses)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("expected:\n%v\ngot:\n%v", tc.expected, got)
			}
		})
	}
}

func TestAggregateReplicaSetStatus(t *testing.T) {
	statuses := []map[string]any{
		{
			"replicas":             int64(4),
			"fullyLabeledReplicas": int64(4),
			"readyReplicas":        int64(3),
			"availableReplicas":    int64(3),
			"observedGeneration":   int64(2),
			"conditions": []any{
				map[string]any{"type": "ReplicaFailure", "status": "True"},
			},
		},
		{
			"replicas":             int64(3),
			"fullyLabeledReplicas": int64(3),
			"readyReplicas":        int64(2),
			"availableReplicas":    int64(2),
			"observedGeneration":   int64(3),
			"conditions": []any{
				map[string]any{"type": "Progressing", "status": "True"},
			},
		},
	}

	expected := map[string]any{
		"replicas":             int64(3),
		"fullyLabeledReplicas": int64(3),
		"readyReplicas":        int64(2),
		"availableReplicas":    int64(2),
		"observedGeneration":   int64(2),
		"conditions": []any{
			map[string]any{"type": "ReplicaFailure", "status": "True"},
		},
	}

	got, err := AggregateReplicaSetStatus(statuses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected:\n%v\ngot:\n%v", expected, got)
	}
}

func TestAggregateDaemonSetStatus(t *testing.T) {
	statuses := []map[string]any{
		{
			"currentNumberScheduled": int64(3),
			"numberMisscheduled":     int64(0),
			"desiredNumberScheduled": int64(3),
			"numberReady":            int64(2),
			"observedGeneration":     int64(2),
			"updatedNumberScheduled": int64(2),
			"numberAvailable":        int64(2),
		},
		{
			"currentNumberScheduled": int64(2),
			"numberMisscheduled":     int64(1),
			"desiredNumberScheduled": int64(2),
			"numberReady":            int64(1),
			"observedGeneration":     int64(3),
			"updatedNumberScheduled": int64(1),
			"numberAvailable":        int64(1),
		},
	}

	expected := map[string]any{
		"currentNumberScheduled": int64(2),
		"numberMisscheduled":     int64(1),
		"desiredNumberScheduled": int64(2),
		"numberReady":            int64(1),
		"observedGeneration":     int64(2),
		"updatedNumberScheduled": int64(1),
		"numberAvailable":        int64(1),
	}

	got, err := AggregateDaemonSetStatus(statuses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("expected:\n%v\ngot:\n%v", expected, got)
	}
}
