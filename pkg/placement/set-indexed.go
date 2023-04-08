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

// NewGenericSetIndex constructs an index given the constituent functionality and,
// optionally, an observer.
// TODO: add the inverse of factor and make it implement Index2
func NewGenericSetIndex[Tuple, Key, Val comparable](
	setObserver MappingReceiver[Key, Set[Val]],
	factoring Factorer[Tuple, Key, Val],
	valSetFactory func() MutableSet[Val],
	rep MutableMap[Key, MutableSet[Val]],
) GenericSetIndex[Tuple, Key, Val] {
	return &genericIndex[Tuple, Key, Val]{
		setObserver:   setObserver,
		factoring:     factoring,
		valSetFactory: valSetFactory,
		rep:           rep,
	}
}

type GenericSetIndex[Tuple, Key, Val comparable] interface {
	MutableSet[Tuple]
	GetIndex1to2() Index2[Key, Val]
}

type genericIndex[Tuple, Key, Val comparable] struct {
	setObserver   MappingReceiver[Key, Set[Val]]
	factoring     Factorer[Tuple, Key, Val]
	valSetFactory func() MutableSet[Val]
	rep           MutableMap[Key, MutableSet[Val]]
}

func (gi *genericIndex[Tuple, Key, Val]) IsEmpty() bool    { return gi.rep.IsEmpty() }
func (gi *genericIndex[Tuple, Key, Val]) LenIsCheap() bool { return false }

func (gi *genericIndex[Tuple, Key, Val]) Len() int {
	var ans int
	gi.rep.Visit(func(tup Pair[Key, MutableSet[Val]]) error {
		ans += tup.Second.Len()
		return nil
	})
	return ans
}

func (gi *genericIndex[Tuple, Key, Val]) Has(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
	vals, ok := gi.rep.Get(key)
	if !ok {
		return false
	}
	return vals.Has(val)
}

func (gi *genericIndex[Tuple, Key, Val]) Visit(visitor func(Tuple) error) error {
	return gi.rep.Visit(func(tup Pair[Key, MutableSet[Val]]) error {
		return tup.Second.Visit(func(val Val) error {
			return visitor(gi.factoring.Second(Pair[Key, Val]{tup.First, val}))
		})
	})
}

func (gi *genericIndex[Tuple, Key, Val]) Add(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
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

func (gi *genericIndex[Tuple, Key, Val]) Remove(tup Tuple) bool {
	key, val := gi.factoring.Factor(tup)
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

func (gi *genericIndex[Tuple, Key, Val]) GetIndex1to2() Index2[Key, Val] {
	return SetIndex2[Key, Val]{gi.rep, gi.valSetFactory}
}

type SetIndex2[First, Second comparable] struct {
	MutableMap[First, MutableSet[Second]]
	secondSetFactory func() MutableSet[Second]
}

var _ MutableIndex2[int, string] = SetIndex2[int, string]{}

func (mi SetIndex2[First, Second]) Get(key First) (Set[Second], bool) {
	set, ok := mi.MutableMap.Get(key)
	return SetReadonly[Second]{set}, ok
}

func (mi SetIndex2[First, Second]) Visit(visitor func(Pair[First, Set[Second]]) error) error {
	return mi.MutableMap.Visit(Func11Compose11(mi.insulateSeconds, visitor))
}

func (mi SetIndex2[First, Second]) insulateSeconds(tup Pair[First, MutableSet[Second]]) Pair[First, Set[Second]] {
	return Pair[First, Set[Second]]{tup.First, SetReadonly[Second]{tup.Second}}
}

func (mi SetIndex2[First, Second]) Visit1to2(first First, visitor func(Second) error) error {
	seconds, ok := mi.Get(first)
	if ok {
		return seconds.Visit(visitor)
	}
	return nil
}

func (mi SetIndex2[First, Second]) Add(key First, val Second) bool {
	vals, ok := mi.MutableMap.Get(key)
	if !ok {
		vals = mi.secondSetFactory()
		mi.MutableMap.Put(key, vals)
		return true
	}
	return vals.Add(val)
}

func (mi SetIndex2[First, Second]) Remove(key First, val Second) bool {
	vals, ok := mi.MutableMap.Get(key)
	if !ok {
		return false
	}
	change := vals.Remove(val)
	if vals.IsEmpty() {
		mi.MutableMap.Delete(key)
	}
	return change
}
