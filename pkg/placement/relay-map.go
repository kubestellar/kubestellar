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
)

// This file defines an implementation of DynamicMapProducer that is simply
// told what to put in the map.  It allows the values returned to the client
// to contain a projection of the values in the map.

type relayMap[Key comparable, OuterVal any, InnerVal any] struct {
	project func(OuterVal) InnerVal
	sync.Mutex
	theMap    map[Key]OuterVal
	consumers []func(Key, InnerVal)
}

func NewRelayMap[Key comparable, Val any]() DynamicMapProducerWithDelete[Key, Val] {
	return NewRelayAndProjectMap[Key, Val, Val](func(x Val) Val { return x })
}

func NewRelayAndProjectMap[Key comparable, OuterVal any, InnerVal any](project func(OuterVal) InnerVal) DynamicMapProducerWithDelete[Key, InnerVal] {
	return &relayMap[Key, OuterVal, InnerVal]{
		project: project,
		theMap:  map[Key]OuterVal{},
	}
}

func (rm *relayMap[Key, OuterVal, InnerVal]) Get(key Key, kont func(InnerVal)) {
	rm.Lock()
	defer rm.Unlock()
	kont(rm.project(rm.theMap[key]))
}

func (rm *relayMap[Key, OuterVal, InnerVal]) MaybeDelete(key Key, shouldDelete func(InnerVal) bool) {
	rm.Lock()
	defer rm.Unlock()
	innerVal := rm.project(rm.theMap[key])
	if shouldDelete(innerVal) {
		delete(rm.theMap, key)
	}
	for _, consumer := range rm.consumers {
		consumer(key, innerVal)
	}
}

func (rm *relayMap[Key, OuterVal, InnerVal]) AddConsumer(consumer func(Key, InnerVal)) {
	rm.Lock()
	defer rm.Unlock()
	rm.consumers = append(rm.consumers, consumer)
	for key, outerVal := range rm.theMap {
		consumer(key, rm.project(outerVal))
	}
}

func (rm *relayMap[Key, OuterVal, InnerVal]) Set(key Key, outerVal OuterVal) {
	innerVal := rm.project(outerVal)
	rm.Lock()
	defer rm.Unlock()
	rm.theMap[key] = outerVal
	for _, consumer := range rm.consumers {
		consumer(key, innerVal)
	}
}

func (rm *relayMap[Key, OuterVal, InnerVal]) Remove(key Key) {
	rm.MaybeDelete(key, func(InnerVal) bool { return true })
}
