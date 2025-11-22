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

func AggregateDaemonSetStatus(statuses []map[string]interface{}) (map[string]interface{}, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	if len(statuses) == 1 {
		return statuses[0], nil
	}

	aggregatedStatus := make(map[string]interface{})

	aggregatedStatus["currentNumberScheduled"] = GetMin(statuses, "currentNumberScheduled")
	aggregatedStatus["numberMisscheduled"] = GetMax(statuses, "numberMisscheduled")
	aggregatedStatus["desiredNumberScheduled"] = GetMin(statuses, "desiredNumberScheduled")
	aggregatedStatus["numberReady"] = GetMin(statuses, "numberReady")
	aggregatedStatus["observedGeneration"] = GetMin(statuses, "observedGeneration")
	aggregatedStatus["updatedNumberScheduled"] = GetMin(statuses, "updatedNumberScheduled")
	aggregatedStatus["numberAvailable"] = GetMin(statuses, "numberAvailable")

	return aggregatedStatus, nil
}
