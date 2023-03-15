/*
Copyright 2023 The KCP Authors.

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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// AssociatorAnnotationKey is the key of an annotation used to point from a workload
// object to an Associator that says how associated objects in edge clusters are identified.
// For a namespaced object, the value of the annotation is the name of the Associator.
// For a non-namespaced object, the value of the annotation is the namespace/name of the Associator.
// When the annotation is absent, no associated objects are identified.
const AssociatorAnnotationKey = "edge.kcp.io/associator"

// Associator specifies how the associated objects of a primary object are identified.
// A primary object is one that is specified in the center and propagated to the edge.
// An associated object is one that first appears at the edge and then propagates to the center.
// An Associator has a collection of ways that associated objects are identified.
//
// +crd
// +genclient
// +kubebuilder:resource:scope=Namespace,shortName=assoc
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Associator struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AssociatorSpec `json:"spec"`
}

// AssociatorSpec lists the ways that associated objects are identified.
// Object X is an associated object if it satisfies any of the listed rules here.
type AssociatorSpec struct {
	Rules []AssociationRule `json:"rules"`
}

// AssociationRule identifies some associated objects.
// This rule associates object A with primary object P if and only if
// both (a) A matches an item in `resources` and
// (b) A refers to P as described by `reference`.
type AssociationRule struct {
	Resources []ResourceGroup `json:"resources"`
	Reference ReferenceRule   `json:"reference"`
}

// ResourceGroup identifies a set of resources in one particular API group
type ResourceGroup struct {
	APIGroup  string   `json:"apiGroup"`
	Resources []string `json:"resources"`
}

// ReferenceRule describes a way for an associated object A to refer to
// a primary object P.  There are four possible ways enumerated here,
// and two of them have additional details.
type ReferenceRule struct {
	Type ReferenceType `json:"type"`

	// `label` is a label key, and the corresponding value is the name of P
	Label string `json:"label,omitempty"`

	// `ownerReference` can supply an additional restriction
	OwnerReference *OwnerReferenceAssociation `json:"ownerReference,omitempty"`
}

// ReferenceType identifies one of the ways that an associated object can refer to a primary object
type ReferenceType string

const (
	// ReferenceTypeLabel means that A has a label referring to P
	ReferenceTypeLabel ReferenceType = "Label"

	// ReferenceTypeOwnerReference means that A has a `metav1.OwnerReference` that refers to P
	ReferenceTypeOwnerReference ReferenceType = "OwnerReference"

	// ReferenceTypeNamespace means that P is namespaced and A is the Namespace object for the namespace that contains P
	ReferenceTypeNamespace ReferenceType = "Namespace"

	// ReferenceTypeCRD means that A is the CustomResourceDefinition object that defines the kind of P
	ReferenceTypeCRD ReferenceType = "CRD"
)

type OwnerReferenceAssociation struct {
	// `controller`, if specified, is the value that the `controller` field of the OwnerReference must have.
	Controller *bool `json:"controller,omitempty"`
}

// AssociatorList is the API type for a list of Associator
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type AssociatorList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Associator `json:"items"`
}
