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
	"testing"
)

func TestMapSetDifferencer(t *testing.T) {
	genSet := func() MapSet[int] {
		ans := NewMapSet[int]()
		for count := 1; count < 10; count++ {
			ans.Add(rand.Intn(20))
		}
		return ans
	}
	reportedGone := NewMapSet[int]()
	reportedNew := NewMapSet[int]()
	reportToGone := SetWriterReverse[int](reportedGone)
	checkingReceiver := SetWriterFork[int](false, reportedNew, reportToGone)
	differ := NewSetDifferenceByMapAndEnum[MapSet[int], int](MapSetAsVisitable[int], checkingReceiver)
	current := NewMapSet[int]()
	for iteration := 1; iteration < 100; iteration++ {
		next := genSet()
		differ.Receive(next)
		expectedGone, _, expectedNew := MapSetSymmetricDifference[int](true, false, true, current, next)
		if !SetEqual[int](expectedGone, reportedGone) {
			t.Fatalf("At iteration %d, wrong goners.  current=%v, next=%v, expected=%v, reported=%v", iteration, current, next, expectedGone, reportedGone)
		}
		if !SetEqual[int](expectedNew, reportedNew) {
			t.Fatalf("At iteration %d, wrong new.  current=%v, next=%v, expected=%v, reported=%v", iteration, current, next, expectedNew, reportedNew)
		}
		current = next
		SetRemoveAll[int](reportedGone, reportedGone)
		SetRemoveAll[int](reportedNew, reportedNew)
	}
}
