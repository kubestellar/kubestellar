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
	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func cleanJob(logger logr.Logger, object *unstructured.Unstructured) {
	objectU := object.UnstructuredContent()
	podLabels, found, err := unstructured.NestedMap(objectU, "spec", "template", "metadata", "labels")
	if err != nil || !found {
		logger.V(3).Error(nil, "Job lacks expected template labels", "namespace", object.GetNamespace(), "name", object.GetName(), "found", found, "err", err)
		return
	}
	delete(podLabels, "batch.kubernetes.io/controller-uid")
	delete(podLabels, "batch.kubernetes.io/job-name")
	delete(podLabels, "controller-uid")
	delete(podLabels, "job-name")
	err = unstructured.SetNestedMap(objectU, podLabels, "spec", "template", "metadata", "labels")
	if err != nil {
		logger.Error(err, "Inconceivable! Failed to update existing pod labels in Job", "namespace", object.GetNamespace(), "name", object.GetName())
	}
	object.SetUnstructuredContent(objectU)
}
