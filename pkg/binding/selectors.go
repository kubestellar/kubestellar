/*
Copyright 2023 The KubeStellar Authors.

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

package binding

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

const emptyBindingPolicyName = ""

// when an object is updated, we iterate over all bindingpolicies and update
// resolutions that are affected by the update. Every changed resolution leads
// to queueing its relevant binding for syncing.
func (c *Controller) updateResolutions(ctx context.Context, objIdentifier util.ObjectIdentifier) error {
	bindingPolicyRuntimeObjects, err := c.listBindingPolicies()
	if err != nil {
		return err
	}

	logger := klog.FromContext(ctx)

	obj, err := c.getObjectFromIdentifier(objIdentifier)
	if errors.IsNotFound(err) {
		// object is deleted, delete from all resolutions it exists in (and enqueue Binding references for the latter)
		return c.removeObjectFromBindingPolicies(objIdentifier, bindingPolicyRuntimeObjects)
	} else if err != nil {
		return fmt.Errorf("failed to get runtime.Object from object identifier (%v): %w", objIdentifier, err)
	}

	objMR := obj.(mrObject)
	objBeingDeleted := isBeingDeleted(obj)
	objGVR := schema.GroupVersionResource{Group: objIdentifier.GVK.Group, Version: objIdentifier.GVK.Version,
		Resource: objIdentifier.Resource,
	}

	isSelectedBySingletonBinding := false

	for _, item := range bindingPolicyRuntimeObjects {
		bindingPolicy, err := runtimeObjectToBindingPolicy(item)
		if err != nil {
			return fmt.Errorf("failed to convert runtime.Object to BindingPolicy: %w", err)
		}

		matchedSome := c.testObject(ctx, objIdentifier, objMR.GetLabels(), bindingPolicy.Spec.Downsync)
		if !matchedSome {
			// if previously selected, remove
			if resolutionUpdated := c.bindingPolicyResolver.RemoveObjectIdentifier(bindingPolicy.GetName(),
				objIdentifier); resolutionUpdated {
				// enqueue binding to be synced since object was removed from its bindingpolicy's resolution
				logger.V(4).Info("Enqueuing Binding for syncing due to the removal of an "+
					"object from its resolution", "binding", bindingPolicy.GetName(),
					"objectIdentifier", objIdentifier)
				c.enqueueBinding(bindingPolicy.GetName())
			} else {
				logger.V(4).Info("Not enqueuing Binding for syncing due to the removal of an "+
					"object from its resolution", "binding", bindingPolicy.GetName(),
					"objectIdentifier", objIdentifier)
			}
			continue
		}

		// obj is selected by bindingpolicy, update the bindingpolicy resolver
		resolutionUpdated, err := c.bindingPolicyResolver.EnsureObjectIdentifierWithVersion(bindingPolicy.GetName(),
			objIdentifier, objMR.GetResourceVersion())
		if err != nil {
			if errorIsBindingPolicyResolutionNotFound(err) {
				// this case can occur if a bindingpolicy resolution was deleted AFTER
				// starting this iteration and BEFORE getting to the NoteObject function,
				// which occurs if a bindingpolicy was deleted in this time-window.
				utilruntime.HandleError(fmt.Errorf("failed to note object (identifier: %v) - %w",
					objIdentifier, err))
				continue
			}

			return fmt.Errorf("failed to update resolution for bindingpolicy %s for object (identifier: %v): %v",
				bindingPolicy.GetName(), objIdentifier, err)
		}

		if resolutionUpdated {
			// enqueue binding to be synced since an object was added to its bindingpolicy's resolution
			logger.V(4).Info("Enqueued Binding for syncing due to a noting of an "+
				"object in its resolution", "binding", bindingPolicy.GetName(),
				"objectIdentifier", objIdentifier, "objBeingDeleted", objBeingDeleted)
			c.enqueueBinding(bindingPolicy.GetName())
		}

		// make sure object has singleton status if needed
		if !objBeingDeleted &&
			c.bindingPolicyResolver.ResolutionRequiresSingletonReportedState(bindingPolicy.GetName()) {
			if err := c.handleSingletonLabel(ctx, obj, objGVR, bindingPolicy.GetName()); err != nil {
				return fmt.Errorf("failed to add singleton label to object: %w", err)
			}

			isSelectedBySingletonBinding = true
		}
	}

	// if the binding-policies matching cycles end and the object is not selected by a singleton binding,
	// we need to remove the singleton label from the object if it exists.
	// NOTE that this takes care of the case where the object was previously selected by a singleton binding
	// and is no longer selected by any binding.
	if !objBeingDeleted && !isSelectedBySingletonBinding {
		if err := c.handleSingletonLabel(ctx, obj, objGVR, emptyBindingPolicyName); err != nil {
			return fmt.Errorf("failed to remove singleton label from object: %w", err)
		}
	}

	return nil
}

// handleSingletonLabel adds or removes the singleton label from the object in the cluster.
// If a bindingPolicyName is provided, the singleton label is added to the object. If the bindingPolicyName is empty,
// the singleton label is removed.
//
// The method parameter `obj` is not mutated by this function.
func (c *Controller) handleSingletonLabel(ctx context.Context, obj runtime.Object, objGVR schema.GroupVersionResource,
	bindingPolicyName string) error {
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("failed to convert runtime.Object to unstructured.Unstructured")
	}

	labels := unstructuredObj.GetLabels() // gets a copy of the labels
	if labels == nil {
		labels = make(map[string]string)
	}

	_, found := labels[util.BindingPolicyLabelSingletonStatusKey]

	if bindingPolicyName == emptyBindingPolicyName {
		if !found {
			return nil
		}
		delete(labels, util.BindingPolicyLabelSingletonStatusKey)
	} else {
		if found {
			return nil
		}
		labels[util.BindingPolicyLabelSingletonStatusKey] = bindingPolicyName
	}

	unstructuredObj = unstructuredObj.DeepCopy() // avoid mutating the original object
	unstructuredObj.SetLabels(labels)

	if unstructuredObj.GetNamespace() == metav1.NamespaceNone {
		_, err := c.dynamicClient.Resource(objGVR).Update(ctx, unstructuredObj, metav1.UpdateOptions{})
		if errors.IsNotFound(err) {
			return nil // object was deleted after getting into this function. This is not an error.
		}
		return err
	}

	_, err := c.dynamicClient.Resource(objGVR).Namespace(unstructuredObj.GetNamespace()).Update(ctx, unstructuredObj,
		metav1.UpdateOptions{})
	if errors.IsNotFound(err) {
		return nil // object was deleted after getting into this function. This is not an error.
	}

	return err
}

func (c *Controller) removeObjectFromBindingPolicies(objIdentifier util.ObjectIdentifier,
	bindingPolicyRuntimeObjects []runtime.Object) error {
	for _, bindingPolicyRuntimeObject := range bindingPolicyRuntimeObjects {
		bindingPolicy, err := runtimeObjectToBindingPolicy(bindingPolicyRuntimeObject)
		if err != nil {
			return fmt.Errorf("failed to convert runtime.Object to BindingPolicy: %w", err)
		}

		if resolutionUpdated := c.bindingPolicyResolver.RemoveObjectIdentifier(bindingPolicy.GetName(),
			objIdentifier); resolutionUpdated {
			// enqueue binding to be synced since object was removed from its bindingpolicy's resolution
			c.enqueueBinding(bindingPolicy.GetName())
		}
	}

	return nil
}
