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

// relayMap is both a mapping receiver and a map provider.
// It caches what it has been given in a local map, and provides that map.
// It can be configured at creation time with a transform from
// the values it is given to the values that it provides.
// It can be configured at creation time to deduplicate receivers.
type relayMap[Key comparable, OuterVal any, InnerVal any] struct {
	dedupReceivers bool
	transform      func(OuterVal) InnerVal
	sync.Mutex
	theMap    map[Key]OuterVal
	receivers []MappingReceiver[Key, InnerVal]
}

type TransformingRelayMap[Key comparable, OuterVal any, InnerVal any] interface {
	MappingReceiver[Key, OuterVal]
	DynamicMapProviderWithRelease[Key, InnerVal]
	Len() int
	Remove(Key)
}

type RelayMap[Key comparable, Val any] TransformingRelayMap[Key, Val, Val]

var _ DynamicMapProviderWithRelease[string, func()] = &relayMap[string, int64, func()]{}

func NewRelayMap[Key comparable, Val any](dedupReceivers bool) RelayMap[Key, Val] {
	return NewRelayAndProjectMap[Key, Val, Val](dedupReceivers, func(x Val) Val { return x })
}

func NewRelayAndProjectMap[Key comparable, OuterVal any, InnerVal any](dedupReceivers bool, transform func(OuterVal) InnerVal) TransformingRelayMap[Key, OuterVal, InnerVal] {
	return &relayMap[Key, OuterVal, InnerVal]{
		dedupReceivers: dedupReceivers,
		transform:      transform,
		theMap:         map[Key]OuterVal{},
	}
}

func (rm *relayMap[Key, OuterVal, InnerVal]) Get(key Key, kont func(InnerVal)) {
	rm.Lock()
	defer rm.Unlock()
	kont(rm.transform(rm.theMap[key]))
}

func (rm *relayMap[Key, OuterVal, InnerVal]) MaybeRelease(key Key, shouldRelease func(InnerVal) bool) {
	rm.Lock()
	defer rm.Unlock()
	innerVal := rm.transform(rm.theMap[key])
	if !shouldRelease(innerVal) {
		return
	}
	delete(rm.theMap, key)
	var outerVal OuterVal
	innerVal = rm.transform(outerVal)
	for _, receiver := range rm.receivers {
		receiver.Set(key, innerVal)
	}
}

func (rm *relayMap[Key, OuterVal, InnerVal]) AddReceiver(receiver MappingReceiver[Key, InnerVal], notifyCurrent bool) {
	rm.Lock()
	defer rm.Unlock()
	if rm.dedupReceivers {
		for _, existing := range rm.receivers {
			if existing == receiver {
				return
			}
		}
	}
	rm.receivers = append(rm.receivers, receiver)
	if !notifyCurrent {
		return
	}
	for key, outerVal := range rm.theMap {
		receiver.Set(key, rm.transform(outerVal))
	}
}

func (rm *relayMap[Key, OuterVal, InnerVal]) Set(key Key, outerVal OuterVal) {
	innerVal := rm.transform(outerVal)
	rm.Lock()
	defer rm.Unlock()
	rm.theMap[key] = outerVal
	for _, receiver := range rm.receivers {
		receiver.Set(key, innerVal)
	}
}

func (rm *relayMap[Key, OuterVal, InnerVal]) Len() int {
	return len(rm.theMap)
}

func (rm *relayMap[Key, OuterVal, InnerVal]) Remove(key Key) {
	rm.MaybeRelease(key, func(InnerVal) bool { return true })
}
