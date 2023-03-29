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

type MapSet[Elt comparable] map[Elt]Empty

func NewMapSet[Elt comparable](elts ...Elt) MapSet[Elt] {
	ans := MapSet[Elt]{}
	for _, elt := range elts {
		ans.Add(elt)
	}
	return ans
}

func (ms MapSet[Elt]) IsEmpty() bool { return len(ms) == 0 }

func (ms MapSet[Elt]) Has(elt Elt) bool {
	_, has := ms[elt]
	return has
}

func (ms MapSet[Elt]) Visit(visitor func(Elt) error) error {
	for element := range ms {
		if err := visitor(element); err != nil {
			return err
		}
	}
	return nil
}

func (ms MapSet[Elt]) Add(elt Elt) bool /* change */ {
	if _, has := ms[elt]; !has {
		ms[elt] = Empty{}
		return true
	}
	return false
}

func (ms MapSet[Elt]) Remove(elt Elt) bool /* change */ {
	if _, has := ms[elt]; has {
		delete(ms, elt)
		return true
	}
	return false
}
