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

package providerclientinterface

import (
	"context"
	"errors"

	kindprovider "github.com/kubestellar/kubestellar/cluster-provider-client/kind"
	v1alpha1apis "github.com/kubestellar/kubestellar/pkg/apis/logicalcluster/v1alpha1"
	clusterprovider "github.com/kubestellar/kubestellar/pkg/clustermanager/provider-client-interface/cluster"
)

// Each provider gets its own namespace named prefixNamespace+providerName
const prefixNamespace = "lcprovider-"

func GetNamespace(providerName string) string {
	return prefixNamespace + providerName
}

// TODO: this is termporary for stage 1. For stage 2 we expect to have a uniform interface for all informers.
func NewProvider(ctx context.Context,
	providerName string,
	providerType v1alpha1apis.ClusterProviderType) (clusterprovider.ProviderClient, error) {
	var newProvider clusterprovider.ProviderClient = nil
	switch providerType {
	case v1alpha1apis.KindProviderType:
		newProvider = kindprovider.New(providerName)
	default:
		err := errors.New("unknown provider type")
		return nil, err
	}
	return newProvider, nil
}
