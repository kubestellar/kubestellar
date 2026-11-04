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

	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func nativeToRefVal(native any) ref.Val {
	return celtypes.DefaultTypeAdapter.NativeToValue(native)
}

// scDataGroupingByList builds a statusCollectorData that groups by a single
// expression whose evaluated value is a list (a composite, non-comparable
// value). Each WEC's groupBy value is provided by wecToList.
func scDataGroupingByList(wecToList map[string][]any) *statusCollectorData {
	wecToData := map[string]*workStatusData{}
	for wec, list := range wecToList {
		wecToData[wec] = &workStatusData{
			groupByEval:        rowFragment{"g": nativeToRefVal(list)},
			combinedFieldsEval: rowFragment{"count": nil},
		}
	}
	return &statusCollectorData{
		collectorSpec: &v1alpha1.StatusCollectorSpec{
			GroupBy: []v1alpha1.NamedExpression{{Name: "g", Def: "obj.spec.list"}},
			CombinedFields: []v1alpha1.NamedAggregator{
				{Name: "count", Type: v1alpha1.AggregatorTypeCount},
			},
		},
		WECToData: wecToData,
	}
}

// TestHandleAggregationGroupByComposite verifies that a groupBy value that is a
// composite (list) type does not panic and that equal values are grouped
// together while distinct values are grouped apart. Before the fix, the raw
// composite value was used as a map key and this panicked with
// "hash of unhashable type".
func TestHandleAggregationGroupByComposite(t *testing.T) {
	// wec1 and wec2 share the same list value; wec3 differs. Expect 2 groups.
	sc := scDataGroupingByList(map[string][]any{
		"wec1": {"a", "b"},
		"wec2": {"a", "b"},
		"wec3": {"a", "c"},
	})

	var result *v1alpha1.NamedStatusCombination
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("handleAggregationReadLocked panicked on composite groupBy value: %v", r)
			}
		}()
		result = handleAggregationReadLocked("sc", sc)
	}()

	if result == nil {
		t.Fatal("expected a NamedStatusCombination, got nil")
	}
	if got := len(result.Rows); got != 2 {
		t.Fatalf("expected 2 groups (two WECs share a value, one differs), got %d rows", got)
	}
}

// TestGroupByValueKey checks the key helper directly: scalars keep distinct
// keys by type and value, composites are reduced to a stable JSON key, and
// equal composites collide while different ones do not.
func TestGroupByValueKey(t *testing.T) {
	if groupByValueKey(nil) != "null" {
		t.Error("nil value should key to \"null\"")
	}

	listAB1 := groupByValueKey(nativeToRefVal([]any{"a", "b"}))
	listAB2 := groupByValueKey(nativeToRefVal([]any{"a", "b"}))
	listAC := groupByValueKey(nativeToRefVal([]any{"a", "c"}))
	if listAB1 != listAB2 {
		t.Errorf("equal lists should produce equal keys: %q vs %q", listAB1, listAB2)
	}
	if listAB1 == listAC {
		t.Errorf("different lists should produce different keys, both were %q", listAB1)
	}

	mapKey := groupByValueKey(nativeToRefVal(map[string]any{"x": 1}))
	if mapKey == listAB1 {
		t.Error("a map and a list should not share a key")
	}

	// A string scalar "a,b" must not collide with the list ["a","b"].
	if groupByValueKey(nativeToRefVal("a,b")) == listAB1 {
		t.Error("scalar string should not collide with a list key")
	}
}
