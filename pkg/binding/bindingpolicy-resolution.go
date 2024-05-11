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

	"golang.org/x/exp/slices"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/abstract"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// bindingPolicyResolution stores the selected object identifiers and
// destinations for a single bindingpolicy. The mutex should be read locked
// before reading, and write locked before writing to any field.
type bindingPolicyResolution struct {
	sync.RWMutex

	objectIdentifierToResourceVersion  map[util.ObjectIdentifier]string
	objectIdentifierToStatusCollectors map[util.ObjectIdentifier]sets.Set[string]

	destinations sets.Set[string]

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

// setStatusCollectorsForObjectIdentifier sets the statuscollector names for
// the given object identifier. The return bool indicates whether the
// resolution was changed.
// The given set is expected not to be mutated during and after this call by
// the caller.
// If the object identifier does not exist in the resolution, nothing happens.
// This function is thread-safe.
func (resolution *bindingPolicyResolution) setStatusCollectorsForObjectIdentifier(objIdentifier util.ObjectIdentifier,
	statusCollectors sets.Set[string]) bool {
	resolution.Lock()
	defer resolution.Unlock()

	if _, exists := resolution.objectIdentifierToResourceVersion[objIdentifier]; !exists {
		return false
	}

	// if the object identifier does not exist in the statuscollectors map, add it
	if _, exists := resolution.objectIdentifierToStatusCollectors[objIdentifier]; !exists {
		resolution.objectIdentifierToStatusCollectors[objIdentifier] = statusCollectors
		return true
	}

	// if the object identifier's statuscollectors set is already equal the given set, do nothing
	if resolution.objectIdentifierToStatusCollectors[objIdentifier].Equal(statusCollectors) {
		return false
	}

	resolution.objectIdentifierToStatusCollectors[objIdentifier] = statusCollectors
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

	return sets.KeySet(resolution.objectIdentifierToResourceVersion)
}

// toBindingSpec converts the resolution to a binding
// spec. This function is thread-safe.
func (resolution *bindingPolicyResolution) toBindingSpec() *v1alpha1.BindingSpec {
	resolution.RLock()
	defer resolution.RUnlock()

	workload := v1alpha1.DownsyncObjectReferences{}
	statusCollectorNames, statusCollectorNameToIndex := resolution.generateStatusCollectorsListAndIndices()

	// iterate over all objects and build workload efficiently. No (GVR, namespace, name) tuple is
	// duplicated in the objectIdentifierToResourceVersion map, due to the uniqueness of the identifiers.
	// Therefore, whenever an object is about to be appended, we simply append.
	for objIdentifier, objResourceVersion := range resolution.objectIdentifierToResourceVersion {
		// get the statuscollector names for the object
		statusCollectorNamesForObj := resolution.objectIdentifierToStatusCollectors[objIdentifier].UnsortedList()

		// check if object is cluster-scoped or namespaced by checking namespace
		if objIdentifier.ObjectName.Namespace == metav1.NamespaceNone {
			workload.ClusterScope = append(workload.ClusterScope, v1alpha1.ClusterScopeDownsyncObject{
				GroupVersionResource: metav1.GroupVersionResource(objIdentifier.GVR()),
				Name:                 objIdentifier.ObjectName.Name,
				ResourceVersion:      objResourceVersion,
				StatusCollectorIndices: abstract.SliceMap(statusCollectorNamesForObj, func(name string) int {
					return statusCollectorNameToIndex[name]
				}),
			})

			continue
		}

		workload.NamespaceScope = append(workload.NamespaceScope, v1alpha1.NamespaceScopeDownsyncObject{
			GroupVersionResource: metav1.GroupVersionResource(objIdentifier.GVR()),
			Name:                 objIdentifier.ObjectName.Name,
			Namespace:            objIdentifier.ObjectName.Namespace,
			ResourceVersion:      objResourceVersion,
			StatusCollectorIndices: abstract.SliceMap(statusCollectorNamesForObj, func(name string) int {
				return statusCollectorNameToIndex[name]
			}),
		})
	}

	// sort workload objects
	sortBindingWorkloadObjects(&workload)

	return &v1alpha1.BindingSpec{
		Workload:         workload,
		Destinations:     destinationsStringSetToSortedDestinations(resolution.destinations),
		StatusCollectors: statusCollectorNames,
	}
}

func (resolution *bindingPolicyResolution) generateStatusCollectorsListAndIndices() ([]string, map[string]int) {
	resolution.RLock()
	defer resolution.RUnlock()

	statusCollectorsSet := sets.New[string]()

	for _, statusCollectors := range resolution.objectIdentifierToStatusCollectors {
		statusCollectorsSet.Insert(statusCollectors.UnsortedList()...)
	}

	statusCollectorsList := statusCollectorsSet.UnsortedList()
	statusCollectorNameToIndex := make(map[string]int, statusCollectorsSet.Len())

	for i, name := range statusCollectorsList {
		statusCollectorNameToIndex[name] = i
	}

	return statusCollectorsList, statusCollectorNameToIndex
}

func (resolution *bindingPolicyResolution) matchesBindingSpec(bindingSpec *v1alpha1.BindingSpec) bool {
	resolution.RLock()
	defer resolution.RUnlock()

	// check destinations
	if !destinationsMatch(resolution.destinations, bindingSpec.Destinations) {
		return false
	}

	// check workload
	if len(resolution.objectIdentifierToResourceVersion) != len(bindingSpec.Workload.ClusterScope)+
		len(bindingSpec.Workload.NamespaceScope) {
		return false
	}

	objectRefToStatusCollectorsSet := bindingObjectRefToStatusCollectorsSetFromWorkload(&bindingSpec.Workload,
		bindingSpec.StatusCollectors)

	for objIdentifier, objResourceVersion := range resolution.objectIdentifierToResourceVersion {
		// check if object ref exists, then check if the status collectors are equal
		if objectRefSetFromWorkload, exists := objectRefToStatusCollectorsSet[objectRef{
			GroupVersionResource: objIdentifier.GVR(),
			ObjectName:           objIdentifier.ObjectName,
			ResourceVersion:      objResourceVersion,
		}]; !exists || !resolution.objectIdentifierToStatusCollectors[objIdentifier].Equal(objectRefSetFromWorkload) {
			return false
		}

	} // this check works because both groups have unique members and are of equal size

	return true
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

type objectRef struct {
	schema.GroupVersionResource
	cache.ObjectName
	ResourceVersion string
}

func bindingObjectRefToStatusCollectorsSetFromWorkload(bindingSpecWorkload *v1alpha1.DownsyncObjectReferences,
	bindingSpecStatusCollectors []string) map[objectRef]sets.Set[string] {
	bindingObjectRefAndVersionToStatusCollectors := make(map[objectRef]sets.Set[string])

	for _, clusterScopeDownsyncObject := range bindingSpecWorkload.ClusterScope {
		bindingObjectRefAndVersionToStatusCollectors[objectRef{
			GroupVersionResource: schema.GroupVersionResource(clusterScopeDownsyncObject.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      clusterScopeDownsyncObject.Name,
				Namespace: metav1.NamespaceNone,
			},
			ResourceVersion: clusterScopeDownsyncObject.ResourceVersion,
		}] = sets.New[string](abstract.SliceMap(clusterScopeDownsyncObject.StatusCollectorIndices, func(index int) string {
			return bindingSpecStatusCollectors[index]
		})...)
	}

	for _, namespacedScopeDownsyncObject := range bindingSpecWorkload.NamespaceScope {
		bindingObjectRefAndVersionToStatusCollectors[objectRef{
			GroupVersionResource: schema.GroupVersionResource(namespacedScopeDownsyncObject.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      namespacedScopeDownsyncObject.Name,
				Namespace: namespacedScopeDownsyncObject.Namespace,
			},
			ResourceVersion: namespacedScopeDownsyncObject.ResourceVersion,
		}] = sets.New[string](abstract.SliceMap(namespacedScopeDownsyncObject.StatusCollectorIndices, func(index int) string {
			return bindingSpecStatusCollectors[index]
		})...)
	}

	return bindingObjectRefAndVersionToStatusCollectors
}

func destinationsStringSetToSortedDestinations(destinationsStringSet sets.Set[string]) []v1alpha1.Destination {
	sortedDestinations := make([]v1alpha1.Destination, 0, len(destinationsStringSet))

	for _, d := range sets.List(destinationsStringSet) {
		sortedDestinations = append(sortedDestinations, v1alpha1.Destination{ClusterId: d})
	}

	return sortedDestinations
}

func sortBindingWorkloadObjects(bindingWorkload *v1alpha1.DownsyncObjectReferences) {
	// sort clusterScopeDownsyncObjects
	slices.SortFunc(bindingWorkload.ClusterScope, func(a, b v1alpha1.ClusterScopeDownsyncObject) bool {
		if a.GroupVersionResource.String() != b.GroupVersionResource.String() {
			return a.GroupVersionResource.String() < b.GroupVersionResource.String()
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.ResourceVersion < b.ResourceVersion
	})
	// sort namespaceScopeDownsyncObjects
	slices.SortFunc(bindingWorkload.NamespaceScope, func(a, b v1alpha1.NamespaceScopeDownsyncObject) bool {
		if a.GroupVersionResource.String() != b.GroupVersionResource.String() {
			return a.GroupVersionResource.String() < b.GroupVersionResource.String()
		}
		objectNameA := cache.NewObjectName(a.Namespace, a.Name).String()
		objectNameB := cache.NewObjectName(b.Namespace, b.Name).String()
		if objectNameA != objectNameB {
			return objectNameA < objectNameB
		}
		return a.ResourceVersion < b.ResourceVersion
	})
}
