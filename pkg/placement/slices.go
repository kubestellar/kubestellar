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

func SliceContains[Elt comparable](slice []Elt, seek Elt) bool {
	for _, elt := range slice {
		if elt == seek {
			return true
		}
	}
	return false
}

func SliceRemoveFunctional[Elt comparable](slice []Elt, seek Elt) []Elt {
	ans := []Elt{}
	for _, elt := range slice {
		if elt != seek {
			ans = append(ans, elt)
		}
	}
	return ans
}

func SliceApply[Elt any](slice []Elt, fn func(Elt)) {
	for _, elt := range slice {
		fn(elt)
	}
}

func SetToSlice[Elt comparable](set Set[Elt]) []Elt {
	ans := make([]Elt, 0, set.Len())
	set.Visit(func(elt Elt) error {
		ans = append(ans, elt)
		return nil
	})
	return ans
}

func SetMapToSlice[Original comparable, Mapped any](set Set[Original], mapfn func(Original) Mapped) []Mapped {
	ans := make([]Mapped, 0, set.Len())
	set.Visit(func(elt Original) error {
		mapped := mapfn(elt)
		ans = append(ans, mapped)
		return nil
	})
	return ans
}

func MapMapToSlice[Key comparable, Val, Mapped any](theMap Map[Key, Val], mapfn func(Key, Val) Mapped) []Mapped {
	ans := make([]Mapped, 0, theMap.Len())
	theMap.Visit(func(tup Pair[Key, Val]) error {
		mapped := mapfn(tup.First, tup.Second)
		ans = append(ans, mapped)
		return nil
	})
	return ans
}
