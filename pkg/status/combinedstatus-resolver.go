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

	"k8s.io/apimachinery/pkg/runtime"
	runtime2 "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

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
	// 1. If `deleted` is true, the passed bindingResolution is guaranteed not to
	// be used, and the associated combinedstatus resolutions are removed from
	// memory. The latter is also done if a resolution no longer requires status
	// collection.
	//
	// 2. Excessive combinedstatus resolutions are removed if they are no
	// longer associated with the binding.
	//
	// 3. For every workload object associated with one or more
	// statuscollectors, a combinedstatus resolution is created/updated.
	// The update may involve adding or removing statuscollectors, and changing
	// the set of destinations associated with the binding.
	//
	// The function uses the workstatus-indexer and statuscollector-lister to update
	// internal state.
	//
	// The returned set contains the identifiers of combinedstatus objects
	// that should be queued for syncing.
	NoteBindingResolution(bindingName string, bindingResolution *binding.Resolution, deleted bool,
		workStatusIndexer cache.Indexer,
		statusCollectorLister controllisters.StatusCollectorLister) sets.Set[util.ObjectIdentifier]

	// NoteStatusCollector notes a statuscollector's spec.
	// The statuscollector is cached on the resolver's level, and is updated
	// for every resolution it is involved with. The statuscollector is assumed
	// to be valid and immutable.
	//
	// If `deleted` is true, only the statuscollector's name is expected to be
	// valid, and the statuscollector is removed from the cache.
	//
	// The function uses the workstatus indexer to update internal state.
	//
	// The returned set contains the identifiers of combinedstatus objects
	// that should be queued for syncing.
	NoteStatusCollector(statusCollector *v1alpha1.StatusCollector, deleted bool,
		workStatusIndexer cache.Indexer) sets.Set[util.ObjectIdentifier]

	// NoteWorkStatus notes a workstatus in the combinedstatus resolutions
	// associated with its source workload object.
	//
	// The returned set contains the identifiers of combinedstatus objects
	// that should be queued for syncing.
	NoteWorkStatus(workStatus *workStatus) sets.Set[util.ObjectIdentifier]

	// ResolutionExists returns true if a combinedstatus resolution is
	// associated with the given name. The name is expected to follow the
	// formatting specified in the API.
	// The function returns a tuple of:
	//
	// - The associated binding's name, if the resolution exists.
	//
	// - The workload object identifier, if the resolution exists.
	//
	// - A boolean indicating whether the resolution exists.
	//
	// The returned pointers are expected to be read-only.
	ResolutionExists(name string) (*string, *util.ObjectIdentifier, bool)
}

// NewCombinedStatusResolver creates a new CombinedStatusResolver.
func NewCombinedStatusResolver(celEvaluator *celEvaluator) CombinedStatusResolver {
	return &combinedStatusResolver{
		celEvaluator:              celEvaluator,
		bindingNameToResolutions:  make(map[string]map[util.ObjectIdentifier]*combinedStatusResolution),
		resolutionNameToKey:       make(map[string]resolutionKey),
		statusCollectorNameToSpec: make(map[string]*v1alpha1.StatusCollectorSpec),
	}
}

// resolutionKey is a key used to identify a combinedstatus resolution.
// It consists of a binding name and a workload object identifier.
type resolutionKey struct {
	bindingName            string
	sourceObjectIdentifier util.ObjectIdentifier
}

