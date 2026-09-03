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

	// Test with arrays and wildcard
	var root2 RootNode
	err = json.Unmarshal([]byte(`{
		"spec": {
			"resources": {
				"GenericItems": [
					{"generictemplate": {"metadata": {"resourceVersion": "123", "name": "item1"}}},
					{"generictemplate": {"metadata": {"resourceVersion": "456", "name": "item2"}}}
				]
			}
		}
	}`), &root2.Value)
	if err != nil {
		t.Fatalf("Failed to parse doc2, err=%s", err.Error())
	}

	expected = []JSONValue{"123", "456"}
	actual = GetQuery(&root2, "$.spec.resources.GenericItems[*].generictemplate.metadata.resourceVersion")
	if !jsonEqualities.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, got %#v", expected, actual)
	}

	// Test with dot-wildcard as well
	actual = GetQuery(&root2, "$.spec.resources.GenericItems.*.generictemplate.metadata.resourceVersion")
	if !jsonEqualities.DeepEqual(expected, actual) {
		t.Errorf("Expected %#v, got %#v", expected, actual)
	}

	// Test Remove using wildcard
	query, err := ParseQuery("$.spec.resources.GenericItems[*].generictemplate.metadata.resourceVersion")
	if err != nil {
		t.Fatalf("Failed to parse remove query, err=%s", err.Error())
	}
	QueryValue(query, &root2, func(node Node) {
		node.Remove()
	})

	// Verify that the fields were removed
	var expectedDoc any
	err = json.Unmarshal([]byte(`{
		"spec": {
			"resources": {
				"GenericItems": [
					{"generictemplate": {"metadata": {"name": "item1"}}},
					{"generictemplate": {"metadata": {"name": "item2"}}}
				]
			}
		}
	}`), &expectedDoc)
	if err != nil {
		t.Fatalf("Failed to parse expectedDoc, err=%s", err.Error())
	}

	if !jsonEqualities.DeepEqual(expectedDoc, *root2.Value) {
		t.Errorf("After Remove, expected %#v, got %#v", expectedDoc, *root2.Value)
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
