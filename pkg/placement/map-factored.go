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

// FactoredMap is a mutable map implemented by two levels of mapping.
// It has a general key decomposition/composition rotator.
type FactoredMap[WholeKey, KeyPartA, KeyPartB, Val any] GenericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, Map[KeyPartB, Val]]

// GenericFactoredMap is a mutable map implemented by two levels of mapping.
// It has a general key decomposition/composition rotator
// and a general representation for the nested maps.
type GenericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, InnerMap any] interface {
	MutableMap[WholeKey, Val]
	GetIndex() GenericFactoredMapIndex[KeyPartA, KeyPartB, Val, InnerMap]
}

// GenericFactoredMapIndex is an index into a factored map where the inner maps
// (from KeyPartB to Val) have a general representation.
type GenericFactoredMapIndex[KeyPartA, KeyPartB, Val, InnerMap any] interface {
	Map[KeyPartA, InnerMap]
	Visit1to2(KeyPartA, func(Pair[KeyPartB, Val]) error) error
}

// NewFactoredMapMap makes a FactoredMap implements by golang maps.
func NewFactoredMapMap[WholeKey, KeyPartA, KeyPartB comparable, Val any](
	keyDecomposer Factorer[WholeKey, KeyPartA, KeyPartB],
	unifiedObserver MapChangeReceiver[WholeKey, Val],
	outerKeysetObserver SetChangeReceiver[KeyPartA],
	outerObserver2 MappingReceiver[KeyPartA, Map[KeyPartB, Val]],
) FactoredMap[WholeKey, KeyPartA, KeyPartB, Val] {
	return NewGenericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MapMap[KeyPartB, Val], Map[KeyPartB, Val]](
		keyDecomposer,
		func(KeyPartA) MapMap[KeyPartB, Val] { return NewMapMap[KeyPartB, Val](nil) },
		func(inner MapMap[KeyPartB, Val]) MutableMap[KeyPartB, Val] { return inner },
		func(inner MapMap[KeyPartB, Val]) Map[KeyPartB, Val] { return inner },
		NewMapMap[KeyPartA, MapMap[KeyPartB, Val]](MapChangeReceiverFuncs[KeyPartA, MapMap[KeyPartB, Val]]{
			OnCreate: func(keyPartA KeyPartA, rest MapMap[KeyPartB, Val]) {
				if outerKeysetObserver != nil {
					outerKeysetObserver(true, keyPartA)
				}
			},
			OnDelete: func(keyPartA KeyPartA, rest MapMap[KeyPartB, Val]) {
				if outerKeysetObserver != nil {
					outerKeysetObserver(false, keyPartA)
				}
			}}),
		unifiedObserver,
		outerObserver2,
	)
}

// NewFactoredMap creates a factored map implemented by the given outer map
// and constructor of inner maps.
func NewGenericFactoredMap[WholeKey, KeyPartA, KeyPartB comparable, Val, MutableInnerMap, InnerMap any](
	keyDecomposer Factorer[WholeKey, KeyPartA, KeyPartB],
	innerMapConstructor func(KeyPartA) MutableInnerMap,
	innerMapAsMap func(MutableInnerMap) MutableMap[KeyPartB, Val],
	insulateInner func(MutableInnerMap) InnerMap,
	outerMap MutableMap[KeyPartA, MutableInnerMap],
	unifiedObserver MapChangeReceiver[WholeKey, Val],
	outerObserver MappingReceiver[KeyPartA, Map[KeyPartB, Val]],
) GenericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, InnerMap] {
	return &genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]{
		keyDecomposer:       keyDecomposer,
		innerMapConstructor: innerMapConstructor,
		innerMapAsMap:       innerMapAsMap,
		insulateInner:       insulateInner,
		outerMap:            outerMap,
		unifiedObserver:     unifiedObserver,
		outerObserver:       outerObserver,
	}
}

type genericFactoredMap[WholeKey, KeyPartA, KeyPartB comparable, Val, MutableInnerMap, InnerMap any] struct {
	keyDecomposer       Factorer[WholeKey, KeyPartA, KeyPartB]
	innerMapConstructor func(KeyPartA) MutableInnerMap
	innerMapAsMap       func(MutableInnerMap) MutableMap[KeyPartB, Val]
	insulateInner       func(MutableInnerMap) InnerMap
	outerMap            MutableMap[KeyPartA, MutableInnerMap]
	unifiedObserver     MapChangeReceiver[WholeKey, Val]
	outerObserver       MappingReceiver[KeyPartA, Map[KeyPartB, Val]]
}

