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

package binding

import (
	"io"
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/pkg/abstract"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// ResolutionBroker allows for the registration of callback functions that
// are called when a resolution is updated.
// It also allows for the retrieval of resolutions for a given bindingPolicyKey.
type ResolutionBroker interface {
	// RegisterCallback adds a new callback function that will be called,
	// on a separate goroutine, when a resolution is updated. The callback
	// function should accept a bindingPolicyKey.
	// There is no deduplication of callbacks.
	RegisterCallback(func(bindingPolicyKey string))

	// GetResolution retrieves the resolution for a given bindingPolicyKey.
	// If no resolution is associated with the given key, nil is returned.
	GetResolution(bindingPolicyKey string) Resolution

	// NotifyCallbacks calls all registered callbacks with the given bindingPolicyKey.
	// The callbacks are called on separate goroutines.
	NotifyCallbacks(bindingPolicyKey string)
}

// Resolution is an internal representation of what a BindingPolicy is calling for.
// A Resolution may change over time; the methods read an attribute's value at the current point in time.
// Thread-safe.
type Resolution interface {
	// GetPolicyUID returns the UID of the BindingPolicy
	// for use in constructing CombinedStatus object names.
	GetPolicyUID() string

	// GetDestinations returns a Set holding the names of the inventory objects.
	// The returned set is immutable.
	GetDestinations() sets.Set[string]

	// GetWorkload returns a Map holding the current workload object references and
	// associated downsyn modalities.
	// The contents of the Map may change over time; the consumer of Iterate2 must not
	// access the map (neither directly nor indirectly).
	GetWorkload() abstract.Map[util.ObjectIdentifier, ObjectData]
}

func NonNilPointerDeference[T any](ptr *T) T { return *ptr }

func RequiresStatusCollection(r Resolution) bool {
	return r.GetWorkload().Iterate2(func(_ util.ObjectIdentifier, data ObjectData) error {
		if len(data.StatusCollectors) > 0 {
			return io.EOF
		}
		return nil
	}) != nil
}

// newResolutionBroker creates a new ResolutionBroker with the given
// bindingPolicyResolutionGetter function.
// The latter function's returned `bindingPolicyResolution` is expected to be
// thread-safe under the constraint that only its methods are used.
func newResolutionBroker(bindingPolicyResolutionGetter func(bindingPolicyKey string) *bindingPolicyResolution,
	allBindingPolicyResolutionKeysGetter func() []string) ResolutionBroker {
	return &resolutionBroker{
		bindingPolicyResolutionGetter:        bindingPolicyResolutionGetter,
		allBindingPolicyResolutionKeysGetter: allBindingPolicyResolutionKeysGetter,
	}
}

// resolutionBroker implements the ResolutionBroker interface.
// The broker requires a bindingPolicyResolutionGetter function that retrieves
// resolutions from a BindingPolicyResolver.
type resolutionBroker struct {
	sync.Mutex
	callbacks []func(bindingPolicyKey string)
	// bindingPolicyResolutionGetter is a function that retrieves the
	// resolution for a given bindingPolicyKey. If no resolution is associated
	// with the given key, nil is returned.
	// The returned `bindingPolicyResolution` is expected to be thread-safe as
	// long as only its methods are used.
	bindingPolicyResolutionGetter func(bindingPolicyKey string) *bindingPolicyResolution

	allBindingPolicyResolutionKeysGetter func() []string
}

// RegisterCallback adds a new callback function that will be called, on a
// separate goroutine, when a resolution is updated. The callback function
// should accept a bindingPolicyKey.
//
// Upon registration, the callback is immediately called with all existing
// resolutions.
// There is no deduplication of callbacks.
func (broker *resolutionBroker) RegisterCallback(callback func(bindingPolicyKey string)) {
	broker.Lock()
	defer broker.Unlock()

	broker.callbacks = append(broker.callbacks, callback)

	for _, bindingPolicyKey := range broker.allBindingPolicyResolutionKeysGetter() {
		go callback(bindingPolicyKey)
	}
}

// GetResolution retrieves the resolution for a given bindingPolicyKey.
// If no resolution is associated with the given key, nil is returned.
func (broker *resolutionBroker) GetResolution(bindingPolicyKey string) Resolution {
	resolution := broker.bindingPolicyResolutionGetter(bindingPolicyKey) //thread-safe
	if resolution == nil {
		return nil
	}
	return resolution
}

// NotifyCallbacks calls all registered callbacks with the given bindingPolicyKey.
// The callbacks are called on separate goroutines.
func (broker *resolutionBroker) NotifyCallbacks(bindingPolicyKey string) {
	broker.Lock()
	defer broker.Unlock()

	for _, callback := range broker.callbacks {
		go callback(bindingPolicyKey)
	}
}
