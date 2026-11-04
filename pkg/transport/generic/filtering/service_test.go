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

func TestCleanServiceRemovesHealthCheckNodePort(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata":   map[string]interface{}{"name": "svc", "namespace": "ns"},
		"spec": map[string]interface{}{
			"type":                  "LoadBalancer",
			"externalTrafficPolicy": "Local",
			"healthCheckNodePort":   int64(30442),
		},
	}}
	cleanService(obj)
	for _, field := range []string{"externalTrafficPolicy", "healthCheckNodePort"} {
		if _, found, _ := unstructured.NestedFieldNoCopy(obj.Object, "spec", field); found {
			t.Errorf("spec.%s was not removed", field)
		}
	}
	if typ, _, _ := unstructured.NestedString(obj.Object, "spec", "type"); typ != "LoadBalancer" {
		t.Errorf("spec.type = %q, want LoadBalancer", typ)
	}
}
