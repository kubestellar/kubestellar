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

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/ocm"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const (
	KSFinalizer = "bindingpolicy.kubestellar.io/kscontroller"
)

// Handle bindingpolicy as follows:
//
// if bindingpolicy is not being deleted:
//   - update the (where) resolution of the bindingpolicy and queue the
//     associated binding for syncing.
//   - requeue workload objects to account for changes in bindingpolicy
//
// otherwise:
//   - if binding policy wants singleton-status reported, requeue all selected
//     workload objects to remove the singleton label.
//   - delete the bindingpolicy's finalizer and remove its resolution.
func (c *Controller) handleBindingPolicy(ctx context.Context, objIdentifier util.ObjectIdentifier) error {
	logger := klog.FromContext(ctx)

	obj, err := c.getObjectFromIdentifier(objIdentifier)
	if errors.IsNotFound(err) {
		// binding policy is deleted, update resolver.
		return c.deleteResolutionForBindingPolicy(ctx, objIdentifier.ObjectName.Name)
	} else if err != nil {
		return fmt.Errorf("failed to get runtime.Object from object identifier (%v): %w", objIdentifier, err)
	}

	bindingPolicy, err := runtimeObjectToBindingPolicy(obj)
	if err != nil {
		return fmt.Errorf("failed to convert runtime.Object to BindingPolicy: %w", err)
	}

	// handle requeing for changes in bindingpolicy, excluding deletion
	if !isBeingDeleted(obj) {
		if err := c.handleBindingPolicyFinalizer(ctx, bindingPolicy); err != nil {
			return fmt.Errorf("failed to handle finalizer for bindingPolicy %s: %w", bindingPolicy.Name, err)
		}

		// note bindingpolicy in resolver to create/update its resolution
		c.bindingPolicyResolver.NoteBindingPolicy(bindingPolicy)

		// update bindingpolicy resolution destinations since bindingpolicy was updated
		clusterSet, err := ocm.FindClustersBySelectors(c.ocmClient, bindingPolicy.Spec.ClusterSelectors)
		if err != nil {
			return fmt.Errorf("failed to ocm.FindClustersBySelectors: %w", err)
		}

		if bindingPolicy.Spec.WantSingletonReportedState {
			// if the bindingpolicy requires a singleton status, then we should only
			// have one destination
			// TODO: this should be removed once we have proper enforcement or error reporting for this
			clusterSet = pickSingleDestination(clusterSet)
		}

		// set destinations and enqueue binding for syncing
		// we can skip handling the error since the call to BindingPolicyResolver::NoteBindingPolicy above
		// guarantees that an error won't be returned here
		_ = c.bindingPolicyResolver.SetDestinations(bindingPolicy.GetName(), clusterSet)
		logger.V(4).Info("Enqueued Binding for syncing, while handling BindingPolicy", "name", bindingPolicy.Name)
		c.enqueueBinding(bindingPolicy.GetName())

		// requeue all objects to account for changes in bindingpolicy.
		// this does not include bindingpolicy/binding objects.
		return c.requeueWorkloadObjects(ctx, bindingPolicy.Name)
	}

	// we delete finalizer if the policy is being deleted (not yet deleted).
	if err = c.handleBindingPolicyFinalizer(ctx, bindingPolicy); err != nil {
		return fmt.Errorf("failed to handle finalizer for bindingPolicy %s: %w", bindingPolicy.Name, err)
	}

	return c.deleteResolutionForBindingPolicy(ctx, objIdentifier.ObjectName.Name)
}

func (c *Controller) deleteResolutionForBindingPolicy(ctx context.Context, bindingPolicyName string) error {
	if c.bindingPolicyResolver.ResolutionRequiresSingletonReportedState(bindingPolicyName) {
		// if the bindingpolicy required a singleton status, all selected objects should
		// be requeued in order to remove the label
		if err := c.requeueSelectedWorkloadObjects(ctx, bindingPolicyName); err != nil {
			return fmt.Errorf("failed to c.requeueForBindingPolicyChanges: %w", err)
		}
	}

	logger := klog.FromContext(ctx)
	c.bindingPolicyResolver.DeleteResolution(bindingPolicyName)
	logger.Info("Deleted resolution for bindingpolicy", "name", bindingPolicyName)

	return nil
}

