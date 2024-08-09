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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// reconcileSingletonByBdg goes over the binding-covered workload objects, decides sync or not for each of the workload objects,
// then maintains their statuses based on the decisions.
func (c *Controller) reconcileSingletonByBdg(ctx context.Context, bdgName string) error {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Reconciling singleton status due to binding change", "binding", bdgName)

	binding, err := c.wdsKsClient.ControlV1alpha1().Bindings().Get(ctx, bdgName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("could not get binding %s from cache", bdgName)
	}

	sync := make(map[util.ObjectIdentifier]bool)
	if err = c.lookForSingletonWorkloadInBdg(ctx, *binding, sync, true); err != nil {
		return err
	}

	allBdgs, err := c.wdsKsClient.ControlV1alpha1().Bindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, bdg := range allBdgs.Items {
		logger.V(2).Info("Got Binding while building singleton state", "name", bdg.Name)
		if bdg.Name == binding.Name {
			continue
		}
		if err = c.lookForSingletonWorkloadInBdg(ctx, bdg, sync, false); err != nil {
			return err
		}
	}

	for wObjIdentifier, v := range sync {
		if err := c.reconcileSingletonWObj(ctx, wObjIdentifier, v); err != nil {
			return err
		}
	}
	return nil
}

func (c *Controller) reconcileSingletonWObj(ctx context.Context, wObjID util.ObjectIdentifier, sync bool) error {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Reconciling singleton workload object", "resource", wObjID.Resource, "objectName", wObjID.ObjectName)

	if !sync {
		emptyStatus := make(map[string]interface{})
		logger.V(2).Info("Cleaning up singleton status", "resource", wObjID.Resource, "objectName", wObjID.ObjectName)
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

// reconcileSingletonByWs decides sync or not for the corresponding workload object,
// then maintains its status based on the decision.
func (c *Controller) reconcileSingletonByWs(ctx context.Context, ref singletonWorkStatusRef) error {
	logger := klog.FromContext(ctx)
	logger.V(2).Info("Reconciling singleton status due to workstatus changes", "name", string(ref.Name))

	wObjID := ref.SourceObjectIdentifier
	wObjGVR := wObjID.GVR()
	lister, found := c.listers.Get(wObjGVR)
	if !found {
		return fmt.Errorf("could not get lister for gvr: %s", wObjGVR)
	}

	var wObj runtime.Object
	var err error
	if wObjID.ObjectName.Namespace != "" {
		wObj, err = lister.ByNamespace(wObjID.ObjectName.Namespace).Get(wObjID.ObjectName.Name)
	} else {
		wObj, err = lister.Get(wObjID.ObjectName.Name)
	}
	if err != nil {
		return err
	}

	labels := wObj.(metav1.Object).GetLabels()
	if v, ok := labels[util.BindingPolicyLabelSingletonStatusKey]; !ok {
		return nil
	} else if v == util.BindingPolicyLabelSingletonStatusValueUnset {
		return c.reconcileSingletonWObj(ctx, wObjID, false)
	}

	sync := map[util.ObjectIdentifier]bool{wObjID: false}
	allBdgs, err := c.wdsKsClient.ControlV1alpha1().Bindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, bdg := range allBdgs.Items {
		logger.V(2).Info("Got Binding while building singleton state", "name", bdg.Name)
		if err = c.lookForSingletonWorkloadInBdg(ctx, bdg, sync, false); err != nil {
			return err
		}
	}

	if !sync[wObjID] {
		logger.V(2).Info("Singleton workload object should not have status synced, cleaning status",
			"resource", ref.SourceObjectIdentifier.Resource, "objectName", ref.SourceObjectIdentifier.ObjectName)
		return c.reconcileSingletonWObj(ctx, wObjID, false)
	} else {
		logger.V(2).Info("Singleton workload object should have status synced, updating status",
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
		return updateObjectStatus(ctx, &wObjID, status, c.listers, c.wdsDynClient)
	}
}

func (c *Controller) lookForSingletonWorkloadInBdg(ctx context.Context, binding v1alpha1.Binding, sync map[util.ObjectIdentifier]bool, init bool) error {
	logger := klog.FromContext(ctx)
	for _, item := range binding.Spec.Workload.NamespaceScope {
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
			if init {
				sync[wObjIdentifier] = false // hold a place
			} else {
				if _, ok = sync[wObjIdentifier]; !ok {
					continue
				}
			}
			if v == util.BindingPolicyLabelSingletonStatusValueSet && binding.GetDeletionTimestamp() == nil && len(binding.Spec.Destinations) > 0 {
				logger.V(2).Info("Singleton workload object should have status synced", "resource", wObjIdentifier.Resource, "objectName", wObjIdentifier.ObjectName, "binding", binding.Name)
				sync[wObjIdentifier] = true
			}
		}
	}
	for _, item := range binding.Spec.Workload.ClusterScope {
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
			if init {
				sync[wObjIdentifier] = false // hold a place
			} else {
				if _, ok = sync[wObjIdentifier]; !ok {
					continue
				}
			}
			if v == util.BindingPolicyLabelSingletonStatusValueSet && binding.GetDeletionTimestamp() == nil && len(binding.Spec.Destinations) > 0 {
				logger.V(2).Info("Singleton workload object should have status synced", "resource", wObjIdentifier.Resource, "objectName", wObjIdentifier.ObjectName, "binding", binding.Name)
				sync[wObjIdentifier] = true
			}
		}
	}
	return nil
}
