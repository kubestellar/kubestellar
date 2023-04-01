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

// This file defines some generic types that embody common patterns.

// These types use the following terminology conventions.
//
// "ThingProducer" actively produces a series of Things.
// "ThingSupplier" passively produces a series of Things.
// "ThingConsumer" actively pulls for things.
// "ThingReceiver" passively receives Things.
// "ThingProvider" provides the Thing service, which may have active and/or passive aspects.
// "Client[Thing]" is a user of an implementation of Thing
// "Mapping" is a single key-value pair or other such association
// "Map" is a complete set of pairs

// DynamicMapProvider holds a mutable map and keeps clients appraised of it.
// The zero value of Val signals a missing entry in the map.
type DynamicMapProvider[Key comparable, Val any] interface {
	// AddReceiver causes the given receiver to be notified of following
	// changes and, if notifyCurrent, the current map contents.
	// If receivers are comparable: depending on the implementation,
	// successive adds of the same receiver have no more effect than
	// the first add, or lead to duplicated callbacks to the receiver.
	// The producer precedes each receiver in the locking order.
	AddReceiver(receiver MappingReceiver[Key, Val], notifyCurrent bool)

	// Get invokes the given function on the value corresponding to the key.
	// This does not count as informing any particular receiver.
	// The producer precedes the function in the locking order.
	Get(Key, func(Val))
}

// DynamicMapProviderGet does a non-CPS Get.
// In situations with concurrency, regular clients (as in AddReceiver)
// get timing splinters if they use this.
func DynamicMapProviderGet[Key comparable, Val any](prod DynamicMapProvider[Key, Val], key Key) Val {
	var ans Val
	prod.Get(key, func(val Val) { ans = val })
	return ans
}

type DynamicMapProviderWithRelease[Key comparable, Val any] interface {
	DynamicMapProvider[Key, Val]

	// MaybeRelease invokes the given function on the value corresponding to the
	// given key and, if the function returns true, may release some internal resources
	// associated with that key.
	// The producer precedes the given function in the locking order.
	MaybeRelease(Key, func(Val) bool)
}

func DynamicMapProviderRelease[Key comparable, Val any](prod DynamicMapProviderWithRelease[Key, Val], key Key) {
	prod.MaybeRelease(key, func(Val) bool { return true })
}

// Map is a finite set of (key,value) pairs
// that has at most one value for any given key.
// The collection may or may not be mutable.
// This view of the collection may or may not have a limited scope of validity.
// This view may or may not have concurrency restrictions.
type Map[Key comparable, Val any] interface {
	Emptyable
	Len() int
	LenIsCheap() bool
	Get(Key) (Val, bool)
	Visitable[Pair[Key, Val]]
}

// MutableMap is a Map that can be written to.
type MutableMap[Key comparable, Val any] interface {
	Map[Key, Val]
	MappingReceiver[Key, Val]
}

// MappingReceiver is something that can be given key/value pairs.
// This is the writable aspect of a Map.
// Some DynamicMapProvider implementations require receivers to be comparable.
type MappingReceiver[Key comparable, Val any] interface {
	Put(Key, Val)
	Delete(Key)
}

// MappingReceiverFunc is a func value that implements MappingReceiver.
// Remember that func values are not comparable.
type MappingReceiverFunc[Key comparable, Val any] func(Key, Val)

func (cf MappingReceiverFunc[Key, Val]) Set(key Key, val Val) { cf(key, val) }

type TransformSetChangeReceiver[Type1 comparable, Type2 comparable] struct {
	Transform func(Type1) Type2
	Inner     SetChangeReceiver[Type2]
}

var _ SetChangeReceiver[int] = &TransformSetChangeReceiver[int, string]{}

func (xr TransformSetChangeReceiver[Type1, Type2]) Add(v1 Type1) bool {
	v2 := xr.Transform(v1)
	return xr.Inner.Add(v2)
}

func (xr TransformSetChangeReceiver[Type1, Type2]) Remove(v1 Type1) bool {
	v2 := xr.Transform(v1)
	return xr.Inner.Remove(v2)
}

type Client[T any] interface {
	SetProvider(T)
}

type DynamicValueProvider[Val any] interface {
	AddReceiver(Receiver[Val])
	Get(func(Val))
}

type Receiver[Val any] interface {
	Receive(Val)
}

type Empty struct{}
