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
