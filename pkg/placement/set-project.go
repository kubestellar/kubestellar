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

// NewSetChangeProjector transforms a receiver of PartA into a receiver of Whole,
// given a Factorer of Whole into PartA and PartB.
// This buffers the set in a MapMap used as an index.
// The booleans returned by partAReceiver are ignored.
func NewSetChangeProjectorByMapMap[Whole any, PartA, PartB comparable](
	factoring Factorer[Whole, PartA, PartB],
	partAReceiver SetWriter[PartA],
) SetWriter[Whole] {
	return NewSetChangeProjector[Whole, PartA, PartB](
		factoring,
		partAReceiver,
		func(observer MapChangeReceiver[PartA, MutableSet[PartB]]) MutableMap[PartA, MutableSet[PartB]] {
			return NewMapMap[PartA, MutableSet[PartB]](observer)
		},
		func(PartA) MutableSet[PartB] { return NewEmptyMapSet[PartB]() },
	)
}

func NewSetChangeProjectorByHashMap[Whole, PartA, PartB any](
	factoring Factorer[Whole, PartA, PartB],
	partAReceiver SetWriter[PartA],
	hashDomainA HashDomain[PartA],
	hashDomainB HashDomain[PartB],
) SetWriter[Whole] {
	return NewSetChangeProjector[Whole, PartA, PartB](
		factoring,
		partAReceiver,
		func(observer MapChangeReceiver[PartA, MutableSet[PartB]]) MutableMap[PartA, MutableSet[PartB]] {
			return NewHashMap[PartA, MutableSet[PartB]](hashDomainA)(observer)
		},
		func(PartA) MutableSet[PartB] { return NewHashSet[PartB](hashDomainB) },
	)
}

// NewSetChangeProjector transforms a receiver of PartA into a receiver of Whole,
// given a Factorer of Whole into PartA and PartB.
// This buffers the set in an index created by the given maker.
// The booleans returned by partAReceiver are ignored.
func NewSetChangeProjector[Whole, PartA, PartB any](
	factoring Factorer[Whole, PartA, PartB],
	partAReceiver SetWriter[PartA],
	repMaker func(MapChangeReceiver[PartA, MutableSet[PartB]]) MutableMap[PartA, MutableSet[PartB]],
	innerSetFactory func(PartA) MutableSet[PartB],
) SetWriter[Whole] {
	// indexerRep ignores the set of PartB and notifies partAReceiver of PartA set change
	indexerRep := repMaker(MapKeySetReceiver[PartA, MutableSet[PartB]](partAReceiver))
	indexer := NewGenericIndexedSet[Whole, PartA, PartB, MutableSet[PartB], Set[PartB]](
		// nil,
		factoring,
		innerSetFactory,
		Identity1[MutableSet[PartB]],
		NewSetReadonly[PartB],
		indexerRep,
	)
	return indexer
}
