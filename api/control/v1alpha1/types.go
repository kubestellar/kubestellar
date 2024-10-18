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
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// TemplateExpansionAnnotationKey, when paired with the value "true" in an annotation of
// a workload object in a WDS, indicates that Go template expansion should be
// bundled with propagation from core to WEC.
//
// Go template expansion means to (1) parse each leaf string of the object as a Go template
// as defined in the Go standard package "text/template", and (2) for each WEC, replace that
// leaf string with the string that results from expanding this template
// (`Template.Execute`) using properties of the WEC.
//
// The properties for a given WEC are collected from the following four sources, in order.
// For a property defined by multiple sources, the first one in this order takes precedence.
// The first source is a ConfigMap object, if it exists, that: (a) has the same name as the WEC's
// inventory object, (b) is in the namespace named "customization-properties", and (c) is
// in the Inventory and Transport Space (ITS). In particular, the string and binary data entries
// whose name is valid as a Go language identifier provide properties.
// The second source is the annotations of the WEC's inventory object,
// when the name (AKA key) of that annotation is valid as a Go language identifier.
// The third source is the labels of the WEC's inventory object,
// when the name (AKA key) of that label is valid as a Go language identifier.
// The fourth source is some built-in definitions, of which there is presently just one:
// the value of the property named "clusterName" is the name of the WEC's inventory object.
//
// Any failure in any template expansion for a given Binding suppresses propagation of
// desired state from that Binding; the previosly propagated desired state from that Binding,
// if any, remains in place in the WEC.
//
// Note that this sort of customization has limited applicability.  It can only be used where
// the un-expanded string passes the validation conditions of the relevant object type.
// For more broadly applicable customization, see Customizer objects.

const TemplateExpansionAnnotationKey string = "control.kubestellar.io/expand-templates"

// PropertyConfigMapNamespace is the namespace in the ITS that holds ConfigMap objects that provide
// WEC properties to be used in customization.
const PropertyConfigMapNamespace = "customization-properties"

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
	// and modulates their downsync.
	// An object is selected if it matches at least one member of this list.
	// When multiple DownsyncPolicyClause match the same workload object:
	// the `createOnly` bits are ORed together, and the StatusCollector reference
	// sets are combined by union.
	Downsync []DownsyncPolicyClause `json:"downsync,omitempty"`

	// WantSingletonReportedState means that for workload objects that are distributed --- taking
	// all BindingPolicies into account --- to exactly one WEC, the object's reported state
	// from the WEC should be written to the object in its WDS.
	// If any of the workload objects are distributed to more or less than 1 WEC then
	// the `.status.errors` of this policy will report that discrepancy for
	// some of them.
	// +optional
	WantSingletonReportedState bool `json:"wantSingletonReportedState,omitempty"`
}

const (
	ValidationErrorKeyPrefix string = "validation-error.kubestellar.io/"

	// BindingPolicyConditionSatisfied means BindingPolicy requirements are satisfied.
	// A BindingPolicy is not satisfied only if the set of selected clusters is empty
	BindingPolicyConditionSatisfied string = "BindingPolicySatisfied"

	// BindingPolicyConditionMisconfigured means BindingPolicy configuration is incorrect.
	BindingPolicyConditionMisconfigured string = "BindingPolicyMisconfigured"
)

// DownsyncPolicyClause identifies some objects (by a predicate)
// and modulates how they are downsynced.
// One modulation is specifying a set of StatusCollectors to apply
// to returned status.
// The other modulation is specifying whether the propagation from WDS to WEC
// involves continual maintenance of the spec or the object is only created
// if it is absent.
type DownsyncPolicyClause struct {
	DownsyncObjectTest `json:",inline"`

	// `createOnly` indicates that in a given WEC, the object is not to be updated
	// if it already exists.
	// +optional
	CreateOnly bool `json:"createOnly,omitempty"`

	// `statusCollection` holds the rules for collecting status from the WECs.
	// +optional
	StatusCollection *StatusCollection `json:"statusCollection,omitempty"`
}

