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
