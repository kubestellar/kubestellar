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

package placement

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubestellar/kubestellar/api/edge/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// PlacementDecisionResolver abstracts the functionality needed from a placement decision resolver,
// to simplify its usage and keep implementation details inaccessible.
type PlacementDecisionResolver interface {
	// GetPlacementDecision returns the placement decision for the given
	// placement key. If a mapping to the key does not exist, it is created.
	GetPlacementDecision(placementKey types.NamespacedName) (*v1alpha1.PlacementDecisionSpec, error)
	// ComparePlacementDecision compares the given placement decision spec
	// with the maintained placement decision for the given placement key.
	ComparePlacementDecision(placementKey types.NamespacedName,
		placementDecisionSpec *v1alpha1.PlacementDecisionSpec) bool
	// UpdateDecisionDataResources updates the maintained placement decision's
	// resources data for the given placement key.
	// If a mapping to the key does not exist, it is created.
	// The return bool indicates whether the placement decision resolution was changed.
	UpdateDecisionDataResources(placementKey types.NamespacedName, obj runtime.Object) (bool, error)
	// UpdateDecisionDataDestinations updates the maintained placement
	// decision's destinations data for the given placement key.
	// If a mapping to the key does not exist, it is created.
	UpdateDecisionDataDestinations(placementKey types.NamespacedName, destinations []string)
	// DeleteDecisionData deletes the maintained placement decision data for
	// the given placement key.
	DeleteDecisionData(placementKey types.NamespacedName)
}

func NewPlacementDecisionResolver(gvkGvrMapper GvkGvrMapper) PlacementDecisionResolver {
	return &placementDecisionResolver{
		gvkGvrMapper: gvkGvrMapper,
		decisions:    make(map[types.NamespacedName]*decisionData),
	}
}

type placementDecisionResolver struct {
	sync.RWMutex
	gvkGvrMapper GvkGvrMapper
	decisions    map[types.NamespacedName]*decisionData
}

// GetPlacementDecision returns the placement decision for the given
// placement key. If a mapping to the key does not exist, it is created.
func (pdr *placementDecisionResolver) GetPlacementDecision(placementKey types.NamespacedName) (
	*v1alpha1.PlacementDecisionSpec, error) {
	pdr.RLock() // lock for reading map
	decisionData, exists := pdr.decisions[placementKey]
	pdr.RUnlock() // unlock after accessing map

	if !exists {
		decisionData = pdr.createDecision(placementKey)
	}

	placementDecisionSpec, err := decisionData.toPlacementDecisionSpec(pdr.gvkGvrMapper)

	if err != nil {
		return nil, fmt.Errorf("failed to create placement-decision-spec for placement %v: %v",
			placementKey, err)
	}

	return placementDecisionSpec, nil
}

// ComparePlacementDecision compares the given placement decision spec
// with the maintained placement decision for the given placement key.
func (pdr *placementDecisionResolver) ComparePlacementDecision(placementKey types.NamespacedName,
	placementDecisionSpec *v1alpha1.PlacementDecisionSpec) bool {
	pdr.RLock() // lock for reading map
	decisionData, exists := pdr.decisions[placementKey]
	pdr.RUnlock() // unlock after accessing map

	if !exists {
		return false
	}

	return decisionData.matchesPlacementDecisionSpec(placementDecisionSpec, pdr.gvkGvrMapper)
}

// UpdateDecisionDataResources updates the maintained placement decision's
// resources data for the given placement key.
// If a mapping to the key does not exist, it is created.
// The return bool indicates whether the placement decision resolution was changed.
func (pdr *placementDecisionResolver) UpdateDecisionDataResources(placementKey types.NamespacedName,
	obj runtime.Object) (bool, error) {
	pdr.RLock() // lock for reading map
	decisionData, exists := pdr.decisions[placementKey]
	pdr.RUnlock() // unlock after accessing map

	if !exists {
		decisionData = pdr.createDecision(placementKey)
	}

	// updateObject is thread-safe
	changed, err := decisionData.updateObject(obj)
	if err != nil {
		return false, fmt.Errorf("failed to update decision data for placement %v: %v", placementKey, err)
	}

	return changed, nil
}

// UpdateDecisionDataDestinations updates the maintained placement
// decision's destinations data for the given placement key.
// If a mapping to the key does not exist, it is created.
func (pdr *placementDecisionResolver) UpdateDecisionDataDestinations(placementKey types.NamespacedName,
	destinations []string) {
	pdr.RLock() // lock for reading map
	decisionData, exists := pdr.decisions[placementKey]
	pdr.RUnlock() // unlock after accessing map

	if !exists {
		decisionData = pdr.createDecision(placementKey)
	}

	decisionData.updateDestinations(destinations)
}

// DeleteDecisionData deletes the maintained placement decision data for the given key.
func (pdr *placementDecisionResolver) DeleteDecisionData(placementKey types.NamespacedName) {
	pdr.Lock() // lock for modifying map
	defer pdr.Unlock()

	delete(pdr.decisions, placementKey)
}

func (pdr *placementDecisionResolver) createDecision(placementKey types.NamespacedName) *decisionData {
	pdr.Lock() // lock for modifying map
	defer pdr.Unlock()

	// double-check existence to handle race conditions (common pattern)
	if decisionData, exists := pdr.decisions[placementKey]; exists {
		return decisionData
	}

	decisionData := &decisionData{
		placementKey:  placementKey,
		mappedObjects: make(map[string]*util.Key),
		destinations:  []string{},
	}
	pdr.decisions[placementKey] = decisionData

	return decisionData
}