// StatusCollection holds the rules for collecting status from the WECs.
type StatusCollection struct {
	// `statusCollectors` is a list of StatusCollectors name references to be applied.
	StatusCollectors []string `json:"statusCollectors,omitempty"`
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
// Validation might not be fully checked by apiservers;
// if not prevented by the apiserver then violations will be reported in `.status.errors`.
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

// BindingPolicyStatus defines the observed state of BindingPolicy
type BindingPolicyStatus struct {
	// +optional
	Conditions []BindingPolicyCondition `json:"conditions"`

	ObservedGeneration int64 `json:"observedGeneration"`

	// +optional
	Errors []string `json:"errors,omitempty"`
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
// +kubebuilder:subresource:status
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

	Status BindingStatus `json:"status,omitempty"`
}

// BindingSpec holds a list of object references with their associated resource versions,
// and a list of destinations which are the resolution of a BindingPolicy's `what` and `where`:
// what objects to propagate and to where.
// All objects referenced in this spec are propagated to all destinations present.
type BindingSpec struct {
	// `workload` is a collection of namespaced and cluster scoped object references and their associated
	// data - resource versions, create-only bits, and statuscollectors - to be propagated to destination clusters.
	Workload DownsyncObjectClauses `json:"workload,omitempty"`

	// `destinations` is a list of cluster-identifiers that the objects should be propagated to.
	// No duplications are allowed in this list.
	// +listType=map
	// +listMapKey=clusterId
	Destinations []Destination `json:"destinations,omitempty"`
}

// DownsyncObjectClauses defines the objects to be down-synced, grouping them by scope.
// It specifies a set of object references with their associated resource versions, to be downsynced.
// Each object reference is associated with a set of statuscollectors that should be applied to it.
type DownsyncObjectClauses struct {
	// `clusterScope` holds a list of references to cluster-scoped objects to downsync and how the
	// downsync is to be modulated.
	// No duplications.
	// +listType=map
	// +listMapKey=group
	// +listMapKey=resource
	// +listMapKey=name
	ClusterScope []ClusterScopeDownsyncClause `json:"clusterScope,omitempty"`

	// `namespaceScope` holds a list of references to namsepace-scoped objects to downsync and how the
	// downsync is to be modulated.
	// No duplications.
	// +listType=map
	// +listMapKey=group
	// +listMapKey=resource
	// +listMapKey=namespace
	// +listMapKey=name
	NamespaceScope []NamespaceScopeDownsyncClause `json:"namespaceScope,omitempty"`
}

// NamespaceScopeDownsyncClause references a specific namespace-scoped object to downsync,
// and the status collectors that should be applied to it.
type NamespaceScopeDownsyncClause struct {
	NamespaceScopeDownsyncObject `json:",inline"`

	// `createOnly` indicates that in a given WEC, the object is not to be updated
	// if it already exists.
	// +optional
	CreateOnly bool `json:"createOnly,omitempty"`

	// `statusCollection` holds the rules of status collection for the object.
	// +optional
	StatusCollection *StatusCollection `json:"statusCollection,omitempty"`
}

// NamespaceScopeDownsyncObject references a specific namespace-scoped object to downsync,
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

// ClusterScopeDownsyncClause references a specific cluster-scoped object to downsync,
// and the status collectors that should be applied to it.
type ClusterScopeDownsyncClause struct {
	ClusterScopeDownsyncObject `json:",inline"`

	// `createOnly` indicates that in a given WEC, the object is not to be updated
	// if it already exists.
	// +optional
	CreateOnly bool `json:"createOnly,omitempty"`

	// `statusCollection` holds the rules of status collection for the object.
	// +optional
	StatusCollection *StatusCollection `json:"statusCollection,omitempty"`
}

// ClusterScopeDownsyncObject references a specific cluster-scoped object to downsync,
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
	ClusterId string `json:"clusterId"`
}

type BindingStatus struct {
	ObservedGeneration int64    `json:"observedGeneration"`
	Errors             []string `json:"errors,omitempty"`
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

// StatusCollector defines one way to collect status about a given workload object from
// the set of WECs that it propagates to.
// This is modeled after an SQL SELECT statement that does aggregation.
//
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName={sc}
type StatusCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StatusCollectorSpec   `json:"spec,omitempty"`
	Status StatusCollectorStatus `json:"status,omitempty"`
}

