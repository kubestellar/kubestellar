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

func loadBalancerLocalService(annotations map[string]interface{}) *unstructured.Unstructured {
	metadata := map[string]interface{}{"name": "svc", "namespace": "ns"}
	if annotations != nil {
		metadata["annotations"] = annotations
	}
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata":   metadata,
		"spec": map[string]interface{}{
			"type":                  "LoadBalancer",
			"externalTrafficPolicy": "Local",
			"internalTrafficPolicy": "Local",
			"sessionAffinity":       "ClientIP",
			"healthCheckNodePort":   int64(30442),
			"ipFamilies":            []interface{}{"IPv4"},
			"ipFamilyPolicy":        "SingleStack",
			"clusterIP":             "10.0.0.1",
			"clusterIPs":            []interface{}{"10.0.0.1"},
			"ports": []interface{}{
				map[string]interface{}{"port": int64(80), "nodePort": int64(31000)},
			},
		},
	}}
}

func TestCleanServiceKeepsTrafficPoliciesAndDropsHealthCheckNodePort(t *testing.T) {
	svc := loadBalancerLocalService(nil)
	cleanService(svc)

	for _, field := range []string{"externalTrafficPolicy", "internalTrafficPolicy", "sessionAffinity"} {
		if _, have, _ := unstructured.NestedString(svc.Object, "spec", field); !have {
			t.Errorf("spec.%s was removed but should be kept", field)
		}
	}
	// A healthCheckNodePort without externalTrafficPolicy=Local is invalid, and
	// externalTrafficPolicy is kept, so the hub-allocated value must be dropped
	// to let the WEC allocate its own.
	for _, field := range []string{"healthCheckNodePort", "ipFamilies", "ipFamilyPolicy", "clusterIP", "clusterIPs"} {
		if _, have, _ := unstructured.NestedFieldNoCopy(svc.Object, "spec", field); have {
			t.Errorf("spec.%s was kept but should be removed", field)
		}
	}
	ports, _, _ := unstructured.NestedSlice(svc.Object, "spec", "ports")
	if np := ports[0].(map[string]interface{})["nodePort"]; np != nil {
		t.Errorf("nodePort = %v, expected nil", np)
	}
}

func TestCleanServicePreserveNodePort(t *testing.T) {
	svc := loadBalancerLocalService(map[string]interface{}{preserveFieldAnnotation: preserveNodePortValue})
	cleanService(svc)

	if hcnp, _, _ := unstructured.NestedInt64(svc.Object, "spec", "healthCheckNodePort"); hcnp != 30442 {
		t.Errorf("spec.healthCheckNodePort = %d, expected 30442", hcnp)
	}
	if etp, _, _ := unstructured.NestedString(svc.Object, "spec", "externalTrafficPolicy"); etp != "Local" {
		t.Errorf("spec.externalTrafficPolicy = %q, expected Local", etp)
	}
	ports, _, _ := unstructured.NestedSlice(svc.Object, "spec", "ports")
	if np := ports[0].(map[string]interface{})["nodePort"]; np != int64(31000) {
		t.Errorf("nodePort = %v, expected 31000", np)
	}
}

func TestCleanServiceKeepsHeadless(t *testing.T) {
	svc := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Service",
		"metadata":   map[string]interface{}{"name": "svc", "namespace": "ns"},
		"spec": map[string]interface{}{
			"clusterIP":  "None",
			"clusterIPs": []interface{}{"None"},
		},
	}}
	cleanService(svc)
	if ip, _, _ := unstructured.NestedString(svc.Object, "spec", "clusterIP"); ip != "None" {
		t.Errorf("spec.clusterIP = %q, expected None", ip)
	}
}
