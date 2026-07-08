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
	"strings"
	"testing"

	celtypes "github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

func TestGetCombinedFieldSubject_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		aggregator  v1alpha1.NamedAggregator
		row         map[string]ref.Val
		expectedVal *float64
		expectedErr string
	}{
		{
			name: "Success - float value",
			aggregator: v1alpha1.NamedAggregator{
				Name: "test-agg",
				Type: v1alpha1.AggregatorTypeSum,
			},
			row: map[string]ref.Val{
				"test-agg": celtypes.Double(12.34),
			},
			expectedVal: func() *float64 { f := 12.34; return &f }(),
			expectedErr: "",
		},
		{
			name: "Success - int value",
			aggregator: v1alpha1.NamedAggregator{
				Name: "test-agg",
				Type: v1alpha1.AggregatorTypeSum,
			},
			row: map[string]ref.Val{
				"test-agg": celtypes.Int(42),
			},
			expectedVal: func() *float64 { f := 42.0; return &f }(),
			expectedErr: "",
		},
		{
			name: "Success - parsable string value",
			aggregator: v1alpha1.NamedAggregator{
				Name: "test-agg",
				Type: v1alpha1.AggregatorTypeSum,
			},
			row: map[string]ref.Val{
				"test-agg": celtypes.String("123.45"),
			},
			expectedVal: func() *float64 { f := 123.45; return &f }(),
			expectedErr: "",
		},
		{
			name: "Failure - unparsable string value",
			aggregator: v1alpha1.NamedAggregator{
				Name: "test-agg",
				Type: v1alpha1.AggregatorTypeSum,
			},
			row: map[string]ref.Val{
				"test-agg": celtypes.String("not-a-number"),
			},
			expectedVal: nil,
			expectedErr: "failed to parse combinedField subject as a float",
		},
		{
			name: "Failure - unexpected type",
			aggregator: v1alpha1.NamedAggregator{
				Name: "test-agg",
				Type: v1alpha1.AggregatorTypeSum,
			},
			row: map[string]ref.Val{
				"test-agg": celtypes.False,
			},
			expectedVal: nil,
			expectedErr: "combinedField subject has unexpected type bool",
		},
		{
			name: "Nil evaluation returns nil without error",
			aggregator: v1alpha1.NamedAggregator{
				Name: "test-agg",
				Type: v1alpha1.AggregatorTypeSum,
			},
			row:         map[string]ref.Val{},
			expectedVal: nil,
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, errStr := getCombinedFieldSubject(tt.aggregator, tt.row)
			if tt.expectedErr == "" {
				if errStr != "" {
					t.Errorf("expected no error, got: %s", errStr)
				}
				if tt.expectedVal == nil {
					if val != nil {
						t.Errorf("expected nil value, got: %v", *val)
					}
				} else {
					if val == nil || *val != *tt.expectedVal {
						t.Errorf("expected value %v, got: %v", *tt.expectedVal, val)
					}
				}
			} else {
				if errStr == "" || !strings.Contains(errStr, tt.expectedErr) {
					t.Errorf("expected error containing %q, got: %q", tt.expectedErr, errStr)
				}
				if val != nil {
					t.Errorf("expected nil value on error, got: %v", val)
				}
			}
		})
	}
}

func TestNoteWorkStatus_ErrorHandling(t *testing.T) {
	// Setup a resolver with a fake/empty listers map to trigger getCombinedContentMap errors
	listers := util.NewConcurrentMap[schema.GroupVersionResource, cache.GenericLister]()
	resolver := NewCombinedStatusResolver(nil, listers)

	r := resolver.(*combinedStatusResolver)

	objID := util.ObjectIdentifier{
		GVK: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		ObjectName: cache.ObjectName{
			Namespace: "default",
			Name:      "my-deploy",
		},
	}

	r.bindingNameToResolutions["binding1"] = map[util.ObjectIdentifier]*combinedStatusResolution{
		objID: {
			Name: "cs-name",
			StatusCollectorNameToData: map[string]*statusCollectorData{
				"sc1": {
					collectorSpec: &v1alpha1.StatusCollectorSpec{
						// sourceObjectRequired=true triggers getObjectMetaAndSpec
						Select: []v1alpha1.NamedExpression{
							{
								Name: "col1",
								Def:  "obj.metadata.name",
							},
						},
					},
				},
			},
		},
	}

	ws := &workStatus{
		workStatusRef: workStatusRef{
			Name:                   "ws-name",
			WECName:                "wec1",
			SourceObjectIdentifier: objID,
		},
	}

	// We expect NoteWorkStatus to fail because the GVR lister is missing from listers
	_, err := r.NoteWorkStatus(context.Background(), ws)
	if err == nil {
		t.Fatal("expected NoteWorkStatus to return an error due to missing lister, but got nil")
	}
	expectedErr := "failed to get combined content map"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error containing %q, got: %v", expectedErr, err)
	}
}
