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
	//
	// This function returns:
	//   - The identifiers of combinedstatus objects that should be queued for
	//     syncing.
	//   - The identifiers of workload objects that should have their
	//     associated workstatuses processed for evaluation.
	//   - The names of statuscollectors that should be fetched.
	NoteBindingResolution(bindingName string, bindingResolution binding.Resolution,
		deleted bool) (sets.Set[util.ObjectIdentifier], sets.Set[util.ObjectIdentifier], sets.Set[string])

	// NoteStatusCollector notes a statuscollector's spec.
	// The statuscollector is cached on the resolver's level, and is updated
	// for every resolution it is involved with.
	// The returned set contains the identifiers of workload objects that
	// should have their associated workstatuses processed for evaluation.
	// The statuscollector is assumed to be valid.
	NoteStatusCollector(statusCollector *v1alpha1.StatusCollector) sets.Set[util.ObjectIdentifier]

	// NoteWorkStatus notes a workstatus's content in the combinedstatus
	// resolution associated with its source workload object.
	// The returned boolean indicates whether the combinedstatus resolution
	// associated with the workstatus was updated.
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
	// statusCollectorNameToSpec is a map of statuscollector names to their
	// specs. This serves as a cache that is the source of truth for
	// statuscollectors that are used in the combinedstatus resolutions.
	// Users of this map are expected not to mutate mapped values.
	statusCollectorNameToSpec map[string]*v1alpha1.StatusCollectorSpec
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
//
// This function returns:
//   - The identifiers of combinedstatus objects that should be queued for
//     syncing.
//   - The identifiers of workload objects that should have their
//     associated workstatuses processed for evaluation.
//   - The names of statuscollectors that should be fetched.
func (c *combinedStatusResolver) NoteBindingResolution(bindingName string, bindingResolution binding.Resolution,
	deleted bool) (sets.Set[util.ObjectIdentifier], sets.Set[util.ObjectIdentifier], sets.Set[string]) {
	c.Lock()
	defer c.Unlock()

	combinedStatusIdentifiersToQueue := sets.Set[util.ObjectIdentifier]{}
	workloadIdentifiersToQueue := sets.Set[util.ObjectIdentifier]{}

	statusCollectorNamesToFetch := sets.Set[string]{}

	// (1)
	if deleted {
		// for every combinedstatus resolution - queue its object identifier for syncing
		// and remove the resolution from memory
		for workloadObjectIdentifier, _ := range c.bindingNameToResolutions[bindingName] {
			combinedStatusIdentifiersToQueue.Insert(getCombinedStatusIdentifier(bindingName, workloadObjectIdentifier))
		}

		delete(c.bindingNameToResolutions, bindingName)
		return combinedStatusIdentifiersToQueue, workloadIdentifiersToQueue, statusCollectorNamesToFetch
	}

	// if the binding resolution is not yet noted - create a new entry
	objectIdentifierToResolutions, exists := c.bindingNameToResolutions[bindingName]
	if !exists {
		objectIdentifierToResolutions = make(map[util.ObjectIdentifier]*combinedStatusResolution)
		c.bindingNameToResolutions[bindingName] = objectIdentifierToResolutions
	}

	// (2) remove excessive combinedstatus resolutions of objects that are no longer
	// associated with the binding resolution
	for objectIdentifier, _ := range objectIdentifierToResolutions {
		if _, exists := bindingResolution.ObjectIdentifierToData[objectIdentifier]; !exists {
			combinedStatusIdentifiersToQueue.Insert(getCombinedStatusIdentifier(bindingName, objectIdentifier))
			delete(objectIdentifierToResolutions, objectIdentifier)
		}
	}

	// (~2+3) create/update combinedstatus resolutions for every object that requires status collection,
	// and delete resolutions that are no longer required
	for objectIdentifier, objectData := range bindingResolution.ObjectIdentifierToData {
		csResolution, exists := objectIdentifierToResolutions[objectIdentifier]
		if len(objectData.StatusCollectors) == 0 {
			if exists { // associated resolution is no longer required
				combinedStatusIdentifiersToQueue.Insert(getCombinedStatusIdentifier(bindingName, objectIdentifier))
				delete(objectIdentifierToResolutions, objectIdentifier)
			}

			continue
		}

		// create resolution entry if missing
		if !exists {
			csResolution = &combinedStatusResolution{statusCollectorNameToData: make(map[string]*statusCollectorData)}
			objectIdentifierToResolutions[objectIdentifier] = csResolution
			// no need to queue the combinedstatus object for syncing since if it will contain
			// information, it will be queued up after the calculations phase
		}

		// update statuscollectors
		removedSome, addedSome := csResolution.setStatusCollectors(
			abstract.SliceToPrimitiveMap(objectData.StatusCollectors,
				func(statusCollectorName string) string { return statusCollectorName }, // keys mapper
				func(statusCollectorName string) *v1alpha1.StatusCollectorSpec { // val mapper
					if statusCollectorSpec, exists := c.statusCollectorNameToSpec[statusCollectorName]; exists {
						return statusCollectorSpec
					}
					// not cached - need to fetch
					statusCollectorNamesToFetch.Insert(statusCollectorName)
					return nil
				}))

		// queue the combinedstatus object for syncing if lost collectors
		if removedSome {
			combinedStatusIdentifiersToQueue.Insert(getCombinedStatusIdentifier(bindingName, objectIdentifier))
		}

		// queue the source object identifier to have its workstatuses evaluated
		if addedSome {
			workloadIdentifiersToQueue.Insert(objectIdentifier)
		}
	}

	return combinedStatusIdentifiersToQueue, workloadIdentifiersToQueue, statusCollectorNamesToFetch
}

