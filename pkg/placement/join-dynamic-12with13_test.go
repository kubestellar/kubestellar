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

func TestDynamicJoin12with13(t *testing.T) {
	logger := klog.Background()
	for super := 1; super <= 10; super++ {
		t.Logf("Starting super=%d", super)
		yzReceiver := NewMapRelation2[string, float32]()
		xyReceiver, xzReceiver := NewDynamicJoin12with13[int, string, float32](logger, yzReceiver)
		xyReln := NewMapRelation2[int, string]()
		xzReln := NewMapRelation2[int, float32]()
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
				pair := Pair[int, float32]{rand.Intn(3), float32(1+rand.Intn(3)) / 4.0}
				switch rand.Intn(2) {
				case 0:
					change = xzReln.Add(pair)
					xzReceiver.Add(pair)
					t.Logf("Added yz=%v, change=%v", pair, change)
				case 1:
					change = xzReln.Remove(pair)
					xzReceiver.Remove(pair)
					t.Logf("Removed yz=%v, change=%v", pair, change)
				}
			}
			yzReln := Relation2Equijoin12with13[int, string, float32](xyReln, xzReln)
			if SetEqual[Pair[string, float32]](yzReceiver, yzReln) {
				t.Logf("At super=%d, iteration=%d, got correct xz=%v", super, iteration, xzReln)
			} else {
				t.Fatalf("At super=%d, iteration=%d, got %v, expected %v", super, iteration, xzReceiver, xzReln)
			}
		}
	}
}
