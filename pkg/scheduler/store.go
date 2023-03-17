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

package scheduler

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
	ep <- st

*/

type internalData struct {
	l sync.Mutex

	// TODO(waltforme): should these map values be slices or maps?
	epsBySelectedLoc map[string][]string // ep <- loc
	locsBySelectedSt map[string][]string // loc <- st
	epsByUsedSt      map[string][]string // ep <- st
}

var store internalData

func init() {
	store = internalData{
		l:                sync.Mutex{},
		epsBySelectedLoc: map[string][]string{},
		locsBySelectedSt: map[string][]string{},
		epsByUsedSt:      map[string][]string{},
	}

	store.manuallyFillIn()
}
