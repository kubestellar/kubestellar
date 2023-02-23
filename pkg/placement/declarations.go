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
	"fmt"
	"strings"

	k8sevents "k8s.io/api/events/v1"
	apimachtypes "k8s.io/apimachinery/pkg/types"

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
//   in the mailbox workspaces.
// - a PlacementProjector that maintains the TMC Placement objects that
//   correspond to the EdgePlacement objects.

// WhereResolver is responsible for keeping given consumers eventually
// consistent with the resolutions of the EdgePlacement "where" predicates.
// This is done by a calls to the given consumers.
// Calls to Get are not associated with any particular consumer.
// Each `*edgeapi.SinglePlacementSlice` points to an immutable object.
type WhereResolver interface {
	AddConsumer(consume func(placementRef edgeapi.ExternalName, resolved ResolvedWhere))
	Get(placementRef edgeapi.ExternalName) ResolvedWhere
}

// WhatResolver is responsible for resolving the "what" predicates of the EdgePlacement
// objects and keeping given consumers eventually consistent with these resolutions.
// This is done by a calls to the given consumers.
// Calls to Get are not associated with any particular consumer.
type WhatResolver interface {
	AddConsumer(consume func(placementRef edgeapi.ExternalName, resolved WorkloadParts))
	Get(placementRef edgeapi.ExternalName) WorkloadParts
}

// SetBinder is a component that keeps track of the mapping
// from EdgePlacement to its pair of (resolved what, resolved where).
// An implementation may be responsible, for example, for keeping a given
// SingleBinder eventually consistent.
// An implementation is likely going to use a SinglePlacementSliceSetReducer.
// Each `*edgeapi.SinglePlacementSlice` points to an immutable object.
type SetBinder interface {
	NoteBinding(placementRef edgeapi.ExternalName, what WorkloadParts, where ResolvedWhere)
	NoteUnbinding(placementRef edgeapi.ExternalName)
}

// WorkloadProjector is kept appraised of what goes where
// and is responsible for maintaining the customized workload
// copies in the mailbox workspaces.
type WorkloadProjector interface {
	SingleBinder
}

// PlacementProjector is responsible for maintaining the TMC Placement
// objects corresponding to given EdgePlacement objects.
// An implementation is likely going to use a SinglePlacementSliceSetReducer.
type PlacementProjector interface {
	SetBinder
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
	ClusterName logicalcluster.Name

	APIGroup string

	// Resource is the lowercase plural way of identifying the kind of object
	Resource string

	Name string
}

// WorkloadPartDetails provides additional details about how the WorkloadPart
// is to be included.
type WorkloadPartDetails struct {
	// APIVersion is the one to use when reading from the source workspace.
	APIVersion string

	// IncludeNamespaceObject is only relevant for a Namespace part, and
	// indicates whether to include the details of the Namespace object;
	// the objects in the namespace are certainly included.
	IncludeNamespaceObject bool
}

// SingleBinder is appraised of individual bindings and unbindings,
// but they may come in batches.
// AddBinding calls are ordered by API machinery dependencies.
// RemoveBinding calls are ordered by the reverse of the API machinery dependencies.
type SingleBinder interface {
	// Transact does a collection of adds and removes.
	Transact(func(SingleBindingOps))
}

type SingleBindingOps interface {
	AddBinding(what WorkloadPart, where SinglePlacement)
	RemoveBinding(what WorkloadPart, where SinglePlacement)
}

type WorkloadPart struct {
	WorkloadPartID
	WorkloadPartDetails
}

// SinglePlacement extends the API struct with the UID of the SyncTarget.
type SinglePlacement struct {
	edgeapi.SinglePlacement
	SyncTargetUID apimachtypes.UID
}

func (sp SinglePlacement) SyncTargetRef() edgeapi.ExternalName {
	return edgeapi.ExternalName{Workspace: sp.Location.Workspace, Name: sp.SyncTargetName}
}

// SinglePlacementSliceSetReducer keeps track of
// a slice of `*edgeapi.SinglePlacementSlice`.
// Typically one of these has a UIDer and
// a SinglePlacementSetChangeConsumer that is kept
// appraised of the set of values in those slices, extended by the
// SyncTarget UIDs.
// Each `*edgeapi.SinglePlacementSlice` points to an immutable object.
type SinglePlacementSliceSetReducer interface {
	Set(ResolvedWhere)
}

// SinglePlacementSetChangeConsumer is something that is kept
// incrementally appraised of a set of SinglePlacement values.
type SinglePlacementSetChangeConsumer interface {
	Add(SinglePlacement)
	Remove(SinglePlacement)
}

// UIDer is a source of mapping from object name to UID.
// One of these is specific to one kind of object,
// which is not namespaced.
type UIDer interface {
	AddConsumer(func(edgeapi.ExternalName, apimachtypes.UID))
	GetUID(edgeapi.ExternalName) apimachtypes.UID
}

func (sp *SinglePlacement) MailboxWorkspaceName() string {
	return sp.Location.Workspace + WSNameSep + string(sp.SyncTargetUID)
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
