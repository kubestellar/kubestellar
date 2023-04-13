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

// This file contains some generic non-uniform (the members are not necessarily all the same type) tuple types
// with useful methods.

type Empty struct{}

type Pair[First, Second any] struct {
	First  First
	Second Second
}

func NewPair[First, Second any](first First, second Second) Pair[First, Second] {
	return Pair[First, Second]{first, second}
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

func PairReverse[First, Second comparable](forward Pair[First, Second]) Pair[Second, First] {
	return Pair[Second, First]{forward.Second, forward.First}
}

func NewPair1Then2[First any, Second any](first First) func(Second) Pair[First, Second] {
	return func(second Second) Pair[First, Second] {
		return Pair[First, Second]{First: first, Second: second}
	}
}

func NewPair2Then1[First any, Second any](second Second) func(First) Pair[First, Second] {
	return func(first First) Pair[First, Second] {
		return Pair[First, Second]{First: first, Second: second}
	}
}

type Triple[First, Second, Third any] struct {
	First  First
	Second Second
	Third  Third
}

func NewTriple[First, Second, Third any](first First, second Second, third Third) Triple[First, Second, Third] {
	return Triple[First, Second, Third]{first, second, third}
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

func (tup Triple[PartA, PartB, PartC]) Reverse() Triple[PartC, PartB, PartA] {
	return NewTriple(tup.Third, tup.Second, tup.First)
}

// Rotator is something that can change one value into an equivalent value and back again.
// To understand the name, think of something that can rotate (forward and back) a point in some coordinate system.
type Rotator[Original, Rotated any] Pair[func(Original) Rotated, func(Rotated) Original]

func (rr Rotator[Original, Rotated]) Reverse() Rotator[Rotated, Original] {
	return Rotator[Rotated, Original]{rr.Second, rr.First}
}

func NewRotator[Original, Rotated any](forward func(Original) Rotated, reverse func(Rotated) Original) Rotator[Original, Rotated] {
	return Rotator[Original, Rotated](NewPair(forward, reverse))
}

func NoRotation[Original any]() Rotator[Original, Original] {
	return Rotator[Original, Original]{
		First:  Identity1[Original],
		Second: Identity1[Original],
	}
}

// Factorer is a Rotator that converts from some Whole type to some Pair type (and back)
type Factorer[Whole, PartA, PartB any] Rotator[Whole, Pair[PartA, PartB]]

func NewFactorer[Whole, PartA, PartB any](forward func(Whole) Pair[PartA, PartB], reverse func(Pair[PartA, PartB]) Whole) Factorer[Whole, PartA, PartB] {
	return Factorer[Whole, PartA, PartB]{forward, reverse}
}

func (factoring Factorer[Whole, PartA, PartB]) Factor(whole Whole) (PartA, PartB) {
	ans := factoring.First(whole)
	return ans.First, ans.Second
}

func (factoring Factorer[Whole, PartA, PartB]) Unfactor(partA PartA, partB PartB) Whole {
	return factoring.Second(Pair[PartA, PartB]{partA, partB})
}

func PairReverser[PartA, PartB any]() Rotator[Pair[PartA, PartB], Pair[PartB, PartA]] {
	return NewRotator(Pair[PartA, PartB].Reverse, Pair[PartB, PartA].Reverse)
}

func PairFactorer[PartA, PartB any]() Factorer[Pair[PartA, PartB], PartA, PartB] {
	return Factorer[Pair[PartA, PartB], PartA, PartB]{
		First:  Identity1[Pair[PartA, PartB]],
		Second: Identity1[Pair[PartA, PartB]],
	}
}

func TripleReverser[PartA, PartB, PartC any]() Rotator[Triple[PartA, PartB, PartC], Triple[PartC, PartB, PartA]] {
	return NewRotator(Triple[PartA, PartB, PartC].Reverse, Triple[PartC, PartB, PartA].Reverse)
}

func TripleFactorerTo23and1[ColX, ColY, ColZ any]() Factorer[Triple[ColX, ColY, ColZ], Pair[ColY, ColZ], ColX] {
	return Factorer[Triple[ColX, ColY, ColZ], Pair[ColY, ColZ], ColX]{
		First: func(tup Triple[ColX, ColY, ColZ]) Pair[Pair[ColY, ColZ], ColX] {
			return Pair[Pair[ColY, ColZ], ColX]{Pair[ColY, ColZ]{tup.Second, tup.Third}, tup.First}
		},
		Second: func(put Pair[Pair[ColY, ColZ], ColX]) Triple[ColX, ColY, ColZ] {
			return Triple[ColX, ColY, ColZ]{put.Second, put.First.First, put.First.Second}
		},
	}
}

func TripleFactorerTo13and2[ColX, ColY, ColZ any]() Factorer[Triple[ColX, ColY, ColZ], Pair[ColX, ColZ], ColY] {
	return NewFactorer(
		func(tup Triple[ColX, ColY, ColZ]) Pair[Pair[ColX, ColZ], ColY] {
			return NewPair(NewPair(tup.First, tup.Third), tup.Second)
		},
		func(put Pair[Pair[ColX, ColZ], ColY]) Triple[ColX, ColY, ColZ] {
			return NewTriple(put.First.First, put.Second, put.First.Second)
		})
}

func TripleFactorerTo1and23[ColX, ColY, ColZ any]() Factorer[Triple[ColX, ColY, ColZ], ColX, Pair[ColY, ColZ]] {
	return NewFactorer(
		func(tup Triple[ColX, ColY, ColZ]) Pair[ColX, Pair[ColY, ColZ]] {
			return NewPair(tup.First, NewPair(tup.Second, tup.Third))
		},
		func(put Pair[ColX, Pair[ColY, ColZ]]) Triple[ColX, ColY, ColZ] {
			return NewTriple(put.First, put.Second.First, put.Second.Second)
		})
}

type Quad[First, Second, Third, Fourth any] struct {
	First  First
	Second Second
	Third  Third
	Fourth Fourth
}

func NewQuad[First, Second, Third, Fourth any](first First, second Second, third Third, fourth Fourth) Quad[First, Second, Third, Fourth] {
	return Quad[First, Second, Third, Fourth]{first, second, third, fourth}
}

func QuadFactorerTo1and234[ColW, ColX, ColY, ColZ any]() Factorer[Quad[ColW, ColX, ColY, ColZ], ColW, Triple[ColX, ColY, ColZ]] {
	return NewFactorer(
		func(tup Quad[ColW, ColX, ColY, ColZ]) Pair[ColW, Triple[ColX, ColY, ColZ]] {
			return NewPair(tup.First, NewTriple(tup.Second, tup.Third, tup.Fourth))
		},
		func(put Pair[ColW, Triple[ColX, ColY, ColZ]]) Quad[ColW, ColX, ColY, ColZ] {
			return NewQuad(put.First, put.Second.First, put.Second.Second, put.Second.Third)
		})
}
