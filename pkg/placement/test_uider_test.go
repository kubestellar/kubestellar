/*
Copyright 2023 The KCP Authors.

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

package placement

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	apimachtypes "k8s.io/apimachinery/pkg/types"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func TestTestUIDer(t *testing.T) {
	var wg sync.WaitGroup
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	uider := NewTestUIDer(rng, &wg)
	testConsumer := &testUIDConsumer{
		current: map[edgeapi.ExternalName]apimachtypes.UID{},
	}
	uider.AddConsumer(testConsumer.Note)
	en1 := edgeapi.ExternalName{Workspace: "ws1", Name: "n1"}
	en2 := edgeapi.ExternalName{Workspace: "ws1", Name: "n2"}
	uid1 := uider.GetUID(en1)
	uid2 := uider.GetUID(en2)
	wg.Wait()
	if len(testConsumer.current) != 2 {
		t.Errorf("Insufficient mappings: %v", testConsumer.current)
	}
	if actual, expected := uider.GetUID(en1), uid1; actual != expected {
		t.Errorf("Got %q instead of %q", actual, expected)
	}
	if actual, expected := uider.GetUID(en2), uid2; actual != expected {
		t.Errorf("Got %q instead of %q", actual, expected)
	}
}

type testUIDConsumer struct {
	sync.Mutex
	current map[edgeapi.ExternalName]apimachtypes.UID
}

func (tc *testUIDConsumer) Note(en edgeapi.ExternalName, uid apimachtypes.UID) {
	tc.Lock()
	defer tc.Unlock()
	tc.current[en] = uid
}
