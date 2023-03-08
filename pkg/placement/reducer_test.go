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
	"time"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	apimachtypes "k8s.io/apimachinery/pkg/types"
)

// exerciseSinglePlacementSliceSetReducer tests a given SinglePlacementSliceSetReducer.
// For a given SinglePlacementSliceSetReducer implementation Foo,
// do the following in TestFoo (or whatever you call it):
// - construct a Foo using a SinglePlacementSet as the receiver
// - call exerciseSinglePlacementSliceSetReducer on the Foo and that receiver.
func exerciseSinglePlacementSliceSetReducer(rng *rand.Rand, initialWhere ResolvedWhere, iterations int, changesPerIteration int, extraPerIteration func(), reducer SinglePlacementSliceSetReducer, receiver SinglePlacementSet) func(*testing.T) {
	return func(t *testing.T) {
		input := initialWhere
		for iteration := 1; iteration <= iterations; iteration++ {
			prevInput := input
			for change := 1; change <= changesPerIteration; change++ {
				input = reviseSinglePlacementSliceSlice(rng, input)
			}
			reducer.Set(input)
			extraPerIteration()
			checker := NewSinglePlacementSet()
			reference := NewSimplePlacementSliceSetReducer(checker)
			reference.Set(input)
			if receiver.Equals(checker) {
				continue
			}
			t.Errorf("Unexpected result: excess=%v, missing=%v, iteration=%d, prevInput=%v, input=%v", receiver.Sub(checker), checker.Sub(receiver), iteration, prevInput, input)
		}
	}
}

func reviseSinglePlacementSliceSlice(rng *rand.Rand, slices ResolvedWhere) ResolvedWhere {
	ans := make(ResolvedWhere, 0, len(slices))
	copy(ans, slices)
	if len(ans) == 0 || rng.Intn(20) == 1 {
		// Add a new slice
		sliceLen := (rng.Intn(12) + 2) / 3
		slice := edgeapi.SinglePlacementSlice{Destinations: []edgeapi.SinglePlacement{}}
		for dest := 1; dest <= sliceLen; dest++ {
			slice.Destinations = append(slice.Destinations, genSinglePlacement(rng))
		}
		sliceIdx := rng.Intn(len(ans) + 1)
		ans = append(ans[:sliceIdx], append([]*edgeapi.SinglePlacementSlice{&slice}, ans[sliceIdx:]...)...)
	} else if rng.Intn(20) != 1 {
		// modify an existing slice
		sliceIdx := rng.Intn(len(ans))
		slice := *ans[sliceIdx]
		newDestinations := make([]edgeapi.SinglePlacement, 0, len(slice.Destinations))
		copy(newDestinations, slice.Destinations)
		if len(slice.Destinations) == 0 || rng.Intn(20) == 1 {
			// Add a new entry
			destIdx := rng.Intn(len(newDestinations) + 1)
			dest := genSinglePlacement(rng)
			newDestinations = append(newDestinations[:destIdx], append([]edgeapi.SinglePlacement{dest}, newDestinations[destIdx:]...)...)
		} else if rng.Intn(20) != 1 {
			// modify an existing SinglePlacement
			destIdx := rng.Intn(len(newDestinations))
			newDestinations[destIdx] = reviseSinglePlacement(rng, newDestinations[destIdx])
		} else {
			// delete an existing entry
			destIdx := rng.Intn(len(newDestinations))
			newDestinations = append(newDestinations[:destIdx], newDestinations[destIdx+1:]...)
		}
		slice.Destinations = newDestinations
	} else {
		// Delete an existing slice
		sliceIdx := rng.Intn(len(ans))
		ans = append(ans[:sliceIdx], ans[sliceIdx+1:]...)
	}
	return ans
}

func reviseSinglePlacement(rng *rand.Rand, sp edgeapi.SinglePlacement) edgeapi.SinglePlacement {
	switch rng.Intn(4) {
	case 0:
		sp.Cluster = fmt.Sprintf("ws%d", rng.Intn(1000))
	case 1:
		sp.LocationName = fmt.Sprintf("loc%d", rng.Intn(1000))
	case 2:
		sp.SyncTargetName = fmt.Sprintf("st%d", rng.Intn(1000))
	default:
		sp.SyncTargetUID = apimachtypes.UID(fmt.Sprintf("uid%d", rng.Intn(1000000)))
	}
	return sp
}

func genSinglePlacement(rng *rand.Rand) edgeapi.SinglePlacement {
	return edgeapi.SinglePlacement{
		Cluster:        fmt.Sprintf("ws%d", rng.Intn(1000)),
		LocationName:   fmt.Sprintf("loc%d", rng.Intn(1000)),
		SyncTargetName: fmt.Sprintf("st%d", rng.Intn(1000)),
		SyncTargetUID:  apimachtypes.UID(fmt.Sprintf("uid%d", rng.Intn(1000000))),
	}
}

func TestSimplePlacementSliceSetReducer(t *testing.T) {
	rs := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(rs)
	testReceiver := NewSinglePlacementSet()
	testReducer := NewSimplePlacementSliceSetReducer(testReceiver)
	sp1 := edgeapi.SinglePlacement{Cluster: "ws-a", LocationName: "loc-a",
		SyncTargetName: "st-a", SyncTargetUID: "u-a"}
	rw1 := ResolvedWhere{&edgeapi.SinglePlacementSlice{
		Destinations: []edgeapi.SinglePlacement{sp1},
	}}
	testReducer.Set(rw1)
	if actual, expected := len(testReceiver), 1; actual != expected {
		t.Errorf("Wrong size after first Set: actual=%d, expected=%d", actual, expected)
	}
	if actual, expected := testReceiver[ExternalName{}.OfSPTarget(sp1)], SPDetails(sp1); actual != expected {
		t.Errorf("Wrong details after first Set: actual=%#v, expected=%#v", actual, expected)
	}
	sp1a := sp1
	sp1a.SyncTargetUID = "u-aa"
	rw1a := ResolvedWhere{&edgeapi.SinglePlacementSlice{
		Destinations: []edgeapi.SinglePlacement{sp1a},
	}}
	testReducer.Set(rw1a)
	if actual, expected := len(testReceiver), 1; actual != expected {
		t.Errorf("Wrong size after first tweak: actual=%d, expected=%d", actual, expected)
	}
	if actual, expected := testReceiver[ExternalName{}.OfSPTarget(sp1)], SPDetails(sp1a); actual != expected {
		t.Errorf("Wrong details after first tweak: actual=%#v, expected=%#v", actual, expected)
	}
	exerciseSinglePlacementSliceSetReducer(rng, rw1, 20, 10, func() {}, testReducer, testReceiver)(t)
}
