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
	// AddConsumer notifies the consumer of the current contents
	// of the map and following changes.
	// Note that if this producer is safe for concurrent use then
	// the consumer can not expect to be able to call Get.
	AddConsumer(func(Key, Val))

	// Get queries the current map state, and does not count
	// as notifying any consumer.
	Get(Key) Val
}

type DynamicMapConsumer[Key comparable, Val any] interface {
	Set(Key, Val)
}

type Client[T any] interface {
	SetProvider(T)
}
