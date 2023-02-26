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
	"fmt"
	"math/rand"
	"sync"

	apimachtypes "k8s.io/apimachinery/pkg/types"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

// testUIDer is a UIDer that makes up an association whenever asked
// for one it has not already made up.
// The WaitGroup is so that tests can know when all the asynchronous processing is done.
type testUIDer struct {
	wg *sync.WaitGroup
	sync.Mutex
	rng       *rand.Rand
	current   []UIDPair
	consumers []func(edgeapi.ExternalName, apimachtypes.UID)
}

type UIDPair struct {
	en  edgeapi.ExternalName
	uid apimachtypes.UID
}

var _ UIDer = &testUIDer{}

func NewTestUIDer(rng *rand.Rand, wg *sync.WaitGroup) *testUIDer {
	return &testUIDer{
		wg:      wg,
		rng:     rng,
		current: []UIDPair{},
	}
}

func (tu *testUIDer) AddConsumer(consumer func(edgeapi.ExternalName, apimachtypes.UID)) {
	tu.Lock()
	defer tu.Unlock()
	tu.consumers = append(tu.consumers, consumer)
}

func (tu *testUIDer) TweakOne(rng *rand.Rand) {
	tu.Lock()
	defer tu.Unlock()
	var en edgeapi.ExternalName
	newUID := apimachtypes.UID(fmt.Sprintf("u%d", tu.rng.Intn(1000000000)))
	if len(tu.current) == 0 || rng.Intn(3) == 0 { // add a new one
		en = edgeapi.ExternalName{
			Workspace: fmt.Sprintf("ws%d", tu.rng.Intn(1000)),
			Name:      fmt.Sprintf("n%d", tu.rng.Intn(1000)),
		}
		tu.current = append(tu.current, UIDPair{en, newUID})
	} else { // modify an existing one
		which := rng.Intn(len(tu.current))
		tu.current[which].uid = newUID
		en = tu.current[which].en
	}
	for _, consumer := range tu.consumers {
		consumer(en, newUID)
	}
}

func (tu *testUIDer) TweakNAndWait(rng *rand.Rand, n int) func() {
	return func() {
		for iter := 1; iter <= n; iter++ {
			tu.TweakOne(rng)
		}
		tu.wg.Wait()
	}
}

func (tu *testUIDer) Get(en edgeapi.ExternalName) apimachtypes.UID {
	tu.Lock()
	defer tu.Unlock()
	if uid, ok := tu.lookupLocked(en); ok {
		return uid
	}
	ans := apimachtypes.UID(fmt.Sprintf("u%d", tu.rng.Intn(1000000000)))
	tu.current = append(tu.current, UIDPair{en, ans})
	// Notify asynchronously in case GetUID was called while some consumer is locked
	tu.wg.Add(1)
	go func() {
		defer tu.wg.Add(-1)
		tu.Lock()
		defer tu.Unlock()
		uid, ok := tu.lookupLocked(en)
		if !ok {
			panic(tu)
		}
		for _, consumer := range tu.consumers {
			consumer(en, uid)
		}
	}()
	return ans
}

func (tu *testUIDer) lookupLocked(en edgeapi.ExternalName) (apimachtypes.UID, bool) {
	for _, pair := range tu.current {
		if pair.en == en {
			return pair.uid, true
		}
	}
	return "", false
}

func (tu *testUIDer) Set(en edgeapi.ExternalName, uid apimachtypes.UID) {
	tu.Lock()
	defer tu.Unlock()
	found := false
	for idx, pair := range tu.current {
		if pair.en == en {
			tu.current[idx].uid = uid
			found = true
			break
		}
	}
	if !found {
		tu.current = append(tu.current, UIDPair{en, uid})
	}
	for _, consumer := range tu.consumers {
		consumer(en, uid)
	}
}