func (c *Controller) requeueSelectedWorkloadObjects(ctx context.Context, bindingPolicyName string) error {
	if !c.bindingPolicyResolver.ResolutionExists(bindingPolicyName) {
		return nil
	}

	// requeue all objects that are selected by the bindingpolicy (are in its resolution)
	objectIdentifiers, err := c.bindingPolicyResolver.GetObjectIdentifiers(bindingPolicyName)
	if err != nil {
		return fmt.Errorf("failed to get object identifiers from bindingpolicy resolver for "+
			"bindingpolicy %s: %w", bindingPolicyName, err)
	}

	logger := klog.FromContext(ctx)
	for objIdentifier := range objectIdentifiers {
		logger.V(4).Info("Enqueuing workload object due to change in BindingPolicy",
			"objectIdentifier", objIdentifier, "bindingPolicyName", bindingPolicyName)
		c.enqueueObjectIdentifier(objIdentifier)
	}

	return nil
}

func (c *Controller) evaluateBindingPoliciesForUpdate(ctx context.Context, clusterId string, oldLabels labels.Set, newLabels labels.Set) {
	logger := klog.FromContext(ctx)

	logger.Info("Evaluating BindingPolicies for cluster", "clusterId", clusterId)
	bindingPolicies, err := c.listBindingPolicies()
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	for _, obj := range bindingPolicies {
		bindingPolicy, err := runtimeObjectToBindingPolicy(obj)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}
		match1, err := util.SelectorsMatchLabels(bindingPolicy.Spec.ClusterSelectors, oldLabels)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}
		match2, err := util.SelectorsMatchLabels(bindingPolicy.Spec.ClusterSelectors, newLabels)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}
		if match1 || match2 {
			logger.V(4).Info("Enqueuing workload object due to cluster and BindingPolicy", "clusterId", clusterId, "bindingPolicyName", bindingPolicy.Name)
			c.enqueueObject(bindingPolicy, util.BindingPolicyResource)
		}
	}
}

func (c *Controller) evaluateBindingPolicies(ctx context.Context, clusterId string, labelsSet labels.Set) {
	logger := klog.FromContext(ctx)

	logger.Info("evaluating BindingPolicies")
	bindingPolicies, err := c.listBindingPolicies()
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	for _, obj := range bindingPolicies {
		bindingPolicy, err := runtimeObjectToBindingPolicy(obj)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}
		match, err := util.SelectorsMatchLabels(bindingPolicy.Spec.ClusterSelectors, labelsSet)
		if err != nil {
			utilruntime.HandleError(err)
			return
		}
		if match {
			logger.V(4).Info("Enqueuing BindingPolicy due to cluster notification", "clusterId", clusterId, "bindingPolicyName", bindingPolicy.Name)
			c.enqueueObject(bindingPolicy, util.BindingPolicyResource)
		}
	}
}

func (c *Controller) listBindingPolicies() ([]runtime.Object, error) {
	lister, found := c.listers.Get(util.GetBindingPolicyGVR())
	if !found {
		return nil, fmt.Errorf("could not get lister for BindingPolicy")
	}

	list, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	return list, nil
}

func runtimeObjectToBindingPolicy(obj runtime.Object) (*v1alpha1.BindingPolicy, error) {
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to convert runtime.Object to unstructured.Unstructured")
	}
	var bindingPolicy *v1alpha1.BindingPolicy
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.UnstructuredContent(), &bindingPolicy); err != nil {
		return nil, err
	}
	return bindingPolicy, nil
}

// read objects from all workload listers and enqueue
// the keys this is useful for when a new bindingpolicy is
// added or a bindingpolicy is updated
func (c *Controller) requeueWorkloadObjects(ctx context.Context, bindingPolicyName string) error {
	logger := klog.FromContext(ctx)

	return c.listers.Iterator(func(key schema.GroupVersionResource, lister cache.GenericLister) error {
		// do not requeue bindingpolicies or bindings
		if key == util.GetBindingPolicyGVR() || key == util.GetBindingGVR() {
			logger.Info("Not enqueuing control object", "key", key)
			return nil // continue iterating
		}

		objs, err := lister.List(labels.Everything())
		if err != nil {
			return fmt.Errorf("failed to list objects for key %v: %w", key, err)
		}

		for _, obj := range objs {
			logger.V(4).Info("Enqueuing workload object due to BindingPolicy",
				"listerKey", key, "obj", util.RefToRuntimeObj(obj),
				"bindingPolicyName", bindingPolicyName)
			c.enqueueObject(obj, key.GroupResource().Resource)
		}

		return nil // continue iterating
	})
}

