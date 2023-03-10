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
	"sync"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

// SimplePlacementSliceSetReducer is the simplest possible
// set differencer for ResolvedWhere
type SimplePlacementSliceSetReducer struct {
	sync.Mutex
	receivers []SetChangeReceiver[edgeapi.SinglePlacement]
	current   SinglePlacementSet
}

var _ Receiver[ResolvedWhere] = &SimplePlacementSliceSetReducer{}

var _ ResolvedWhereDifferencerConstructor = NewSimplePlacementSliceSetReducer

func NewSimplePlacementSliceSetReducer(receiver SetChangeReceiver[edgeapi.SinglePlacement]) Receiver[ResolvedWhere] {
	ans := &SimplePlacementSliceSetReducer{
		receivers: []SetChangeReceiver[edgeapi.SinglePlacement]{receiver},
		current:   NewSinglePlacementSet(),
	}
	return ans
}

func (spsr *SimplePlacementSliceSetReducer) Receive(newSlices ResolvedWhere) {
	spsr.Lock()
	defer spsr.Unlock()
	for key, val := range spsr.current {
		sp := val.Complete(key)
		for _, receiver := range spsr.receivers {
			receiver.Remove(sp)
		}
	}
	spsr.setLocked(newSlices)
}

func (spsr *SimplePlacementSliceSetReducer) setLocked(newSlices ResolvedWhere) {
	spsr.current = NewSinglePlacementSet()
	enumerateSinglePlacementSlices(newSlices, func(apiSP edgeapi.SinglePlacement) {
		syncTargetID := ExternalName{}.OfSPTarget(apiSP)
		syncTargetDetails := SPDetails(apiSP)
		spsr.current[syncTargetID] = syncTargetDetails
		for _, receiver := range spsr.receivers {
			receiver.Add(apiSP)
		}
	})
}

func enumerateSinglePlacementSlices(slices []*edgeapi.SinglePlacementSlice, receiver func(edgeapi.SinglePlacement)) {
	for _, slice := range slices {
		if slice != nil {
			for _, sp := range slice.Destinations {
				receiver(sp)
			}
		}
	}
}
