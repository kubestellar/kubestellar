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

package placement

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// syncPlacementDecision syncs a placement-decision object with what is resolved by the placement-resolver.
func (c *Controller) syncPlacementDecision(key util.Key) error {
	var unstructuredObj *unstructured.Unstructured
	if !c.placementResolver.ResolutionExists(key.NamespacedName.Name) {
		// if a resolution is not associated to the placement-decision's name
		// then the placement has been deleted, and the placement-decision
		// will eventually be garbage collected. We can safely ignore this.

		return nil
	}

	obj, err := c.getObjectFromKey(key)
	if errors.IsNotFound(err) {
		unstructuredObj = util.EmptyUnstructuredObjectFromKey(key)
	} else if err != nil {
		return fmt.Errorf("failed to get runtime.Object from key with gvk (%v) and namespaced-name (%v): %w",
			key.GVK, key.NamespacedName, err)
	} else {
		// perform the type assertion only if getObjectFromKey did not fail
		var ok bool
		unstructuredObj, ok = obj.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("the given runtime.Object (%#v) is not a pointer to Unstructured", obj)
		}
	}

	placementDecision, err := unstructuredObjectToPlacementDecision(unstructuredObj)
	if err != nil {
		return fmt.Errorf("failed to convert from Unstructured to PlacementDecision: %w", err)
	}

	// placement decision name matches that of the placement 1:1, therefore its NamespacedName is the same.
	placementIdentifier := placementDecision.GetName()

	// generate placement decision spec from resolver
	generatedPlacementDecisionSpec, err := c.placementResolver.GeneratePlacementDecision(placementIdentifier)
	if err != nil {
		return fmt.Errorf("failed to generate PlacementDecisionSpec: %w", err)
	}

	// calculate if the resolved decision is different from the current one
	if !c.placementResolver.ComparePlacementDecision(placementIdentifier, &placementDecision.Spec) {
		// update the placement decision object in the cluster by updating spec
		if err = c.updateOrCreatePlacementDecision(placementDecision, generatedPlacementDecisionSpec); err != nil {
			return fmt.Errorf("failed to update or create placement decision: %w", err)
		}

		return nil
	}

	c.logger.Info("placement decision is up to date", "name", placementDecision.GetName())
	return nil
}

// updateOrCreatePlacementDecision updates or creates a placement-decision object in the cluster.
// If the object already exists, it is updated. Otherwise, it is created.
func (c *Controller) updateOrCreatePlacementDecision(pd *v1alpha1.PlacementDecision,
	generatedPlacementDecisionSpec *v1alpha1.PlacementDecisionSpec) error {
	// use the passed placement decision and set its spec
	pd.Spec = *generatedPlacementDecisionSpec

	// set owner reference
	ownerReference, err := c.placementResolver.GetOwnerReference(pd.GetName())
	if err != nil {
		return fmt.Errorf("failed to get OwnerReference: %w", err)
	}
	pd.SetOwnerReferences([]metav1.OwnerReference{ownerReference})

	// update or create placement decision
	unstructuredPlacementDecision, err := placementDecisionToUnstructuredObject(pd)
	if err != nil {
		return fmt.Errorf("failed to convert PlacementDecision to Unstructured: %w", err)
	}

	_, err = c.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Version:  pd.GetObjectKind().GroupVersionKind().Version,
		Resource: util.PlacementDecisionResource,
	}).Update(context.Background(), unstructuredPlacementDecision, metav1.UpdateOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.dynamicClient.Resource(schema.GroupVersionResource{
				Group:    v1alpha1.SchemeGroupVersion.Group,
				Version:  pd.GetObjectKind().GroupVersionKind().Version,
				Resource: util.PlacementDecisionResource,
			}).Create(context.Background(), unstructuredPlacementDecision, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create placement decision: %w", err)
			}

			c.logger.Info("created placement decision", "name", pd.GetName())
			return nil
		} else {
			return fmt.Errorf("failed to update placement decision: %w", err)
		}
	}

	c.logger.Info("updated placement decision", "name", pd.GetName())
	return nil
}

func unstructuredObjectToPlacementDecision(unstructuredObj *unstructured.Unstructured) (*v1alpha1.PlacementDecision, error) {
	var placementDecision *v1alpha1.PlacementDecision
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.UnstructuredContent(),
		&placementDecision); err != nil {
		return nil, fmt.Errorf("failed to convert Unstructured to PlacementDecision: %w", err)
	}

	return placementDecision, nil
}

func placementDecisionToUnstructuredObject(placementDecision *v1alpha1.PlacementDecision) (*unstructured.Unstructured, error) {
	innerObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(placementDecision)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PlacementDecision to map[string]interface{}: %w", err)
	}

	return &unstructured.Unstructured{
		Object: innerObj,
	}, nil
}
