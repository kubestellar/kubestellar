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
	"github.com/kubestellar/kubestellar/pkg/util"
)

func (c *Controller) syncCombinedStatus(ctx context.Context, ref string) error {
	logger := klog.FromContext(ctx)

	ns, name, err := cache.SplitMetaNamespaceKey(ref)
	if err != nil {
		return err
	}
	logger.V(5).Info("Syncing CombinedStatus", "ns", ns, "name", name)

	bindingName, sourceObjectIdentifier, exists := c.combinedStatusResolver.ResolutionExists(name) // name is unique
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

	generatedCombinedStatus := c.combinedStatusResolver.CompareCombinedStatus(bindingName,
		sourceObjectIdentifier, combinedStatus)
	if generatedCombinedStatus == nil {
		logger.V(4).Info("CombinedStatus is up-to-date", "ns", ns, "name", name, "binding", bindingName, "sourceObjectIdentifier", sourceObjectIdentifier)
		return nil
	}

	generatedCombinedStatus.ResourceVersion = combinedStatus.ResourceVersion // in case of update
	if err = c.updateOrCreateCombinedStatus(ctx, bindingName, sourceObjectIdentifier, generatedCombinedStatus); err != nil {
		return fmt.Errorf("failed to update or create CombinedStatus: %w", err)
	} // all the call's exit routes log the event

	return nil
}

func (c *Controller) updateOrCreateCombinedStatus(ctx context.Context,
	bindingName string, sourceObjectIdentifier util.ObjectIdentifier,
	generatedCombinedStatus *v1alpha1.CombinedStatus) error {
	logger := klog.FromContext(ctx)

	if generatedCombinedStatus.ResourceVersion != "" {
		csEcho, err := c.combinedStatusClient.Namespace(generatedCombinedStatus.Namespace).Update(ctx,
			generatedCombinedStatus, metav1.UpdateOptions{FieldManager: ControllerName})
		if err != nil {
			if errors.IsNotFound(err) {
				logger.V(2).Info("CombinedStatus not found (update skipped)",
					"binding", bindingName, "sourceObjectIdentifier", sourceObjectIdentifier,
					"ns", generatedCombinedStatus.Namespace, "name", generatedCombinedStatus.Name)
				return nil // the object was deleted during the syncing procedure, event will be handled
			}

			return fmt.Errorf("failed to update CombinedStatus (ns, name = %v, %v): %w",
				generatedCombinedStatus.Namespace, generatedCombinedStatus.Name, err)
		}

		logger.V(2).Info("Updated CombinedStatus", "ns", generatedCombinedStatus.Namespace,
			"name", generatedCombinedStatus.Name, "resourceVersion", csEcho.ResourceVersion,
			"binding", bindingName, "sourceObjectIdentifier", sourceObjectIdentifier)
		return nil
	}

	csEcho, err := c.combinedStatusClient.Namespace(generatedCombinedStatus.Namespace).Create(ctx,
		generatedCombinedStatus, metav1.CreateOptions{FieldManager: ControllerName})
	if err != nil {
		return fmt.Errorf("failed to create CombinedStatus (ns, name = %v, %v): %w",
			generatedCombinedStatus.Namespace, generatedCombinedStatus.Name, err)
	}

	logger.V(2).Info("Created CombinedStatus", "ns", generatedCombinedStatus.Namespace,
		"name", generatedCombinedStatus.Name, "resourceVersion", csEcho.ResourceVersion,
		"binding", bindingName, "sourceObjectIdentifier", sourceObjectIdentifier)
	return nil
}

func (c *Controller) deleteCombinedStatus(ctx context.Context, ns, name string) error {
	logger := klog.FromContext(ctx)

	err := c.combinedStatusClient.Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(2).Info("CombinedStatus not found (deletion skipped)", "ns", ns, "name", name)
			return nil
		}

		return fmt.Errorf("failed to delete CombinedStatus (ns, name = %v, %v): %w", ns, name, err)
	}

	logger.V(2).Info("Deleted CombinedStatus", "ns", ns, "name", name)
	return nil
}

func getCombinedStatusName(bindingUID, sourceObjectUID string) string {
	// The name of the CombinedStatus object is the concatenation of:
	// - the UID of the workload object
	// - the string ":"
	// - the UID of the BindingPolicy object.
	return fmt.Sprintf("%s.%s", sourceObjectUID, bindingUID)
}
