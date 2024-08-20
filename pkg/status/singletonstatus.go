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

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

func (c *Controller) reconcileSingletonByBdg(ctx context.Context, bdgName string) error {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Reconciling singleton status due to binding change", "binding", bdgName)

	wObjIDs, requested, nWECs := c.bindingPolicyResolver.GetSingletonReportedStateRequestsForBinding(bdgName)
	for i := range wObjIDs {
		wObjID, r, n := wObjIDs[i], requested[i], nWECs[i]
		if !r || n != 1 {
			logger.V(4).Info("Singleton workload object should not have status synced, cleaning",
				"resource", wObjID.Resource, "objectName", wObjID.ObjectName,
				"requested", r, "nWECs", n)
			if err := c.reconcileSingletonWObj(ctx, wObjID, false); err != nil {
				return err
			}
		} else {
			logger.V(4).Info("Singleton workload object should have status synced, updating",
				"resource", wObjID.Resource, "objectName", wObjID.ObjectName)
			if err := c.reconcileSingletonWObj(ctx, wObjID, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Controller) reconcileSingletonByWS(ctx context.Context, ref singletonWorkStatusRef) error {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Reconciling singleton status due to workstatus changes", "name", string(ref.Name))
	wObjID := ref.SourceObjectIdentifier

	requested, nWECs := c.bindingPolicyResolver.GetSingletonReportedStateRequestForObject(wObjID)
	if !requested || nWECs != 1 {
		logger.V(4).Info("Singleton workload object should not have status synced, cleaning",
			"resource", ref.SourceObjectIdentifier.Resource, "objectName", ref.SourceObjectIdentifier.ObjectName,
			"requested", requested, "nWECs", nWECs)
		return c.reconcileSingletonWObj(ctx, wObjID, false)
	} else {
		logger.V(4).Info("Singleton workload object should have status synced, updating",
			"resource", ref.SourceObjectIdentifier.Resource, "objectName", ref.SourceObjectIdentifier.ObjectName)
		wsObj, err := c.workStatusLister.ByNamespace(ref.WECName).Get(ref.Name)
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
		logger.V(4).Info("Updating singleton status", "workstatus", ref.Name, "wecName", ref.WECName)
		return updateObjectStatus(ctx, &wObjID, status, c.listers, c.wdsDynClient)
	}
}

func (c *Controller) reconcileSingletonWObj(ctx context.Context, wObjID util.ObjectIdentifier, sync bool) error {
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Reconciling singleton workload object", "resource", wObjID.Resource, "objectName", wObjID.ObjectName)

	if !sync {
		emptyStatus := make(map[string]interface{})
		logger.V(4).Info("Cleaning up singleton status", "resource", wObjID.Resource, "objectName", wObjID.ObjectName)
		return updateObjectStatus(ctx, &wObjID, emptyStatus, c.listers, c.wdsDynClient)
	}

	list, _ := c.workStatusLister.ByNamespace("").List(labels.Everything())
	for _, obj := range list {
		wsRef, err := runtimeObjectToWorkStatusRef(obj)
		if err != nil {
			return err
		}
		sourceObjID := wsRef.SourceObjectIdentifier
		if sourceObjID != wObjID {
			continue
		}
		status, err := util.GetWorkStatusStatus(obj)
		if err != nil {
			return err
		}
		if status == nil {
			return nil
		}
		logger.V(4).Info("Updating singleton status", "workstatus", wsRef.Name, "wecName", wsRef.WECName)
		return updateObjectStatus(ctx, &wsRef.SourceObjectIdentifier, status, c.listers, c.wdsDynClient)
	}
	return nil
}
