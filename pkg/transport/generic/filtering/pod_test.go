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
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCleanPodRemovesNodeName(t *testing.T) {
	object := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name": "test-pod",
		},
		"spec": map[string]interface{}{
			"nodeName":      "kubeflex-control-plane",
			"restartPolicy": "Never",
		},
	}}

	NewObjectFilteringMap().CleanObjectSpecifics(object)

	if _, found, err := unstructured.NestedFieldNoCopy(object.Object, "spec", "nodeName"); err != nil {
		t.Fatalf("reading spec.nodeName: %v", err)
	} else if found {
		t.Fatal("spec.nodeName was not removed from the Pod")
	}

	if got, found, err := unstructured.NestedString(object.Object, "spec", "restartPolicy"); err != nil {
		t.Fatalf("reading spec.restartPolicy: %v", err)
	} else if !found || got != "Never" {
		t.Fatalf("spec.restartPolicy = %q (found %t), want %q", got, found, "Never")
	}
}

func TestCleanPodWithoutNodeName(t *testing.T) {
	object := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]interface{}{
			"name": "test-pod",
		},
		"spec": map[string]interface{}{
			"restartPolicy": "Never",
		},
	}}

	NewObjectFilteringMap().CleanObjectSpecifics(object)

	if _, found, err := unstructured.NestedFieldNoCopy(object.Object, "spec", "nodeName"); err != nil {
		t.Fatalf("reading spec.nodeName: %v", err)
	} else if found {
		t.Fatal("spec.nodeName unexpectedly appeared")
	}
}
