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

// MapLocker wraps locking around an inner Map.
// The outer Map does not allow a consumer of Iterate2 to
// access the outer map.
type MapLocker[Key, Val any] struct {
	mutex *sync.RWMutex // never nil
	inner Map[Key, Val]
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
	return &MapLocker[Key, Val]{mutex: mutex, inner: inner}
}

func (ml *MapLocker[Key, Val]) Length() int {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	return ml.inner.Length()
}

func (ml *MapLocker[Key, Val]) Get(key Key) (Val, bool) {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	return ml.inner.Get(key)
}

func (ml *MapLocker[Key, Val]) Iterate2(yield func(Key, Val) error) error {
	ml.mutex.RLock()
	defer ml.mutex.RUnlock()
	return ml.inner.Iterate2(yield)
}
