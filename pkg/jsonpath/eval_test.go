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
	"encoding/json"
	"testing"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
)

func TestEval(t *testing.T) {
	for _, testCase := range []struct {
		inputStr       string
		path           []Selector
		definite       bool
		replacementStr string
		expectStr      string
	}{
		{`{"abc": {"abc": {"def": 3, "ghi":4}}, "def": ["a", "b", "c", "d", "e"]}`,
			[]Selector{{Type: SelectorRecurse}, {Type: SelectorName, Name: "def"}},
			true, "47",
			`{"abc": {"abc": {"def": 47, "ghi":4}}, "def": 47}`},
		{`{"abc": {"abc": {"def": 3, "ghi":4}}, "def": ["a", "b", "c", "d", "e"]}`,
			[]Selector{{Type: SelectorRecurse}, {Type: SelectorName, Name: "abc"}},
			true, "47",
			`{"abc": 47, "def": ["a", "b", "c", "d", "e"]}`},
		{`{"abc": {"abc": {"def": 3, "ghi":4}}, "def": ["a", "b", "c", "d", "e"]}`,
			[]Selector{{Type: SelectorName, Name: "def"}, {Type: SelectorRange, Range: &Range{1, ptr(2), 1}}},
			true, "47",
			`{"abc": {"abc": {"def": 3, "ghi":4}}, "def": ["a", 47, "c", "d", "e"]}`},
		{`{"abc": {"abc": {"def": 3, "ghi":4}}, "def": ["a", "b", "c", "d", "e"]}`,
			[]Selector{{Type: SelectorName, Name: "def"}, {Type: SelectorRange, Range: &Range{2, nil, 2}}},
			true, "47",
			`{"abc": {"abc": {"def": 3, "ghi":4}}, "def": ["a", "b", 47, "d", 47]}`},
		{`[{"ab":1}, {"ab":2}, {"ab":3}, {"ab":4}, {"ab":5}]`,
			[]Selector{{Type: SelectorRange, Range: &Range{2, nil, 2}}, {Type: SelectorName, Name: "ab"}},
			true, "47",
			`[{"ab":1}, {"ab":2}, {"ab":47}, {"ab":4}, {"ab":47}]`},
		{`[{"ab":1}, {"ab":2}, {"xy":3}, {"ab":4}, {"ab":5}]`,
			[]Selector{{Type: SelectorRange, Range: &Range{2, nil, 2}}, {Type: SelectorName, Name: "xy"}},
			true, "47",
			`[{"ab":1}, {"ab":2}, {"xy":47}, {"ab":4}, {"ab":5}]`},
		{`[{"ab":1}, {"ab":2}, {"ab":3}, {"ab":4}, {"ab":5}]`,
			[]Selector{{Type: SelectorRange, Range: &Range{2, ptr(3), 1}}, {Type: SelectorName, Name: "xy"}},
			true, "47",
			`[{"ab":1}, {"ab":2}, {"ab":3, "xy":47}, {"ab":4}, {"ab":5}]`},
		{`[{"ab":1}, {"ab":2}, {"ab":3}, {"ab":4}, {"ab":5}]`,
			[]Selector{{Type: SelectorRange, Range: &Range{2, ptr(3), 1}}, {Type: SelectorName, Name: "xy"}},
			false, "47",
			`[{"ab":1}, {"ab":2}, {"ab":3}, {"ab":4}, {"ab":5}]`},
	} {
		var inputVal JSONValue
		err := json.Unmarshal([]byte(testCase.inputStr), &inputVal)
		if err != nil {
			panic(err)
		}
		var replacementVal JSONValue
		err = json.Unmarshal([]byte(testCase.replacementStr), &replacementVal)
		if err != nil {
			panic(err)
		}
		var expectVal JSONValue
		err = json.Unmarshal([]byte(testCase.expectStr), &expectVal)
		if err != nil {
			panic(err)
		}
		outputVal := Apply(inputVal, testCase.path, testCase.definite, func(JSONValue) JSONValue { return replacementVal })
		if !apiequality.Semantic.DeepEqual(outputVal, expectVal) {
			t.Errorf("Failed case path=%+v definite=%v replacement=%+v expect=%+v: got %+v", testCase.path, testCase.definite, replacementVal, expectVal, outputVal)
		} else {
			t.Logf("Success for case path=%+v definite=%v replacement=%+v expect=%+v: got %+v", testCase.path, testCase.definite, replacementVal, expectVal, outputVal)
		}
	}
}
