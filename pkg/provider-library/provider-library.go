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

package providerlibrary

import (
	"context"
	"sync"

	clusterproviderclient "github.com/kcp-dev/edge-mc/cluster-provider-client"
	lcv1alpha1apis "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
	edgeclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
)

var (
	providerList map[string]clusterproviderclient.ProviderClient
	lock         sync.Mutex
)

func New() {
	lock.Lock()
	defer lock.Unlock()
	providerList = make(map[string]clusterproviderclient.ProviderClient)
}

func GetProvider(ctx context.Context,
	clusterclientset edgeclient.Interface,
	providerName string,
	providerType lcv1alpha1apis.ClusterProviderType) (clusterproviderclient.ProviderClient, error) {
	lock.Lock()
	defer lock.Unlock()

	provider, exists := providerList[providerName]
	if !exists {
		newProvider, err := clusterproviderclient.NewProvider(ctx, clusterclientset, providerName, providerType)
		if err != nil {
			return nil, err
		}
		providerList[providerName] = newProvider
		w, err := newProvider.Watch()
		if err != nil {
			return nil, err
		}
		go clusterproviderclient.ProcessProviderWatchEvents(ctx, w, clusterclientset, providerName)
		provider = newProvider
	}
	return provider, nil
}
