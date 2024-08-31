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

// MapMapValues composes a Map with a function.
// A consumer of Iterate2 of the composition can access the composition
// or the inner map iff a consumer of `inner.Iterate2` can.
func MapMapValues[Key, Val, Mapped any](inner Map[Key, Val], mapper func(Val) Mapped) Map[Key, Mapped] {
	return MapValueMapper[Key, Val, Mapped]{inner: inner, mapVal: mapper}
}

// MapValueMapper is the composition of a Map and a function.
// A consumer of Iterate2 of the composition can access the composition
// or the inner map iff a consumer of `inner.Iterate2` can.
type MapValueMapper[Key, Val, Mapped any] struct {
	inner  Map[Key, Val]
	mapVal func(Val) Mapped
}

var _ Map[func() int, func() string] = MapValueMapper[func() int, func() bool, func() string]{}

func (wrapper MapValueMapper[Key, Val, Mapped]) Length() int {
	return wrapper.inner.Length()
}

func (wrapper MapValueMapper[Key, Val, Mapped]) Get(key Key) (Mapped, bool) {
	val, have := wrapper.inner.Get(key)
	if !have {
		var zero Mapped
		return zero, false
	}
	return wrapper.mapVal(val), true
}

func (wrapper MapValueMapper[Key, Val, Mapped]) Iterate2(yield func(Key, Mapped) error) error {
	return wrapper.inner.Iterate2(func(key Key, val Val) error {
		mapped := wrapper.mapVal(val)
		return yield(key, mapped)
	})
}
