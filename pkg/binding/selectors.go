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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/pkg/ocm"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// matches an object to each bindingpolicy and returns the list of matching clusters (if any)
// the list of bindingpolicies that manage the object and if singleton status required
// (the latter forces the selection  of only one cluster)
// TODO (maroon): this should be deleted when transport is ready
func (c *Controller) matchSelectors(obj runtime.Object) (sets.Set[string], []string, bool, error) {
	managedByBindingPolicyList := []string{}
	objMR := obj.(mrObject)
	bindingpolicies, err := c.listBindingPolicies()
	if err != nil {
		return nil, nil, false, err
	}

	clusterSet := sets.New[string]()
	// if a bindingpolicy wants single reported status we force to select only one cluster
	wantSingletonStatus := false
	for _, item := range bindingpolicies {
		bindingpolicy, err := runtimeObjectToBindingPolicy(item)
		if err != nil {
			return nil, nil, false, err
		}
		matchedSome := c.testObject(objMR, bindingpolicy.Spec.Downsync)
		if !matchedSome {
			continue
		}
		// WantSingletonReportedState for multiple bindingpolicies are OR'd
		if bindingpolicy.Spec.WantSingletonReportedState {
			wantSingletonStatus = true
		}

		managedByBindingPolicyList = append(managedByBindingPolicyList, bindingpolicy.GetName())
		c.logger.Info("Matched", "object", util.RefToRuntimeObj(obj), "for bindingpolicy", bindingpolicy.GetName())

		clusters, err := ocm.FindClustersBySelectors(c.ocmClient, bindingpolicy.Spec.ClusterSelectors)
		if err != nil {
			return nil, nil, false, err
		}

		if clusters == nil {
			continue
		}

		clusterSet = clusterSet.Union(clusters)
	}

	return clusterSet, managedByBindingPolicyList, wantSingletonStatus, nil
}

// when an object is updated, we iterate over all bindingpolices and update bindingpolicy resolutions that
// are affected by the update. Every affected binding is then queued to be synced.
func (c *Controller) updateDecisions(obj runtime.Object) error {
	bindingpolices, err := c.listBindingPolicies()
	if err != nil {
		return err
	}

	objMR := obj.(mrObject)

	for _, item := range bindingpolices {
		bindingpolicy, err := runtimeObjectToBindingPolicy(item)
		if err != nil {
			return err
		}

		matchedSome := c.testObject(objMR, bindingpolicy.Spec.Downsync)
		if !matchedSome {
			// if previously selected, remove
			// TODO: optimize
			if resolutionUpdated := c.bindingPolicyResolver.RemoveObject(bindingpolicy.GetName(), obj); resolutionUpdated {
				// enqueue binding to be synced since object was removed from its bindingpolicy's resolution
				c.logger.V(4).Info("enqueued Binding for syncing due to the removal of an "+
					"object from its resolution", "binding", bindingpolicy.GetName(),
					"object", util.RefToRuntimeObj(obj))
				c.enqueueBinding(bindingpolicy.GetName())
			}
			continue
		}

		// obj is selected by bindingpolicy, update the bindingpolicy resolver
		resolutionUpdated, err := c.bindingPolicyResolver.NoteObject(bindingpolicy.GetName(), obj)
		if err != nil {
			if errorIsBindingPolicyResolutionNotFound(err) {
				// this case can occur if a bindingpolicy resolution was deleted AFTER
				// starting this iteration and BEFORE getting to the NoteObject function,
				// which occurs if a bindingpolicy was deleted in this time-window.
				utilruntime.HandleError(fmt.Errorf("failed to note object (%s) - %w",
					util.RefToRuntimeObj(obj), err))
				continue
			}

			return fmt.Errorf("failed to update resolution for bindingpolicy %s for object %v: %v",
				bindingpolicy.GetName(), util.RefToRuntimeObj(obj), err)
		}

		if resolutionUpdated {
			// enqueue binding to be synced since an object was added to its bindingpolicy's resolution
			c.logger.V(4).Info("enqueued Binding for syncing due to a noting of an "+
				"object in its resolution", "binding", bindingpolicy.GetName(),
				"object", util.RefToRuntimeObj(obj))
			c.enqueueBinding(bindingpolicy.GetName())
		}
	}

	return nil
}
