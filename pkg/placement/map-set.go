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

type MapSet[Elt comparable] map[Elt]Empty

var _ MutableSet[string] = MapSet[string]{}

func NewMapSet[Elt comparable](elts ...Elt) MapSet[Elt] {
	ans := NewEmptyMapSet[Elt]()
	for _, elt := range elts {
		ans.Add(elt)
	}
	return ans
}

func NewEmptyMapSet[Elt comparable]() MapSet[Elt] {
	return MapSet[Elt]{}
}

func MapSetCopy[Elt comparable](source Visitable[Elt]) MapSet[Elt] {
	ans := NewEmptyMapSet[Elt]()
	SetAddAll[Elt](ans, source)
	return ans
}

func MapSetCopier[Elt comparable]() Reducer[Elt, MapSet[Elt]] {
	return StatefulReducer(NewEmptyMapSet[Elt], MapSetAddNoResult[Elt], Identity1[MapSet[Elt]])
}

func MapSetAsVisitable[Elt comparable](ms MapSet[Elt]) Visitable[Elt] { return ms }

func (ms MapSet[Elt]) IsEmpty() bool    { return len(ms) == 0 }
func (ms MapSet[Elt]) Len() int         { return len(ms) }
func (ms MapSet[Elt]) LenIsCheap() bool { return true }

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

func MapSetAddNoResult[Elt comparable](set MapSet[Elt], elt Elt) {
	set.Add(elt)
}

func MapSetSymmetricDifference[Elt comparable](wantLeftMinusRight, wantIntersection, wantRightMinusLeft bool, left, right Set[Elt]) (MapSet[Elt], MapSet[Elt], MapSet[Elt]) {
	var leftMinusRight, intersection, rightMinusLeft MapSet[Elt]
	if wantLeftMinusRight {
		leftMinusRight = NewEmptyMapSet[Elt]()
	}
	if wantIntersection {
		intersection = NewEmptyMapSet[Elt]()
	}
	if wantRightMinusLeft {
		rightMinusLeft = NewEmptyMapSet[Elt]()
	}
	if wantLeftMinusRight {
		left.Visit(func(leftElt Elt) error {
			if right.Has(leftElt) {
				if wantIntersection {
					intersection.Add(leftElt)
				}
			} else {
				leftMinusRight.Add(leftElt)
			}
			return nil
		})
	}
	if wantRightMinusLeft || wantIntersection && !wantLeftMinusRight {
		right.Visit(func(rightElt Elt) error {
			if left.Has(rightElt) {
				if wantIntersection && !wantLeftMinusRight {
					intersection.Add(rightElt)
				}
			} else if wantRightMinusLeft {
				rightMinusLeft.Add(rightElt)
			}
			return nil
		})
	}
	return leftMinusRight, intersection, rightMinusLeft
}
