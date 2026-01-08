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

func AggregateStatefulSetStatus(statuses []map[string]any) (map[string]any, error) {

	aggregatedStatus := make(map[string]any)

	// Aggregate readiness-related numeric fields using minimum, per spec
	aggregatedStatus["readyReplicas"] = GetMin(statuses, "readyReplicas")
	aggregatedStatus["currentReplicas"] = GetMin(statuses, "currentReplicas")
	aggregatedStatus["updatedReplicas"] = GetMin(statuses, "updatedReplicas")
	aggregatedStatus["observedGeneration"] = GetMin(statuses, "observedGeneration")

	// Aggregate conditions using the generic condition aggregation rules
	aggregatedStatus["conditions"] = AggregateConditions(statuses)

	// Aggregate revision fields only if all WECs report the same value
	if value, ok := GetCommonString(statuses, "currentRevision"); ok {
		aggregatedStatus["currentRevision"] = value
	}

	if value, ok := GetCommonString(statuses, "updateRevision"); ok {
		aggregatedStatus["updateRevision"] = value
	}

	return aggregatedStatus, nil
}