// NoteWorkStatus notes a workstatus's content in the combinedstatus
// resolution associated with its source workload object.
// The returned boolean indicates whether the combinedstatus resolution
// associated with the workstatus was updated.
func (c *combinedStatusResolver) NoteWorkStatus(ctx context.Context, workStatusIdentifier util.ObjectIdentifier,
	workStatusContent *runtime.RawExtension) bool {
	return false // TODO
}

// NoteStatusCollector notes a statuscollector's spec.
// The statuscollector is cached on the resolver's level, and is updated
// for every resolution it is involved with.
// The returned set contains the identifiers of workload objects that
// should have their associated workstatuses processed for evaluation.
// The statuscollector is assumed to be valid.
func (c *combinedStatusResolver) NoteStatusCollector(statusCollector *v1alpha1.StatusCollector) sets.Set[util.ObjectIdentifier] {
	return nil // TODO
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

func statusCollectorSpecsMatch(spec1, spec2 *v1alpha1.StatusCollectorSpec) bool {
	if spec1.Limit != spec2.Limit {
		return false
	}

	if spec1.Filter != spec2.Filter {
		return false
	}

	// check clauses lengths
	if len(spec1.GroupBy) != len(spec2.GroupBy) ||
		len(spec1.CombinedFields) != len(spec2.CombinedFields) ||
		len(spec1.Select) != len(spec2.Select) {
		return false
	}
	// compare contents: all names of expressions must common, and their expressions must be equal.
	// select-clause first since groupBy and combinedFields would be empty if select is not
	selectClauseMap := namedExpressionSliceToMap(spec1.Select)
	for _, ne := range spec2.Select {
		if expr, ok := selectClauseMap[ne.Name]; !ok || expr != ne.Def {
			return false
		}
	}

	groupByMap := namedExpressionSliceToMap(spec1.GroupBy)
	for _, ne := range spec2.GroupBy {
		if expr, ok := groupByMap[ne.Name]; !ok || expr != ne.Def {
			return false
		}
	}

	combinedFieldsMap := abstract.SliceToPrimitiveMap(spec1.CombinedFields,
		func(na v1alpha1.NamedAggregator) string { return na.Name },
		func(na v1alpha1.NamedAggregator) v1alpha1.NamedAggregator { return na })
	for _, na := range spec2.CombinedFields {
		if aggregator, ok := combinedFieldsMap[na.Name]; !ok ||
			aggregator.Type != na.Type || aggregator.Subject != na.Subject {
			return false
		}
	}

	return true
}

// namedExpressionSliceToMap converts a slice of NamedExpressions to a map,
// where the key is the name of the expression and the value is the expression
// itself.
func namedExpressionSliceToMap(slice []v1alpha1.NamedExpression) map[string]v1alpha1.Expression {
	result := make(map[string]v1alpha1.Expression, len(slice))
	for _, ne := range slice {
		result[ne.Name] = ne.Def
	}

	return result
}
