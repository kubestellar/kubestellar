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

package util

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestParseResourcesString tests the ParseResourcesString function
func TestParseResourcesString(t *testing.T) {
	// Define some test cases with inputs and expected outputs
	testCases := []struct {
		name     string
		input    string
		expected []schema.GroupVersionResource
		err      bool
	}{
		{
			name:  "valid input with api group",
			input: "pods.core/v1, deployments.apps/v1, services.core/v1",
			expected: []schema.GroupVersionResource{
				{
					Group:    "core",
					Version:  "v1",
					Resource: "pods",
				},
				{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
				{
					Group:    "core",
					Version:  "v1",
					Resource: "services",
				},
			},
			err: false,
		},
		{
			name:  "valid input without api group",
			input: "pods/v1, nodes/v1, events/v1",
			expected: []schema.GroupVersionResource{
				{
					Group:    "",
					Version:  "v1",
					Resource: "pods",
				},
				{
					Group:    "",
					Version:  "v1",
					Resource: "nodes",
				},
				{
					Group:    "",
					Version:  "v1",
					Resource: "events",
				},
			},
			err: false,
		},
		{
			name:  "valid input with api group with dot",
			input: "flowschemas.flowcontrol.apiserver.k8s.io/v1beta1, prioritylevelconfigurations.flowcontrol.apiserver.k8s.io/v1beta1",
			expected: []schema.GroupVersionResource{
				{
					Group:    "flowcontrol.apiserver.k8s.io",
					Version:  "v1beta1",
					Resource: "flowschemas",
				},
				{
					Group:    "flowcontrol.apiserver.k8s.io",
					Version:  "v1beta1",
					Resource: "prioritylevelconfigurations",
				},
			},
			err: false,
		},
		{
			name:  "valid input with extra quotes",
			input: "\"flowschemas.flowcontrol.apiserver.k8s.io/v1beta1\"",
			expected: []schema.GroupVersionResource{
				{
					Group:    "flowcontrol.apiserver.k8s.io",
					Version:  "v1beta1",
					Resource: "flowschemas",
				},
			},
			err: false,
		},
		{
			name:     "invalid resource format",
			input:    "pods.core/v1, deployments.apps/v1, services",
			expected: nil,
			err:      true,
		},
		{
			name:     "invalid group/version format",
			input:    "pods.core/v1, deployments.apps/v1, services.core",
			expected: nil,
			err:      true,
		},
		{
			name:     "invalid resource/version format",
			input:    "pods.core/v1, deployments.apps/v1, services/",
			expected: nil,
			err:      true,
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
			err:      false,
		},
	}

	// Iterate over the test cases, run the function with the input and compare expected vs. returned
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := ParseResourcesString(tc.input)

			// Check if the error matches the expected error
			if (err != nil) != tc.err {
				t.Fatalf("returned error does not match: %v", err)
			}

			// Check if the output matches the expected output
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("expected: %v, got: %v", tc.expected, actual)
			}
		})
	}
}

// TestIsResourceAllowed tests the IsResourceAllowed function
func TestIsResourceAllowed(t *testing.T) {
	// Define some test cases with inputs and expected outputs
	testCases := []struct {
		name             string
		resource         schema.GroupVersionResource
		allowedResources []schema.GroupVersionResource
		expected         bool
	}{
		{
			name: "resource is allowed by non-empty slice",
			resource: schema.GroupVersionResource{
				Group:    "core",
				Version:  "v1",
				Resource: "pods",
			},
			allowedResources: []schema.GroupVersionResource{
				{
					Group:    "core",
					Version:  "v1",
					Resource: "pods",
				},
				{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
			},
			expected: true,
		},
		{
			name: "Any version is allowed",
			resource: schema.GroupVersionResource{
				Group:    "core",
				Version:  "v1beta1",
				Resource: "pods",
			},
			allowedResources: []schema.GroupVersionResource{
				{
					Group:    "core",
					Version:  "*",
					Resource: "pods",
				},
			},
			expected: true,
		},
		{
			name: "placements are always allowed for non empty list",
			resource: schema.GroupVersionResource{
				Group:    "control.kubestellar.io",
				Version:  "v1alpha1",
				Resource: "placements",
			},
			allowedResources: []schema.GroupVersionResource{
				{
					Group:    "core",
					Version:  "*",
					Resource: "pods",
				},
			},
			expected: true,
		},
		// disabled until https://github.com/kubestellar/kubestellar/issues/1705 is resolved
		// to avoid client-side throttling
		// {
		// 	name: "customresourcedefinitions are always allowed for non empty list",
		// 	resource: schema.GroupVersionResource{
		// 		Group:    "apiextensions.k8s.io",
		// 		Version:  "v1",
		// 		Resource: "customresourcedefinitions",
		// 	},
		// 	allowedResources: []schema.GroupVersionResource{
		// 		{
		// 			Group:    "core",
		// 			Version:  "*",
		// 			Resource: "pods",
		// 		},
		// 	},
		// 	expected: true,
		// },
		{
			name: "resource is not allowed by non-empty slice",
			resource: schema.GroupVersionResource{
				Group:    "core",
				Version:  "v1",
				Resource: "nodes",
			},
			allowedResources: []schema.GroupVersionResource{
				{
					Group:    "core",
					Version:  "v1",
					Resource: "pods",
				},
				{
					Group:    "apps",
					Version:  "v1",
					Resource: "deployments",
				},
			},
			expected: false,
		},
		{
			name: "resource is allowed by empty slice",
			resource: schema.GroupVersionResource{
				Group:    "core",
				Version:  "v1",
				Resource: "nodes",
			},
			allowedResources: []schema.GroupVersionResource{},
			expected:         true,
		},
		{
			name: "resource is allowed by nil slice",
			resource: schema.GroupVersionResource{
				Group:    "core",
				Version:  "v1",
				Resource: "nodes",
			},
			allowedResources: nil,
			expected:         true,
		},
	}

	// Iterate over the test cases and run the function with the input
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsResourceAllowed(tc.resource, tc.allowedResources)

			// Check if the output matches the expected output
			if actual != tc.expected {
				t.Errorf("expected: %v, got: %v", tc.expected, actual)
			}
		})
	}
}
