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

// BasicClientModNamespace is the commonly used methods of the typed stubs for a given object type.
// No methods for subresources are included here.
// These are the methods that a cluster-scoped kind of object has,
// and the methods that a namespace-scoped kind of object has after specializing to a namespace.
type BasicClientModNamespace[Single MRObject, List runtime.Object] interface {
	Create(ctx context.Context, object Single, opts metav1.CreateOptions) (Single, error)
	Update(ctx context.Context, object Single, opts metav1.UpdateOptions) (Single, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (Single, error)
	List(ctx context.Context, opts metav1.ListOptions) (List, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result Single, err error)
}

// ClientModNamespace is the commonly used methods of the typed stubs for a given object type.
// These are the methods that a cluster-scoped kind of object has,
// and the methods that a namespace-scoped kind of object has after specializing to a namespace.
type ClientModNamespace[Single MRObject, List runtime.Object] interface {
	BasicClientModNamespace[Single, List]
	UpdateStatus(ctx context.Context, object Single, opts metav1.UpdateOptions) (Single, error)
}

// BasicNamespacedClient is similar to the interface of the typed stubs for a namespace-scoped kind of object,
// but uses a fixed name for the method that specializes to a namespace.
type BasicNamespacedClient[Single MRObject, List runtime.Object] interface {
	Namespace(string) BasicClientModNamespace[Single, List]
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

type MeasuredBasicClientModNamespace[Single MRObject, List runtime.Object] interface {
	BasicClientModNamespace[Single, List]
	ClientResourceMetrics
}

type MeasuredClientModNamespace[Single MRObject, List runtime.Object] interface {
	ClientModNamespace[Single, List]
	ClientResourceMetrics
}

type MeasuredBasicNamespacedClient[Single MRObject, List runtime.Object] interface {
	BasicNamespacedClient[Single, List]
	ClientResourceMetrics
}

type MeasuredNamespacedClient[Single MRObject, List runtime.Object] interface {
	NamespacedClient[Single, List]
	ClientResourceMetrics
}

type wrappedBasicClientMetrics[Single MRObject, List runtime.Object] struct {
	ClientResourceMetrics
	Inner BasicClientModNamespace[Single, List]
}

type wrappedClientMetrics[Single MRObject, List runtime.Object] struct {
	wrappedBasicClientMetrics[Single, List]
	Inner ClientModNamespace[Single, List]
}

var _ MeasuredClientModNamespace[MRObject, MRObject] = &wrappedClientMetrics[MRObject, MRObject]{}

func NewWrappedBasicClusterScopedClient[Single MRObject, List runtime.Object](cm ClientMetrics, gvr schema.GroupVersionResource, inner BasicClientModNamespace[Single, List]) MeasuredBasicClientModNamespace[Single, List] {
	return &wrappedBasicClientMetrics[Single, List]{cm.ResourceMetrics(gvr), inner}
}

func NewWrappedClusterScopedClient[Single MRObject, List runtime.Object](cm ClientMetrics, gvr schema.GroupVersionResource, inner ClientModNamespace[Single, List]) MeasuredClientModNamespace[Single, List] {
	return &wrappedClientMetrics[Single, List]{
		wrappedBasicClientMetrics: wrappedBasicClientMetrics[Single, List]{cm.ResourceMetrics(gvr), inner},
		Inner:                     inner,
	}
}

func NewWrappedBasicNamespacedClient[Single MRObject, List runtime.Object](cm ClientMetrics, gvr schema.GroupVersionResource, namespaceFn func(string) BasicClientModNamespace[Single, List]) MeasuredBasicNamespacedClient[Single, List] {
	return &wrappedBasicNamespacingMetrics[Single, List]{
		namespaceFn:           namespaceFn,
		ClientResourceMetrics: cm.ResourceMetrics(gvr),
	}
}

func NewWrappedNamespacedClient[Single MRObject, List runtime.Object](cm ClientMetrics, gvr schema.GroupVersionResource, namespaceFn func(string) ClientModNamespace[Single, List]) MeasuredNamespacedClient[Single, List] {
	return &wrappedNamespacingMetrics[Single, List]{
		wrappedBasicNamespacingMetrics: wrappedBasicNamespacingMetrics[Single, List]{cm.ResourceMetrics(gvr), func(ns string) BasicClientModNamespace[Single, List] { return namespaceFn(ns) }},
		namespaceFn:                    namespaceFn,
	}
}

type wrappedBasicNamespacingMetrics[Single MRObject, List runtime.Object] struct {
	ClientResourceMetrics
	namespaceFn func(string) BasicClientModNamespace[Single, List]
}

type wrappedNamespacingMetrics[Single MRObject, List runtime.Object] struct {
	wrappedBasicNamespacingMetrics[Single, List]
	namespaceFn func(string) ClientModNamespace[Single, List]
}

func (wnm *wrappedBasicNamespacingMetrics[Single, List]) Namespace(namespace string) BasicClientModNamespace[Single, List] {
	inner := wnm.namespaceFn(namespace)
	return &wrappedBasicClientMetrics[Single, List]{wnm.ClientResourceMetrics, inner}
}

func (wnm *wrappedNamespacingMetrics[Single, List]) Namespace(namespace string) ClientModNamespace[Single, List] {
	inner := wnm.namespaceFn(namespace)
	return &wrappedClientMetrics[Single, List]{
		wrappedBasicClientMetrics: wrappedBasicClientMetrics[Single, List]{
			wnm.ClientResourceMetrics, inner},
		Inner: inner}
}

var errPanic = errors.New("panic")

func (wcm *wrappedBasicClientMetrics[Single, List]) Create(ctx context.Context, object Single, opts metav1.CreateOptions) (Single, error) {
	var ans Single
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("create", err, time.Since(start)) }()
	ans, err = wcm.Inner.Create(ctx, object, opts)
	return ans, err
}

func (wcm *wrappedBasicClientMetrics[Single, List]) Update(ctx context.Context, object Single, opts metav1.UpdateOptions) (Single, error) {
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

func (wcm *wrappedBasicClientMetrics[Single, List]) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("delete", err, time.Since(start)) }()
	err = wcm.Inner.Delete(ctx, name, opts)
	return err
}

func (wcm *wrappedBasicClientMetrics[Single, List]) Get(ctx context.Context, name string, opts metav1.GetOptions) (Single, error) {
	var ans Single
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("get", err, time.Since(start)) }()
	ans, err = wcm.Inner.Get(ctx, name, opts)
	return ans, err
}

func (wcm *wrappedBasicClientMetrics[Single, List]) List(ctx context.Context, opts metav1.ListOptions) (List, error) {
	var ans List
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("list", err, time.Since(start)) }()
	ans, err = wcm.Inner.List(ctx, opts)
	return ans, err
}

func (wcm *wrappedBasicClientMetrics[Single, List]) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var ans watch.Interface
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("watch", err, time.Since(start)) }()
	ans, err = wcm.Inner.Watch(ctx, opts)
	return ans, err
}

func (wcm *wrappedBasicClientMetrics[Single, List]) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (Single, error) {
	var ans Single
	err := errPanic
	start := time.Now()
	defer func() { wcm.ResourceRecord("patch", err, time.Since(start)) }()
	ans, err = wcm.Inner.Patch(ctx, name, pt, data, opts, subresources...)
	return ans, err
}
