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
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const bindingPolicyResolutionNotFoundErrorPrefix = "bindingpolicy resolution is not found"

// A BindingPolicyResolver holds a collection of bindingpolicy resolutions.
// The collection is indexed by bindingPolicyKey strings, which are the names of
// the bindingpolicy objects. The resolution for a given key can be updated,
// exported and compared to the binding representation.
// All functions in this interface are thread-safe, and nothing mutates any
// method-parameter during a call to one of them.
type BindingPolicyResolver interface {
	// GenerateBinding returns the binding for the given
	// bindingpolicy key.
	//
	// If no resolution is associated with the given key, nil is returned.
	GenerateBinding(bindingPolicyKey string) *v1alpha1.BindingSpec
	// GetOwnerReference returns the owner reference for the given
	// bindingpolicy key. If no resolution is associated with the given key, an
	// error is returned.
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

	// EnsureObjectIdentifierWithVersion ensures that an object's identifier is
	// in the resolution for the given bindingpolicy key, and is associated
	// with the given resourceVersion.
	//
	// The returned bool indicates whether the bindingpolicy resolution was
	// changed. If no resolution is associated with the given key, an error is
	// returned.
	EnsureObjectIdentifierWithVersion(bindingPolicyKey string, objIdentifier util.ObjectIdentifier,
		resourceVersion string) (bool, error)
	// RemoveObjectIdentifier ensures the absence of the given object
	// identifier from the resolution for the given bindingpolicy key.
	//
	// The returned bool indicates whether the bindingpolicy resolution was
	// changed. If no resolution is associated with the given key, false is
	// returned.
	RemoveObjectIdentifier(bindingPolicyKey string, objIdentifier util.ObjectIdentifier) bool
	// GetObjectIdentifiers returns the object identifiers associated with the
	// given bindingpolicy key.
	// If no resolution is associated with the given key, an error is returned.
	GetObjectIdentifiers(bindingPolicyKey string) (sets.Set[util.ObjectIdentifier], error)

	// SetDestinations updates the maintained bindingpolicy's
	// destinations resolution for the given bindingpolicy key.
	// The given destinations set is expected not to be mutated during and
	// after this call by the caller.
	// If no resolution is associated with the given key, an error is returned.
	SetDestinations(bindingPolicyKey string, destinations sets.Set[string]) error

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

func NewBindingPolicyResolver() BindingPolicyResolver {
	return &bindingPolicyResolver{
		bindingPolicyToResolution: make(map[string]*bindingPolicyResolution),
	}
}

type bindingPolicyResolver struct {
	sync.RWMutex
	bindingPolicyToResolution map[string]*bindingPolicyResolution
}

// GenerateBinding returns the binding for the given
// bindingpolicy key.
//
// If no resolution is associated with the given key, nil is returned.
func (resolver *bindingPolicyResolver) GenerateBinding(bindingPolicyKey string) *v1alpha1.BindingSpec {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return nil
	}

	// thread-safe
	return bindingPolicyResolution.toBindingSpec()
}

// GetOwnerReference returns the owner reference for the given
// bindingpolicy key. If no resolution is associated with the given key, an
// error is returned.
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

	return bindingPolicyResolution.matchesBindingSpec(bindingSpec)
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

// EnsureObjectIdentifierWithVersion ensures that an object's identifier is
// in the resolution for the given bindingpolicy key, and is associated
// with the given resourceVersion.
//
// The returned bool indicates whether the bindingpolicy resolution was
// changed. If no resolution is associated with the given key, an error is
// returned.
func (resolver *bindingPolicyResolver) EnsureObjectIdentifierWithVersion(bindingPolicyKey string,
	objIdentifier util.ObjectIdentifier, resourceVersion string) (bool, error) {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		// bindingPolicyKey is not associated with any resolution
		return false, fmt.Errorf("%s - bindingpolicy-key: %s", bindingPolicyResolutionNotFoundErrorPrefix,
			bindingPolicyKey)
	}

	// ensureObjectIdentifier is thread-safe
	return bindingPolicyResolution.ensureObjectIdentifierWithVersion(objIdentifier, resourceVersion), nil
}

// RemoveObjectIdentifier ensures the absence of the given object
// identifier from the resolution for the given bindingpolicy key.
//
// The returned bool indicates whether the bindingpolicy resolution was
// changed. If no resolution is associated with the given key, false is
// returned.
func (resolver *bindingPolicyResolver) RemoveObjectIdentifier(bindingPolicyKey string,
	objIdentifier util.ObjectIdentifier) bool {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return false
	}

	// removeObjectIdentifier is thread-safe
	return bindingPolicyResolution.removeObjectIdentifier(objIdentifier)
}

// GetObjectIdentifiers returns a copy of the object identifiers associated
// with the given bindingpolicy key.
// If no resolution is associated with the given key, an error is returned.
func (resolver *bindingPolicyResolver) GetObjectIdentifiers(bindingPolicyKey string) (sets.Set[util.ObjectIdentifier],
	error) {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return nil, fmt.Errorf("%s - bindingpolicy-key: %s", bindingPolicyResolutionNotFoundErrorPrefix,
			bindingPolicyKey)
	}

	// getObjectIdentifiers is thread-safe
	return bindingPolicyResolution.getObjectIdentifiers(), nil
}

// SetDestinations updates the maintained bindingpolicy's
// destinations resolution for the given bindingpolicy key.
// The given destinations set is expected not to be mutated during and
// after this call by the caller.
// If no resolution is associated with the given key, an error is returned.
func (resolver *bindingPolicyResolver) SetDestinations(bindingPolicyKey string,
	destinations sets.Set[string]) error {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		return fmt.Errorf("%s - bindingpolicy-key: %s", bindingPolicyResolutionNotFoundErrorPrefix,
			bindingPolicyKey)
	}

	bindingPolicyResolution.setDestinations(destinations)
	return nil
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
		objectIdentifierToResourceVersion: make(map[util.ObjectIdentifier]string),
		destinations:                      sets.New[string](),
		ownerReference:                    ownerReference,
		requiresSingletonReportedState:    bindingpolicy.Spec.WantSingletonReportedState,
	}
	resolver.bindingPolicyToResolution[bindingpolicy.GetName()] = bindingPolicyResolution

	return bindingPolicyResolution
}

func errorIsBindingPolicyResolutionNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), bindingPolicyResolutionNotFoundErrorPrefix)
}
