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
	"time"
)

// TestIteratorNoDeadlockOnConcurrentWrite verifies that Iterator does not
// deadlock when a concurrent writer is waiting for the write lock.
func TestIteratorNoDeadlockOnConcurrentWrite(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)

	done := make(chan struct{})
	firstYield := true

	err := m.Iterator(func(k string, v int) error {
		if firstYield {
			firstYield = false
			go func() {
				m.Set("c", 3)
				close(done)
			}()
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Errorf("concurrent Set deadlocked")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := m.Get("c"); !ok {
		t.Fatal("concurrent Set during Iterator did not take effect")
	}
}

// TestIteratorSnapshotSemantics verifies that mutations after Iterator starts
// are not visible within the iteration.
func TestIteratorSnapshotSemantics(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	m.Set("a", 1)

	seen := map[string]int{}
	err := m.Iterator(func(k string, v int) error {
		m.Set("b", 2) // mutation during yield -- must not appear in this run
		seen[k] = v
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := seen["b"]; ok {
		t.Fatal("snapshot was not taken: mutation during iteration was visible")
	}
}

// TestConcurrentMap_Basic verifies standard operations under parallel load.
func TestConcurrentMap_Basic(t *testing.T) {
	cm := NewConcurrentMap[string, int]()
	var wg sync.WaitGroup

	// Concurrent Writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			cm.Set("key", val)
		}(i)
	}

	// Concurrent Readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cm.Get("key")
			_ = cm.Len()
		}()
	}

	wg.Wait()
}

// TestConcurrentMap_Iterator_SafeMutation ensures that calling mutation
// methods within or alongside the iterator does not cause a deadlock.
func TestConcurrentMap_Iterator_SafeMutation(t *testing.T) {
	cm := NewConcurrentMap[string, int]()
	cm.Set("key1", 100)
	cm.Set("key2", 200)

	err := cm.Iterator(func(k string, v int) error {
		if k == "key1" {
			// Simulating mutation during iteration (would deadlock previously)
			cm.Set("key1_updated", 150)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("unexpected error during iteration: %v", err)
	}

	if _, ok := cm.Get("key1_updated"); !ok {
		t.Error("expected key1_updated to be set successfully")
	}
}
