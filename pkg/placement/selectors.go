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

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/pkg/ocm"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// matches an object to each placement and returns the list of matching clusters (if any)
// the list of placements that manage the object and if singleton status required
// (the latter forces the selection  of only one cluster)
// TODO (maroon): this should be deleted when transport is ready
func (c *Controller) matchSelectors(obj runtime.Object) (sets.Set[string], []string, bool, error) {
	managedByPlacementList := []string{}
	objMR := obj.(mrObject)
	placements, err := c.listPlacements()
	if err != nil {
		return nil, nil, false, err
	}

	clusterSet := sets.New[string]()
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

		clusters, err := ocm.FindClustersBySelectors(c.ocmClient, placement.Spec.ClusterSelectors)
		if err != nil {
			return nil, nil, false, err
		}

		if clusters == nil {
			continue
		}

		clusterSet = clusterSet.Union(clusters)
	}

	return clusterSet, managedByPlacementList, wantSingletonStatus, nil
}

// when an object is updated, we iterate over all placements and update placement-resolutions that
// are affected by the update. Every affected placement decision is then queued to be synced.
func (c *Controller) updateDecisions(obj runtime.Object) error {
	placements, err := c.listPlacements()
	if err != nil {
		return err
	}

	objMR := obj.(mrObject)

	for _, item := range placements {
		placement, err := runtimeObjectToPlacement(item)
		if err != nil {
			return err
		}

		matchedSome := c.testObject(objMR, placement.Spec.Downsync)
		if !matchedSome {
			// if previously selected, remove
			// TODO: optimize
			if resolutionUpdated := c.placementResolver.RemoveObject(placement.GetName(), obj); resolutionUpdated {
				// enqueue placement-decision to be synced since object was removed from its placement's resolution
				c.logger.V(4).Info("enqueued PlacementDecision for syncing due to the removal of an "+
					"object from its resolution", "placement-decision", placement.GetName(),
					"object", util.GenerateObjectInfoString(obj))
				c.enqueuePlacementDecision(placement.GetName())
			}
			continue
		}

		// obj is selected by placement, update the placement decision resolver
		resolutionUpdated, err := c.placementResolver.NoteObject(placement.GetName(), obj)
		if err != nil {
			if errorIsPlacementResolutionNotFound(err) {
				// this case can occur if a placement resolution was deleted AFTER
				// starting this iteration and BEFORE getting to the NoteObject function,
				// which occurs if a placement was deleted in this time-window.
				utilruntime.HandleError(fmt.Errorf("failed to note object (%s) - %w",
					util.GenerateObjectInfoString(obj), err))
				continue
			}

			return fmt.Errorf("failed to update resolution for placement %s for object %v: %v",
				placement.GetName(), util.GenerateObjectInfoString(obj), err)
		}

		if resolutionUpdated {
			// enqueue placement-decision to be synced since an object was added to its placement's resolution
			c.logger.V(4).Info("enqueued PlacementDecision for syncing due to a noting of an "+
				"object in its resolution", "placement-decision", placement.GetName(),
				"object", util.GenerateObjectInfoString(obj))
			c.enqueuePlacementDecision(placement.GetName())
		}
	}

	return nil
}
