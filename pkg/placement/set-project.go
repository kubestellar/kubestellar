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
// This buffers the set in an index.
// The booleans returned by partAReceiver are ignored.
func NewSetChangeProjector[Whole, PartA, PartB comparable](
	factoring Factorer[Whole, PartA, PartB],
	partAReceiver SetChangeReceiver[PartA],
) SetChangeReceiver[Whole] {
	// indexerRep ignores the set of PartB and notifies partAReceiver of PartA set change
	indexerRep := NewMapMap[PartA, MutableSet[PartB]](MapKeySetReceiver[PartA, MutableSet[PartB]](partAReceiver))
	indexer := NewGenericIndexedSet[Whole, PartA, PartB](
		// nil,
		factoring,
		func() MutableSet[PartB] { return NewEmptyMapSet[PartB]() },
		indexerRep,
	)
	return indexer
}
