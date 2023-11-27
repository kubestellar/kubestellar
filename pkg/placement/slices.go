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

package placement

import (
	"fmt"
	"strings"
)

func SliceCopy[Elt any](original []Elt) []Elt {
	if original == nil {
		return original
	}
	return append([]Elt{}, original...)
}

func SliceContains[Elt comparable](slice []Elt, seek Elt) bool {
	for _, elt := range slice {
		if elt == seek {
			return true
		}
	}
	return false
}

func SliceContainsParametric[Elt any](isEqual func(Elt, Elt) bool, slice []Elt, seek Elt) bool {
	for _, elt := range slice {
		if isEqual(elt, seek) {
			return true
		}
	}
	return false
}

func SliceRemoveFunctional[Elt comparable](slice []Elt, seek Elt) []Elt {
	ans := []Elt{}
	for _, elt := range slice {
		if elt != seek {
			ans = append(ans, elt)
		}
	}
	return ans
}

func SliceEqual[Elt comparable](a, b []Elt) bool {
	if len(a) != len(b) {
		return false
	}
	for index, elta := range a {
		if elta != b[index] {
			return false
		}
	}
	return true
}

func SliceApply[Elt any](slice []Elt, fn func(Elt)) {
	for _, elt := range slice {
		fn(elt)
	}
}

func VisitableToSlice[Elt any](set Visitable[Elt]) []Elt {
	ans := []Elt{}
	set.Visit(func(elt Elt) error {
		ans = append(ans, elt)
		return nil
	})
	return ans
}

// VisitableStringer wraps a given set with particular String() behavior.
// NB: you only want to apply this to a set that is safe for concurrent access,
// and you probably only want to apply it to an immutable set.
func VisitableStringer[Elt any](set Visitable[Elt]) VisitableStringerVal[Elt] {
	return VisitableStringerVal[Elt]{set}
}

type VisitableStringerVal[Elt any] struct {
	set Visitable[Elt]
}

func (vs VisitableStringerVal[Elt]) String() string {
	var ans strings.Builder
	ans.WriteRune('{')
	first := true
	vs.set.Visit(func(elt Elt) error {
		if first {
			first = false
		} else {
			ans.WriteString(", ")
		}
		eltStr := fmt.Sprintf("%v", elt)
		ans.WriteString(eltStr)
		return nil
	})
	ans.WriteRune('}')
	return ans.String()
}

func VisitableMap[Domain, Range any](visitable Visitable[Domain], mapFn func(Domain) Range) Visitable[Range] {
	return VisitableFunc[Range](func(visitor func(Range) error) error {
		return visitable.Visit(func(input Domain) error {
			mapped := mapFn(input)
			return visitor(mapped)
		})
	})
}

func VisitableMapFnReduceOr[Elt any](visitable Visitable[Elt], test func(Elt) bool) bool {
	return visitable.Visit(func(elt Elt) error {
		if test(elt) {
			return errStop
		}
		return nil
	}) != nil
}

func VisitableTransformToSlice[Original, Transformed any](set Visitable[Original], xform func(Original) Transformed) []Transformed {
	ans := []Transformed{}
	set.Visit(func(elt Original) error {
		mapped := xform(elt)
		ans = append(ans, mapped)
		return nil
	})
	return ans
}

func MapTransformToSlice[Key, Val, Transformed any](theMap Map[Key, Val], xform func(Key, Val) Transformed) []Transformed {
	ans := make([]Transformed, 0, theMap.Len())
	theMap.Visit(func(tup Pair[Key, Val]) error {
		mapped := xform(tup.First, tup.Second)
		ans = append(ans, mapped)
		return nil
	})
	return ans
}

type VisitableFunc[Elt any] func(func(Elt) error) error

var _ Visitable[string] = VisitableFunc[string](nil)

func (vf VisitableFunc[Elt]) Visit(visitor func(Elt) error) error {
	return vf(visitor)
}

type Slice[Elt any] []Elt

var _ Visitable[float64] = Slice[float64]{}

// NewSlice makes a Visitable from a slice
func NewSlice[Elt any](elts ...Elt) Slice[Elt] { return Slice[Elt](elts) }

func (slice Slice[Elt]) IsEmpty() bool    { return len(slice) == 0 }
func (slice Slice[Elt]) Len() int         { return len(slice) }
func (slice Slice[Elt]) LenIsCheap() bool { return true }

func (slice Slice[Elt]) Visit(visitor func(Elt) error) error {
	for _, elt := range slice {
		if err := visitor(elt); err != nil {
			return err
		}
	}
	return nil
}

func (slice Slice[Elt]) Append(elt Elt) Slice[Elt] {
	return append(slice, elt)
}

type SliceSet[Elt comparable] struct{ Slice[Elt] }

var _ Set[int64] = SliceSet[int64]{}

// NewSliceSet makes a Set from a slice that has no duplicates
func NewSliceSet[Elt comparable](elts ...Elt) SliceSet[Elt] { return SliceSet[Elt]{NewSlice(elts...)} }

func (slice SliceSet[Elt]) Has(seek Elt) bool {
	for _, have := range slice.Slice {
		if seek == have {
			return true
		}
	}
	return false
}
