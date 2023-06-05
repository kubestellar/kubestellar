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
)

// ClusterProviderInfo represents a provider.
//
// +crd
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster,shortName=ecl
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterProviderInfo struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `spec` describes a provider.
	// +optional
	Spec ClusterProviderInfoSpec `json:"spec,omitempty"`

	// `status` describes the status of the cluster object.
	// +optional
	Status ClusterProviderInfoStatus `json:"status"`
}

// ClusterProviderInfoSpec describes a cluster.
type ClusterProviderInfoSpec struct {
	// ProviderType is the type of the provider of the cluster.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="cluster is immutable"
	ProviderType ClusterProviderType `json:"ProviderType"`

	// ProviderName is the name of the provider of the cluster.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="cluster is immutable"
	ProviderName string `json:"ProviderName"`

	// TODO: this should be stored as a secret!
	// Config is the provider config
	// +kubebuilder:format:uri
	Config string `json:"Config,omitempty"`
}

// ClusterProviderInfoStatus describes a cluster.
type ClusterProviderInfoStatus struct {
}

// ClusterProviderInfoList is the API type for a list of ClusterProviderInfo
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ClusterProviderInfoList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterProviderInfo `json:"items"`
}
