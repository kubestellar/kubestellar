/*
Copyright 2023 The KubeStellar Authors.

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

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DownsyncWorkloadPartSlice reports the identities of some of the
// objects that match the "what predicate" of an EdgePlacement and
// the modalities of their downsync.
// Collectively all the DownsyncWorkloadPartSlice objects that share
// the same value of the "edge.kubestellar.io/source-placement" label
// report on the identities of all the objects that match the "what predicate"
// of the EdgePlacement named by that label.
//
// +crd
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster,shortName=dwps,path=downsyncworkloadpartslices
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DownsyncWorkloadPartSlice struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `parts` is the object identities being reported and their associated modalities.
	// +patchStrategy=merge
	// +patchMergeKey=group,version,resource,namespace,name
	Parts []DownsyncWorkloadPart `json:"parts" patchStrategy:"merge" patchMergeKey:"group,version,resource,namespace,name"`
}

// DownsyncWorkloadPart identifies one workload object and its downsync modalities.
type DownsyncWorkloadPart struct {
	metav1.GroupVersionResource `json:",inline"`
	Namespace                   string `json:"namespace"`
	Name                        string `json:"name"`

	// `returnSingletonState` indicates that the user intends this object to go to exactly 1 WEC
	// and the reported state from there should be returned all the way to the WDS.
	// +optional
	ReturnSingletonState bool `json:"returnSingletonState,omitempty"`

	// `createOnly` indicates that the WDS supplies only the initial value for the desired state.
	// +optional
	CreateOnly bool `json:"createOnly,omitempty"`
}

// DownsyncWorkloadPartSliceList is the API type for a list of DownsyncWorkloadPartSlice
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DownsyncWorkloadPartSliceList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DownsyncWorkloadPartSlice `json:"items"`
}
