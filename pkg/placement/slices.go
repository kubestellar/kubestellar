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
