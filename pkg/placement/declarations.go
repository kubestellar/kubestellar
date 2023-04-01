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

package placement

import (
	"context"
	"fmt"
	"strings"
	"sync"

	k8sevents "k8s.io/api/events/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8ssets "k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

// This file contains declarations of the interfaces between the main parts
// of the placement translator.  Those are as follows.
//
// - a WhereResolver that monitors the "where" resolutions in the
//   SinglePlacementSlice API objects.
// - a WhatResolver that monitors the EdgePlacement objects and resolves
//   their "what" predicates.
// - a SetBinder that keeps track of the bindings between resolved where
//   and resolved what.
// - a WorkloadProjector that maintains the customized workload copies
//   and syncer configuration objects in the mailbox workspaces.

// We are talking here about an assembly of several components, each
// of which likely has private data protected by a mutex.
// We thus must pay attention to avoiding deadlock.
// We do that by declaring and respecting a partial order among mutexes.
// When we say that mutex A precedes B in the locking order, this means that
// it is forbidden for a goroutine to invoke `A.Lock()` while holding B locked.
// Stated in a positive but little less precise way: a goroutine that is
// going to lock both A and B must lock A first.

// The particular locking order chosen here generally follows the pattern
// that components that drive activity precede components that get driven,
// so that this relationship can be synchronous.  For example, providers
// of maps generally precede clients of those maps.

// WhereResolver is responsible for keeping given receiver eventually
// consistent with the resolution of the "where" predicate for each EdgePlacement
// (identified by cluster and name).
type WhereResolver func(MappingReceiver[ExternalName, ResolvedWhere]) Runnable

// WhatResolver is responsible for keeping its receiver eventually consistent
// with the resolution of the "what" predicate of each EdgePlacement
// (identified by cluster ane name).
type WhatResolver func(MappingReceiver[ExternalName, WorkloadParts]) Runnable

// SetBinder is a component that is kept appraised of the "what" and "where"
// resolutions and reorganizing and picking API versions to guide the
// workload projector.
// The implementation may atomize the resolved "what" and "where"
// using differencers constructed by a SetDifferencerConstructor.
// The implementation may use a BindingOrganizer to get from the atomized
// "what" and "where" to the ProjectionMappingReceiver behavior.
type SetBinder func(workloadReceiver ProjectionMappingReceiver) (
	whatReceiver MappingReceiver[ExternalName, WorkloadParts],
	whereReceiver MappingReceiver[ExternalName, ResolvedWhere])

// WorkloadProjector is kept appraised of what goes where
// and is responsible for maintaining (a) the customized workload
// copies in the mailbox workspaces and (b) the syncer configuration objects.
type WorkloadProjector ProjectionMappingReceiver

type ProjectionMappingReceiver MappingReceiver[ProjectionKey, *ProjectionPerCluster]

// Runnable is something that can run until a given context is closed
type Runnable interface {
	Run(context.Context)
}

// RunAll is a Runnable that runs all the constituent Runnables
type RunAll []Runnable

var _ Runnable = RunAll{}

func (ra RunAll) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(len(ra))
	for _, runnable := range ra {
		go func(runnable Runnable) {
			runnable.Run(ctx)
			wg.Done()
		}(runnable)
	}
	wg.Wait()
}

func GetNamespacesBuiltIntoEdgeClusters() k8ssets.String {
	// TODO: Make this configurable
	return k8ssets.NewString("default")
}

func GetNamespacesBuiltIntoMailboxes() k8ssets.String {
	// TODO: see if more need to go here
	return k8ssets.NewString("default")
}

// AssemplePlacementTranslator puts together the top-level pieces.
func AssemplePlacementTranslator(
	whatResolver WhatResolver,
	whereResolver WhereResolver,
	setBinder SetBinder,
	workloadProjector WorkloadProjector,
) Runnable {
	whatReceiver, whereReceiver := setBinder(workloadProjector)
	runWhat := whatResolver(whatReceiver)
	runWhere := whereResolver(whereReceiver)
	return RunAll{runWhat, runWhere}
}

// ResolvedWhere identifies the set of SyncTargets that match a certain
// EdgePlacement's "where" predicate.
// Each `*edgeapi.SinglePlacementSlice` points to an immutable object.
type ResolvedWhere []*edgeapi.SinglePlacementSlice

