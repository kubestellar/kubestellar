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

// NewMapMap makes a new Map implemented by a Map, optionally with a given observer.
// For providing initial values see AddArgs, AddAllByVisit, MapMapCopy below.
func NewMapMap[Key comparable, Val any](observer MapChangeReceiver[Key, Val]) MapMap[Key, Val] {
	return MintMapMap[Key, Val](map[Key]Val{}, observer)
}

// NewMapMapFactory makes a func that returns new Map implemented by a Map, optionally with a given observer.
// In other words, this is the curried form of NewMapMap.
func NewMapMapFactory[Key comparable, Val any](observer MapChangeReceiver[Key, Val]) func() MutableMap[Key, Val] {
	return func() MutableMap[Key, Val] { return NewMapMap[Key, Val](observer) }
}

// MintMapMap takes the given map and obervers puts the MapMap seal of approval on them.
// It is the user's responsibility to not surprise themself.
func MintMapMap[Key comparable, Val any](theMap map[Key]Val, observer MapChangeReceiver[Key, Val]) MapMap[Key, Val] {
	return MapMap[Key, Val]{theMap: theMap, observer: observer}
}

// AddArgs adds the given pairs and returns the map itself.
// The pairs are added in the order given.
func (mm MapMap[Key, Val]) AddArgs(args ...Pair[Key, Val]) MapMap[Key, Val] {
	for _, arg := range args {
		mm.Put(arg.First, arg.Second)
	}
	return mm
}

// AddAllByVisit enumerates the given Visitable and adds every pair enumerated, in order,
// and returns the Map itself.
func (mm MapMap[Key, Val]) AddAllByVisit(what Visitable[Pair[Key, Val]]) MapMap[Key, Val] {
	what.Visit(func(tup Pair[Key, Val]) error {
		mm.Put(tup.First, tup.Second)
		return nil
	})
	return mm
}

// MapMapCopy creates a new MapMap holding what the given Visitable reported at construction time.
func MapMapCopy[Key comparable, Val any](observer MapChangeReceiver[Key, Val], other Visitable[Pair[Key, Val]]) MapMap[Key, Val] {
	return NewMapMap[Key, Val](observer).AddAllByVisit(other)
}

var _ MutableMap[int, string] = MapMap[int, string]{}

type MapMap[Key comparable, Val any] struct {
	theMap   map[Key]Val
	observer MapChangeReceiver[Key, Val]
}

func (mm MapMap[Key, Val]) IsEmpty() bool    { return len(mm.theMap) == 0 }
func (mm MapMap[Key, Val]) LenIsCheap() bool { return true }
func (mm MapMap[Key, Val]) Len() int         { return len(mm.theMap) }

func (mm MapMap[Key, Val]) Get(key Key) (Val, bool) {
	val, ok := mm.theMap[key]
	return val, ok
}

func (mm MapMap[Key, Val]) Visit(visitor func(Pair[Key, Val]) error) error {
	for key, val := range mm.theMap {
		if err := visitor(Pair[Key, Val]{key, val}); err != nil {
			return err
		}
	}
	return nil
}

func (mm MapMap[Key, Val]) Put(key Key, val Val) {
	oldVal, had := mm.theMap[key]
	mm.theMap[key] = val
	if mm.observer == nil {
	} else if had {
		mm.observer.Update(key, oldVal, val)
	} else {
		mm.observer.Create(key, val)
	}
}

func (mm MapMap[Key, Val]) Delete(key Key) {
	oldVal, had := mm.theMap[key]
	if had {
		delete(mm.theMap, key)
		if mm.observer != nil {
			mm.observer.DeleteWithFinal(key, oldVal)
		}
	}
}