type combinedStatusResolver struct {
	celEvaluator *celEvaluator

	sync.RWMutex

	// resolutions is a map of resolution keys to their
	// combinedstatus resolutions.
	bindingNameToResolutions map[string]map[util.ObjectIdentifier]*combinedStatusResolution
	// resolutionNameToKey is a map of resolution names to their keys.
	resolutionNameToKey map[string]resolutionKey
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
// 1. If `deleted` is true, the passed bindingResolution is guaranteed not to
// be used, and the associated combinedstatus resolutions are removed from
// memory. The latter is also done if a resolution no longer requires status
// collection.
//
// 2. Excessive combinedstatus resolutions are removed if they are no
// longer associated with the binding.
//
// 3. For every workload object associated with one or more
// statuscollectors, a combinedstatus resolution is created/updated.
// The update may involve adding or removing statuscollectors, and changing
// the set of destinations associated with the binding.
//
// The function uses the workstatus-indexer and statuscollector-lister to update
// internal state.
//
// The returned set contains the identifiers of combinedstatus objects
// that should be queued for syncing.
func (c *combinedStatusResolver) NoteBindingResolution(bindingName string, bindingResolution *binding.Resolution,
	deleted bool, workStatusIndexer cache.Indexer,
	statusCollectorLister controllisters.StatusCollectorLister) sets.Set[util.ObjectIdentifier] {
	c.Lock()
	defer c.Unlock()

	combinedStatusIdentifiersToQueue := sets.New[util.ObjectIdentifier]()
	workloadIdentifiersToEvaluate := sets.New[util.ObjectIdentifier]()

	// (1)
	if deleted {
		return c.deleteResolutionsForBindingWriteLocked(bindingName)
	}

	destinationsSet := sets.New(abstract.SliceMap(bindingResolution.Destinations,
		func(destination v1alpha1.Destination) string { return destination.ClusterId })...)

	// if the binding resolution is not yet noted - create a new entry
	objectIdentifierToResolution, exists := c.bindingNameToResolutions[bindingName]
	if !exists {
		objectIdentifierToResolution = make(map[util.ObjectIdentifier]*combinedStatusResolution,
			len(bindingResolution.ObjectIdentifierToData))
		c.bindingNameToResolutions[bindingName] = objectIdentifierToResolution
	}

	// (2) remove excessive combinedstatus resolutions of objects that are no longer
	// associated with the binding resolution
	for objectIdentifier, resolution := range objectIdentifierToResolution {
		if _, exists := bindingResolution.ObjectIdentifierToData[objectIdentifier]; !exists {
			combinedStatusIdentifiersToQueue.Insert(util.IdentifierForCombinedStatus(resolution.getName(),
				objectIdentifier.ObjectName.Namespace))
			delete(objectIdentifierToResolution, objectIdentifier)
			delete(c.resolutionNameToKey, resolution.getName())

			fmt.Printf("$$STATUS$$: removed resolution %s for %v\n\n", resolution.getName(), objectIdentifier)
		}
	}

	// (~2+3) create/update combinedstatus resolutions for every object that requires status collection,
	// and delete resolutions that are no longer required
	for objectIdentifier, objectData := range bindingResolution.ObjectIdentifierToData {
		fmt.Printf("$$STATUS$$: noting object: \n%v\n%v for binding resolution %s\n", objectIdentifier, objectData, bindingName)
		csResolution, exists := objectIdentifierToResolution[objectIdentifier]
		if len(objectData.StatusCollectors) == 0 {
			if exists { // associated resolution is no longer required
				combinedStatusIdentifiersToQueue.Insert(util.IdentifierForCombinedStatus(csResolution.getName(),
					objectIdentifier.ObjectName.Namespace))

				delete(objectIdentifierToResolution, objectIdentifier)
				delete(c.resolutionNameToKey, csResolution.getName())

				fmt.Printf("$$STATUS$$: removed2 resolution %s for %v\n\n", csResolution.getName(), objectIdentifier)
			}

			continue
		}

		// create resolution entry if missing
		if !exists {
			csResolution = &combinedStatusResolution{
				name:                      getCombinedStatusName(bindingResolution.UID, objectData.UID),
				statusCollectorNameToData: make(map[string]*statusCollectorData),
			}
			objectIdentifierToResolution[objectIdentifier] = csResolution
			c.resolutionNameToKey[csResolution.getName()] = resolutionKey{bindingName, objectIdentifier}

			fmt.Printf("$$STATUS$$: created resolution %s for %v, %v\n\n", csResolution.getName(), objectIdentifier, c.resolutionNameToKey)
		}

		// fetch missing statuscollector specs
		c.fetchMissingStatusCollectorSpecsLocked(statusCollectorLister, objectData.StatusCollectors)

		// update statuscollectors
		removedCollectors, addedCollectors := csResolution.setStatusCollectors(c.statusCollectorNameToSpecFromCache(objectData.StatusCollectors))

		// update destinations
		removedDestinations, newDestinationsSet := csResolution.setCollectionDestinations(destinationsSet)

		// should queue the combinedstatus object for syncing if lost collectors / destinations
		if removedCollectors || removedDestinations {
			combinedStatusIdentifiersToQueue.Insert(util.IdentifierForCombinedStatus(csResolution.getName(),
				objectIdentifier.ObjectName.Namespace))
		}

		// should evaluate workstatuses if added/updated collectors or added destinations
		if addedCollectors || len(newDestinationsSet) > 0 {
			workloadIdentifiersToEvaluate.Insert(objectIdentifier) // TODO: this can be optimized through tightening
		}
	}

	fmt.Printf("$$STATUS$$: noted binding resolution %s, updated %v combined statuses\n\n", bindingName,
		combinedStatusIdentifiersToQueue)

	// evaluate workstatuses associated with members of workloadIdentifiersToEvaluate and return the combinedstatus
	// identifiers that should be queued for syncing
	return combinedStatusIdentifiersToQueue.Union(c.evaluateWorkStatusesPerBindingReadLocked(bindingName,
		workloadIdentifiersToEvaluate, destinationsSet, workStatusIndexer))
}

// deleteResolutionsForBindingWriteLocked deletes all combinedstatus resolutions associated with the given binding name.
// The method returns the identifiers of combinedstatus objects that should be queued for syncing (deletion).
// The method is expected to be called with the write lock held.
func (c *combinedStatusResolver) deleteResolutionsForBindingWriteLocked(bindingName string) sets.Set[util.ObjectIdentifier] {
	combinedStatusIdentifiersToQueue := sets.New[util.ObjectIdentifier]()

	resolutions, exists := c.bindingNameToResolutions[bindingName]
	if !exists {
		return combinedStatusIdentifiersToQueue
	}

	for objectIdentifier, resolution := range resolutions {
		combinedStatusIdentifiersToQueue.Insert(util.IdentifierForCombinedStatus(resolution.getName(),
			objectIdentifier.ObjectName.Namespace))
		delete(c.resolutionNameToKey, resolution.getName())

		fmt.Printf("$$STATUS$$: removed3 resolution %s for %v\n\n", resolution.getName(), objectIdentifier)
	}

	delete(c.bindingNameToResolutions, bindingName)

	return combinedStatusIdentifiersToQueue
}

// NoteWorkStatus notes a workstatus in the combinedstatus resolutions
// associated with its source workload object.
//
// If the workstatus's status field is nil, the workstatus is removed from
// resolutions it affects.
//
// The returned set contains the identifiers of combinedstatus objects
// that should be queued for syncing.
// TODO: handle errors
func (c *combinedStatusResolver) NoteWorkStatus(workStatus *workStatus) sets.Set[util.ObjectIdentifier] {
	c.RLock()
	defer c.RUnlock()

	combinedStatusIdentifiersToQueue := sets.New[util.ObjectIdentifier]()

	// update resolutions sensitive to the workstatus
	for _, resolutions := range c.bindingNameToResolutions {
		resolution, exists := resolutions[workStatus.sourceObjectIdentifier]
		if !exists {
			continue
		}

		// this call logs errors, but does not return them for now
		if resolution.evaluateWorkStatus(c.celEvaluator, workStatus.wecName, workStatus.status) {
			combinedStatusIdentifiersToQueue.Insert(util.IdentifierForCombinedStatus(resolution.getName(),
				workStatus.sourceObjectIdentifier.ObjectName.Namespace))
		}
	}

	fmt.Printf("$$STATUS$$: noted workstatus %v, updated %v combined statuses\n", workStatus, combinedStatusIdentifiersToQueue)

	return combinedStatusIdentifiersToQueue
}

// NoteStatusCollector notes a statuscollector's spec.
// The statuscollector is cached on the resolver's level, and is updated
// for every resolution it is involved with. The statuscollector is assumed
// to be valid and immutable.
//
// If `deleted` is true, only the statuscollector's name is expected to be
// valid, and the statuscollector is removed from the cache.
//
// The function uses the workstatus indexer to update internal state.
//
// The returned set contains the identifiers of combinedstatus objects
// that should be queued for syncing.
func (c *combinedStatusResolver) NoteStatusCollector(statusCollector *v1alpha1.StatusCollector, deleted bool,
	workStatusIndexer cache.Indexer) sets.Set[util.ObjectIdentifier] {
	c.Lock()
	defer c.Unlock()

	currentSpec := c.statusCollectorNameToSpec[statusCollector.Name]
	if !deleted && currentSpec != nil && statusCollectorSpecsMatch(currentSpec, &statusCollector.Spec) {
		return nil // already cached and the spec has not changed
	}

	combinedStatusIdentifiersToQueue := sets.New[util.ObjectIdentifier]()
	// update resolutions that use the statuscollector
	// this call cannot add an association that was not already present.
	// if deleted, the association is removed.
	for bindingName, resolutions := range c.bindingNameToResolutions {
		for workloadObjectIdentifier, resolution := range resolutions {
			if deleted {
				if resolution.removeStatusCollector(statusCollector.Name) {
					combinedStatusIdentifiersToQueue.Insert(util.IdentifierForCombinedStatus(resolution.getName(),
						workloadObjectIdentifier.ObjectName.Namespace))
				}

				continue
			}

			if resolution.updateStatusCollector(statusCollector.Name, &statusCollector.Spec) { // true if changed
				// evaluate ALL workstatuses associated with the (binding, workload object) pair
				combinedStatusIdentifiersToQueue.Insert(c.evaluateWorkStatusesPerBindingReadLocked(bindingName,
					sets.New(workloadObjectIdentifier), resolution.collectionDestinations,
					workStatusIndexer).UnsortedList()...)
			}
		}
	}

	if !deleted {
		c.statusCollectorNameToSpec[statusCollector.Name] = &statusCollector.Spec // readonly
	} else {
		delete(c.statusCollectorNameToSpec, statusCollector.Name)
	}

	return combinedStatusIdentifiersToQueue
}

// ResolutionExists returns true if a combinedstatus resolution is
// associated with the given name. The name is expected to follow the
// formatting specified in the API.
// The function returns a tuple of:
//
// - The associated binding's name, if the resolution exists.
//
// - The workload object identifier, if the resolution exists.
//
// - A boolean indicating whether the resolution exists.
//
// The returned pointers are expected to be read-only.
func (c *combinedStatusResolver) ResolutionExists(name string) (*string, *util.ObjectIdentifier, bool) {
	c.RLock()
	defer c.RUnlock()

	key, exists := c.resolutionNameToKey[name]
	if !exists {
		return nil, nil, false
	}

	return &key.bindingName, &key.sourceObjectIdentifier, true
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
			return
		}

		c.statusCollectorNameToSpec[statusCollectorName] = &statusCollector.Spec // readonly
	}
}

