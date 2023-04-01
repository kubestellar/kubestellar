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
	"fmt"
	"strings"
)

type Relation2[First, Second comparable] interface {
	Set[Pair[First, Second]]
	Visit1to2(First, func(Second) error) error
	Visit2to1(Second, func(First) error) error
}

type MutableRelation2[First, Second comparable] interface {
	Relation2[First, Second]
	MutableSet[Pair[First, Second]]
}

type Pair[First any, Second any] struct {
	First  First
	Second Second
}

func (tup Pair[First, Second]) String() string {
	var ans strings.Builder
	ans.WriteRune('(')
	ans.WriteString(fmt.Sprintf("%v", tup.First))
	ans.WriteString(", ")
	ans.WriteString(fmt.Sprintf("%v", tup.Second))
	ans.WriteRune(')')
	return ans.String()
}

func (tup Pair[First, Second]) Reverse() Pair[Second, First] {
	return Pair[Second, First]{First: tup.Second, Second: tup.First}
}

func AddFirstFunc[First any, Second any](first First) func(Second) Pair[First, Second] {
	return func(second Second) Pair[First, Second] {
		return Pair[First, Second]{First: first, Second: second}
	}
}

func AddSecondFunc[First any, Second any](second Second) func(First) Pair[First, Second] {
	return func(first First) Pair[First, Second] {
		return Pair[First, Second]{First: first, Second: second}
	}
}

type Triple[First any, Second any, Third any] struct {
	First  First
	Second Second
	Third  Third
}

func (tup Triple[First, Second, Third]) String() string {
	var ans strings.Builder
	ans.WriteRune('(')
	ans.WriteString(fmt.Sprintf("%v", tup.First))
	ans.WriteString(", ")
	ans.WriteString(fmt.Sprintf("%v", tup.Second))
	ans.WriteString(", ")
	ans.WriteString(fmt.Sprintf("%v", tup.Third))
	ans.WriteRune(')')
	return ans.String()
}

func Relation2LenFromVisit[First, Second comparable](reln Relation2[First, Second]) int {
	var ans int
	reln.Visit(func(_ Pair[First, Second]) error {
		ans++
		return nil
	})
	return ans
}

func Relation2Reverse[First, Second comparable](forward Relation2[First, Second]) Relation2[Second, First] {
	return ReverseRelation2[First, Second]{baseReverseRelation2[First, Second]{forward}}
}

var _ Relation2[int, string] = ReverseRelation2[string, int]{}

type ReverseRelation2[First, Second comparable] struct {
	baseReverseRelation2[First, Second]
}

type baseReverseRelation2[First, Second comparable] struct {
	forward Relation2[First, Second]
}

func (rr baseReverseRelation2[First, Second]) IsEmpty() bool {
	return rr.forward.IsEmpty()
}

func (rr baseReverseRelation2[First, Second]) LenIsCheap() bool {
	return rr.forward.LenIsCheap()
}

func (rr baseReverseRelation2[First, Second]) Len() int {
	return rr.forward.Len()
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
