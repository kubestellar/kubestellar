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
	"fmt"
	"io"
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/abstract"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// ResolutionBroker allows for the registration of callback functions that
// are called when a resolution is updated.
// It also allows for the retrieval of resolutions for a given bindingPolicyKey.
type ResolutionBroker interface {
	// RegisterCallback adds a new tuple of callbacks.
	RegisterCallbacks(ResolutionCallbacks) error

	// GetResolution retrieves the resolution for a given bindingPolicyKey.
	// If no resolution is associated with the given key, nil is returned.
	GetResolution(bindingPolicyKey string) Resolution

	// NotifyBindingPolicyCallbacks calls all registered callbacks with the given bindingPolicyKey.
	// The callbacks are called the same goroutine as this call.
	NotifyBindingPolicyCallbacks(bindingPolicyKey string)

	// NotifySingletonRequestCallbacks calls all the registered callbacks for the
	// given policy and workload object, in the same goroutine.
	NotifySingletonRequestCallbacks(bindingPolicyKey string, objId util.ObjectIdentifier)
}

type ResolutionCallbacks struct {
	BindingPolicyChanged                 func(bindingPolicyKey string)
	SingletonReportedStateRequestChanged func(bindingPolicyKey string, objId util.ObjectIdentifier)
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
		if len(data.Modulation.StatusCollectors) > 0 {
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

	callbackses []ResolutionCallbacks

	// bindingPolicyResolutionGetter is a function that retrieves the
	// resolution for a given bindingPolicyKey. If no resolution is associated
	// with the given key, nil is returned.
	// The returned `bindingPolicyResolution` is expected to be thread-safe as
	// long as only its methods are used.
	bindingPolicyResolutionGetter func(bindingPolicyKey string) *bindingPolicyResolution

	// allBindingPolicyResolutionKeysGetter returns the keys of all the currently existing resolutions.
	allBindingPolicyResolutionKeysGetter func() []string
}

// RegisterCallbacks adds a new tuple of functions that will be called when
// certain things change. The callbacks are done in the goroutine triggering them.
// The broker should be empty at registration time.
// There is no deduplication of callbacks.
func (broker *resolutionBroker) RegisterCallbacks(callbacks ResolutionCallbacks) error {
	broker.Lock()
	defer broker.Unlock()

	broker.callbackses = append(broker.callbackses, callbacks)
	klog.Infof("RegisterCallbacks(%p=%v)(%p); total is now %d", broker, broker, callbacks, len(broker.callbackses))
	return nil
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

// NotifyBindingPolicyCallbacks calls all registered callbacks with the given bindingPolicyKey.
// The callbacks are called synchronously.
func (broker *resolutionBroker) NotifyBindingPolicyCallbacks(bindingPolicyKey string) {
	broker.Lock()
	defer broker.Unlock()

	for _, callbacks := range broker.callbackses {
		callbacks.BindingPolicyChanged(bindingPolicyKey)
	}
}

func (broker *resolutionBroker) NotifySingletonRequestCallbacks(bindingPolicyKey string, objId util.ObjectIdentifier) {
	broker.Lock()
	defer broker.Unlock()
	klog.InfoS("Relaying singleton request callback", "broker", fmt.Sprintf("%p=%v", broker, broker), "binding", bindingPolicyKey, "objId", objId, "numCallbacks", len(broker.callbackses))

	for idx, callbacks := range broker.callbackses {
		klog.InfoS("Relaying singleton request callback", "binding", bindingPolicyKey, "objId", objId, "idx", idx, "callbacks", callbacks)
		callbacks.SingletonReportedStateRequestChanged(bindingPolicyKey, objId)
	}

}
