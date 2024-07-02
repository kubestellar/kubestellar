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

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func (c *Controller) syncCombinedStatus(ctx context.Context, ref string) error {
	logger := klog.FromContext(ctx)

	ns, name, err := cache.SplitMetaNamespaceKey(ref)
	if err != nil {
		return err
	}

	bindingName, sourceObjectIdentifier, exists := c.combinedStatusResolver.ResolutionExists(ref)
	if !exists {
		// if a resolution is not associated to the combined status, then it must be deleted
		return c.deleteCombinedStatus(ctx, ns, name)
	}

	combinedStatus, err := c.combinedStatusLister.CombinedStatuses(ns).Get(name)
	if errors.IsNotFound(err) {
		// object must be created
		combinedStatus = &v1alpha1.CombinedStatus{
			TypeMeta: metav1.TypeMeta{
				Kind:       "CombinedStatus",
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
	} else if err != nil {
		return fmt.Errorf("failed to get CombinedStatus from informer cache (ns=%v, name=%v): %w", ns, name, err)
	}

	if c.combinedStatusResolver.CompareCombinedStatus(*bindingName, *sourceObjectIdentifier, combinedStatus) {
		// no change
		return nil
	}

	// update combined status
	generatedCombinedStatus := c.combinedStatusResolver.GenerateCombinedStatus(*bindingName, *sourceObjectIdentifier)
	if generatedCombinedStatus == nil {
		return fmt.Errorf("syncing CombinedStatus was stopped because it has no counterpart resolution")
	}

	if err = c.updateOrCreateCombinedStatus(ctx, combinedStatus, generatedCombinedStatus); err != nil {
		return fmt.Errorf("failed to update or create CombinedStatus: %w", err)
	}

	logger.Info("Synced CombinedStatus", "ns", ns, "name", name)
	return nil
}

func (c *Controller) updateOrCreateCombinedStatus(ctx context.Context, combinedStatus,
	generatedCombinedStatus *v1alpha1.CombinedStatus) error {
	// set labels
	combinedStatus.Labels = generatedCombinedStatus.Labels
	// set results
	combinedStatus.Results = generatedCombinedStatus.Results

	logger := klog.FromContext(ctx)
	csEcho, err := c.wdsKsClient.ControlV1alpha1().CombinedStatuses(combinedStatus.Namespace).Update(ctx,
		combinedStatus, metav1.UpdateOptions{FieldManager: controllerName})

	if err != nil {
		if errors.IsNotFound(err) {
			csEcho, err = c.wdsKsClient.ControlV1alpha1().CombinedStatuses(combinedStatus.Namespace).Create(ctx,
				combinedStatus, metav1.CreateOptions{FieldManager: controllerName})
			if err != nil {
				return fmt.Errorf("failed to create CombinedStatus (ns, name = %v, %v): %w",
					combinedStatus.Namespace, combinedStatus.Name, err)
			}

			logger.Info("Created CombinedStatus", "ns", combinedStatus.Namespace,
				"name", combinedStatus.Name, "resourceVersion", csEcho.ResourceVersion)
			return nil
		} else {

			return fmt.Errorf("failed to update CombinedStatus (ns, name = %v, %v): %w",
				combinedStatus.Namespace, combinedStatus.Name, err)
		}
	}

	logger.Info("Updated CombinedStatus", "ns", combinedStatus.Namespace,
		"name", combinedStatus.Name, "resourceVersion", csEcho.ResourceVersion)
	return nil
}

func (c *Controller) deleteCombinedStatus(ctx context.Context, ns, name string) error {
	logger := klog.FromContext(ctx)

	err := c.wdsKsClient.ControlV1alpha1().CombinedStatuses(ns).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to delete CombinedStatus (ns, name = %v, %v): %w", ns, name, err)
	}

	logger.Info("Deleted CombinedStatus", "ns", ns, "name", name)
	return nil
}

func getCombinedStatusName(bindingUID, sourceObjectUID string) string {
	// The name of the CombinedStatus object is the concatenation of:
	// - the UID of the workload object
	// - the string ":"
	// - the UID of the BindingPolicy object.
	return fmt.Sprintf("%s:%s", sourceObjectUID, bindingUID)
}
