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

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SyncerConfig tells a syncer what to sync down and up.
// There is a 1:1:1:1 relation between:
// - edge cluster
// - mailbox workspace
// - syncer
// - syncer config.
//
// +crd
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster,shortName=escfg
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SyncerConfig struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec SyncerConfigSpec `json:"spec,omitempty"`

	// +optional
	Status SyncerConfigStatus `json:"status,omitempty"`
}

// SyncerConfigSpec is instructions to the syncer
type SyncerConfigSpec struct {
	// +optional
	NamespaceScope NamespaceScopeDownsyncs `json:"namespaceScope,omitempty"`

	// `clusterScope` holds a list of individual cluster-scoped objects
	// to downsync, organized by resource.
	// Remember that a "resource" is a kind/type/sort of objects,
	// not an individual object.
	// +optional
	ClusterScope []ClusterScopeDownsyncResource `json:"clusterScope,omitempty"`

	// `upsync` identifies objects to upsync.
	// An object matches `upsync` if and only if it matches at least one member of `upsync`.
	// The syncer identifies matching objects in the edge cluster.
	// The syncer reads and writese the matching objects using the
	// API version preferred in the edge cluster.
	// +optional
	Upsync []UpsyncSet `json:"upsync,omitempty"`
}

// NamespaceScopeDownsyncs describes what namespace-scoped objects
// to downsync.
// Note that it is factored into two orthogonal parts,
// one identifying namespaces and one identifying resources.
// An object is to be downsynced iff it matches both parts.
type NamespaceScopeDownsyncs struct {
	// `namespaces` is the names of the namespaces to downsync.
	// Empty list means to downsync no namespace contents.
	// Whether the particulars of the Namespace object itself
	// are to be downsynced are controlled by the `clusterScope`;
	// if not then downsync will ensure that the namespace exists
	// but take no further care to make it match upstream.
	// +optional
	Namespaces []string `json:"namespaces,omitempty"`

	// `resources` lists the namespace-scoped resources to downsync.
	// Empty list means none of them.
	// +optional
	Resources []NamespaceScopeDownsyncResource `json:"resources,omitempty"`
}

type NamespaceScopeDownsyncResource struct {
	// GroupResource holds the API group and resource name.
	metav1.GroupResource `json:",inline"`

	// `apiVeresion` holds just the version, not the group too.
	// This is the version to use both upstream and downstream.
	APIVersion string `json:"apiVersion"`
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

type SyncerConfigStatus struct {
	// A timestamp indicating when the syncer last reported status.
	// +optional
	LastSyncerHeartbeatTime *metav1.Time `json:"lastSyncerHeartbeatTime,omitempty"`
}

// SyncerConfigList is the API type for a list of SyncerConfig
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SyncerConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SyncerConfig `json:"items"`
}
