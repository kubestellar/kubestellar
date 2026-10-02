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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/cel-go/common/types"
	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func TestCheckExpression(t *testing.T) {
	evaluator, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("Failed to initialize CEL evaluator: %v", err)
	}

	tests := []struct {
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
			name:       "empty expression should fail validation",
			expression: func() *v1alpha1.Expression { e := v1alpha1.Expression(""); return &e }(),
			expectErr:  true,
			errSubstr:  "Syntax error",
		},
		{
			name:       "valid expression referencing obj",
			expression: func() *v1alpha1.Expression { e := v1alpha1.Expression("obj.metadata.name == 'foo'"); return &e }(),
			expectErr:  false,
		},
		{
			name: "valid expression referencing returned and inventory",
			expression: func() *v1alpha1.Expression {
				e := v1alpha1.Expression("returned.status.replicas > 0 && inventory.name == 'cluster1'")
				return &e
			}(),
			expectErr: false,
		},
		{
			name:       "invalid syntax expression",
			expression: func() *v1alpha1.Expression { e := v1alpha1.Expression("obj.metadata.name =="); return &e }(),
			expectErr:  true,
			errSubstr:  "mismatched input",
		},
		{
			name: "type check mismatch (int compared to string)",
			expression: func() *v1alpha1.Expression {
				e := v1alpha1.Expression("inventory.name == 123")
				return &e
			}(),
			expectErr: true,
			errSubstr: "found no matching overload for '_==_'",
		},
		{
			name: "type check mismatch (comparing map to int)",
			expression: func() *v1alpha1.Expression {
				e := v1alpha1.Expression("obj == 123")
				return &e
			}(),
			expectErr: true,
			errSubstr: "found no matching overload for '_==_'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := evaluator.CheckExpression(tt.expression)
			if (err != nil) != tt.expectErr {
				t.Errorf("CheckExpression() error = %v, expectErr = %v", err, tt.expectErr)
				return
			}
			if err != nil && tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("CheckExpression() error = %v, expected to contain %q", err, tt.errSubstr)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	evaluator, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("Failed to initialize CEL evaluator: %v", err)
	}

	// Setup evaluation context simulating a typical WDS/WEC status reconciliation scenario
	now := time.Now()
	protoTimestamp := timestamppb.New(now)

	contextMap := map[string]interface{}{
		// obj represents the desired workload state in the WDS (core cluster)
		"obj": map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "my-deployment",
				"namespace": "default",
				"labels": map[string]interface{}{
					"app": "nginx",
				},
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
			},
		},
		// returned represents the reported state back from the WEC (managed cluster)
		"returned": map[string]interface{}{
			"status": map[string]interface{}{
				"readyReplicas":   int64(2),
				"updatedReplicas": int64(2),
				"conditions": []interface{}{
					map[string]interface{}{
						"type":   "Ready",
						"status": "True",
					},
					map[string]interface{}{
						"type":   "Progressing",
						"status": "True",
					},
				},
			},
		},
		// inventory represents WEC cluster record
		"inventory": map[string]string{
			"name": "cluster-east",
		},
		// propagation metadata
		"propagation": map[string]interface{}{
			"lastReturnedUpdateTimestamp": protoTimestamp,
		},
	}

	tests := []struct {
		name       string
		expression v1alpha1.Expression
		expectVal  interface{}
		expectErr  bool
		errSubstr  string
	}{
		// KubeStellar Business Logic Scenarios
		{
			name:       "ArgoCD-style Ready condition list traversal check",
			expression: "returned.status.conditions.exists(c, c.type == 'Ready' && c.status == 'True')",
			expectVal:  true,
			expectErr:  false,
		},
		{
			name:       "replica synchronization status check (mismatched states)",
			expression: "obj.spec.replicas == returned.status.readyReplicas",
			expectVal:  false,
			expectErr:  false,
		},
		{
			name:       "replica synchronization status check (offset/arithmetic check)",
			expression: "obj.spec.replicas == returned.status.readyReplicas + 1",
			expectVal:  true,
			expectErr:  false,
		},
		{
			name:       "propagation timestamp exists check",
			expression: "propagation.lastReturnedUpdateTimestamp != null",
			expectVal:  true,
			expectErr:  false,
		},
		// Standard CEL engine access patterns
		{
			name:       "evaluate boolean equality",
			expression: "obj.metadata.name == 'my-deployment'",
			expectVal:  true,
			expectErr:  false,
		},
		{
			name:       "evaluate string map property",
			expression: "inventory.name == 'cluster-east'",
			expectVal:  true,
			expectErr:  false,
		},
		// Map safety & error boundary tests
		{
			name:       "accessing missing keys in map returns error (verifies error-on-missing-field constraint)",
			expression: "obj.metadata.missingKey == 'foo'",
			expectVal:  nil,
			expectErr:  true,
			errSubstr:  "no such key",
		},
		{
			name:       "safely checking missing keys using standard in operator",
			expression: "'missingKey' in obj.metadata",
			expectVal:  false,
			expectErr:  false,
		},
		// Error Handling and Compilation Boundaries
		{
			name:       "invalid expression syntax error",
			expression: "obj.metadata.name == ",
			expectVal:  nil,
			expectErr:  true,
			errSubstr:  "mismatched input",
		},
		{
			name:       "expression type checking compilation error",
			expression: "inventory.name == 123",
			expectVal:  nil,
			expectErr:  true,
			errSubstr:  "found no matching overload",
		},
		{
			name:       "verifies KubeStellar does not panic on runtime evaluation errors (e.g. division by zero)",
			expression: "obj.spec.replicas / 0 == 0",
			expectVal:  nil,
			expectErr:  true,
			errSubstr:  "division by zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := evaluator.Evaluate(tt.expression, contextMap)
			if (err != nil) != tt.expectErr {
				t.Errorf("Evaluate() error = %v, expectErr = %v", err, tt.expectErr)
				return
			}
			if tt.expectErr {
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("Evaluate() error = %v, expected to contain %q", err, tt.errSubstr)
				}
				return
			}

			if val == nil {
				t.Errorf("Expected non-nil value for expression %s", tt.expression)
				return
			}

			// Direct assertions on primitives (types.NullValue / bool / int64 / string)
			var actual interface{}
			if val == types.NullValue {
				actual = nil
			} else {
				actual = val.Value()
			}

			if actual != tt.expectVal {
				t.Errorf("Evaluate() got = %v (type %T), want = %v (type %T)", actual, actual, tt.expectVal, tt.expectVal)
			}
		})
	}
}
