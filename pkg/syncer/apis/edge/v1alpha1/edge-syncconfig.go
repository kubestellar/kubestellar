package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +crd
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:resource:scope=Cluster,shortName=esc
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EdgeSyncConfig struct {
	/*
	 */
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EdgeSyncConfigSpec   `json:"spec,omitempty"`
	Status EdgeSyncConfigStatus `json:"status,omitempty"`
}

// EdgeSyncConfigSpec defines the desired state of EdgeSyncConfig
type EdgeSyncConfigSpec struct {
	// Downsncyed resource list
	DownSyncedResources []EdgeSyncConfigResource `json:"downSyncedResources,omitempty"`

	// Upsncyed resource list
	UpSyncedResources []EdgeSyncConfigResource `json:"upSyncedResources,omitempty"`

	// Conversions
	Conversions []EdgeSynConversion `json:"conversions,omitempty"`
}

// Resource specifies down/up synced resource with exact GVK and name (and namespace if not cluster scoped resource)
type EdgeSyncConfigResource struct {
	// Kind of down/up synced resource
	Kind string `json:"kind,omitempty"`

	// Group of down/up synced resource
	Group string `json:"group,omitempty"`

	// Version of down/up synced resource
	Version string `json:"version,omitempty"`

	// Name of down/up synced resource
	Name string `json:"name,omitempty"`

	// Namespace of down/up synced resource if it's not cluster scoped resource
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// Resource to be renatured
type EdgeSynConversion struct {
	// Resource representation in upstream
	Upstream EdgeSyncConfigResource `json:"upstream,omitempty"`
	// Resource representation in downstream
	Downstream EdgeSyncConfigResource `json:"downstream,omitempty"`
}

// EdgeSyncConfigStatus defines the observed state of EdgeSyncConfig
type EdgeSyncConfigStatus struct {
	// A timestamp indicating when the syncer last reported status.
	// +optional
	LastSyncerHeartbeatTime *metav1.Time `json:"lastSyncerHeartbeatTime,omitempty"`
}

// EdgePlacementList is the API type for a list of EdgePlacement
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type EdgeSyncConfigList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EdgeSyncConfig `json:"items"`
}
