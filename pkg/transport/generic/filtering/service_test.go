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

package filtering

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newServiceWithClusterIPAndNodePort(annotations map[string]interface{}) *unstructured.Unstructured {
	object := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata": map[string]interface{}{
			"name": "my-svc",
		},
		"spec": map[string]interface{}{
			"clusterIP":  "10.96.100.100",
			"clusterIPs": []interface{}{"10.96.100.100"},
			"ports": []interface{}{
				map[string]interface{}{
					"port":     int64(80),
					"nodePort": int64(32222),
				},
			},
		},
	}
	if len(annotations) > 0 {
		object["metadata"].(map[string]interface{})["annotations"] = annotations
	}
	return &unstructured.Unstructured{Object: object}
}

func TestCleanServiceStripsClusterIPAndNodePortByDefault(t *testing.T) {
	svc := newServiceWithClusterIPAndNodePort(nil)

	cleanService(svc)

	if _, have, _ := unstructured.NestedString(svc.Object, "spec", "clusterIP"); have {
		t.Errorf("expected spec.clusterIP to be removed")
	}
	if _, have, _ := unstructured.NestedStringSlice(svc.Object, "spec", "clusterIPs"); have {
		t.Errorf("expected spec.clusterIPs to be removed")
	}
	ports, _, _ := unstructured.NestedSlice(svc.Object, "spec", "ports")
	if got := ports[0].(map[string]interface{})["nodePort"]; got != nil {
		t.Errorf("expected nodePort to be nil, got %v", got)
	}
}

func TestCleanServicePreservesClusterIPWithAnnotation(t *testing.T) {
	svc := newServiceWithClusterIPAndNodePort(map[string]interface{}{
		preserveFieldAnnotation: "clusterip",
	})

	cleanService(svc)

	clusterIP, have, _ := unstructured.NestedString(svc.Object, "spec", "clusterIP")
	if !have || clusterIP != "10.96.100.100" {
		t.Errorf("expected spec.clusterIP to be preserved as 10.96.100.100, got %q (present=%v)", clusterIP, have)
	}
	clusterIPs, have, _ := unstructured.NestedStringSlice(svc.Object, "spec", "clusterIPs")
	if !have || len(clusterIPs) != 1 || clusterIPs[0] != "10.96.100.100" {
		t.Errorf("expected spec.clusterIPs to be preserved as [10.96.100.100], got %v (present=%v)", clusterIPs, have)
	}
	// nodePort is preserved only via its own value, not implied by clusterip.
	ports, _, _ := unstructured.NestedSlice(svc.Object, "spec", "ports")
	if got := ports[0].(map[string]interface{})["nodePort"]; got != nil {
		t.Errorf("expected nodePort to still be nil, got %v", got)
	}
}

func TestCleanServicePreservesBothWithCommaSeparatedAnnotation(t *testing.T) {
	svc := newServiceWithClusterIPAndNodePort(map[string]interface{}{
		preserveFieldAnnotation: "nodeport, clusterip",
	})

	cleanService(svc)

	clusterIP, have, _ := unstructured.NestedString(svc.Object, "spec", "clusterIP")
	if !have || clusterIP != "10.96.100.100" {
		t.Errorf("expected spec.clusterIP to be preserved, got %q (present=%v)", clusterIP, have)
	}
	ports, _, _ := unstructured.NestedSlice(svc.Object, "spec", "ports")
	nodePort := ports[0].(map[string]interface{})["nodePort"]
	nodePortInt, ok := nodePort.(int64)
	if !ok || nodePortInt != 32222 {
		t.Errorf("expected nodePort to be preserved as 32222, got %v", nodePort)
	}
}
