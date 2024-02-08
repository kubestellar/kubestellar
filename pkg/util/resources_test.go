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
)

// TestParseResourceGroupsString tests the TestParseResourceGroupsString function
func TestParseResourceGroupsString(t *testing.T) {
	// Define some test cases with inputs and expected outputs
	testCases := []struct {
		name     string
		input    string
		expected map[string]bool
	}{
		{
			name:  "valid input with api group",
			input: "apps, networking.k8s.io, policy",
			expected: map[string]bool{
				"apps":              true,
				"networking.k8s.io": true,
				"policy":            true,
			},
		},
		{
			name:  "valid input with empty api group",
			input: "apps, ,policy",
			expected: map[string]bool{
				"apps":   true,
				"":       true,
				"policy": true,
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
	}

	// Iterate over the test cases, run the function with the input and compare expected vs. returned
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ParseResourceGroupsString(tc.input)

			// Check if the output matches the expected output
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("expected: %v, got: %v", tc.expected, actual)
			}
		})
	}
}

// TestIsResourceGroupAllowed tests the IsResourceGroupAllowed function
func TestIsResourceGroupAllowed(t *testing.T) {
	// Define some test cases with inputs and expected outputs
	testCases := []struct {
		name                  string
		resourceGroup         string
		allowedResourceGroups map[string]bool
		expected              bool
	}{
		{
			name:          "resource group is allowed by non-empty map",
			resourceGroup: "apps",
			allowedResourceGroups: map[string]bool{
				"apps":              true,
				"networking.k8s.io": true,
				"policy":            true,
			},
			expected: true,
		},
		{
			name:                  "resource group is allowed by nil map",
			resourceGroup:         "apps",
			allowedResourceGroups: nil,
			expected:              true,
		},
		{
			name:          "kubestellar resource group is always allowed",
			resourceGroup: "control.kubestellar.io",
			allowedResourceGroups: map[string]bool{
				"apps":              true,
				"networking.k8s.io": true,
				"policy":            true,
			},
			expected: true,
		},
	}

	// Iterate over the test cases and run the function with the input
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsResourceGroupAllowed(tc.resourceGroup, tc.allowedResourceGroups)

			// Check if the output matches the expected output
			if actual != tc.expected {
				t.Errorf("expected: %v, got: %v", tc.expected, actual)
			}
		})
	}
}
