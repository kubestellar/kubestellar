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

// PlacementDecision is mapped 1:1 to a single Placement resource.
// The decision resource reflects the resolution of the Placement's selectors,
// and explicitly reflects which resources should go to what destinations.
//
// +crd
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster,shortName=pd
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PlacementDecision struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `spec` explicitly describes a desired binding between workloads and Locations.
	// It reflects the resolution of a Placement's selectors.
	Spec PlacementDecisionSpec `json:"spec,omitempty"`

	// `status` describes the status of the process of binding workloads to Locations.
	// +optional
	Status PlacementDecisionStatus `json:"status,omitempty"`
}

// PlacementDecisionSpec holds a list of resources and a list of destinations which are the resolution
// of a Placement's `what` and `where`: what resources to propagate and to where.
// All resources present in this spec are propagated to all destinations present.
type PlacementDecisionSpec struct {
	// `Workload` is a collection of namespaced-scoped resources and a collection of cluster-scoped resources to be propagated to destination clusters.
	Workload PlacementDecisionSpecResources `json:"workload,omitempty"`

	// `destinations` is a list of cluster-identifiers that the resources should be propagated to.
	Destinations []Destination `json:"destinations,omitempty"`
}

// PlacementDecisionSpecResources explicitly defines the resources to be down-synced.
// The ClusterScope list defines the cluster-scope resources, while NamespacedObjects packs individual resources
// identifiable by namespace & name.
type PlacementDecisionSpecResources struct {
	// `clusterScope` holds a list of individual cluster-scoped objects
	// to downsync, organized by resource.
	// Remember that a "resource" is a kind/type/sort of objects,
	// not an individual object.
	// +optional
	ClusterScope []ClusterScopeDownsyncResource `json:"clusterScope,omitempty"`

	// `NamespaceScope` matches if and only if at least one member matches.
	// +optional
	NamespaceScope []NamespaceScopeDownsyncObjects `json:"namespaceScope,omitempty"`
}

// NamespaceScopeDownsyncObjects matches some objects of one particular namespaced resource.
type NamespaceScopeDownsyncObjects struct {
	// GroupResource holds the API group and resource name.
	metav1.GroupResource `json:",inline"`

	// `apiVeresion` holds just the version, not the group too.
	// This is the version to use both upstream and downstream.
	APIVersion string `json:"apiVersion"`

	// `objectsByNamespace` matches by namespace and name.
	// An object matches the list if and only if the object matches at least one member of the list.
	// Thus, no object matches the empty list.
	// +optional
	ObjectsByNamespace []NamespaceAndNames `json:"objectsByNamespace,omitempty"`
}

// NamespaceAndNames identifies some objects of an implied resource that is namespaced.
// The objects are all in the same namespace.
type NamespaceAndNames struct {
	// `namespace` identifies the namespace
	Namespace string `json:"namespace"`

	// `names` holds the names of the objects that match.
	// Empty list means none of them.
	// +optional
	Names []string `json:"names,omitempty"`
}

type ClusterScopeDownsyncResource struct {
	// GroupResource holds the API group and resource name.
	metav1.GroupResource `json:",inline"`

	// `apiVeresion` holds just the version, not the group too.
	// This is the version to use both upstream and downstream.
	APIVersion string `json:"apiVersion"`

	// `objects` holds the names of the objects of this kind to downsync.
	// Empty list means none of them.
	// +optional
	Objects []string `json:"objects,omitempty"`
}

// Destination wraps the identifiers required to uniquely identify a destination cluster.
type Destination struct {
	Namespace string `json:"namespace,omitempty"`
}

type PlacementDecisionStatus struct {
	// `specGeneration` identifies the generation of the spec that this is the status for.
	// Zero means that no status has yet been written here.
	// +optional
	SpecGeneration int32 `json:"specGeneration,omitempty"`

	// `propagatedWorkloadsCount` is the number of destinations that received all the resources in the `spec.workload`.
	PropagatedWorkloadsCount int `json:"propagatedWorkloadsCount,omitempty"`
}

// PlacementDecisionList is the API type for a list of PlacementDecision
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type PlacementDecisionList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PlacementDecision `json:"items"`
}