// finalizer logic
func (c *Controller) handleBindingPolicyFinalizer(ctx context.Context, bindingPolicy *v1alpha1.BindingPolicy) error {
	if bindingPolicy.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(bindingPolicy, KSFinalizer) {
			controllerutil.RemoveFinalizer(bindingPolicy, KSFinalizer)
			if err := updateBindingPolicy(ctx, c.dynamicClient, bindingPolicy); err != nil {
				if errors.IsNotFound(err) {
					// object was deleted after getting into this function. This is not an error.
					return nil
				}

				return err
			}
		}
		return nil
	}

	if !controllerutil.ContainsFinalizer(bindingPolicy, KSFinalizer) {
		controllerutil.AddFinalizer(bindingPolicy, KSFinalizer)
		if err := updateBindingPolicy(ctx, c.dynamicClient, bindingPolicy); err != nil {
			return err
		}
	}
	return nil
}

func updateBindingPolicy(ctx context.Context, client dynamic.Interface, bindingPolicy *v1alpha1.BindingPolicy) error {
	gvr := schema.GroupVersionResource{
		Group:    v1alpha1.GroupVersion.Group,
		Version:  v1alpha1.GroupVersion.Version,
		Resource: util.BindingPolicyResource,
	}

	innerObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(bindingPolicy)
	if err != nil {
		return fmt.Errorf("failed to convert BindingPolicy to unstructured: %w", err)
	}

	unstructuredObj := &unstructured.Unstructured{
		Object: innerObj,
	}

	_, err = client.Resource(gvr).Namespace("").Update(ctx, unstructuredObj, metav1.UpdateOptions{})
	return err
}

type mrObject interface {
	metav1.Object
	runtime.Object
}

func (c *Controller) testObject(ctx context.Context, objIdentifier util.ObjectIdentifier, objLabels map[string]string,
	tests []v1alpha1.DownsyncObjectTest) bool {
	gvr := schema.GroupVersionResource{
		Group:    objIdentifier.GVK.Group,
		Version:  objIdentifier.GVK.Version,
		Resource: objIdentifier.Resource,
	}

	logger := klog.FromContext(ctx)

	var objNS *corev1.Namespace
	for _, test := range tests {
		if test.APIGroup != nil && (*test.APIGroup) != objIdentifier.GVK.Group {
			continue
		}
		if len(test.Resources) > 0 && !(SliceContains(test.Resources, "*") ||
			SliceContains(test.Resources, gvr.Resource)) {
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
				objNS, err = c.kubernetesClient.CoreV1().Namespaces().Get(context.TODO(),
					objIdentifier.ObjectName.Namespace, metav1.GetOptions{})
				if err != nil {
					logger.Info("Object namespace not found, assuming object does not match",
						"object identifier", objIdentifier)
					continue
				}
			}
			if !labelsMatchAny(logger, objNS.Labels, test.NamespaceSelectors) {
				continue
			}
		}
		return true
	}
	return false
}

func labelsMatchAny(logger logr.Logger, labelSet map[string]string, selectors []metav1.LabelSelector) bool {
	for _, ls := range selectors {
		sel, err := metav1.LabelSelectorAsSelector(&ls)
		if err != nil {
			logger.Info("Failed to convert LabelSelector to labels.Selector", "ls", ls, "err", err)
			continue
		}
		if sel.Matches(labels.Set(labelSet)) {
			return true
		}
	}
	return false
}

func ALabelSelectorIsEmpty(selectors ...metav1.LabelSelector) bool {
	for _, sel := range selectors {
		if len(sel.MatchExpressions) == 0 && len(sel.MatchLabels) == 0 {
			return true
		}
	}
	return false
}

func SliceContains[Elt comparable](slice []Elt, seek Elt) bool {
	for _, elt := range slice {
		if elt == seek {
			return true
		}
	}
	return false
}

// sort by name and pick first cluster so that the choice is deterministic based on names
func pickSingleDestination(clusterSet sets.Set[string]) sets.Set[string] {
	return sets.New(sets.List(clusterSet)[0])
}
