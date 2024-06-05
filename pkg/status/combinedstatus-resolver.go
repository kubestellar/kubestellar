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
	"fmt"
	"sync"

	runtime2 "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/abstract"
	"github.com/kubestellar/kubestellar/pkg/binding"
	controllisters "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

type CombinedStatusResolver interface {
	// GenerateCombinedStatus generates a CombinedStatus object for the given
	// binding name and workload object identifier.
	// If no resolution is associated with the given combination, nil is returned.
	GenerateCombinedStatus(bindingName string, objectIdentifier util.ObjectIdentifier) *v1alpha1.CombinedStatus

	// CompareCombinedStatus compares the given CombinedStatus object with the
	// one associated with the given binding name and workload object identifier.
	// True is returned in case of a match, false otherwise.
	CompareCombinedStatus(bindingName string, objectIdentifier util.ObjectIdentifier,
		combinedStatus *v1alpha1.CombinedStatus) bool

	// NoteBindingResolution notes a binding resolution for status collection.
	//
	// 1. If `deleted` is true, the associated combinedstatus resolutions are
	// removed from memory. The same is done if a resolution no longer requires
	// status collection.
	//
	// 2. Excessive combinedstatus resolutions are removed if they are no
	// longer associated with the binding.
	//
	// 3. For every workload object associated with one or more
	// statuscollectors, a combinedstatus resolution is created/updated.
	// The update may involve adding or removing statuscollectors, and changing
	// the set of destinations associated with the binding.
	//
	// The function uses the statuscollector lister and the cached workstatus
	// objects to update internal state. The given array is assumed to contain
	// all objects that may be relevant to the binding.
	//
	// The returned set contains the identifiers of combinedstatus objects
	// that should be queued for syncing.
	NoteBindingResolution(bindingName string, bindingResolution binding.Resolution, deleted bool,
		workStatuses []*workStatus,
		statusCollectorLister controllisters.StatusCollectorLister) sets.Set[util.ObjectIdentifier]

	// NoteStatusCollector notes a statuscollector's spec.
	// The statuscollector is cached on the resolver's level, and is updated
	// for every resolution it is involved with. The statuscollector is assumed
	// to be valid and immutable.
	//
	// The given array is assumed to contain all objects that may be associated
	// with any combinedstatus resolution.
	// The array and its contents must not be mutated during this call.
	//
	// The returned set contains the identifiers of combinedstatus objects
	// that should be queued for syncing.
	NoteStatusCollector(statusCollector *v1alpha1.StatusCollector,
		workStatuses []*workStatus) sets.Set[util.ObjectIdentifier]

	// NoteWorkStatus notes a workstatus in the combinedstatus resolution
	// associated with its source workload object.
	//
	// The returned set contains the identifiers of combinedstatus objects
	// that should be queued for syncing.
	NoteWorkStatus(workStatus *workStatus) sets.Set[util.ObjectIdentifier]

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

// GenerateCombinedStatus generates a CombinedStatus object for the given
// binding name and workload object identifier.
// If no resolution is associated with the given combination, nil is returned.
func (c *combinedStatusResolver) GenerateCombinedStatus(bindingName string,
	objectIdentifier util.ObjectIdentifier) *v1alpha1.CombinedStatus {
	c.RLock()
	defer c.RUnlock()

	if resolutions, exists := c.bindingNameToResolutions[bindingName]; exists {
		if resolution, exists := resolutions[objectIdentifier]; exists {
			return resolution.generateCombinedStatus(bindingName, objectIdentifier)
		}
	}

	return nil
}

// CompareCombinedStatus compares the given CombinedStatus object with the
// one associated with the given binding name and workload object identifier.
// True is returned in case of a match, false otherwise.
func (c *combinedStatusResolver) CompareCombinedStatus(bindingName string,
	objectIdentifier util.ObjectIdentifier, combinedStatus *v1alpha1.CombinedStatus) bool {
	c.RLock()
	defer c.RUnlock()

	if resolutions, exists := c.bindingNameToResolutions[bindingName]; exists {
		if resolution, exists := resolutions[objectIdentifier]; exists {
			return resolution.compareCombinedStatus(combinedStatus)
		}
	}

	return false
}

// NoteBindingResolution notes a binding resolution for status collection.
//
// 1. If `deleted` is true, the associated combinedstatus resolutions are
// removed from memory. The same is done if a resolution no longer requires
// status collection.
//
// 2. Excessive combinedstatus resolutions are removed if they are no
// longer associated with the binding.
//
// 3. For every workload object associated with one or more
// statuscollectors, a combinedstatus resolution is created/updated.
// The update may involve adding or removing statuscollectors, and changing
// the set of destinations associated with the binding.
//
// The function uses the statuscollector lister and the cached workstatus
// objects to update internal state. The given array is assumed to contain
// all objects that may be relevant to the binding.
//
// The returned set contains the identifiers of combinedstatus objects
// that should be queued for syncing.
func (c *combinedStatusResolver) NoteBindingResolution(bindingName string, bindingResolution binding.Resolution,
	deleted bool, workStatuses []*workStatus,
	statusCollectorLister controllisters.StatusCollectorLister) sets.Set[util.ObjectIdentifier] {
	c.Lock()
	defer c.Unlock()

	combinedStatusIdentifiersToQueue := sets.New[util.ObjectIdentifier]()
	workloadIdentifiersToEvaluate := sets.New[util.ObjectIdentifier]()
	// destinations set to collect from, shared for all workload objects
	collectionDestinations := sets.New(abstract.SliceMap(bindingResolution.Destinations,
		func(destination v1alpha1.Destination) string { return destination.ClusterId })...)

	// (1)
	if deleted {
		// for every combinedstatus resolution - queue its object identifier for syncing
		// and remove the resolution from memory
		for workloadObjectIdentifier, _ := range c.bindingNameToResolutions[bindingName] {
			combinedStatusIdentifiersToQueue.Insert(getCombinedStatusIdentifier(bindingName, workloadObjectIdentifier))
		}

		delete(c.bindingNameToResolutions, bindingName)
		return combinedStatusIdentifiersToQueue
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
			csResolution = &combinedStatusResolution{
				statusCollectorNameToData: make(map[string]*statusCollectorData),
			}
			objectIdentifierToResolutions[objectIdentifier] = csResolution
		}

		// fetch missing statuscollector specs
		c.fetchMissingStatusCollectorSpecsLocked(statusCollectorLister, objectData.StatusCollectors)

		// update statuscollectors
		removedCollectors, addedCollectors := csResolution.setStatusCollectors(
			abstract.SliceToPrimitiveMap(objectData.StatusCollectors,
				func(statusCollectorName string) string { return statusCollectorName }, // keys mapper
				func(statusCollectorName string) *v1alpha1.StatusCollectorSpec { // val mapper
					return c.statusCollectorNameToSpec[statusCollectorName] // if missing after fetch, ignored
				}))

		// update destinations
		removedDestinations, newDestinationsSet := csResolution.setCollectionDestinations(collectionDestinations)

		// should queue the combinedstatus object for syncing if lost collectors / destinations
		if removedCollectors || removedDestinations {
			combinedStatusIdentifiersToQueue.Insert(getCombinedStatusIdentifier(bindingName, objectIdentifier))
		}

		// should evaluate workstatuses if added/updated collectors or added destinations
		// it's possible to evaluate workloads coming from new destinations only if no collectors were removed (TODO)
		if addedCollectors || len(newDestinationsSet) > 0 {
			workloadIdentifiersToEvaluate.Insert(objectIdentifier)
		}
	}

	// evaluate workstatuses associated with members of workloadIdentifiersToEvaluate and return the combinedstatus
	// identifiers that should be queued for syncing
	return combinedStatusIdentifiersToQueue.Union(c.evaluateWorkStatusesPerBindingLocked(bindingName,
		workloadIdentifiersToEvaluate, collectionDestinations, workStatuses))
}