// evaluateWorkStatusesPerBindingReadLocked evaluates workstatuses associated
// with the given workload identifiers and destinations.
// The returned set contains the identifiers of combinedstatus objects that
// should be queued for syncing.
// The method is expected to be called with the read lock held.
func (c *combinedStatusResolver) evaluateWorkStatusesPerBindingReadLocked(bindingName string,
	workloadObjIdentifiersToEvaluate sets.Set[util.ObjectIdentifier], destinations sets.Set[string],
	workStatusIndexer cache.Indexer) sets.Set[util.ObjectIdentifier] {
	combinedStatusesToQueue := sets.Set[util.ObjectIdentifier]{}

	for workloadObjIdentifier := range workloadObjIdentifiersToEvaluate {
		for destination := range destinations {
			// fetch workstatus
			indexKey := util.KeyFromSourceRefAndWecName(util.SourceRefFromObjectIdentifier(workloadObjIdentifier),
				destination)

			objs, err := workStatusIndexer.ByIndex(workStatusIdentificationIndexKey, indexKey) // one obj expected
			if err != nil {
				runtime2.HandleError(fmt.Errorf("failed to get workstatus with indexKey %s: %w", indexKey, err))
				continue
			}

			if len(objs) == 0 {
				continue
			}

			workStatus, err := runtimeObjectToWorkStatus(objs[0].(runtime.Object))
			if err != nil {
				runtime2.HandleError(fmt.Errorf("failed to convert runtime.Object to workStatus: %w", err))
				continue
			}

			// evaluate workstatus
			csResolution := c.bindingNameToResolutions[bindingName][workStatus.sourceObjectIdentifier]
			if csResolution.evaluateWorkStatus(c.celEvaluator, workStatus.wecName, workStatus.status) {
				combinedStatusesToQueue.Insert(util.IdentifierForCombinedStatus(csResolution.getName(),
					workloadObjIdentifier.ObjectName.Namespace))
			}
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

func (c *combinedStatusResolver) statusCollectorNameToSpecFromCache(names []string) map[string]v1alpha1.StatusCollectorSpec {
	result := make(map[string]v1alpha1.StatusCollectorSpec, len(names))
	for _, name := range names {
		spec, ok := c.statusCollectorNameToSpec[name]
		if !ok {
			continue
		}

		result[name] = *spec
	}

	return result
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
