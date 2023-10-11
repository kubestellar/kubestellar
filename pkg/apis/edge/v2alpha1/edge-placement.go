/*
Copyright 2022 The KubeStellar Authors.

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

// EdgePlacement exists in the center and binds (a) a collection of
// Locations with (b) both (b1) objects in the center to downsync,
// and (b2) a way of identifying objects
// (in edge clusters) to upsync (propagate from edge toward center).
// For an object subject to downsync, its state is divided into
// the following four categories.
//  1. Desired state, which flows from center to edge;
//  2. Reported state, which flows from edge to center;
//  3. System-managed state, determined by Kubernetes and beyond the
//     control of KubeStellar;
//  4. Immobile state, which also does not propagate but is not determined
//     by Kubernetes.
//
// Both downsyncing and upsyncing are with all of the Locations.  This
// is not entirely unrelated to a TMC Placement, which directs the
// selected Namespaces to propagate to _one_ of the selected
// Locations.
//
// The objects to downsync are those in selected namespaces plus
// selected non-namespaced objects.
//
// For upsync, the matching objects originate in the edge clusters and
// propagate to the corresponding mailbox workspaces while summaries
// of them go to the workload management workspace (as prescribed by
// the summarization API).
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
// +kubebuilder:resource:scope=Cluster,shortName=epl
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

// EdgePlacementSpec holds a desired binding between (a) a set of Locations and
// (b) a way of identifying objects to downsync and a way of identifying objects to upsync.
type EdgePlacementSpec struct {
	// `locationSelectors` identifies the relevant Location objects in terms of their labels.
	// A Location is relevant if and only if it passes any of the LabelSelectors in this field.
	LocationSelectors []metav1.LabelSelector `json:"locationSelectors,omitempty"`

	// `downsync` selects the objects to bind with the selected Locations for downsync,
	// and helps define which parts are immobile.
	// An object is selected if it matches at least one member of this list.
	Downsync []DownsyncObjectSelector `json:"downsync,omitempty"`

	// WantSingletonReportedState indicates that (a) the number of selected locations is intended
	// to be 1 and (b) the reported state of each downsynced object should be returned back to
	// the object in this space.
	// When multiple EdgePlacement objects match the same workload object,
	// the OR of these booleans rules.
	WantSingletonReportedState bool `json:"wantSingletonReportedState,omitempty"`

	// `upsync` identifies objects to upsync.
	// An object matches `upsync` if and only if it matches at least one member of `upsync`.
	// +optional
	Upsync []UpsyncSet `json:"upsync,omitempty"`
}

// ExecutingCountKey is the name (AKA key) of an annotation on a workload object.
// This annotation is written by the KubeStellar implementation to report on
// the number of executing copies of that object.
// This annotation is maintained while that number is intended to be 1
// (see the `WantSingletonReportedState` field above).
// The value of this annotation is a string representing the number of
// executing copies.  While this annotation is present with the value "1",
// the reported state is being returned into this workload object (the design
// of an API object typically assumes that it is taking effect in just one cluster).
// For reported state from a general number of executing copies, see the
// mailboxwatch library and the aspiration for summarization.
const ExecutingCountKey string = "kubestellar.io/executing-count"

// DownsyncObjectSelector is a set of criteria that characterize matching objects,
// and information about which part(s) of the matching objects are immobile.
// An object matches if:
// - the `apiGroup` criterion is satisfied;
// - the `resources` criterion is satisfied;
// - the `namespaceSelectors` criterion is satisfied; and
// - EITHER the `objectNames` criterion or the `labelSelectors` criterion matches.
type DownsyncObjectSelector struct {
	// `apiGroup` is the API group of the referenced object, empty string for the core API group.
	APIGroup string `json:"apiGroup,omitempty"`

	// `resources` is a list of lowercase plural names for the sorts of objects to match.
	// An entry of `"*"` means that all match.
	// Empty list means nothing matches.
	Resources []string `json:"resources"`

	// `namespaceSelectors` tests the labels on the object's namespace.
	// A non-namespaced object matches if and only if this list is empty.
	// A namespaced object matches if and only if the namespace's labels match
	// at least one member of this list.
	NamespaceSelectors []metav1.LabelSelector `json:"namespaceSelectors"`

	// `objectNames` is a list of object names that match.
	// An entry of `"*"` means that all match.
	// Empty list means nothing matches.
	ObjectNames []string `json:"objectNames,omitempty"`

	// `labelSelectors` allows matching objects by a rule rather than listing individuals.
	// An object maches if and only if its labels match at least one member of this list.
	LabelSelectors []metav1.LabelSelector `json:"labelSelectors,omitempty"`

	// `immobileParts` identifies parts of the matched objects that are immobile.
	// That is, they do not propagate to or from WECs.
	// A part of an object is immobile iff it is listed in any of the DownsyncObjectSelectors
	// that match that object.
	ImmobileParts []LiteralPath `json:"immobileParts,omitempty"`
}

// LiteralPath identifies a part of a JSON data structure, by nested indexing.
// Indexing object `O` with path `[fieldname, index2, ... indexN]` means
// indexing `O.fieldname` with path `[index2, ... indexN]`.
// Indexing array `A` with path `[index1, index2, ... indexN]` means
// indexing `A[index1]` with path `[index2, ... indexN]`.
// Indexing value `V` with the empty path means `V`.
type LiteralPath []string

// UpsyncSet specifies a set of objects,
// which may be namespaced or cluster-scoped,
// from one particular API group.
// An object is in this set if:
// - its API group is the one listed;
// - its resource (lowercase plural form of object type) is one of those listed;
// - EITHER the resource is cluster-scoped OR the object's namespace matches `namespaces`; and
// - the object's name matches `names`.
type UpsyncSet struct {
	// `apiGroup` is the API group of the referenced object, empty string for the core API group.
	APIGroup string `json:"apiGroup,omitempty"`

	// `resources` is a list of lowercase plural names for the sorts of objects to match.
	// An entry of `"*"` means that all match.
	// Empty list means nothing matches.
	Resources []string `json:"resources"`

	// `namespaces` is a list of acceptable namespaces.
	// An entry of `"*"` means that all match.
	// Empty list means nothing matches (you probably do not want this
	// for namespaced resources).
	Namespaces []string `json:"namespaces,omitempty"`

	// `Names` is a list of objects that match by name.
	// An entry of `"*"` means that all match.
	// Empty list means nothing matches (you probably never want an empty list).
	Names []string `json:"names,omitempty"`
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
