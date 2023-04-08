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

// MapRelation2 is a 2-ary relation represented by an index on the first column.
// It is mutable.
// It is not safe for concurrent access.
type MapRelation2[First comparable, Second comparable] struct {
	GenericSetIndex[Pair[First, Second], First, Second]
}

var _ MutableRelation2[string, float64] = &MapRelation2[string, float64]{}

// NewGenericRelation2Index constructs a Relation2 that is represented
// by an index on the first column.
// The representation is based on golang `map`s.
func NewMapRelation2[First, Second comparable](pairs ...Pair[First, Second]) *MapRelation2[First, Second] {
	return NewGenericRelation2Index[First, Second](
		func() MutableSet[Second] { return NewEmptyMapSet[Second]() },
		NewMapMap[First, MutableSet[Second]](nil),
		pairs...,
	)
}

// NewGenericRelation3Index constructs a set of triples
// that is represented by two layers of indexing.
// The representation is based on golang `map`s.
func NewMapRelation3Index[First, Second, Third comparable]() *MapRelation2[First, Pair[Second, Third]] {
	return NewGenericRelation3Index[First, Second, Third](
		func() MutableSet[Third] { return NewEmptyMapSet[Third]() },
		func() MutableMap[Second, MutableSet[Third]] {
			return NewMapMap[Second, MutableSet[Third]](nil)
		},
		NewMapMap[First, MutableSet[Pair[Second, Third]]](nil))
}
