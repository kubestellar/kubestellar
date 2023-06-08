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

package vvrv

import (
	"strconv"
	"sync"
	"time"

	data "github.com/kcp-dev/edge-mc/pkg/placement"
)

func NewResourceVersionVector() *RVV {
	return &RVV{
		next:    1,
		vv:      data.NewMapOverlay[string, string](),
		recents: data.NewMapRecent[string, data.Map[string, string]](time.Second*10, 30, time.Now),
	}
}

// RVV maintains a vector of ResourceVersion and maps each one
// to a ResourceVersion that identifies the vector.
// This map to vectors has limited memory.
type RVV struct {
	sync.Mutex
	next    int64
	vv      data.MapFunctional[string, string]
	recents data.MutableMap[string, data.Map[string, string]]
}

// Update changes one entry in the vector and returns the ResourceVersion
// that identifies the updated vector.
func (vr *RVV) Update(subId, subRV string) string {
	vr.Lock()
	defer vr.Unlock()
	vr.vv = vr.vv.Put(subId, subRV)
	ans := strconv.FormatInt(vr.next, 10)
	vr.next++
	vr.recents.Put(ans, vr.vv)
	return ans
}

// Get returns the associated vector, if one is still available.
func (vr *RVV) Get(rv string) (data.Map[string, string], bool) {
	vr.Lock()
	defer vr.Unlock()
	ans, have := vr.recents.Get(rv)
	return ans, have
}
