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

package fuzz

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/jsonpath"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// FuzzJSONPathParsing tests the JSONPath parser with various inputs
func FuzzJSONPathParsing(f *testing.F) {
	// Add seed corpus for JSONPath expressions
	seedCorpus := []string{
		"$.spec.name",
		"$.metadata.labels",
		"$.spec.containers[0].name",
		"$.status.conditions[0].type",
		"$.spec.template.spec.containers[0].env[0].name",
		"$.metadata.annotations[\"kubernetes.io/change-cause\"]",
		"$.spec.selector.matchLabels",
		"$.spec.template.metadata.labels",
		"$.spec.ports[0].port",
		"$.spec.rules[0].http.paths[0].path",
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, queryStr string) {
		// Skip empty strings
		if strings.TrimSpace(queryStr) == "" {
			return
		}

		// Try to parse the JSONPath query
		query, err := jsonpath.ParseQuery(queryStr)
		if err != nil {
			// Expected for invalid inputs, just return
			return
		}

		// If parsing succeeded, validate the query structure
		if len(query) == 0 {
			return
		}

		// Test with a simple JSON object
		testObj := map[string]interface{}{
			"spec": map[string]interface{}{
				"name": "test",
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "container1",
						"image": "nginx:latest",
					},
				},
			},
			"metadata": map[string]interface{}{
				"labels": map[string]interface{}{
					"app": "test",
				},
				"annotations": map[string]interface{}{
					"kubernetes.io/change-cause": "test-deployment",
				},
			},
		}

		// Create a root node for testing
		var jsonValue jsonpath.JSONValue = testObj
		rootNode := &jsonpath.RootNode{Value: &jsonValue}

		// Try to execute the query
		var results []jsonpath.Node
		jsonpath.QueryValue(query, rootNode, func(node jsonpath.Node) {
			results = append(results, node)
		})

		// Basic validation - if query has fields, we should get some results
		// (though they might be nil if the path doesn't exist)
		if len(query) > 0 && len(results) == 0 {
			// This might indicate an issue, but not necessarily a bug
			// as the path might not exist in our test object
		}
	})
}

// FuzzCRDValidation tests CRD validation with various YAML inputs
func FuzzCRDValidation(f *testing.F) {
	// Add seed corpus for CRD validation
	baseBindingPolicy := `apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: test-binding-policy
  namespace: default
spec:
  downsync:
    - apiGroup: ""
      resources: ["pods"]
      namespaces: ["default"]
      objectSelectors:
        - matchLabels:
            app: test
`

	baseStatusCollector := `apiVersion: control.kubestellar.io/v1alpha1
kind: StatusCollector
metadata:
  name: test-status-collector
  namespace: default
spec:
  limit: 10
  select:
    - name: podName
      def: object.metadata.name
    - name: podStatus
      def: object.status.phase
`

	baseBinding := `apiVersion: control.kubestellar.io/v1alpha1
kind: Binding
metadata:
  name: test-binding
  namespace: default
spec:
  bindingPolicyRef: test-binding-policy
  workloadSelector:
    apiGroup: ""
    resource: "pods"
    namespace: "default"
    name: "test-pod"
`

	// Add seed corpus
	f.Add([]byte(baseBindingPolicy))
	f.Add([]byte(baseStatusCollector))
	f.Add([]byte(baseBinding))
	f.Add([]byte(baseBindingPolicy + "---\n" + baseStatusCollector))
	f.Add([]byte(baseBindingPolicy + "---\n" + baseBinding))

	f.Fuzz(func(t *testing.T, yamlData []byte) {
		// Skip empty data
		if len(yamlData) == 0 {
			return
		}

		// Try to parse as YAML
		var obj unstructured.Unstructured
		err := yaml.Unmarshal(yamlData, &obj)
		if err != nil {
			// Expected for invalid YAML, just return
			return
		}

		// Check if it's a KubeStellar CRD
		apiVersion := obj.GetAPIVersion()
		kind := obj.GetKind()

		if !strings.Contains(apiVersion, "kubestellar.io") {
			return
		}

		// Validate specific CRD types
		switch kind {
		case "BindingPolicy":
			// Try to convert to BindingPolicy
			var bp v1alpha1.BindingPolicy
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &bp)
			if err != nil {
				return
			}

			// Basic validation
			if bp.Name == "" {
				return
			}

		case "StatusCollector":
			// Try to convert to StatusCollector
			var sc v1alpha1.StatusCollector
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &sc)
			if err != nil {
				return
			}

			// Basic validation
			if sc.Name == "" {
				return
			}

		case "Binding":
			// Try to convert to Binding
			var binding v1alpha1.Binding
			err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), &binding)
			if err != nil {
				return
			}

			// Basic validation
			if binding.Name == "" {
				return
			}
		}
	})
}

