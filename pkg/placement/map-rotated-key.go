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

// RotateKeyMap rotates the key space of a Map.
// The given Map has keys that are rotated forward, according to the given Rotator, from the returned Map.
func RotateKeyMap[Original, Rotated comparable, Val any](rotator Rotator[Original, Rotated], inner Map[Rotated, Val]) Map[Original, Val] {
	return rotatedKeyMap[Original, Rotated, Val]{inner, rotator}
}

type rotatedKeyMap[Original, Rotated comparable, Val any] struct {
	Map[Rotated, Val]
	rotator Rotator[Original, Rotated]
}

func (rkm rotatedKeyMap[Original, Rotated, Val]) Get(key Original) (Val, bool) {
	rotatedKey := rkm.rotator.First(key)
	return rkm.Map.Get(rotatedKey)
}

func (rkm rotatedKeyMap[Original, Rotated, Val]) Visit(visitor func(Pair[Original, Val]) error) error {
	return rkm.Map.Visit(func(rotatedTup Pair[Rotated, Val]) error {
		originalKey := rkm.rotator.Second(rotatedTup.First)
		return visitor(Pair[Original, Val]{originalKey, rotatedTup.Second})
	})
}

func RotatedKeyMutableMap[Original, Rotated comparable, Val any](rotator Rotator[Original, Rotated], inner MutableMap[Rotated, Val]) MutableMap[Original, Val] {
	return rotatedKeyMutableMap[Original, Rotated, Val]{inner, rotator}
}

type rotatedKeyMutableMap[Original, Rotated comparable, Val any] struct {
	MutableMap[Rotated, Val]
	rotator Rotator[Original, Rotated]
}

func (rkm rotatedKeyMutableMap[Original, Rotated, Val]) Get(key Original) (Val, bool) {
	rotatedKey := rkm.rotator.First(key)
	return rkm.MutableMap.Get(rotatedKey)
}

func (rkm rotatedKeyMutableMap[Original, Rotated, Val]) Visit(visitor func(Pair[Original, Val]) error) error {
	return rkm.MutableMap.Visit(func(rotatedTup Pair[Rotated, Val]) error {
		originalKey := rkm.rotator.Second(rotatedTup.First)
		return visitor(Pair[Original, Val]{originalKey, rotatedTup.Second})
	})
}

func (rkm rotatedKeyMutableMap[Original, Rotated, Val]) Put(key Original, val Val) {
	rotatedKey := rkm.rotator.First(key)
	rkm.MutableMap.Put(rotatedKey, val)
}

func (rkm rotatedKeyMutableMap[Original, Rotated, Val]) Delete(key Original) {
	rotatedKey := rkm.rotator.First(key)
	rkm.MutableMap.Delete(rotatedKey)
}

// RotateMappingReceiver rotates the key space of a MappingReceiver.
func RotateMappingReceiver[Original, Rotated comparable, Val any](rotator Rotator[Original, Rotated], inner MappingReceiver[Rotated, Val]) MappingReceiver[Original, Val] {
	return rotatedMappingReceiver[Original, Rotated, Val]{
		rotator: rotator,
		inner:   inner,
	}
}

type rotatedMappingReceiver[Original, Rotated comparable, Val any] struct {
	rotator Rotator[Original, Rotated]
	inner   MappingReceiver[Rotated, Val]
}

func (rmr rotatedMappingReceiver[Original, Rotated, Val]) Put(key Original, val Val) {
	rotatedKey := rmr.rotator.First(key)
	rmr.inner.Put(rotatedKey, val)
}

func (rmr rotatedMappingReceiver[Original, Rotated, Val]) Delete(key Original) {
	rotatedKey := rmr.rotator.First(key)
	rmr.inner.Delete(rotatedKey)
}
