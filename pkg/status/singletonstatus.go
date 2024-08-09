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
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

type singletonState struct {
	sync.RWMutex
	wObjSync map[util.ObjectIdentifier]bool
}

func (ss *singletonState) get(id util.ObjectIdentifier) bool {
	ss.Lock()
	defer ss.Unlock()

	return ss.wObjSync[id]
}

func (ss *singletonState) set(id util.ObjectIdentifier) {
	ss.Lock()
	defer ss.Unlock()

	ss.wObjSync[id] = true
}

func (ss *singletonState) unset(id util.ObjectIdentifier) {
	ss.Lock()
	defer ss.Unlock()

	ss.wObjSync[id] = false
}

func (ss *singletonState) addIfNotExist(id util.ObjectIdentifier) {
	ss.Lock()
	defer ss.Unlock()

	if _, ok := ss.wObjSync[id]; !ok {
		ss.wObjSync[id] = false
	}
}

func (ss *singletonState) delete(id util.ObjectIdentifier) {
	ss.Lock()
	defer ss.Unlock()

	delete(ss.wObjSync, id)
}

func (c *Controller) buildSingletonStateAndOptionallyReconcile(ctx context.Context, reconcile bool) error {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Building the desired state for singleton statuses")

	for wObjIdentifier := range c.singletonState.wObjSync {
		gvr := schema.GroupVersionResource{Group: wObjIdentifier.GVK.Group, Version: wObjIdentifier.GVK.Version, Resource: wObjIdentifier.Resource}
		lister, found := c.listers.Get(gvr)
		if !found {
			return fmt.Errorf("could not get lister for gvr: %s", gvr)
		}
		_, err := lister.ByNamespace(wObjIdentifier.ObjectName.Namespace).Get(wObjIdentifier.ObjectName.Name)
		if err != nil {
			if errors.IsNotFound(err) {
				c.singletonState.delete(wObjIdentifier)
			}
			return err
		}
	}

	allBdgs, err := c.wdsKsClient.ControlV1alpha1().Bindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, bdg := range allBdgs.Items {
		logger.V(2).Info("Got Binding while building singleton state", "name", bdg.Name)
		for _, item := range bdg.Spec.Workload.NamespaceScope {
			nsWObjRef := item.NamespaceScopeDownsyncObject
			lister, found := c.listers.Get(schema.GroupVersionResource(nsWObjRef.GroupVersionResource))
			if !found {
				return fmt.Errorf("could not get lister for gvr: %s", nsWObjRef.GroupVersionResource)
			}
			nsWObj, err := lister.ByNamespace(nsWObjRef.Namespace).Get(nsWObjRef.Name)
			if err != nil {
				return err
			}
			labels := nsWObj.(metav1.Object).GetLabels()
			if v, ok := labels[util.BindingPolicyLabelSingletonStatusKey]; ok {
				wObjIdentifier := util.ObjectIdentifier{
					GVK:        schema.GroupVersionKind{Group: nsWObjRef.Group, Version: nsWObjRef.Version, Kind: nsWObj.GetObjectKind().GroupVersionKind().Kind},
					Resource:   nsWObjRef.Resource,
					ObjectName: cache.NewObjectName(nsWObjRef.Namespace, nsWObjRef.Name),
				}
				if v == util.BindingPolicyLabelSingletonStatusValueSet && len(bdg.Spec.Destinations) > 0 {
					logger.V(2).Info("Singleton workload object should have status synced", "resource", wObjIdentifier.Resource, "objectName", wObjIdentifier.ObjectName, "binding", bdg.Name)
					c.singletonState.set(wObjIdentifier)
				} else {
					logger.V(2).Info("Singleton workload object status syncing is not driven by Binding", "resource", wObjIdentifier.Resource, "objectName", wObjIdentifier.ObjectName, "binding", bdg.Name)
					c.singletonState.addIfNotExist(wObjIdentifier)
				}
			}
		}
		for _, item := range bdg.Spec.Workload.ClusterScope {
			cWObjRef := item.ClusterScopeDownsyncObject
			lister, found := c.listers.Get(schema.GroupVersionResource(cWObjRef.GroupVersionResource))
			if !found {
				return fmt.Errorf("could not get lister for gvr: %s", cWObjRef.GroupVersionResource)
			}
			clusterWObj, err := lister.Get(cWObjRef.Name)
			if err != nil {
				return err
			}
			labels := clusterWObj.(metav1.Object).GetLabels()
			if v, ok := labels[util.BindingPolicyLabelSingletonStatusKey]; ok {
				wObjIdentifier := util.ObjectIdentifier{
					GVK:        schema.GroupVersionKind{Group: cWObjRef.Group, Version: cWObjRef.Version, Kind: clusterWObj.GetObjectKind().GroupVersionKind().Kind},
					Resource:   cWObjRef.Resource,
					ObjectName: cache.NewObjectName("", cWObjRef.Name),
				}
				if v == util.BindingPolicyLabelSingletonStatusValueSet && len(bdg.Spec.Destinations) > 0 {
					logger.V(2).Info("Singleton workload object should have status synced", "resource", wObjIdentifier.Resource, "objectName", wObjIdentifier.ObjectName, "binding", bdg.Name)
					c.singletonState.set(wObjIdentifier)
				} else {
					logger.V(2).Info("Singleton workload object status syncing is not driven by Binding", "resource", wObjIdentifier.Resource, "objectName", wObjIdentifier.ObjectName, "binding", bdg.Name)
					c.singletonState.addIfNotExist(wObjIdentifier)
				}
			}
		}
	}

	if !reconcile {
		return nil
	}

	for wObjIdentifier := range c.singletonState.wObjSync {
		if err := c.reconcileSingletonByWObj(ctx, wObjIdentifier); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) reconcileSingletonByWObj(ctx context.Context, wObjID util.ObjectIdentifier) error {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Reconciling singleton workload object", "resource", wObjID.Resource, "objectName", wObjID.ObjectName)

	if !c.singletonState.get(wObjID) {
		emptyStatus := make(map[string]interface{})
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
		logger.V(2).Info("Updating singleton status", "workstatus", wsRef.Name, "wecName", wsRef.WECName)
		return updateObjectStatus(ctx, &wsRef.SourceObjectIdentifier, status, c.listers, c.wdsDynClient)
	}
	return nil
}

func (c *Controller) reconcileSingletonByBdg(ctx context.Context, ref bindingRef) error {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Reconciling singleton status due to binding changes", "name", string(ref))
	return c.buildSingletonStateAndOptionallyReconcile(ctx, true)
}

func (c *Controller) reconcileSingletonByWs(ctx context.Context, ref singletonWorkStatusRef) error {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Reconciling singleton status due to workstatus changes", "name", string(ref.Name))
	obj, _ := c.workStatusLister.ByNamespace(ref.WECName).Get(ref.Name)
	status, _ := util.GetWorkStatusStatus(obj)
	if sync, ok := c.singletonState.wObjSync[ref.SourceObjectIdentifier]; !ok {
		logger.V(2).Info("Not a singleton workload object", "objectIdentifier", ref.SourceObjectIdentifier)
		return nil
	} else if !sync {
		logger.V(2).Info("Singleton workload object should not have status synced, not updating status",
			"resource", ref.SourceObjectIdentifier.Resource, "objectName", ref.SourceObjectIdentifier.ObjectName)
		return nil
	} else {
		logger.V(2).Info("Singleton workload object should have status synced, updating status",
			"resource", ref.SourceObjectIdentifier.Resource, "objectName", ref.SourceObjectIdentifier.ObjectName)
		return updateObjectStatus(ctx, &ref.SourceObjectIdentifier, status, c.listers, c.wdsDynClient)
	}
}
