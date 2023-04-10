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

func Relation2Equijoin12with13[First, Second, Third comparable](left Relation2[First, Second], right Relation2[First, Third]) Relation2[Second, Third] {
	ans := NewMapRelation2[Second, Third]()
	left.Visit(func(xy Pair[First, Second]) error {
		right.GetIndex1to2().Visit1to2(xy.First, func(z Third) error {
			ans.Add(Pair[Second, Third]{xy.Second, z})
			return nil
		})
		return nil
	})
	return ans
}

func MapEquijoin12With13[Key comparable, ValLeft, ValRight any](left Map[Key, ValLeft], right Map[Key, ValRight]) Map[Key, Pair[ValLeft, ValRight]] {
	ans := NewMapMap[Key, Pair[ValLeft, ValRight]](nil)
	left.Visit(func(kl Pair[Key, ValLeft]) error {
		rightVal, has := right.Get(kl.First)
		if has {
			ans.Put(kl.First, Pair[ValLeft, ValRight]{kl.Second, rightVal})
		}
		return nil
	})
	return ans
}

func JoinByVisitSquared[Left, Right, Joint any](
	left Visitable[Left],
	right Visitable[Right],
	match func(Left, Right) (Joint, bool),
) Visitable[Joint] {
	return joinByVisitSquared[Left, Right, Joint]{left, right, match}
}

type joinByVisitSquared[Left, Right, Joint any] struct {
	left  Visitable[Left]
	right Visitable[Right]
	match func(Left, Right) (Joint, bool)
}

func (jbv joinByVisitSquared[Left, Right, Joint]) Visit(visitor func(Joint) error) error {
	return jbv.left.Visit(func(leftElt Left) error {
		return jbv.right.Visit(func(rightElt Right) error {
			joint, isMatch := jbv.match(leftElt, rightElt)
			if !isMatch {
				return nil
			}
			return visitor(joint)
		})
	})
}
