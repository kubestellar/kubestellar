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

	"k8s.io/klog/v2"
)

func (c *Controller) syncBinding(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)

	isDeleted := false
	resolution := c.bindingResolutionBroker.GetResolution(key)
	if resolution == nil {
		// If a binding key gets here and no resolution exists, then isDeleted can be set to true.
		isDeleted = true
	}

	// NoteBindingResolution does not use the resolution if isDeleted is true
	changedCombinedStatuses := c.combinedStatusResolver.NoteBindingResolution(key, resolution, isDeleted,
		c.workStatusIndexer, c.statusCollectorLister)
	for combinedStatus := range changedCombinedStatuses {
		c.workqueue.AddAfter(combinedStatusRef(combinedStatus.ObjectName.AsNamespacedName().String()), queueingDelay)
	}

	logger.Info("Handled Binding", "key", key)
	return nil
}
