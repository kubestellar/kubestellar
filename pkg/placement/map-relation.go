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

// NewMapRelation2 constructs a Relation2 that is represented
// by an index on the first column.
// The representation is based on golang `map`s.
func NewMapRelation2[First, Second comparable](pairs ...Pair[First, Second]) SingleIndexedRelation2[First, Second] {
	return NewSingleIndexedRelation2[First, Second](
		func(First) MutableSet[Second] { return NewEmptyMapSet[Second]() },
		NewMapMap[First, MutableSet[Second]](nil),
		pairs...,
	)
}

// NewHashRelation2 constructs a Relation2 that is represented
// by an index on the first column.
// The representation is based on HashMaps.
func NewHashRelation2[First, Second any](hashDomainFirst HashDomain[First], hashDomainSecond HashDomain[Second], pairs ...Pair[First, Second]) SingleIndexedRelation2[First, Second] {
	return NewSingleIndexedRelation2(
		func(First) MutableSet[Second] { return NewHashSet(hashDomainSecond) },
		NewHashMap[First, MutableSet[Second]](hashDomainFirst)(nil),
		pairs...,
	)
}

func NewMapRelation3[First, Second, Third comparable]() SingleIndexedRelation3[First, Second, Third] {
	gis := NewGenericIndexedSet[Triple[First, Second, Third], First, Pair[Second, Third],
		GenericMutableIndexedSet[Pair[Second, Third], Second, Third, Set[Third]],
		GenericIndexedSet[Pair[Second, Third], Second, Third, Set[Third]],
	](
		TripleFactorerTo1and23[First, Second, Third](),
		func(First) GenericMutableIndexedSet[Pair[Second, Third], Second, Third, Set[Third]] {
			return NewGenericIndexedSet[Pair[Second, Third], Second, Third, MapSet[Third], Set[Third]](
				PairFactorer[Second, Third](),
				func(Second) MapSet[Third] { return NewEmptyMapSet[Third]() },
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
	return SingleIndexedRelation3[First, Second, Third]{gis}
}

func NewMapRelation4[First, Second, Third, Fourth comparable]() SingleIndexedRelation4[First, Second, Third, Fourth] {
	gis := NewGenericIndexedSet[Quad[First, Second, Third, Fourth], First, Triple[Second, Third, Fourth],
		GenericMutableIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
			GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]],
		GenericIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
			GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]],
	](
		QuadFactorerTo1and234[First, Second, Third, Fourth](),
		func(First) GenericMutableIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
			GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]] {
			return NewGenericIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
				GenericMutableIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]],
				GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]],
			](
				TripleFactorerTo1and23[Second, Third, Fourth](),
				func(Second) GenericMutableIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]] {
					return NewGenericIndexedSet[Pair[Third, Fourth], Third, Fourth, MapSet[Fourth], Set[Fourth]](
						PairFactorer[Third, Fourth](),
						func(Third) MapSet[Fourth] { return NewEmptyMapSet[Fourth]() },
						func(thirds MapSet[Fourth]) MutableSet[Fourth] { return thirds },
						func(thirds MapSet[Fourth]) Set[Fourth] { return NewSetReadonly[Fourth](thirds) },
						NewMapMap[Third, MapSet[Fourth]](nil),
					)
				},
				func(mutable34 GenericMutableIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]) MutableSet[Pair[Third, Fourth]] {
					return mutable34
				},
				func(mutable34 GenericMutableIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]) GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]] {
					return mutable34.AsReadonly()
				},

				NewMapMap[Second, GenericMutableIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]](nil),
			)
		},
		func(mutable234 GenericMutableIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
			GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]],
		) MutableSet[Triple[Second, Third, Fourth]] {
			return mutable234
		},
		func(mutable234 GenericMutableIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
			GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]],
		) GenericIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
			GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]] {
			return mutable234.AsReadonly()
		},
		NewMapMap[First,
			GenericMutableIndexedSet[Triple[Second, Third, Fourth], Second, Pair[Third, Fourth],
				GenericIndexedSet[Pair[Third, Fourth], Third, Fourth, Set[Fourth]]],
		](nil),
	)
	return SingleIndexedRelation4[First, Second, Third, Fourth]{gis}
}
