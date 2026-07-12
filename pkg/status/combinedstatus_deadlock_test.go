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

package status

import (
	"sync"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// TestCompareCombinedStatusNoRecursiveReadLock is a regression test for the
// recursive read-lock deadlock (issue #3899). compareCombinedStatus held the
// resolution's read lock and then called generateCombinedStatus, which
// re-acquired the same read lock. sync.RWMutex is not reentrant: if a writer
// calls Lock in the window between the two RLock calls, the second RLock blocks
// behind the writer, which blocks behind the first reader -> deadlock.
//
// The test hammers compareCombinedStatus while a writer goroutine repeatedly
// takes the write lock on the same resolution. On the buggy code this hangs; on
// the fixed code it completes well within the timeout.
func TestCompareCombinedStatusNoRecursiveReadLock(t *testing.T) {
	res := &combinedStatusResolution{
		Name:                      "cs",
		StatusCollectorNameToData: map[string]*statusCollectorData{},
		CollectionDestinations:    sets.New[string]("wec1"),
	}

	objID := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Resource:   "deployments",
		ObjectName: cache.NewObjectName("ns", "name"),
	}
	status := &v1alpha1.CombinedStatus{}

	done := make(chan struct{})
	stop := make(chan struct{})
	var wg sync.WaitGroup

	// Writer: contends for the write lock on the same resolution so a Lock can
	// land between the two read-lock acquisitions of the old code path.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
			}
			res.Lock()
			res.StatusCollectorNameToData["sc"] = nil
			res.Unlock()
		}
	}()

	// Readers: repeatedly run the compare path.
	go func() {
		for i := 0; i < 2000; i++ {
			res.compareCombinedStatus(status, "bp", objID)
		}
		close(done)
	}()

	select {
	case <-done:
		close(stop)
		wg.Wait()
	case <-time.After(15 * time.Second):
		t.Fatal("compareCombinedStatus deadlocked: recursive read-lock with a waiting writer")
	}
}
