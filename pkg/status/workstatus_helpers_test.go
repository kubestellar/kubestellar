/*
Copyright 2025 The KubeStellar Authors.

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
	"encoding/json"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetObjectStatusLastUpdateTime(t *testing.T) {
	makeRaw := func(fields map[string]interface{}) *metav1.FieldsV1 {
		raw, err := json.Marshal(fields)
		if err != nil {
			t.Fatalf("json.Marshal: %v", err)
		}
		return &metav1.FieldsV1{Raw: raw}
	}

	now := metav1.NewTime(time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC))
	earlier := metav1.NewTime(time.Date(2025, 1, 10, 8, 0, 0, 0, time.UTC))

	cases := []struct {
		name          string
		managedFields []metav1.ManagedFieldsEntry
		wantZero      bool
		wantTime      *metav1.Time
	}{
		{
			name:          "no managed fields returns zero time",
			managedFields: nil,
			wantZero:      true,
		},
		{
			name: "entry without f:status key is ignored",
			managedFields: []metav1.ManagedFieldsEntry{
				{
					FieldsType: "FieldsV1",
					FieldsV1:   makeRaw(map[string]interface{}{"f:metadata": map[string]interface{}{}}),
					Time:       &now,
				},
			},
			wantZero: true,
		},
		{
			name: "entry with f:status key is picked up",
			managedFields: []metav1.ManagedFieldsEntry{
				{
					FieldsType: "FieldsV1",
					FieldsV1:   makeRaw(map[string]interface{}{"f:status": map[string]interface{}{}}),
					Time:       &now,
				},
			},
			wantZero: false,
			wantTime: &now,
		},
		{
			name: "latest time among multiple status entries wins",
			managedFields: []metav1.ManagedFieldsEntry{
				{
					FieldsType: "FieldsV1",
					FieldsV1:   makeRaw(map[string]interface{}{"f:status": map[string]interface{}{}}),
					Time:       &earlier,
				},
				{
					FieldsType: "FieldsV1",
					FieldsV1:   makeRaw(map[string]interface{}{"f:status": map[string]interface{}{}}),
					Time:       &now,
				},
			},
			wantZero: false,
			wantTime: &now,
		},
		{
			name: "nil FieldsV1 entry is skipped",
			managedFields: []metav1.ManagedFieldsEntry{
				{
					FieldsType: "FieldsV1",
					FieldsV1:   nil,
					Time:       &now,
				},
			},
			wantZero: true,
		},
		{
			name: "non-FieldsV1 type is skipped",
			managedFields: []metav1.ManagedFieldsEntry{
				{
					FieldsType: "FieldsV2",
					FieldsV1:   makeRaw(map[string]interface{}{"f:status": map[string]interface{}{}}),
					Time:       &now,
				},
			},
			wantZero: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &metav1.ObjectMeta{ManagedFields: tc.managedFields}
			got := getObjectStatusLastUpdateTime(obj)

			if tc.wantZero {
				if got != nil && !got.IsZero() {
					t.Errorf("expected zero time, got %v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil time, got nil")
			}
			if !got.Equal(tc.wantTime) {
				t.Errorf("got %v, want %v", got, tc.wantTime)
			}
		})
	}
}
