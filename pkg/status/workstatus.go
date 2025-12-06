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
	"encoding/json"
	"fmt"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

const workStatusIdentificationIndexKey = "workStatusIdentificationIndex"

type workStatus struct {
	workStatusRef
	status         map[string]interface{}
	lastUpdateTime *metav1.Time
}

func (ws *workStatus) Content() map[string]interface{} {
	return map[string]interface{}{
		"status": ws.status,
	}
}

func (c *Controller) syncWorkStatus(ctx context.Context, ref workStatusRef) error {
	logger := klog.FromContext(ctx)
	wsON := ref.ObjectName()

	if err := c.updateWorkStatusToObject(ctx, wsON); err != nil {
		return err
	}

	workStatus := &workStatus{
		workStatusRef: ref, //readonly
		status:        nil,
	}

	obj, err := c.workStatusLister.ByNamespace(ref.WECName).Get(ref.Name)
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
		workStatus.lastUpdateTime = getObjectStatusLastUpdateTime(obj.(metav1.Object))
	}

	combinedStatusSet := c.combinedStatusResolver.NoteWorkStatus(ctx, workStatus) // nil .status is equivalent to deleted
	for combinedStatus := range combinedStatusSet {
		logger.V(5).Info("Enqueuing reference to CombinedStatus while syncing WorkStatus", "combinedStatusRef", combinedStatus.ObjectName, "workStatusRef", ref)
		c.workqueue.AddAfter(combinedStatusRef(combinedStatus.ObjectName.AsNamespacedName().String()), queueingDelay)
	}

	return nil
}

