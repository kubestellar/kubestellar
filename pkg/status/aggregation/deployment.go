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

func AggregateDeploymentStatus(statuses []map[string]interface{}) (map[string]interface{}, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	if len(statuses) == 1 {
		return statuses[0], nil
	}

	aggregatedStatus := make(map[string]interface{})

	aggregatedStatus["replicas"] = GetMin(statuses, "replicas")
	aggregatedStatus["updatedReplicas"] = GetMin(statuses, "updatedReplicas")
	aggregatedStatus["availableReplicas"] = GetMin(statuses, "availableReplicas")
	aggregatedStatus["readyReplicas"] = GetMin(statuses, "readyReplicas")
	aggregatedStatus["unavailableReplicas"] = GetMax(statuses, "unavailableReplicas")
	aggregatedStatus["observedGeneration"] = GetMin(statuses, "observedGeneration")

	aggregatedStatus["conditions"] = aggregateConditions(statuses)

	return aggregatedStatus, nil
}

func aggregateConditions(statuses []map[string]interface{}) []interface{} {
	// we focus on type Progressing with containing reason ProgressDeadlineExceeded which is checked by Argo for determining health
	// If any of the statuses contains it we return its condition as aggregated condition

	for _, status := range statuses {
		conditions, ok := status["conditions"].([]interface{})

		if !ok {
			continue
		}

		// cond is the single elements from conditions
		for _, cond := range conditions {
			c, ok := cond.(map[string]interface{})
			if !ok {
				continue
			}

			if c["type"] == "Progressing" && c["reason"] == "ProgressDeadlineExceeded" {
				return []interface{}{cond}
			}

		}
	}

	return []interface{}{}

}
