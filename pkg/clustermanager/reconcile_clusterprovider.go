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

package clustermanager

import (
	"errors"

	lcv1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
)

func (c *controller) reconcileClusterProviderDesc(key string) error {
	providerObj, exists, err := c.clusterProviderInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	provider, ok := providerObj.(*lcv1alpha1.ClusterProviderDesc)
	if !ok {
		return errors.New("unexpected object type. expected ClusterProviderDesc")
	}

	if !exists {
		c.logger.V(2).Info("ClusterProviderDesc deleted", "resource", provider.Name)
		//TODO handle ClusterProviderDesc deleted
	} else {
		//TODO handle ClusterProviderDesc added/updated
		c.logger.V(2).Info("reconcile ClusterProviderDesc", "resource", provider.Name)
	}
	return nil
}
