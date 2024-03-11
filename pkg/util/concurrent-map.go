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

package util

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ConcurrentMap is a thread-safe map.
type ConcurrentMap[K comparable, V any] interface {
	// Set sets the value for the given key.
	Set(key K, value V)
	// Remove removes the value for the given key.
	Remove(key K)
	// Get gets the value for the given key.
	// The second return value is true if the key exists in the map, otherwise false.
	// Getting a key does not guarantee that no other goroutine will also work with it.
	Get(key K) (V, bool)
	// Keys returns a copy of all keys in the map at the time of the call.
	// Nothing guarantees that when used, these keys are still in the map.
	Keys() sets.Set[K]
	// Len returns the number of items in the map.
	Len() int
}

// NewConcurrentMap creates a new ConcurrentMap.
func NewConcurrentMap[K comparable, V any]() ConcurrentMap[K, V] {
	return &rwMutexMap[K, V]{
		m: make(map[K]V),
	}
}

type rwMutexMap[K comparable, V any] struct {
	sync.RWMutex
	m map[K]V
}

// Set sets the value for the given key.
func (mm *rwMutexMap[K, V]) Set(key K, value V) {
	mm.Lock()
	defer mm.Unlock()

	mm.m[key] = value
}

// Remove removes the value for the given key.
func (mm *rwMutexMap[K, V]) Remove(key K) {
	mm.Lock()
	defer mm.Unlock()

	delete(mm.m, key)
}

// Get gets the value for the given key.
// The second return value is true if the key exists in the map, otherwise false.
// Getting a key does not guarantee that no other goroutine will also work with it.
func (mm *rwMutexMap[K, V]) Get(key K) (V, bool) {
	mm.RLock()
	defer mm.RUnlock()

	value, ok := mm.m[key]
	return value, ok
}

// Keys returns a copy of all keys in the map at the time of the call.
// Nothing guarantees that when used, these keys are still in the map.
func (mm *rwMutexMap[K, V]) Keys() sets.Set[K] {
	mm.RLock()
	defer mm.RUnlock()

	return sets.KeySet(mm.m)
}

// Len returns the number of items in the map.
func (mm *rwMutexMap[K, V]) Len() int {
	mm.RLock()
	defer mm.RUnlock()

	return len(mm.m)
}
