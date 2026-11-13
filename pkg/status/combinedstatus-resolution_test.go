/*
Copyright 2026 The KubeStellar Authors.

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

func namedAgg(aggType v1alpha1.AggregatorType) v1alpha1.NamedAggregator {
	return v1alpha1.NamedAggregator{Name: "field", Type: aggType}
}

func rowFragmentWith(subject ref.Val) rowFragment {
	if subject == nil {
		return rowFragment{}
	}
	return rowFragment{"field": subject}
}

func destination(clusterId string) v1alpha1.Destination {
	return v1alpha1.Destination{ClusterId: clusterId}
}

func TestCalculateCombinedFieldAggregation(t *testing.T) {
	testCases := []struct {
		name    string
		agg     v1alpha1.NamedAggregator
		rows    map[v1alpha1.Destination]rowFragment
		wantNum string
		wantErr string
	}{
		{
			name:    "COUNT counts all rows",
			agg:     namedAgg(v1alpha1.AggregatorTypeCount),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(celtypes.Int(5)), destination("w2"): rowFragmentWith(celtypes.Int(7))},
			wantNum: "2",
		},
		{
			name:    "SUM skips rows without a subject",
			agg:     namedAgg(v1alpha1.AggregatorTypeSum),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(celtypes.Int(5)), destination("w2"): rowFragmentWith(nil)},
			wantNum: "5",
		},
		{
			name:    "AVG uses the number of valid subjects as the denominator",
			agg:     namedAgg(v1alpha1.AggregatorTypeAvg),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(celtypes.Int(5)), destination("w2"): rowFragmentWith(celtypes.Int(5)), destination("w3"): rowFragmentWith(nil)},
			wantNum: "5",
		},
		{
			name:    "AVG of no valid subjects reports an error",
			agg:     namedAgg(v1alpha1.AggregatorTypeAvg),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(nil), destination("w2"): rowFragmentWith(nil)},
			wantNum: "NaN",
			wantErr: "no values to average",
		},
		{
			name:    "MIN ignores rows without a subject",
			agg:     namedAgg(v1alpha1.AggregatorTypeMin),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(celtypes.Int(9)), destination("w2"): rowFragmentWith(celtypes.Int(3)), destination("w3"): rowFragmentWith(nil)},
			wantNum: "3",
		},
		{
			name:    "MIN of no valid subjects reports an error instead of +Inf",
			agg:     namedAgg(v1alpha1.AggregatorTypeMin),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(nil)},
			wantNum: "+Inf",
			wantErr: "no values to take minimum of",
		},
		{
			name:    "MAX ignores rows without a subject",
			agg:     namedAgg(v1alpha1.AggregatorTypeMax),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(celtypes.Int(9)), destination("w2"): rowFragmentWith(celtypes.Int(3)), destination("w3"): rowFragmentWith(nil)},
			wantNum: "9",
		},
		{
			name:    "MAX of no valid subjects reports an error instead of -Inf",
			agg:     namedAgg(v1alpha1.AggregatorTypeMax),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(nil)},
			wantNum: "-Inf",
			wantErr: "no values to take maximum of",
		},
		{
			name:    "aggregation over string subjects parsed as floats",
			agg:     namedAgg(v1alpha1.AggregatorTypeSum),
			rows:    map[v1alpha1.Destination]rowFragment{destination("w1"): rowFragmentWith(celtypes.String("2.5")), destination("w2"): rowFragmentWith(celtypes.String("0.5"))},
			wantNum: "3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value, errStr := calculateCombinedFieldAggregation(tc.agg, tc.rows)
			if value.Number == nil || *value.Number != tc.wantNum {
				t.Fatalf("expected number %q, got %v", tc.wantNum, value.Number)
			}
			if errStr != tc.wantErr {
				t.Fatalf("expected error %q, got %q", tc.wantErr, errStr)
			}
		})
	}
}
