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

// FactoredMap is a map implemented by two levels of mapping.
// It has a general key decomposition/composition rotator.
type FactoredMap[WholeKey, KeyPartA, KeyPartB comparable, Val any] interface {
	MutableMap[WholeKey, Val]
	GetIndex() FactoredMapIndex[KeyPartA, KeyPartB, Val]
}

// FactoredMapIndex is an index into a factored map.
type FactoredMapIndex[KeyPartA, KeyPartB comparable, Val any] interface {
	Map[KeyPartA, Map[KeyPartB, Val]]
	Visit1to2(KeyPartA, func(Pair[KeyPartB, Val]) error) error
}

// NewFactoredMapMap makes a FactoredMap implements by golang maps.
func NewFactoredMapMap[WholeKey, KeyPartA, KeyPartB comparable, Val any](
	keyDecomposer Factorer[WholeKey, KeyPartA, KeyPartB],
	unifiedObserver MapChangeReceiver[WholeKey, Val],
	outerKeysetObserver SetChangeReceiver[KeyPartA],
	outerObserver2 MappingReceiver[KeyPartA, Map[KeyPartB, Val]],
) FactoredMap[WholeKey, KeyPartA, KeyPartB, Val] {
	return NewFactoredMap[WholeKey, KeyPartA, KeyPartB, Val](
		keyDecomposer,
		NewMapMapFactory[KeyPartB, Val](nil),
		NewMapMap[KeyPartA, MutableMap[KeyPartB, Val]](MapChangeReceiverFuncs[KeyPartA, MutableMap[KeyPartB, Val]]{
			OnCreate: func(keyPartA KeyPartA, rest MutableMap[KeyPartB, Val]) {
				if outerKeysetObserver != nil {
					outerKeysetObserver(true, keyPartA)
				}
			},
			OnDelete: func(keyPartA KeyPartA, rest MutableMap[KeyPartB, Val]) {
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
func NewFactoredMap[WholeKey, KeyPartA, KeyPartB comparable, Val any](
	keyDecomposer Factorer[WholeKey, KeyPartA, KeyPartB],
	innerMapConstructor func() MutableMap[KeyPartB, Val],
	outerMap MutableMap[KeyPartA, MutableMap[KeyPartB, Val]],
	unifiedObserver MapChangeReceiver[WholeKey, Val],
	outerObserver MappingReceiver[KeyPartA, Map[KeyPartB, Val]],
) FactoredMap[WholeKey, KeyPartA, KeyPartB, Val] {
	return &factoredMap[WholeKey, KeyPartA, KeyPartB, Val]{
		keyDecomposer:       keyDecomposer,
		innerMapConstructor: innerMapConstructor,
		outerMap:            outerMap,
		unifiedObserver:     unifiedObserver,
		outerObserver:       outerObserver,
	}
}

type factoredMap[WholeKey, KeyPartA, KeyPartB comparable, Val any] struct {
	keyDecomposer       Factorer[WholeKey, KeyPartA, KeyPartB]
	innerMapConstructor func() MutableMap[KeyPartB, Val]
	outerMap            MutableMap[KeyPartA, MutableMap[KeyPartB, Val]]
	unifiedObserver     MapChangeReceiver[WholeKey, Val]
	outerObserver       MappingReceiver[KeyPartA, Map[KeyPartB, Val]]
}

func (fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]) IsEmpty() bool {
	return fm.outerMap.IsEmpty()
}
func (fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]) LenIsCheap() bool { return false }

func (fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]) Len() int {
	var ans int
	fm.outerMap.Visit(func(outerMapping Pair[KeyPartA, MutableMap[KeyPartB, Val]]) error {
		ans += outerMapping.Second.Len()
		return nil
	})
	return ans
}

func (fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]) Visit(wholeVisitor func(Pair[WholeKey, Val]) error) error {
	return fm.outerMap.Visit(func(outerMapping Pair[KeyPartA, MutableMap[KeyPartB, Val]]) error {
		return outerMapping.Second.Visit(func(innerMapping Pair[KeyPartB, Val]) error {
			decomposedKey := Pair[KeyPartA, KeyPartB]{outerMapping.First, innerMapping.First}
			wholeKey := fm.keyDecomposer.Second(decomposedKey)
			return wholeVisitor(Pair[WholeKey, Val]{wholeKey, innerMapping.Second})
		})
	})
}

func (fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]) Get(wholeKey WholeKey) (Val, bool) {
	decomposedKey := fm.keyDecomposer.First(wholeKey)
	innerMap, ok := fm.outerMap.Get(decomposedKey.First)
	if !ok {
		var zeroVal Val
		return zeroVal, false
	}
	return innerMap.Get(decomposedKey.Second)
}

