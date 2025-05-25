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
	if jobAnnotations, found, _ := unstructured.NestedMap(objectU, "metadata", "annotations"); found {
		if cleanJobAnnotations(jobAnnotations) {
			_ = unstructured.SetNestedMap(objectU, jobAnnotations, "metadata", "annotations")
			changed = true
		}
	}
	if jobLabels, found, _ := unstructured.NestedMap(objectU, "metadata", "labels"); found {
		if cleanJobLabels(jobLabels) {
			_ = unstructured.SetNestedMap(objectU, jobLabels, "metadata", "labels")
			changed = true
		}
	}
	_, foundSelector, _ := unstructured.NestedFieldNoCopy(objectU, "spec", "selector")
	if foundSelector {
		unstructured.RemoveNestedField(objectU, "spec", "selector")
		changed = true
	}
	if _, found, _ := unstructured.NestedFieldNoCopy(objectU, "spec", "suspend"); found {
		unstructured.RemoveNestedField(objectU, "spec", "suspend")
		changed = true
	}
	podLabels, foundlabels, _ := unstructured.NestedMap(objectU, "spec", "template", "metadata", "labels")
	if foundlabels {
		if cleanJobLabels(podLabels) {
			_ = unstructured.SetNestedMap(objectU, podLabels, "spec", "template", "metadata", "labels")
			changed = true
		}
	}
	if _, foundStatus, _ := unstructured.NestedFieldNoCopy(objectU, "status"); foundStatus {
		unstructured.RemoveNestedField(objectU, "status")
		changed = true
	}
	if changed {
		object.SetUnstructuredContent(objectU)
	}
}

func cleanJobAnnotations(annotations map[string]any) bool {
	changed := false
	for key := range annotations {
		if key == "batch.kubernetes.io/job-tracking" {
			delete(annotations, key)
			changed = true
		}
	}
	return changed
}

func cleanJobLabels(labels map[string]any) bool {
	changed := false
	for key := range labels {
		if badJobKey(key) {
			delete(labels, key)
			changed = true
		}
	}
	return changed
}
