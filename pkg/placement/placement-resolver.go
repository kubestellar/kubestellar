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
	"strings"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const placementResolutionNotFoundErrorPrefix = "placement resolution is not found"

// A PlacementResolver holds a collection of placement resolutions.
// The collection is indexed by placementKey string, the resolver does not
// care what the strings are. The resolution for a given key can be updated,
// exported and compared to the PlacementDecision representation.
// All functions in this interface are thread-safe, and nothing mutates any
// method-parameter during a call to one of them.
type PlacementResolver interface {
	// GeneratePlacementDecision returns the placement decision for the given
	// placement key. This function can fail due to internal caches temporarily being
	// out of sync.
	//
	// If no resolution is associated with the given key, an error is returned.
	GeneratePlacementDecision(placementKey string) (*v1alpha1.PlacementDecisionSpec, error)
	// GetOwnerReference returns the owner reference for the given placement key.
	// If no resolution is associated with the given key, an error is returned.
	GetOwnerReference(placementKey string) (metav1.OwnerReference, error)
	// ComparePlacementDecision compares the given placement decision spec
	// with the maintained placement decision for the given placement key.
	// The returned value is true only if:
	//
	// - The destinations in the PlacementDecisionSpec are an exact match
	//of those in the resolution.
	//
	// - The same is true for every selected object.
	//
	// It is possible to output a false negative due to a temporary state of
	// internal caches being out of sync.
	ComparePlacementDecision(placementKey string,
		placementDecisionSpec *v1alpha1.PlacementDecisionSpec) bool

	// NotePlacement associates a new resolution with the given placement,
	// if none is associated.
	NotePlacement(placement *v1alpha1.Placement)

	// NoteObject updates the maintained placement's objects resolution for the
	// given placement key. If the object is being deleted, it is removed from
	// the resolution if exists.
	//
	// The returned bool indicates whether the placement resolution was changed.
	// If no resolution is associated with the given key, an error is returned.
	NoteObject(placementKey string, obj runtime.Object) (bool, error)
	// RemoveObject removes the given object from the maintained placement's
	// objects resolution for the given placement key.
	//
	// The returned bool indicates whether the placement resolution was changed.
	RemoveObject(placementKey string, obj runtime.Object) bool
	// SetDestinations updates the maintained placement's
	// destinations resolution for the given placement key.
	// The given destinations set is expected not to be mutated after this call.
	SetDestinations(placementKey string, destinations sets.Set[string])

	// ResolutionExists returns true if a resolution is associated with the
	// given placement key.
	ResolutionExists(placementKey string) bool
	// DeleteResolution deletes the resolution associated with the given key,
	// if it exists.
	DeleteResolution(placementKey string)
}

func NewPlacementResolver(gvkGvrMapper util.GvkGvrMapper) PlacementResolver {
	return &placementResolver{
		gvkGvrMapper:          gvkGvrMapper,
		placementToResolution: make(map[string]*placementResolution),
	}
}

type placementResolver struct {
	sync.RWMutex
	gvkGvrMapper          util.GvkGvrMapper
	placementToResolution map[string]*placementResolution
}

// GeneratePlacementDecision returns the placement decision for the given
// placement key. If a key is not associated to a resolution, the latter is
// created. This function can fail due to internal caches temporarily being
// out of sync.
func (resolver *placementResolver) GeneratePlacementDecision(placementKey string) (
	*v1alpha1.PlacementDecisionSpec, error) {
	placementResolution := resolver.getResolution(placementKey) // thread-safe

	if placementResolution == nil {
		return nil, fmt.Errorf("%s - placement-key: %s", placementResolutionNotFoundErrorPrefix, placementKey)
	}

	placementDecisionSpec, err := placementResolution.toPlacementDecisionSpec(resolver.gvkGvrMapper)

	if err != nil {
		return nil, fmt.Errorf("failed to create PlacementDecisionSpec for placement %v: %w",
			placementKey, err)
	}

	return placementDecisionSpec, nil
}

// GetOwnerReference returns the owner reference for the given placement key.
// If no resolution is associated with the given key, an error is returned.
func (resolver *placementResolver) GetOwnerReference(placementKey string) (metav1.OwnerReference, error) {
	placementResolution := resolver.getResolution(placementKey) // thread-safe

	if placementResolution == nil {
		return metav1.OwnerReference{}, fmt.Errorf("%s - placement-key: %s", placementResolutionNotFoundErrorPrefix, placementKey)
	}

	return *placementResolution.ownerReference, nil
}

