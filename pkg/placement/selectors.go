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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubestellar/kubestellar/pkg/ocm"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// matches an object to each placement and returns the list of matching clusters (if any)
// the list of placements that manage the object and if singleton status required
// (the latter forces the selection  of only one cluster)
func (c *Controller) matchSelectors(obj runtime.Object) ([]string, []string, bool, error) {
	managedByPlacementList := []string{}
	mObj := obj.(metav1.Object)
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
		matchAllSelectors := false
		for _, downsync := range placement.Spec.Downsync {
			for _, s := range downsync.LabelSelectors {
				selector, err := metav1.LabelSelectorAsSelector(&s)
				if err != nil {
					return nil, nil, false, err
				}
				if selector.Matches(labels.Set(mObj.GetLabels())) {
					matchAllSelectors = true
				} else {
					matchAllSelectors = false
					break
				}
			}
		}
		if !matchAllSelectors {
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
