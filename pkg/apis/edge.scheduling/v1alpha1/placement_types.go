/*
Copyright 2022.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PlacementSpec defines the desired state of Placement
type PlacementSpec struct {
	// `locationSelectors` identifies the relevant Location objects in terms of their labels.
	// A Location is relevant if and only if it passes any of the LabelSelectors in this field.
	LocationSelectors []metav1.LabelSelector `json:"locationSelectors,omitempty"`

	// `namespaceSelector` identifies the relevant Namespace objects in terms of their labels.
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// locationWorkspace is an absolute reference to a workspace for the location.
	// +optional
	// +kubebuilder:validation:Pattern:="^root(:[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$"
	LocationWorkspace string `json:"locationWorkspace,omitempty"`
}

// PlacementStatus defines the observed state of Placement
type PlacementStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Placement is the Schema for the placements API
type Placement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlacementSpec   `json:"spec,omitempty"`
	Status PlacementStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PlacementList contains a list of Placement
type PlacementList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Placement `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Placement{}, &PlacementList{})
}
