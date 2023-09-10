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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Space represents a cluster.
//
// +crd
// +genclient
// +kubebuilder:resource:scope=Namespaced,shortName=spa
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Space struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `spec` describes a cluster.
	// +optional
	Spec SpaceSpec `json:"spec,omitempty"`

	// `status` describes the status of the cluster object.
	// +optional
	Status SpaceStatus `json:"status"`
}

// SpaceSpec describes a cluster.
type SpaceSpec struct {
	// SpaceProviderDescName is a reference to a SpaceProviderDesc resource
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="SpaceProviderDescName is immutable"
	SpaceProviderDescName string `json:"SpaceProviderDescName"`

	// Managed identifies whether a cluster is managed (true) or unmanaged (false).
	// Currently this is immutable.
	// A space can be created through the ClusterManager (managed) or
	// discovered/imported (unmanaged).
	// +kubebuilder:default=true
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Managed is immutable"
	Managed bool `json:"Managed"`
}

// SpacePhaseType is the type of the current phase of the cluster.
//
// +kubebuilder:validation:Enum=Initializing;NotReady;Ready
type SpacePhaseType string

const (
	SpacePhaseInitializing SpacePhaseType = "Initializing"
	SpacePhaseNotReady     SpacePhaseType = "NotReady"
	SpacePhaseReady        SpacePhaseType = "Ready"
)

// SpaceStatus represents information about the status of a cluster.
type SpaceStatus struct {
	// Phase of the space (Initializing,NotReady,Ready).
	// +kubebuilder
	Phase SpacePhaseType `json:"Phase,omitempty"`

	// Cluster config from the kube config file in string format.
	// +kubebuilder
	SpaceConfig string `json:"ClusterConfig,omitempty"`
}

// SpaceList is the API type for a list of Space
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SpaceList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Space `json:"items"`
}
