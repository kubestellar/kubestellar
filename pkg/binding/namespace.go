/*
Copyright 2026 The KubeStellar Authors.

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
	"maps"

	"github.com/go-logr/logr"

	k8scoreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

var namespaceGVR = k8scoreapi.SchemeGroupVersion.WithResource("namespaces")

// handleNamespaceUpdate requeues the workload objects in a namespace when a
// change to the namespace's labels changes whether some BindingPolicy's
// `namespaceSelectors` match it. Without this, objects in the namespace would
// only be re-evaluated when they themselves change.
func (c *Controller) handleNamespaceUpdate(ctx context.Context, oldObj, newObj any) {
	logger := klog.FromContext(ctx)
	oldNS := oldObj.(metav1.Object)
	newNS := newObj.(metav1.Object)
	oldLabels := oldNS.GetLabels()
	newLabels := newNS.GetLabels()
	if maps.Equal(oldLabels, newLabels) {
		return
	}

	bindingPolicies, err := c.listBindingPolicies()
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	if !namespaceSelectionChanged(logger, bindingPolicies, oldLabels, newLabels) {
		logger.V(5).Info("Namespace labels changed without affecting any BindingPolicy", "namespace", newNS.GetName(),
			"oldLabels", oldLabels, "newLabels", newLabels)
		return
	}

	logger.V(4).Info("Requeuing workload objects in namespace due to change in namespaceSelectors match",
		"namespace", newNS.GetName(), "oldLabels", oldLabels, "newLabels", newLabels)
	if err := c.requeueNamespaceObjects(ctx, newNS.GetName()); err != nil {
		utilruntime.HandleError(err)
	}
}

// namespaceSelectionChanged returns whether some downsync clause of some of the given
// BindingPolicies has `namespaceSelectors` whose match differs between the two label sets.
// Clauses without namespaceSelectors, or with an empty selector, match regardless of
// namespace labels and are therefore ignored.
func namespaceSelectionChanged(logger logr.Logger, bindingPolicies []*v1alpha1.BindingPolicy,
	oldLabels, newLabels map[string]string) bool {
	for _, bindingPolicy := range bindingPolicies {
		for _, clause := range bindingPolicy.Spec.Downsync {
			selectors := clause.NamespaceSelectors
			if len(selectors) == 0 || ALabelSelectorIsEmpty(selectors...) {
				continue
			}
			if labelsMatchAny(logger, oldLabels, selectors) != labelsMatchAny(logger, newLabels, selectors) {
				return true
			}
		}
	}
	return false
}

// requeueNamespaceObjects enqueues all the workload objects in the given namespace.
// BindingPolicy and Binding objects are not workload objects and are skipped.
func (c *Controller) requeueNamespaceObjects(ctx context.Context, namespace string) error {
	logger := klog.FromContext(ctx)

	return c.listers.Iterator(func(key schema.GroupVersionResource, lister cache.GenericLister) error {
		if key == util.GetBindingPolicyGVR() || key == util.GetBindingGVR() {
			return nil // continue iterating
		}

		objs, err := lister.ByNamespace(namespace).List(labels.Everything())
		if err != nil {
			return fmt.Errorf("failed to list objects for key %v in namespace %q: %w", key, namespace, err)
		}

		for _, obj := range objs {
			logger.V(5).Info("Enqueuing workload object due to namespace label change",
				"listerKey", key, "obj", util.RefToRuntimeObj(obj), "namespace", namespace)
			c.enqueueObject(obj, key.GroupResource().Resource)
		}

		return nil // continue iterating
	})
}