func (fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]) Put(wholeKey WholeKey, val Val) {
	decomposedKey := fm.keyDecomposer.First(wholeKey)
	innerMap, ok := fm.outerMap.Get(decomposedKey.First)
	if !ok {
		innerMap = fm.innerMapConstructor()
		innerMap.Put(decomposedKey.Second, val)
		fm.outerMap.Put(decomposedKey.First, innerMap)
		if fm.unifiedObserver != nil {
			fm.unifiedObserver.Create(wholeKey, val)
		}
		if fm.outerObserver != nil {
			fm.outerObserver.Put(decomposedKey.First, innerMap)
		}
		return
	}
	if fm.unifiedObserver != nil || fm.outerObserver != nil {
		oldVal, had := innerMap.Get(decomposedKey.Second)
		innerMap.Put(decomposedKey.Second, val)
		if had {
			if fm.unifiedObserver != nil {
				fm.unifiedObserver.Update(wholeKey, oldVal, val)
			}
			if fm.outerObserver != nil {
				fm.outerObserver.Put(decomposedKey.First, innerMap)
			}
		} else {
			if fm.unifiedObserver != nil {
				fm.unifiedObserver.Create(wholeKey, val)
			}
			if fm.outerObserver != nil {
				fm.outerObserver.Put(decomposedKey.First, innerMap)
			}
		}
		return
	}
	innerMap.Put(decomposedKey.Second, val)
}

func (fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]) Delete(wholeKey WholeKey) {
	decomposedKey := fm.keyDecomposer.First(wholeKey)
	innerMap, ok := fm.outerMap.Get(decomposedKey.First)
	if !ok {
		return
	}
	oldVal, had := innerMap.Get(decomposedKey.Second)
	if !had {
		return
	}
	innerMap.Delete(decomposedKey.Second)
	if innerMap.IsEmpty() {
		fm.outerMap.Delete(decomposedKey.First)
		if fm.outerObserver != nil {
			fm.outerObserver.Delete(decomposedKey.First)
		}
	} else if fm.outerObserver != nil {
		fm.outerObserver.Put(decomposedKey.First, innerMap)
	}
	if fm.unifiedObserver != nil {
		fm.unifiedObserver.DeleteWithFinal(wholeKey, oldVal)
	}
}

func (fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]) GetIndex() FactoredMapIndex[KeyPartA, KeyPartB, Val] {
	return factoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val]{fm}
}

type factoredMapIndex[WholeKey, KeyPartA, KeyPartB comparable, Val any] struct {
	fm *factoredMap[WholeKey, KeyPartA, KeyPartB, Val]
}

func (fmi factoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val]) IsEmpty() bool {
	return fmi.fm.IsEmpty()
}

func (fmi factoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val]) LenIsCheap() bool {
	return fmi.fm.outerMap.LenIsCheap()
}

func (fmi factoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val]) Len() int {
	return fmi.fm.outerMap.Len()
}

func (fmi factoredMapIndex[KWholeKey, KeyPartA, KeyPartB, Val]) Get(keyPartA KeyPartA) (Map[KeyPartB, Val], bool) {
	return fmi.fm.outerMap.Get(keyPartA)
}

func (fmi factoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val]) Visit(visitor func(Pair[KeyPartA, Map[KeyPartB, Val]]) error) error {
	return fmi.fm.outerMap.Visit(func(outerTup Pair[KeyPartA, MutableMap[KeyPartB, Val]]) error {
		return visitor(Pair[KeyPartA, Map[KeyPartB, Val]]{outerTup.First, MapReadonly[KeyPartB, Val](outerTup.Second)})
	})
}

func (fmi factoredMapIndex[WholeKey, KeyPartA, KeyPartB, Val]) Visit1to2(keyPartA KeyPartA, visitor func(Pair[KeyPartB, Val]) error) error {
	innerMap, has := fmi.fm.outerMap.Get(keyPartA)
	if !has {
		return nil
	}
	return innerMap.Visit(visitor)
}
