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
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

const workStatusIdentificationIndexKey = "workStatusIdentificationIndex"

type workStatus struct {
	workStatusRef
	status map[string]interface{}
}

func (c *Controller) syncWorkStatus(ctx context.Context, ref workStatusRef) error {
	logger := klog.FromContext(ctx)

	workStatus := &workStatus{
		workStatusRef: ref, //readonly
		status:        nil,
	}

	obj, err := c.workStatusLister.ByNamespace(ref.wecName).Get(ref.name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get workstatus (%v): %w", ref, err)
		} // if not found, the above workstatus will reflect the fact
	} else {
		status, err := util.GetWorkStatusStatus(obj)
		if err != nil {
			logger.Error(err, "Failed to get status from workstatus", "workStatusRef", ref)
		}

		workStatus.status = status // might be nil
	}

	combinedStatusSet := c.combinedStatusResolver.NoteWorkStatus(workStatus) // nil .status is equivalent to deleted
	for combinedStatus := range combinedStatusSet {
		c.workqueue.AddAfter(combinedStatusRef(combinedStatus.ObjectName.AsNamespacedName().String()), queueingDelay)
	}

	return nil
}

func (c *Controller) syncSingletonWorkStatus(ctx context.Context, ref singletonWorkStatusRef) error {
	logger := klog.FromContext(ctx)

	workStatusObj, err := c.workStatusLister.ByNamespace(ref.wecName).Get(ref.name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get workstatus (%v): %w", ref, err)
		}
	}

	if workStatusObj == nil { // make sure status of source obj is deleted
		emptyStatus := make(map[string]interface{})
		if err = updateObjectStatus(ctx, &ref.sourceObjectIdentifier, emptyStatus,
			c.listers, c.wdsDynClient); err != nil {
			return err
		}

		return c.syncWorkStatus(ctx, workStatusRef(ref))
	}

	// only process workstatues with the label for single reported status
	statusLabelVal, ok := workStatusObj.(metav1.Object).GetLabels()[util.BindingPolicyLabelSingletonStatusKey]
	if !ok {
		return nil
	}

	// remove the status if singleton status label value is unset
	if statusLabelVal == util.BindingPolicyLabelSingletonStatusValueUnset {
		emptyStatus := make(map[string]interface{})
		return updateObjectStatus(ctx, &ref.sourceObjectIdentifier, emptyStatus,
			c.listers, c.wdsDynClient)
	}

	status, err := util.GetWorkStatusStatus(workStatusObj)
	if err != nil {
		return err
	}

	if status == nil {
		// when status is updated, a new event will trigger update
		return nil
	}

	logger.Info("Updating singleton status", "objectIdentifier", ref.sourceObjectIdentifier)
	if err = updateObjectStatus(ctx, &ref.sourceObjectIdentifier, status, c.listers, c.wdsDynClient); err != nil {
		return err
	}

	return c.syncWorkStatus(ctx, workStatusRef(ref))
}

func updateObjectStatus(ctx context.Context, objectIdentifier *util.ObjectIdentifier, status map[string]interface{},
	listers util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister], wdsDynClient dynamic.Interface) error {
	logger := klog.FromContext(ctx)

	gvr := objectIdentifier.GVR()
	lister, found := listers.Get(gvr)
	if !found {
		logger.V(5).Info("Could not find lister for gvr", "gvr", gvr)
		return nil
	}

	obj, err := lister.ByNamespace(objectIdentifier.ObjectName.Namespace).Get(objectIdentifier.ObjectName.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(5).Info("Could not find object", "objectIdentifier", objectIdentifier)
			return nil
		}

		return fmt.Errorf("failed to get object (%v): %w", objectIdentifier, err)
	}

	unstrObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("object cannot be cast to *unstructured.Unstructured: object: %s", util.RefToRuntimeObj(obj))
	}

	// set the status and update the object
	unstrObj.Object["status"] = status

	if objectIdentifier.ObjectName.Namespace == "" {
		_, err = wdsDynClient.Resource(gvr).UpdateStatus(ctx, unstrObj, metav1.UpdateOptions{})
	} else {
		_, err = wdsDynClient.Resource(gvr).Namespace(objectIdentifier.ObjectName.Namespace).UpdateStatus(ctx,
			unstrObj, metav1.UpdateOptions{})
	}
	if err != nil {
		// if resource not found it may mean no status subresource - try to patch the status
		if errors.IsNotFound(err) {
			return util.PatchStatus(ctx, unstrObj, status, objectIdentifier.ObjectName.Namespace, gvr, wdsDynClient)
			// PatchStatus returns nil if the full object is not found
		}
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func runtimeObjectToWorkStatus(obj runtime.Object) (*workStatus, error) {
	ref, err := runtimeObjectToWorkStatusRef(obj)
	if err != nil {
		return nil, err
	}

	status, err := util.GetWorkStatusStatus(obj)
	if err != nil {
		return nil, err
	}

	return &workStatus{
		workStatusRef: *ref,
		status:        status,
	}, nil
}

func runtimeObjectToWorkStatusRef(obj runtime.Object) (*workStatusRef, error) {
	name := obj.(metav1.Object).GetName()
	wecName := obj.(metav1.Object).GetNamespace()

	sourceRef, err := util.GetWorkStatusSourceRef(obj)
	if err != nil {
		return nil, err
	}

	objIdentifier := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: sourceRef.Group, Version: sourceRef.Version, Kind: sourceRef.Kind},
		Resource:   sourceRef.Resource,
		ObjectName: cache.NewObjectName(sourceRef.Namespace, sourceRef.Name),
	}

	return &workStatusRef{
		name:                   name,
		wecName:                wecName,
		sourceObjectIdentifier: objIdentifier,
	}, nil
}
