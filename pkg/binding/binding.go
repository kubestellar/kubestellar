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
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

// syncBinding syncs a binding object with what is resolved by the bindingpolicy resolver.
func (c *Controller) syncBinding(ctx context.Context, bindingName string) error {
	logger := klog.FromContext(ctx)

	if !c.bindingPolicyResolver.ResolutionExists(bindingName) {
		// if a resolution is not associated to the binding's name
		// then the bindingpolicy has been deleted, and the binding
		// will eventually be garbage collected. We can safely ignore this.

		c.bindingPolicyResolver.Broker().NotifyCallbacks(bindingName)
		return nil
	}

	binding, err := c.bindingLister.Get(bindingName)
	// `*binding` is immutable
	if errors.IsNotFound(err) {
		// a resolution exists and the object is not found, therefore it is deleted and should be created
		binding = &v1alpha1.Binding{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Binding",
				APIVersion: v1alpha1.GroupVersion.String()},
			ObjectMeta: metav1.ObjectMeta{
				Name: bindingName},
		}
	} else if err != nil {
		return fmt.Errorf("failed to get Binding from informer cache (name=%v): %w", bindingName, err)
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

		// notify the bindingpolicy resolution broker that the binding has been updated
		c.bindingPolicyResolver.Broker().NotifyCallbacks(bindingPolicyIdentifier)
		return nil
	}

	logger.Info("binding is up to date", "name", binding.GetName())
	return nil
}

// updateOrCreateBinding updates or creates a binding object in the cluster.
// If the object already exists, it is updated. Otherwise, it is created.
// The given `bdg *v1alpha1.Binding` points to immutable storage.
func (c *Controller) updateOrCreateBinding(ctx context.Context, bdg *v1alpha1.Binding,
	generatedBindingSpec *v1alpha1.BindingSpec) error {
	bdg = bdg.DeepCopy()
	// use the passed binding and set its spec
	bdg.Spec = *generatedBindingSpec

	// set owner reference
	ownerReference, err := c.bindingPolicyResolver.GetOwnerReference(bdg.GetName())
	if err != nil {
		return fmt.Errorf("failed to get OwnerReference: %w", err)
	}
	bdg.SetOwnerReferences([]metav1.OwnerReference{ownerReference})

	logger := klog.FromContext(ctx)
	bdgEcho, err := c.controlClient.Bindings().Update(ctx, bdg, metav1.UpdateOptions{FieldManager: ControllerName})

	if err != nil {
		if errors.IsNotFound(err) {
			bdgEcho, err = c.controlClient.Bindings().Create(ctx, bdg, metav1.CreateOptions{FieldManager: ControllerName})
			if err != nil {
				return fmt.Errorf("failed to create binding (name=%s): %w", bdg.Name, err)
			}

			logger.Info("created binding", "name", bdg.GetName(), "resourceVersion", bdgEcho.ResourceVersion)
			return nil
		} else {
			return fmt.Errorf("failed to update binding: %w", err)
		}
	}

	logger.Info("updated binding", "name", bdg.GetName(), "resourceVersion", bdgEcho.ResourceVersion)
	return nil
}
