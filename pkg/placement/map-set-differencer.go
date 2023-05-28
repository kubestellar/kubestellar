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

func ResolvedWhatAsVisitable(rw WorkloadParts) Visitable[Pair[WorkloadPartID, WorkloadPartDetails]] {
	return MintMapMap(rw, nil)
}

func ResolvedWhereAsVisitable(rw ResolvedWhere) Visitable[SinglePlacement] { return rw }

// func (parts WorkloadParts) Visit(visitor func(WorkloadPart) error) error {
// 	for partID, partDetails := range parts {
// 		part := WorkloadPart{partID, partDetails}
// 		if err := visitor(part); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

func (rw ResolvedWhere) IsEmpty() bool {
	for _, sps := range rw {
		if len(sps.Destinations) > 0 {
			return false
		}
	}
	return true
}

func (rw ResolvedWhere) Len() int {
	var ans int
	for _, sps := range rw {
		ans += len(sps.Destinations)
	}
	return ans
}

func (rw ResolvedWhere) LenIsCheap() bool {
	return true // some day this may be more difficult to answer, but not today
}

func (rw ResolvedWhere) Has(seek SinglePlacement) bool {
	return VisitableHas[SinglePlacement](rw, seek)
}

func (rw ResolvedWhere) Visit(visitor func(SinglePlacement) error) error {
	for _, sps := range rw {
		for _, sp := range sps.Destinations {
			if err := visitor(sp); err != nil {
				return err
			}
		}
	}
	return nil
}

var _ ResolvedWhereDifferencerConstructor = NewResolvedWhereDifferencer
var _ DownsyncsDifferencerConstructor = NewWorkloadPartsDifferencer

func NewResolvedWhereDifferencer(eltChangeReceiver SetWriter[SinglePlacement]) Receiver[ResolvedWhere] {
	return NewSetDifferenceByMapAndEnum(ResolvedWhereAsVisitable, eltChangeReceiver)
}

func NewWorkloadPartsDifferencer(mappingReceiver MapChangeReceiver[WorkloadPartID, WorkloadPartDetails]) Receiver[WorkloadParts] {
	return NewMapDifferenceByMapAndEnum(ResolvedWhatAsVisitable, mappingReceiver)
}

func NewSetDifferenceByMapAndEnum[SetType any, Elt comparable](visitablize func(SetType) Visitable[Elt], eltChangeReceiver SetWriter[Elt]) Receiver[SetType] {
	return setDifferenceByMapAndEnum[SetType, Elt]{
		visitablize:       visitablize,
		eltChangeReceiver: eltChangeReceiver,
		current:           NewMapSet[Elt](),
	}
}

func NewMapDifferenceByMapAndEnum[MapType any, Key, Val comparable](visitablize func(MapType) Visitable[Pair[Key, Val]], mappingChangeReceiver MapChangeReceiver[Key, Val]) Receiver[MapType] {
	return mapDifferenceByMapAndEnum[MapType, Key, Val]{
		visitablize: visitablize,
		current:     NewMapMap(mappingChangeReceiver),
	}
}

type setDifferenceByMapAndEnum[SetType any, Elt comparable] struct {
	visitablize       func(SetType) Visitable[Elt]
	eltChangeReceiver SetWriter[Elt]
	current           MutableSet[Elt]
}

type mapDifferenceByMapAndEnum[MapType any, Key, Val comparable] struct {
	visitablize func(MapType) Visitable[Pair[Key, Val]]
	current     MutableMap[Key, Val]
}

func (dme setDifferenceByMapAndEnum[SetType, Elt]) Receive(nextA SetType) {
	nextS := dme.visitablize(nextA)
	SetUpdateToMatch(dme.current, nextS, dme.eltChangeReceiver)
}

func (dme mapDifferenceByMapAndEnum[MapType, Key, Val]) Receive(nextA MapType) {
	nextS := dme.visitablize(nextA)
	MapUpdateToMatch(dme.current, nextS)
}

func SetUpdateToMatch[Elt comparable](current MutableSet[Elt], target Visitable[Elt], eltChangeReceiver SetWriter[Elt]) {
	goners := MapSetCopy[Elt](current)
	target.Visit(func(newElt Elt) error {
		if current.Add(newElt) {
			eltChangeReceiver.Add(newElt)
		}
		goners.Remove(newElt)
		return nil
	})
	goners.Visit(func(oldElt Elt) error {
		current.Remove(oldElt)
		eltChangeReceiver.Remove(oldElt)
		return nil
	})
}

func MapUpdateToMatch[Key, Val comparable](current MutableMap[Key, Val], target Visitable[Pair[Key, Val]]) {
	goners := MapMapCopy[Key, Val](nil, current)
	target.Visit(func(tup Pair[Key, Val]) error {
		current.Put(tup.First, tup.Second)
		goners.Delete(tup.First)
		return nil
	})
	goners.Visit(func(oldTup Pair[Key, Val]) error {
		current.Delete(oldTup.First)
		return nil
	})
}