// ComparePlacementDecision compares the given placement decision spec
// with the maintained placement decision for the given placement key.
// The returned value is true only if:
//
// - The destinations in the PlacementDecisionSpec are an exact match
// of those in the resolution.
//
// - The same is true for every selected object.
//
// It is possible to output a false negative due to a temporary state of
// internal caches being out of sync.
func (resolver *placementResolver) ComparePlacementDecision(placementKey string,
	placementDecisionSpec *v1alpha1.PlacementDecisionSpec) bool {
	placementResolution := resolver.getResolution(placementKey) // thread-safe

	if placementResolution == nil {
		return false
	}

	return placementResolution.matchesPlacementDecisionSpec(placementDecisionSpec, resolver.gvkGvrMapper)
}

// NotePlacement associates a new resolution with the given placement,
// if none is associated.
func (resolver *placementResolver) NotePlacement(placement *v1alpha1.Placement) {
	if resolution := resolver.getResolution(placement.GetName()); resolution != nil {
		return
	}

	resolver.createResolution(placement)
}

// NoteObject updates the maintained placement's objects resolution for the
// given placement key. If the object is being deleted, it is removed from
// the resolution if exists.
//
// The returned bool indicates whether the placement resolution was changed.
func (resolver *placementResolver) NoteObject(placementKey string,
	obj runtime.Object) (bool, error) {
	placementResolution := resolver.getResolution(placementKey) // thread-safe

	if placementResolution == nil {
		// placementKey is not associated with any resolution
		return false, fmt.Errorf("%s - placement-key: %s", placementResolutionNotFoundErrorPrefix, placementKey)
	}

	// noteObject is thread-safe
	changed, err := placementResolution.noteObject(obj)
	if err != nil {
		return false, fmt.Errorf("failed to update resolution for placement %v: %w", placementKey, err)
	}

	return changed, nil
}

// RemoveObject removes the given object from the maintained placement's
// objects resolution for the given placement key.
//
// The returned bool indicates whether the placement resolution was changed.
func (resolver *placementResolver) RemoveObject(placementKey string,
	obj runtime.Object) bool {
	placementResolution := resolver.getResolution(placementKey) // thread-safe

	if placementResolution == nil {
		return false
	}

	// removeObject is thread-safe
	return placementResolution.removeObject(obj)
}

// SetDestinations updates the maintained placement's
// destinations resolution for the given placement key.
// The given destinations set is expected not to be mutated after this call.
func (resolver *placementResolver) SetDestinations(placementKey string,
	destinations sets.Set[string]) {
	placementResolution := resolver.getResolution(placementKey) // thread-safe

	if placementResolution == nil {
		return
	}

	placementResolution.setDestinations(destinations)
}

// ResolutionExists returns true if a resolution is associated with the
// given placement key.
func (resolver *placementResolver) ResolutionExists(placementKey string) bool {
	if resolver.getResolution(placementKey) == nil {
		return false
	}

	return true
}

// DeleteResolution deletes the resolution associated with the given key,
// if it exists.
func (resolver *placementResolver) DeleteResolution(placementKey string) {
	resolver.Lock() // lock for modifying map
	defer resolver.Unlock()

	delete(resolver.placementToResolution, placementKey)
}

// getResolution retrieves the resolution associated with the given key.
// If the resolution does not exist, nil is returned.
func (resolver *placementResolver) getResolution(placementKey string) *placementResolution {
	resolver.RLock()         // lock for reading map
	defer resolver.RUnlock() // unlock after accessing map

	return resolver.placementToResolution[placementKey]
}

func (resolver *placementResolver) createResolution(placement *v1alpha1.Placement) *placementResolution {
	resolver.Lock() // lock for modifying map
	defer resolver.Unlock()

	// double-check existence to handle race conditions (common pattern)
	if placementResolution, exists := resolver.placementToResolution[placement.GetName()]; exists {
		return placementResolution
	}

	ownerReference := metav1.NewControllerRef(placement, placement.GroupVersionKind())
	ownerReference.BlockOwnerDeletion = &[]bool{false}[0]

	placementResolution := &placementResolution{
		objectIdentifierToKey: make(map[string]*util.Key),
		destinations:          sets.New[string](),
		workloadGeneration:    1,
		ownerReference:        ownerReference,
	}
	resolver.placementToResolution[placement.GetName()] = placementResolution

	return placementResolution
}

func errorIsPlacementResolutionNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), placementResolutionNotFoundErrorPrefix)
}
