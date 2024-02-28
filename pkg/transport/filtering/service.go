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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	preserveFieldAnnotation = "control.kubestellar.io/preserve"
	preserveNodePortValue   = "nodeport"
)

func cleanService(object *unstructured.Unstructured) {
	// Fields to remove
	fieldsToDelete := []string{"clusterIP", "clusterIPs", "ipFamilies",
		"externalTrafficPolicy", "internalTrafficPolicy", "ipFamilyPolicy", "sessionAffinity"}

	for _, field := range fieldsToDelete {
		unstructured.RemoveNestedField(object.Object, "spec", field)
	}

	// Set the nodePort to an empty string unelss the annotation "kubestellar.io/annotations/preserve=nodeport" is present
	if !(object.GetAnnotations() != nil && object.GetAnnotations()[preserveFieldAnnotation] == preserveNodePortValue) {
		if ports, found, _ := unstructured.NestedSlice(object.Object, "spec", "ports"); found {
			for i, port := range ports {
				if portMap, ok := port.(map[string]interface{}); ok {
					portMap["nodePort"] = nil
					ports[i] = portMap
				}
			}
			unstructured.SetNestedSlice(object.Object, ports, "spec", "ports")
		}
	}
}
