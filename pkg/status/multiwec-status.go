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

package status

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/status/aggregation"
	"github.com/kubestellar/kubestellar/pkg/util"
)

func (c *Controller) handleMultiWEC(ctx context.Context, wObjID util.ObjectIdentifier, qualifiedWEC sets.Set[string]) error {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Implement multiwec handling logic", "object", wObjID, "qualifiedWEC", util.K8sSet4Log(qualifiedWEC))

	var wsObjects []cache.ObjectName
	c.workStatusToObject.ReadInverse().ContGet(wObjID, func(wsONSet sets.Set[cache.ObjectName]) {
		wsObjects = make([]cache.ObjectName, 0, qualifiedWEC.Len())
		for wsON := range wsONSet {
			if qualifiedWEC.Has(wsON.Namespace) {
				wsObjects = append(wsObjects, wsON)
			}
		}
	})

	if len(wsObjects) == 0 {
		if err := c.updateObjectStatus(ctx, wObjID, nil, c.listers, true); err != nil {
			return err
		}
		logger.V(4).Info("No workstatus found for workload object",
			"object", wObjID, "numWS", len(wsObjects))
		return nil
	}

	// collect status
	statuses := make([]map[string]any, 0, len(wsObjects))
	for _, wsON := range wsObjects {
		wsObj, err := c.workStatusLister.ByNamespace(wsON.Namespace).Get(wsON.Name)
		if err != nil {
			logger.V(4).Info("Failed to get WorkStatus", "workStatus", wsON, "error", err)
			continue
		}
		status, err := util.GetWorkStatusStatus(wsObj)
		if err != nil {
			logger.V(4).Info("Failed to extract status from WorkStatus", "workStatus", wsON, "error", err)
			continue
		}
		if status == nil {
			continue
		}
		statuses = append(statuses, status)
	}

	if len(statuses) == 0 {
		if err := c.updateObjectStatus(ctx, wObjID, nil, c.listers, true); err != nil {
			return err
		}
		logger.V(4).Info("No workstatus found for workload object",
			"object", wObjID, "numWS", len(wsObjects))
		return nil
	}

	if len(statuses) == 1 {
		if err := c.updateObjectStatus(ctx, wObjID, statuses[0], c.listers, false); err != nil {
			return err
		}
		return nil
	}

	// Aggregate for known kind that Kubestellar handle specially which are mainly kind available as built-in healthchecks for Argocd
	var aggregatedStatus map[string]any
	var errAggregate error

	switch wObjID.GVK.Group {
	case appsv1.GroupName:
		switch wObjID.GVK.Kind {
		case "Deployment":
			aggregatedStatus, errAggregate = aggregation.AggregateDeploymentStatus(statuses)
		case "ReplicaSet":
			aggregatedStatus, errAggregate = aggregation.AggregateReplicaSetStatus(statuses)
		case "DaemonSet":
			aggregatedStatus, errAggregate = aggregation.AggregateDaemonSetStatus(statuses)
		}
	default:
		// for generic

	}

	if errAggregate != nil {
		return errAggregate
	}
	if aggregatedStatus == nil {
		return nil
	}

	if err := c.updateObjectStatus(ctx, wObjID, aggregatedStatus, c.listers, true); err != nil {
		return err
	}

	logger.V(4).Info("Updated multiWEC aggregated status for workload object", "objId", wObjID, "workStatus", wsObjects)

	return nil
}
