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
	// WrapObjects gets slice of objects and wraps them into a single wrapped object.
	// In case slice is empty, the function should return an empty wrapped object.
	WrapObjects(objects []*unstructured.Unstructured) runtime.Object
}

// TransportWithCreateOnly is a subtype of Transport that knows how to handle
// the create-only option.
// The presence of this subtype is an intermediate development step only, required
// because of the factoring into multiple repos.
// The first step is a release of ks/ks exposing this new type.
// After the ks/OTP repo has made a release that passes values that implement this type
// to the generic code (this package here), the generic code will be changed so that the
// `Transport` interface has just the `WrapObjectsHavingCreateOnly` method and the `TransportWithCreateOnly`
// interface will be deleted. After a release
// of that, the ks/OTP repo can be changed so that its Transport implementation does
// not implement the older method (`WrapObjects`).
type TransportWithCreateOnly interface {
	Transport

	// WrapObjectsHavingCreateOnly gets slice of Wrapee and wraps them into a single wrapped object.
	// The "HavingCreateOnly" is short for "having a create-only bit",
	// not for "having the create-only bit set to true".
	// That is, this method honors the create-only bit whatever it is set to.
	// In case slice is empty, the function should return an empty wrapped object.
	WrapObjectsHavingCreateOnly(objects []Wrapee) runtime.Object
}

// Wrapee is a workload object to wrap and its associated create-only bit
type Wrapee struct {
	Object     *unstructured.Unstructured
	CreateOnly bool
}

func (wr Wrapee) GetObject() *unstructured.Unstructured { return wr.Object }

func (wr Wrapee) GetCreateOnly() bool { return wr.CreateOnly }

func NewWrapee(object *unstructured.Unstructured, createOnly bool) Wrapee {
	return Wrapee{object, createOnly}
}
