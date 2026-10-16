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
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestConcurrentMapBasicOperations(t *testing.T) {
	cm := NewConcurrentMap[string, int]()

	if l := cm.Len(); l != 0 {
		t.Fatalf("expected initial Len to be 0, got %d", l)
	}

	val, ok := cm.Get("key1")
	if ok || val != 0 {
		t.Fatalf("expected Get on non-existent key to return (0, false), got (%v, %v)", val, ok)
	}

	// Set a key
	cm.Set("key1", 100)
	if l := cm.Len(); l != 1 {
		t.Fatalf("expected Len to be 1 after Set, got %d", l)
	}
	val, ok = cm.Get("key1")
	if !ok || val != 100 {
		t.Fatalf("expected Get('key1') to return (100, true), got (%v, %v)", val, ok)
	}

	// Set another key
	cm.Set("key2", 200)
	if l := cm.Len(); l != 2 {
		t.Fatalf("expected Len to be 2, got %d", l)
	}

	// Overwrite existing key
	cm.Set("key1", 150)
	if l := cm.Len(); l != 2 {
		t.Fatalf("expected Len to remain 2 after overwrite, got %d", l)
	}
	val, ok = cm.Get("key1")
	if !ok || val != 150 {
		t.Fatalf("expected Get('key1') to return updated (150, true), got (%v, %v)", val, ok)
	}

	// Remove existing key
	cm.Remove("key1")
	if l := cm.Len(); l != 1 {
		t.Fatalf("expected Len to be 1 after Remove, got %d", l)
	}
	val, ok = cm.Get("key1")
	if ok || val != 0 {
		t.Fatalf("expected Get on removed key to return (0, false), got (%v, %v)", val, ok)
	}

	// Remove non-existent key
	cm.Remove("non-existent")
	if l := cm.Len(); l != 1 {
		t.Fatalf("expected Len to remain 1 after Removing non-existent key, got %d", l)
	}

	// Remove remaining key
	cm.Remove("key2")
	if l := cm.Len(); l != 0 {
		t.Fatalf("expected Len to be 0 after removing all keys, got %d", l)
	}
}

func TestConcurrentMapIterator(t *testing.T) {
	cm := NewConcurrentMap[string, string]()

	// Test iteration on empty map
	err := cm.Iterator(func(k, v string) error {
		return fmt.Errorf("should not be called for empty map")
	})
	if err != nil {
		t.Fatalf("expected nil error on empty map iteration, got %v", err)
	}

	// Populate map
	expectedData := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}
	for k, v := range expectedData {
		cm.Set(k, v)
	}

	// Test successful iteration over all elements
	visited := make(map[string]string)
	err = cm.Iterator(func(k, v string) error {
		visited[k] = v
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error on full iteration, got %v", err)
	}
	if len(visited) != len(expectedData) {
		t.Fatalf("expected %d visited items, got %d", len(expectedData), len(visited))
	}
	for k, v := range expectedData {
		if visited[k] != v {
			t.Fatalf("expected visited[%s] = %s, got %s", k, v, visited[k])
		}
	}

	// Test early termination when yield returns error
	targetErr := errors.New("stop iteration")
	visitedCount := 0
	err = cm.Iterator(func(k, v string) error {
		visitedCount++
		return targetErr
	})
	if !errors.Is(err, targetErr) {
		t.Fatalf("expected error %v, got %v", targetErr, err)
	}
	if visitedCount != 1 {
		t.Fatalf("expected iteration to stop after 1 call, but was called %d times", visitedCount)
	}
}

func TestConcurrentMapConcurrency(t *testing.T) {
	cm := NewConcurrentMap[int, int]()
	const numGoroutines = 50
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := gID*opsPerGoroutine + j
				cm.Set(key, j)

				if val, ok := cm.Get(key); ok && val != j {
					t.Errorf("goroutine %d: key %d expected value %d, got %d", gID, key, j, val)
				}

				_ = cm.Len()

				// Perform iterator call
				_ = cm.Iterator(func(k, v int) error {
					return nil
				})

				if j%2 == 0 {
					cm.Remove(key)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify Len calculation matches remaining items
	expectedLen := 0
	_ = cm.Iterator(func(k, v int) error {
		expectedLen++
		return nil
	})

	if actualLen := cm.Len(); actualLen != expectedLen {
		t.Fatalf("expected Len() %d to match Iterator count %d", actualLen, expectedLen)
	}
}
