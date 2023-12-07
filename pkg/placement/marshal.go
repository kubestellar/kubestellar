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

import (
	"encoding/json"
	"strings"
)

func MarshalMap[Key comparable, Val any](it map[Key]Val) ([]byte, error) {
	if it == nil {
		return []byte("null"), nil
	}
	var builder strings.Builder
	enc := json.NewEncoder(&builder)
	builder.WriteRune('[')
	first := true
	for key, val := range it {
		item := struct{ Key, Val any }{key, val}
		if first {
			first = false
		} else {
			builder.WriteString(", ")
		}
		err := enc.Encode(item)
		if err != nil {
			errS := err.Error()
			enc.Encode(errS)
		}
	}
	builder.WriteRune(']')
	return []byte(builder.String()), nil
}

func MarshalSet[Key comparable](it map[Key]Empty) ([]byte, error) {
	if it == nil {
		return []byte("null"), nil
	}
	var builder strings.Builder
	enc := json.NewEncoder(&builder)
	builder.WriteRune('[')
	first := true
	for key := range it {
		if first {
			first = false
		} else {
			builder.WriteString(", ")
		}
		err := enc.Encode(key)
		if err != nil {
			errS := err.Error()
			enc.Encode(errS)
		}
	}
	builder.WriteRune(']')
	return []byte(builder.String()), nil
}
