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

// Regression test for the objectIsQueried infinite loop (issue #3848).
// Before the fix, inputs where the first occurrence of obj is an embedded
// prefix of a longer word caused an infinite loop: strings.Index returns a
// relative offset within the sliced substring, but the original code used it
// as an absolute index for isWholeWord, so the search oscillated between two
// wrong positions forever.
package status

import "testing"

func TestObjectIsQueried(t *testing.T) {
	cases := []struct {
		query string
		obj   string
		want  bool
	}{
		// Whole-word occurrences
		{"select foo from bar", "foo", true},
		{"foo", "foo", true},
		// Not a whole word (embedded inside another word)
		{"foobar", "foo", false},
		{"barfoo", "foo", false},
		// Surrounded by punctuation (not alphanumeric) — still a match
		{"(foo)", "foo", true},
		// Key regression case: first occurrence embedded, second standalone.
		// The original code looped forever on this input.
		{"foobar foo", "foo", true},
		// Object not present at all
		{"select bar from baz", "qux", false},
		// Empty query
		{"", "foo", false},
	}
	for _, tc := range cases {
		t.Run(tc.query+"/"+tc.obj, func(t *testing.T) {
			q := tc.query
			got := objectIsQueried(&q, tc.obj)
			if got != tc.want {
				t.Errorf("objectIsQueried(%q, %q) = %v, want %v", tc.query, tc.obj, got, tc.want)
			}
		})
	}
}

func TestIsWholeWord(t *testing.T) {
	cases := []struct {
		s      string
		idx    int
		length int
		want   bool
	}{
		// "foo" alone at position 0
		{"foo", 0, 3, true},
		// "foo" at start followed by space
		{"foo bar", 0, 3, true},
		// "foo" embedded in "foobar"
		{"foobar", 0, 3, false},
		// "foo" at the end preceded by letter
		{"barfoo", 3, 3, false},
		// "foo" in "(foo)"
		{"(foo)", 1, 3, true},
		// "bar" in "bar,baz"
		{"bar,baz", 0, 3, true},
	}
	for _, tc := range cases {
		t.Run(tc.s, func(t *testing.T) {
			s := tc.s
			got := isWholeWord(&s, tc.idx, tc.length)
			if got != tc.want {
				t.Errorf("isWholeWord(%q, %d, %d) = %v, want %v", tc.s, tc.idx, tc.length, got, tc.want)
			}
		})
	}
}
