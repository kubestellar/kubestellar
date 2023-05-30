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

package jsonpath

// JSONValue is something that can be produced by encoding/json.Unmarshal(bytes, map[string]any{}).
// That is: `bool`, `float64`, `string`, `nil`, `[]any`, or `map[string]any`.
type JSONValue = any

// Apply returns the result of applying the given function to the
// places in the given data selected by the given path.  Selection by
// a path prefix will introduce selected map members that do not
// already exist IFF `definite` and every selector in the prefix is
// "definite".  The definite selectors are SelectorName, SelectorList,
// and SelectorRange where afterEnd is non-nil and equal to start+1
// and stride is 1.
// For SelectorRecurse, a parent is visited before its children.
func Apply(data JSONValue, path []Selector, definite bool, fn func(JSONValue) JSONValue) JSONValue {
	if len(path) == 0 {
		return fn(data)
	}
	sel := path[0]
	switch sel.Type {
	case SelectorName:
		if typed, ok := data.(map[string]any); ok {
			if definite {
				typed[sel.Name] = Apply(typed[sel.Name], path[1:], definite, fn)
			} else if elt, ok := typed[sel.Name]; ok {
				typed[sel.Name] = Apply(elt, path[1:], definite, fn)
			}
		}
	case SelectorRange:
		definite = definite && sel.Range.afterEnd != nil && *sel.Range.afterEnd == sel.Range.start+1 && sel.Range.stride == 1
		if typed, ok := data.([]any); ok {
			limit := len(typed)
			if sel.Range.afterEnd != nil && *sel.Range.afterEnd < limit {
				limit = *sel.Range.afterEnd
			}
			for index := sel.Range.start; index < limit; index += sel.Range.stride {
				typed[index] = Apply(typed[index], path[1:], definite, fn)
			}
		}
	case SelectorList:
		for _, sub := range sel.List {
			data = Apply(data, append([]Selector{sub}, path[1:]...), definite, fn)
		}
	case SelectorEveryChild:
		switch typed := data.(type) {
		case []any:
			for index := range typed {
				typed[index] = Apply(typed[index], path[1:], false, fn)
			}
		case map[string]any:
			for key := range typed {
				typed[key] = Apply(typed[key], path[1:], false, fn)
			}
		}
	case SelectorRecurse:
		data = Apply(data, path[1:], false, fn)
		switch typed := data.(type) {
		case []any:
			for index := range typed {
				typed[index] = Apply(typed[index], path, false, fn)
			}
		case map[string]any:
			for key := range typed {
				typed[key] = Apply(typed[key], path, false, fn)
			}
		}
	}
	return data
}
