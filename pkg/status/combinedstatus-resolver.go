/*
Copyright 2024 The KubeStellar Authors.

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

package status

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/abstract"
	"github.com/kubestellar/kubestellar/pkg/binding"
	"github.com/kubestellar/kubestellar/pkg/util"
)

type CombinedStatusResolver interface {
	// TODO: API for generating/comparing combinedstatus resolutions

	// NoteBindingResolution notes a binding resolution for status collection.
	//
	// 1. If `deleted` is true, the associated combinedstatus resolutions are
	// removed from memory. The same is done if a resolution no longer requires
	// status collection.
	//
	// 2. Excessive combinedstatus resolutions are removed if they are no longer
	// associated with the binding.
	//
	// 3. For every workload object associated with one or more statuscollectors, a
	// combinedstatus resolution is created/updated.
	// In the case of creation, the binding.Resolution's OwnerReference is used
	// to identify the bindingpolicy that the resolution is associated with.
	//
	// The returned array contains:
	//   - The identifiers of workload objects that should have their
	//     associated workstatuses processed for evaluation.
	//   - The identifiers of statuscollectors that need to be processed and
	//     cached.
	NoteBindingResolution(bindingResolution binding.Resolution, deleted bool) sets.Set[util.ObjectIdentifier]

	// NoteStatusCollector notes a statuscollector's spec.
	// The statuscollector is cached on the resolver's level, and is updated
	// for every resolution it is involved with.
	// The returned array contains the identifiers of workload objects that
	// should have their associated workstatuses processed for evaluation.
	// The statuscollector is assumed to be valid.
	NoteStatusCollector(statusCollector *v1alpha1.StatusCollector) bool

	// NoteWorkStatus notes a workstatus's content in the combinedstatus
	// resolution associated with its source workload object.
	// The returned boolean indicates whether the combinedstatus resolution
	// was updated.
	NoteWorkStatus(ctx context.Context, workStatusIdentifier util.ObjectIdentifier,
		workStatusContent *runtime.RawExtension) bool

	// ResolutionExists returns true if a combinedstatus resolution exists for
	// the given binding name and workload object identifier.
	ResolutionExists(bindingName string, objectIdentifier util.ObjectIdentifier) bool
}

// NewCombinedStatusResolver creates a new CombinedStatusResolver.
func NewCombinedStatusResolver(celEvaluator *celEvaluator) (CombinedStatusResolver, error) {
	celEvaluator, err := newCELEvaluator()
	if err != nil {
		return nil, fmt.Errorf("failed to create CEL evaluator: %w", err)
	}

	return &combinedStatusResolver{
		celEvaluator:             celEvaluator,
		bindingNameToResolutions: make(map[string]map[util.ObjectIdentifier]*combinedStatusResolution),
	}, nil
}

type combinedStatusResolver struct {
	sync.RWMutex
	celEvaluator *celEvaluator
	// bindingNameToResolutions is a map of binding names to their resolution
	// entries. The latter is a map of object identifiers to their
	// combinedstatus resolutions.
	bindingNameToResolutions map[string]map[util.ObjectIdentifier]*combinedStatusResolution
	// statusCollectorNameToClauses is a map of statuscollector names to their
	// clauses. This serves as a cache that is the source of truth for
	// statuscollectors that are used in the combinedstatus resolutions.
	statusCollectorNameToClauses map[string]statusCollectorClauses
}

type statusCollectorClauses struct {
	filter *v1alpha1.Expression

	groupBy        []v1alpha1.NamedExpression
	combinedFields []v1alpha1.NamedAggregator

	selectClause []v1alpha1.NamedExpression

	limit int64
}

// NoteBindingResolution notes a binding resolution for status collection.
//
// 1. If `deleted` is true, the associated combinedstatus resolutions are
// removed from memory. The same is done if a resolution no longer requires
// status collection.
//
// 2. Excessive combinedstatus resolutions are removed if they are no longer
// associated with the binding.
//
// 3. For every workload object associated with one or more statuscollectors, a
// combinedstatus resolution is created/updated.
// In the case of creation, the binding.Resolution's OwnerReference is used
// to identify the bindingpolicy that the resolution is associated with.
//
// The returned array contains:
//   - The identifiers of workload objects that should have their
//     associated workstatuses processed for evaluation.
//   - The identifiers of statuscollectors that need to be processed and
//     cached.
func (c *combinedStatusResolver) NoteBindingResolution(bindingResolution binding.Resolution,
	deleted bool) sets.Set[util.ObjectIdentifier] {
	c.Lock()
	defer c.Unlock()

	identifiersToQueue := sets.Set[util.ObjectIdentifier]{}

	// (1)
	if deleted {
		// for every combinedstatus resolution - queue its object identifier for syncing
		// and remove the resolution from memory
		for _, combinedStatusResolution := range c.bindingNameToResolutions[bindingResolution.Name] {
			identifiersToQueue.Insert(combinedStatusResolution.getIdentifier())
		}

		delete(c.bindingNameToResolutions, bindingResolution.Name)
		return identifiersToQueue
	}

	// if the binding resolution is not yet noted - create a new entry
	objectIdentifierToResolutions, exists := c.bindingNameToResolutions[bindingResolution.Name]
	if !exists {
		objectIdentifierToResolutions = make(map[util.ObjectIdentifier]*combinedStatusResolution)
		c.bindingNameToResolutions[bindingResolution.Name] = objectIdentifierToResolutions
	}

	// (2) remove excessive combinedstatus resolutions of objects that are no longer
	// associated with the binding resolution
	for objectIdentifier, csResolution := range objectIdentifierToResolutions {
		if _, exists := bindingResolution.ObjectIdentifierToData[objectIdentifier]; !exists {
			identifiersToQueue.Insert(csResolution.getIdentifier())
			delete(objectIdentifierToResolutions, objectIdentifier)
		}
	}

	// (~2+3) create/update combinedstatus resolutions for every object that requires status collection,
	// and delete resolutions that are no longer required
	for objectIdentifier, objectData := range bindingResolution.ObjectIdentifierToData {
		csResolution, exists := objectIdentifierToResolutions[objectIdentifier]
		if len(objectData.StatusCollectors) == 0 {
			if exists { // associated resolution is no longer required
				identifiersToQueue.Insert(csResolution.getIdentifier())
				delete(objectIdentifierToResolutions, objectIdentifier)
			}

			continue
		}

		// create resolution entry if missing
		if !exists {
			csResolution = &combinedStatusResolution{
				statusCollectorNameToData:       make(map[string]*statusCollectorData),
				workStatusToStatusCollectorName: make(map[util.ObjectIdentifier]sets.Set[string]),
				ownerReference:                  &bindingResolution.OwnerReference,
			}
			objectIdentifierToResolutions[objectIdentifier] = csResolution
			identifiersToQueue.Insert(csResolution.getIdentifier()) // should be queued up to create itself
		}

		// update statuscollectors
		queueCombinedStatus, queueObjWorkStatuses := csResolution.setStatusCollectors(
			abstract.SliceToMap(objectData.StatusCollectors,
				func(statusCollectorName string) string { return statusCollectorName }, // keys mapper
				func(statusCollectorName string) *statusCollectorClauses { // val mapper
					if clauses, exists := c.statusCollectorNameToClauses[statusCollectorName]; exists {
						return clauses.clone()
					}
					// not cached - need to queue
					identifiersToQueue.Insert(util.IdentifierForStatusCollector(statusCollectorName))
					return nil
				}))

		// queue the combinedstatus object for syncing if changed
		if queueCombinedStatus {
			identifiersToQueue.Insert(csResolution.getIdentifier())
		}

		// queue the source object identifier to have its workstatuses evaluated
		if queueObjWorkStatuses {
			identifiersToQueue.Insert(objectIdentifier)
		}
	}

	return identifiersToQueue
}

// NoteWorkStatus notes a workstatus's content in the combinedstatus
// resolution associated with its source workload object.
// The returned boolean indicates whether the combinedstatus resolution
// was updated.
func (c *combinedStatusResolver) NoteWorkStatus(ctx context.Context, workStatusIdentifier util.ObjectIdentifier,
	workStatusContent *runtime.RawExtension) bool {
	return false // TODO
}

// NoteStatusCollector notes a statuscollector's spec.
// The statuscollector is cached on the resolver's level, and is updated
// for every resolution it is involved with.
// The returned array contains the identifiers of workload objects that
// should have their associated workstatuses processed for evaluation.
// The statuscollector is assumed to be valid.
func (c *combinedStatusResolver) NoteStatusCollector(statusCollector *v1alpha1.StatusCollector) bool {
	return false // TODO
}

// ResolutionExists returns true if a combinedstatus resolution exists for
// the given binding name and workload object identifier.
func (c *combinedStatusResolver) ResolutionExists(bindingName string, objectIdentifier util.ObjectIdentifier) bool {
	c.RLock()
	defer c.RUnlock()

	if resolutions, exists := c.bindingNameToResolutions[bindingName]; exists {
		_, exists = resolutions[objectIdentifier]
		return exists
	}

	return false
}

func (scc *statusCollectorClauses) equals(other *statusCollectorClauses) bool {
	if scc.limit != other.limit {
		return false
	}

	if scc.filter != other.filter {
		return false
	}

	// check clauses lengths
	if len(scc.groupBy) != len(other.groupBy) ||
		len(scc.combinedFields) != len(other.combinedFields) ||
		len(scc.selectClause) != len(other.selectClause) {
		return false
	}
	// compare contents: all names of expressions must common, and their expressions must be equal
	// select-clause first since groupBy and combinedFields would be empty if select is not
	selectClauseMap := abstract.SliceToMap(scc.selectClause,
		func(ne v1alpha1.NamedExpression) string { return ne.Name },
		func(ne v1alpha1.NamedExpression) v1alpha1.Expression { return ne.Def })
	for _, ne := range other.selectClause {
		if expr, ok := selectClauseMap[ne.Name]; !ok || expr != ne.Def {
			return false
		}
	}

	groupByMap := abstract.SliceToMap(scc.groupBy, func(ne v1alpha1.NamedExpression) string { return ne.Name },
		func(ne v1alpha1.NamedExpression) v1alpha1.Expression { return ne.Def })
	for _, ne := range other.groupBy {
		if expr, ok := groupByMap[ne.Name]; !ok || expr != ne.Def {
			return false
		}
	}

	combinedFieldsMap := abstract.SliceToMap(scc.combinedFields,
		func(na v1alpha1.NamedAggregator) string { return na.Name },
		func(na v1alpha1.NamedAggregator) v1alpha1.NamedAggregator { return na })
	for _, na := range other.combinedFields {
		if aggregator, ok := combinedFieldsMap[na.Name]; !ok ||
			aggregator.Type != na.Type || aggregator.Subject != na.Subject {
			return false
		}
	}

	return true
}

func (scc *statusCollectorClauses) clone() *statusCollectorClauses {
	return &statusCollectorClauses{
		filter:         scc.filter,
		groupBy:        abstract.SliceCopy(scc.groupBy),
		combinedFields: abstract.SliceCopy(scc.combinedFields),
		selectClause:   abstract.SliceCopy(scc.selectClause),
		limit:          scc.limit,
	}
}
