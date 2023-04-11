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
type MapRelation2[First, Second any] struct {
	GenericMutableIndexedSet[Pair[First, Second], First, Second, Set[Second]]
}

var _ MutableRelation2[string, float64] = &MapRelation2[string, float64]{}

// NewMapRelation2 constructs a Relation2 that is represented
// by an index on the first column.
// The representation is based on golang `map`s.
func NewMapRelation2[First, Second comparable](pairs ...Pair[First, Second]) *MapRelation2[First, Second] {
	return NewGenericRelation2Index[First, Second](
		func() MutableSet[Second] { return NewEmptyMapSet[Second]() },
		NewMapMap[First, MutableSet[Second]](nil),
		pairs...,
	)
}

// NewHashRelation2 constructs a Relation2 that is represented
// by an index on the first column.
// The representation is based on HashMaps.
func NewHashRelation2[First, Second any](hashDomainFirst HashDomain[First], hashDomainSecond HashDomain[Second], pairs ...Pair[First, Second]) *MapRelation2[First, Second] {
	return NewGenericRelation2Index[First, Second](
		func() MutableSet[Second] { return NewHashSet(hashDomainSecond) },
		NewHashMap[First, MutableSet[Second]](hashDomainFirst)(nil),
		pairs...,
	)
}

type MapRelation3[First, Second, Third comparable] struct {
	GenericMutableIndexedSet[Triple[First, Second, Third], First, Pair[Second, Third],
		GenericIndexedSet[Pair[Second, Third], Second, Third, Set[Third]]]
}

func NewMapRelation3[First, Second, Third comparable]() MapRelation3[First, Second, Third] {
	gis := NewGenericIndexedSet[Triple[First, Second, Third], First, Pair[Second, Third],
		GenericMutableIndexedSet[Pair[Second, Third], Second, Third, Set[Third]],
		GenericIndexedSet[Pair[Second, Third], Second, Third, Set[Third]],
	](
		TripleFactorerTo1and23[First, Second, Third](),
		func() GenericMutableIndexedSet[Pair[Second, Third], Second, Third, Set[Third]] {
			return NewGenericIndexedSet[Pair[Second, Third], Second, Third, MapSet[Third], Set[Third]](
				PairFactorer[Second, Third](),
				NewEmptyMapSet[Third],
				func(thirds MapSet[Third]) MutableSet[Third] { return thirds },
				func(thirds MapSet[Third]) Set[Third] { return NewSetReadonly[Third](thirds) },
				NewMapMap[Second, MapSet[Third]](nil),
			)
		},
		func(mutable23 GenericMutableIndexedSet[Pair[Second, Third], Second, Third, Set[Third]]) MutableSet[Pair[Second, Third]] {
			return mutable23
		},
		func(mutable23 GenericMutableIndexedSet[Pair[Second, Third], Second, Third, Set[Third]]) GenericIndexedSet[Pair[Second, Third], Second, Third, Set[Third]] {
			return mutable23.AsReadonly()
		},
		NewMapMap[First, GenericMutableIndexedSet[Pair[Second, Third], Second, Third, Set[Third]]](nil),
	)
	return MapRelation3[First, Second, Third]{gis}
}

func (mr MapRelation3[First, Second, Third]) Get1to2to3(first First) Index2[Second, Third, Set[Third]] {
	inner, has := mr.GenericMutableIndexedSet.GetIndex1to2().Get(first)
	if !has {
		return nil
	}
	return inner.GetIndex1to2()
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
