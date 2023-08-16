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
	"fmt"

	ksclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
	edgev1alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v1alpha1"
	// edgev1alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v1alpha1"
)

type (
	OneClusterClient ksclientset.Interface
	AllClusterClient edgeclientset.ClusterInterface

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

func IdentifyEpAccessScope[A EpAccess](access A) string {
	switch any(access).(type) {
	case string:
		return "a string"
	default:
		if a, ok := any(access).(*edgev1alpha1informers.EdgePlacementInformer); ok {
			fmt.Printf("pointer of edgev1alpha1informers.EdgePlacementInformer at %v\n", a)
			return "one cluster EdgePlacement Access"
		} else if a, ok := any(access).(*edgev1alpha1informers.EdgePlacementClusterInformer); ok {
			fmt.Printf("pointer of edgev1alpha1informers.EdgePlacementClusterInformer at %v\n", a)
			return "all Cluster EdgePlacement Access"
		}
	}
	fmt.Println("something mystery")
	return "something mystery"
}
