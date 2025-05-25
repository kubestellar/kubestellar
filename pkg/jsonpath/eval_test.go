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

package jsonpath

import (
	"encoding/json"
	"testing"

	k8sreflect "k8s.io/apimachinery/third_party/forked/golang/reflect"
)

func TestEval(t *testing.T) {
	var root1 RootNode
	err := json.Unmarshal([]byte(`{"abc": 1, "def":{"x":"yz", "abc":true}, "ghi":null}`), &root1.Value)
	if err != nil {
		t.Fatalf("Failed to parse doc1, err=%s", err.Error())
	}
	expected := []JSONValue{float64(1)}
	actual := GetQuery(&root1, "$.abc")
	if !jsonEqualities.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, got %#v", expected, actual)
	}
	expected = []JSONValue{true}
	actual = GetQuery(&root1, `$["def"].abc`)
	if !jsonEqualities.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, got %#v", expected, actual)
	}
	expected = []JSONValue{map[string]any{"x": "yz", "abc": true}}
	actual = GetQuery(&root1, `$.def`)
	if !jsonEqualities.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, got %#v", expected, actual)
	}
}

var jsonEqualities = k8sreflect.Equalities{}

func GetQuery(root Node, pathS string) []JSONValue {
	query, err := ParseQuery(pathS)
	if err != nil {
		panic(err)
	}
	ans := []JSONValue{}
	QueryValue(query, root, func(node Node) {
		val, ok := node.Get()
		if ok {
			ans = append(ans, val)
		}
	})
	return ans
}
