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
		results []Segment
		goodEnd func(error) bool
	}{
		{"", nil, badEnd},
		{`'xyz`,
			nil,
			badEnd},
		{`$.xyz`,
			[]Segment{{Kind: SegmentKindField, Name: "xyz"}},
			cleanEOF},
		{`$["foo.bar/baz"]`,
			[]Segment{{Kind: SegmentKindField, Name: "foo.bar/baz"}},
			cleanEOF},
		{`$["foo.bar/baz"].zork`,
			[]Segment{{Kind: SegmentKindField, Name: "foo.bar/baz"}, {Kind: SegmentKindField, Name: "zork"}},
			cleanEOF},
		{`$.zot["foo.bar/baz"]`,
			[]Segment{{Kind: SegmentKindField, Name: "zot"}, {Kind: SegmentKindField, Name: "foo.bar/baz"}},
			cleanEOF},
		{`$.`, nil, badEnd},
		{`$[`, nil, badEnd},
		{`$[]`, nil, badEnd},
		{`$.spec.resources.GenericItems[*].generictemplate.metadata.resourceVersion`,
			[]Segment{
				{Kind: SegmentKindField, Name: "spec"},
				{Kind: SegmentKindField, Name: "resources"},
				{Kind: SegmentKindField, Name: "GenericItems"},
				{Kind: SegmentKindWildcard},
				{Kind: SegmentKindField, Name: "generictemplate"},
				{Kind: SegmentKindField, Name: "metadata"},
				{Kind: SegmentKindField, Name: "resourceVersion"},
			},
			cleanEOF},
		{`$.spec.resources.GenericItems.*.metadata.resourceVersion`,
			[]Segment{
				{Kind: SegmentKindField, Name: "spec"},
				{Kind: SegmentKindField, Name: "resources"},
				{Kind: SegmentKindField, Name: "GenericItems"},
				{Kind: SegmentKindWildcard},
				{Kind: SegmentKindField, Name: "metadata"},
				{Kind: SegmentKindField, Name: "resourceVersion"},
			},
			cleanEOF},
	} {
		query, err := ParseQuery(testCase.source)
		for idx, good := range testCase.results {
			if idx >= len(query) {
				t.Errorf("For source %q, parse produced only %d results", testCase.source, len(query))
				continue
			}
			if query[idx] != good {
				t.Errorf("For source %q, segment %d is %#v but expected %#v", testCase.source, idx, query[idx], good)
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
