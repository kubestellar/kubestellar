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

// BindingPolicySpec defines the desired state of BindingPolicy
type BindingPolicySpec struct {
	// `clusterSelectors` identifies the relevant Cluster objects in terms of their labels.
	// A Cluster is relevant if and only if it passes any of the LabelSelectors in this field.
	ClusterSelectors []metav1.LabelSelector `json:"clusterSelectors,omitempty"`

	// We agreed not to expose NumberOfClusters in release 0.20, to avoid confusions.
	// We may or may not support it in later releases per future discussions.
	// NumberOfClusters represents the desired number of ManagedClusters to be selected which meet the
	// BindingPolicy's requirements.
	// 1) If not specified, all Clusters which meet the BindingPolicy's requirements will be selected;
	// 2) Otherwise if the number of Clusters meet the BindingPolicy's requirements is larger than
	//    NumberOfClusters, a random subset with desired number of ManagedClusters will be selected;
	// 3) If the number of Clusters meet the BindingPolicy's requirements is equal to NumberOfClusters,
	//    all of them will be selected;
	// 4) If the number of Clusters meet the BindingPolicy's requirements is less than NumberOfClusters,
	//    all of them will be selected, and the status of condition `BindingPolicyConditionSatisfied` will be
	//    set to false;
	// +optional
	// NumberOfClusters *int32 `json:"numberOfClusters,omitempty"`

	// `downsync` selects the objects to bind with the selected Locations for downsync.
	// An object is selected if it matches at least one member of this list.
	// +optional
	Downsync []DownsyncObjectTest `json:"downsync,omitempty"`

	// WantSingletonReportedState indicates that (a) the number of selected locations is intended
	// to be 1 and (b) the reported state of each downsynced object should be returned back to
	// the object in this space.
	// When multiple BindingPolicy objects match the same workload object,
	// the OR of these booleans rules.
	// +optional
	WantSingletonReportedState bool `json:"wantSingletonReportedState,omitempty"`
}

// BindingPolicyStatus defines the observed state of BindingPolicy
type BindingPolicyStatus struct {
	Conditions         []BindingPolicyCondition `json:"conditions"`
	ObservedGeneration int64                    `json:"observedGeneration"`
}

// BindingPolicy defines in which ways the workload objects ('what') and the destinations ('where') are bound together.
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,shortName={bp}
type BindingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BindingPolicySpec   `json:"spec,omitempty"`
	Status BindingPolicyStatus `json:"status,omitempty"`
}

const (
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
	ExecutingCountKey string = "kubestellar.io/executing-count"

	ValidationErrorKeyPrefix string = "validation-error.kubestellar.io/"

	// BindingPolicyConditionSatisfied means BindingPolicy requirements are satisfied.
	// A BindingPolicy is not satisfied only if the set of selected clusters is empty
	BindingPolicyConditionSatisfied string = "BindingPolicySatisfied"

	// BindingPolicyConditionMisconfigured means BindingPolicy configuration is incorrect.
	BindingPolicyConditionMisconfigured string = "BindingPolicyMisconfigured"
)

// DownsyncObjectTest is a set of criteria that characterize matching objects.
// An object matches if:
// - the `apiGroup` criterion is satisfied;
// - the `resources` criterion is satisfied;
// - the `namespaces` criterion is satisfied;
// - the `namespaceSelectors` criterion is satisfied;
// - the `objectNames` criterion is satisfied; and
// - the `objectSelectors` criterion is satisfied.
// At least one of the fields must make some discrimination;
// it is not valid for every field to match all objects.
// Validation might not be fully checked by apiservers until the Kubernetes dependency is release 1.25;
// in the meantime validation error messages will appear
// in annotations whose key is `validation-error.kubestellar.io/{number}`.
type DownsyncObjectTest struct {
	// `apiGroup` is the API group of the referenced object, empty string for the core API group.
	// `nil` matches every API group.
	// +optional
	APIGroup *string `json:"apiGroup"`

	// `resources` is a list of lowercase plural names for the sorts of objects to match.
	// An entry of `"*"` means that all match.
	// If this list contains `"*"` then it should contain nothing else.
	// Empty list is a special case, it matches every object.
	// +optional
	Resources []string `json:"resources,omitempty"`

	// `namespaces` is a list of acceptable names for the object's namespace.
	// An entry of `"*"` means that any namespace is acceptable;
	// this is the only way to match a cluster-scoped object.
	// If this list contains `"*"` then it should contain nothing else.
	// Empty list is a special case, it matches every object.
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// `namespaceSelectors` a list of label selectors.
	// For a namespaced object, at least one of these label selectors has to match
	// the labels of the Namespace object that defines the namespace of the object that this DownsyncObjectTest is testing.
	// For a cluster-scoped object, at least one of these label selectors must be `{}`.
	// Empty list is a special case, it matches every object.
	// +optional
	NamespaceSelectors []metav1.LabelSelector `json:"namespaceSelectors,omitempty"`

	// `objectSelectors` is a list of label selectors.
	// At least one of them must match the labels of the object being tested.
	// Empty list is a special case, it matches every object.
	// +optional
	ObjectSelectors []metav1.LabelSelector `json:"objectSelectors,omitempty"`

	// `objectNames` is a list of object names that match.
	// An entry of `"*"` means that all match.
	// If this list contains `"*"` then it should contain nothing else.
	// Empty list is a special case, it matches every object.
	// +optional
	ObjectNames []string `json:"objectNames,omitempty"`
}

