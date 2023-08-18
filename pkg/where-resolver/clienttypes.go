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

package where_resolver

import (
	edgev1alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v1alpha1"
)

type ClusterProvider string

const (
	ClusterProviderKCP  ClusterProvider = "kcp"
	ClusterProviderKube ClusterProvider = "kubernetes"
)

type (
	EpAccess interface {
		*edgev1alpha1informers.EdgePlacementInformer | *edgev1alpha1informers.EdgePlacementClusterInformer
	}
	SpsAccess interface {
		*edgev1alpha1informers.SinglePlacementSliceInformer | *edgev1alpha1informers.SinglePlacementSliceClusterInformer
	}
	LocAccess interface {
		*edgev1alpha1informers.LocationInformer | *edgev1alpha1informers.LocationClusterInformer
	}
	StAccess interface {
		*edgev1alpha1informers.SyncTargetInformer | *edgev1alpha1informers.SyncTargetClusterInformer
	}
)

func EpAccessScopeCheck[A EpAccess](access A, provider ClusterProvider) bool {
	switch any(access).(type) {
	default:
		if _, ok := any(access).(*edgev1alpha1informers.EdgePlacementInformer); ok {
			return provider == ClusterProviderKube
		} else if _, ok := any(access).(*edgev1alpha1informers.EdgePlacementClusterInformer); ok {
			return provider == ClusterProviderKCP
		}
	}
	return false
}
