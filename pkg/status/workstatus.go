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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/pkg/util"
)

type workStatus struct {
	wecName                string
	sourceObjectIdentifier util.ObjectIdentifier
	status                 map[string]interface{}
}

func convertToWorkStatusesList(objList []runtime.Object) ([]*workStatus, error) {
	var workStatuses []*workStatus
	for _, ws := range objList {
		wStatus, err := convertToWorkStatus(ws)
		if err != nil {
			return workStatuses, err
		}
		workStatuses = append(workStatuses, wStatus)
	}

	return workStatuses, nil
}

func convertToWorkStatus(obj runtime.Object) (*workStatus, error) {
	// wecName is the WorkStatus namespace
	wecName := obj.(metav1.Object).GetNamespace()

	status, err := util.GetWorkStatusStatus(obj)
	if err != nil {
		return nil, err
	}
	sourceRef, err := util.GetWorkStatusSourceRef(obj)
	if err != nil {
		return nil, err
	}

	objIdentifier := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: sourceRef.Group, Version: sourceRef.Version, Kind: sourceRef.Kind},
		Resource:   sourceRef.Resource,
		ObjectName: cache.NewObjectName(sourceRef.Namespace, sourceRef.Name),
	}

	return &workStatus{
		wecName:                wecName,
		sourceObjectIdentifier: objIdentifier,
		status:                 status}, nil
}

func (c *Controller) syncWorkStatus(ctx context.Context, key string) error {
	ns, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}
	obj, err := getObject(c.workStatusLister, ns, name)
	if err != nil {
		// The resource no longer exist, which means it has been deleted.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("object %#v in work queue no longer exists", key))
			return nil
		}
		return err
	}

	ws, err := convertToWorkStatus(obj)
	if err != nil {
		return err
	}
	combinedStatusSet := c.combinedStatusResolver.NoteWorkStatus(ws)
	for combinedStatus := range combinedStatusSet {
		// TODO: combinedStatusResolver.NoteWorkStatus can return a list of strings(ns/name)
		c.workqueue.AddAfter(combinedStatusRef(combinedStatus.ObjectName.AsNamespacedName().String()), queueingDelay)
	}

	// only process workstatues with the label for single reported status
	statusLabelVal, ok := obj.(metav1.Object).GetLabels()[util.BindingPolicyLabelSingletonStatusKey]
	if !ok {
		return nil
	}

	sourceRef, err := util.GetWorkStatusSourceRef(obj)
	if err != nil {
		return err
	}

	// remove the status if singleton status label value is unset
	if statusLabelVal == util.BindingPolicyLabelSingletonStatusValueUnset {
		emptyStatus := make(map[string]interface{})
		return updateObjectStatus(ctx, sourceRef, emptyStatus, c.listers, c.wdsDynClient)
	}

	status, err := util.GetWorkStatusStatus(obj)
	if err != nil {
		// status gets updated after workstatus is created, it's ok to requeue
		return err
	}

	c.logger.Info("updating singleton status", "kind", sourceRef.Kind, "name", sourceRef.Name, "namespace", sourceRef.Namespace)
	return updateObjectStatus(ctx, sourceRef, status, c.listers, c.wdsDynClient)
}

func updateObjectStatus(ctx context.Context, objRef *util.SourceRef, status map[string]interface{},
	listers util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister], wdsDynClient dynamic.Interface) error {

	gvr := schema.GroupVersionResource{Group: objRef.Group, Version: objRef.Version, Resource: objRef.Resource}
	lister, found := listers.Get(gvr)
	if !found {
		return fmt.Errorf("could not find lister for gvr %s", gvr)
	}

	obj, err := getObject(lister, objRef.Namespace, objRef.Name)
	if err != nil {
		return err
	}

	unstrObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("object cannot be cast to *unstructured.Unstructured: object: %s", util.RefToRuntimeObj(obj))
	}

	// set the status and update the object
	unstrObj.Object["status"] = status

	if objRef.Namespace == "" {
		_, err = wdsDynClient.Resource(gvr).UpdateStatus(ctx, unstrObj, metav1.UpdateOptions{})
	} else {
		_, err = wdsDynClient.Resource(gvr).Namespace(objRef.Namespace).UpdateStatus(ctx, unstrObj, metav1.UpdateOptions{})
	}
	if err != nil {
		// if resource not found it may mean no status subresource - try to patch the status
		if errors.IsNotFound(err) {
			return util.PatchStatus(ctx, unstrObj, status, objRef.Namespace, gvr, wdsDynClient)
		}
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// TODO - move these to a common lib
func getObject(lister cache.GenericLister, namespace, name string) (runtime.Object, error) {
	if namespace != "" {
		return lister.ByNamespace(namespace).Get(name)
	}
	return lister.Get(name)
}
