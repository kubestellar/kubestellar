/*
Copyright 2024 The KubeStellar Authors.

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

package abstract

import "sync"

// PrimitiveMapGet exposes the indexing functionality of a primitive map as a func
func PrimitiveMapGet[Key comparable, Val any](rep map[Key]Val) func(Key) (Val, bool) {
	return func(key Key) (Val, bool) {
		val, has := rep[key]
		return val, has
	}
}

func PrimitiveMapEqual[Key, Val comparable](map1, map2 map[Key]Val) bool {
	if len(map1) != len(map2) {
		return false
	}
	for key, val1 := range map1 {
		val2, have2 := map2[key]
		if !(have2 && val1 == val2) {
			return false
		}
	}
	return true
}

// PrimitiveMapValMap creates a new map by applying a mapper function to each value in the source map.
func PrimitiveMapValMap[Key comparable, Val, Mapped any](source map[Key]Val,
	mapper func(Val) Mapped) map[Key]Mapped {
	result := make(map[Key]Mapped, len(source))
	for key, val := range source {
		result[key] = mapper(val)
	}

	return result
}

// PrimitiveMapSafeValMap creates a new map by applying a mapper function to each value in the source map.
// The source map is read-locked during the operation using the provided lock.
func PrimitiveMapSafeValMap[Key comparable, Val, Mapped any](lock *sync.RWMutex, source map[Key]Val,
	mapper func(Val) Mapped) map[Key]Mapped {
	lock.RLock()
	defer lock.RUnlock()

	return PrimitiveMapValMap(source, mapper)
}

func PrimitiveMapKeySlice[Key comparable, Val any](rep map[Key]Val) []Key {
	keys := make([]Key, 0, len(rep))
	for key := range rep {
		keys = append(keys, key)
	}
	return keys
}
