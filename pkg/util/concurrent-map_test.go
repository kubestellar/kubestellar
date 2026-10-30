package util

import (
	"testing"
	"time"
)

func TestIteratorNoDeadlock(t *testing.T) {
	m := NewConcurrentMap[string, string]()
	m.Set("a", "1")

	done := make(chan struct{})
	go func() {
		err := m.Iterator(func(k, v string) error {
			m.Get(k)
			return nil
		})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock detected")
	}
}
