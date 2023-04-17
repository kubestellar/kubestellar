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
	"sort"
	"testing"
)

// This tests NewFactoredMapMapAggregator.
// The outer map is from uint16 to float32.
// The factorer decomposes the uint16 into two uint8.
// The aggregation function takes a map uint8->float32 and returns
// a sorted slice of the sum of the uint8 and the float32.
func TestAggregate(t *testing.T) {
	factorer := NewFactorer(
		func(whole uint16) Pair[uint8, uint8] { return NewPair(uint8(whole/10), uint8(whole%10)) },
		func(parts Pair[uint8, uint8]) uint16 { return uint16(parts.First)*256 + uint16(parts.Second) })
	valmap := func(right uint8, val float32) float64 { return float64(right) + float64(val) }
	solve := func(left uint8, problem Map[uint8, float32]) []float64 {
		slice := MapTransformToSlice(problem, valmap)
		sort.Float64s(slice)
		return slice
	}
	for super := 1; super <= 10; super++ {
		got := NewMapMap[uint8, []float64](nil)
		fm := NewFactoredMapMapAggregator[uint16, uint8, uint8, float32, []float64](
			factorer,
			nil,
			nil,
			solve,
			got,
		)
		aggregate := func(input Map[uint16, float32]) Map[uint8, []float64] {
			ans := NewMapMap[uint8, []float64](nil)
			input.Visit(func(tup Pair[uint16, float32]) error {
				parts := factorer.First(tup.First)
				inner, _ := ans.Get(parts.First)
				inner = append(inner, valmap(parts.Second, tup.Second))
				sort.Float64s(inner)
				ans.Put(parts.First, inner)
				return nil
			})
			return ans
		}
		wholeMap := NewMapMap[uint16, float32](nil)
		for iteration := 1; iteration <= 100; iteration++ {
			growness := 70 - iteration/2
			if wholeMap.Len() == 0 || rand.Intn(100) < growness {
				key := uint16(rand.Intn(100))
				val := float32(rand.Intn(63)+1) / 64
				wholeMap.Put(key, val)
				fm.Put(key, val)
			} else {
				goner, _ := VisitableGet[Pair[uint16, float32]](wholeMap, rand.Intn(wholeMap.Len()))
				wholeMap.Delete(goner.First)
				fm.Delete(goner.First)
			}
			expected := aggregate(wholeMap)
			if !MapEqualParametric[uint8, []float64](SliceEqual[float64])(expected, got) {
				t.Fatalf("At super=%d iteration=%d, expected %v but got %v", super, iteration, expected, got)
			}
		}
	}
}
