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

func badJobKey(key string) bool {
	return key == "controller-uid" || key == "batch.kubernetes.io/controller-uid"
}

func cleanJob(object *unstructured.Unstructured) {
	objectU := object.UnstructuredContent()
	changed := false
	selector, foundSelector, _ := unstructured.NestedMap(objectU, "spec", "selector")
	if foundSelector {
		if matchLabels, found, _ := unstructured.NestedMap(selector, "matchLabels"); found {
			if cleanLabels(matchLabels) {
				_ = unstructured.SetNestedMap(objectU, matchLabels, "spec", "selector", "matchLabels")
				changed = true
			}
		}
		if matchExpressions, found, _ := unstructured.NestedSlice(selector, "matchExpressions"); found {
			cleanedExpressions := make([]any, 0, len(matchExpressions))
			for _, expr := range matchExpressions {
				exprM := expr.(map[string]any)
				key, _, _ := unstructured.NestedString(exprM, "key")
				if !badJobKey(key) {
					cleanedExpressions = append(cleanedExpressions, expr)
				}
			}
			if len(cleanedExpressions) != len(matchExpressions) {
				unstructured.SetNestedSlice(objectU, cleanedExpressions, "spec", "selector", "matchExpressions")
				changed = true
			}
		}
	}
	podLabels, foundlabels, _ := unstructured.NestedMap(objectU, "spec", "template", "metadata", "labels")
	if foundlabels {
		if cleanLabels(podLabels) {
			_ = unstructured.SetNestedMap(objectU, podLabels, "spec", "template", "metadata", "labels")
			changed = true
		}
	}
	if changed {
		object.SetUnstructuredContent(objectU)
	}
}

func cleanLabels(labels map[string]any) bool {
	changed := false
	for key := range labels {
		if badJobKey(key) {
			delete(labels, key)
			changed = true
		}
	}
	return changed
}