// +kubebuilder:object:root=true

// BindingPolicyList contains a list of BindingPolicies
type BindingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BindingPolicy `json:"items"`
}

// Binding is mapped 1:1 to a single BindingPolicy object.
// Binding reflects the resolution of the BindingPolicy's selectors,
// and explicitly reflects which objects should go to what destinations.
//
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName={bdg}
type Binding struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `spec` explicitly describes a desired binding between workloads and Locations.
	// It reflects the resolution of a BindingPolicy's selectors.
	Spec BindingSpec `json:"spec,omitempty"`
}

// BindingSpec holds a list of objects and a list of destinations which are the resolution
// of a BindingPolicy's `what` and `where`: what objects to propagate and to where.
// All objects present in this spec are propagated to all destinations present.
type BindingSpec struct {
	// `Workload` is a collection of namespaced-scoped objects and a collection of cluster-scoped objects to be propagated to destination clusters.
	Workload DownsyncObjectReferences `json:"workload,omitempty"`

	// `destinations` is a list of cluster-identifiers that the objects should be propagated to.
	Destinations []Destination `json:"destinations,omitempty"`
}

// DownsyncObjectReferences explicitly defines the objects to be down-synced.
// The ClusterScope list defines the cluster-scope objects, NamespaceScope list
// defines the namespace-scope objects and WorkloadGeneration represents the
// generation of the objects in the ClusterScope and NamespaceScope lists.
// Upon a change in any of workload objects that should be distributed
// (e.g., spec, annotations or labels) the workload generation field should be incremented.
type DownsyncObjectReferences struct {
	// `clusterScope` holds a list of individual cluster-scoped objects
	// to downsync, organized by resource.
	// Remember that a "resource" is a kind/type/sort of objects,
	// not an individual object.
	// +optional
	ClusterScope []ClusterScopeDownsyncObjects `json:"clusterScope,omitempty"`

	// `NamespaceScope` matches if and only if at least one member matches.
	// +optional
	NamespaceScope []NamespaceScopeDownsyncObjects `json:"namespaceScope,omitempty"`

	// `WorkloadGeneration` is a sequence number representing a specific generation of
	// the workload objects to be downsynced.
	// For example, if ClusterScope and NamespaceScope lists haven't changed but at least
	// one object has changed, this field should be incremented.
	// Upon a change in any of workload objects that should be distributed
	// (e.g., spec, annotations or labels) the workload generation field should be incremented.
	// `WorkloadGeneration` field is monotonically increasing.
	// +optional
	WorkloadGeneration int64 `json:"workloadGeneration,omitempty"`
}

// NamespaceScopeDownsyncObjects matches some objects of one particular namespaced object.
type NamespaceScopeDownsyncObjects struct {
	// GroupVersionResource holds the API group, API version and resource name.
	metav1.GroupVersionResource `json:",inline"`

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

type ClusterScopeDownsyncObjects struct {
	// GroupVersionResource holds the API group, API version and resource name.
	metav1.GroupVersionResource `json:",inline"`

	// `objectNames` holds the names of the objects of this kind to downsync.
	// Empty list means none of them.
	// +optional
	ObjectNames []string `json:"objectNames,omitempty"`
}

// Destination wraps the identifiers required to uniquely identify a destination cluster.
type Destination struct {
	ClusterId string `json:"clusterId,omitempty"`
}

// BindingList is the API type for a list of Binding
//
// +kubebuilder:object:root=true
type BindingList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Binding `json:"items"`
}
