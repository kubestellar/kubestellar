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

package transport

import (
	"testing"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

// newTestController returns a controller wired with just the fields the delete
// handlers touch: a logger and a workqueue to observe enqueues.
func newTestController() *genericTransportController {
	return &genericTransportController{
		logger:    logr.Discard(),
		workqueue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test"),
	}
}

// drainKeys returns every key currently sitting in the workqueue.
func drainKeys(t *testing.T, q workqueue.RateLimitingInterface) []any {
	t.Helper()
	keys := []any{}
	for q.Len() > 0 {
		key, shutdown := q.Get()
		if shutdown {
			break
		}
		keys = append(keys, key)
		q.Done(key)
	}
	return keys
}

// TestDeleteHandlersUnwrapTombstone verifies that the informer delete handlers
// accept a cache.DeletedFinalStateUnknown tombstone and enqueue the embedded
// object instead of panicking. client-go delivers tombstones by value, so a
// pointer type assertion (*cache.DeletedFinalStateUnknown) silently fails to
// match and lets the raw tombstone reach the typed cast below, which panics.
func TestDeleteHandlersUnwrapTombstone(t *testing.T) {
	binding := &v1alpha1.Binding{ObjectMeta: metav1.ObjectMeta{Name: "binding-1"}}
	wrapped := &unstructured.Unstructured{}
	wrapped.SetName("mw-1")
	wrapped.SetLabels(map[string]string{originOwnerReferenceLabel: "owner-binding"})

	tests := []struct {
		name    string
		invoke  func(c *genericTransportController, obj any)
		obj     any
		wantKey any
	}{
		{
			name:    "binding direct",
			invoke:  func(c *genericTransportController, obj any) { c.handleBinding(tombstoneObject(obj), "delete") },
			obj:     binding,
			wantKey: "binding-1",
		},
		{
			name:    "binding tombstone",
			invoke:  func(c *genericTransportController, obj any) { c.handleBinding(tombstoneObject(obj), "delete") },
			obj:     cache.DeletedFinalStateUnknown{Key: "binding-1", Obj: binding},
			wantKey: "binding-1",
		},
		{
			name:    "wrapped object direct",
			invoke:  func(c *genericTransportController, obj any) { c.handleWrappedObject(tombstoneObject(obj), "delete") },
			obj:     wrapped,
			wantKey: "owner-binding",
		},
		{
			name:    "wrapped object tombstone",
			invoke:  func(c *genericTransportController, obj any) { c.handleWrappedObject(tombstoneObject(obj), "delete") },
			obj:     cache.DeletedFinalStateUnknown{Key: "mw-1", Obj: wrapped},
			wantKey: "owner-binding",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := newTestController()
			defer c.workqueue.ShutDown()

			// Would panic before the fix for the tombstone cases.
			tc.invoke(c, tc.obj)

			keys := drainKeys(t, c.workqueue)
			if len(keys) != 1 || keys[0] != tc.wantKey {
				t.Fatalf("expected workqueue to contain %q, got %v", tc.wantKey, keys)
			}
		})
	}
}

// TestTombstoneObject checks the unwrap helper in isolation.
func TestTombstoneObject(t *testing.T) {
	binding := &v1alpha1.Binding{ObjectMeta: metav1.ObjectMeta{Name: "b"}}

	if got := tombstoneObject(binding); got != binding {
		t.Errorf("expected a non-tombstone value to pass through unchanged")
	}

	tombstone := cache.DeletedFinalStateUnknown{Key: "b", Obj: binding}
	if got := tombstoneObject(tombstone); got != binding {
		t.Errorf("expected tombstone to be unwrapped to its embedded object, got %#v", got)
	}
}
