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
	"sync"
	"testing"
)

func TestConcurrentMap_SetGet(t *testing.T) {
	m := NewConcurrentMap[string, int]()

	if _, ok := m.Get("a"); ok {
		t.Fatalf("expected key 'a' to not exist in empty map")
	}

	m.Set("a", 1)
	v, ok := m.Get("a")
	if !ok {
		t.Fatalf("expected key 'a' to exist")
	}
	if v != 1 {
		t.Fatalf("expected value 1, got %d", v)
	}

	m.Set("a", 2)
	v, ok = m.Get("a")
	if !ok || v != 2 {
		t.Fatalf("expected updated value 2, got %d, ok=%v", v, ok)
	}
}

func TestConcurrentMap_Remove(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	m.Set("a", 1)
	m.Remove("a")

	if _, ok := m.Get("a"); ok {
		t.Fatalf("expected key 'a' to be removed")
	}

	// Removing a non-existent key should not panic.
	m.Remove("nonexistent")
}

func TestConcurrentMap_Len(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	if m.Len() != 0 {
		t.Fatalf("expected empty map len 0, got %d", m.Len())
	}

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	if m.Len() != 3 {
		t.Fatalf("expected len 3, got %d", m.Len())
	}

	m.Remove("b")
	if m.Len() != 2 {
		t.Fatalf("expected len 2 after remove, got %d", m.Len())
	}

	m.Set("a", 100) // update, not insert
	if m.Len() != 2 {
		t.Fatalf("expected len unchanged after update, got %d", m.Len())
	}
}

func TestConcurrentMap_Iterator(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	want := map[string]int{"a": 1, "b": 2, "c": 3}
	for k, v := range want {
		m.Set(k, v)
	}

	got := make(map[string]int)
	err := m.Iterator(func(k string, v int) error {
		got[k] = v
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d items, got %d", len(want), len(got))
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("key %s: expected %d, got %d", k, v, got[k])
		}
	}
}

func TestConcurrentMap_Iterator_ErrorStopsIteration(t *testing.T) {
	m := NewConcurrentMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	sentinel := errors.New("stop")
	calls := 0
	err := m.Iterator(func(k string, v int) error {
		calls++
		return sentinel
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected iteration to stop after first call, got %d calls", calls)
	}
}

func TestConcurrentMap_Concurrency(t *testing.T) {
	m := NewConcurrentMap[int, int]()
	const goroutines = 50
	const opsPerGoroutine = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				key := id*opsPerGoroutine + i
				m.Set(key, key*2)
				if v, ok := m.Get(key); !ok {
					t.Errorf("key %d: expected key to exist after Set", key)
				} else if v != key*2 {
					t.Errorf("key %d: expected %d, got %d", key, key*2, v)
				}
				m.Len()
				if err := m.Iterator(func(k, v int) error { return nil }); err != nil {
					t.Errorf("unexpected iterator error: %v", err)
				}
				m.Remove(key)
			}
		}(g)
	}

	wg.Wait()

	if m.Len() != 0 {
		t.Fatalf("expected empty map after concurrent removes, got len %d", m.Len())
	}
}

func TestConcurrentMap_TypeParameters(t *testing.T) {
	type customKey struct {
		Namespace string
		Name      string
	}
	type customValue struct {
		Data int
	}

	m := NewConcurrentMap[customKey, customValue]()
	k := customKey{Namespace: "ns", Name: "obj"}
	m.Set(k, customValue{Data: 42})

	v, ok := m.Get(k)
	if !ok || v.Data != 42 {
		t.Fatalf("expected custom struct value, got %+v, ok=%v", v, ok)
	}
}
