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
	upstreamcache "k8s.io/client-go/tools/cache"
)

// TypedContextualListWatcher is a ListWatcher that takes a Context
// and returns a specific type of list object.
// This may be all-cluster or specific to one cluster.
// For an all-cluster example,
// see https://github.com/kcp-dev/kcp/blob/v0.11.0/pkg/client/clientset/versioned/cluster/typed/scheduling/v1alpha1/placement.go#L47-L48 ;
// `PlacementClusterInterface` is a subtype of `TypedContextualListWatcher[*schedulingv1alpha1.PlacementList]`.
type TypedContextualListWatcher[ListType runtime.Object] interface {
	List(ctx context.Context, opts metav1.ListOptions) (ListType, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
}

// NewListerWatcher binds a TypedContextualListWatcher with a Context and
// adjusts the List return type to match the upstream interface
func NewListerWatcher[ListType runtime.Object](ctx context.Context, clw TypedContextualListWatcher[ListType]) upstreamcache.ListerWatcher {
	return &upstreamcache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return clw.List(ctx, opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return clw.Watch(ctx, opts)
		},
	}
}
