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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubestellar/kubestellar/api/edge/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/ocm"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// matches an object to each placement and returns the list of matching clusters (if any)
// the list of placements that manage the object and if singleton status required
// (the latter forces the selection  of only one cluster)
// TODO (maroon): this should be deleted when transport is ready
func (c *Controller) matchSelectors(obj runtime.Object) ([]string, []string, bool, error) {
	managedByPlacementList := []string{}
	objMR := obj.(mrObject)
	placements, err := c.listPlacements()
	if err != nil {
		return nil, nil, false, err
	}
	clustersMap := map[string]string{}
	// if a placement wants single reported status we force to select only one cluster
	wantSingletonStatus := false
	for _, item := range placements {
		placement, err := runtimeObjectToPlacement(item)
		if err != nil {
			return nil, nil, false, err
		}
		matchedSome := c.testObject(objMR, placement.Spec.Downsync)
		if !matchedSome {
			continue
		}
		// WantSingletonReportedState for multiple placement are OR'd
		if placement.Spec.WantSingletonReportedState {
			wantSingletonStatus = true
		}
		managedByPlacementList = append(managedByPlacementList, placement.GetName())
		c.logger.Info("Matched", "object", util.GenerateObjectInfoString(obj), "for placement", placement.GetName())
		list, err := ocm.ListClustersBySelectors(c.ocmClient, placement.Spec.ClusterSelectors)
		if err != nil {
			return nil, nil, false, err
		}
		if list == nil {
			continue
		}
		for _, s := range list {
			clustersMap[s] = ""
		}
	}
	return GetKeys(clustersMap), managedByPlacementList, wantSingletonStatus, nil
}

func GetKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// when an object is updated, we iterate over all placements and update decisions that
// are affected by the update. Every affected placement decision is then queued to by synced.
// TODO: when PlacementDecision API supports singleton statuses, this function should be updated.
func (c *Controller) updateDecisions(obj runtime.Object) error {
	placements, err := c.listPlacements()
	if err != nil {
		return err
	}

	objMR := obj.(mrObject)

	changedDecisions := map[types.NamespacedName]interface{}{}

	for _, item := range placements {
		placement, err := runtimeObjectToPlacement(item)
		if err != nil {
			return err
		}

		matchedSome := c.testObject(objMR, placement.Spec.Downsync)
		if !matchedSome {
			continue
		}

		// obj is selected by placement, update the placement decision resolver
		resolutionUpdated, err := c.placementDecisionResolver.UpdateDecisionDataResources(namespacedNameFromObjectMeta(placement.ObjectMeta),
			obj)
		if err != nil {
			return fmt.Errorf("failed to update decisions for object %v: %v",
				util.GenerateObjectInfoString(obj), err)
		}

		if resolutionUpdated {
			changedDecisions[namespacedNameFromObjectMeta(placement.ObjectMeta)] = nil
		}
	}

	// if no decision changed, we can return
	if len(changedDecisions) == 0 {
		return nil
	}

	// get all placement decisions
	placementDecisions, err := c.listPlacementDecisions()
	if err != nil {
		return fmt.Errorf("failed to update decisions: %v", err)
	}

	handledDecisions := map[types.NamespacedName]interface{}{}

	// queue all placement decisions that exist are affected by the update
	for _, pdObj := range placementDecisions {
		id := types.NamespacedName{
			Namespace: pdObj.(mrObject).GetNamespace(),
			Name:      pdObj.(mrObject).GetName(),
		}

		if _, found := changedDecisions[id]; found {
			// enqueue the placement decision for syncing
			c.enqueueObject(pdObj, true)
			// mark as handled
			handledDecisions[id] = nil
		}
	}

	// create all missing placement decisions by enqueuing them as new objects
	// at a later stage in the pipeline a worker will realize that the object does not exist
	// and will create it
	for namespacedName, _ := range changedDecisions {
		if _, found := handledDecisions[namespacedName]; found {
			continue
		}

		// enqueue the placement decision for creation
		c.enqueueObject(&v1alpha1.PlacementDecision{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespacedName.Namespace,
				Name:      namespacedName.Name,
			},
		}, false)
	}

	return nil
}
