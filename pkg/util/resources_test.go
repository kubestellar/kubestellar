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
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
)

// TestParseAPIGroupsString tests the TestParseAPIGroupsString function
func TestParseAPIGroupsString(t *testing.T) {
	// Define some test cases with inputs and expected outputs
	testCases := []struct {
		name     string
		input    string
		expected sets.Set[string]
	}{
		{
			name:     "valid input with api group",
			input:    "apps,networking.k8s.io,policy",
			expected: sets.New("apps", "networking.k8s.io", "policy", "control.kubestellar.io", "apiextensions.k8s.io"),
		},
		{
			name:     "valid input with empty api group",
			input:    "apps,,policy",
			expected: sets.New("apps", "", "policy", "control.kubestellar.io", "apiextensions.k8s.io"),
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
			actual := ParseAPIGroupsString(tc.input)

			// Check if the output matches the expected output
			if !actual.Equal(tc.expected) {
				t.Errorf("expected: %v, got: %v", tc.expected, actual)
			}
		})
	}
}

// TestIsAPIGroupAllowed tests the IsAPIGroupAllowed function
func TestIsAPIGroupAllowed(t *testing.T) {
	// Define some test cases with inputs and expected outputs
	testCases := []struct {
		name             string
		resourceGroup    string
		allowedAPIGroups sets.Set[string]
		expected         bool
	}{
		{
			name:             "api group is allowed by non-empty map",
			resourceGroup:    "apps",
			allowedAPIGroups: sets.New("apps", "networking.k8s.io", "policy"),
			expected:         true,
		},
		{
			name:             "api group is allowed by nil map",
			resourceGroup:    "apps",
			allowedAPIGroups: nil,
			expected:         true,
		},
	}

	// Iterate over the test cases and run the function with the input
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsAPIGroupAllowed(tc.resourceGroup, tc.allowedAPIGroups)

			// Check if the output matches the expected output
			if actual != tc.expected {
				t.Errorf("expected: %v, got: %v", tc.expected, actual)
			}
		})
	}
}
