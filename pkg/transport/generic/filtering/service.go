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
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/kubestellar/kubestellar/pkg/abstract"
)

const (
	preserveFieldAnnotation = "control.kubestellar.io/preserve"
	preserveNodePortValue   = "nodeport"
	preserveClusterIPValue  = "clusterip"
)

func cleanService(object *unstructured.Unstructured) {
	// Fields to remove
	fieldsToDelete := []string{"ipFamilies",
		"externalTrafficPolicy", "internalTrafficPolicy", "ipFamilyPolicy", "sessionAffinity"}

	for _, field := range fieldsToDelete {
		unstructured.RemoveNestedField(object.Object, "spec", field)
	}

	preserveClusterIP := hasPreserveValue(object, preserveClusterIPValue)

	// Keep headless Services headless, remove cluster IPs from others unless explicitly preserved.
	if !preserveClusterIP {
		if val, have, _ := unstructured.NestedString(object.Object, "spec", "clusterIP"); have && val != "None" {
			unstructured.RemoveNestedField(object.Object, "spec", "clusterIP")
		}
		if val, have, _ := unstructured.NestedStringSlice(object.Object, "spec", "clusterIPs"); have {
			newVal := abstract.NewSliceByFilter(val, func(ip string) bool { return ip == "None" })
			if len(newVal) == 0 {
				unstructured.RemoveNestedField(object.Object, "spec", "clusterIPs")
			} else {
				unstructured.SetNestedStringSlice(object.Object, newVal, "spec", "clusterIPs")
			}
		}
	}

	// Set the nodePort to an empty string unless the annotation "control.kubestellar.io/preserve=nodeport" is present
	if !hasPreserveValue(object, preserveNodePortValue) {
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

func hasPreserveValue(object *unstructured.Unstructured, want string) bool {
	annotations := object.GetAnnotations()
	if annotations == nil {
		return false
	}
	for _, value := range strings.FieldsFunc(annotations[preserveFieldAnnotation], func(r rune) bool { return r == ',' || r == ' ' }) {
		if value == want {
			return true
		}
	}
	return false
}
