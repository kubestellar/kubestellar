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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// when an object is updated, we iterate over all bindingpolicies and update
// resolutions that are affected by the update. Every changed resolution leads
// to queueing its relevant binding for syncing.
func (c *Controller) updateResolutions(ctx context.Context, objIdentifier util.ObjectIdentifier) error {
	bindingPolicies, err := c.listBindingPolicies()
	if err != nil {
		return err
	}

	logger := klog.FromContext(ctx)

	obj, err := c.getObjectFromIdentifier(objIdentifier)
	if errors.IsNotFound(err) {
		logger.V(3).Info("Removing non-existent object from resolutions", "object", objIdentifier, "numPolicies", len(bindingPolicies))
		// object is deleted, delete from all resolutions it exists in (and enqueue Binding references for the latter)
		return c.removeObjectFromBindingPolicies(ctx, objIdentifier, bindingPolicies)
	} else if err != nil {
		return fmt.Errorf("failed to get runtime.Object from object identifier (%v): %w", objIdentifier, err)
	}

	objMR := obj.(mrObject)
	objBeingDeleted := isBeingDeleted(obj)

	for _, bindingPolicy := range bindingPolicies {
		if !c.bindingPolicyResolver.ResolutionExists(bindingPolicy.GetName()) {
			continue // resolution does not exist, skip
		}

		matchedAny, modFromPolicy := c.testObject(ctx, bindingPolicy.GetName(), objIdentifier, objMR.GetLabels(), bindingPolicy.Spec.Downsync)
		if !matchedAny {
			// if previously selected, remove
			if resolutionUpdated := c.bindingPolicyResolver.RemoveObjectIdentifier(bindingPolicy.GetName(),
				objIdentifier); resolutionUpdated {
				// enqueue binding to be synced since object was removed from its bindingpolicy's resolution
				logger.V(4).Info("Enqueuing Binding for syncing due to the removal of an "+
					"object from its resolution", "binding", bindingPolicy.GetName(),
					"objectIdentifier", objIdentifier)
				c.enqueueBinding(bindingPolicy.GetName())
			} else {
				logger.V(5).Info("Not enqueuing Binding for syncing, because its resolution continues "+
					"to not include workload object", "binding", bindingPolicy.GetName(),
					"objectIdentifier", objIdentifier)
			}
			continue
		}
		logger.V(5).Info("BindingPolicy matched workload object", "policy", bindingPolicy.Name, "objIdentifier", objIdentifier)

		// obj is selected by bindingpolicy, update the bindingpolicy resolver
		resolutionUpdated, err := c.bindingPolicyResolver.EnsureObjectData(bindingPolicy.GetName(),
			objIdentifier, string(objMR.GetUID()), objMR.GetResourceVersion(), modFromPolicy)
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
			logger.V(5).Info("Enqueued Binding for syncing due to a noting of an "+
				"object in its resolution", "binding", bindingPolicy.GetName(),
				"objectIdentifier", objIdentifier, "objBeingDeleted", objBeingDeleted,
				"resourceVersion", objMR.GetResourceVersion())
			c.enqueueBinding(bindingPolicy.GetName())
		} else {
			logger.V(5).Info("Not enqueuing Binding, due to no change in resolution",
				"binding", bindingPolicy.GetName(),
				"objectIdentifier", objIdentifier, "objBeingDeleted", objBeingDeleted,
				"resourceVersion", objMR.GetResourceVersion())

		}
	}

	return nil
}

func (c *Controller) removeObjectFromBindingPolicies(ctx context.Context, objIdentifier util.ObjectIdentifier,
	bindingPolicies []*v1alpha1.BindingPolicy) error {
	logger := klog.FromContext(ctx)
	for _, bindingPolicy := range bindingPolicies {
		if resolutionUpdated := c.bindingPolicyResolver.RemoveObjectIdentifier(bindingPolicy.GetName(),
			objIdentifier); resolutionUpdated {
			// enqueue binding to be synced since object was removed from its bindingpolicy's resolution
			logger.V(5).Info("Enqueuing Binding due to deletion of matching object", "bindingPolicy", bindingPolicy.Name, "object", objIdentifier)
			c.enqueueBinding(bindingPolicy.GetName())
		} else {
			logger.V(5).Info("Not enqueuing Binding due to deletion of non-matching object", "bindingPolicy", bindingPolicy.Name, "object", objIdentifier)
		}
	}

	return nil
}

// testObject tests if the object matches the given tests.
// The returned tuple is:
//   - bool: whether the object matches ANY of the tests
//   - bool: whether any test that matches the object also says CreateOnly==true
//   - sets.Set[string]: the UNION of the statuscollector names that appear within
//     EACH of the tests that the object matches
func (c *Controller) testObject(ctx context.Context, bindingName string, objIdentifier util.ObjectIdentifier, objLabels map[string]string,
	tests []v1alpha1.DownsyncPolicyClause) (bool, DownsyncModulation) {

	logger := klog.FromContext(ctx)

	var matched bool
	mod := ZeroDownsyncModulation()

	var objNS *corev1.Namespace
	for _, test := range tests {
		if test.APIGroup != nil && (*test.APIGroup) != objIdentifier.GVK.Group {
			continue
		}
		if len(test.Resources) > 0 && !(SliceContains(test.Resources, "*") ||
			SliceContains(test.Resources, objIdentifier.Resource)) {
			continue
		}
		if len(test.Namespaces) > 0 && !(SliceContains(test.Namespaces, "*") ||
			SliceContains(test.Namespaces, objIdentifier.ObjectName.Namespace)) {
			continue
		}
		if len(test.ObjectNames) > 0 && !(SliceContains(test.ObjectNames, "*") ||
			SliceContains(test.ObjectNames, objIdentifier.ObjectName.Name)) {
			continue
		}
		if len(test.ObjectSelectors) > 0 && !labelsMatchAny(c.logger, objLabels, test.ObjectSelectors) {
			continue
		}
		if len(test.NamespaceSelectors) > 0 && !ALabelSelectorIsEmpty(test.NamespaceSelectors...) {
			if objNS == nil {
				var err error
				objNS, err = c.namespaceClient.Get(ctx,
					objIdentifier.ObjectName.Namespace, metav1.GetOptions{})
				if err != nil {
					logger.V(3).Info("Object namespace not found, assuming object does not match",
						"object identifier", objIdentifier, "binding", bindingName)
					continue
				}
			}
			if !labelsMatchAny(logger, objNS.Labels, test.NamespaceSelectors) {
				continue
			}
		}

		klog.FromContext(ctx).V(5).Info("Workload object matched clause", "objIdentifier", objIdentifier, "objLabels", objLabels, "clause", test, "binding", bindingName)
		// test is a match
		matched = true
		mod.AddExternal(test.DownsyncModulation)
	}

	return matched, mod
}
