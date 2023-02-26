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
	apimachtypes "k8s.io/apimachinery/pkg/types"
)

// SimplePlacementSliceSetReducer is the simplest possible
// implementation of SinglePlacementSliceSetReducer.
type SimplePlacementSliceSetReducer struct {
	uider UIDer
	sync.Mutex
	consumers []SinglePlacementSetChangeConsumer
	enhanced  SinglePlacementSet
}

var _ SinglePlacementSliceSetReducer = &SimplePlacementSliceSetReducer{}

func NewSimplePlacementSliceSetReducer(uider UIDer, consumers ...SinglePlacementSetChangeConsumer) *SimplePlacementSliceSetReducer {
	ans := &SimplePlacementSliceSetReducer{
		uider:     uider,
		consumers: consumers,
		enhanced:  NewSinglePlacementSet(),
	}
	uider.AddConsumer(ans.noteUID)
	return ans
}

func (spsr *SimplePlacementSliceSetReducer) Set(newSlices ResolvedWhere) {
	spsr.Lock()
	defer spsr.Unlock()
	for key, val := range spsr.enhanced {
		sp := val.Complete(key)
		for _, consumer := range spsr.consumers {
			consumer.Remove(sp)
		}
	}
	spsr.setLocked(newSlices)
}

func (spsr *SimplePlacementSliceSetReducer) setLocked(newSlices ResolvedWhere) {
	spsr.enhanced = NewSinglePlacementSet()
	enumerateSinglePlacementSlices(newSlices, func(apiSP edgeapi.SinglePlacement) {
		syncTargetID := edgeapi.ExternalName{Workspace: apiSP.Location.Workspace, Name: apiSP.SyncTargetName}
		syncTargetUID := spsr.uider.Get(syncTargetID)
		syncTargetDetails := SinglePlacementDetails{
			LocationName:  apiSP.Location.Name,
			SyncTargetUID: syncTargetUID,
		}
		spsr.enhanced[syncTargetID] = syncTargetDetails
		fullSP := SinglePlacement{SinglePlacement: apiSP, SyncTargetUID: syncTargetUID}
		for _, consumer := range spsr.consumers {
			consumer.Add(fullSP)
		}
	})
}

func (spsr *SimplePlacementSliceSetReducer) noteUID(en edgeapi.ExternalName, uid apimachtypes.UID) {
	spsr.Lock()
	defer spsr.Unlock()
	if details, ok := spsr.enhanced[en]; ok {
		details.SyncTargetUID = uid
		spsr.enhanced[en] = details
		fullSP := details.Complete(en)
		for _, consumer := range spsr.consumers {
			consumer.Add(fullSP)
		}

	}
}

func enumerateSinglePlacementSlices(slices []*edgeapi.SinglePlacementSlice, consumer func(edgeapi.SinglePlacement)) {
	for _, slice := range slices {
		if slice != nil {
			for _, sp := range slice.Destinations {
				consumer(sp)
			}
		}
	}
}
