package util

import (
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
