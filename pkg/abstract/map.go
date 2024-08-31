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
type Map[Key, Val any] interface {
	// Length returns the current number of entries in the map
	Length() int

	// Get queries the map with a given key.
	// If the map has a value for that key then that value
	// and `true` are returned; otherwise the zero value of Val
	// and `false` are returned.
	Get(Key) (Val, bool)

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
	// If the consumer ever returns a non-nil error then the iteration
	// halts and that error is returned.
	// Otherwise `nil` is returned.
	Iterate2(func(Key, Val) error) error
}
