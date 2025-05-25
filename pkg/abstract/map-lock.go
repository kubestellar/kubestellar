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
)

// MapToLockedLocker wraps a lock around a MapToLocked to create one
// that is safe for concurrent access.
// The outer Map does not allow a consumer of Iterate2 to
// access the outer map.
type MapToLockedLocker[Key, Val any] struct {
	mutex *sync.RWMutex // never nil
	inner MapToLocked[Key, Val]
}

var _ MapToLocked[func() int, func() bool] = &MapToLockedLocker[func() int, func() bool]{}

func NewMapToLockedLocker[Key, Val any](mutex *sync.RWMutex, inner MapToLocked[Key, Val]) *MapToLockedLocker[Key, Val] {
	if mutex == nil {
		mutex = &sync.RWMutex{}
	}
	return &MapToLockedLocker[Key, Val]{mutex: mutex, inner: inner}
}

func (ml *MapToLockedLocker[Key, Val]) Length() int {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	return ml.inner.Length()
}

func (ml *MapToLockedLocker[Key, Val]) ContGet(key Key, cont func(Val)) {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	ml.inner.ContGet(key, cont)
}

func (ml *MapToLockedLocker[Key, Val]) Iterate2(yield func(Key, Val) error) error {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	return ml.inner.Iterate2(yield)
}

// MapLocker wraps locking around an inner Map.
// The outer Map does not allow a consumer of Iterate2 to
// access the outer map.
type MapLocker[Key, Val any] struct {
	MapToLockedLocker[Key, Val]
	asMap Map[Key, Val]
}

// Assert that `*MapLocker` implements Map
var _ Map[func() int, func() bool] = &MapLocker[func() int, func() bool]{}

// NewMapLocker wraps locking around a given Map.
// If the given mutex is `nil` then this will allocate a new one to use
// for this purpose.
func NewMapLocker[Key, Val any](mutex *sync.RWMutex, inner Map[Key, Val]) *MapLocker[Key, Val] {
	if mutex == nil {
		mutex = &sync.RWMutex{}
	}
	return &MapLocker[Key, Val]{MapToLockedLocker[Key, Val]{mutex: mutex, inner: inner}, inner}
}

func (ml *MapLocker[Key, Val]) Get(key Key) (Val, bool) {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	return ml.asMap.Get(key)
}
