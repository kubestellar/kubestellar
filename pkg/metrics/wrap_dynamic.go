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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

type MeasuredDynamicClient interface {
	dynamic.Interface
	ClientMetrics
}

type wrappedDynamicClient struct {
	ClientMetrics
	Inner dynamic.Interface
}

type wrappedDynamicResourceInteface struct {
	wrappedDynamicModNamespace
	Inner dynamic.NamespaceableResourceInterface
}

type wrappedDynamicModNamespace struct {
	ClientResourceMetrics
	Inner dynamic.ResourceInterface
}

func NewWrappedDynamicClient(clientMetrics ClientMetrics, inner dynamic.Interface) MeasuredDynamicClient {
	return &wrappedDynamicClient{
		Inner:         inner,
		ClientMetrics: clientMetrics,
	}
}

func (wdc *wrappedDynamicClient) Resource(gvr schema.GroupVersionResource) dynamic.NamespaceableResourceInterface {
	inner := wdc.Inner.Resource(gvr)
	return &wrappedDynamicResourceInteface{
		wrappedDynamicModNamespace: wrappedDynamicModNamespace{
			ClientResourceMetrics: wdc.ClientMetrics.ResourceMetrics(gvr),
			Inner:                 inner,
		},
		Inner: inner,
	}
}

func (wdr *wrappedDynamicResourceInteface) Namespace(namespace string) dynamic.ResourceInterface {
	return &wrappedDynamicModNamespace{
		ClientResourceMetrics: wdr.ClientResourceMetrics,
		Inner:                 wdr.Inner.Namespace(namespace),
	}
}

func (wdn *wrappedDynamicModNamespace) Apply(ctx context.Context, name string, object *unstructured.Unstructured, options metav1.ApplyOptions, subresources ...string) (*unstructured.Unstructured, error) {
	var ans *unstructured.Unstructured
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("apply", err, time.Since(start)) }()
	ans, err = wdn.Inner.Apply(ctx, name, object, options, subresources...)
	return ans, err

}

func (wdn *wrappedDynamicModNamespace) ApplyStatus(ctx context.Context, name string, object *unstructured.Unstructured, options metav1.ApplyOptions) (*unstructured.Unstructured, error) {
	var ans *unstructured.Unstructured
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("apply_status", err, time.Since(start)) }()
	ans, err = wdn.Inner.Apply(ctx, name, object, options)
	return ans, err

}

func (wdn *wrappedDynamicModNamespace) Create(ctx context.Context, object *unstructured.Unstructured, opts metav1.CreateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	var ans *unstructured.Unstructured
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("create", err, time.Since(start)) }()
	ans, err = wdn.Inner.Create(ctx, object, opts, subresources...)
	return ans, err
}

func (wdn *wrappedDynamicModNamespace) Update(ctx context.Context, object *unstructured.Unstructured, opts metav1.UpdateOptions, subresources ...string) (*unstructured.Unstructured, error) {
	var ans *unstructured.Unstructured
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("update", err, time.Since(start)) }()
	ans, err = wdn.Inner.Update(ctx, object, opts, subresources...)
	return ans, err
}

func (wdn *wrappedDynamicModNamespace) UpdateStatus(ctx context.Context, object *unstructured.Unstructured, opts metav1.UpdateOptions) (*unstructured.Unstructured, error) {
	var ans *unstructured.Unstructured
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("update_status", err, time.Since(start)) }()
	ans, err = wdn.Inner.UpdateStatus(ctx, object, opts)
	return ans, err
}

func (wdn *wrappedDynamicModNamespace) Delete(ctx context.Context, name string, opts metav1.DeleteOptions, subresources ...string) error {
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("delete", err, time.Since(start)) }()
	err = wdn.Inner.Delete(ctx, name, opts, subresources...)
	return err
}

func (wdn *wrappedDynamicModNamespace) DeleteCollection(ctx context.Context, options metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("delete_collection", err, time.Since(start)) }()
	err = wdn.Inner.DeleteCollection(ctx, options, listOptions)
	return err
}

func (wdn *wrappedDynamicModNamespace) Get(ctx context.Context, name string, opts metav1.GetOptions, subresources ...string) (*unstructured.Unstructured, error) {
	var ans *unstructured.Unstructured
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("get", err, time.Since(start)) }()
	ans, err = wdn.Inner.Get(ctx, name, opts, subresources...)
	return ans, err
}

func (wdn *wrappedDynamicModNamespace) List(ctx context.Context, opts metav1.ListOptions) (*unstructured.UnstructuredList, error) {
	var ans *unstructured.UnstructuredList
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("list", err, time.Since(start)) }()
	ans, err = wdn.Inner.List(ctx, opts)
	return ans, err
}

func (wdn *wrappedDynamicModNamespace) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var ans watch.Interface
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("watch", err, time.Since(start)) }()
	ans, err = wdn.Inner.Watch(ctx, opts)
	return ans, err
}

func (wdn *wrappedDynamicModNamespace) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*unstructured.Unstructured, error) {
	var ans *unstructured.Unstructured
	err := errPanic
	start := time.Now()
	defer func() { wdn.ResourceRecord("patch", err, time.Since(start)) }()
	ans, err = wdn.Inner.Patch(ctx, name, pt, data, opts, subresources...)
	return ans, err
}
