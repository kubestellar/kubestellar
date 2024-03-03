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
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// bindingPolicyResolution stores the selected object identifiers and
// destinations for a single bindingpolicy. The mutex should be read locked
// before reading, and write locked before writing to any field.
type bindingPolicyResolution struct {
	sync.RWMutex

	objectIdentifierToResourceVersion map[util.ObjectIdentifier]string
	destinations                      sets.Set[string]

	// ownerReference identifies the bindingpolicy that this resolution is
	// associated with as an owning object.
	ownerReference *metav1.OwnerReference

	// requiresSingletonReportedState indicates whether the bindingpolicy
	// that this resolution is associated with requires singleton status.
	requiresSingletonReportedState bool
}

// ensureObjectIdentifierWithVersion ensures that an object identifier exists
// in the resolution and is associated with the given resource version.
// The returned bool indicates whether the resolution was changed.
// This function is thread-safe.
func (resolution *bindingPolicyResolution) ensureObjectIdentifierWithVersion(objIdentifier util.ObjectIdentifier,
	resourceVersion string) bool {
	resolution.Lock()
	defer resolution.Unlock()

	currentResourceVersion := resolution.objectIdentifierToResourceVersion[objIdentifier]
	if currentResourceVersion == resourceVersion {
		return false
	}

	resolution.objectIdentifierToResourceVersion[objIdentifier] = resourceVersion
	return true
}

// removeObjectIdentifier removes an object identifier from the resolution if it
// exists. The return bool indicates whether the resolution was changed.
// This function is thread-safe.
func (resolution *bindingPolicyResolution) removeObjectIdentifier(objIdentifier util.ObjectIdentifier) bool {
	resolution.Lock()
	defer resolution.Unlock()

	if _, exists := resolution.objectIdentifierToResourceVersion[objIdentifier]; !exists {
		return false
	}

	delete(resolution.objectIdentifierToResourceVersion, objIdentifier)
	return true
}

// setDestinations updates the destinations list in the resolution.
// The given destinations set is expected not to be mutated after this call.
func (resolution *bindingPolicyResolution) setDestinations(destinations sets.Set[string]) {
	resolution.Lock()
	defer resolution.Unlock()

	resolution.destinations = destinations
}

// getObjectIdentifiers returns a copy of the object identifiers in the resolution.
func (resolution *bindingPolicyResolution) getObjectIdentifiers() sets.Set[util.ObjectIdentifier] {
	resolution.RLock()
	defer resolution.RUnlock()

	objIdentifiers := sets.New[util.ObjectIdentifier]()
	for objIdentifier := range resolution.objectIdentifierToResourceVersion {
		objIdentifiers.Insert(objIdentifier)
	}

	return objIdentifiers
}

// toBindingSpec converts the resolution to a binding
// spec. This function is thread-safe.
func (resolution *bindingPolicyResolution) toBindingSpec() (*v1alpha1.BindingSpec, error) {
	resolution.RLock()
	defer resolution.RUnlock()

	workload := v1alpha1.DownsyncObjectReferences{}

	// iterate over all objects and build workload efficiently. No (GVR, namespace, name) tuple is
	// duplicated in the objectIdentifierToResourceVersion map, due to the uniqueness of the identifiers.
	// Therefore, whenever an object is about to be appended, we simply append.
	for objIdentifier, objResourceVersion := range resolution.objectIdentifierToResourceVersion {
		// check if object is cluster-scoped or namespaced by checking namespace
		if objIdentifier.ObjectName.Namespace == metav1.NamespaceNone {
			workload.ClusterScope = append(workload.ClusterScope, v1alpha1.ClusterScopeDownsyncObject{
				GroupVersionResource: metav1.GroupVersionResource(objIdentifier.GVR()),
				Name:                 objIdentifier.ObjectName.Name,
				ResourceVersion:      objResourceVersion,
			})

			continue
		}

		workload.NamespaceScope = append(workload.NamespaceScope, v1alpha1.NamespaceScopeDownsyncObject{
			GroupVersionResource: metav1.GroupVersionResource(objIdentifier.GVR()),
			Name:                 objIdentifier.ObjectName.Name,
			Namespace:            objIdentifier.ObjectName.Namespace,
			ResourceVersion:      objResourceVersion,
		})
	}

	return &v1alpha1.BindingSpec{
		Workload:     workload,
		Destinations: destinationsStringSetToDestinations(resolution.destinations),
	}, nil
}

func (resolution *bindingPolicyResolution) matchesBindingSpec(bindingSpec *v1alpha1.BindingSpec) bool {
	resolution.RLock()
	defer resolution.RUnlock()

	// check destinations
	if !destinationsMatch(resolution.destinations, bindingSpec.Destinations) {
		return false
	}

	// check workload
	resolutionBindingObjectRefSet := bindingObjectRefSetFromResolution(resolution)
	workloadBindingObjectRefSet := bindingObjectRefSetFromBindingWorkload(&bindingSpec.Workload)

	return resolutionBindingObjectRefSet.Equal(workloadBindingObjectRefSet)
}

// destinationsMatch returns true if the destinations in the resolution
// match the destinations in the binding spec.
func destinationsMatch(resolvedDestinations sets.Set[string], bindingDestinations []v1alpha1.Destination) bool {
	if len(resolvedDestinations) != len(bindingDestinations) {
		return false
	}

	for _, destination := range bindingDestinations {
		if !resolvedDestinations.Has(destination.ClusterId) {
			return false
		}
	}

	return true
}

type bindingObjectRef struct {
	schema.GroupVersionResource
	cache.ObjectName
	ResourceVersion string
}

func bindingObjectRefSetFromResolution(resolution *bindingPolicyResolution) sets.Set[bindingObjectRef] {
	bindingObjectRefSet := sets.New[bindingObjectRef]()

	for objIdentifier, objResourceVersion := range resolution.objectIdentifierToResourceVersion {
		bindingObjectRefSet.Insert(bindingObjectRef{
			GroupVersionResource: objIdentifier.GVR(),
			ObjectName:           objIdentifier.ObjectName,
			ResourceVersion:      objResourceVersion,
		})
	}

	return bindingObjectRefSet
}

func bindingObjectRefSetFromBindingWorkload(bindingSpecWorkload *v1alpha1.DownsyncObjectReferences) sets.Set[bindingObjectRef] {
	bindingObjectRefSet := sets.New[bindingObjectRef]()

	for _, clusterScopeDownsyncObject := range bindingSpecWorkload.ClusterScope {
		bindingObjectRefSet.Insert(bindingObjectRef{
			GroupVersionResource: schema.GroupVersionResource(clusterScopeDownsyncObject.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      clusterScopeDownsyncObject.Name,
				Namespace: metav1.NamespaceNone,
			},
			ResourceVersion: clusterScopeDownsyncObject.ResourceVersion,
		})
	}

	for _, namespacedScopeDownsyncObject := range bindingSpecWorkload.NamespaceScope {
		bindingObjectRefSet.Insert(bindingObjectRef{
			GroupVersionResource: schema.GroupVersionResource(namespacedScopeDownsyncObject.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      namespacedScopeDownsyncObject.Name,
				Namespace: namespacedScopeDownsyncObject.Namespace,
			},
			ResourceVersion: namespacedScopeDownsyncObject.ResourceVersion,
		})
	}

	return bindingObjectRefSet
}

func destinationsStringSetToDestinations(destinations sets.Set[string]) []v1alpha1.Destination {
	dests := make([]v1alpha1.Destination, 0, len(destinations))
	for d := range destinations {
		dests = append(dests, v1alpha1.Destination{ClusterId: d})
	}

	return dests
}
