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

	timeout := 1 * time.Second
	if deadline, ok := t.Deadline(); ok {
		if remaining := time.Until(deadline); remaining < timeout {
			timeout = remaining / 2
		}
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
	case <-timer.C:
		t.Fatal("deadlock detected")
	}
}

func TestIteratorSafeMutation(t *testing.T) {
	m := NewConcurrentMap[string, string]()
	m.Set("a", "1")

	err := m.Iterator(func(k, v string) error {
		m.Set("b", "2")
		m.Remove("a")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := m.Get("b")
	if !ok || val != "2" {
		t.Errorf("expected key 'b' to be '2', got %q, %v", val, ok)
	}
	_, ok = m.Get("a")
	if ok {
		t.Errorf("expected key 'a' to be removed")
	}
}
