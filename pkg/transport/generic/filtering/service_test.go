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

func TestCleanServicePreservesStaticClusterIPWhenAnnotated(t *testing.T) {
	service := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				preserveFieldAnnotation: "clusterip,nodeport",
			},
		},
		"spec": map[string]interface{}{
			"clusterIP":  "10.96.0.42",
			"clusterIPs": []interface{}{"10.96.0.42"},
			"ports": []interface{}{
				map[string]interface{}{"port": int64(80), "nodePort": int64(32222)},
			},
		},
	}}

	cleanService(service)

	if got, _, _ := unstructured.NestedString(service.Object, "spec", "clusterIP"); got != "10.96.0.42" {
		t.Fatalf("clusterIP = %q, want preserved static IP", got)
	}
	if got, _, _ := unstructured.NestedStringSlice(service.Object, "spec", "clusterIPs"); len(got) != 1 || got[0] != "10.96.0.42" {
		t.Fatalf("clusterIPs = %#v, want preserved static IPs", got)
	}
	ports, _, _ := unstructured.NestedSlice(service.Object, "spec", "ports")
	if got := ports[0].(map[string]interface{})["nodePort"]; got != int64(32222) {
		t.Fatalf("nodePort = %#v, want preserved static port", got)
	}
}

func TestCleanServiceStillStripsClusterIPByDefault(t *testing.T) {
	service := &unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{
			"clusterIP":  "10.96.0.42",
			"clusterIPs": []interface{}{"10.96.0.42"},
		},
	}}

	cleanService(service)

	if _, found, _ := unstructured.NestedString(service.Object, "spec", "clusterIP"); found {
		t.Fatalf("clusterIP was not stripped by default")
	}
	if _, found, _ := unstructured.NestedStringSlice(service.Object, "spec", "clusterIPs"); found {
		t.Fatalf("clusterIPs was not stripped by default")
	}
}
