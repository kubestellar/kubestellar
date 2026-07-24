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

package status

import (
	"testing"
)

func TestObjectIsQueried(t *testing.T) {
	tests := []struct {
		name   string
		query  string
		obj    string
		expect bool
	}{
		{
			name:   "exact match",
			query:  "foo",
			obj:    "foo",
			expect: true,
		},
		{
			name:   "infinite loop regression: embedded prefix then standalone",
			query:  "foobar foo",
			obj:    "foo",
			expect: true,
		},
		{
			name:   "embedded only, no standalone",
			query:  "foobar baz",
			obj:    "foo",
			expect: false,
		},
		{
			name:   "embedded only, no standalone anywhere",
			query:  "foobar bazfoo qux",
			obj:    "foo",
			expect: false,
		},
		{
			name:   "standalone at start",
			query:  "foo bar baz",
			obj:    "foo",
			expect: true,
		},
		{
			name:   "standalone at end",
			query:  "bar baz foo",
			obj:    "foo",
			expect: true,
		},
		{
			name:   "not found",
			query:  "bar baz qux",
			obj:    "foo",
			expect: false,
		},
		{
			name:   "empty query",
			query:  "",
			obj:    "foo",
			expect: false,
		},
		{
			name:   "embedded prefix then standalone after many words",
			query:  "xyzfoo abc def foo",
			obj:    "foo",
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := objectIsQueried(&tt.query, tt.obj)
			if result != tt.expect {
				t.Errorf("objectIsQueried(%q, %q) = %v, want %v", tt.query, tt.obj, result, tt.expect)
			}
		})
	}
}