// updateObjectStatus puts returned `.status` into the given workload object
// and updates the label indicating that this has been done, or
// updates both to indicate that singleton status return has not been done.
// `status == nil` indicates that singleton status return is not desired.
func (c *Controller) updateObjectStatus(ctx context.Context, objectIdentifier util.ObjectIdentifier, status map[string]interface{},
	listers util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister], isMultiWEC bool) error {
	logger := klog.FromContext(ctx)

	if util.WEC2WDSExceptions.Has(objectIdentifier.GVK.GroupKind()) {
		logger.V(4).Info("Status from WEC shouldn't have authority to overwrite status in WDS", "objectIdentifier", objectIdentifier)
		return nil
	}

	gvr := objectIdentifier.GVR()
	lister, found := listers.Get(gvr)
	if !found {
		logger.V(4).Info("Could not find lister for gvr", "gvr", gvr)
		return nil
	}

	var obj runtime.Object
	var err error
	var listerIfc cache.GenericNamespaceLister = lister
	if objectIdentifier.ObjectName.Namespace != "" {
		listerIfc = lister.ByNamespace(objectIdentifier.ObjectName.Namespace)
	}
	obj, err = listerIfc.Get(objectIdentifier.ObjectName.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(4).Info("Did not update workload object because it is not in local cache, presumably because it was recently deleted", "objectIdentifier", objectIdentifier)
			return nil
		}
		return fmt.Errorf("failed to get object (%v): %w", objectIdentifier, err)
	}

	unstrObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("object cannot be cast to *unstructured.Unstructured: object: %s", util.RefToRuntimeObj(obj))
	}
	wantReturn := status != nil
	labels := unstrObj.GetLabels()
	_, haveSingleton := labels[util.BindingPolicyLabelSingletonStatusKey]
	_, haveMultiWEC := labels[util.BindingPolicyLabelMultiWECStatusKey]

	// check which label to add
	wantSingleton := wantReturn && !isMultiWEC
	wantMultiWEC := wantReturn && isMultiWEC

	if !wantReturn && !haveMultiWEC && !haveSingleton {
		logger.V(5).Info("Workload object neither wants nor has returned status", "objectIdentifier", objectIdentifier)
		return nil
	}

	if wantSingleton && !haveSingleton {
		err = c.handleStatusReturnLabel(ctx, unstrObj, objectIdentifier.GVR(), true, util.BindingPolicyLabelSingletonStatusKey)
		if err != nil {
			return err
		}
	}
	if wantMultiWEC && !haveMultiWEC {
		err = c.handleStatusReturnLabel(ctx, unstrObj, objectIdentifier.GVR(), true, util.BindingPolicyLabelMultiWECStatusKey)
		if err != nil {
			return err
		}
	}
	if status == nil {
		status = map[string]any{}
	}
	if apiequality.Semantic.DeepEqual(unstrObj.Object["status"], status) {
		logger.V(5).Info("Workload object found to already have intended status", "objectIdentifier", objectIdentifier)
	} else {
		// set the status and update the object
		unstrObj.Object["status"] = status

		rscIfc := util.DynamicForResource(c.wdsDynClient, gvr, unstrObj.GetNamespace())
		_, err = rscIfc.UpdateStatus(ctx, unstrObj, metav1.UpdateOptions{FieldManager: ControllerName})
		if err != nil {
			// if resource not found it may mean no status subresource - try to patch the status
			if errors.IsNotFound(err) {
				return util.PatchStatus(ctx, unstrObj, status, objectIdentifier.ObjectName.Namespace, gvr, c.wdsDynClient)
				// PatchStatus returns nil if the full object is not found
			}
			return fmt.Errorf("failed to update status of %v: %w", objectIdentifier, err)
		}
		logger.V(5).Info("Updated status of workload object", "objectIdentifier", objectIdentifier)
	}

	if haveSingleton && !wantReturn {
		err = c.handleStatusReturnLabel(ctx, unstrObj, objectIdentifier.GVR(), false, util.BindingPolicyLabelSingletonStatusKey)
		if err != nil {
			return err
		}
	}

	if haveMultiWEC && !wantReturn {
		err = c.handleStatusReturnLabel(ctx, unstrObj, objectIdentifier.GVR(), false, util.BindingPolicyLabelMultiWECStatusKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handleStatusReturnLabel(ctx context.Context, unstructuredObj *unstructured.Unstructured,
	objGVR schema.GroupVersionResource, wantLabel bool, labelKey string) error {

	labels := unstructuredObj.GetLabels() // gets a copy of the labels
	_, foundLabel := labels[labelKey]
	if foundLabel == wantLabel {
		return nil
	}
	message := fmt.Sprintf("Added %s label to workload object", labelKey)
	if wantLabel {
		if labels == nil {
			labels = map[string]string{}
		}
		labels[labelKey] = "true"
	} else {
		message = fmt.Sprintf("Removed %s label from workload object", labelKey)
		delete(labels, labelKey)
	}
	unstructuredObj = unstructuredObj.DeepCopy() // avoid mutating the original object
	unstructuredObj.SetLabels(labels)
	namespace := unstructuredObj.GetNamespace()
	rscIfc := util.DynamicForResource(c.wdsDynClient, objGVR, namespace)
	_, err := rscIfc.Update(ctx, unstructuredObj, metav1.UpdateOptions{FieldManager: ControllerName})
	if errors.IsNotFound(err) {
		return nil // object was deleted after getting into this function. This is not an error.
	}
	if err == nil {
		klog.FromContext(ctx).V(5).Info(message, "gvr", objGVR, "namespace", namespace, "name", unstructuredObj.GetName())
	}
	return err
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
		workStatusRef:  *ref,
		status:         status,
		lastUpdateTime: getObjectStatusLastUpdateTime(obj.(metav1.Object)),
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
		Name:                   name,
		WECName:                wecName,
		SourceObjectIdentifier: objIdentifier,
	}, nil
}

func getObjectStatusLastUpdateTime(metaObj metav1.Object) *metav1.Time {
	// iterate through all managedFields entries to find the one that updated the status
	latestTime := &metav1.Time{}

	for _, field := range metaObj.GetManagedFields() {
		if field.FieldsType == "FieldsV1" &&
			field.FieldsV1 != nil {
			// parse the FieldsV1.Raw to a map to check if "f:status" is present
			var fieldsMap map[string]interface{}
			if err := json.Unmarshal(field.FieldsV1.Raw, &fieldsMap); err != nil {
				continue
			}

			if _, ok := fieldsMap["f:status"]; ok {
				if field.Time.After(latestTime.Time) {
					latestTime = field.Time
				}
			}
		}
	}

	return latestTime
}
