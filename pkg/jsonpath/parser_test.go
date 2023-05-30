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

func TestParser(t *testing.T) {
	for _, testCase := range []struct {
		source  string
		expect  []Selector
		goodErr func(error) bool
	}{
		{`$.a[1]`,
			[]Selector{{Type: SelectorName, Name: "a"}, {Type: SelectorRange, Range: rangeOf1(1)}},
			noErr},
		{`$.a[1:4]`,
			[]Selector{{Type: SelectorName, Name: "a"}, {Type: SelectorRange, Range: &Range{1, ptr(int(4)), 1}}},
			noErr},
		{`$.a[:4]`,
			[]Selector{{Type: SelectorName, Name: "a"}, {Type: SelectorRange, Range: &Range{0, ptr(int(4)), 1}}},
			noErr},
		{`$[1:4:9]`,
			[]Selector{{Type: SelectorRange, Range: &Range{1, ptr(int(4)), 9}}},
			noErr},
		{`$[1::9]`,
			[]Selector{{Type: SelectorRange, Range: &Range{1, nil, 9}}},
			noErr},
		{`$[:4:9]`,
			[]Selector{{Type: SelectorRange, Range: &Range{0, ptr(int(4)), 9}}},
			noErr},
		{`$[::9]`,
			[]Selector{{Type: SelectorRange, Range: &Range{0, nil, 9}}},
			noErr},
		{`$[::]`,
			[]Selector{{Type: SelectorRange, Range: &Range{0, nil, 1}}},
			noErr},
		{`$.1`,
			[]Selector{{Type: SelectorRange, Range: rangeOf1(1)}},
			noErr},
		{`$.*`,
			[]Selector{{Type: SelectorEveryChild}},
			noErr},
		{`$[*]`,
			[]Selector{{Type: SelectorEveryChild}},
			noErr},
		{`$..*`,
			[]Selector{{Type: SelectorRecurse}, {Type: SelectorEveryChild}},
			noErr},
	} {
		parsed, err := ParseString(testCase.source)
		if !testCase.goodErr(err) {
			t.Errorf("Failed case %q: got parsed=%v, wrong error %#+v", testCase.source, parsed, err)
			continue
		}
		if !parsed.Equals(testCase.expect) {
			t.Errorf("Failed case %q: expected %v, got %v", testCase.source, testCase.expect, parsed)
			continue
		}
		t.Logf("Passed case %q", testCase.source)
	}
}

func noErr(err error) bool {
	return err == nil
}
