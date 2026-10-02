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
	"testing"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func TestInventoryWithNonStringValue(t *testing.T) {
	evaluator, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("failed to create evaluator: %v", err)
	}

	// inventory values are not always strings; they can carry numeric
	// fields such as CPU count or memory. This expression was rejected
	// before the fix because inventory was typed as map[string]string.
	expr := v1alpha1.Expression(`inventory["cpu"] == 4`)

	err = evaluator.CheckExpression(&expr)
	if err != nil {
		t.Fatalf("CheckExpression rejected a valid expression: %v", err)
	}

	objMap := map[string]interface{}{
		sourceObjectKey:    map[string]interface{}{},
		returnedKey:        map[string]interface{}{},
		inventoryKey:       map[string]interface{}{"cpu": int64(4)},
		propagationMetaKey: map[string]interface{}{},
	}

	result, err := evaluator.Evaluate(expr, objMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if result.Value() != true {
		t.Errorf("expected true, got %v", result.Value())
	}
}

func TestInventoryWithStringValueStillWorks(t *testing.T) {
	evaluator, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("failed to create evaluator: %v", err)
	}

	// existing string-only inventory users must not be broken by the fix.
	expr := v1alpha1.Expression(`inventory["region"] == "us-east-1"`)

	err = evaluator.CheckExpression(&expr)
	if err != nil {
		t.Fatalf("CheckExpression rejected a valid expression: %v", err)
	}

	objMap := map[string]interface{}{
		sourceObjectKey:    map[string]interface{}{},
		returnedKey:        map[string]interface{}{},
		inventoryKey:       map[string]interface{}{"region": "us-east-1"},
		propagationMetaKey: map[string]interface{}{},
	}

	result, err := evaluator.Evaluate(expr, objMap)
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}

	if result.Value() != true {
		t.Errorf("expected true, got %v", result.Value())
	}
}
