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

func AggregateJobStatus(statuses []map[string]any) (map[string]any, error) {

	aggregatedStatus := make(map[string]any)

	// Aggregate Job numeric fields using minimum, per spec
	aggregatedStatus["active"] = GetMin(statuses, "active")
	aggregatedStatus["succeeded"] = GetMin(statuses, "succeeded")
	aggregatedStatus["failed"] = GetMin(statuses, "failed")

	// Aggregate conditions using the generic condition aggregation rules
	aggregatedStatus["conditions"] = AggregateConditions(statuses)

	return aggregatedStatus, nil
}
