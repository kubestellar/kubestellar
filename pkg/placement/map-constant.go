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

func NewMapToConstant[Key comparable, Val any](keys Set[Key], val Val) Map[Key, Val] {
	return MapToConstant[Key, Val]{keys: keys, val: val}
}

type MapToConstant[Key comparable, Val any] struct {
	keys Set[Key]
	val  Val
}

var _ Map[string, func()] = MapToConstant[string, func()]{}

func (mtc MapToConstant[Key, Val]) IsEmpty() bool    { return mtc.keys.IsEmpty() }
func (mtc MapToConstant[Key, Val]) LenIsCheap() bool { return mtc.keys.LenIsCheap() }
func (mtc MapToConstant[Key, Val]) Len() int         { return mtc.keys.Len() }

func (mtc MapToConstant[Key, Val]) Get(key Key) (Val, bool) {
	if mtc.keys.Has(key) {
		return mtc.val, true
	}
	var val Val
	return val, false
}

func (mtc MapToConstant[Key, Val]) Visit(visitor func(Pair[Key, Val]) error) error {
	return mtc.keys.Visit(func(key Key) error {
		return visitor(Pair[Key, Val]{key, mtc.val})
	})
}
