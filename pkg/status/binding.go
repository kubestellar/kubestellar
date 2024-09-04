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

	"k8s.io/klog/v2"
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

	if err := c.reconcileSingletonByBdg(ctx, key); err != nil {
		return err
	}

	logger.V(5).Info("Synced Binding", "key", key)
	return nil
}
