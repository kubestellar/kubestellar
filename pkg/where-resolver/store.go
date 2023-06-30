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

package where_resolver

import (
	"sync"
)

/*
The relationship between relevant API objects (arrows indicate selections/references):
ep -> loc -> st
ep: edgeplacement; loc: location; st: synctarget

To facilitate incremental updates, we need to maintain internal data structures:

	ep <- loc <- st

or in other words,

	ep <- loc
	loc <- st

*/

type empty struct{}

type internalData struct {
	l                sync.Mutex
	epsBySelectedLoc map[string]map[string]empty // ep <- loc
	locsBySelectedSt map[string]map[string]empty // loc <- st
}

var store internalData

func init() {
	store = internalData{
		l:                sync.Mutex{},
		epsBySelectedLoc: map[string]map[string]empty{},
		locsBySelectedSt: map[string]map[string]empty{},
	}
}

func unionTwo(a, b map[string]empty) map[string]empty {
	u := map[string]empty{}
	for k, v := range a {
		u[k] = v
	}
	for k, v := range b {
		u[k] = v
	}
	return u
}

func (d *internalData) findEpsUsedSt(st string) map[string]empty {
	eps := map[string]empty{}
	locs := d.locsBySelectedSt[st]
	for l := range locs {
		eps = unionTwo(eps, d.epsBySelectedLoc[l])
	}
	return eps
}

func (d *internalData) dropEp(epKey string) {
	for _, eps := range d.epsBySelectedLoc {
		delete(eps, epKey)
	}
}

func (d *internalData) dropLoc(locKey string) {
	delete(d.epsBySelectedLoc, locKey)
	for _, locs := range d.locsBySelectedSt {
		delete(locs, locKey)
	}
}

func (d *internalData) dropSt(stKey string) {
	delete(d.locsBySelectedSt, stKey)
}
