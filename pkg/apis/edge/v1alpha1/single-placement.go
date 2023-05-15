/*
Copyright 2022 The KCP Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachtypes "k8s.io/apimachinery/pkg/types"
)

// SourcePlacementLabelKey is the key of the label used in a SinglePlacementSlice
// to reference the EdgePlacement that it is part of the response to.
// We use a label rather than a field because field selectors do not work
// on a resource defined by a CRD.
const SourcePlacementLabelKey string = "edge.kcp.io/source-placement"

// SinglePlacementSlice is the interface between "scheduling" and syncing.
// For a given EdgePlacement object, the scheduler figures out which Locations
// match that EdgePlacement and selects the SyncTarget to use within that Location, then
// puts these results into some SinglePlacementSlice API objects.
// We use potentially multiple API objects so that no one of them has to get
// very big.
//
// +crd
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster,shortName=sps,path=singleplacementslices
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SinglePlacementSlice struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `destinations` holds some of the results of scheduling.
	Destinations []SinglePlacement `json:"destinations"`
}

// SinglePlacement describes one Location that matches the relevant EdgePlacement.
type SinglePlacement struct {
	// Cluster is the logicacluster.Name of the logical cluster that contains
	// both the Location and the SyncTarget.
	Cluster string `json:"cluster"`

	LocationName string `json:"locationName"`

	// `syncTargetName` identifies the relevant SyncTarget at the Location
	SyncTargetName string `json:"syncTargetName"`

	SyncTargetUID apimachtypes.UID `json:"syncTargetUID"`
}

// SinglePlacementSliceList is the API type for a list of SinglePlacementSlice
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SinglePlacementSliceList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SinglePlacementSlice `json:"items"`
}