// StatusCollectorSpec defines the desired state of StatusCollector.
type StatusCollectorSpec struct {
	// `filter`, if given, is applied first.
	// It must evaluate to a boolean or null (which is treated as false).
	// This is like the WHERE clause in an SQL SELECT statement.
	// +optional
	Filter *Expression `json:"filter,omitempty"`

	// `groupBy` says how to group workload objects for aggregation (if there is any).
	// Each expression must evaluate to an atomic value.
	// `groupBy` must be empty if `combinedFields` is.
	// +optional
	GroupBy []NamedExpression `json:"groupBy,omitempty"`

	// `combinedFields` defines the aggregations to do, if any.
	// `combinedFields` must be empty if `select` is not.
	// +optional
	CombinedFields []NamedAggregator `json:"combinedFields,omitempty"`

	// `select` defines named values to extract from each object.
	// `select` must be empty when `combinedFields` is not.
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
//
// - For `type=="COUNT"`, `subject` is omitted and the aggregate is the count
// of those objects that are not `null`.
//
// - For the other types, `subject` is required and SHOULD
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
// The AVG of no values is NaN;
// for the other types the aggregation of no values is the identity element of the combining operation
// in `float64`.
type AggregatorType string

const (
	AggregatorTypeCount AggregatorType = "COUNT"
	AggregatorTypeSum   AggregatorType = "SUM"
	AggregatorTypeAvg   AggregatorType = "AVG"
	AggregatorTypeMin   AggregatorType = "MIN"
	AggregatorTypeMax   AggregatorType = "MAX"
)

// Expression is written in the [Common Expression Language](https://cel.dev/).
// See github.com/google/cel-go for the Go implementation used in Kubernetes,
// and https://kubernetes.io/docs/reference/using-api/cel/ about CEL's uses in Kubernetes.
// The expression will be type-checked against the schema for the object type at hand,
// using the Kubernetes library code for converting an OpenAPI schema to a CEL type
// (e.g., https://github.com/kubernetes/apiserver/blob/v0.28.2/pkg/cel/common/schemas.go#L40).
// Parsing errors are posted to the status.Errors of the StatusCollector.
// Type checking errors are posted to the status.Errors of the Binding and BindingPolicy.
type Expression string

// ExpressionContext defines what an Expression can reference regarding the workload object.
type ExpressionContext struct {
	// `inventory` holds the inventory record for the workload object.
	Inventory InventoryRecord `json:"inventory"`

	// `obj` holds a copy of the workload object as read from the WDS.
	Obj runtime.RawExtension `json:"obj"`

	// `returned` holds the fragment of the workload object that was returned to the core from the WEC.
	Returned ReturnedState `json:"returned"`

	// `propagation` holds data about the current state of the work on propagating
	// the object's state from WDS to WEC and from WEC to WDS.
	Propagation PropagationData `json:"propagation"`
}

// InventoryRecord is what appears in the inventory for a given WEC.
type InventoryRecord struct {
	// the name of the WEC.
	Name string `json:"name"`
}

type ReturnedState struct {
	Status runtime.RawExtension `json:"status"`
}

type PropagationData struct {
	// `lastReturnedUpdateTimestamp` is the time of the last update to any
	// of the returned object state in the core.
	// Before the first such update, this holds the zero value of `time.Time`.
	LastReturnedUpdateTimestamp metav1.Time `json:"lastReturnedUpdateTimestamp"`

	// `lastGeneration` is that last `ObjectMeta.Generation` from the WDS that
	// propagated to the WEC. This is not to imply that it was successfully applied there;
	// for that, see `lastGenerationIsApplied`.
	// Zero means that none has yet propagated there.
	LastGeneration int64 `json:"lastGeneration"`

	// `lastGenerationIsApplied` indicates whether `lastGeneration` has been successfully
	// applied in the WEC.
	LastGenerationIsApplied bool `json:"lastGenerationIsApplied"`

	// `lastCurrencyUpdateTime` is the time of the latest update to either
	// `lastGeneration` or `lastGenerationIsApplied`. More precisely, it is
	// the time when the core became informed of the update.
	// Before the first such update, this holds the zero value of `time.Time`.
	LastCurrencyUpdateTime metav1.Time `json:"lastCurrencyUpdateTime"`
}

// StatusCollectorStatus defines the observed state of StatusCollector.
type StatusCollectorStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	// +optional
	Errors []string `json:"errors,omitempty"`
}

