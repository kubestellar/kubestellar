/*
Copyright 2026 The KubeStellar Authors.

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

package binding

import (
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

func newTestBindingPolicyResolver(t *testing.T) BindingPolicyResolver {
	t.Helper()
	resolver := NewBindingPolicyResolver()
	resolver.NoteBindingPolicy(&v1alpha1.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "bp1", UID: types.UID("uid-1")},
	})
	return resolver
}

func TestGetSingletonReportedStateRequestsForBinding(t *testing.T) {
	resolver := newTestBindingPolicyResolver(t)

	obj1 := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Resource:   "deployments",
		ObjectName: cache.ObjectName{Namespace: "ns1", Name: "deploy1"},
	}
	obj2 := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Resource:   "deployments",
		ObjectName: cache.ObjectName{Namespace: "ns1", Name: "deploy2"},
	}

	modulation := DownsyncModulation{
		WantSingletonReportedState: true,
		StatusCollectors:           sets.New[string](),
	}
	if _, err := resolver.EnsureObjectData("bp1", obj1, "uid-1", "rv-1", modulation); err != nil {
		t.Fatalf("failed to ensure object data for obj1: %v", err)
	}
	if _, err := resolver.EnsureObjectData("bp1", obj2, "uid-2", "rv-2", ZeroDownsyncModulation()); err != nil {
		t.Fatalf("failed to ensure object data for obj2: %v", err)
	}
	if err := resolver.SetDestinations("bp1", sets.New("wec1", "wec2")); err != nil {
		t.Fatalf("failed to set destinations: %v", err)
	}

	statuses := resolver.GetSingletonReportedStateRequestsForBinding("bp1")
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d: %v", len(statuses), statuses)
	}

	for _, status := range statuses {
		if status.ObjectId == obj1 {
			if !status.WantSingletonReportedState {
				t.Errorf("expected singleton reported state requested for %v", obj1)
			}
			if status.NumWECs != 2 {
				t.Errorf("expected 2 qualified WECs for %v, got %d", obj1, status.NumWECs)
			}
		} else if status.ObjectId == obj2 {
			if status.WantSingletonReportedState {
				t.Errorf("expected no singleton reported state requested for %v", obj2)
			}
			if status.NumWECs != 0 {
				t.Errorf("expected 0 qualified WECs for %v, got %d", obj2, status.NumWECs)
			}
		} else {
			t.Errorf("unexpected object id %v in statuses", status.ObjectId)
		}
	}
}

func TestGetSingletonReportedStateRequestsForBindingNoResolution(t *testing.T) {
	resolver := NewBindingPolicyResolver()
	if statuses := resolver.GetSingletonReportedStateRequestsForBinding("no-such-policy"); statuses != nil {
		t.Fatalf("expected nil statuses for missing resolution, got %v", statuses)
	}
}

// TestGetResolutionLockedWhileWriterQueued guards against the recursive
// read-lock deadlock that occurred when GetSingletonReportedStateRequestsForBinding
// re-acquired the resolver's read lock (via getResolution and
// GetReportedStateRequestForObject) while a writer was queued. Go's sync.RWMutex
// blocks new readers as soon as a writer is waiting, so a nested read lock held
// while a writer queues behind the outer read lock deadlocks forever. The
// locked helpers used by the fixed implementation must not lock again.
func TestGetResolutionLockedWhileWriterQueued(t *testing.T) {
	resolver := newTestBindingPolicyResolver(t)
	concrete := resolver.(*bindingPolicyResolver)

	concrete.RLock()

	writerQueued := make(chan struct{})
	writerDone := make(chan struct{})
	go func() {
		concrete.Lock()
		defer concrete.Unlock()
		close(writerDone)
	}()
	// The writer is now waiting behind our read lock; the RWMutex will block
	// any further read-lock attempts until the writer runs.
	close(writerQueued)
	<-writerQueued
	time.Sleep(20 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		concrete.getResolutionLocked("bp1")
		concrete.getReportedStateRequestForObjectLocked(util.ObjectIdentifier{})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("locked helpers blocked while a writer was queued; recursive read lock would deadlock")
	}

	concrete.RUnlock()
	<-writerDone
}

// TestGetSingletonReportedStateRequestsForBindingConcurrent exercises the
// function under realistic concurrent resolution mutations and guards against
// the recursive read-lock deadlock that occurred when the function re-acquired
// the resolver's read lock while a writer was queued.
func TestGetSingletonReportedStateRequestsForBindingConcurrent(t *testing.T) {
	resolver := newTestBindingPolicyResolver(t)

	obj1 := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Resource:   "deployments",
		ObjectName: cache.ObjectName{Namespace: "ns1", Name: "deploy1"},
	}
	modulation := DownsyncModulation{
		WantSingletonReportedState: true,
		StatusCollectors:           sets.New[string](),
	}
	if _, err := resolver.EnsureObjectData("bp1", obj1, "uid-1", "rv-1", modulation); err != nil {
		t.Fatalf("failed to ensure object data: %v", err)
	}
	if err := resolver.SetDestinations("bp1", sets.New("wec1", "wec2")); err != nil {
		t.Fatalf("failed to set destinations: %v", err)
	}

	stop := make(chan struct{})
	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			resolver.DeleteResolution("bp1")
			resolver.NoteBindingPolicy(&v1alpha1.BindingPolicy{
				ObjectMeta: metav1.ObjectMeta{Name: "bp1", UID: types.UID("uid-1")},
			})
			if _, err := resolver.EnsureObjectData("bp1", obj1, "uid-1", "rv-1", modulation); err != nil {
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				resolver.GetSingletonReportedStateRequestsForBinding("bp1")
			}
		}()
	}

	go func() {
		time.Sleep(2 * time.Second)
		close(stop)
	}()
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("GetSingletonReportedStateRequestsForBinding deadlocked under concurrent writes")
	}
}
