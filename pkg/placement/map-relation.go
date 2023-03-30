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

import "errors"

type Relation2[First, Second comparable] interface {
	IsEmpty() bool
	Has(Pair[First, Second]) bool
	Visit(func(Pair[First, Second]) error) error
	Visit1to2(First, func(Second) error) error
	Visit2to1(Second, func(First) error) error
}

type MutableRelation2[First, Second comparable] interface {
	Relation2[First, Second]
	SetChangeReceiver[Pair[First, Second]]
}

var errStop = errors.New("it is done")

func Relation2LessOrEqual[First, Second comparable](reln1, reln2 Relation2[First, Second]) bool {
	return reln1.Visit(func(tup Pair[First, Second]) error {
		if !reln2.Has(tup) {
			return errStop
		}
		return nil
	}) == nil
}

func Relation2Compare[First, Second comparable](reln1, reln2 Relation2[First, Second]) Comparison {
	return Comparison{
		LessOrEqual:    Relation2LessOrEqual(reln1, reln2),
		GreaterOrEqual: Relation2LessOrEqual(reln2, reln1),
	}
}

func Relation2Equal[First, Second comparable](reln1, reln2 Relation2[First, Second]) bool {
	return Relation2Compare[First, Second](reln1, reln2).IsEqual()
}

type Comparison struct{ LessOrEqual, GreaterOrEqual bool }

func (comp Comparison) Reverse() Comparison {
	return Comparison{LessOrEqual: comp.GreaterOrEqual, GreaterOrEqual: comp.LessOrEqual}
}

func (comp Comparison) IsEqual() bool           { return comp.LessOrEqual && comp.GreaterOrEqual }
func (comp Comparison) IsStrictlyLess() bool    { return comp.LessOrEqual && !comp.GreaterOrEqual }
func (comp Comparison) IsStrictlyGreater() bool { return comp.GreaterOrEqual && !comp.LessOrEqual }
func (comp Comparison) IsRelated() bool         { return comp.LessOrEqual || comp.GreaterOrEqual }

func Relation2Reverse[First, Second comparable](forward Relation2[First, Second]) Relation2[Second, First] {
	return ReverseRelation2[First, Second]{baseReverseRelation2[First, Second]{forward}}
}

type ReverseRelation2[First, Second comparable] struct {
	baseReverseRelation2[First, Second]
}

type baseReverseRelation2[First, Second comparable] struct {
	forward Relation2[First, Second]
}

func (rr baseReverseRelation2[First, Second]) IsEmpty() bool {
	return rr.forward.IsEmpty()
}

func (rr baseReverseRelation2[First, Second]) Has(tup Pair[Second, First]) bool {
	return rr.forward.Has(tup.Reverse())
}

func (rr baseReverseRelation2[First, Second]) Visit(visitor func(Pair[Second, First]) error) error {
	return rr.forward.Visit(func(tup Pair[First, Second]) error { return visitor(tup.Reverse()) })
}

func (rr baseReverseRelation2[First, Second]) Visit1to2(second Second, visitor func(First) error) error {
	return rr.forward.Visit2to1(second, visitor)
}

func (rr baseReverseRelation2[First, Second]) Visit2to1(first First, visitor func(Second) error) error {
	return rr.forward.Visit1to2(first, visitor)
}

type ReverseMutableRelation2[First, Second comparable] struct {
	baseReverseRelation2[First, Second]
	forward MutableRelation2[First, Second]
}

func MutableRelation2Reverse[First, Second comparable](forward MutableRelation2[First, Second]) MutableRelation2[Second, First] {
	return ReverseMutableRelation2[First, Second]{baseReverseRelation2[First, Second]{forward}, forward}
}

func (rr ReverseMutableRelation2[First, Second]) Add(tup Pair[Second, First]) bool {
	return rr.forward.Add(tup.Reverse())
}

func (rr ReverseMutableRelation2[First, Second]) Remove(tup Pair[Second, First]) bool {
	return rr.forward.Remove(tup.Reverse())
}

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

type relation2WithObservers[First, Second comparable] struct {
	Relation2[First, Second]
	inner     MutableRelation2[First, Second]
	observers []SetChangeReceiver[Pair[First, Second]]
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
