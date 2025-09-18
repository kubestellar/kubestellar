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

package abstract

import (
	"math/rand"
	"sync"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Pair[First, Second any] struct {
	First  First
	Second Second
}

func NewPair[First, Second any](first First, second Second) Pair[First, Second] {
	return Pair[First, Second]{first, second}
}
func FuzzTestLockedMapToComparable(f *testing.F) {
	f.Add(int64(42))
	f.Add(int64(446))
	f.Fuzz(func(t *testing.T, seed int64) {
		rg := rand.New(rand.NewSource(seed))
		var randMutex sync.Mutex
		intn := func(bound int) int {
			randMutex.Lock()
			defer randMutex.Unlock()
			return rg.Intn(bound)
		}
		for round := 1; round <= 16; round++ {
			t.Logf("Start of round %d", round)
			lmc := NewLockedMapToComparable(nil,
				NewPrimitiveMapToComparable[int, int]())
			N := 24
			goChan := make(chan sets.Empty, N)
			allowed := sets.New[Pair[int, int]]()
			var allowedLock sync.Mutex
			var wgTrigger, wgDone sync.WaitGroup
			wgTrigger.Add(N)
			wgDone.Add(N)
			for i := 0; i < N; i++ {
				go func() {
					kv := NewPair(intn(N), intn(N))
					add := intn(12) < 8
					wgTrigger.Done()
					<-goChan
					if add {
						lmc.Put(kv.First, kv.Second)
					} else {
						lmc.Delete(kv.First)
					}
					if add {
						allowedLock.Lock()
						allowed.Insert(kv)
						allowedLock.Unlock()
					}
					wgDone.Done()
				}()
			}
			wgTrigger.Wait()
			close(goChan)
			wgDone.Wait()
			mtcCheckConsistency(t, lmc)
			lmc.Iterate2(func(key, val int) error {
				kv := NewPair(key, val)
				if !allowed.Has(kv) {
					t.Errorf("Found disallowed entry %v", kv)
				}
				return nil
			})
		}
	})
}
func mtcCheckConsistency(t *testing.T, mtc MutableMapToComparable[int, int]) {
	inverse := mtc.ReadInverse()
	mtc.Iterate2(func(key, val int) error {
		inverse.ContGet(val, func(keys sets.Set[int]) {
			if !keys.Has(key) {
				t.Errorf("Forward entry %v:%v not found in reverse", key, val)
			}
		})
		return nil
	})
	inverse.Iterate2(func(val int, keys sets.Set[int]) error {
		for key := range keys {
			fwdVal, have := mtc.Get(key)
			if !have || fwdVal != val {
				t.Errorf("Reverse entry %v:%v mismatches forward result %v,%v", key, val, fwdVal, have)
			}
		}
		return nil
	})
}
