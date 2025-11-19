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
	"k8s.io/klog/v2"

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

	// NoteBindingPolicy ensures that the resolver has an entry whose key
	// is the given BindingPolicy's name.
	// If an entry is introduced, it is introduced with empty destination set
	// and no workload references.
	// `*bindingPolicy` is immutable.
	// Concurrent calls for the same BindingPolicy name are not allowed.
	NoteBindingPolicy(bindingpolicy *v1alpha1.BindingPolicy)

	// EnsureObjectData ensures that an object's identifier is
	// in the resolution for the given bindingpolicy key, and is associated
	// with the given resource-version and DownsyncModulation.
	// The modulation's StatusCollector name set is immutable.
	//
	// The returned bool indicates whether the bindingpolicy resolution was
	// changed. If no resolution is associated with the given key, an error is
	// returned.
	EnsureObjectData(bindingPolicyKey string, objIdentifier util.ObjectIdentifier,
		objUID, resourceVersion string, modulation DownsyncModulation) (bool, error)
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
	// Must not be called concurrently with any call that can add a resolution
	// with the same name.
	SetDestinations(bindingPolicyKey string, destinations sets.Set[string]) error

	// ResolutionExists returns true if a resolution is associated with the
	// given bindingpolicy key.
	ResolutionExists(bindingPolicyKey string) bool

	// GetReportedStateRequestForObject returns the combined effects of all
	// the resolutions regarding singleton reported state and multi-WEC reported state requests for a given workload object.
	// First and third is the `bool` indicating whether any BindingPolicy requests singleton or multi-WEC reported state return respectively
	// for the given object.
	// When first or third value is true then the second or fourth are the set of qualified WECs bound to that object respectively,
	// otherwise the second or fourth value is empty.
	// Returns: (wantSingleton, qualifiedWECsSingleton, wantMultiWEC, qualifiedWECsMulti)
	GetReportedStateRequestForObject(util.ObjectIdentifier) (bool, sets.Set[string], bool, sets.Set[string])

	// GetSingletonReportedStateRequestsForBinding calls GetReportedStateRequestForObject
	// for each of workload objects in the resolution if the resolution exists.
	// For each workload object, it returns the objectID, whether singleton reported state is requested and
	// the number of qualified WECs.
	// If the resolution doesn't exist then returns `nil`.
	GetSingletonReportedStateRequestsForBinding(bindingPolicyKey string) []SingletonReportedStateReturnStatus

	// DeleteResolution deletes the resolution associated with the given key,
	// if it exists.
	DeleteResolution(bindingPolicyKey string)

	// Broker returns a ResolutionBroker for the resolver.
	Broker() ResolutionBroker
}

// DownsyncModulation is a convenient internal representation of v1alpha1.DownsyncModulation
type DownsyncModulation struct {
	CreateOnly                 bool
	StatusCollectors           sets.Set[string]
	WantSingletonReportedState bool
	WantMultiWECReportedState  bool
}

func ZeroDownsyncModulation() DownsyncModulation {
	return DownsyncModulation{StatusCollectors: sets.New[string]()}
}

func DownsyncModulationFromExternal(external v1alpha1.DownsyncModulation) DownsyncModulation {
	return DownsyncModulation{
		CreateOnly:                 external.CreateOnly,
		StatusCollectors:           sets.New(external.StatusCollectors...),
		WantSingletonReportedState: external.WantSingletonReportedState,
		WantMultiWECReportedState:  external.WantMultiWECReportedState,
	}
}

func (dm *DownsyncModulation) ToExternal() v1alpha1.DownsyncModulation {
	return v1alpha1.DownsyncModulation{
		CreateOnly:                 dm.CreateOnly,
		StatusCollectors:           sets.List(dm.StatusCollectors),
		WantSingletonReportedState: dm.WantSingletonReportedState,
		WantMultiWECReportedState:  dm.WantMultiWECReportedState,
	}
}

func (left *DownsyncModulation) Equal(right DownsyncModulation) bool {
	return left.CreateOnly == right.CreateOnly &&
		left.WantSingletonReportedState == right.WantSingletonReportedState &&
		left.WantMultiWECReportedState == right.WantMultiWECReportedState &&
		left.StatusCollectors.Equal(right.StatusCollectors)
}

func (dm *DownsyncModulation) AddExternal(external v1alpha1.DownsyncModulation) {
	dm.CreateOnly = dm.CreateOnly || external.CreateOnly
	dm.StatusCollectors.Insert(external.StatusCollectors...)
	dm.WantSingletonReportedState = dm.WantSingletonReportedState || external.WantSingletonReportedState
	dm.WantMultiWECReportedState = dm.WantMultiWECReportedState || external.WantMultiWECReportedState
}

