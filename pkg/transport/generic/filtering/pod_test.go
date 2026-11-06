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

func TestCleanPodStripsNodeName(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata":   map[string]any{"name": "mypod"},
		"spec": map[string]any{
			"nodeName": "hub-node",
			"containers": []any{map[string]any{
				"name":  "c",
				"image": "nginx:1.27.1",
			}},
		},
	}}
	cleanPod(obj)

	if _, found, _ := unstructured.NestedString(obj.Object, "spec", "nodeName"); found {
		t.Errorf("expected spec.nodeName to be removed")
	}
	containers, found, err := unstructured.NestedSlice(obj.Object, "spec", "containers")
	if err != nil || !found || len(containers) == 0 {
		t.Fatalf("expected containers to be present, got %v (err=%v)", containers, err)
	}
	c0, ok := containers[0].(map[string]any)
	if !ok || c0["image"] != "nginx:1.27.1" {
		t.Errorf("expected container image to be preserved, got %v (ok=%v)", c0, ok)
	}
}
