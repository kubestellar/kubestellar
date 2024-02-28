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

package binding

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

const bindingPolicyResolutionNotFoundErrorPrefix = "bindingpolicy resolution is not found"

// A BindingPolicyResolver holds a collection of bindingpolicy resolutions.
// The collection is indexed by bindingPolicyKey string, the resolver does not
// care what the strings are. The resolution for a given key can be updated,
// exported and compared to the Binding representation.
// All functions in this interface are thread-safe, and nothing mutates any
// method-parameter during a call to one of them.
type BindingPolicyResolver interface {
	// GenerateBinding returns the binding for the given
	// bindingpolicy key. This function can fail due to internal caches temporarily being
	// out of sync.
	//
	// If no resolution is associated with the given key, an error is returned.
	GenerateBinding(bindingPolicyKey string) (*v1alpha1.BindingSpec, error)
	// GetOwnerReference returns the owner reference for the given bindingpolicy key.
	// If no resolution is associated with the given key, an error is returned.
	GetOwnerReference(bindingPolicyKey string) (metav1.OwnerReference, error)
	// CompareBinding compares the given binding spec
	// with the maintained binding for the given bindingpolicy key.
	// The returned value is true only if:
	//
	// - The destinations in the BindingSpec are an exact match
	//of those in the resolution.
	//
	// - The same is true for every selected object.
	//
	// It is possible to output a false negative due to a temporary state of
	// internal caches being out of sync.
	CompareBinding(bindingPolicyKey string,
		bindingSpec *v1alpha1.BindingSpec) bool

	// NoteBindingPolicy associates a new resolution with the given
	// bindingpolicy, if none is associated. This method maintains the
	// singleton status reporting requirement in the resolution.
	NoteBindingPolicy(bindingpolicy *v1alpha1.BindingPolicy)

	// NoteObject updates the maintained bindingpolicy's objects resolution for the
	// given bindingpolicy key. If the object is being deleted, it is removed from
	// the resolution if exists.
	//
	// The returned bool indicates whether the bindingpolicy resolution was changed.
	// If no resolution is associated with the given key, an error is returned.
	NoteObject(bindingPolicyKey string, obj runtime.Object) (bool, error)
	// RemoveObject removes the given object from the maintained bindingpolicy's
	// objects resolution for the given bindingpolicy key.
	//
	// The returned bool indicates whether the bindingpolicy resolution was changed.
	RemoveObject(bindingPolicyKey string, obj runtime.Object) bool
	// SetDestinations updates the maintained bindingpolicy's
	// destinations resolution for the given bindingpolicy key.
	// The given destinations set is expected not to be mutated after this call.
	SetDestinations(bindingPolicyKey string, destinations sets.Set[string])
	// GetObjectKeys returns the objects associated with the given
	// bindingpolicy key.
	// If no resolution is associated with the given key, an error is returned.
	GetObjectKeys(bindingPolicyKey string) ([]*util.Key, error)

	// ResolutionExists returns true if a resolution is associated with the
	// given bindingpolicy key.
	ResolutionExists(bindingPolicyKey string) bool
	// ResolutionRequiresSingletonReportedState returns true if the
	// bindingpolicy associated with the given key requires a singleton
	// reported state, and it satisfies the conditions on this requirement.
	//
	// This means that if true is returned, then the singleton status reporting
	// requirement is effective.
	ResolutionRequiresSingletonReportedState(bindingPolicyKey string) bool
	// DeleteResolution deletes the resolution associated with the given key,
	// if it exists.
	DeleteResolution(bindingPolicyKey string)
}

func NewBindingPolicyResolver(gvkGvrMapper util.GvkGvrMapper) BindingPolicyResolver {
	return &bindingPolicyResolver{
		gvkGvrMapper:              gvkGvrMapper,
		bindingPolicyToResolution: make(map[string]*bindingPolicyResolution),
	}
}

type bindingPolicyResolver struct {
	sync.RWMutex
	gvkGvrMapper              util.GvkGvrMapper
	bindingPolicyToResolution map[string]*bindingPolicyResolution
}

// GenerateBinding returns the binding for the given
// bindingpolicy key. If a key is not associated to a resolution, the latter is
// created. This function can fail due to internal caches temporarily being
// out of sync.
func (resolver *bindingPolicyResolver) GenerateBinding(bindingPolicyKey string) (
	*v1alpha1.BindingSpec, error) {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return nil, fmt.Errorf("%s - bindingpolicy-key: %s", bindingPolicyResolutionNotFoundErrorPrefix,
			bindingPolicyKey)
	}

	bindingSpec, err := bindingPolicyResolution.toBindingSpec(resolver.gvkGvrMapper)

	if err != nil {
		return nil, fmt.Errorf("failed to create BindingSpec for bindingpolicy %v: %w",
			bindingPolicyKey, err)
	}

	return bindingSpec, nil
}

// GetOwnerReference returns the owner reference for the given bindingpolicy key.
// If no resolution is associated with the given key, an error is returned.
func (resolver *bindingPolicyResolver) GetOwnerReference(bindingPolicyKey string) (metav1.OwnerReference, error) {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return metav1.OwnerReference{}, fmt.Errorf("%s - bindingpolicy-key: %s",
			bindingPolicyResolutionNotFoundErrorPrefix, bindingPolicyKey)
	}

	return *bindingPolicyResolution.ownerReference, nil
}

// CompareBinding compares the given binding spec
// with the maintained binding for the given bindingpolicy key.
// The returned value is true only if:
//
// - The destinations in the BindingSpec are an exact match
// of those in the resolution.
//
// - The same is true for every selected object.
//
// It is possible to output a false negative due to a temporary state of
// internal caches being out of sync.
func (resolver *bindingPolicyResolver) CompareBinding(bindingPolicyKey string,
	bindingSpec *v1alpha1.BindingSpec) bool {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return false
	}

	return bindingPolicyResolution.matchesBindingSpec(bindingSpec, resolver.gvkGvrMapper)
}

