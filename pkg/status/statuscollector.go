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

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (c *Controller) syncStatusCollector(ctx context.Context, key string) error {
	isDeleted := false

	statusCollector, err := c.statusCollectorLister.Get(key)
	if err != nil {
		// The resource no longer exist, which means it has been deleted.
		if errors.IsNotFound(err) {
			isDeleted = true
			statusCollector = &v1alpha1.StatusCollector{}
			statusCollector.Name = key
		}
		return err
	}

	// TODO: NoteStatusCollector assumes that the given object is valid. The check should be done here.
	combinedStatusSet := c.combinedStatusResolver.NoteStatusCollector(statusCollector, isDeleted, c.workStatusLister)
	for combinedStatus := range combinedStatusSet {
		// TODO: combinedStatusResolver.NoteStatusCollector can return a list of strings(ns/name)
		c.workqueue.AddAfter(combinedStatusRef(combinedStatus.ObjectName.AsNamespacedName().String()), queueingDelay)
	}

	c.logger.Info("Handled StatusCollector", "key", key)
	return nil
}
