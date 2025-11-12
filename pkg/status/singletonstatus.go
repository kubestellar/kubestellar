/*
Copyright 2024 The KubeStellar Authors.

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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

func (c *Controller) updateWorkStatusToObject(ctx context.Context, workStatusON cache.ObjectName) error {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Reconciling singleton status due to workstatus changes", "workStatus", workStatusON)
	wsObj, err := c.workStatusLister.ByNamespace(workStatusON.Namespace).Get(workStatusON.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			wsObj = nil
		} else {
			return err
		}
	}
	if wsObj == nil {
		objId, had := c.workStatusToObject.Delete(workStatusON)
		if had {
			logger.V(5).Info("Enqueuing workload object due to absense of WorkStatus", "workStatus", workStatusON, "workloadObject", objId)
			c.workqueue.Add(workloadObjectRef{objId})
		}
	} else {
		source, err := util.GetWorkStatusSourceRef(wsObj)
		if err != nil {
			return err
		}
		objId := util.ObjectIdentifierFromSourceRef(source)
		formerSource, had := c.workStatusToObject.Put(workStatusON, objId)
		logger.V(5).Info("Enqueuing workload object due to existence of WorkStatus", "workStatus", workStatusON, "workloadObject", objId)
		c.workqueue.Add(workloadObjectRef{objId})
		if had {
			logger.V(5).Info("Enqueuing workload object due to change in WorkStatus", "workStatus", workStatusON, "workloadObject", formerSource)
			c.workqueue.Add(workloadObjectRef{formerSource})
		}
	}
	return nil
}

func (c *Controller) syncWorkloadObject(ctx context.Context, wObjID util.ObjectIdentifier) error {
	logger := klog.FromContext(ctx)
	isSingletonRequested, nWECs := c.bindingPolicyResolver.GetSingletonReportedStateRequestForObject(wObjID)

	isMultiWECRequested, nWECsMulti := c.bindingPolicyResolver.GetMultiWECReportedStateRequestForObject(wObjID)

	logger.V(4).Info("Workload object (singleton status requested)", "object", wObjID, "isSingletonRequested", isSingletonRequested, "nWECs", nWECs,
		"Workload object (multiWEC status requested)", "object", wObjID, "isMultiWECRequested", isMultiWECRequested, "nWECsMulti", nWECsMulti)

	if (isMultiWECRequested || isSingletonRequested) && nWECs == 1 {
		logger.V(4).Info("Either singleton or multiWEC status is requested and nWEC == 1", "object: ", wObjID, "nWEC", nWECs)
		return c.handleSingleton(ctx, wObjID)
	}

	if isMultiWECRequested && nWECs > 1 {
		logger.V(4).Info("multiWEC status is requested and nWEC != 1", "object: ", wObjID, "nWEC", nWECs)
		return c.handleMultiWEC(ctx, wObjID, nWECs)
	}

	logger.V(4).Info("None of the condition mentioned in doc for execution of handleSingleton and handleMultiWEC function is matched.", wObjID)
	if err := c.updateObjectStatus(ctx, wObjID, nil, c.listers); err != nil {
		return err
	}

	return nil
}

func (c *Controller) handleSingleton(ctx context.Context, wObjID util.ObjectIdentifier) error {
	logger := klog.FromContext(ctx)
	var wsONs [2]cache.ObjectName
	var numWS int

	c.workStatusToObject.ReadInverse().ContGet(wObjID, func(wsONSet sets.Set[cache.ObjectName]) {
		numWS = wsONSet.Len()
		var idx int
		for it := range wsONSet {
			wsONs[idx] = it
			idx++
			if idx > 1 {
				break
			}
		}
	})

	if numWS != 1 {
		if err := c.updateObjectStatus(ctx, wObjID, nil, c.listers); err != nil {
			return err
		}
		logger.V(4).Info("Cleaned singleton status for workload object",
			"object", wObjID, "numWS", numWS)
		return nil
	}
	wsON := wsONs[0]
	wsObj, err := c.workStatusLister.ByNamespace(wsON.Namespace).Get(wsON.Name)
	if err != nil {
		return err
	}
	status, err := util.GetWorkStatusStatus(wsObj)
	if err != nil {
		return err
	}
	if status == nil {
		return nil
	}
	if err := c.updateObjectStatus(ctx, wObjID, status, c.listers); err != nil {
		return err
	}
	logger.V(4).Info("Updated singleton status for workload object", "objId", wObjID, "workStatus", wsON)
	return nil
}
