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

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

// MutableMapToComparable is an indexed mutable map where
// the domain and range types are both comparable.
type MutableMapToComparable[Key, Val comparable] interface {
	MutableMap[Key, Val]

	ReadInverse() MapToLocked[Val, sets.Set[Key]]
}

type IndexedMapToComparable[Key, Val comparable] struct {
	forward MutableMap[Key, Val]
	reverse MutableMap[Val, sets.Set[Key]]
}

func NewPrimitiveMapToComparable[Key, Val comparable]() *IndexedMapToComparable[Key, Val] {
	return &IndexedMapToComparable[Key, Val]{
		forward: AsPrimitiveMap(map[Key]Val{}),
		reverse: AsPrimitiveMap(map[Val]sets.Set[Key]{}),
	}
}

var _ MutableMapToComparable[string, int] = &IndexedMapToComparable[string, int]{}

func (imc *IndexedMapToComparable[Key, Val]) Length() int {
	return imc.forward.Length()
}

func (imc *IndexedMapToComparable[Key, Val]) ContGet(key Key, cont func(Val)) {
	imc.forward.ContGet(key, cont)
}

func (imc *IndexedMapToComparable[Key, Val]) Get(key Key) (Val, bool) {
	return imc.forward.Get(key)
}

func (imc *IndexedMapToComparable[Key, Val]) Put(key Key, val Val) (Val, bool) {
	oldVal, had := imc.forward.Put(key, val)
	if had {
		oldKeys, _ := imc.reverse.Get(oldVal)
		oldKeys.Delete(key)
	}
	valKeys, hadVal := imc.reverse.Get(val)
	if hadVal {
		valKeys.Insert(key)
	} else {
		valKeys = sets.New[Key](key)
		imc.reverse.Put(val, valKeys)
	}
	return oldVal, had
}

func (imc *IndexedMapToComparable[Key, Val]) Delete(key Key) (Val, bool) {
	oldVal, had := imc.forward.Delete(key)
	if had {
		oldKeys, _ := imc.reverse.Get(oldVal)
		oldKeys.Delete(key)
	}
	return oldVal, had
}

func (imc *IndexedMapToComparable[Key, Val]) Iterate2(yield func(Key, Val) error) error {
	return imc.forward.Iterate2(yield)
}

func (imc *IndexedMapToComparable[Key, Val]) ReadInverse() MapToLocked[Val, sets.Set[Key]] {
	return imc.reverse
}

type LockedMapToComparable[Key, Val comparable] struct {
	lock    *sync.RWMutex
	inner   MutableMapToComparable[Key, Val]
	inverse MapToLocked[Val, sets.Set[Key]]
}

func NewLockedMapToComparable[Key, Val comparable](lock *sync.RWMutex, inner MutableMapToComparable[Key, Val]) *LockedMapToComparable[Key, Val] {
	if lock == nil {
		lock = new(sync.RWMutex)
	}
	return &LockedMapToComparable[Key, Val]{
		lock:    lock,
		inner:   inner,
		inverse: NewMapToLockedLocker(lock, inner.ReadInverse())}
}

var _ MutableMapToComparable[int, bool] = &LockedMapToComparable[int, bool]{}

func (lmc *LockedMapToComparable[Key, Val]) ContGet(key Key, cont func(Val)) {
	lmc.lock.RLock()
	defer lmc.lock.RUnlock()
	lmc.inner.ContGet(key, cont)
}

func (lmc *LockedMapToComparable[Key, Val]) Get(key Key) (Val, bool) {
	lmc.lock.RLock()
	defer lmc.lock.RUnlock()
	return lmc.inner.Get(key)
}

func (lmc *LockedMapToComparable[Key, Val]) Iterate2(yield func(Key, Val) error) error {
	lmc.lock.RLock()
	defer lmc.lock.RUnlock()
	return lmc.inner.Iterate2(yield)
}

func (lmc *LockedMapToComparable[Key, Val]) Length() int {
	lmc.lock.RLock()
	defer lmc.lock.RUnlock()
	return lmc.inner.Length()
}

func (lmc *LockedMapToComparable[Key, Val]) Put(key Key, val Val) (Val, bool) {
	lmc.lock.Lock()
	defer lmc.lock.Unlock()
	return lmc.inner.Put(key, val)
}

func (lmc *LockedMapToComparable[Key, Val]) Delete(key Key) (Val, bool) {
	lmc.lock.Lock()
	defer lmc.lock.Unlock()
	return lmc.inner.Delete(key)
}

func (lmc *LockedMapToComparable[Key, Val]) ReadInverse() MapToLocked[Val, sets.Set[Key]] {
	return lmc.inverse
}
