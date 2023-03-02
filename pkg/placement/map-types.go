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

// DynamicMapProducer holds a mutable map and keeps consumers appraised of it.
// The zero value of Val signals a missing entry in the map.
type DynamicMapProducer[Key comparable, Val any] interface {
	// AddConsumer causes the given consumer to be notified of following
	// changes and, if notifyCurrent, the current map contents.
	// If consumers are comparable: depending on the implementation,
	// successive adds of the same consumer have no more effect than
	// the first add, or lead to duplicated callbacks to the consumer.
	// The producer precedes each consumer in the locking order.
	AddConsumer(consumer DynamicMapConsumer[Key, Val], notifyCurrent bool)

	// Get invokes the given function on the value corresponding to the key.
	// This does not count as informing any particular consumer.
	// The producer precedes the function in the locking order.
	Get(Key, func(Val))
}

// DynamicMapProducerGet does a non-CPS Get.
// In situations with concurrency, regular consumers (as in AddConsumer)
// get timing splinters if they use this.
func DynamicMapProducerGet[Key comparable, Val any](prod DynamicMapProducer[Key, Val], key Key) Val {
	var ans Val
	prod.Get(key, func(val Val) { ans = val })
	return ans
}

type DynamicMapProducerWithRelease[Key comparable, Val any] interface {
	DynamicMapProducer[Key, Val]

	// MaybeRelease invokes the given function on the value corresponding to the
	// given key and, if the function returns true, may release some internal resources
	// associated with that key.
	// The producer precedes the given function in the locking order.
	MaybeRelease(Key, func(Val) bool)
}

// DynamicMapConsumer is given map entries by a DynamicMapProducer.
// Some DynamicMapProducer implementations require consumers to be comparable.
type DynamicMapConsumer[Key comparable, Val any] interface {
	Set(Key, Val)
}

// DynamicMapConsumerFunc is a func value that implements DynamicMapConsumer.
// Remember that func values are not comparable.
type DynamicMapConsumerFunc[Key comparable, Val any] func(Key, Val)

func (cf DynamicMapConsumerFunc[Key, Val]) Set(key Key, val Val) { cf(key, val) }

type Client[T any] interface {
	SetProvider(T)
}

type DynamicValueProducer[Val any] interface {
	AddConsumer(func(Val))
	Get() Val
}

type Empty struct{}
