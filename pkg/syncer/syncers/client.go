/*
Copyright 2022 The KCP Authors.

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

package syncers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"
)

type Client struct {
	ResourceClient dynamic.NamespaceableResourceInterface
	scope          meta.RESTScope
}

func (c *Client) IsNamespaced() bool {
	return c.scope == meta.RESTScopeNamespace
}

func (c *Client) Create(resource edgev1alpha1.EdgeSyncConfigResource, unstObj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	var createdObj *unstructured.Unstructured
	var err error
	if c.IsNamespaced() {
		createdObj, err = c.ResourceClient.Namespace(resource.Namespace).Create(context.Background(), unstObj, v1.CreateOptions{})
	} else {
		createdObj, err = c.ResourceClient.Create(context.Background(), unstObj, v1.CreateOptions{})
	}
	return createdObj, err
}

func (c *Client) Get(resource edgev1alpha1.EdgeSyncConfigResource) (*unstructured.Unstructured, error) {
	var unstObj *unstructured.Unstructured
	var err error
	if c.IsNamespaced() {
		unstObj, err = c.ResourceClient.Namespace(resource.Namespace).Get(context.Background(), resource.Name, v1.GetOptions{})
	} else {
		unstObj, err = c.ResourceClient.Get(context.Background(), resource.Name, v1.GetOptions{})
	}
	return unstObj, err
}

func (c *Client) Update(resource edgev1alpha1.EdgeSyncConfigResource, unstObj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	var updatedObj *unstructured.Unstructured
	var err error
	if c.IsNamespaced() {
		// updatedObj, err = c.ResourceClient.Namespace(resource.Namespace).Apply(context.Background(), unstObj.GetName(), unstObj, v1.ApplyOptions{FieldManager: "application/apply-patch"})
		updatedObj, err = c.ResourceClient.Namespace(resource.Namespace).Update(context.Background(), unstObj, v1.UpdateOptions{})
	} else {
		updatedObj, err = c.ResourceClient.Update(context.Background(), unstObj, v1.UpdateOptions{})
	}
	return updatedObj, err
}

func (c *Client) UpdateStatus(resource edgev1alpha1.EdgeSyncConfigResource, unstObj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	var updatedObj *unstructured.Unstructured
	var err error
	if c.IsNamespaced() {
		// updatedObj, err = c.ResourceClient.Namespace(resource.Namespace).Apply(context.Background(), unstObj.GetName(), unstObj, v1.ApplyOptions{FieldManager: "application/apply-patch"})
		updatedObj, err = c.ResourceClient.Namespace(resource.Namespace).UpdateStatus(context.Background(), unstObj, v1.UpdateOptions{})
	} else {
		updatedObj, err = c.ResourceClient.UpdateStatus(context.Background(), unstObj, v1.UpdateOptions{})
	}
	return updatedObj, err
}
