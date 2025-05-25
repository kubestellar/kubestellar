/*
Copyright 2024 The KubeStellar Authors.

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

package abstract

// Map abstracts reading from a map from Key to Val.
// The Key values must be comparable in some way
// (which is not exposed in this interface).
// The map may or may not vary over time.
// Access through this interface may or may not be thread-safe.
// The continuation passed to ContGet may enable access to
// the Val after return.
type Map[Key, Val any] interface {
	// Users of this map may retain Val values;
	// they are immutable in this shallow sense.
	MapToLocked[Key, Val]

	// Get queries the map with a given key.
	// If the map has a value for that key then that value
	// and `true` are returned; otherwise the zero value of Val
	// and `false` are returned.
	Get(Key) (Val, bool)
}

// MapToLocked is a map to values that may or may not be protected by a lock.
type MapToLocked[Key, Val any] interface {
	// Length returns the current number of entries in the map
	Length() int

	// Get queries the map with a given key.
	// If the map has a value for that key then the given func (a
	// "continuation"; hence the "Cont" in the name of this method) is called
	// with that value. The call is in the same goroutine and the func may not
	// make any calls on the map. The callee may not enable anything to access
	// the given Val after the callee returns.
	ContGet(Key, func(Val))

	// Iterate2 calls the given consumer with entries in the map,
	// sequentially in some order.
	// That is: (a) the first call to the consumer happens after the
	// call to (but not the return from) this method;
	// (b) each of the other calls to the consumer happens after
	// the return from the previous call to the consumer; and
	// (c) the return from this call happens after the return from
	// the last call to consumer.
	// The consumer may or may not be allowed to access the map
	// (comments on particular maps should say whether such access is allowed).
	// The consumer may or may not be allowed to enable access
	// to the Val after returning (commands on particular maps should say).
	// If the consumer ever returns a non-nil error then the iteration
	// halts and that error is returned.
	// Otherwise `nil` is returned.
	Iterate2(func(Key, Val) error) error
}

// MapWriter is an interface to a mutable map.
type MapWriter[Key, Val any] interface {
	// Put sets the mapping for the given key, and returns
	// the previous mapping (or zero) and a bool indicating
	// whether there already was a mapping for that key.
	Put(Key, Val) (Val, bool)

	// Delete removes the mapping for a given key if there was one.
	// Delete returns that mapping and true if there was one
	// otherwise zero and false.
	Delete(Key) (Val, bool)
}

// MutableMap is a Map with write methods.
type MutableMap[Key, Val any] interface {
	Map[Key, Val]
	MapWriter[Key, Val]
}
