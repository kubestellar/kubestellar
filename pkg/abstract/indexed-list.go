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

type IndexedList[Elt comparable] struct {
	Index map[Elt]int
	List  []Elt
}

func NewIndexedList[Elt comparable](elts ...Elt) *IndexedList[Elt] {
	sl := &IndexedList[Elt]{
		Index: make(map[Elt]int, len(elts)),
		List:  make([]Elt, 0, len(elts)),
	}
	for _, elt := range elts {
		sl.Insert(elt)
	}
	return sl
}

func (sl *IndexedList[Elt]) Len() int {
	return len(sl.List)
}

func (sl *IndexedList[Elt]) Has(elt Elt) bool {
	_, has := sl.Index[elt]
	return has
}

func (sl *IndexedList[Elt]) Insert(elt Elt) bool {
	if sl.Has(elt) {
		return false
	}
	sl.Index[elt] = len(sl.List)
	sl.List = append(sl.List, elt)
	return true
}

func (sl *IndexedList[Elt]) Delete(elt Elt) bool {
	idx, has := sl.Index[elt]
	if !has {
		return false
	}
	lastIdx := len(sl.List) - 1
	delete(sl.Index, elt)
	if idx != lastIdx {
		movedElt := sl.List[lastIdx]
		sl.List[idx] = movedElt
		sl.Index[movedElt] = idx
	}
	sl.List = sl.List[:lastIdx]
	return true
}

func (sl *IndexedList[Elt]) Ith(i int) Elt {
	return sl.List[i]
}
