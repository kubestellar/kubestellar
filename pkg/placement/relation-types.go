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

type Relation2[First, Second comparable] interface {
	Set[Pair[First, Second]]
	GetIndex1to2() Index2[First, Second]
}

type MutableRelation2[First, Second comparable] interface {
	Relation2[First, Second]
	MutableSet[Pair[First, Second]]
}

type Index2[Key, Val comparable] interface {
	Map[Key, Set[Val]]
	Visit1to2(Key, func(Val) error) error
}

type MutableIndex2[Key, Val comparable] interface {
	Index2[Key, Val]
	Add(Key, Val) bool
	Remove(Key, Val) bool
}

type relation2WithObservers[First, Second comparable] struct {
	Relation2[First, Second]
	inner     MutableRelation2[First, Second]
	observers []SetChangeReceiver[Pair[First, Second]]
}

var _ MutableRelation2[int, string] = &relation2WithObservers[int, string]{}

func (rwo *relation2WithObservers[First, Second]) Add(tup Pair[First, Second]) bool {
	if rwo.inner.Add(tup) {
		for _, observer := range rwo.observers {
			observer.Add(tup)
		}
		return true
	}
	return false
}

func (rwo *relation2WithObservers[First, Second]) Remove(tup Pair[First, Second]) bool {
	if rwo.inner.Remove(tup) {
		for _, observer := range rwo.observers {
			observer.Remove(tup)
		}
		return true
	}
	return false
}
