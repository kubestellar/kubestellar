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
	"sort"
	"time"
)

// GetCommonString returns the string value for the given field
// only if all statuses contain the same non-empty value.
func GetCommonString(statuses []map[string]any, field string) (string, bool) {
	if len(statuses) == 0 {
		return "", false
	}

	value, _ := statuses[0][field].(string)
	if value == "" {
		return "", false
	}

	for _, status := range statuses[1:] {
		v, _ := status[field].(string)
		if v != value {
			return "", false
		}
	}

	return value, true
}

// AggregateConditions aggregates conditions across multiple WECs
// using the generic condition aggregation rules defined in the
// Multi-WEC Aggregated Status design.
func AggregateConditions(statuses []map[string]any) []any {
	conditionSets := make(map[string][]map[string]any)

	for _, status := range statuses {
		conds, ok := status["conditions"].([]any)
		if !ok {
			continue
		}

		for _, c := range conds {
			cond, ok := c.(map[string]any)
			if !ok {
				continue
			}

			condType, ok := cond["type"].(string)
			if !ok || condType == "" {
				continue
			}

			conditionSets[condType] = append(conditionSets[condType], cond)
		}
	}

	// Sort condition types for deterministic output order
	types := make([]string, 0, len(conditionSets))
	for t := range conditionSets {
		types = append(types, t)
	}
	sort.Strings(types)

	result := make([]any, 0, len(conditionSets))

	for _, t := range types {
		if agg := aggregateConditionSet(conditionSets[t]); agg != nil {
			result = append(result, agg)
		}
	}

	return result
}

func aggregateConditionSet(conds []map[string]any) map[string]any {
	var (
		hasFalse   bool
		hasUnknown bool
		latestTime time.Time
		latestCond map[string]any
	)

	for _, c := range conds {
		status, _ := c["status"].(string)

		switch status {
		case "False":
			hasFalse = true
		case "Unknown":
			hasUnknown = true
		}

		ts, _ := c["lastTransitionTime"].(string)
		if ts == "" {
			continue
		}

		t, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			continue
		}

		if latestCond == nil || t.After(latestTime) {
			latestTime = t
			latestCond = c
		}
	}

	if latestCond == nil {
		return nil
	}

	finalStatus := "True"
	if hasFalse {
		finalStatus = "False"
	} else if hasUnknown {
		finalStatus = "Unknown"
	}

	aggregated := make(map[string]any)
	for k, v := range latestCond {
		aggregated[k] = v
	}
	aggregated["status"] = finalStatus

	return aggregated
}
