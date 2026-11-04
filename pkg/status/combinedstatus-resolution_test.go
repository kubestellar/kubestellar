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
	"context"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	machtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

func TestCombinedStatusResolutionDestinations(t *testing.T) {
	csr := &combinedStatusResolution{
		Name:                      "test-cs",
		workloadObjectUID:         machtypes.UID("uid-1"),
		StatusCollectorNameToData: make(map[string]*statusCollectorData),
		CollectionDestinations:    sets.New[string]("wec1", "wec2"),
	}

	// Test removing wec2, adding wec3
	removed, added := csr.setCollectionDestinations(sets.New[string]("wec1", "wec3"))
	if !removed {
		t.Error("expected removed to be true")
	}
	if !reflect.DeepEqual(added, sets.New[string]("wec3")) {
		t.Errorf("expected added to contain wec3, got: %v", added)
	}
	if !csr.CollectionDestinations.Equal(sets.New[string]("wec1", "wec3")) {
		t.Errorf("expected CollectionDestinations to be updated, got: %v", csr.CollectionDestinations)
	}

	// Test no change
	removed, added = csr.setCollectionDestinations(sets.New[string]("wec1", "wec3"))
	if removed || added != nil {
		t.Error("expected no changes")
	}
}

func TestCombinedStatusResolutionStatusCollectors(t *testing.T) {
	csr := &combinedStatusResolution{
		Name:                      "test-cs",
		workloadObjectUID:         machtypes.UID("uid-1"),
		StatusCollectorNameToData: make(map[string]*statusCollectorData),
		CollectionDestinations:    sets.New[string]("wec1"),
	}

	spec1 := &v1alpha1.StatusCollectorSpec{Limit: 10}
	spec2 := &v1alpha1.StatusCollectorSpec{Limit: 20}

	// Add collectors
	removed, added := csr.setStatusCollectors(map[string]*v1alpha1.StatusCollectorSpec{
		"sc1": spec1,
		"sc2": nil, // SC doesn't exist yet but is relevant
	})
	if removed {
		t.Error("expected removed to be false")
	}
	if !added {
		t.Error("expected added to be true")
	}

	if _, ok := csr.StatusCollectorNameToData["sc1"]; !ok {
		t.Error("expected sc1 to be present")
	}
	if csr.StatusCollectorNameToData["sc1"].collectorSpec != spec1 {
		t.Error("expected spec1 to be set on sc1")
	}

	// Update collector
	updated := csr.updateStatusCollector("sc1", spec2)
	if !updated {
		t.Error("expected update to be true")
	}
	if csr.StatusCollectorNameToData["sc1"].collectorSpec != spec2 {
		t.Error("expected spec2 to be set on sc1 after update")
	}

	// Note absence
	absent := csr.noteStatusCollectorAbsence("sc1")
	if !absent {
		t.Error("expected absent update to be true")
	}
	if csr.StatusCollectorNameToData["sc1"] != nil {
		t.Error("expected sc1 data to be nil after noting absence")
	}

	// Test nil -> non-nil transition of a collector (e.g. sc2 becomes present)
	removed, added = csr.setStatusCollectors(map[string]*v1alpha1.StatusCollectorSpec{
		"sc1": nil,
		"sc2": spec1, // previously was nil, now has spec1
	})
	if !added {
		t.Error("expected added to be true when a missing collector is provided a spec")
	}
	if csr.StatusCollectorNameToData["sc2"] == nil {
		t.Error("expected sc2 data to be allocated and non-nil")
	} else if csr.StatusCollectorNameToData["sc2"].collectorSpec != spec1 {
		t.Error("expected spec1 to be set on sc2")
	}
}

