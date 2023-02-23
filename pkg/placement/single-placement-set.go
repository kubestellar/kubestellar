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

func (sp SinglePlacement) Details() SinglePlacementDetails {
	return SinglePlacementDetails{LocationName: sp.Location.Name, SyncTargetUID: sp.SyncTargetUID}
}

func (spd SinglePlacementDetails) Complete(en edgeapi.ExternalName) SinglePlacement {
	return SinglePlacement{
		SinglePlacement: spd.ToAPI(en),
		SyncTargetUID:   spd.SyncTargetUID,
	}
}

func (spd SinglePlacementDetails) ToAPI(en edgeapi.ExternalName) edgeapi.SinglePlacement {
	return edgeapi.SinglePlacement{
		Location: edgeapi.ExternalName{
			Workspace: en.Workspace,
			Name:      spd.LocationName},
		SyncTargetName: en.Name,
	}
}

// SinglePlacementSet maps ID of SyncTarget to the rest of the information
// in each SinglePlacement.
type SinglePlacementSet map[edgeapi.ExternalName]SinglePlacementDetails

var _ SinglePlacementSetChangeConsumer = NewSinglePlacementSet()

func NewSinglePlacementSet() SinglePlacementSet {
	return SinglePlacementSet{}
}

func (sps SinglePlacementSet) Add(sp SinglePlacement) {
	key := sp.SyncTargetRef()
	sps[key] = sp.Details()
}

func (sps SinglePlacementSet) Remove(sp SinglePlacement) {
	key := sp.SyncTargetRef()
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
