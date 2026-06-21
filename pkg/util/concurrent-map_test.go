package util

import (
	"sync"
	"testing"
)

// TestIteratorNoDeadlockOnConcurrentWrite verifies that Iterator does not
// deadlock when a concurrent writer is waiting for the write lock.
func TestIteratorNoDeadlockOnConcurrentWrite(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)

	var wg sync.WaitGroup
	wg.Add(1)

	err := m.Iterator(func(k string, v int) error {
		// Spawn writer while iterator is executing yield.
		// Pre-fix: deadlock. Post-fix: completes.
		go func() {
			defer wg.Done()
			m.Set("c", 3)
		}()
		wg.Wait()
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
	_ = m.Iterator(func(k string, v int) error {
		m.Set("b", 2) // mutation during yield -- must not appear in this run
		seen[k] = v
		return nil
	})

	if _, ok := seen["b"]; ok {
		t.Fatal("snapshot was not taken: mutation during iteration was visible")
	}
}
