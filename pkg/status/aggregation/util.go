package aggregation

import "math"

func GetMin(statuses []map[string]interface{}, field string) int64 {
	min := int64(math.MaxInt64)
	found := false
	for _, status := range statuses {
		if val, ok := status[field].(int64); ok {
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
		if val, ok := status[field].(int64); ok {
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