// SingletonReportedStateReturnStatus reports the resolver's state regarding
// requests for return of singleton reported state for a particular object.
type SingletonReportedStateReturnStatus struct {
	ObjectId                   util.ObjectIdentifier
	WantSingletonReportedState bool
	NumWECs                    int
}

func NewBindingPolicyResolver() BindingPolicyResolver {
	bpResolver := &bindingPolicyResolver{
		bindingPolicyToResolution: make(map[string]*bindingPolicyResolution),
	}
	bpResolver.broker = newResolutionBroker(bpResolver.getResolution, bpResolver.getAllResolutionKeys)

	return bpResolver
}

type bindingPolicyResolver struct {
	broker ResolutionBroker

	// Hold this mutex while accessing bindingPolicyToResolution.
	// This mutex may be held while acquiring a bindingPolicyResolution's mutex,
	// but not vice-versa.
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

	return bindingPolicyResolution.getOwnerReference(), nil
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

func (resolver *bindingPolicyResolver) NoteBindingPolicy(bindingpolicy *v1alpha1.BindingPolicy) {
	if resolution := resolver.getResolution(bindingpolicy.GetName()); resolution != nil {
		return
	}
	// Because concurrent calls with the same BindingPolicy name are not allowed,
	// it is guaranteed that createResolution will not find an existing entry ---
	// which means that we do not have to worry about updating that existing entry.
	resolver.createResolution(bindingpolicy)
}

// EnsureObjectData ensures that an object's identifier is
// in the resolution for the given bindingpolicy key, and is associated
// with the given resource-version, create-only bit, and statuscollectors set.
// The given set is expected not to be mutated during and after this call
// by the caller.
//
// The returned bool indicates whether the bindingpolicy resolution was
// changed. If no resolution is associated with the given key, an error is
// returned.
func (resolver *bindingPolicyResolver) EnsureObjectData(bindingPolicyKey string, objIdentifier util.ObjectIdentifier,
	objUID, resourceVersion string, modulation DownsyncModulation) (bool, error) {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe

	if bindingPolicyResolution == nil {
		// bindingPolicyKey is not associated with any resolution
		return false, fmt.Errorf("%s - bindingpolicy-key: %s", bindingPolicyResolutionNotFoundErrorPrefix,
			bindingPolicyKey)
	}

	// Now the resolver's mutex is not held, so the resolution just fetched could be concurrently
	// deleted and even a replacement inserted --- causing the following code to update a positively
	// wrong resolution. However, in that case there will be calls to `controller::reconcile` that
	// get the replacement fully updated.

	// ensureObjectIdentifier is thread-safe
	return bindingPolicyResolution.ensureObjectData(objIdentifier, objUID, resourceVersion, modulation), nil
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
	// The resolver's mutex is no longer held by this goroutine, so the resolution
	// could be concurrently deleted and even a replacement introduced --- so that
	// the following code will update a positively wrong resolution. However, we
	// expect that in this situation there will be later calls to
	// `controller::reconcile` that cause complete re-evaluation of the BindingPolicy.

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

	bindingPolicyResolution.RLock()
	defer bindingPolicyResolution.RUnlock()

	return sets.KeySet(bindingPolicyResolution.objectIdentifierToData), nil
}

func (resolver *bindingPolicyResolver) SetDestinations(bindingPolicyKey string,
	destinations sets.Set[string]) error {
	bindingPolicyResolution := resolver.getResolution(bindingPolicyKey) // thread-safe
	// Now the resolver's mutex is not held, so the resolution just fetched could be removed.
	// The prohibition against calling concurrently with methods that add a resolution ensures
	// that the following code will not update a positively wrong resolution.
	if bindingPolicyResolution == nil {
		return fmt.Errorf("%s - bindingpolicy-key: %s", bindingPolicyResolutionNotFoundErrorPrefix,
			bindingPolicyKey)
	}

	bindingPolicyResolution.Lock()
	defer bindingPolicyResolution.Unlock()

	bindingPolicyResolution.destinations = destinations
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

// GetReportedStateRequestForObject returns four things.
// First and third is the `bool` indicating whether any BindingPolicy requests singleton or multi-WEC reported state return respectively
// for the given object.
// If those are true then the second and fourth are the set of qualified WECs bound to that object respectively,
// otherwise the second or fourth value is empty.
func (resolver *bindingPolicyResolver) GetReportedStateRequestForObject(objId util.ObjectIdentifier) (bool, sets.Set[string], bool, sets.Set[string]) {
	resolver.RWMutex.RLock()
	defer resolver.RWMutex.RUnlock()

	var singletonRequested bool
	var multiWECRequested bool
	singletonWECs := sets.New[string]()
	multiWECs := sets.New[string]()

	// Single loop to check both singleton and multi-WEC requests and collect WECs
	for _, resolution := range resolver.bindingPolicyToResolution {
		singMatches, singRequest, singDests := resolution.getSingletonReportedStateRequestForObject(objId)
		if singMatches && singRequest {
			singletonRequested = true
			singletonWECs = singletonWECs.Union(singDests)
		}

		multiMatches, multiRequest, multiDests := resolution.getMultiWECReportedStateRequestForObject(objId)
		if multiMatches && multiRequest {
			multiWECRequested = true
			multiWECs = multiWECs.Union(multiDests)
		}
	}

	return singletonRequested, singletonWECs, multiWECRequested, multiWECs
}

func (resolver *bindingPolicyResolver) GetSingletonReportedStateRequestsForBinding(bindingPolicyKey string) []SingletonReportedStateReturnStatus {
	resolver.RWMutex.RLock()
	defer resolver.RWMutex.RUnlock()

	resolution := resolver.getResolution(bindingPolicyKey)
	if resolution == nil {
		return nil
	}
	objIds := resolution.getWorkloadReferences()
	ans := make([]SingletonReportedStateReturnStatus, len(objIds))
	for idx, objId := range objIds {
		want, qualifiedWECsSingleton, _, _ := resolver.GetReportedStateRequestForObject(objId)
		ans[idx] = SingletonReportedStateReturnStatus{objId, want, qualifiedWECsSingleton.Len()}
	}
	return ans
}

// DeleteResolution deletes the resolution associated with the given key,
// if it exists.
func (resolver *bindingPolicyResolver) DeleteResolution(bindingPolicyKey string) {
	resolver.Lock() // lock for modifying map
	defer resolver.Unlock()

	delete(resolver.bindingPolicyToResolution, bindingPolicyKey)
	resolver.broker.NotifyBindingPolicyCallbacks(bindingPolicyKey)
}

// Broker returns a ResolutionBroker for the resolver.
func (resolver *bindingPolicyResolver) Broker() ResolutionBroker {
	return resolver.broker
}

// getResolution retrieves the resolution associated with the given key.
// If the resolution does not exist, nil is returned.
func (resolver *bindingPolicyResolver) getResolution(bindingPolicyKey string) *bindingPolicyResolution {
	resolver.RLock()         // lock for reading map
	defer resolver.RUnlock() // unlock after accessing map

	return resolver.bindingPolicyToResolution[bindingPolicyKey]
}

// getAllResolutionKeys returns all keys associated with the maintained
// bindingpolicy resolutions.
func (resolver *bindingPolicyResolver) getAllResolutionKeys() []string {
	resolver.RLock()         // lock for reading map
	defer resolver.RUnlock() // unlock after accessing map

	keys := make([]string, 0, len(resolver.bindingPolicyToResolution))
	for key := range resolver.bindingPolicyToResolution {
		keys = append(keys, key)
	}

	return keys
}

// `*bindingPolicy` is immutable
func (resolver *bindingPolicyResolver) createResolution(bindingpolicy *v1alpha1.BindingPolicy) *bindingPolicyResolution {
	resolver.Lock() // lock for modifying map
	defer resolver.Unlock()

	// double-check existence to handle race conditions (common pattern)
	if bindingPolicyResolution, exists := resolver.bindingPolicyToResolution[bindingpolicy.GetName()]; exists {
		return bindingPolicyResolution
	}

	ownerReference := metav1.NewControllerRef(bindingpolicy, v1alpha1.SchemeGroupVersion.WithKind(util.BindingPolicyKind))
	ownerReference.BlockOwnerDeletion = &[]bool{false}[0]

	bindingPolicyResolution := &bindingPolicyResolution{
		singletonRequestChangeConsumer: func(objId util.ObjectIdentifier) {
			resolver.broker.NotifySingletonRequestCallbacks(bindingpolicy.Name, objId)
		},
		objectIdentifierToData: make(map[util.ObjectIdentifier]*ObjectData),
		destinations:           sets.New[string](),
		ownerReference:         ownerReference,
	}
	klog.InfoS("Created bindingPolicyResolution", "binding", bindingpolicy.Name, "resolution", fmt.Sprintf("%p", bindingPolicyResolution))
	resolver.bindingPolicyToResolution[bindingpolicy.GetName()] = bindingPolicyResolution

	return bindingPolicyResolution
}

func errorIsBindingPolicyResolutionNotFound(err error) bool {
	return strings.HasPrefix(err.Error(), bindingPolicyResolutionNotFoundErrorPrefix)
}
