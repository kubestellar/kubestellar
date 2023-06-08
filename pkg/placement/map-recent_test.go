/*
Copyright 2023 The KubeStellar Authors.

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
	"testing"
	"time"
)

func TestMapRecent(t *testing.T) {
	for round := 1; round <= 10; round++ {
		expect := []Pair[int, int]{}
		current := NewMapRecent[int, int](time.Second, 10, time.Now)
		for iter := 1; iter <= 100; iter++ {
			key := rand.Intn(50)
			val := rand.Intn(1000000)
			current.Put(key, val)
			if len(expect) >= 10 {
				expect = expect[:9]
			}
			expect = append([]Pair[int, int]{NewPair(key, val)}, expect...)
			checked := NewMapSet[int]()
			for _, pair := range expect {
				if checked.Has(pair.First) {
					continue
				}
				got, has := current.Get(pair.First)
				if !has {
					t.Errorf("Failure at round %v, iteration %v: current=%v, lacks entry for %v", round, iter, current, pair.First)
				} else if got != pair.Second {
					t.Errorf("Failure at round %v, iteration %v: current=%v, lacks entry for %v is %v but expected %v", round, iter, current, pair.First, got, pair.Second)
				}
				checked.Add(pair.First)
			}
		}
	}
}
