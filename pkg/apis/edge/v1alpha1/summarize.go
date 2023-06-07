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

// +crd
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Summarizer defines a collection of ways to summarize a collection of objects.
type Summarizer struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Groupers is a collection of named ways to summarize objects.
	// +listType=map
	// +listMapKey=name
	// +patchStrategy=merge
	// +patchMergeKey=name
	Groupers []Grouper `json:"groupers" patchStrategy:"merge" patchMergeKey:"name"`

	// `stalenessThresholdSecs` is the threshold used to derive
	// the `stale` bit in SyncerSyntheticStatus values.
	StalenessThresholdSecs int32 `json:"stalenessThresholdSecs"`

	// `broken` defines which objects are considered to be "broken" ---
	// that is, which contribute to the list of broken objects in
	// the corresponding Summary object.
	Broken OrOfAndsOfTests `json:"broken"`
}

// Grouper is a named way to summarize objects.
// The corresponding summary is a map from group-by tuple to aggregate tuple.
type Grouper struct {
	// `name` is the name of this way of summarizing
	Name string `json:"name"`

	// `groupBy` defines the group-by tuples.
	// Each group-by tuple has an entry for each member of `groupBy`.
	// The summary has an entry for every combination of member values
	// that exists among the objects being summarized.
	// TODO: do we need to define equality between JSON arrays and/or objects,
	// or can we restrict the values here to atomic ones?
	// +listType=map
	// +listMapKey=name
	// +patchStrategy=merge
	// +patchMergeKey=name
	GroupBy []NamedExpression `json:"groupBy" patchStrategy:"merge" patchMergeKey:"name"`

	// `aggregators` defines the aggregate tuple.
	// Each aggregate tuple has an for each member of `aggregators`,
	// which defines how to aggregate the objects with the corresponding
	// group-by tuple.
	// +listType=map
	// +listMapKey=name
	// +patchStrategy=merge
	// +patchMergeKey=name
	Aggregators []NamedAggregator `json:"aggregators" patchStrategy:"merge" patchMergeKey:"name"`
}

// NamedExpression pairs a name with a way of extracting a value from a JSON object.
type NamedExpression struct {
	Name string     `json:"name"`
	Def  Expression `json:"def"`
}

// NamedAggregator pairs a name with a way to aggregate over some objects.
// For `type=="COUNT"`, `subject` is omitted and the aggregate is the count
// of those objects.  For the other types, `subject` is required and SHOULD
// evaluate to a numeric value; exceptions are handled as follows.
// For a string value: if it parses as an int64 or float64 then that is used.
// Otherwise this is an error condition and a value of 0 is used.
// TODO: define how those error conditions are reported to the user.
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

// Expression is a JSONPath expression of how to extract a value
// from an object.  Naturally, the JSON representation of that object
// is what is relevant here.  Although a general JSONPath expression can
// evaluate to zero, one, or more values, the intent here is to use
// only JSONPath expressions that evaluate to one value.
// If more than one value results, the first value is used.
// If zero values result, the value `null` is used.
// Every object is implicitly extended with a member whose
// name is "syncerStatus" and whose value is the JSON
// rendering of the corresponding SyncerSyntheticStatus.
type Expression string

// SyncerRawStatusName is the name of an annotation that the syncer
// maintains on the objects it downsyncs from or upsyncs to mailbox
// workspaces.  The value is the JSON rendering of a SyncerRawStatus.
const SyncerRawStatusName = "edge.kcp.io/syncer-status"

// SyncerRawStatus is maintained by the syncer, on objects that
// it works on in the mailbox workspace.
type SyncerRawStatus struct {
	// `lastSyncerReportTime` is when the syncer last reported on
	// its work on the relevant object.
	LastSyncerReportTime metav1.Time `json:"lastSyncerReportTime"`

	// `syncedGeneration` is meaningful only for downsynced objects.
	// It reports on the `metadata.generation` (in the mailbox workspace) of
	// the last desired state that was successfully written to the edge copy.
	// +optional
	SyncedGeneration string `json:"syncedGeneration,omitempty"`
}

