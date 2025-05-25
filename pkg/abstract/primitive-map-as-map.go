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

// PrimitiveMap is a MutableMap implemented by a Go primitve map.
type PrimitiveMap[Key comparable, Val any] map[Key]Val

// Assert that PrimitiveMap implements MutableMap
var _ MutableMap[int, func()] = PrimitiveMap[int, func()]{}

// AsPrimitiveMap makes a primitive map look like a PrimitiveMap.
func AsPrimitiveMap[Key comparable, Val any](pm map[Key]Val) PrimitiveMap[Key, Val] {
	return PrimitiveMap[Key, Val](pm)
}

func (pm PrimitiveMap[Key, Val]) Length() int { return len(pm) }

func (pm PrimitiveMap[Key, Val]) Get(key Key) (Val, bool) {
	val, have := pm[key]
	return val, have
}

func (pm PrimitiveMap[Key, Val]) ContGet(key Key, cont func(Val)) {
	val, have := pm[key]
	if have {
		cont(val)
	}
}

func (pm PrimitiveMap[Key, Val]) Iterate2(yield func(Key, Val) error) error {
	for key, val := range pm {
		err := yield(key, val)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pm PrimitiveMap[Key, Val]) Put(key Key, val Val) (Val, bool) {
	old, had := pm[key]
	pm[key] = val
	return old, had
}

func (pm PrimitiveMap[Key, Val]) Delete(key Key) (Val, bool) {
	old, had := pm[key]
	delete(pm, key)
	return old, had
}
