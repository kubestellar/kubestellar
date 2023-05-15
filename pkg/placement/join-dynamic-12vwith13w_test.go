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

	"k8s.io/klog/v2"
)

func exerciseDynamicJoin12VWith13W[Col1, Col2, Col3, Val, Wal comparable](
	t *testing.T,
	logger klog.Logger,
	newLeftMap func() MutableMap[Pair[Col1, Col2], Val],
	newRightMap func() MutableMap[Pair[Col1, Col3], Wal],
	newLeftTuple func() Pair[Pair[Col1, Col2], Val],
	newRightTuple func() Pair[Pair[Col1, Col3], Wal],
	newJoint func() MutableMap[Triple[Col1, Col2, Col3], Pair[Val, Wal]],
	newDynamicJoin func(MappingReceiver[Triple[Col1, Col2, Col3], Pair[Val, Wal]]) (MappingReceiver[Pair[Col1, Col2], Val], MappingReceiver[Pair[Col1, Col3], Wal]),
) {
	for super := 1; super <= 10; super++ {
		t.Logf("Starting super=%d", super)
		// yzReceiver := NewMapRelation2[string, float32]()
		joinReceiver := newJoint()
		leftReceiver, rightReceiver := newDynamicJoin(joinReceiver)
		leftMap := newLeftMap()
		rightMap := newRightMap()
		for iteration := 1; iteration <= 100; iteration++ {
			switch rand.Intn(2) {
			case 0:
				switch {
				case leftMap.IsEmpty() || rand.Intn(2) == 0:
					leftTuple := newLeftTuple()
					leftReceiver.Put(leftTuple.First, leftTuple.Second)
					leftMap.Put(leftTuple.First, leftTuple.Second)
					t.Logf("Added on the left: %v", leftTuple)
				default:
					leftTuple, _ := VisitableGet[Pair[Pair[Col1, Col2], Val]](leftMap, rand.Intn(leftMap.Len()))
					leftReceiver.Delete(leftTuple.First)
					leftMap.Delete(leftTuple.First)
					t.Logf("Removed on the left: %v", leftTuple.First)
				}
			case 1:
				switch {
				case rightMap.IsEmpty() || rand.Intn(2) == 0:
					rightTuple := newRightTuple()
					rightReceiver.Put(rightTuple.First, rightTuple.Second)
					rightMap.Put(rightTuple.First, rightTuple.Second)
					t.Logf("Added on the right: %v", rightTuple)
				default:
					rightTuple, _ := VisitableGet[Pair[Pair[Col1, Col3], Wal]](rightMap, rand.Intn(rightMap.Len()))
					rightReceiver.Delete(rightTuple.First)
					rightMap.Delete(rightTuple.First)
					t.Logf("Removed on the right: %v", rightTuple.First)
				}
			}
			expectedJoin := newJoint()
			vj := JoinByVisitSquared[Pair[Pair[Col1, Col2], Val],
				Pair[Pair[Col1, Col3], Wal],
				Pair[Triple[Col1, Col2, Col3], Pair[Val, Wal]],
			](
				leftMap, rightMap,
				func(leftTuple Pair[Pair[Col1, Col2], Val], rightTuple Pair[Pair[Col1, Col3], Wal]) (Pair[Triple[Col1, Col2, Col3], Pair[Val, Wal]], bool) {
					if leftTuple.First.First == rightTuple.First.First {
						return Pair[Triple[Col1, Col2, Col3], Pair[Val, Wal]]{
							Triple[Col1, Col2, Col3]{leftTuple.First.First, leftTuple.First.Second, rightTuple.First.Second},
							Pair[Val, Wal]{leftTuple.Second, rightTuple.Second},
						}, true
					}
					return Pair[Triple[Col1, Col2, Col3], Pair[Val, Wal]]{}, false
				},
			)
			MapAddAll(expectedJoin, vj)
			MapEnumerateDifferences[Triple[Col1, Col2, Col3], Pair[Val, Wal]](expectedJoin, joinReceiver,
				MapChangeReceiverFuncs[Triple[Col1, Col2, Col3], Pair[Val, Wal]]{
					OnCreate: func(key Triple[Col1, Col2, Col3], val Pair[Val, Wal]) {
						t.Fatalf("At super=%d, iteration=%d, extra key %v, val %v", super, iteration, key, val)
					},
					OnUpdate: func(key Triple[Col1, Col2, Col3], goodVal, badVal Pair[Val, Wal]) {
						t.Fatalf("At super=%d, iteration=%d, key %v, expected val %v, got val %v", super, iteration, key, goodVal, badVal)
					},
					OnDelete: func(key Triple[Col1, Col2, Col3], val Pair[Val, Wal]) {
						t.Fatalf("At super=%d, iteration=%d, missing key %v, val %v", super, iteration, key, val)
					},
				})
		}
	}
}
