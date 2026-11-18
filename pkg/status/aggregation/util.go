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

import "math"

// fieldValue returns the int64 value of field in status. Kubernetes omits
// omitempty integer status fields (such as readyReplicas) when they are 0, so
// a missing field counts as 0 rather than being skipped.
func fieldValue(status map[string]interface{}, field string) (int64, bool) {
	raw, present := status[field]
	if !present || raw == nil {
		return 0, true
	}
	val, ok := raw.(int64)
	return val, ok
}

func GetMin(statuses []map[string]interface{}, field string) int64 {
	min := int64(math.MaxInt64)
	found := false
	for _, status := range statuses {
		if val, ok := fieldValue(status, field); ok {
			if val < min {
				min = val
			}
			found = true
		}
	}
	if !found {
		return 0
	}
	return min
}

func GetMax(statuses []map[string]interface{}, field string) int64 {
	max := int64(math.MinInt64)
	found := false
	for _, status := range statuses {
		if val, ok := fieldValue(status, field); ok {
			if val > max {
				max = val
			}
			found = true
		}
	}
	if !found {
		return 0
	}
	return max
}
