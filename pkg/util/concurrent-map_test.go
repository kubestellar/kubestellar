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
	"testing"
)

func TestConcurrentMap_SetAndGet(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	// Test Set and Get
	cm.Set("key1", 100)
	cm.Set("key2", 200)

	val, ok := cm.Get("key1")
	if !ok {
		t.Errorf("expected key1 to exist")
	}
	if val != 100 {
		t.Errorf("expected value 100, got %d", val)
	}

	val, ok = cm.Get("key2")
	if !ok {
		t.Errorf("expected key2 to exist")
	}
	if val != 200 {
		t.Errorf("expected value 200, got %d", val)
	}
}

func TestConcurrentMap_GetNonExistent(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	_, ok := cm.Get("nonexistent")
	if ok {
		t.Errorf("expected non-existent key to return false")
	}
}

func TestConcurrentMap_Remove(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	cm.Set("key1", 100)
	cm.Set("key2", 200)

	// Remove existing key
	cm.Remove("key1")
	_, ok := cm.Get("key1")
	if ok {
		t.Errorf("expected key1 to be removed")
	}

	// Remove non-existent key (should not panic)
	cm.Remove("nonexistent")
}

func TestConcurrentMap_Len(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	// Empty map
	if cm.Len() != 0 {
		t.Errorf("expected empty map to have length 0, got %d", cm.Len())
	}

	// Add items
	cm.Set("key1", 100)
	cm.Set("key2", 200)
	cm.Set("key3", 300)

	if cm.Len() != 3 {
		t.Errorf("expected length 3, got %d", cm.Len())
	}

	// Remove item
	cm.Remove("key1")
	if cm.Len() != 2 {
		t.Errorf("expected length 2 after removal, got %d", cm.Len())
	}
}

func TestConcurrentMap_Iterator(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	cm.Set("key1", 100)
	cm.Set("key2", 200)
	cm.Set("key3", 300)

	count := 0
	sum := 0

	err := cm.Iterator(func(key string, value int) error {
		count++
		sum += value
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error from iterator: %v", err)
	}

	if count != 3 {
		t.Errorf("expected to iterate 3 items, got %d", count)
	}

	if sum != 600 {
		t.Errorf("expected sum 600, got %d", sum)
	}
}

func TestConcurrentMap_IteratorError(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	cm.Set("key1", 100)
	cm.Set("key2", 200)
	cm.Set("key3", 300)

	expectedErr := "test error"
	err := cm.Iterator(func(key string, value int) error {
		if key == "key2" {
			return &testError{msg: expectedErr}
		}
		return nil
	})

	if err == nil {
		t.Errorf("expected error from iterator")
	}

	if err.Error() != expectedErr {
		t.Errorf("expected error message %q, got %q", expectedErr, err.Error())
	}
}

func TestConcurrentMap_ConcurrentAccess(t *testing.T) {
	cm := NewConcurrentMap[int, int]()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cm.Set(i, i*10)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cm.Get(i)
		}(i)
	}

	// Concurrent removes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cm.Remove(i)
		}(i)
	}

	wg.Wait()

	// Verify final state
	if cm.Len() != 50 {
		t.Errorf("expected 50 items after concurrent operations, got %d", cm.Len())
	}
}

func TestConcurrentMap_Overwrite(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	cm.Set("key1", 100)
	cm.Set("key1", 200) // Overwrite

	val, ok := cm.Get("key1")
	if !ok {
		t.Errorf("expected key1 to exist")
	}
	if val != 200 {
		t.Errorf("expected overwritten value 200, got %d", val)
	}

	if cm.Len() != 1 {
		t.Errorf("expected length 1 after overwrite, got %d", cm.Len())
	}
}

func TestConcurrentMap_IteratorEmpty(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	count := 0
	err := cm.Iterator(func(key string, value int) error {
		count++
		return nil
	})

	if err != nil {
		t.Errorf("unexpected error from iterator on empty map: %v", err)
	}

	if count != 0 {
		t.Errorf("expected to iterate 0 items on empty map, got %d", count)
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
