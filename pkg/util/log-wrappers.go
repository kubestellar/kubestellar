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

package util

import (
	"encoding/json"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/util/sets"
)

// PrimitiveMap4LogWrapper implements logr.Marshaler by converting
// to a new map by converting the domain values by JSON marshaling.
type PrimitiveMap4LogWrapper[Domain comparable, Range any] struct{ base map[Domain]Range }

var _ logr.Marshaler = PrimitiveMap4LogWrapper[*int, func()]{}

// PrimitiveMap4Log wraps the given value to make it suitable for use as a logr data value
func PrimitiveMap4Log[Domain comparable, Range any](base map[Domain]Range) PrimitiveMap4LogWrapper[Domain, Range] {
	return PrimitiveMap4LogWrapper[Domain, Range]{base}
}

func (pm PrimitiveMap4LogWrapper[Domain, Range]) MarshalLog() any {
	forLog := make(map[string]Range, len(pm.base))
	for key, val := range pm.base {
		enc, err := json.Marshal(key)
		if err != nil {
			forLog[err.Error()] = val
		} else {
			forLog[string(enc)] = val
		}
	}
	return forLog
}

// K8sSet4LogWrapper implements logr.Marshaler by converting
// to a slice (in non-determistic order).
type K8sSet4LogWrapper[Domain comparable] struct{ sets.Set[Domain] }

var _ logr.Marshaler = K8sSet4LogWrapper[int]{}

// K8sSet4Log wraps the given Set with logr.Marshaler behavior.
// The conversion for marshaling is `kset.UnsortedList()`.
func K8sSet4Log[Domain comparable](kset sets.Set[Domain]) K8sSet4LogWrapper[Domain] {
	return K8sSet4LogWrapper[Domain]{kset}
}

func (kset K8sSet4LogWrapper[Domain]) MarshalLog() any {
	return kset.UnsortedList()
}
