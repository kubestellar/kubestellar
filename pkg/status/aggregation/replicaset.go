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

func AggregateReplicaSetStatus(statuses []map[string]any) (map[string]any, error) {

	aggregatedStatus := make(map[string]any)

	aggregatedStatus["replicas"] = GetMin(statuses, "replicas")
	aggregatedStatus["fullyLabeledReplicas"] = GetMin(statuses, "fullyLabeledReplicas")
	aggregatedStatus["readyReplicas"] = GetMin(statuses, "readyReplicas")
	aggregatedStatus["availableReplicas"] = GetMin(statuses, "availableReplicas")
	aggregatedStatus["observedGeneration"] = GetMin(statuses, "observedGeneration")

	aggregatedStatus["conditions"] = aggregateReplicaSetConditions(statuses)

	return aggregatedStatus, nil
}

func aggregateReplicaSetConditions(statuses []map[string]any) []any {
	// Focus on type ReplicaFailure condition which indicates issues with replica creation
	// If any of the statuses contains it we return its condition as aggregated condition

	for _, status := range statuses {
		conditions, ok := status["conditions"].([]any)

		if !ok {
			continue
		}

		// cond is the single elements from conditions
		for _, cond := range conditions {
			c, ok := cond.(map[string]any)
			if !ok {
				continue
			}

			if c["type"] == "ReplicaFailure" {
				return []any{cond}
			}

		}
	}

	return []any{}

}
