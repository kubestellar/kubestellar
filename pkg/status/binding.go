/*
Copyright 2024 The KubeStellar Authors.

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

package status

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func (c *Controller) syncBinding(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)

	isDeleted := false
	resolution := c.bindingPolicyResolver.Broker().GetResolution(key)
	if resolution == nil {
		// If a binding key gets here and no resolution exists, then isDeleted can be set to true.
		isDeleted = true
	}
	logger.V(5).Info("In syncBinding", "bindingName", key, "isDeleted", isDeleted, "resolutionIsNil", resolution == nil, "resolution", resolution, "resolutionType", fmt.Sprintf("%T", resolution))

	// NoteBindingResolution does not use the resolution if isDeleted is true
	changedCombinedStatuses := c.combinedStatusResolver.NoteBindingResolution(ctx, key, resolution, isDeleted,
		c.workStatusIndexer, c.statusCollectorLister)
	for combinedStatus := range changedCombinedStatuses {
		logger.V(5).Info("Enqueuing CombinedStatus due to sync of Binding", "combinedStatus", combinedStatus.ObjectName, "bindingName", key)
		c.workqueue.AddAfter(combinedStatusRef(combinedStatus.ObjectName.AsNamespacedName().String()), queueingDelay)
	}

	// If at least one StatusCollector object is missing, use a Condition to show a list of missing object(s),
	// o.w. use the Condition to indicate all StatusCollector object(s) are available.
	missingStatusCollectors := c.combinedStatusResolver.MissingStatusCollectors(key)
	if len(missingStatusCollectors) > 0 {
		logger.V(4).Info("Missing StatusCollector(s)", "missingStatusCollectors", missingStatusCollectors, "binding", key)
	}
	bdg, err := c.bindingLister.Get(key)
	if err != nil {
		return fmt.Errorf("failed to get Binding %s from cache: %w", key, err)
	}
	if err = c.createOrUpdateStatusCollectorAvailableCondition(ctx, bdg, missingStatusCollectors); err != nil {
		return fmt.Errorf("failed to update status for Binding %s: %w", key, err)
	}

	logger.V(5).Info("Synced Binding", "key", key)
	return nil
}

// createOrUpdateStatusCollectorAvailableCondition maintains a Condition of type StatusCollectorsAvailable
// in a Binding object's status.
// missingSCs, a slice of the missing StatusCollector object name(s), must be sorted.
func (c *Controller) createOrUpdateStatusCollectorAvailableCondition(ctx context.Context, bdg *v1alpha1.Binding, missingSCs []string) error {
	// compose tentative condition where LastTransitionTime is TBD
	conditionTentative := v1alpha1.BindingPolicyCondition{}
	if len(missingSCs) != 0 {
		conditionTentative = v1alpha1.BindingPolicyCondition{
			Type:    v1alpha1.TypeStatusCollectorsAvailable,
			Status:  corev1.ConditionFalse,
			Reason:  v1alpha1.ReasonReconcileError,
			Message: fmt.Sprintf("Missing StatusCollector(s) %s", missingSCs),
		}
	} else {
		conditionTentative = v1alpha1.BindingPolicyCondition{
			Type:    v1alpha1.TypeStatusCollectorsAvailable,
			Status:  corev1.ConditionTrue,
			Reason:  v1alpha1.ReasonReconcileSuccess,
			Message: "All StatusCollector(s) are available",
		}
	}
	// create or update if necessary
	bdgWithProposedCondition := bdg.DeepCopy()
	conditions, changed := v1alpha1.SetCondition(bdgWithProposedCondition.Status.Conditions, conditionTentative)
	if !changed {
		return nil
	}
	bdgWithProposedCondition.Status.Conditions = conditions
	if _, err := c.bindingClient.UpdateStatus(ctx, bdgWithProposedCondition, metav1.UpdateOptions{FieldManager: ControllerName}); err != nil {
		return err
	}
	return nil
}
