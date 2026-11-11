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

	"github.com/kubestellar/kubestellar/pkg/abstract"
)

const (
	preserveFieldAnnotation = "control.kubestellar.io/preserve"
	preserveNodePortValue   = "nodeport"
)

func cleanService(object *unstructured.Unstructured) {
	// Fields to remove. externalTrafficPolicy, internalTrafficPolicy, and
	// sessionAffinity are deliberately not here: they are user intent that is
	// meaningful on every cluster, and removing externalTrafficPolicy would make
	// a surviving healthCheckNodePort invalid (the API server only accepts
	// healthCheckNodePort for type=LoadBalancer with externalTrafficPolicy=Local).
	fieldsToDelete := []string{"ipFamilies", "ipFamilyPolicy"}

	for _, field := range fieldsToDelete {
		unstructured.RemoveNestedField(object.Object, "spec", field)
	}

	// Keep headless Services headless, remove cluster IPs from others.
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

	// Clear the nodePort and healthCheckNodePort values unless the annotation
	// "control.kubestellar.io/preserve=nodeport" is present. Both are node port
	// numbers that are allocated by the API server of the cluster where the Service
	// lives, so a value allocated on the hub is not meaningful on a WEC.
	if !(object.GetAnnotations() != nil && object.GetAnnotations()[preserveFieldAnnotation] == preserveNodePortValue) {
		unstructured.RemoveNestedField(object.Object, "spec", "healthCheckNodePort")
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
