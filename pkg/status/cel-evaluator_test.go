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

package status

import (
	"strings"
	"testing"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func TestNewCELEvaluator(t *testing.T) {
	eval, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("failed to create CEL evaluator: %v", err)
	}
	if eval == nil {
		t.Fatal("expected non-nil CEL evaluator")
	}
	if eval.env == nil {
		t.Fatal("expected non-nil CEL environment")
	}
}

func TestCheckExpression(t *testing.T) {
	eval, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("failed to create CEL evaluator: %v", err)
	}

	testCases := []struct {
		name       string
		expression *v1alpha1.Expression
		expectErr  bool
		errSubstr  string
	}{
		{
			name:       "nil expression",
			expression: nil,
			expectErr:  false,
		},
		{
			name:       "valid expression - check obj name",
			expression: exprPtr("obj.metadata.name == 'test'"),
			expectErr:  false,
		},
		{
			name:       "valid expression - check returned replicas",
			expression: exprPtr("returned.status.replicas == 3"),
			expectErr:  false,
		},
		{
			name:       "invalid syntax (parsing error)",
			expression: exprPtr("obj.metadata.name == 'test"),
			expectErr:  true,
			errSubstr:  "failed to parse expression",
		},
		{
			name:       "invalid check (unknown variable)",
			expression: exprPtr("unknown_variable == 'test'"),
			expectErr:  true,
			errSubstr:  "failed to check expression",
		},
		{
			name:       "invalid check (type mismatch)",
			expression: exprPtr("obj.metadata.name == 123"),
			expectErr:  false, // CEL maps dynamic typed keys, so obj.metadata.name is dyn type and can be compared to int in check phase, but might fail at evaluation.
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := eval.CheckExpression(tc.expression)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tc.errSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	eval, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("failed to create CEL evaluator: %v", err)
	}

	contextMap := map[string]interface{}{
		"obj": map[string]interface{}{
			"metadata": map[string]interface{}{
				"name": "my-service",
			},
			"spec": map[string]interface{}{
				"port": int64(80),
			},
		},
		"returned": map[string]interface{}{
			"status": map[string]interface{}{
				"replicas":          int64(3),
				"availableReplicas": int64(3),
			},
		},
		"inventory": map[string]string{
			"name": "wec1",
		},
		"propagation": map[string]interface{}{
			"lastReturnedUpdateTimestamp": "2026-07-22T20:00:00Z",
		},
	}

	testCases := []struct {
		name       string
		expression v1alpha1.Expression
		expectErr  bool
		errSubstr  string
		expected   interface{}
	}{
		{
			name:       "evaluate bool - true",
			expression: v1alpha1.Expression("returned.status.replicas == 3"),
			expected:   true,
		},
		{
			name:       "evaluate bool - false",
			expression: v1alpha1.Expression("returned.status.replicas == 5"),
			expected:   false,
		},
		{
			name:       "evaluate string extraction",
			expression: v1alpha1.Expression("obj.metadata.name"),
			expected:   "my-service",
		},
		{
			name:       "evaluate int extraction",
			expression: v1alpha1.Expression("obj.spec.port"),
			expected:   int64(80),
		},
		{
			name:       "evaluate inventory variable",
			expression: v1alpha1.Expression("inventory.name"),
			expected:   "wec1",
		},
		{
			name:       "evaluate propagation variable",
			expression: v1alpha1.Expression("propagation.lastReturnedUpdateTimestamp"),
			expected:   "2026-07-22T20:00:00Z",
		},
		{
			name:       "invalid syntax",
			expression: v1alpha1.Expression("returned.status.replicas =="),
			expectErr:  true,
			errSubstr:  "failed to parse expression",
		},
		{
			name:       "invalid variable in expression",
			expression: v1alpha1.Expression("unknown_var"),
			expectErr:  true,
			errSubstr:  "failed to check expression",
		},
		{
			name:       "runtime error - divide by zero",
			expression: v1alpha1.Expression("1 / 0 == 0"),
			expectErr:  true,
			errSubstr:  "failed to evaluate expression",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := eval.Evaluate(tc.expression, contextMap)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tc.errSubstr) {
					t.Errorf("expected error containing %q, got: %v", tc.errSubstr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if val == nil {
					t.Fatal("expected non-nil value")
				}
				got := val.Value()
				if got != tc.expected {
					t.Errorf("expected value %v (%T), got %v (%T)", tc.expected, tc.expected, got, got)
				}
			}
		})
	}
}

func exprPtr(s string) *v1alpha1.Expression {
	e := v1alpha1.Expression(s)
	return &e
}