// WorkloadParts identifies a workload prescription and provides
// ephemeral details of how to access it.
// A workload prescription is the things that match the "what" predicate
// of an EdgePlacement.
//
// Every WorkloadParts that appears in the interfaces here is immutable.
//
// In the case of a Namespace object, this implies that
// all the objects in that namespace are included.
//
// A workload may include objects of kinds that are built into
// the edge cluster.  By built-in we mean that these kinds are both
// already known (regardless of whether it is via being built into
// the apiserver or added to it by either form of aggregation) and
// not managed by edge workload management.
// It is the user's responsibility to make the "what" predicate
// match the corresponding CRD when the workload includes an object
// of a kind that is not built into the edge cluster.
type WorkloadParts map[WorkloadPartID]WorkloadPartDetails

// WorkloadPartID identifies part of a workload.
type WorkloadPartID struct {
	APIGroup string

	// Resource is the lowercase plural way of identifying the kind of object
	Resource string

	Name string
}

// WorkloadPartDetails provides additional details about how the WorkloadPart
// is to be included.
type WorkloadPartDetails struct {
	// APIVersion is version (no group) that the source workspace prefers to serve.
	APIVersion string

	// IncludeNamespaceObject is only interesting for a Namespace part, and
	// indicates whether to include the details of the Namespace object;
	// the objects in the namespace are certainly included.
	// For other parts, this field holds `false`.
	IncludeNamespaceObject bool
}

type WorkloadPart struct {
	WorkloadPartID
	WorkloadPartDetails
}

// ProjectionKey identifies the topmost level of organization,
// the combinatin of the destination and the API group and resource.
type ProjectionKey struct {
	metav1.GroupResource
	Destination edgeapi.SinglePlacement
}

// ProjectionPerCluster is the second level of organization.
// It identifies the API version to use currently and holds
// the map provider that gets to the lowest level of organization.
type ProjectionPerCluster struct {
	// APIVersion is the version to read.  Just the version, no group included
	APIVersion string

	// PerSourceCluster drives awareness of the relevant logical clusters
	// and the work to do for each.
	// This provider (a) requires receivers to be comparable and (b) deduplicates
	// additions of receivers.
	PerSourceCluster DynamicMapProvider[logicalcluster.Name, ProjectionDetails]
}

// ProjectionDetails modulates projection
type ProjectionDetails struct {

	// For namespaced resoruces, Namespaces can optionally be non-nil to restrict
	// the namespaces read from.
	Namespaces *k8ssets.String

	// For non-namespaced objects, Names can optionally be non-nil to restrict
	// the objects handled.
	Names *k8ssets.String
}

func (pd ProjectionDetails) String() string {
	var builder strings.Builder
	builder.WriteRune('{')
	if pd.Namespaces != nil {
		builder.WriteString("Namespaces: ")
		builder.WriteString(fmt.Sprintf("%v", pd.Namespaces.List()))
	}
	if pd.Names != nil {
		if pd.Namespaces != nil {
			builder.WriteString(", ")
		}
		builder.WriteString("Names: ")
		builder.WriteString(fmt.Sprintf("%v", pd.Names.List()))

	}
	builder.WriteRune('}')
	return builder.String()
}

// SetBinderConstructor is a likely signature for the final assembly of a SetBinder.
// The two set differencer constructors will be called to create set differencers
// that translate new whole values of ResolvedWhat and ResolvedWhere into
// elemental differences.
// The BindingOrganizer produces a pipe stage that is given those elemental
// differences and re-organizes them and solves the workload conflicts to
// supply input to a ProjectionMappingReceiver.
type SetBinderConstructor func(
	logger klog.Logger,
	resolvedWhatDifferencerConstructor ResolvedWhatDifferencerConstructor,
	resolvedWhereDifferencerConstructor ResolvedWhereDifferencerConstructor,
	bindingOrganizer BindingOrganizer,
	discovery APIMapProvider,
	resourceModes ResourceModes,
	eventHandler EventHandler,
) SetBinder

// SetDifferencerConstructor is a function that is given a receiver of set
// differences and returns a receiver of sets that keeps track of the latest
// set and keeps the difference receiver informed of differences as they arrive.
// The set differencer precedes the set difference receiver in the locking order.
type SetDifferencerConstructor[Set any, Element comparable] func(SetChangeReceiver[Element]) Receiver[Set]

type ResolvedWhatDifferencerConstructor = SetDifferencerConstructor[WorkloadParts, WorkloadPart]

type ResolvedWhereDifferencerConstructor = SetDifferencerConstructor[ResolvedWhere, edgeapi.SinglePlacement]

// BindingOrganizer takes a ProjectionMappingReceiver and produces a SingleBinder
// that takes the atomized bindings and reorganizes them and resolves
// the API group version issue to feed the ProjectionMappingReceiver.
// A SetBinder implementation will likely use one of these to drive its
// ProjectionMappingReceiver, feeding the SingleBinder atomized changes
// from the incoming ResolvedWhat and ResolvedWhere values.
// The given EventHandler is given events that the organizer produces
// and publishes them somewhere.
type BindingOrganizer func(discovery APIMapProvider, resourceModes ResourceModes, eventHandler EventHandler, projectionMappingReceiver ProjectionMappingReceiver) SingleBinder

