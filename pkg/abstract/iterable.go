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

// Iterable2 is something that can iterate over some collection of key,value pairs
type Iterable2[Key, Val any] interface {
	Iterator2() Iterator2[Key, Val]
}

// Iterator2 is a function for iterating over some collection of key,value pairs.
// It calls yield on the pairs in some sequence,
// stopping early if yield returns `false`.
// This is intended to align with what I see now in https://go.dev/wiki/RangefuncExperiment .
type Iterator2[Key, Val any] func(yield func(Key, Val) bool)
