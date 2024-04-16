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
	actual = GetQuery(&root1, `$ ['def'].abc`)
	if !jsonEqualities.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, got %#v", expected, actual)
	}
	expected = []JSONValue{true}
	actual = GetQuery(&root1, `$ [?@.x=="yz"].abc`)
	if !jsonEqualities.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, got %#v", expected, actual)
	}
}

func GetQuery(root Node, pathS string) []JSONValue {
	lexer := NewLexer(pathS)
	parser, err := NewParser(lexer)
	if err != nil {
		panic(err)
	}
	query, err := parser.ParseQuery()
	if err != nil {
		panic(err)
	}
	ans := []JSONValue{}
	QueryValue(query, root, root, func(node Node) { ans = append(ans, node.Get()) })
	return ans
}