// SingleBinder is appraised of individual bindings and unbindings,
// but they may come in batches.
// AddBinding calls are ordered by API machinery dependencies.
// RemoveBinding calls are ordered by the reverse of the API machinery dependencies.
type SingleBinder interface {
	// Transact does a collection of adds and removes.
	Transact(func(SingleBindingOps))
}

type SingleBindingOps SetChangeReceiver[Pair[WorkloadPart, edgeapi.SinglePlacement]]

// APIMapProvider provides API information on a cluster-by-cluster basis,
// as needed by clients.
// This information comes from runtime monitoring of the API resources
// of the clusters.
type APIMapProvider interface {
	// AddClient adds a client for a cluster.
	// All clients for the same cluster get the same provider.
	AddClient(cluster logicalcluster.Name, client Client[ScopedAPIProvider])

	// RemoveClient removes a client for a cluster.
	// Clients must be comparable.
	// Removing the last client for a given cluster causes release of
	// internal computational resources.
	RemoveClient(cluster logicalcluster.Name, client Client[ScopedAPIProvider])
}

// ScopedAPIGroupVersioner is specific to one logical cluster and
// provides a map from API group name to details about it in that cluster.
// A nil pointer means that the group is not defined in the cluster.
type ScopedAPIProvider DynamicMapProvider[string, *APIGroupInfo]

type APIGroupInfo struct {
	Versions  DynamicValueProvider[APIGroupVersions]
	Resources APIResourceDetailsProvider
}

// APIGroupVersions tells about the versions available for the group.
type APIGroupVersions struct {
	// Versions are ordered as semantic versions.
	// This slice is immutable.
	Versions []metav1.GroupVersionForDiscovery

	PreferredVersion metav1.GroupVersionForDiscovery
}

// APIResourceDetailsProvider reveals details about the resoureces in a
// given API group in a given cluster.
// A resource is identified in the usual way, the lowercase plural.
// A nil pointer means the resource is not defined there.
type APIResourceDetailsProvider DynamicMapProvider[string, *ResourceDetails]

// ResourceDetails holds the information needed here about a resource
type ResourceDetails struct {
	Namespaced        bool
	SupportsInformers bool
}

// ResourceModes tells the handling of the given resource.
// This information comes from platform configuration and code.
// Immutable.
type ResourceModes func(metav1.GroupResource) ResourceMode

// ResourceMode describes how a given resource is handled regarding
// propagation and denaturing.
type ResourceMode struct {
	PropagationMode PropagationMode
	NatureMode      NatureMode
	BuiltinToEdge   bool
}

// DefaultResourceMode is the handling for every user-defined resource
var DefaultResourceMode = ResourceMode{
	PropagationMode: GoesToEdge,
	NatureMode:      NaturalyDenatured,
	BuiltinToEdge:   false,
}

// PropagationMode describes the relationship between present-in-center and present-in-edge
type PropagationMode string

const (
	ErrorInCenter    PropagationMode = "error"
	TolerateInCenter PropagationMode = "tolerate"
	GoesToEdge       PropagationMode = "propagate"
)

// NatureMode describes the stance regarding whether a resource is denatured in the center.
// All resources that go to the edge are natured (not denatured) at the edge.
type NatureMode string

const (
	// NaturalyDenatured is a resource that is denatured in the center without any special
	// effort in this code.
	NaturalyDenatured NatureMode = "NaturallyDenatured"

	// NaturallyNatured is a resource that is natured in the center and should be that way.
	NaturallyNatured NatureMode = "NaturallyNatured"

	// ForciblyDenatured is a resource that would be given an undesired interpretation in the center
	// if stored normally in the center, so has to be stored differently in the center (but not
	// at the edge).
	ForciblyDenatured NatureMode = "ForciblyDenatured"
)

func SPMailboxWorkspaceName(sp edgeapi.SinglePlacement) string {
	return sp.Cluster + WSNameSep + string(sp.SyncTargetUID)
}

const WSNameSep = "-mb-"

// EventHandler can be given Event objects.
type EventHandler interface {
	HandleEvent(*k8sevents.Event)
}

func (where ResolvedWhere) String() string {
	var builder strings.Builder
	builder.WriteRune('[')
	for idx, slice := range where {
		if idx > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%v", slice))
	}
	builder.WriteRune(']')
	return builder.String()
}
