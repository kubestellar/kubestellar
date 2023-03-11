/*
Copyright 2023 The KCP Authors.

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

package scheduler

import (
	"context"

	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
)

func (c *controller) reconcileOnSyncTarget(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)
	ws, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(key)
	if err != nil {
		logger.Error(err, "invalid key")
		return err
	}
	logger = logger.WithValues("workspace", ws, "syncTarget", name)
	logger.V(2).Info("reconciling")

	/*
		On synctarget change:
		- find all its loc(s)
		- compose all sp(s)

		- find all its ep(s)
		- for each of its obsolete ep, remove all sp(s)
		- for each of its ongoing ep, update all sp(s)
		- for each of its new ep, add all sp(s)

		Need data structure:
		- map from a synctarget to its locations, say 'locSelectSt'
		- map from a synctarget to its eps, say 'epSelectSt'
	*/

	return nil
}
