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

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// syncBinding syncs a binding object with what is resolved by the bindingpolicy resolver.
func (c *Controller) syncBinding(ctx context.Context, objIdentifier util.ObjectIdentifier) error {
	logger := klog.FromContext(ctx)

	var unstructuredObj *unstructured.Unstructured
	if !c.bindingPolicyResolver.ResolutionExists(objIdentifier.ObjectName.Name) {
		// if a resolution is not associated to the binding's name
		// then the bindingpolicy has been deleted, and the binding
		// will eventually be garbage collected. We can safely ignore this.

		return nil
	}

	obj, err := c.getObjectFromIdentifier(objIdentifier)
	if errors.IsNotFound(err) {
		// a resolution exists and the object is not found, therefore it is deleted and should be created
		unstructuredObj = util.EmptyUnstructuredObjectFromIdentifier(objIdentifier)
	} else if err != nil {
		return fmt.Errorf("failed to get runtime.Object from object identifier (%v): %w", objIdentifier, err)
	} else {
		// perform the type assertion only if getObjectFromIdentifier did not fail
		var ok bool
		unstructuredObj, ok = obj.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("the given runtime.Object (%#v) is not a pointer to Unstructured", obj)
		}
	}

	binding, err := unstructuredObjectToBinding(unstructuredObj)
	if err != nil {
		return fmt.Errorf("failed to convert from Unstructured to Binding: %w", err)
	}

	// binding name matches that of the bindingpolicy 1:1, therefore its NamespacedName is the same.
	bindingPolicyIdentifier := binding.GetName()

	// generate binding spec from resolver
	generatedBindingSpec := c.bindingPolicyResolver.GenerateBinding(bindingPolicyIdentifier)
	if generatedBindingSpec == nil { // resolution does not exist, abort syncing
		return fmt.Errorf("syncing Binding was stopped because it has no counterpart resolution")
	}

	// calculate if the resolved decision is different from the current one
	if !c.bindingPolicyResolver.CompareBinding(bindingPolicyIdentifier, &binding.Spec) {
		// update the binding object in the cluster by updating spec
		if err = c.updateOrCreateBinding(ctx, binding, generatedBindingSpec); err != nil {
			return fmt.Errorf("failed to update or create binding: %w", err)
		}

		return nil
	}

	logger.Info("binding is up to date", "name", binding.GetName())
	return nil
}

// updateOrCreateBinding updates or creates a binding object in the cluster.
// If the object already exists, it is updated. Otherwise, it is created.
func (c *Controller) updateOrCreateBinding(ctx context.Context, bdg *v1alpha1.Binding,
	generatedBindingSpec *v1alpha1.BindingSpec) error {
	// use the passed binding and set its spec
	bdg.Spec = *generatedBindingSpec

	// set owner reference
	ownerReference, err := c.bindingPolicyResolver.GetOwnerReference(bdg.GetName())
	if err != nil {
		return fmt.Errorf("failed to get OwnerReference: %w", err)
	}
	bdg.SetOwnerReferences([]metav1.OwnerReference{ownerReference})

	// update or create binding
	unstructuredBinding, err := bindingToUnstructuredObject(bdg)
	if err != nil {
		return fmt.Errorf("failed to convert Binding to Unstructured: %w", err)
	}

	logger := klog.FromContext(ctx)

	_, err = c.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Version:  bdg.GetObjectKind().GroupVersionKind().Version,
		Resource: util.BindingResource,
	}).Update(ctx, unstructuredBinding, metav1.UpdateOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.dynamicClient.Resource(schema.GroupVersionResource{
				Group:    v1alpha1.SchemeGroupVersion.Group,
				Version:  bdg.GetObjectKind().GroupVersionKind().Version,
				Resource: util.BindingResource,
			}).Create(ctx, unstructuredBinding, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create binding: %w", err)
			}

			logger.Info("created binding", "name", bdg.GetName())
			return nil
		} else {
			return fmt.Errorf("failed to update binding: %w", err)
		}
	}

	logger.Info("updated binding", "name", bdg.GetName())
	return nil
}

func unstructuredObjectToBinding(unstructuredObj *unstructured.Unstructured) (*v1alpha1.Binding, error) {
	var binding *v1alpha1.Binding
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.UnstructuredContent(),
		&binding); err != nil {
		return nil, fmt.Errorf("failed to convert Unstructured to Binding: %w", err)
	}

	return binding, nil
}

func bindingToUnstructuredObject(binding *v1alpha1.Binding) (*unstructured.Unstructured, error) {
	innerObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(binding)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Binding to map[string]interface{}: %w", err)
	}

	return &unstructured.Unstructured{
		Object: innerObj,
	}, nil
}