// NoteWorkStatus notes a workstatus in the combinedstatus resolution
// associated with its source workload object.
//
// The returned set contains the identifiers of combinedstatus objects
// that should be queued for syncing.
func (c *combinedStatusResolver) NoteWorkStatus(workStatus *workStatus) sets.Set[util.ObjectIdentifier] {
	c.RLock()
	defer c.RUnlock()

	combinedStatusIdentifiersToQueue := sets.New[util.ObjectIdentifier]()
	bindingNameToError := make(map[string]error) // TODO: use

	// update resolutions sensitive to the workstatus
	for bindingName, resolutions := range c.bindingNameToResolutions {
		resolution, exists := resolutions[workStatus.sourceObjectIdentifier]
		if !exists {
			continue
		}

		changed, err := resolution.evaluateWorkStatus(c.celEvaluator, workStatus.wecName, workStatus.status)
		if err != nil { //
			bindingNameToError[bindingName] = err
		}

		if changed {
			combinedStatusIdentifiersToQueue.Insert(getCombinedStatusIdentifier(bindingName,
				workStatus.sourceObjectIdentifier))
		}
	}

	return combinedStatusIdentifiersToQueue
}

// NoteStatusCollector notes a statuscollector's spec.
// The statuscollector is cached on the resolver's level, and is updated
// for every resolution it is involved with. The statuscollector is assumed
// to be valid and immutable.
//
// The given array is assumed to contain all objects that may be associated
// with any combinedstatus resolution.
// The array and its contents must not be mutated during this call.
//
// The returned set contains the identifiers of combinedstatus objects
// that should be queued for syncing.
func (c *combinedStatusResolver) NoteStatusCollector(statusCollector *v1alpha1.StatusCollector,
	workStatuses []*workStatus) sets.Set[util.ObjectIdentifier] {
	c.Lock()
	defer c.Unlock()

	currentSpec := c.statusCollectorNameToSpec[statusCollector.Name]
	if currentSpec != nil && statusCollectorSpecsMatch(currentSpec, &statusCollector.Spec) {
		return nil // already cached and the spec has not changed
	}

	// update the cache
	c.statusCollectorNameToSpec[statusCollector.Name] = &statusCollector.Spec // readonly

	combinedStatusIdentifiersToQueue := sets.New[util.ObjectIdentifier]()
	// update resolutions that use the statuscollector
	// this call cannot add an association that was not already present
	for bindingName, resolutions := range c.bindingNameToResolutions {
		for workloadObjectIdentifier, resolution := range resolutions {
			if resolution.updateStatusCollector(statusCollector.Name, &statusCollector.Spec) { // true if changed
				// evaluate ALL workstatuses associated with the (binding, workload object) pair
				combinedStatusIdentifiersToQueue.Insert(c.evaluateWorkStatusesPerBindingLocked(bindingName,
					sets.New(workloadObjectIdentifier), resolution.collectionDestinations,
					workStatuses).UnsortedList()...)
			}
		}
	}

	return combinedStatusIdentifiersToQueue
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

// fetchMissingStatusCollectorSpecs fetches the missing statuscollector specs
// from the given lister and updates the cache.
// The method is expected to be called with the write lock held.
func (c *combinedStatusResolver) fetchMissingStatusCollectorSpecsLocked(statusCollectorLister controllisters.StatusCollectorLister,
	statusCollectorNames []string) {
	for _, statusCollectorName := range statusCollectorNames {
		if _, exists := c.statusCollectorNameToSpec[statusCollectorName]; exists {
			continue // this method is not responsible for keeping the cache up-to-date
		}

		statusCollector, err := statusCollectorLister.Get(statusCollectorName)
		if err != nil {
			// fetch error should not disturb the flow.
			// a missing spec will be reconciled when the status collector is created/updated.
			runtime2.HandleError(fmt.Errorf("failed to get statuscollector %s: %w", statusCollectorName, err))
		}

		c.statusCollectorNameToSpec[statusCollectorName] = &statusCollector.Spec // readonly
	}
}

// evaluateWorkStatusesPerBindingLocked evaluates workstatuses associated with the given
// workload identifiers and destinations. Workstatuses that are not associated
// with the two are skipped.
// The returned set contains the identifiers of combinedstatus objects that
// should be queued for syncing.
// The method is expected to be called with the read lock held.
func (c *combinedStatusResolver) evaluateWorkStatusesPerBindingLocked(bindingName string,
	workloadIdentifiersToEvaluate sets.Set[util.ObjectIdentifier], destinations sets.Set[string],
	workStatuses []*workStatus) sets.Set[util.ObjectIdentifier] {
	workloadObjectToError := make(map[util.ObjectIdentifier]error) // TODO: use
	combinedStatusesToQueue := sets.Set[util.ObjectIdentifier]{}

	for _, workStatus := range workStatuses {
		// skip workstatuses that are not associated with the binding's destinations
		if destinations.Has(workStatus.wecName) {
			continue
		}

		// skip workstatuses that are not associated with the workload objects
		if !workloadIdentifiersToEvaluate.Has(workStatus.sourceObjectIdentifier) {
			continue
		}

		// evaluate workstatus
		csResolution := c.bindingNameToResolutions[bindingName][workStatus.sourceObjectIdentifier]
		changed, err := csResolution.evaluateWorkStatus(c.celEvaluator, workStatus.wecName, workStatus.status)
		if err != nil {
			workloadObjectToError[workStatus.sourceObjectIdentifier] = err
		}

		if changed {
			combinedStatusesToQueue.Insert(getCombinedStatusIdentifier(bindingName, workStatus.sourceObjectIdentifier))
		}
	}

	return combinedStatusesToQueue
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
