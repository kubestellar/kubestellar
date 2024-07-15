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
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/abstract"
)

func (c *Controller) syncStatusCollector(ctx context.Context, ref string) error {
	logger := klog.FromContext(ctx)
	isDeleted := false

	statusCollector, err := c.statusCollectorLister.Get(ref)
	if err != nil {
		// The resource no longer exist, which means it has been deleted.
		if !errors.IsNotFound(err) {
			return err
		}

		isDeleted = true // not found, should be deleted
		statusCollector = &v1alpha1.StatusCollector{}
		statusCollector.Name = ref
	}

	// Validate the StatusCollector
	if errs := c.validateStatusCollector(statusCollector); len(errs) > 0 {
		if err := c.updateStatusCollectorErrors(ctx, statusCollector, errs); err != nil {
			return err
		}

		isDeleted = true // invalid statuscollector, should be deleted if it exists
	} else if statusCollector.Status.Errors != nil {
		if err := c.updateStatusCollectorErrors(ctx, statusCollector, nil); err != nil {
			return err
		}
	}

	combinedStatusSet := c.combinedStatusResolver.NoteStatusCollector(statusCollector, isDeleted, c.workStatusIndexer)
	for combinedStatus := range combinedStatusSet {
		c.workqueue.AddAfter(combinedStatusRef(combinedStatus.ObjectName.AsNamespacedName().String()), queueingDelay)
	}

	logger.Info("Synced StatusCollector", "ref", ref)
	return nil
}

func (c *Controller) validateStatusCollector(statusCollector *v1alpha1.StatusCollector) []error {
	var errs []error
	// groupBy & CombinedFields empty if select is not
	if len(statusCollector.Spec.Select) > 0 &&
		(len(statusCollector.Spec.GroupBy) > 0 || len(statusCollector.Spec.CombinedFields) > 0) {
		errs = append(errs, fmt.Errorf("groupBy and combinedFields must be empty if select is not"))
	}
	// groupBy empty if combinedFields is
	if len(statusCollector.Spec.CombinedFields) == 0 && len(statusCollector.Spec.GroupBy) > 0 {
		errs = append(errs, fmt.Errorf("groupBy must be empty if combinedFields is"))
	}

	// structure must be valid before we get to parsing errors
	if len(errs) > 0 {
		return errs
	}

	// validate filter expression
	if err := c.celEvaluator.CheckExpression(statusCollector.Spec.Filter); err != nil {
		errs = append(errs, fmt.Errorf("filter expression invalid: %w", err))
	}

	// validate select expression
	for _, selectExpr := range statusCollector.Spec.Select {
		if err := c.celEvaluator.CheckExpression(&selectExpr.Def); err != nil {
			errs = append(errs, fmt.Errorf("select expression (%s) invalid: %w", selectExpr.Name, err))
		}
	}

	// validate groupBy expression
	for _, groupByExpr := range statusCollector.Spec.GroupBy {
		if err := c.celEvaluator.CheckExpression(&groupByExpr.Def); err != nil {
			errs = append(errs, fmt.Errorf("groupBy expression (%s) invalid: %w", groupByExpr.Name, err))
		}
	}

	// validate combinedFields expression
	for _, combinedField := range statusCollector.Spec.CombinedFields {
		if combinedField.Type == v1alpha1.AggregatorTypeCount {
			if combinedField.Subject != nil {
				errs = append(errs,
					fmt.Errorf("combinedField expression (%s) invalid: subject must be nil for %s type",
						combinedField.Name, combinedField.Type))
			}

			continue
		}

		if combinedField.Subject == nil {
			errs = append(errs, fmt.Errorf("combinedField expression (%s) invalid: subject must be set",
				combinedField.Name))
			continue
		}

		if err := c.celEvaluator.CheckExpression(combinedField.Subject); err != nil {
			errs = append(errs, fmt.Errorf("combinedField expression (%s) subject invalid: %w",
				combinedField.Name, err))
		}
	}

	return errs
}

func (c *Controller) updateStatusCollectorErrors(ctx context.Context, statusCollector *v1alpha1.StatusCollector,
	errs []error) error {
	logger := klog.FromContext(ctx)

	statusCollector.Status.Errors = abstract.SliceMap(errs, func(err error) string {
		return err.Error()
	})

	scEcho, err := c.wdsKsClient.ControlV1alpha1().StatusCollectors().UpdateStatus(ctx,
		statusCollector, metav1.UpdateOptions{FieldManager: controllerName})

	if err != nil {
		if errors.IsNotFound(err) {
			// StatusCollector has been deleted
			return nil
		} else {
			return fmt.Errorf("failed to update StatusCollector status (ns, name = %s, %s): %w",
				statusCollector.Namespace, statusCollector.Name, err)
		}
	}

	logger.Info("Updated StatusCollector status", "ns", statusCollector.Namespace,
		"name", statusCollector.Name, "resourceVersion", scEcho.ResourceVersion)
	return nil
}
