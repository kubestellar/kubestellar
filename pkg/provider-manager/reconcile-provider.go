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

package providermanager

import (
	"errors"

	"k8s.io/apimachinery/pkg/util/runtime"

	lcv1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
)

func (c *controller) handleAdd(objProvider interface{}) {
	// TODO: _ should be providerDesc. Changing to _ to avoid the compile check until we start using this variable.
	_, ok := objProvider.(*lcv1alpha1.ClusterProviderDesc)
	if !ok {
		// TODO: Is HandleError() better than changing the return type?
		err := errors.New("unexpected object type")
		runtime.HandleError(err)
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	// TODO: Initiate discovery
}

// handleDelete deletes a provider and its associated clusters (when managed)
// TODO: delete associated clusters when managed
func (c *controller) handleDelete(nameProvider string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if _, ok := c.listProviders[nameProvider]; !ok {
		return
	}
	delete(c.listProviders, nameProvider)
}
