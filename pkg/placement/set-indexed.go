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

// GenericIndexedSet is a readonly set of Tuple whose representation is based on
// factoring each Tuple into Key and Val parts and using a map from
// Key to set of Val.
// Making the Val set type a type parameter allows the use of this in
// contexts where the Val set itself has interesting properties beyond
// just being a set of Val.
// For example, this construction can be thus nested in a way that makes
// the inner structure available.
type GenericIndexedSet[Tuple, Key, Val comparable, ValSet any] interface {
	Set[Tuple]
	GetIndex1to2() Index2[Key, Val, ValSet]
}

// GenericMutableIndexedSet is a GenericIndexedSet that also provides write access.
type GenericMutableIndexedSet[Tuple, Key, Val comparable, ValSet any] interface {
	GenericIndexedSet[Tuple, Key, Val, ValSet]
	MutableSet[Tuple]

	// AsReadonly returns a view that does not support writes
	AsReadonly() GenericIndexedSet[Tuple, Key, Val, ValSet]
}

// NewGenericIndexedSet constructs an index given the constituent functionality.
// The returned value implements GenericMutableIndexedSet.
func NewGenericIndexedSet[Tuple, Key, Val comparable, ValMutableSet, ValSet any](
	factoring Factorer[Tuple, Key, Val],
	valSetFactory func() ValMutableSet,
	valMutableSetAsSet func(ValMutableSet) MutableSet[Val],
	insulateValSet func(ValMutableSet) ValSet,
	rep MutableMap[Key, ValMutableSet],
) *genericMutableIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet] {
	return &genericMutableIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]{
		genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]{
			// setObserver:   setObserver,
			factoring:          factoring,
			valSetFactory:      valSetFactory,
			valMutableSetAsSet: valMutableSetAsSet,
			insulateValSet:     insulateValSet,
			rep:                rep,
		}}
}

var _ GenericMutableIndexedSet[complex64, int, string, float32] = &genericMutableIndexedSet[complex64, int, string, byte, float32]{}

func GenericMutableIndexedSetToReadonly[Tuple, Key, Val comparable, ValMutableSet, ValSet any](gi *genericMutableIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) GenericIndexedSet[Tuple, Key, Val, ValSet] {
	return &gi.genericIndexedSet
}

func (gi *genericMutableIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) AsReadonly() GenericIndexedSet[Tuple, Key, Val, ValSet] {
	return &gi.genericIndexedSet
}

type genericIndexedSet[Tuple, Key, Val comparable, ValMutableSet, ValSet any] struct {
	factoring          Factorer[Tuple, Key, Val]
	valSetFactory      func() ValMutableSet
	valMutableSetAsSet func(ValMutableSet) MutableSet[Val]
	insulateValSet     func(ValMutableSet) ValSet
	rep                MutableMap[Key, ValMutableSet]
}

type genericMutableIndexedSet[Tuple, Key, Val comparable, ValMutableSet, ValSet any] struct {
	genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]
}

func (gi *genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) IsEmpty() bool {
	return gi.rep.IsEmpty()
}
func (gi *genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) LenIsCheap() bool { return false }

func (gi *genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) Len() int {
	var ans int
	gi.rep.Visit(func(tup Pair[Key, ValMutableSet]) error {
		vals := gi.valMutableSetAsSet(tup.Second)
		ans += vals.Len()
		return nil
	})
	return ans
}

func (gi *genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) Has(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
	valSet, ok := gi.rep.Get(key)
	if !ok {
		return false
	}
	vals := gi.valMutableSetAsSet(valSet)
	return vals.Has(val)
}

func (gi *genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) Visit(visitor func(Tuple) error) error {
	return gi.rep.Visit(func(tup Pair[Key, ValMutableSet]) error {
		vals := gi.valMutableSetAsSet(tup.Second)
		return vals.Visit(func(val Val) error {
			return visitor(gi.factoring.Second(Pair[Key, Val]{tup.First, val}))
		})
	})
}

func (gi *genericMutableIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) Add(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
	valSet, ok := gi.rep.Get(key)
	if !ok {
		valSet = gi.valSetFactory()
		gi.rep.Put(key, valSet)
	}
	vals := gi.valMutableSetAsSet(valSet)
	return vals.Add(val)
}

func (gi *genericMutableIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) Remove(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
	valSet, ok := gi.rep.Get(key)
	if !ok {
		return false
	}
	vals := gi.valMutableSetAsSet(valSet)
	if !vals.Remove(val) {
		return false
	}
	if vals.IsEmpty() {
		gi.rep.Delete(key)
	}
	return true
}

func (gi *genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]) GetIndex1to2() Index2[Key, Val, ValSet] {
	return genericIndexedSetIndex[Tuple, Key, Val, ValMutableSet, ValSet]{gi}
}

type genericIndexedSetIndex[Tuple, Key, Val comparable, ValMutableSet, ValSet any] struct {
	*genericIndexedSet[Tuple, Key, Val, ValMutableSet, ValSet]
}

var _ Index2[int, string, MapSet[string]] = genericIndexedSetIndex[complex64, int, string, MapSet[int], MapSet[string]]{}

func (gi genericIndexedSetIndex[Tuple, Key, Val, ValMutableSet, ValSet]) Get(key Key) (ValSet, bool) {
	set, ok := gi.rep.Get(key)
	return gi.insulateValSet(set), ok
}

func (mi genericIndexedSetIndex[Tuple, Key, Val, ValMutableSet, ValSet]) Visit(visitor func(Pair[Key, ValSet]) error) error {
	return mi.rep.Visit(Func11Compose11(mi.insulateSeconds, visitor))
}

func (mi genericIndexedSetIndex[Tuple, Key, Val, ValMutableSet, ValSet]) insulateSeconds(tup Pair[Key, ValMutableSet]) Pair[Key, ValSet] {
	return NewPair(tup.First, mi.insulateValSet(tup.Second))
}

func (mi genericIndexedSetIndex[Tuple, Key, Val, ValMutableSet, ValSet]) Visit1to2(key Key, visitor func(Val) error) error {
	valSet, ok := mi.rep.Get(key)
	if ok {
		vals := mi.valMutableSetAsSet(valSet)
		return vals.Visit(visitor)
	}
	return nil
}
