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

type SpaceProviderType string

const (
	KindProviderType     SpaceProviderType = "kind"
	KubeflexProviderType SpaceProviderType = "kubeflex"
)

// SpaceProviderDesc represents a provider.
//
// +crd
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster,shortName=spd
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SpaceProviderDesc struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `spec` describes a provider.
	// +optional
	Spec SpaceProviderDescSpec `json:"spec,omitempty"`

	// `status` describes the status of the provider object.
	// +optional
	Status SpaceProviderDescStatus `json:"status"`
}

// SpaceProviderDescSpec describes a space provider.
// TODO: We are currently only listing the type and config.
// There will be additional fields in the future.
type SpaceProviderDescSpec struct {
	// ProviderType is the type of the provider of the space.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="space provider type is immutable"
	ProviderType SpaceProviderType `json:"ProviderType"`

	// TODO: this should be stored as a secret!
	// Config is the provider config
	// +kubebuilder
	Config string `json:"Config,omitempty"`

	// SpacePrefixForDiscovery contains the prefix used during space discovery.
	// +kubebuilder
	SpacePrefixForDiscovery string `json:"SpacePrefixForDiscovery,omitempty"`
}

// SpaceProviderDescPhaseType is the type of the current phase of the provider.
//
// +kubebuilder:validation:Enum=Initializing;Ready
type SpaceProviderDescPhaseType string

const (
	SpaceProviderDescPhaseInitializing SpaceProviderDescPhaseType = "Initializing"
	SpaceProviderDescPhaseReady        SpaceProviderDescPhaseType = "Ready"
)

// SpaceProviderDescStatus describes a provider status.
// TODO: in the future we may want to hold a list of space resources that
// use this provider.
type SpaceProviderDescStatus struct {
	// Phase of the provider (Initializing,Ready).
	// +kubebuilder:default=Initializing
	Phase SpaceProviderDescPhaseType `json:"Phase"`
}

// SpaceProviderDescList is the API type for a list of SpaceProviderDesc
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SpaceProviderDescList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpaceProviderDesc `json:"items"`
}
