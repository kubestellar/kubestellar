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
	corev1 "k8s.io/api/core/v1"
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

// SpaceType identifies the type of the space (managed, unmanaged, imported)
// +kubebuilder:validation:Enum=managed;unmanaged;imported
type SpaceType string

const (
	SpaceTypeManaged   SpaceType = "managed"
	SpaceTypeUnmanaged SpaceType = "unmanaged"
	SpaceTypeImported  SpaceType = "imported"
)

// SpaceSpec describes a cluster.
type SpaceSpec struct {
	// SpaceProviderDescName is a reference to a SpaceProviderDesc resource
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="SpaceProviderDescName is immutable"
	// +optional
	SpaceProviderDescName string `json:"SpaceProviderDescName"`

	// Type identifies the space type.
	// A space can be created through the ClusterManager (managed), discovered (unmanaged), or imported.
	// +kubebuilder:default=managed
	Type SpaceType `json:"Type"`

	// Access indicate whether the space is going to be accessed from within the cluster the space resides on
	// or externally
	// +kubebuilder:default=Both
	// +optional
	AccessScope AccessScopeType `json:"accessscopetype,omitempty"`
}

// +kubebuilder:validation:Enum=InCluster;External;Both
type AccessScopeType string

const (
	AccessScopeInCluster AccessScopeType = "InCluster"
	AccessScopeExternal  AccessScopeType = "External"
	AccessScopeBoth      AccessScopeType = "Both"
)

// SpacePhaseType is the type of the current phase of the cluster.
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

	InClusterSecretRef *corev1.SecretReference `json:"inClusterSecretRef,omitempty"`

	ExternalSecretRef *corev1.SecretReference `json:"externalSecretRef,omitempty"`
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
