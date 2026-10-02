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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// cleanPersistentVolume removes plane-local fields from a PersistentVolume
// before it is wrapped into transport objects.
//
// A statically-bound PV carries a spec.claimRef with only name+namespace in
// the seed YAML. The PV binder then stamps cluster-local identity
// (uid, resourceVersion) onto claimRef. That identity must never propagate:
// the edge binder attaches the *spoke* PVC uid itself, and enforcing hub-side
// identity on the edge breaks static binding (PV stays Available, PVC goes
// Lost; see kubestellar/kubestellar#3977). This matters in particular for
// host-type WDS, where hub controllers may stamp the live WDS object.
// Static binding intent (claimRef name/namespace, volumeName, storageClassName,
// capacity, accessModes, persistentVolumeReclaimPolicy, nodeAffinity, …) is kept.
func cleanPersistentVolume(object *unstructured.Unstructured) {
	objectU := object.UnstructuredContent()
	changed := false
	if _, found, _ := unstructured.NestedFieldNoCopy(objectU, "spec", "claimRef", "uid"); found {
		unstructured.RemoveNestedField(objectU, "spec", "claimRef", "uid")
		changed = true
	}
	if _, found, _ := unstructured.NestedFieldNoCopy(objectU, "spec", "claimRef", "resourceVersion"); found {
		unstructured.RemoveNestedField(objectU, "spec", "claimRef", "resourceVersion")
		changed = true
	}
	if _, foundStatus, _ := unstructured.NestedFieldNoCopy(objectU, "status"); foundStatus {
		unstructured.RemoveNestedField(objectU, "status")
		changed = true
	}
	if changed {
		object.SetUnstructuredContent(objectU)
	}
}
