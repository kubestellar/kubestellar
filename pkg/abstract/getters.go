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

// Getter is anything that can Get.
// The keys must be comparable in some way, not necessarily
// using Go's built-in `==`.
type Getter[Key, Val any] interface {
	// Get returns the associated value and `true`, if there is an associated value,
	// otherwise an unspecified value and `false`.
	Get(Key) (Val, bool)
}

// ConstGetter maps every key to the same value
type ConstGetter[Key, Val any] struct{ val Val }

var _ Getter[func(), func(string)] = ConstGetter[func(), func(string)]{}

// NewConstGetter returns a constant Getter that returns the given value
func NewConstGetter[Key, Val any](val Val) ConstGetter[Key, Val] {
	return ConstGetter[Key, Val]{val: val}
}

func (cg ConstGetter[Key, Val]) Get(Key) (Val, bool) { return cg.val, true }

// GetterSeries is a Getter that is based on an underlying series of Getters.
// Each is consulted in turn, until the first one finds the key.
type GetterSeries[Key, Val any] []Map[Key, Val]

var _ Getter[func(), func(string)] = GetterSeries[func(), func(string)]{}

// NewGetterSeries makes a GetterSeries
func NewGetterSeries[Key, Val any](base ...Map[Key, Val]) GetterSeries[Key, Val] {
	return GetterSeries[Key, Val](base)
}
func (gs GetterSeries[Key, Val]) Get(key Key) (Val, bool) {
	for _, elt := range gs {
		val, has := elt.Get(key)
		if has {
			return val, true
		}
	}
	var zero Val
	return zero, false
}
