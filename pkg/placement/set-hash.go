/*
Copyright 2023 The KubeStellar Authors.

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

// NewHashSet creates a new set based on a hash map
func NewHashSet[Elt any](domain HashDomain[Elt], elts ...Elt) MutableSet[Elt] {
	theMap := NewHashMap[Elt, Empty](domain)(nil)
	ans := NewSetByMapToEmpty(theMap)
	for _, elt := range elts {
		ans.Add(elt)
	}
	return ans
}

func HashSetCopy[Elt any](domain HashDomain[Elt]) func(Visitable[Elt]) MutableSet[Elt] {
	return func(source Visitable[Elt]) MutableSet[Elt] {
		ans := NewHashSet(domain)
		source.Visit(func(elt Elt) error {
			ans.Add(elt)
			return nil
		})
		return ans
	}
}
