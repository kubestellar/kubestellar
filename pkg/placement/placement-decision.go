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
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubestellar/kubestellar/api/edge/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// updateTimestampAnnotationKey is the annotation key used to mark the last
// update timestamp of a placement-decision. This is used to trigger watchers
// on the placement-decision resource.
const updateTimestampAnnotationKey = "transport.kubestellar.io/lastUpdateTimestamp"

// handlePlacementDecision syncs a placement-decision object with what is resolved by the placement-decision-resolver.
func (c *Controller) handlePlacementDecision(obj runtime.Object) error {
	placementDecision, err := runtimeObjectToPlacementDecision(obj)
	if err != nil {
		return fmt.Errorf("failed to handle placement-decision: %v", err)
	}

	// placement decision name matches that of the placement 1:1, therefore its NamespacedName is the same.
	placementDecisionIdentifier := namespacedNameFromObjectMeta(placementDecision.ObjectMeta)

	// get placement decision spec from resolver
	placementDecisionSpec, err := c.placementDecisionResolver.GetPlacementDecision(placementDecisionIdentifier)
	if err != nil {
		return fmt.Errorf("failed to get placement decision spec: %v", err)
	}

	// calculate if the resolved decision is different from the current one
	if !c.placementDecisionResolver.ComparePlacementDecision(placementDecisionIdentifier, placementDecisionSpec) {
		// update the placement decision object in the cluster by updating spec
		if err = c.updateOrCreatePlacementDecision(placementDecision, placementDecisionSpec); err != nil {
			return fmt.Errorf("failed to update or create placement decision: %v", err)
		}
	}

	return nil
}

// updateOrCreatePlacementDecision updates or creates a placement-decision object in the cluster.
// If the object already exists, it is updated. Otherwise, it is created.
func (c *Controller) updateOrCreatePlacementDecision(pd *v1alpha1.PlacementDecision,
	placementDecisionSpec *v1alpha1.PlacementDecisionSpec) error {
	// add finalizer to placement decision if missing
	addFinalizerToPlacementDecision(pd)
	// add update-timestamp annotation in order to trigger any watcher to receive an event on the placement-decision
	// resource. since the objects/destinations list may not have changed, but an object-internal change
	// could have triggered the queuing of affected placement-decisions.
	updatePlacementDecisionTimestampAnnotation(pd)

	unstructuredPlacementDecision, err := placementDecisionSpecToUnstructuredObject(pd, placementDecisionSpec)
	if err != nil {
		return fmt.Errorf("failed to update or create placement decision: %v", err)
	}

	_, err = c.dynamicClient.Resource(schema.GroupVersionResource{
		Group:    v1alpha1.SchemeGroupVersion.Group,
		Version:  pd.GetObjectKind().GroupVersionKind().Version,
		Resource: util.PlacementDecisionResource,
	}).Update(c.ctx, unstructuredPlacementDecision, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update placement decision: %v", err)
	}

	if err != nil {
		if errors.IsNotFound(err) {
			_, err = c.dynamicClient.Resource(schema.GroupVersionResource{
				Group:    v1alpha1.SchemeGroupVersion.Group,
				Version:  pd.GetObjectKind().GroupVersionKind().Version,
				Resource: util.PlacementDecisionResource,
			}).Create(c.ctx, unstructuredPlacementDecision, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create placement decision: %v", err)
			}
		} else {
			return fmt.Errorf("failed to update placement decision: %v", err)
		}
	}

	return nil
}

func (c *Controller) listPlacementDecisions() ([]runtime.Object, error) {
	pdLister := c.listers[util.KeyForGroupVersionKind(v1alpha1.SchemeGroupVersion.Group,
		v1alpha1.SchemeGroupVersion.Version, util.PlacementDecisionKind)]
	if pdLister == nil {
		return nil, fmt.Errorf("could not get lister for placememt-decision")
	}
	lister := *pdLister

	list, err := lister.List(labels.Nothing()) // dont need labels
	if err != nil {
		return nil, err
	}
	return list, nil
}

func runtimeObjectToPlacementDecision(obj runtime.Object) (*v1alpha1.PlacementDecision, error) {
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to convert runtime.Object to unstructured.Unstructured")
	}

	var placementDecision *v1alpha1.PlacementDecision
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.UnstructuredContent(),
		&placementDecision); err != nil {
		return nil, fmt.Errorf("failed to convert unstructured.Unstructured to PlacementDecision: %v", err)
	}

	return placementDecision, nil
}

func placementDecisionSpecToUnstructuredObject(placementDecision *v1alpha1.PlacementDecision,
	spec *v1alpha1.PlacementDecisionSpec) (*unstructured.Unstructured, error) {
	// replace spec
	placementDecision.Spec = *spec

	innerObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(placementDecision)
	if err != nil {
		return nil, fmt.Errorf("failed to convert PlacementDecisionSpec to unstructured: %v", err)
	}

	return &unstructured.Unstructured{
		Object: innerObj,
	}, nil
}

func addFinalizerToPlacementDecision(pd *v1alpha1.PlacementDecision) {
	pd.ObjectMeta.SetFinalizers(append(pd.ObjectMeta.GetFinalizers(), KSFinalizer))
}

func updatePlacementDecisionTimestampAnnotation(pd *v1alpha1.PlacementDecision) {
	pd.ObjectMeta.Annotations[updateTimestampAnnotationKey] = metav1.Now().String()
}
