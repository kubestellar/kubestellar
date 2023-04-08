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

// NewGenericIndexedSet constructs an index given the constituent functionality and,
// optionally, an observer.
func NewGenericIndexedSet[Tuple, Key, Val comparable](
	// setObserver MappingReceiver[Key, Set[Val]],
	factoring Factorer[Tuple, Key, Val],
	valSetFactory func() MutableSet[Val],
	rep MutableMap[Key, MutableSet[Val]],
) GenericIndexedSet[Tuple, Key, Val] {
	return &genericIndexedSet[Tuple, Key, Val]{
		// setObserver:   setObserver,
		factoring:     factoring,
		valSetFactory: valSetFactory,
		rep:           rep,
	}
}

type GenericIndexedSet[Tuple, Key, Val comparable] interface {
	MutableSet[Tuple]
	GetIndex1to2() Index2[Key, Val]
}

type genericIndexedSet[Tuple, Key, Val comparable] struct {
	setObserver   MappingReceiver[Key, Set[Val]]
	factoring     Factorer[Tuple, Key, Val]
	valSetFactory func() MutableSet[Val]
	rep           MutableMap[Key, MutableSet[Val]]
}

func (gi *genericIndexedSet[Tuple, Key, Val]) IsEmpty() bool    { return gi.rep.IsEmpty() }
func (gi *genericIndexedSet[Tuple, Key, Val]) LenIsCheap() bool { return false }

func (gi *genericIndexedSet[Tuple, Key, Val]) Len() int {
	var ans int
	gi.rep.Visit(func(tup Pair[Key, MutableSet[Val]]) error {
		ans += tup.Second.Len()
		return nil
	})
	return ans
}

func (gi *genericIndexedSet[Tuple, Key, Val]) Has(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
	vals, ok := gi.rep.Get(key)
	if !ok {
		return false
	}
	return vals.Has(val)
}

func (gi *genericIndexedSet[Tuple, Key, Val]) Visit(visitor func(Tuple) error) error {
	return gi.rep.Visit(func(tup Pair[Key, MutableSet[Val]]) error {
		return tup.Second.Visit(func(val Val) error {
			return visitor(gi.factoring.Second(Pair[Key, Val]{tup.First, val}))
		})
	})
}

func (gi *genericIndexedSet[Tuple, Key, Val]) Add(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
	return genericIndexedSetIndex[Tuple, Key, Val]{gi}.AddX(key, val)
}

func (gi *genericIndexedSet[Tuple, Key, Val]) Remove(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
	return genericIndexedSetIndex[Tuple, Key, Val]{gi}.RemoveX(key, val)
}

func (gi *genericIndexedSet[Tuple, Key, Val]) GetIndex1to2() Index2[Key, Val] {
	return genericIndexedSetIndex[Tuple, Key, Val]{gi}
}

type genericIndexedSetIndex[Tuple, Key, Val comparable] struct {
	*genericIndexedSet[Tuple, Key, Val]
}

var _ Index2[int, string] = genericIndexedSetIndex[complex64, int, string]{}

func (mi genericIndexedSetIndex[Tuple, Key, Val]) Get(key Key) (Set[Val], bool) {
	set, ok := mi.rep.Get(key)
	return SetReadonly[Val]{set}, ok
}

func (mi genericIndexedSetIndex[Tuple, Key, Val]) Visit(visitor func(Pair[Key, Set[Val]]) error) error {
	return mi.rep.Visit(Func11Compose11(mi.insulateSeconds, visitor))
}

func (mi genericIndexedSetIndex[Tuple, Key, Val]) insulateSeconds(tup Pair[Key, MutableSet[Val]]) Pair[Key, Set[Val]] {
	return Pair[Key, Set[Val]]{tup.First, SetReadonly[Val]{tup.Second}}
}

func (mi genericIndexedSetIndex[Tuple, Key, Val]) Visit1to2(key Key, visitor func(Val) error) error {
	vals, ok := mi.Get(key)
	if ok {
		return vals.Visit(visitor)
	}
	return nil
}

func (gi genericIndexedSetIndex[Tuple, Key, Val]) AddX(key Key, val Val) bool {
	vals, ok := gi.rep.Get(key)
	if !ok {
		vals = gi.valSetFactory()
		gi.rep.Put(key, vals)
	}
	if !vals.Add(val) {
		return false
	}
	if gi.setObserver != nil {
		gi.setObserver.Put(key, vals)
	}
	return true
}

func (gi genericIndexedSetIndex[Tuple, Key, Val]) RemoveX(key Key, val Val) bool {
	vals, ok := gi.rep.Get(key)
	if !ok {
		return false
	}
	if !vals.Remove(val) {
		return false
	}
	if vals.IsEmpty() {
		gi.rep.Delete(key)
		if gi.setObserver != nil {
			gi.setObserver.Delete(key)
		}
	} else if gi.setObserver != nil {
		gi.setObserver.Put(key, vals)
	}
	return true
}
