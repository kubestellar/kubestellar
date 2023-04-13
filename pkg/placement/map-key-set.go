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

// MapKeySet produces a readonly view of a given Map's key set.
// For the passive version, see MapKeySetReceiver and MapKeySetReceiverLossy.
func MapKeySet[Key comparable, Val any](theMap Map[Key, Val]) Set[Key] {
	return mapKeySet[Key, Val]{theMap}
}

type mapKeySet[Key comparable, Val any] struct{ theMap Map[Key, Val] }

func (mks mapKeySet[Key, Val]) IsEmpty() bool    { return mks.theMap.IsEmpty() }
func (mks mapKeySet[Key, Val]) LenIsCheap() bool { return mks.theMap.LenIsCheap() }
func (mks mapKeySet[Key, Val]) Len() int         { return mks.theMap.Len() }

func (mks mapKeySet[Key, Val]) Has(key Key) bool {
	_, has := mks.theMap.Get(key)
	return has
}

func (mks mapKeySet[Key, Val]) Visit(visitor func(Key) error) error {
	return mks.theMap.Visit(func(tup Pair[Key, Val]) error {
		return visitor(tup.First)
	})
}

// MapKeySetReceiver extends a SetChangeReceiver for keys into a MapChangeReceiver that simply ignores the associated values.
func MapKeySetReceiver[Key comparable, Val any](ksr SetChangeReceiver[Key]) MapChangeReceiver[Key, Val] {
	return MapChangeReceiverFuncs[Key, Val]{
		OnCreate: func(key Key, _ Val) { ksr.Add(key) },
		OnDelete: func(key Key, _ Val) { ksr.Remove(key) },
	}
}

// MapKeySetReceiverLossy extends a SetChangeReceiver for keys into a MappingReceiver that simply ignores the associated values.
// It is lossy in that it may make redundant calls to Add.
func MapKeySetReceiverLossy[Key comparable, Val any](ksr SetChangeReceiver[Key]) MappingReceiver[Key, Val] {
	return MappingReceiverFuncs[Key, Val]{
		OnPut:    func(key Key, _ Val) { ksr.Add(key) },
		OnDelete: func(key Key) { ksr.Remove(key) },
	}
}
