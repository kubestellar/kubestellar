/*
Copyright 2026 The KubeStellar Authors.

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

// MapKeysToAny returns a `[]any“ containing the keys of the given map
func MapKeysToAny[Domain, Range any](input MapToLocked[Domain, Range]) []any {
	ans := make([]any, 0, input.Length())
	_ = input.Iterate2(func(d Domain, r Range) error { ans = append(ans, d); return nil })
	return ans
}
