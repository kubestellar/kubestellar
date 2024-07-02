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

package abstract

import (
	"k8s.io/apimachinery/pkg/util/sets"
)

// SliceDelete removes an entry, identified by position, from a slice.
// The given position must be valid.
func SliceDelete[Elt any](slice *[]Elt, index int) {
	lastIndex := len(*slice) - 1
	if index != lastIndex {
		(*slice)[index] = (*slice)[lastIndex]
	}
	*slice = (*slice)[:lastIndex]
}

// NewSliceByFilter makes a new slice that contains the members of the given slice that pass the given filter
func NewSliceByFilter[Elt any](input []Elt, good func(Elt) bool) []Elt {
	if input == nil {
		return nil
	}
	newSlice := make([]Elt, 0, len(input))
	for _, elt := range input {
		if good(elt) {
			newSlice = append(newSlice, elt)
		}
	}
	return newSlice
}

// SliceCopy copies a given slice into new storage
func SliceCopy[Elt any](input []Elt) []Elt {
	if input == nil {
		return nil
	}
	return append(make([]Elt, 0, len(input)), input...)
}

func SliceEqual[Elt comparable](slice1, slice2 []Elt) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for idx, elt1 := range slice1 {
		if elt1 != slice2[idx] {
			return false
		}
	}
	return true
}

func SliceMap[Domain, Range any](slice []Domain, fn func(Domain) Range) []Range {
	if slice == nil {
		return nil
	}
	ans := make([]Range, len(slice))
	for idx, elt := range slice {
		ans[idx] = fn(elt)
	}
	return ans
}

func SliceToPrimitiveMap[Domain, Range any, Key comparable](slice []Domain, keyFn func(Domain) Key,
	ValFn func(Domain) Range) map[Key]Range {
	if slice == nil {
		return nil
	}

	ans := make(map[Key]Range, len(slice))
	for _, elt := range slice {
		ans[keyFn(elt)] = ValFn(elt)
	}
	return ans
}

func SliceMapToK8sSet[Domain any, Range comparable](slice []Domain, fn func(Domain) Range) sets.Set[Range] {
	if slice == nil {
		return nil
	}
	ans := sets.New[Range]()
	for _, elt := range slice {
		ans.Insert(fn(elt))
	}
	return ans
}
