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

// EdgePlacementDecision exists in the center and is bound to a single EdgePlacement resource.
// The decision resource reflects the resolution of the bounded EdgePlacement's placement selectors,
// and explicitly reflects which resources should go to what destinations.
//
// +crd
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster,shortName=epl
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EdgePlacementDecision struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `spec` explicitly describes a desired binding between workloads and Locations.
	// It reflects the resolution of an EdgePlacement's placement selectors.
	// +optional
	Spec EdgePlacementDecisionSpec `json:"spec,omitempty"`

	// `status` describes the status of the process of binding workloads to Locations.
	// +optional
	Status EdgePlacementDecisionStatus `json:"status,omitempty"`
}

// EdgePlacementDecisionSpec holds a list of resources and a list of destinations which are the resolution
// of an EdgePlacement's `what` and `where`: what resources to propagate and to where.
// All resources present in this spec are propagated to all destinations present.
type EdgePlacementDecisionSpec struct {
	// `Workload` is a collection of namespaced and cluster-scoped resources to be propagated to destination clusters.
	Workload EdgePlacementDecisionSpecResources `json:"workload,omitempty"`

	// `destinations` is a list of cluster-identifiers that the resources should be propagated to.
	Destinations []Destination `json:"destinations,omitempty"`
}

// EdgePlacementDecisionSpecResources explicitly defines the resources to be down-synced.
// The ClusterScope list defines the cluster-scope resources, while NamespacedObjects packs individual resources
// identifiable by namespace & name.
type EdgePlacementDecisionSpecResources struct {
	// `clusterScope` holds a list of individual cluster-scoped objects
	// to downsync, organized by resource.
	// Remember that a "resource" is a kind/type/sort of objects,
	// not an individual object.
	// +optional
	ClusterScope []ClusterScopeDownsyncResource `json:"clusterScope,omitempty"`

	// `namespacedObjects` matches if and only if at least one member matches.
	// +optional
	NamespacedObjects []NamespaceScopeDownsyncObjects `json:"namespacedObjects,omitempty"`
}

// Destination wraps the identifiers required to uniquely identify a destination cluster.
type Destination struct {
	// Cluster is the logicalcluster.Name of the logical cluster
	Cluster string `json:"cluster"`

	LocationName string `json:"locationName"`

	Namespace string `json:"namespace,omitempty"`
}

type EdgePlacementDecisionStatus struct {
	// `specGeneration` identifies the generation of the spec that this
	// is the status for.
	// Zero means that no status has yet been written here.
	// +optional
	SpecGeneration int32 `json:"specGeneration,omitempty"`

	// `propagatedWorkloadsCount` is the number of destinations that received all the resources in the `spec.workload`.
	PropagatedWorkloadsCount int `json:"propagatedWorkloadsCount,omitempty"`
}

// EdgePlacementDecisionList is the API type for a list of EdgePlacementDecision
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EdgePlacementDecisionList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgePlacementDecision `json:"items"`
}
