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

package transport

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestNewWrapee(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deployment",
				"namespace": "default",
			},
		},
	}

	wr := NewWrapee(obj, true)

	if wr.Object != obj {
		t.Errorf("expected object to be set correctly")
	}

	if !wr.CreateOnly {
		t.Errorf("expected CreateOnly to be true")
	}
}

func TestNewWrapee_CreateOnlyFalse(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-configmap",
				"namespace": "default",
			},
		},
	}

	wr := NewWrapee(obj, false)

	if wr.Object != obj {
		t.Errorf("expected object to be set correctly")
	}

	if wr.CreateOnly {
		t.Errorf("expected CreateOnly to be false")
	}
}

func TestWrapee_GetObject(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "kube-system",
			},
		},
	}

	wr := Wrapee{Object: obj, CreateOnly: false}

	retrievedObj := wr.GetObject()

	if retrievedObj != obj {
		t.Errorf("expected GetObject to return the same object")
	}
}

func TestWrapee_GetID(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "nginx-deployment",
				"namespace": "production",
			},
		},
	}

	wr := Wrapee{Object: obj, CreateOnly: true}

	id := wr.GetID()

	expectedGK := schema.GroupKind{
		Group: "apps",
		Kind:  "Deployment",
	}

	if id.GK != expectedGK {
		t.Errorf("expected GroupKind %v, got %v", expectedGK, id.GK)
	}

	if id.OR.Name != "nginx-deployment" {
		t.Errorf("expected object name 'nginx-deployment', got '%s'", id.OR.Name)
	}

	if id.OR.Namespace != "production" {
		t.Errorf("expected namespace 'production', got '%s'", id.OR.Namespace)
	}
}

func TestWrapee_GetID_CoreGroup(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      "my-service",
				"namespace": "default",
			},
		},
	}

	wr := Wrapee{Object: obj, CreateOnly: false}

	id := wr.GetID()

	expectedGK := schema.GroupKind{
		Group: "",
		Kind:  "Service",
	}

	if id.GK != expectedGK {
		t.Errorf("expected GroupKind %v, got %v", expectedGK, id.GK)
	}

	if id.OR.Name != "my-service" {
		t.Errorf("expected object name 'my-service', got '%s'", id.OR.Name)
	}
}

func TestWrapee_GetID_ClusterScoped(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "rbac.authorization.k8s.io/v1",
			"kind":       "ClusterRole",
			"metadata": map[string]interface{}{
				"name": "admin",
			},
		},
	}

	wr := Wrapee{Object: obj, CreateOnly: false}

	id := wr.GetID()

	expectedGK := schema.GroupKind{
		Group: "rbac.authorization.k8s.io",
		Kind:  "ClusterRole",
	}

	if id.GK != expectedGK {
		t.Errorf("expected GroupKind %v, got %v", expectedGK, id.GK)
	}

	if id.OR.Name != "admin" {
		t.Errorf("expected object name 'admin', got '%s'", id.OR.Name)
	}

	if id.OR.Namespace != "" {
		t.Errorf("expected empty namespace for cluster-scoped object, got '%s'", id.OR.Namespace)
	}
}

func TestWrapee_CreateOnlyField(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name": "test-config",
			},
		},
	}

	wr1 := Wrapee{Object: obj, CreateOnly: true}
	if !wr1.CreateOnly {
		t.Errorf("expected CreateOnly to be true")
	}

	wr2 := Wrapee{Object: obj, CreateOnly: false}
	if wr2.CreateOnly {
		t.Errorf("expected CreateOnly to be false")
	}
}