// StatusCollectorList is the API type for a list of StatusCollector.
//
// +kubebuilder:object:root=true
type StatusCollectorList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StatusCollector `json:"items"`
}

// CombinedStatus holds the combined status from the WECs for one particular (workload object, BindingPolicy) pair.
// The namespace of the CombinedStatus object is the namespace of the workload object,
// or "kubestellar-report" if the workload object has no namespace.
// The name of the CombinedStatus object is the concatenation of:
// - the UID of the workload object
// - the string "."
// - the UID of the BindingPolicy object.
// The CombinedStatus object has the following labels:
// - "status.kubestellar.io/api-group" holding the API Group (not verison) of the workload object;
// - "status.kubestellar.io/resource" holding the resource (lowercase plural) of the workload object;
// - "status.kubestellar.io/namespace" holding the namespace of the workload object;
// - "status.kubestellar.io/name" holding the name of the workload object;
// - "status.kubestellar.io/binding-policy" holding the name of the BindingPolicy object.
//
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName={cs}
// +kubebuilder:printcolumn:name="SUBJECT_GROUP",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/api-group']"
// +kubebuilder:printcolumn:name="SUBJECT_RSC",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/resource']"
// +kubebuilder:printcolumn:name="SUBJECT_NS",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/namespace']"
// +kubebuilder:printcolumn:name="SUBJECT_NAME",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/name']"
// +kubebuilder:printcolumn:name="BINDINGPOLICY",type="string",JSONPath=".metadata.labels['status\\.kubestellar\\.io/policy']"
type CombinedStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `results` has an entry for every applicable StatusCollector.
	// +optional
	Results []NamedStatusCombination `json:"results,omitempty"`
}

// NamedStatusCombination holds the rows that come from evaluating one StatusCollector.
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
	Object *v1.JSON `json:"object,omitempty"`

	// +optional
	Array *v1.JSON `json:"array,omitempty"`
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

// CombinedStatusList is the API type for a list of CombinedStatus.
//
// +kubebuilder:object:root=true
type CombinedStatusList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CombinedStatus `json:"items"`
}

// CustomTransform describes how to select and transform some objects
// on their way from WDS to WEC, without regard to the WEC (i.e.,
// not changes that are specific to the individual WEC).
// The transformation specified here is in addition to, and follows,
// whatever is built into KubeStellar for that object.
//
// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName={ct},categories={all}
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SUBJECT_GROUP",type="string",JSONPath=".spec.apiGroup"
// +kubebuilder:printcolumn:name="SUBJECT_RESOURCE",type="string",JSONPath=".spec.resource"
type CustomTransform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomTransformSpec   `json:"spec,omitempty"`
	Status CustomTransformStatus `json:"status,omitempty"`
}

// CustomTransformSpec selects some objects and describes how to transform them.
// The selected objects are those that match the `apiGroup` and `resource` fields.
type CustomTransformSpec struct {
	// `apiGroup` holds just the group, not also the version
	APIGroup string `json:"apiGroup"`

	// `resource` is the lowercase plural way of identifying a sort of object.
	// "subresources" can not be directly bound to, only whole (top-level) objects.
	Resource string `json:"resource"`

	// `remove` is a list of JSONPath expressions (https://goessner.net/articles/JsonPath/)
	// that identify part of the object to remove if present.
	// Only a subset of JSONPath is supported.
	// The expression used in a filter must be a conjunction of field == literal tests.
	// Examples:
	// - "$.spec.resources.GenericItems[*].generictemplate.metadata.resourceVersion"
	// - "$.store.book[?(@.author == 'Kilgore Trout' && @.category == 'fiction')].price"
	// +optional
	Remove []string `json:"remove,omitempty"`
}

type CustomTransformStatus struct {
	ObservedGeneration int64 `json:"observedGeneration"`

	// +optional
	Errors []string `json:"errors,omitempty"`

	// +optional
	Warnings []string `json:"warnings,omitempty"`
}

// CustomTransformList is the API type for a list of CustomTransform
//
// +kubebuilder:object:root=true
type CustomTransformList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomTransform `json:"items"`
}
