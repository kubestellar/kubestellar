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
	isSingletonRequested, qualifiedWECsSingleton, isMultiWECRequested, qualifiedWECsMulti := c.bindingPolicyResolver.GetReportedStateRequestForObject(wObjID)

	logger.V(4).Info("Workload object reported state request", "object", wObjID, "isSingletonRequested", isSingletonRequested, "qualifiedWECsSingleton", util.K8sSet4Log(qualifiedWECsSingleton),
		"isMultiWECRequested", isMultiWECRequested, "qualifiedWECsMulti", util.K8sSet4Log(qualifiedWECsMulti))

	if isMultiWECRequested && isSingletonRequested {
		// if both are requested for same object then we can combine all qualified clusters(qualifiedWECsSingleton and qualifiedWECsMulti)
		// then call handleMultiWEC function
		// TODO: Implement combine all qualified WECs.
		if qualifiedWECsSingleton.Len() == 1 && qualifiedWECsMulti.Len() == 1 {
			return c.handleSingleton(ctx, wObjID, qualifiedWECsSingleton)
		}
		return c.handleMultiWEC(ctx, wObjID, qualifiedWECsMulti)
	}

	if (isSingletonRequested && qualifiedWECsSingleton.Len() == 1) || (isMultiWECRequested && qualifiedWECsMulti.Len() == 1) {
		qualifiedWECs := qualifiedWECsSingleton
		if isMultiWECRequested && qualifiedWECsMulti.Len() == 1 {
			qualifiedWECs = qualifiedWECsMulti
		}
		return c.handleSingleton(ctx, wObjID, qualifiedWECs)
	}

	if isMultiWECRequested && qualifiedWECsMulti.Len() > 0 {
		return c.handleMultiWEC(ctx, wObjID, qualifiedWECsMulti)
	}

	logger.V(4).Info("neither singleton nor multi-WEC reported state return applies", "object", wObjID)
	if err := c.updateObjectStatus(ctx, wObjID, nil, c.listers); err != nil {
		return err
	}

	return nil
}

func (c *Controller) handleSingleton(ctx context.Context, wObjID util.ObjectIdentifier, qualifiedWEC sets.Set[string]) error {
	logger := klog.FromContext(ctx)
	var wsON cache.ObjectName
	var numWS int

	// Get the WEC name
	var qualifiedWECName string
	for wecName := range qualifiedWEC {
		qualifiedWECName = wecName
		break
	}

	c.workStatusToObject.ReadInverse().ContGet(wObjID, func(wsONSet sets.Set[cache.ObjectName]) {
		for it := range wsONSet {
			if it.Namespace == qualifiedWECName {
				wsON = it
				numWS++
				if numWS > 1 {
					break
				}
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
