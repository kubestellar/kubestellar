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

package filtering

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func pvObject(t *testing.T, content map[string]any) *unstructured.Unstructured {
	t.Helper()
	obj := &unstructured.Unstructured{Object: content}
	obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("PersistentVolume"))
	return obj
}

func TestCleanPersistentVolume(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name: "strips binder-owned claimRef identity and status, keeps static binding intent",
			input: map[string]any{
				"spec": map[string]any{
					"capacity":                      map[string]any{"storage": "1Gi"},
					"accessModes":                   []any{"ReadWriteOnce"},
					"persistentVolumeReclaimPolicy": "Retain",
					"storageClassName":              "manual",
					"volumeName":                    "ursa-csipv-template",
					"claimRef":                      map[string]any{"name": "ursa-claim", "namespace": "default", "uid": "hub-local-uid", "resourceVersion": "12345"},
					"persistentVolumeSource":        map[string]any{"csi": map[string]any{"driver": "example.com/csi"}},
				},
				"status": map[string]any{"phase": "Available"},
			},
			expected: map[string]any{
				"spec": map[string]any{
					"capacity":                      map[string]any{"storage": "1Gi"},
					"accessModes":                   []any{"ReadWriteOnce"},
					"persistentVolumeReclaimPolicy": "Retain",
					"storageClassName":              "manual",
					"volumeName":                    "ursa-csipv-template",
					"claimRef":                      map[string]any{"name": "ursa-claim", "namespace": "default"},
					"persistentVolumeSource":        map[string]any{"csi": map[string]any{"driver": "example.com/csi"}},
				},
			},
		},
		{
			name:     "no claimRef and no status is a no-op",
			input:    map[string]any{"spec": map[string]any{"capacity": map[string]any{"storage": "1Gi"}}},
			expected: map[string]any{"spec": map[string]any{"capacity": map[string]any{"storage": "1Gi"}}},
		},
		{
			name:     "claimRef without binder fields is untouched",
			input:    map[string]any{"spec": map[string]any{"claimRef": map[string]any{"name": "ursa-claim", "namespace": "default"}}},
			expected: map[string]any{"spec": map[string]any{"claimRef": map[string]any{"name": "ursa-claim", "namespace": "default"}}},
		},
	}

	filteringMap := NewObjectFilteringMap()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := pvObject(t, tt.input)
			filteringMap.CleanObjectSpecifics(obj)
			// SetGroupVersionKind injects apiVersion/kind into the live object;
			// mirror them into the expectation before comparing.
			want := map[string]any{
				"apiVersion": "v1",
				"kind":       "PersistentVolume",
			}
			for k, v := range tt.expected {
				want[k] = v
			}
			if !reflect.DeepEqual(obj.Object, want) {
				t.Errorf("mismatch:\n got: %#v\nwant: %#v", obj.Object, want)
			}
		})
	}
}

func TestCleanObjectSpecificsUnknownGVKIsNoOp(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{"spec": map[string]any{"foo": "bar"}}}
	obj.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("ConfigMap"))
	before := obj.DeepCopy().Object
	NewObjectFilteringMap().CleanObjectSpecifics(obj)
	if !reflect.DeepEqual(obj.Object, before) {
		t.Errorf("expected no change for unregistered GVK, got %#v", obj.Object)
	}
}
