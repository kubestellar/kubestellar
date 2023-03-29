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
	"testing"

	"k8s.io/klog/v2"
)

func TestDynamicJoin(t *testing.T) {
	logger := klog.Background()
	for super := 1; super <= 10; super++ {
		t.Logf("Starting super=%d", super)
		xzReceiver := NewMapRelation2[int, float32]()
		xyReceiver, yzReceiver := NewDynamicJoin[int, string, float32](logger, xzReceiver)
		xyReln := NewMapRelation2[int, string]()
		yzReln := NewMapRelation2[string, float32]()
		for iteration := 1; iteration <= 100; iteration++ {
			var change bool
			switch rand.Intn(2) {
			case 0:
				pair := Pair[int, string]{rand.Intn(3), fmt.Sprintf("%d", rand.Intn(3))}
				switch rand.Intn(2) {
				case 0:
					change = xyReln.Add(pair)
					xyReceiver.Add(pair)
					t.Logf("Added xy=%v, change=%v", pair, change)
				default:
					change = xyReln.Remove(pair)
					xyReceiver.Remove(pair)
					t.Logf("Removed xy=%v, change=%v", pair, change)
				}
			case 1:
				pair := Pair[string, float32]{fmt.Sprintf("%d", rand.Intn(3)), float32(1+rand.Intn(3)) / 4.0}
				switch rand.Intn(2) {
				case 0:
					change = yzReln.Add(pair)
					yzReceiver.Add(pair)
					t.Logf("Added yz=%v, change=%v", pair, change)
				case 1:
					change = yzReln.Remove(pair)
					yzReceiver.Remove(pair)
					t.Logf("Removed yz=%v, change=%v", pair, change)
				}
			}
			xzReln := Relation2Equijoin[int, string, float32](xyReln, yzReln)
			if Relation2Equal[int, float32](xzReceiver, xzReln) {
				t.Logf("At super=%d, iteration=%d, got correct xz=%v", super, iteration, xzReln)
			} else {
				t.Fatalf("At super=%d, iteration=%d, got %v, expected %v", super, iteration, xzReceiver, xzReln)
			}
		}
	}
}

func Relation2Equijoin[First, Second, Third comparable](left Relation2[First, Second], right Relation2[Second, Third]) Relation2[First, Third] {
	ans := NewMapRelation2[First, Third]()
	left.Visit(func(xy Pair[First, Second]) error {
		right.Visit1to2(xy.Second, func(z Third) error {
			ans.Add(Pair[First, Third]{xy.First, z})
			return nil
		})
		return nil
	})
	return ans
}
