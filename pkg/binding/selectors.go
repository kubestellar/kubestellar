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
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

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
		logger.V(3).Info("Removing non-existent object from resolutions", "object", objIdentifier, "numPolicies", len(bindingPolicyRuntimeObjects))
		// object is deleted, delete from all resolutions it exists in (and enqueue Binding references for the latter)
		return c.removeObjectFromBindingPolicies(ctx, objIdentifier, bindingPolicyRuntimeObjects)
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

		if !c.bindingPolicyResolver.ResolutionExists(bindingPolicy.GetName()) {
			continue // resolution does not exist, skip
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
				// the BindingPolicyResolver::ResolutionExists call and BEFORE getting to the NoteObject function,
				// which occurs if a bindingpolicy was deleted in this time-window.
				logger.V(4).Info("skipped EnsureObjectIdentifierWithVersion for object because "+
					"bindingpolicy was deleted", "objectIdentifier", objIdentifier,
					"bindingpolicy", bindingPolicy.GetName())
				continue
			}

			return fmt.Errorf("failed to update resolution for bindingpolicy %s for object (identifier: %v): %v",
				bindingPolicy.GetName(), objIdentifier, err)
		}

		if resolutionUpdated {
			// enqueue binding to be synced since an object was added to its bindingpolicy's resolution
			logger.V(4).Info("Enqueued Binding for syncing due to a noting of an "+
				"object in its resolution", "binding", bindingPolicy.GetName(),
				"objectIdentifier", objIdentifier, "objBeingDeleted", objBeingDeleted,
				"resourceVersion", objMR.GetResourceVersion())
			c.enqueueBinding(bindingPolicy.GetName())
		} else {
			logger.V(5).Info("Not enqueuing Binding due to no change in resolution",
				"binding", bindingPolicy.GetName(),
				"objectIdentifier", objIdentifier, "objBeingDeleted", objBeingDeleted,
				"resourceVersion", objMR.GetResourceVersion())

		}

		// make sure object has singleton status if needed
		if !objBeingDeleted && c.bindingPolicyResolver.ResolutionRequiresSingletonReportedState(bindingPolicy.GetName()) {
			if err := c.handleSingletonLabel(ctx, obj, objGVR, util.BindingPolicyLabelSingletonStatusValueSet); err != nil {
				return fmt.Errorf("failed to add singleton label to object: %w", err)
			}

			isSelectedBySingletonBinding = true
		}
	}

	// if the binding-policies matching cycles end and the object is not selected by a singleton binding,
	// we need to update the singleton label value for the object.
	// NOTE that this takes care of the case where the object was previously selected by a singleton binding
	// and is no longer selected by any binding.
	if !objBeingDeleted && !isSelectedBySingletonBinding {
		if err := c.handleSingletonLabel(ctx, obj, objGVR, util.BindingPolicyLabelSingletonStatusValueUnset); err != nil {
			return fmt.Errorf("failed to update singleton label for object: %w", err)
		}
	}

	return nil
}

// handleSingletonLabel adds the singleton label to the object in the cluster,
// or matches the label value to the expectedLabelValue if needed.
//
// The method parameter `obj` is not mutated by this function.
func (c *Controller) handleSingletonLabel(ctx context.Context, obj runtime.Object, objGVR schema.GroupVersionResource,
	expectedLabelValue string) error {
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("failed to convert runtime.Object to unstructured.Unstructured")
	}

	labels := unstructuredObj.GetLabels() // gets a copy of the labels
	if labels == nil {
		labels = make(map[string]string)
	}

	val, found := labels[util.BindingPolicyLabelSingletonStatusKey]

	if !found && expectedLabelValue == util.BindingPolicyLabelSingletonStatusValueUnset {
		return nil
	}
	if found && val == expectedLabelValue {
		return nil
	}
	labels[util.BindingPolicyLabelSingletonStatusKey] = expectedLabelValue

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

func (c *Controller) removeObjectFromBindingPolicies(ctx context.Context, objIdentifier util.ObjectIdentifier,
	bindingPolicyRuntimeObjects []runtime.Object) error {
	logger := klog.FromContext(ctx)
	for _, bindingPolicyRuntimeObject := range bindingPolicyRuntimeObjects {
		bindingPolicy, err := runtimeObjectToBindingPolicy(bindingPolicyRuntimeObject)
		if err != nil {
			return fmt.Errorf("failed to convert runtime.Object to BindingPolicy: %w", err)
		}

		if resolutionUpdated := c.bindingPolicyResolver.RemoveObjectIdentifier(bindingPolicy.GetName(),
			objIdentifier); resolutionUpdated {
			// enqueue binding to be synced since object was removed from its bindingpolicy's resolution
			logger.V(4).Info("Enqueuing Binding due to deletion of matching object", "bindingPolicy", bindingPolicy.Name, "object", objIdentifier)
			c.enqueueBinding(bindingPolicy.GetName())
		} else {
			logger.V(5).Info("Not enqueuing Binding due to deletion of non-matching object", "bindingPolicy", bindingPolicy.Name, "object", objIdentifier)
		}
	}

	return nil
}
