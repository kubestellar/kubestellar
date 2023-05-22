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
// These are among the generated code for a cluster-scoped resource.
// For an example,
// see https://github.com/kcp-dev/kcp/blob/v0.11.0/pkg/client/clientset/versioned/cluster/typed/scheduling/v1alpha1/placement.go#L47-L48 ;
// `PlacementClusterInterface` is a subtype of `ClusterListerWatcher[PlacementInterface, *schedulingv1alpha1.PlacementList]`.
type ClusterListerWatcher[Scoped ScopedListerWatcher[ListType], ListType runtime.Object] interface {
	Cluster(logicalcluster.Path) Scoped
}

// NamespacedClusterListerWatcher is something that can map a cluster to a Namespacer for that cluster.
// These are among the generated code for a namespace-scoped resource.
// For an example, see the generated code for ReplicaSet at
// https://github.com/kcp-dev/client-go/blob/aeff170a288b8e135b5db640d08e6e43da0f46ec/kubernetes/typed/apps/v1/replicaset.go#L45 ;
// ReplicaSetClusterInterface is a subtype of NamespacedClusterListerWatcher[ReplicaSetNamespacer, appsv1client.ReplicaSetInterface, *ReplicaSetList]
type NamespacedClusterListerWatcher[NSer Namespacer[Scoped, ListType],
	Scoped ScopedListerWatcher[ListType],
	ListType runtime.Object,
] interface {
	Cluster(logicalcluster.Path) NSer
}

// Namespacer is something that takes a namespace (possibly metav1.NamespaceAll) and returns a ScopedListerWatcher.
// For an example, see the generated code for ReplicaSet at
// https://github.com/kcp-dev/client-go/blob/aeff170a288b8e135b5db640d08e6e43da0f46ec/kubernetes/typed/apps/v1/replicaset.go#L75 ;
// ReplicaSetsNamespacer is a subtype of Namespacer[appsv1client.ReplicaSetInterface, *ReplicaSetList] .
type Namespacer[Scoped ScopedListerWatcher[ListType], ListType runtime.Object] interface {
	Namespace(string) Scoped
}

// FixNamespace makes a ClusterListerWatcher out of a NamespacedClusterListerWatcher and
// a namespace value (possibly metav1.NamespaceAll).
func FixNamespace[NSer Namespacer[Scoped, ListType],
	Scoped ScopedListerWatcher[ListType],
	ListType runtime.Object,
](ncl NamespacedClusterListerWatcher[NSer, Scoped, ListType], namespace string,
) ClusterListerWatcher[Scoped, ListType] {
	return &nsFixer[NSer, Scoped, ListType]{ncl, namespace}
}

type nsFixer[NSer Namespacer[Scoped, ListType],
	Scoped ScopedListerWatcher[ListType],
	ListType runtime.Object,
] struct {
	ncl       NamespacedClusterListerWatcher[NSer, Scoped, ListType]
	namespace string
}

func (nsf *nsFixer[NSer, Scoped, ListType]) Cluster(cluster logicalcluster.Path) Scoped {
	cn := nsf.ncl.Cluster(cluster)
	return cn.Namespace(nsf.namespace)
}