func TestEvaluateWorkStatusAndAggregation(t *testing.T) {
	celEval, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("failed to create CEL evaluator: %v", err)
	}

	objID := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Resource:   "deployments",
		ObjectName: cache.ObjectName{Namespace: "default", Name: "my-app"},
	}

	// We set up a status collector with select/filter
	filterExpr := v1alpha1.Expression("returned.status.availableReplicas > 0")
	selectSpec := &v1alpha1.StatusCollectorSpec{
		Filter: &filterExpr,
		Select: []v1alpha1.NamedExpression{
			{Name: "replicas", Def: v1alpha1.Expression("returned.status.replicas")},
		},
		Limit: 10,
	}

	csr := &combinedStatusResolution{
		Name:                      "my-app.my-binding",
		workloadObjectUID:         machtypes.UID("workload-uid"),
		StatusCollectorNameToData: make(map[string]*statusCollectorData),
		CollectionDestinations:    sets.New[string]("wec1", "wec2"),
	}

	csr.setStatusCollectors(map[string]*v1alpha1.StatusCollectorSpec{
		"sc-select": selectSpec,
	})

	ctx := context.Background()

	// 1. Evaluate with a WEC not in destinations (should be ignored)
	updated := csr.evaluateWorkStatus(ctx, celEval, objID, "my-binding", "wec3", map[string]interface{}{
		"returned": map[string]interface{}{
			"status": map[string]interface{}{
				"replicas":          int64(3),
				"availableReplicas": int64(3),
			},
		},
	})
	if updated {
		t.Error("expected no update for wec3 (not in destinations)")
	}

	// 2. Evaluate with WEC1 (matches filter)
	updated = csr.evaluateWorkStatus(ctx, celEval, objID, "my-binding", "wec1", map[string]interface{}{
		"returned": map[string]interface{}{
			"status": map[string]interface{}{
				"replicas":          int64(3),
				"availableReplicas": int64(3),
			},
		},
	})
	if !updated {
		t.Error("expected update for wec1")
	}

	// 3. Evaluate with WEC2 (does not match filter)
	updated = csr.evaluateWorkStatus(ctx, celEval, objID, "my-binding", "wec2", map[string]interface{}{
		"returned": map[string]interface{}{
			"status": map[string]interface{}{
				"replicas":          int64(2),
				"availableReplicas": int64(0),
			},
		},
	})
	if updated {
		t.Error("expected no update for wec2 since the filter failed and no prior data existed")
	}
	if _, ok := csr.StatusCollectorNameToData["sc-select"].WECToData["wec2"]; ok {
		t.Error("expected wec2 not to be in WECToData since filter failed")
	}

	// 4. Generate combined status
	cs := csr.generateCombinedStatus("my-binding", objID)
	if cs.Name != csr.Name {
		t.Errorf("expected combined status name %q, got %q", csr.Name, cs.Name)
	}
	if len(cs.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(cs.Results))
	}

	result := cs.Results[0]
	if result.Name != "sc-select" {
		t.Errorf("expected result name 'sc-select', got %q", result.Name)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row in result, got %d", len(result.Rows))
	}

	// Verify labels
	if cs.Labels["status.kubestellar.io/binding-policy"] != "my-binding" {
		t.Errorf("missing or incorrect binding-policy label: %v", cs.Labels)
	}

	// 5. Test Aggregation (COUNT, SUM, AVG)
	subjectExpr := v1alpha1.Expression("returned.status.replicas")
	aggSpec := &v1alpha1.StatusCollectorSpec{
		CombinedFields: []v1alpha1.NamedAggregator{
			{Name: "count", Type: v1alpha1.AggregatorTypeCount},
			{Name: "sum", Type: v1alpha1.AggregatorTypeSum, Subject: &subjectExpr},
			{Name: "avg", Type: v1alpha1.AggregatorTypeAvg, Subject: &subjectExpr},
			{Name: "min", Type: v1alpha1.AggregatorTypeMin, Subject: &subjectExpr},
			{Name: "max", Type: v1alpha1.AggregatorTypeMax, Subject: &subjectExpr},
		},
		Limit: 10,
	}

	csr.setStatusCollectors(map[string]*v1alpha1.StatusCollectorSpec{
		"sc-select": selectSpec,
		"sc-agg":    aggSpec,
	})

	// Evaluate WEC1 and WEC2 for aggregation
	csr.evaluateWorkStatus(ctx, celEval, objID, "my-binding", "wec1", map[string]interface{}{
		"returned": map[string]interface{}{
			"status": map[string]interface{}{
				"replicas": int64(3),
			},
		},
	})
	csr.evaluateWorkStatus(ctx, celEval, objID, "my-binding", "wec2", map[string]interface{}{
		"returned": map[string]interface{}{
			"status": map[string]interface{}{
				"replicas": int64(5),
			},
		},
	})

	csAgg := csr.generateCombinedStatus("my-binding", objID)
	// We should have sc-select (from WEC1) and sc-agg
	var aggResult *v1alpha1.NamedStatusCombination
	for i := range csAgg.Results {
		if csAgg.Results[i].Name == "sc-agg" {
			aggResult = &csAgg.Results[i]
		}
	}

	if aggResult == nil {
		t.Fatal("expected sc-agg result")
	}

	if len(aggResult.Rows) != 1 {
		t.Fatalf("expected 1 row in aggregation, got %d", len(aggResult.Rows))
	}

	row := aggResult.Rows[0]
	// Order of columns should match aggSpec.CombinedFields: [count, sum, avg, min, max]
	// count = 2
	// sum = 8
	// avg = 4
	// min = 3
	// max = 5
	expectedVals := []string{"2", "8", "4", "3", "5"}
	for i, valStr := range expectedVals {
		val := row.Columns[i]
		if val.Type != v1alpha1.TypeNumber || *val.Number != valStr {
			t.Errorf("expected column %d to be number %q, got: %+v", i, valStr, val)
		}
	}
}

func TestCompareCombinedStatus(t *testing.T) {
	celEval, err := newCELEvaluator()
	if err != nil {
		t.Fatalf("failed to create CEL evaluator: %v", err)
	}

	objID := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Resource:   "deployments",
		ObjectName: cache.ObjectName{Namespace: "default", Name: "my-app"},
	}

	selectSpec := &v1alpha1.StatusCollectorSpec{
		Select: []v1alpha1.NamedExpression{
			{Name: "replicas", Def: v1alpha1.Expression("returned.status.replicas")},
		},
		Limit: 10,
	}

	csr := &combinedStatusResolution{
		Name:                      "my-app.my-binding",
		workloadObjectUID:         machtypes.UID("workload-uid"),
		StatusCollectorNameToData: make(map[string]*statusCollectorData),
		CollectionDestinations:    sets.New[string]("wec1"),
	}

	csr.setStatusCollectors(map[string]*v1alpha1.StatusCollectorSpec{
		"sc-select": selectSpec,
	})

	ctx := context.Background()
	csr.evaluateWorkStatus(ctx, celEval, objID, "my-binding", "wec1", map[string]interface{}{
		"returned": map[string]interface{}{
			"status": map[string]interface{}{
				"replicas": int64(3),
			},
		},
	})

	// 1. Compare with nil or different labels -> should return generated status
	dummyCS := &v1alpha1.CombinedStatus{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{}, // incorrect labels
		},
	}
	diff := csr.compareCombinedStatus(dummyCS, "my-binding", objID)
	if diff == nil {
		t.Error("expected non-nil diff due to invalid labels")
	}

	// 2. Compare with identical status -> should return nil
	generated := csr.generateCombinedStatus("my-binding", objID)
	diff = csr.compareCombinedStatus(generated, "my-binding", objID)
	if diff != nil {
		t.Errorf("expected nil diff, got: %+v", diff)
	}
}
