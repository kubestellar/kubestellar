/*
Copyright 2024 The KubeStellar Authors.

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

package metrics

import (
	"context"
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
)

type MRObject interface {
	metav1.Object
	runtime.Object
}

// ClientModNamespace is the commonly used methods of the typed stubs for a given object type.
// These are the methods that a cluster-scoped kind of object has,
// and the methods that a namespace-scoped kind of object has after specializing to a namespace.
type ClientModNamespace[Single MRObject, List runtime.Object] interface {
	Create(ctx context.Context, object Single, opts metav1.CreateOptions) (Single, error)
	Update(ctx context.Context, object Single, opts metav1.UpdateOptions) (Single, error)
	UpdateStatus(ctx context.Context, object Single, opts metav1.UpdateOptions) (Single, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (Single, error)
	List(ctx context.Context, opts metav1.ListOptions) (List, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result Single, err error)
}

// NamespacedClient is similar to the interface of the typed stubs for a namespace-scoped kind of object,
// but uses a fixed name for the method that specializes to a namespace.
type NamespacedClient[Single MRObject, List runtime.Object] interface {
	Namespace(string) ClientModNamespace[Single, List]
}

type namespacedClient[Single MRObject, List runtime.Object] struct {
	namespaceFn func(string) ClientModNamespace[Single, List]
}

var _ NamespacedClient[MRObject, MRObject] = &namespacedClient[MRObject, MRObject]{}

func (nsc *namespacedClient[Single, List]) Namespace(namespace string) ClientModNamespace[Single, List] {
	return nsc.namespaceFn(namespace)
}

type MeasuredClientModNamespace[Single MRObject, List runtime.Object] interface {
	ClientModNamespace[Single, List]
	ClientResourceMetrics
}

type MeasuredNamespacedClient[Single MRObject, List runtime.Object] interface {
	NamespacedClient[Single, List]
	ClientResourceMetrics
}

type wrappedClientMetrics[Single MRObject, List runtime.Object] struct {
	ClientResourceMetrics
	Inner ClientModNamespace[Single, List]
}

var _ MeasuredClientModNamespace[MRObject, MRObject] = &wrappedClientMetrics[MRObject, MRObject]{}

func NewWrappedClusterScopedClient[Single MRObject, List runtime.Object](cm ClientMetrics, gvr schema.GroupVersionResource, inner ClientModNamespace[Single, List]) MeasuredClientModNamespace[Single, List] {
	return &wrappedClientMetrics[Single, List]{
		Inner:                 inner,
		ClientResourceMetrics: cm.ResourceMetrics(gvr),
	}
}

func NewWrappedNamespacedClient[Single MRObject, List runtime.Object](cm ClientMetrics, gvr schema.GroupVersionResource, namespaceFn func(string) ClientModNamespace[Single, List]) MeasuredNamespacedClient[Single, List] {
	return &wrappedNamespacingMetrics[Single, List]{
		namespaceFn:           namespaceFn,
		ClientResourceMetrics: cm.ResourceMetrics(gvr),
	}
}

type wrappedNamespacingMetrics[Single MRObject, List runtime.Object] struct {
	ClientResourceMetrics
	namespaceFn func(string) ClientModNamespace[Single, List]
}

func (wnm *wrappedNamespacingMetrics[Single, List]) Namespace(namespace string) ClientModNamespace[Single, List] {
	inner := wnm.namespaceFn(namespace)
	return &wrappedClientMetrics[Single, List]{
		Inner:                 inner,
		ClientResourceMetrics: wnm.ClientResourceMetrics}
}

var errPanic = errors.New("panic")

func (wcm *wrappedClientMetrics[Single, List]) Create(ctx context.Context, object Single, opts metav1.CreateOptions) (Single, error) {
	var ans Single
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("create", err, time.Since(start)) }()
	ans, err = wcm.Inner.Create(ctx, object, opts)
	return ans, err
}

func (wcm *wrappedClientMetrics[Single, List]) Update(ctx context.Context, object Single, opts metav1.UpdateOptions) (Single, error) {
	var ans Single
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("update", err, time.Since(start)) }()
	ans, err = wcm.Inner.Update(ctx, object, opts)
	return ans, err
}

func (wcm *wrappedClientMetrics[Single, List]) UpdateStatus(ctx context.Context, object Single, opts metav1.UpdateOptions) (Single, error) {
	var ans Single
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("update_status", err, time.Since(start)) }()
	ans, err = wcm.Inner.UpdateStatus(ctx, object, opts)
	return ans, err
}

func (wcm *wrappedClientMetrics[Single, List]) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("delete", err, time.Since(start)) }()
	err = wcm.Inner.Delete(ctx, name, opts)
	return err
}

func (wcm *wrappedClientMetrics[Single, List]) Get(ctx context.Context, name string, opts metav1.GetOptions) (Single, error) {
	var ans Single
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("get", err, time.Since(start)) }()
	ans, err = wcm.Inner.Get(ctx, name, opts)
	return ans, err
}

func (wcm *wrappedClientMetrics[Single, List]) List(ctx context.Context, opts metav1.ListOptions) (List, error) {
	var ans List
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("list", err, time.Since(start)) }()
	ans, err = wcm.Inner.List(ctx, opts)
	return ans, err
}

func (wcm *wrappedClientMetrics[Single, List]) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var ans watch.Interface
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("watch", err, time.Since(start)) }()
	ans, err = wcm.Inner.Watch(ctx, opts)
	return ans, err
}

func (wcm *wrappedClientMetrics[Single, List]) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (Single, error) {
	var ans Single
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("patch", err, time.Since(start)) }()
	ans, err = wcm.Inner.Patch(ctx, name, pt, data, opts, subresources...)
	return ans, err
}
