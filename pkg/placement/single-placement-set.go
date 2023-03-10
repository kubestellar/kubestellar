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
	apimachtypes "k8s.io/apimachinery/pkg/types"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

type SinglePlacementDetails struct {
	LocationName  string
	SyncTargetUID apimachtypes.UID
}

func SPDetails(sp edgeapi.SinglePlacement) SinglePlacementDetails {
	return SinglePlacementDetails{LocationName: sp.LocationName, SyncTargetUID: sp.SyncTargetUID}
}

func (spd SinglePlacementDetails) Complete(en ExternalName) edgeapi.SinglePlacement {
	return edgeapi.SinglePlacement{
		Cluster:        en.Cluster.String(),
		LocationName:   spd.LocationName,
		SyncTargetName: en.Name,
		SyncTargetUID:  spd.SyncTargetUID,
	}
}

// SinglePlacementSet is an alaternate representation of "resolved where";
// it maps ID of SyncTarget to the rest of the information
// in each edgeapi.SinglePlacement.
type SinglePlacementSet map[ExternalName]SinglePlacementDetails

var _ SetChangeReceiver[edgeapi.SinglePlacement] = NewSinglePlacementSet()

func NewSinglePlacementSet() SinglePlacementSet {
	return SinglePlacementSet{}
}

func (sps SinglePlacementSet) Add(sp edgeapi.SinglePlacement) {
	key := ExternalName{}.OfSPTarget(sp)
	sps[key] = SPDetails(sp)
}

func (sps SinglePlacementSet) Remove(sp edgeapi.SinglePlacement) {
	key := ExternalName{}.OfSPTarget(sp)
	delete(sps, key)
}

func (sps SinglePlacementSet) Equals(other SinglePlacementSet) bool {
	if len(sps) != len(other) {
		return false
	}
	for key, val := range sps {
		if val != other[key] {
			return false
		}
	}
	return true
}

func (sps SinglePlacementSet) Sub(other SinglePlacementSet) SinglePlacementSet {
	ans := NewSinglePlacementSet()
	for key, val := range sps {
		if other[key] != val {
			ans[key] = val
		}
	}
	return ans
}
