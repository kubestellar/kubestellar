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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

type Transport interface {
	// WrapObjects gets slice of Wrapee and wraps them into a single wrapped object.
	// In case slice is empty, the function should return an empty wrapped object (not nil).
	// `kindToResource` has an answer for every `GroupKind` of the wrapees.
	// `kindToResource` can be nil if `len(objects) == 0`
	WrapObjects(objects []Wrapee, kindToResource func(schema.GroupKind) string) runtime.Object

	// UnwrapObjects extracts the Gloss for a given bundle.
	// This almost an the inverse of WrapObjects, but does not return the full contents of each workload object.
	// `kindToResource` is a typical Map.Get function.
	UnwrapObjects(wrapped runtime.Object, kindToResource func(schema.GroupKind) (string, bool)) (Gloss, error)
}

// Wrapee is a workload object to wrap and its associated create-only bit
type Wrapee struct {
	Object     *unstructured.Unstructured
	CreateOnly bool
}

// Gloss is a set of identities of workload objects
type Gloss = sets.Set[util.GKObjRef]

func (wr Wrapee) GetObject() *unstructured.Unstructured { return wr.Object }

func (wr Wrapee) GetID() util.GKObjRef {
	return util.GKObjRef{
		GK: wr.Object.GroupVersionKind().GroupKind(),
		OR: klog.KObj(wr.Object),
	}
}

func NewWrapee(object *unstructured.Unstructured, createOnly bool) Wrapee {
	return Wrapee{object, createOnly}
}