func (fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) IsEmpty() bool {
	return fm.outerMap.IsEmpty()
}
func (fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) LenIsCheap() bool {
	return false
}

func (fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Len() int {
	var ans int
	fm.outerMap.Visit(func(outerMapping Pair[KeyPartA, MutableInnerMap]) error {
		ans += fm.innerMapAsMap(outerMapping.Second).Len()
		return nil
	})
	return ans
}

func (fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Visit(wholeVisitor func(Pair[WholeKey, Val]) error) error {
	return fm.outerMap.Visit(func(outerMapping Pair[KeyPartA, MutableInnerMap]) error {
		return fm.innerMapAsMap(outerMapping.Second).Visit(func(innerMapping Pair[KeyPartB, Val]) error {
			decomposedKey := Pair[KeyPartA, KeyPartB]{outerMapping.First, innerMapping.First}
			wholeKey := fm.keyDecomposer.Second(decomposedKey)
			return wholeVisitor(Pair[WholeKey, Val]{wholeKey, innerMapping.Second})
		})
	})
}

func (fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Get(wholeKey WholeKey) (Val, bool) {
	decomposedKey := fm.keyDecomposer.First(wholeKey)
	innerMap, ok := fm.outerMap.Get(decomposedKey.First)
	if !ok {
		var zeroVal Val
		return zeroVal, false
	}
	return fm.innerMapAsMap(innerMap).Get(decomposedKey.Second)
}

func (fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Put(wholeKey WholeKey, val Val) {
	decomposedKey := fm.keyDecomposer.First(wholeKey)
	innerMap, ok := fm.outerMap.Get(decomposedKey.First)
	if !ok {
		innerMap = fm.innerMapConstructor(decomposedKey.First)
		innerMapAsMap := fm.innerMapAsMap(innerMap)
		innerMapAsMap.Put(decomposedKey.Second, val)
		fm.outerMap.Put(decomposedKey.First, innerMap)
		if fm.unifiedObserver != nil {
			fm.unifiedObserver.Create(wholeKey, val)
		}
		if fm.outerObserver != nil {
			fm.outerObserver.Put(decomposedKey.First, innerMapAsMap)
		}
		return
	}
	innerMapAsMap := fm.innerMapAsMap(innerMap)
	if fm.unifiedObserver != nil || fm.outerObserver != nil {
		oldVal, had := innerMapAsMap.Get(decomposedKey.Second)
		innerMapAsMap.Put(decomposedKey.Second, val)
		if had {
			if fm.unifiedObserver != nil {
				fm.unifiedObserver.Update(wholeKey, oldVal, val)
			}
			if fm.outerObserver != nil {
				fm.outerObserver.Put(decomposedKey.First, innerMapAsMap)
			}
		} else {
			if fm.unifiedObserver != nil {
				fm.unifiedObserver.Create(wholeKey, val)
			}
			if fm.outerObserver != nil {
				fm.outerObserver.Put(decomposedKey.First, innerMapAsMap)
			}
		}
		return
	}
	innerMapAsMap.Put(decomposedKey.Second, val)
}

func (fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Delete(wholeKey WholeKey) {
	decomposedKey := fm.keyDecomposer.First(wholeKey)
	innerMap, ok := fm.outerMap.Get(decomposedKey.First)
	if !ok {
		return
	}
	innerMapAsMap := fm.innerMapAsMap(innerMap)
	oldVal, had := innerMapAsMap.Get(decomposedKey.Second)
	if !had {
		return
	}
	innerMapAsMap.Delete(decomposedKey.Second)
	if innerMapAsMap.IsEmpty() {
		fm.outerMap.Delete(decomposedKey.First)
		if fm.outerObserver != nil {
			fm.outerObserver.Delete(decomposedKey.First)
		}
	} else if fm.outerObserver != nil {
		fm.outerObserver.Put(decomposedKey.First, innerMapAsMap)
	}
	if fm.unifiedObserver != nil {
		fm.unifiedObserver.DeleteWithFinal(wholeKey, oldVal)
	}
}

func (fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) GetIndex() GenericFactoredMapIndex[KeyPartA, KeyPartB, Val, InnerMap] {
	return genericFactoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]{fm}
}

type genericFactoredMapIndex[WholeKey, KeyPartA, KeyPartB comparable, Val, MutableInnerMap, InnerMap any] struct {
	fm *genericFactoredMap[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]
}

func (fmi genericFactoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) IsEmpty() bool {
	return fmi.fm.IsEmpty()
}

func (fmi genericFactoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) LenIsCheap() bool {
	return fmi.fm.outerMap.LenIsCheap()
}

func (fmi genericFactoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Len() int {
	return fmi.fm.outerMap.Len()
}

func (fmi genericFactoredMapIndex[KWholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Get(keyPartA KeyPartA) (InnerMap, bool) {
	innerMap, have := fmi.fm.outerMap.Get(keyPartA)
	if !have {
		var zero InnerMap
		return zero, false
	}
	return fmi.fm.insulateInner(innerMap), true
}

func (fmi genericFactoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Visit(visitor func(Pair[KeyPartA, InnerMap]) error) error {
	return fmi.fm.outerMap.Visit(func(outerTup Pair[KeyPartA, MutableInnerMap]) error {
		innerMap := fmi.fm.insulateInner(outerTup.Second)
		return visitor(NewPair(outerTup.First, innerMap))
	})
}

func (fmi genericFactoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val, MutableInnerMap, InnerMap]) Visit1to2(keyPartA KeyPartA, visitor func(Pair[KeyPartB, Val]) error) error {
	innerMap, has := fmi.fm.outerMap.Get(keyPartA)
	if !has {
		return nil
	}
	innerMapAsMap := fmi.fm.innerMapAsMap(innerMap)
	return innerMapAsMap.Visit(visitor)
}

type SingleIndexedMap2[First, Second, Val any] GenericFactoredMap[Pair[First, Second], First, Second, Val, Map[Second, Val]]

func NewSingleIndexedMapMap2[First, Second comparable, Val any]() SingleIndexedMap2[First, Second, Val] {
	gfm := NewGenericFactoredMap[Pair[First, Second], First, Second, Val, MutableMap[Second, Val], Map[Second, Val]](
		PairFactorer[First, Second](),
		func(First) MutableMap[Second, Val] { return NewMapMap[Second, Val](nil) },
		Identity1[MutableMap[Second, Val]],
		MutableMapToReadonly[Second, Val],
		NewMapMap[First, MutableMap[Second, Val]](nil),
		nil, nil)
	return gfm
}

type SingleIndexedMap3[First, Second, Third, Val any] GenericFactoredMap[Triple[First, Second, Third],
	First, Pair[Second, Third], Val,
	GenericFactoredMap[Pair[Second, Third], Second, Third, Val, Map[Third, Val]],
]

func NewSingleIndexedMapMap3[First, Second, Third comparable, Val any]() SingleIndexedMap3[First, Second, Third, Val] {
	gfm := NewGenericFactoredMap[Triple[First, Second, Third], First, Pair[Second, Third], Val,
		GenericFactoredMap[Pair[Second, Third], Second, Third, Val, Map[Third, Val]],
		GenericFactoredMap[Pair[Second, Third], Second, Third, Val, Map[Third, Val]],
	](
		TripleFactorerTo1and23[First, Second, Third](),
		func(First) GenericFactoredMap[Pair[Second, Third], Second, Third, Val, Map[Third, Val]] {
			return NewGenericFactoredMap[Pair[Second, Third], Second, Third, Val, MutableMap[Third, Val], Map[Third, Val]](
				PairFactorer[Second, Third](),
				func(Second) MutableMap[Third, Val] { return NewMapMap[Third, Val](nil) },
				Identity1[MutableMap[Third, Val]],
				MutableMapToReadonly[Third, Val],
				NewMapMap[Second, MutableMap[Third, Val]](nil),
				nil, nil)
		},
		func(inner GenericFactoredMap[Pair[Second, Third], Second, Third, Val, Map[Third, Val]]) MutableMap[Pair[Second, Third], Val] {
			return inner
		},
		Identity1[GenericFactoredMap[Pair[Second, Third], Second, Third, Val, Map[Third, Val]]],
		NewMapMap[First, GenericFactoredMap[Pair[Second, Third], Second, Third, Val, Map[Third, Val]]](nil),
		nil, nil)
	return gfm
}
