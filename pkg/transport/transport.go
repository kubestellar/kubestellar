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

package transport

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Transport interface {
	// WrapObjects gets slice of Wrapee and wraps them into a single wrapped object.
	// In case slice is empty, the function should return an empty wrapped object (not nil).
	WrapObjects(objects []Wrapee) runtime.Object
}

// Wrapee is a workload object to wrap and its associated create-only bit
type Wrapee struct {
	Object     *unstructured.Unstructured
	Resource   string // as in schema.GroupVersionResource
	CreateOnly bool
}

func (wr Wrapee) GetObject() *unstructured.Unstructured { return wr.Object }

func NewWrapee(object *unstructured.Unstructured, resource string, createOnly bool) Wrapee {
	return Wrapee{object, resource, createOnly}
}
