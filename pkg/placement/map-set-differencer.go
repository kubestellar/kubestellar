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

import (
	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func ResolvedWhatAsVisitable(rw WorkloadParts) Visitable[WorkloadPart]             { return rw }
func ResolvedWhereAsVisitable(rw ResolvedWhere) Visitable[edgeapi.SinglePlacement] { return rw }

func (parts WorkloadParts) Visit(visitor func(WorkloadPart) error) error {
	for partID, partDetails := range parts {
		part := WorkloadPart{partID, partDetails}
		if err := visitor(part); err != nil {
			return err
		}
	}
	return nil
}

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

func (rw ResolvedWhere) Has(seek edgeapi.SinglePlacement) bool {
	return VisitableHas[edgeapi.SinglePlacement](rw, seek)
}

func (rw ResolvedWhere) Visit(visitor func(edgeapi.SinglePlacement) error) error {
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
var _ ResolvedWhatDifferencerConstructor = NewResolvedWhatDifferencer

func NewResolvedWhereDifferencer(eltChangeReceiver SetChangeReceiver[edgeapi.SinglePlacement]) Receiver[ResolvedWhere] {
	return NewDifferenceByMapAndEnum[ResolvedWhere, edgeapi.SinglePlacement](ResolvedWhereAsVisitable, eltChangeReceiver)
}

func NewResolvedWhatDifferencer(eltChangeReceiver SetChangeReceiver[WorkloadPart]) Receiver[WorkloadParts] {
	return NewDifferenceByMapAndEnum[WorkloadParts, WorkloadPart](ResolvedWhatAsVisitable, eltChangeReceiver)
}

func NewDifferenceByMapAndEnum[SetType any, Elt comparable](visitablize func(SetType) Visitable[Elt], eltChangeReceiver SetChangeReceiver[Elt]) Receiver[SetType] {
	return differenceByMapAndEnum[SetType, Elt]{
		visitablize:       visitablize,
		eltChangeReceiver: eltChangeReceiver,
		current:           NewMapSet[Elt](),
	}
}

type differenceByMapAndEnum[SetType any, Elt comparable] struct {
	visitablize       func(SetType) Visitable[Elt]
	eltChangeReceiver SetChangeReceiver[Elt]
	current           MutableSet[Elt]
}

func (dme differenceByMapAndEnum[SetType, Elt]) Receive(nextA SetType) {
	nextS := dme.visitablize(nextA)
	SetUpdateToMatch[Elt](dme.current, nextS, dme.eltChangeReceiver)
}

func SetUpdateToMatch[Elt comparable](current MutableSet[Elt], target Visitable[Elt], eltChangeReceiver SetChangeReceiver[Elt]) {
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