// NoteBindingPolicy associates a new resolution with the given
// bindingpolicy, if none is associated. This method maintains the
// singleton status reporting requirement in the resolution.
func (resolver *bindingPolicyResolver) NoteBindingPolicy(bindingpolicy *v1alpha1.BindingPolicy) {
	if resolution := resolver.getResolution(bindingpolicy.GetName()); resolution != nil {
		resolution.requiresSingletonReportedState = bindingpolicy.Spec.WantSingletonReportedState
		return
	}

	resolver.createResolution(bindingpolicy)
}

// NoteObject updates the maintained bindingpolicy's objects resolution for the
// given bindingpolicy key. If the object is being deleted, it is removed from
// the resolution if exists.
//
// The returned bool indicates whether the bindingpolicy resolution was changed.
func (resolver *bindingPolicyResolver) NoteObject(bindingPolicyKey string,
	obj runtime.Object) (bool, error) {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		// bindingPolicyKey is not associated with any resolution
		return false, fmt.Errorf("%s - bindingpolicy-key: %s", bindingPolicyResolutionNotFoundErrorPrefix,
			bindingPolicyKey)
	}

	// noteObject is thread-safe
	changed, err := bindingPolicyResolution.noteObject(obj)
	if err != nil {
		return false, fmt.Errorf("failed to update resolution for bindingpolicy %v: %w", bindingPolicyKey, err)
	}

	return changed, nil
}

// RemoveObject removes the given object from the maintained bindingpolicy's
// objects resolution for the given bindingpolicy key.
//
// The returned bool indicates whether the bindingpolicy resolution was changed.
func (resolver *bindingPolicyResolver) RemoveObject(bindingPolicyKey string,
	obj runtime.Object) bool {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return false
	}

	// removeObject is thread-safe
	return bindingPolicyResolution.removeObject(obj)
}

// SetDestinations updates the maintained bindingpolicy's
// destinations resolution for the given bindingpolicy key.
// The given destinations set is expected not to be mutated after this call.
func (resolver *bindingPolicyResolver) SetDestinations(bindingPolicyKey string,
	destinations sets.Set[string]) {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return
	}

	bindingPolicyResolution.setDestinations(destinations)
}

// GetObjectKeys returns the objects associated with the given
// bindingpolicy key.
// If no resolution is associated with the given key, an error is returned.
func (resolver *bindingPolicyResolver) GetObjectKeys(bindingPolicyKey string) ([]*util.Key, error) {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return nil, fmt.Errorf("%s - bindingpolicy-key: %s", bindingPolicyResolutionNotFoundErrorPrefix,
			bindingPolicyKey)
	}

	// getObjectKeys is thread-safe
	return bindingPolicyResolution.getObjectKeys(), nil
}

// ResolutionExists returns true if a resolution is associated with the
// given bindingpolicy key.
func (resolver *bindingPolicyResolver) ResolutionExists(bindingPolicyKey string) bool {
	if resolver.getResolution(bindingPolicyKey) == nil {
		return false
	}

	return true
}

// ResolutionRequiresSingletonReportedState returns true if the
// bindingpolicy associated with the given key requires a singleton
// reported state, and it satisfies the conditions on this requirement.
//
// This means that if true is returned, then the singleton status reporting
// requirement is effective.
func (resolver *bindingPolicyResolver) ResolutionRequiresSingletonReportedState(bindingPolicyKey string) bool {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return false
	}

	return bindingPolicyResolution.requiresSingletonReportedState
}

// DeleteResolution deletes the resolution associated with the given key,
// if it exists.
func (resolver *bindingPolicyResolver) DeleteResolution(bindingPolicyKey string) {
	resolver.Lock() // lock for modifying map
	defer resolver.Unlock()

	delete(resolver.bindingPolicyToResolution, bindingPolicyKey)
}

// getResolution retrieves the resolution associated with the given key.
// If the resolution does not exist, nil is returned.
func (resolver *bindingPolicyResolver) getResolution(bindingPolicyKey string) *bindingPolicyResolution {
	resolver.RLock()         // lock for reading map
	defer resolver.RUnlock() // unlock after accessing map

	return resolver.bindingPolicyToResolution[bindingPolicyKey]
}

func (resolver *bindingPolicyResolver) createResolution(bindingpolicy *v1alpha1.BindingPolicy) *bindingPolicyResolution {
	resolver.Lock() // lock for modifying map
	defer resolver.Unlock()

	// double-check existence to handle race conditions (common pattern)
	if bindingPolicyResolution, exists := resolver.bindingPolicyToResolution[bindingpolicy.GetName()]; exists {
		return bindingPolicyResolution
	}

	ownerReference := metav1.NewControllerRef(bindingpolicy, bindingpolicy.GroupVersionKind())
	ownerReference.BlockOwnerDeletion = &[]bool{false}[0]

	bindingPolicyResolution := &bindingPolicyResolution{
		objectIdentifierToKey:          make(map[string]*util.Key),
		destinations:                   sets.New[string](),
		workloadGeneration:             1,
		ownerReference:                 ownerReference,
		requiresSingletonReportedState: bindingpolicy.Spec.WantSingletonReportedState,
	}
	resolver.bindingPolicyToResolution[bindingpolicy.GetName()] = bindingPolicyResolution

	return bindingPolicyResolution
}

func errorIsBindingPolicyResolutionNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), bindingPolicyResolutionNotFoundErrorPrefix)
}
