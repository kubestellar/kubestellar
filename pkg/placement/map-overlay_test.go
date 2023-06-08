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
)

type mapCheck[Key, Value any] struct{ expected, actual Map[Key, Value] }

func ExerciseMapFunctional(t *testing.T, New func() MapFunctional[int, int]) {
	for round := 1; round <= 10; round++ {
		expected := NewMapMap[int, int](nil)
		actual := New()
		checks := []mapCheck[int, int]{{expected, actual}}
		for iter := 1; iter <= 100; iter++ {
			expected = MapMapCopy[int, int](nil, expected)
			if expected.Len() == 0 || rand.Intn(3) != 1 {
				key := rand.Intn(50)
				val := rand.Intn(1000000)
				expected.Put(key, val)
				actual = actual.Put(key, val)
			} else {
				index := rand.Intn(expected.Len())
				goner, _ := VisitableGet[Pair[int, int]](expected, index)
				expected.Delete(goner.First)
				actual = actual.Delete(goner.First)
			}
			checks = append(checks, mapCheck[int, int]{expected, actual})
			for _, check := range checks {
				if MapEqual(check.expected, check.actual) {
					t.Logf("Success in round %v, iteration %v: expected=%v, actual=%v", round, iter, check.expected, check.actual)
				} else {
					t.Errorf("Failure in round %v, iteration %v: expected=%v, actual=%v", round, iter, check.expected, check.actual)
					MapEqual(check.expected, check.actual)
				}
			}
		}
	}
}

func TestMapOverlay(t *testing.T) {
	ExerciseMapFunctional(t, NewMapOverlay[int, int])
}
