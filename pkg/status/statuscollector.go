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

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func (c *Controller) syncStatusCollector(ctx context.Context, ref string) error {
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

	// TODO: validate the statusCollector if not nil
	combinedStatusSet := c.combinedStatusResolver.NoteStatusCollector(statusCollector, isDeleted, c.workStatusIndexer)
	for combinedStatus := range combinedStatusSet {
		c.workqueue.AddAfter(combinedStatusRef(combinedStatus.ObjectName.AsNamespacedName().String()), queueingDelay)
	}

	c.logger.Info("Synced StatusCollector", "ref", ref)
	return nil
}
