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

	clusterproviderclient "github.com/kcp-dev/edge-mc/cluster-provider-client"
	lcv1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
)

func (c *controller) handleAdd(provider interface{}) {
	providerInfo, ok := provider.(*lcv1alpha1.ClusterProviderDesc)
	if !ok {
		// TODO: Is HandleError() better than changing the return type?
		err := errors.New("unexpected object type")
		runtime.HandleError(err)
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	// Provider descriptions are cluster wide and unique. The uniqueness is
	// determined by the provider description object name and enforced by
	// kubernetes.
	provider, err := clusterproviderclient.GetProviderClient(c.ctx, c.clientset, providerInfo.Spec.ProviderType, providerInfo.Name)
	if err != nil {
		runtime.HandleError(err)
	}
}

// handleDelete deletes cluster from the cache maps
func (c *controller) handleDelete(nameProvider string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.listProviders[nameProvider]; !ok {
		return
	}
	delete(c.listProviders, nameProvider)
}
