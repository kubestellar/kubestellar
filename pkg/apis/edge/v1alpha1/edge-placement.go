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

// EdgePlacement binds a collection of (a) Namespaces and non-namespaced objects
// to (b) a collection of Locations.
// An EdgePlacement object appears in the center and directs all the syncable
// objects in the selected Namespaces and all the selected non-namespaced objects
// to propagate to _all_ of the selected Locations
// (one SyncTarget in each such Location).
// This is not entirely unrelated to a TMC Placement, which directs the selected
// Namespaces to propagate to _one_ of the selected Locations.
//
// Overlap between EdgePlacements is allowed:
// two different EdgePlacement objects may select intersecting Location sets
// and/or intersecting Namespace sets.
// This is not problematic because:
//   - propagation _into_ a destination is additive;
//   - propagation _from_ a source is additive;
//   - two directives to propagate the same object to the same destination are
//     simply redundant (they, by design, can not conflict).
//
// +crd
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EdgePlacement struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `spec` describes a desired binding between workload and Locations.
	// Unlike a TMC Placement, there is an inherent multiplicity and dynamicity
	// in the set of Locations that will be synced to and this field
	// never shifts into immutability.
	// +optional
	Spec EdgePlacementSpec `json:"spec,omitempty"`

	// `status` describes the status of the process of binding
	// workload to Locations.
	// +optional
	Status EdgePlacementStatus `json:"status,omitempty"`
}

// EdgePlacementSpec holds a desired binding between (a) Namespaces and non-namespaced objects
// and (b) Locations.
type EdgePlacementSpec struct {
	// `locationWorkspaceSelector` identifies the workspaces in which to look for Location
	// objects, in terms of labels on the Workspace objects.
	LocationWorkspaceSelector metav1.LabelSelector `json:"locationWorkspaceSelector,omitempty"`

	// `locationSelectors` identifies the relevant Location objects in terms of their labels.
	// A Location is relevant if and only if it passes any of the LabelSelectors in this field.
	LocationSelectors []metav1.LabelSelector `json:"locationSelectors,omitempty"`

	// `namespaceSelector` identifies the relevant Namespace objects in terms of their labels.
	NamespaceSelector metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// `nonNamespacedObjects` defines the non-namespaced objects to bind with the selected Locations.
	// +optional
	NonNamespacedObjects []NonNamespacedObjectReferenceSet `json:"nonNamespacedObjects,omitempty"`
}

// NonNamespacedObjectReferenceSet specifies a set of non-namespaced objects
// from one particular API group.
// An object is in this set if:
// - its API group is the one listed;
// - its resource (lowercase plural form of object type) is one of those listed; and
// - EITHER its name is listed OR its labels match one of the label selectors.
type NonNamespacedObjectReferenceSet struct {
	// `apiGroup` is the API group of the referenced object, empty string for the core API group.
	APIGroup string `json:"apiGroup,omitempty"`

	// `resources` is a list of lowercase plural names for the sorts of objects to match.
	// An entry of `"*"` means that all match.
	// Empty list means nothing matches.
	Resources []string `json:"resources"`

	// `resourceNames` is a list of objects that match by name.
	// An entry of `"*"` means that all match.
	// Empty list means nothing matches.
	ResourceNames []string `json:"resourceNames,omitempty"`

	// `labelSelectors` allows matching objects by a rule rather than listing individuals.
	LabelSelectors []metav1.LabelSelector `json:"labelSelectors,omitempty"`
}

type EdgePlacementStatus struct {
	// `specGeneration` identifies the generation of the spec that this
	// is the status for.
	// Zero means that no status has yet been written here.
	// +optional
	SpecGeneration int32 `json:"specGeneration,omitempty"`

	// `matchingLocationCount` is the number of Locations that satisfy the spec's
	// `locationSelectors`.
	MatchingLocationCount int32 `json:"matchingLocationCount"`
}

// EdgePlacementList is the API type for a list of EdgePlacement
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EdgePlacementList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgePlacement `json:"items"`
}
