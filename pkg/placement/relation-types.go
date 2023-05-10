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

// Relation2 is a set of 2-tuples.
type Relation2[First, Second any] interface {
	Set[Pair[First, Second]]
	GetIndex1to2() Index2[First, Second, Set[Second]]
}

type MutableRelation2[First, Second any] interface {
	Relation2[First, Second]
	MutableSet[Pair[First, Second]]
}

type Index2[Key, Val, ValSet any] interface {
	Map[Key, ValSet]
	Visit1to2(Key, func(Val) error) error
}

type MutableIndex2[Key, Val, ValSet any] interface {
	Index2[Key, Val, ValSet]
	Add(Key, Val) bool
	Remove(Key, Val) bool
}

func Relation2WithObservers[First, Second any](inner MutableRelation2[First, Second], observers ...SetWriter[Pair[First, Second]]) MutableRelation2[First, Second] {
	return &relation2WithObservers[First, Second]{inner, inner, observers}
}

type relation2WithObservers[First, Second any] struct {
	Relation2[First, Second]
	inner     MutableRelation2[First, Second]
	observers []SetWriter[Pair[First, Second]]
}

var _ MutableRelation2[int, string] = &relation2WithObservers[int, string]{}

func (rwo *relation2WithObservers[First, Second]) Add(tup Pair[First, Second]) bool {
	if rwo.inner.Add(tup) {
		for _, observer := range rwo.observers {
			observer.Add(tup)
		}
		return true
	}
	return false
}

func (rwo *relation2WithObservers[First, Second]) Remove(tup Pair[First, Second]) bool {
	if rwo.inner.Remove(tup) {
		for _, observer := range rwo.observers {
			observer.Remove(tup)
		}
		return true
	}
	return false
}

// SingleIndexedRelation2 is a 2-ary relation represented by an index on the first column.
// It is mutable.
// It is not safe for concurrent access.
type SingleIndexedRelation2[First, Second any] struct {
	GenericMutableIndexedSet[Pair[First, Second], First, Second, Set[Second]]
}

var _ MutableRelation2[string, float64] = SingleIndexedRelation2[string, float64]{}

// NewSingleIndexedRelation2 constructs a SingleIndexedRelation2.
// The caller supplies the map implementations used in the index.
func NewSingleIndexedRelation2[First, Second any](
	secondSetFactory func(First) MutableSet[Second],
	rep MutableMap[First, MutableSet[Second]],
	pairs ...Pair[First, Second]) SingleIndexedRelation2[First, Second] {
	wholeSet := NewGenericIndexedSet(
		PairFactorer[First, Second](),
		secondSetFactory,
		Identity1[MutableSet[Second]],
		NewSetReadonly[Second],
		rep,
	)
	ans := SingleIndexedRelation2[First, Second]{
		GenericMutableIndexedSet: wholeSet,
	}
	for _, pair := range pairs {
		ans.Add(pair)
	}
	return ans
}

// SingleIndexedRelation3 is a 3-ary relation represented by one nested index that maps
// a First value to an index from Second to Third.
// It is mutable.
// It is not safe for concurrent access.
type SingleIndexedRelation3[First, Second, Third any] struct {
	GenericMutableIndexedSet[Triple[First, Second, Third], First, Pair[Second, Third],
		GenericIndexedSet[Pair[Second, Third], Second, Third, Set[Third]]]
}

// SingleIndexedRelation4 is a 4-ary relation represented by one nested index that maps
// a First value to an index from Second to index from Third to Fourth.
// It is mutable.
// It is not safe for concurrent access.
type SingleIndexedRelation4[First, Second, Third, Fourth any] struct {
	GenericMutableIndexedSet[Quad[First, Second, Third, Fourth], First, Triple[Second, Third, Fourth],
		GenericIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
			GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]]]
}
