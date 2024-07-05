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
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
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
	GetResolution(bindingPolicyKey string) *Resolution

	// NotifyCallbacks calls all registered callbacks with the given bindingPolicyKey.
	// The callbacks are called on separate goroutines.
	NotifyCallbacks(bindingPolicyKey string)
}

// Resolution is a struct that represents the resolution of a binding policy.
// It contains the destinations and object data for the resolution.
type Resolution struct {
	UID string
	// Destinations is a list of destinations that are the "where" part of the
	// resolution.
	Destinations []v1alpha1.Destination
	// ObjectIdentifierToData is a map of object identifiers to their
	// corresponding ObjectData.
	ObjectIdentifierToData map[util.ObjectIdentifier]ObjectData
}

// ObjectData is a struct that represents the data associated with an object
// in a resolution.
type ObjectData struct {
	UID              string
	ResourceVersion  string
	StatusCollectors []string
}

func (r *Resolution) RequiresStatusCollection() bool {
	for _, data := range r.ObjectIdentifierToData {
		if len(data.StatusCollectors) > 0 {
			return true
		}
	}

	return false
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
func (broker *resolutionBroker) GetResolution(bindingPolicyKey string) *Resolution {
	bindingPolicyResolution := broker.bindingPolicyResolutionGetter(bindingPolicyKey) //thread-safe

	if bindingPolicyResolution == nil {
		return nil
	}

	return &Resolution{
		UID:          string(bindingPolicyResolution.ownerReference.UID), // necessarily exists
		Destinations: bindingPolicyResolution.getDestinationsList(),
		ObjectIdentifierToData: abstract.PrimitiveMapSafeValMap(&bindingPolicyResolution.RWMutex,
			bindingPolicyResolution.objectIdentifierToData,
			func(data *objectData) ObjectData {
				return ObjectData{
					UID:              string(data.UID),
					ResourceVersion:  data.ResourceVersion,
					StatusCollectors: sets.List(data.StatusCollectors), // members are string copies
				}
			}), // while this function breaks the constraint, it maintains its own concurrency safety
		// by using the PrimitiveMapSafeValMap which transforms a map safely using its read-lock.
	}
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