// SyncerSyntheticStatus is implicitly derived during summarization.
type SyncerSyntheticStatus struct {
	// `stale` indicates whether the time since `lastSyncerReportTime`
	// is longer than the staleness threshold.
	Stale bool `json:"stale"`

	// `generationCurrent` reports whether `syncedGeneration` equals
	// the current generation.
	GenerationCurrent bool `json:"generationCurrent"`

	// `exists` reports whether `syncedGeneration` is non-empty.
	Exists bool `json:"exists"`
}

// OrOfAndsOfTests is a disjunctive normal form predicate over
// objects as they appear in a mailbox workspace and implicitly
// extended by "syncerStatus".
type OrOfAndsOfTests []AndOfTests

// AndOfTests is a conjunction of atomic tests.
type AndOfTests []ObjectTest

// ObjectTest is an atomic test of an object in a mailbox
// workspace implicitly extended by "syncerStatus".
// This is an equality test, which may be negated
// to make it an inequality test.
// The value is expressed in JSON.
// TODO: do we have to handle non-atomic values here?
type ObjectTest struct {
	Part   Expression `json:"part"`
	Value  string     `json:"value"`
	Negate bool       `json:"negate,omitempty"`
}

//+kubebuilder:object:root=true

// SummarizerList is the API type for a list of Summarizer
type SummarizerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Summarizer `json:"items"`
}

// The following label names are for labels that link a Summary object
// to the reason for its existence.
// NOTE WELL: because a label value can not be longer than 63 characters,
// this implies an extra limit on the namespaces and names of related objets.
const (
	// LabelSubjectResource is the Summary object label name that is paired
	// with the the `metav1.GroupVersionResource` of the objects being summarized.
	// The syntax of the value is "resource.version.group" with the group
	// part being empty for the Kubernetes core API group.
	LabelSubjectResource = "edge.kcp.io/subject-resource"

	// LabelSubjectNamespace is the Summary object label name that is paired
	// with the namespace of the objects being summarized --- if they are namespaced.
	// For non-namespaced objects, the Summary has no such label.
	LabelSubjectNamespace = "edge.kcp.io/subject-namespace"

	// LabelSubjectName is the Summary object label name that is paired with
	// the name of the objects being summarized.
	LabelSubjectName = "edge.kcp.io/subject-name"

	// LabelSummarizerName is the Summary object label name that is paired with
	// the name of the Summarizer that configured the Summary.
	LabelSummarizerName = "edge.kcp.io/summarizer-name"

	// LabelEdgePlacement is the Summary object label name that is paired with
	// the name of the EdgePlacement requesting that summary.
	LabelEdgePlacement = "edge.kcp.io/edge-placement"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StatusSummary holds the summary of status from the edge for one particular object
// and EdgePlacement.  The object in question is either an object downsynced from the
// center or collection of same-named objects upsynced from the edge.
// A given StatusSummary appears in the same namespace as the Summarizer used to produce
// that summary, with a name chosen by the implementation.
// Labels, defined above, link this object to the subject of summarization and the
// EdgePlacement and Summarizer responsible; they are not referenced in
// `meta.ownerReferences` (deletion is done in a non-generic way).
// The summary holds aggregation results and a size-capped list of broken objects.
type StatusSummary struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Groupings holds aggregations of the status from the edge for the associated object.
	// There is a Grouping here for every member of `groupers` in the Summarizer.
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge
	Groupings []Grouping `patchStrategy:"merge" patchMergeKey:"name" json:"counts,omitempty"`

	// `brokens` holds a length-capped list of edge clusters where the
	// the subject has a status that is classified as "broken".
	// If there are more than 20 such Locations, only 20 of them are listed here
	// (no promise here about _which_ 20).
	Brokens []string `json:"brokens,omitempty"`
}

// Grouping is the result of summarizing some objects according to one Grouper.
type Grouping struct {
	// `name` is the name of the Grouper.
	Name string `json:"name"`

	// `grouops` holds the map from group-by tuple to aggregate tuple.
	Groups []Group `json:"groups"`
}

// Group is the association of one group-by tuple with one aggregate tuple
type Group struct {
	// `groupedBy` is one group-by tuple
	GroupedBy []NamedValue `json:"groupedBy"`

	// `aggregates` is one aggregate tuple
	Aggregates []NamedValue `json:"aggregates"`
}

// NamedValue pairs a name with a value expressed in JSON
type NamedValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StatusSummaryList is the API type for a list of StatusSummary
type StatusSummaryList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StatusSummary `json:"items"`
}
