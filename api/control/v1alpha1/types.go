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

	// `downsync` selects the objects to bind with the selected WECs for downsync,
	// and describes how to combine the status returned from those WECs for each of the
	// selected objects.
	// An object is selected if it matches at least one member of this list.
	// All of the associated StatusReturns are applied to the object; they should
	// have disjoint combiner names.
	// +optional
	Downsync []DownsyncObjectTestAndStatusReturn `json:"downsync,omitempty"`

	// WantSingletonReportedState means that for objects that are distributed --- taking
	// all BindingPolicies into account --- to exactly one WEC, the object's reported state
	// from the WEC should be written to the object in its WDS.
	// WantSingletonReportedState connotes an expectation that indeed the object will
	// propagate to exactly one WEC, but there is no guaranteed reaction when this
	// expetation is not met.
	// +optional
	WantSingletonReportedState bool `json:"wantSingletonReportedState,omitempty"`
}

// BindingPolicyStatus defines the observed state of BindingPolicy
type BindingPolicyStatus struct {
	Conditions         []BindingPolicyCondition `json:"conditions"`
	ObservedGeneration int64                    `json:"observedGeneration"`
	Errors             []string                 `json:"errors,omitempty"`
	Warnings           []string                 `json:"warnings,omitempty"`
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

// DownsyncObjectTestAndStatusReturn identifies some objects (by a predicate)
// and asks for some combined status to be returned from those objects.
type DownsyncObjectTestAndStatusReturn struct {
	DownsyncObjectTest `json:",inline"`
	StatusReturn       `json:",inline"`
}

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

// BindingSpec holds a list of object references with their associated resource versions,
// and a list of destinations which are the resolution of a BindingPolicy's `what` and `where`:
// what objects to propagate and to where.
// All objects referenced in this spec are propagated to all destinations present.
type BindingSpec struct {
	// `workload` is a collection of namespaced and cluster scoped object references and their associated
	// resource versions, to be propagated to destination clusters.
	Workload DownsyncObjectReferences `json:"workload,omitempty"`

	// `destinations` is a list of cluster-identifiers that the objects should be propagated to.
	Destinations []Destination `json:"destinations,omitempty"`
}

// DownsyncObjectReferences defines the objects to be down-synced, grouping them by scope.
// It specifies a set of object references with their associated resource versions, to be downsynced.
// This effectively acts as a map from object reference to ResourceVersion.
type DownsyncObjectReferences struct {
	// `clusterScope` holds a list of cluster-scoped object references with their associated
	// resource versions to downsync.
	ClusterScope []ClusterScopeDownsyncObject `json:"clusterScope,omitempty"`

	// `namespaceScope` holds a list of namespace-scoped object references with their associated
	// resource versions to downsync.
	NamespaceScope []NamespaceScopeDownsyncObject `json:"namespaceScope,omitempty"`
}

// NamespaceScopeDownsyncObject represents a specific namespace-scoped object to downsync,
// identified by its GroupVersionResource, namespace, and name. The ResourceVersion specifies
// the exact version of the object to downsync.
type NamespaceScopeDownsyncObject struct {
	metav1.GroupVersionResource `json:",inline"`
	// `namespace` of the object to downsync.
	Namespace string `json:"namespace"`
	// `name` of the object to downsync.
	Name string `json:"name"`
	// `resourceVersion` is the version of the resource to downsync.
	ResourceVersion string `json:"resourceVersion"`
}

// ClusterScopeDownsyncObject represents a specific cluster-scoped object to downsync,
// identified by its GroupVersionResource and name. The ResourceVersion specifies the
// exact version of the object to downsync.
type ClusterScopeDownsyncObject struct {
	metav1.GroupVersionResource `json:",inline"`
	// `name` of the object to downsync.
	Name string `json:"name"`
	// `resourceVersion` is the version of the resource to downsync.
	ResourceVersion string `json:"resourceVersion"`
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

// StatusReturn says how to combine status from the WECs.
type StatusReturn struct {
	// `stalenessThresholdSecs` is the threshold used to derive
	// the `stale` bit in PropagationMeta values.
	// Default value is 120.
	StalenessThresholdSecs int32 `json:"stalenessThresholdSecs"`

	// `infoAgeQuantumSeconds` is defines the rounding granularity
	// for PropagationMeta.InfoAgeSeconds.
	// Default value is `stalenessThresholdSecs`.
	InfoAgeQuantumSeconds int64 `json:"infoAgeQuantumSeconds,omitempty"`

	// `combiners` defines a named collection of ways to combine status from the WECs
	Combiners []NamedStatusCombiner `json:"aggregators,omitempty"`
}

// NamedStatusCombiner defines one way to collect status about a given workload object from
// the set of WECs that it propagates to.
// This is modeled after an SQL SELECT statement that does aggregation.
type NamedStatusCombiner struct {
	Name string `json:"name"`

	// `filter`, if given, is applied first.
	// It must evaluate to a boolean or null (which is treated as false).
	// This is like the WHERE clause in an SQL SELECT statement.
	// +optional
	Filter *Expression `json:"filter,omitempty"`

	// `groupBy` says how to group workload objects for aggregation (if there is any).
	// Each expression must evaluate to an atomic value.
	// +optional
	GroupBy []NamedExpression `json:"groupBy,omitempty"`

	// `combinedFields` defines the aggregations to do, if any.
	// `combinedFields` must be empty if `select` is not.
	// +optional
	CombinedFields []NamedAggregator `json:"combinedFields,omitempty"`

	// `select` defines named values to extract from each object.
	// `select` must be emtpy when `combinedFields` is not.
	// +optional
	Select []NamedExpression `json:"select,omitempty"`

	// `limit` limits the number of rows returned.
	// The default value is 20.
	Limit int64 `json:"limit"`
}

// NamedExpression pairs a name with a way of extracting a value from a JSON object.
type NamedExpression struct {
	Name string     `json:"name"`
	Def  Expression `json:"def"`
}

// NamedAggregator pairs a name with a way to aggregate over some objects.
// For `type=="COUNT"`, `subject` is omitted and the aggregate is the count
// of those objects that are not `null`.
// For the other types, `subject` is required and SHOULD
// evaluate to a numeric value; exceptions are handled as follows.
// For a string value: if it parses as an int64 or float64 then that is used.
// Otherwise this is an error condition: a value of 0 is used, and the error
// is reported in the BindingPolicyStatus.Errors (not necessarily repeated for each WEC).
type NamedAggregator struct {
	Name string         `json:"name"`
	Type AggregatorType `json:"type"`

	// +optional
	Subject *Expression `json:"subject,omitempty"`
}

// AggregatorType indicates what sort of aggregation is to be done.
type AggregatorType string

const (
	AggregatorTypeCount AggregatorType = "COUNT"

	AggregatorTypeSum AggregatorType = "SUM"

	AggregatorTypeAvg AggregatorType = "AVG"

	AggregatorTypeMin AggregatorType = "MIN"

	AggregatorTypeMax AggregatorType = "MAX"
)

// Expression is some value to derive from an augmented workload object from a WEC.
// The augmentation is inline addition of Augmentation.
// An Expression is either a JSONPath identifying a value to extract
// or a boolean combination of comparisons of atomic values.
// While the expression here in the Go type system admits other values,
// the controller will accept only the restricted set stated above.
type Expression struct {
	Op   ExpressionOperator `json:"op"`
	Path string             `json:"path,omitempty"`
	Args []Expression       `json:"args,omitempty"`
}

type ExpressionOperator string

const (
	OperatorPath  ExpressionOperator = "Path" // JSONPath string
	OperatorOr    ExpressionOperator = "Or"
	OperatorAnd   ExpressionOperator = "And"
	OperatorNot   ExpressionOperator = "Not"
	OperatorEqual ExpressionOperator = "Equal"
)

// Augmentation defines what is implicitly added to every workload object from a WEC,
// for purposes of status collection.
type Augmentation struct {
	Inventory   InventoryObjectReference `json:"inventory"`
	Propagation PropagationMeta          `json:"propagation"`
}

// InventoryObjectReference identifies an inventory object
type InventoryObjectReference struct {
	// `name` is the inventory object's name
	Name string `json:"name"`
}

// PropagationMeta is the last reported metadata about one workload object at one WEC.
// Some fields compare with the current state of the workload object in its WDS.
type PropagationMeta struct {
	// `infoAgeSeconds` is the age of this report, rounded up as specified in the StatusReturn
	InfoAgeSeconds int64 `json:"infoAgeSeconds"`

	// `stale` indicates whether `infoAgeSeconds > stalenessThresholdSecs`.
	Stale bool `json:"stale"`

	// `generationIsCurrent` tells whether the `metadata.Generation` in the WEC
	// equals the generation in the WDS.
	// This is `nil` when the `metadata.Generation` in the WEC is unknown.
	// +optional
	GenerationIsCurrent *bool `json:"generationIsCurrent,omitempty"`
}

// CombinedStatus holds the combined status from the WECs for one particular (workload object, BindingPolicy) pair.
// The namespace of the CombinedStatus object = the namespace of the workload object,
// or "kubestellar-report" if the workload object has no namespace.
// The name of the CombinedStatus object is the concatenation of:
// - the UID of the workload object
// - the string ":"
// - the UID of the BindingPolicy object.
// The CombinedStatus object has the following labels:
// - "status.kubestellar.io/api-group" holding the API Group (not verison) of the workload object;
// - "status.kubestellar.io/resource" holding the resource (lowercase plural) of the workload object;
// - "status.kubestellar.io/namespace" holding the namespace of the workload object;
// - "status.kubestellar.io/name" holding the name of the workload object;
// - "status.kubestellar.io/policy" holding the name of the BindingPolicy object.
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName={cs}
// +kubebuilder:printcolumn:name="WGROUP",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/api-group']"
// +kubebuilder:printcolumn:name="WRSC",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/resource']"
// +kubebuilder:printcolumn:name="WNS",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/namespace']"
// +kubebuilder:printcolumn:name="WNAME",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/name']"
// +kubebuilder:printcolumn:name="POLICY",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/policy']"
type CombinedStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `results` has an entry for every applicable NamedStatusCombiner.
	// +optional
	Results []NamedStatusCombination `json:"results,omitempty"`
}

// NamedStatusCombination holds the rows that come from evaluating one NamedStatusCombiner.
type NamedStatusCombination struct {
	Name string `json:"name"`

	ColumnNames []string `json:"columnNames"`

	// +optional
	Rows []StatusCombinationRow `json:"rows,omitempty"`
}

type StatusCombinationRow struct {
	Columns []Value `json:"columns"`
}

// Value holds a JSON value. This is a union type.
type Value struct {
	Type ValueType `json:"type"`

	// +optional
	String *string `json:"string,omitempty"`

	// Integer or floating-point, in JavaScript Object Notation.
	// +optional
	Number *string `json:"float,omitempty"`

	// +optional
	Bool *bool `json:"bool,omitempty"`

	// +optional
	Object map[string]Value `json:"object,omitempty"`

	// +optional
	Array []Value `json:"array,omitempty"`
}

type ValueType string

const (
	TypeString ValueType = "String"
	TypeNumber ValueType = "Number"
	TypeBool   ValueType = "Bool"
	TypeNull   ValueType = "Null"
	TypeObject ValueType = "Object"
	TypeArray  ValueType = "Array"
)
