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

// MapRelation2 is a 2-ary relation represented by an index on the first column.
// It is mutable.
// It is not safe for concurrent access.
type MapRelation2[First comparable, Second comparable] struct {
	// by1 MutableIndex2[First, Second]
	rep MutableMap[First, MutableSet[Second]]
	GenericSetIndex[Pair[First, Second], First, Second]
}

var _ MutableRelation2[string, float64] = &MapRelation2[string, float64]{}

func NewMapRelation2[First comparable, Second comparable](pairs ...Pair[First, Second]) *MapRelation2[First, Second] {
	rep := NewMapMap[First, MutableSet[Second]](nil)
	wholeSet := NewGenericSetIndex[Pair[First, Second], First, Second](
		nil,
		PairFactorer[First, Second](),
		func() MutableSet[Second] { return NewEmptyMapSet[Second]() },
		rep,
	)
	ans := &MapRelation2[First, Second]{
		rep:             rep,
		GenericSetIndex: wholeSet,
	}
	for _, pair := range pairs {
		ans.Add(pair)
	}
	return ans
}

func Relation2WithObservers[First, Second comparable](inner MutableRelation2[First, Second], observers ...SetChangeReceiver[Pair[First, Second]]) MutableRelation2[First, Second] {
	return &relation2WithObservers[First, Second]{inner, inner, observers}
}
