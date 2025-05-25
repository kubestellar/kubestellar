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

import (
	"testing"
)

func TestLexer(t *testing.T) {
CaseLoop:
	for _, testCase := range []struct {
		source  string
		results []string
		goodEnd func(error) bool
	}{
		{"", nil, badEnd},
		{`'xyz`,
			nil,
			badEnd},
		{`$.xyz`,
			[]string{"xyz"},
			cleanEOF},
		{`$["foo.bar/baz"]`,
			[]string{"foo.bar/baz"},
			cleanEOF},
		{`$["foo.bar/baz"].zork`,
			[]string{"foo.bar/baz", "zork"},
			cleanEOF},
		{`$.zot["foo.bar/baz"]`,
			[]string{"zot", "foo.bar/baz"},
			cleanEOF},
		{`$.`, nil, badEnd},
		{`$[`, nil, badEnd},
		{`$[]`, nil, badEnd},
	} {
		query, err := ParseQuery(testCase.source)
		for idx, good := range testCase.results {
			if idx >= len(query) {
				t.Errorf("For source %q, parse produced only %d results", testCase.source, len(query))
				continue
			}
			if query[idx] != good {
				t.Errorf("For source %q, segment %d is %q but expected token %q", testCase.source, idx, query[idx], good)
				continue CaseLoop
			}
		}
		if !testCase.goodEnd(err) {
			t.Errorf("For source %q, Parse returned wrong err=%#+v", testCase.source, err)
		} else {
			t.Logf("Success for source %q", testCase.source)
		}
	}
}

func cleanEOF(err error) bool {
	return err == nil
}

func badEnd(err error) bool {
	return err != nil
}
