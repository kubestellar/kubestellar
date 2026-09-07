/*
Copyright 2023 The KubeStellar Authors.

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

package transport

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

func TestNewWrapee(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-cm",
				"namespace": "default",
			},
		},
	}

	testCases := []struct {
		name       string
		object     *unstructured.Unstructured
		createOnly bool
	}{
		{
			name:       "createOnly is true",
			object:     obj,
			createOnly: true,
		},
		{
			name:       "createOnly is false",
			object:     obj,
			createOnly: false,
		},
		{
			name:       "nil object",
			object:     nil,
			createOnly: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapee := NewWrapee(tc.object, tc.createOnly)
			if wrapee.Object != tc.object {
				t.Errorf("expected Object %v, got %v", tc.object, wrapee.Object)
			}
			if wrapee.CreateOnly != tc.createOnly {
				t.Errorf("expected CreateOnly %v, got %v", tc.createOnly, wrapee.CreateOnly)
			}
		})
	}
}

func TestWrapeeGetObject(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "my-deploy",
				"namespace": "prod",
			},
		},
	}

	wrapee := NewWrapee(obj, true)
	if got := wrapee.GetObject(); got != obj {
		t.Errorf("expected GetObject() to return %v, got %v", obj, got)
	}

	nilWrapee := NewWrapee(nil, false)
	if got := nilWrapee.GetObject(); got != nil {
		t.Errorf("expected GetObject() on nil Wrapee to return nil, got %v", got)
	}
}

func TestWrapeeGetID(t *testing.T) {
	testCases := []struct {
		name        string
		object      *unstructured.Unstructured
		expectedGK  schema.GroupKind
		expectedRef klog.ObjectRef
	}{
		{
			name: "Namespaced object with Group",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "apps/v1",
					"kind":       "Deployment",
					"metadata": map[string]interface{}{
						"name":      "nginx-deployment",
						"namespace": "test-namespace",
					},
				},
			},
			expectedGK: schema.GroupKind{
				Group: "apps",
				Kind:  "Deployment",
			},
			expectedRef: klog.ObjectRef{
				Namespace: "test-namespace",
				Name:      "nginx-deployment",
			},
		},
		{
			name: "Cluster-scoped object with Group",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "rbac.authorization.k8s.io/v1",
					"kind":       "ClusterRole",
					"metadata": map[string]interface{}{
						"name": "cluster-admin-role",
					},
				},
			},
			expectedGK: schema.GroupKind{
				Group: "rbac.authorization.k8s.io",
				Kind:  "ClusterRole",
			},
			expectedRef: klog.ObjectRef{
				Namespace: "",
				Name:      "cluster-admin-role",
			},
		},
		{
			name: "Core Group resource with empty group",
			object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Pod",
					"metadata": map[string]interface{}{
						"name":      "my-pod",
						"namespace": "kube-system",
					},
				},
			},
			expectedGK: schema.GroupKind{
				Group: "",
				Kind:  "Pod",
			},
			expectedRef: klog.ObjectRef{
				Namespace: "kube-system",
				Name:      "my-pod",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapee := NewWrapee(tc.object, false)
			id := wrapee.GetID()

			expectedID := util.GKObjRef{
				GK: tc.expectedGK,
				OR: tc.expectedRef,
			}

			if id != expectedID {
				t.Errorf("expected GetID() = %#v, got %#v", expectedID, id)
			}
		})
	}
}