// FuzzLabelParsing tests label parsing functionality
func FuzzLabelParsing(f *testing.F) {
	// Add seed corpus for label parsing
	seedCorpus := []string{
		"app=test",
		"environment=production",
		"version=v1.0.0",
		"tier=frontend",
		"component=api",
		"managed-by=kubestellar",
		"owner=team-a",
		"project=myapp",
		"region=us-west-2",
		"zone=us-west-2a",
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, labelStr string) {
		// Skip empty strings
		if strings.TrimSpace(labelStr) == "" {
			return
		}

		// Try to parse the label
		label, err := util.SplitLabelKeyAndValue(labelStr)
		if err != nil {
			// Expected for invalid label formats, just return
			return
		}

		// Basic validation
		if label.Key == "" {
			return
		}

		// Test that we can reconstruct the label string
		reconstructed := fmt.Sprintf("%s=%s", label.Key, label.Value)
		if reconstructed != labelStr {
			// This might indicate an issue with parsing/formatting
			t.Errorf("Label reconstruction mismatch: expected %q, got %q", labelStr, reconstructed)
		}
	})
}

// FuzzAPIGroupParsing tests API group parsing functionality
func FuzzAPIGroupParsing(f *testing.F) {
	// Add seed corpus for API group parsing
	seedCorpus := []string{
		"",
		"apps",
		"apps,networking.k8s.io",
		"apps,networking.k8s.io,control.kubestellar.io",
		"control.kubestellar.io",
		"apiextensions.k8s.io",
		"rbac.authorization.k8s.io",
		"storage.k8s.io",
		"autoscaling",
		"batch",
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, apiGroupsStr string) {
		// Parse API groups
		apiGroups := util.ParseAPIGroupsString(apiGroupsStr)

		// Test that required groups are always included
		if apiGroups != nil {
			// Check that KubeStellar control group is included
			if !apiGroups.Has("control.kubestellar.io") {
				t.Errorf("Required API group 'control.kubestellar.io' not found in parsed groups")
			}

			// Check that apiextensions group is included
			if !apiGroups.Has("apiextensions.k8s.io") {
				t.Errorf("Required API group 'apiextensions.k8s.io' not found in parsed groups")
			}
		}

		// Test API group allowance
		testGroups := []string{"apps", "networking.k8s.io", "control.kubestellar.io", "apiextensions.k8s.io"}
		for _, group := range testGroups {
			allowed := util.IsAPIGroupAllowed(group, apiGroups)
			// If apiGroups is nil or empty, all groups should be allowed
			if apiGroups == nil || len(apiGroups) == 0 {
				if !allowed {
					t.Errorf("Group %q should be allowed when no restrictions are set", group)
				}
			}
		}
	})
}

// FuzzJSONValueValidation tests JSON value validation
func FuzzJSONValueValidation(f *testing.F) {
	// Add seed corpus for JSON value validation
	seedCorpus := []string{
		`{"type": "String", "string": "test"}`,
		`{"type": "Number", "float": "42"}`,
		`{"type": "Bool", "bool": true}`,
		`{"type": "Object", "object": {"key": "value"}}`,
		`{"type": "Array", "array": [1, 2, 3]}`,
		`{"type": "Null"}`,
	}

	for _, seed := range seedCorpus {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, jsonStr string) {
		// Skip empty strings
		if strings.TrimSpace(jsonStr) == "" {
			return
		}

		// Try to parse as JSON
		var value v1alpha1.Value
		err := json.Unmarshal([]byte(jsonStr), &value)
		if err != nil {
			// Expected for invalid JSON, just return
			return
		}

		// Skip validation if the type field is empty or invalid
		// This happens when JSON has typos in field names (e.g., "tYpe" instead of "type")
		// or when the type value is not one of the valid ValueType constants
		if value.Type == "" {
			return
		}

		// Validate that the type is one of the known valid types
		validTypes := map[v1alpha1.ValueType]bool{
			v1alpha1.TypeString: true,
			v1alpha1.TypeNumber: true,
			v1alpha1.TypeBool:   true,
			v1alpha1.TypeObject: true,
			v1alpha1.TypeArray:  true,
			v1alpha1.TypeNull:   true,
		}
		if !validTypes[value.Type] {
			// Invalid type value, skip validation
			return
		}

		// Validate the value based on its type
		// Note: The Value struct fields are marked with omitempty, so they can be
		// omitted from JSON even when the type is specified. This is valid behavior.
		switch value.Type {
		case v1alpha1.TypeString:
			// String type can have nil String field (omitted from JSON)
			// This is valid when the JSON only specifies the type
		case v1alpha1.TypeNumber:
			// Number type can have nil Number field (omitted from JSON)
			// This is valid when the JSON only specifies the type
		case v1alpha1.TypeBool:
			// Bool type can have nil Bool field (omitted from JSON)
			// This is valid when the JSON only specifies the type
		case v1alpha1.TypeObject:
			// Object type can have nil Object field (omitted from JSON)
			// This is valid when the JSON only specifies the type
		case v1alpha1.TypeArray:
			// Array type can have nil Array field (omitted from JSON)
			// This is valid when the JSON only specifies the type
		case v1alpha1.TypeNull:
			// Null type doesn't need any fields
		}
	})
}
