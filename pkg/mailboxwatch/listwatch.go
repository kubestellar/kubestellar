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

package mailboxwatch

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/kcp-dev/logicalcluster/v3"
)

// ScopedListerWatcher is a ListWatcher that takes a Context
// and returns a specific type of list object.
// It is specific to one cluster.
// For an example,
// see https://github.com/kcp-dev/kcp/blob/v0.11.0/pkg/client/clientset/versioned/typed/scheduling/v1alpha1/placement.go#L48-L49 ;
// `PlacementInterface` is a subtype of `ScopedListerWatcher[*schedulingv1alpha1.PlacementList]`.
type ScopedListerWatcher[ListType runtime.Object] interface {
	List(ctx context.Context, opts metav1.ListOptions) (ListType, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

// ClusterListerWatcher is something that can map a cluster to a ScopedListerWatcher for that cluster.
// For an example,
// see https://github.com/kcp-dev/kcp/blob/v0.11.0/pkg/client/clientset/versioned/cluster/typed/scheduling/v1alpha1/placement.go#L47-L48 ;
// `PlacementClusterInterface` is a subtype of `ClusterListerWatcher[PlacementInterface, *schedulingv1alpha1.PlacementList]`.
type ClusterListerWatcher[Scoped ScopedListerWatcher[ListType], ListType runtime.Object] interface {
	Cluster(logicalcluster.Path) Scoped
}
