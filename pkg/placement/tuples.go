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

type Triple[First comparable, Second comparable, Third comparable] struct {
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

// Rotator is something that can change one value into an equivalent value and back again.
// To understand the name, think of something that can rotate (forward and back) a point in some coordinate system.
type Rotator[Original, Rotated comparable] Pair[func(Original) Rotated, func(Rotated) Original]

func (rr Rotator[Original, Rotated]) Reverse() Rotator[Rotated, Original] {
	return Rotator[Rotated, Original]{rr.Second, rr.First}
}

func NoRotation[Original comparable]() Rotator[Original, Original] {
	return Rotator[Original, Original]{
		First:  Identity1[Original],
		Second: Identity1[Original],
	}
}

// Factorer is a Rotator that converts from some Whole type to some Pair type (and back)
type Factorer[Whole, PartA, PartB comparable] Rotator[Whole, Pair[PartA, PartB]]

func (factoring Factorer[Whole, PartA, PartB]) Factor(whole Whole) (PartA, PartB) {
	ans := factoring.First(whole)
	return ans.First, ans.Second
}

func (factoring Factorer[Whole, PartA, PartB]) Unfactor(partA PartA, partB PartB) Whole {
	return factoring.Second(Pair[PartA, PartB]{partA, partB})
}

func PairFactorer[PartA, PartB comparable]() Factorer[Pair[PartA, PartB], PartA, PartB] {
	return Factorer[Pair[PartA, PartB], PartA, PartB]{
		First:  Identity1[Pair[PartA, PartB]],
		Second: Identity1[Pair[PartA, PartB]],
	}
}

func TripleFactorerTo23and1[ColX, ColY, ColZ comparable]() Factorer[Triple[ColX, ColY, ColZ], Pair[ColY, ColZ], ColX] {
	return Factorer[Triple[ColX, ColY, ColZ], Pair[ColY, ColZ], ColX]{
		First: func(tup Triple[ColX, ColY, ColZ]) Pair[Pair[ColY, ColZ], ColX] {
			return Pair[Pair[ColY, ColZ], ColX]{Pair[ColY, ColZ]{tup.Second, tup.Third}, tup.First}
		},
		Second: func(put Pair[Pair[ColY, ColZ], ColX]) Triple[ColX, ColY, ColZ] {
			return Triple[ColX, ColY, ColZ]{put.Second, put.First.First, put.First.Second}
		},
	}
}
