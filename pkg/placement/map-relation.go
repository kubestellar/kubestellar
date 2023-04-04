/*
Copyright 2023 The KCP Authors.

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

// MapRelation2 is a relation represented by two map-based indices.
// It is mutable.
// It is not safe for concurrent access.
type MapRelation2[First comparable, Second comparable] struct {
	by1 map[First]MapSet[Second]
	by2 map[Second]MapSet[First]
}

var _ MutableRelation2[string, float64] = &MapRelation2[string, float64]{}

func NewMapRelation2[First comparable, Second comparable](pairs ...Pair[First, Second]) *MapRelation2[First, Second] {
	return &MapRelation2[First, Second]{
		by1: map[First]MapSet[Second]{},
		by2: map[Second]MapSet[First]{},
	}
}

func (mr MapRelation2[First, Second]) IsEmpty() bool {
	return len(mr.by1) == 0
}

func (mr MapRelation2[First, Second]) LenIsCheap() bool { return false }
func (mr MapRelation2[First, Second]) Len() int         { return Relation2LenFromVisit[First, Second](mr) }

func (mr MapRelation2[First, Second]) Has(tup Pair[First, Second]) bool {
	seconds := mr.by1[tup.First]
	return seconds != nil && seconds.Has(tup.Second)
}

func (mr MapRelation2[First, Second]) Visit(visitor func(Pair[First, Second]) error) error {
	for first, seconds := range mr.by1 {
		if err := seconds.Visit(func(second Second) error {
			return visitor(Pair[First, Second]{first, second})
		}); err != nil {
			return err
		}
	}
	return nil
}

func (mr MapRelation2[First, Second]) Visit1to2(first First, visitor func(Second) error) error {
	seconds := mr.by1[first]
	if seconds != nil {
		return seconds.Visit(visitor)
	}
	return nil
}

func (mr MapRelation2[First, Second]) Visit2to1(second Second, visitor func(First) error) error {
	firsts := mr.by2[second]
	if firsts != nil {
		return firsts.Visit(visitor)
	}
	return nil
}

func (mr MapRelation2[First, Second]) Add(tup Pair[First, Second]) bool {
	if addToIndex(mr.by1, tup.First, tup.Second) {
		return addToIndex(mr.by2, tup.Second, tup.First)
	}
	return false
}

func addToIndex[First, Second comparable](index map[First]MapSet[Second], key First, val Second) bool {
	vals := index[key]
	if vals == nil {
		vals = NewMapSet(val)
		index[key] = vals
		return true
	}
	return vals.Add(val)
}

func (mr MapRelation2[First, Second]) Remove(tup Pair[First, Second]) bool {
	if delFromIndex(mr.by1, tup.First, tup.Second) {
		return delFromIndex(mr.by2, tup.Second, tup.First)
	}
	return false
}

func delFromIndex[First, Second comparable](index map[First]MapSet[Second], key First, val Second) bool {
	vals := index[key]
	if vals == nil {
		return false
	}
	change := vals.Remove(val)
	if vals.IsEmpty() {
		delete(index, key)
	}
	return change
}

func Relation2WithObservers[First, Second comparable](inner MutableRelation2[First, Second], observers ...SetChangeReceiver[Pair[First, Second]]) MutableRelation2[First, Second] {
	return &relation2WithObservers[First, Second]{inner, inner, observers}
}
