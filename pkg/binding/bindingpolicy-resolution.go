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
// destinations and modulations for a single bindingpolicy. The mutex should be read locked
// before reading, and write locked before writing to any field.
type bindingPolicyResolution struct {
	sync.RWMutex

	// This map is mutable, but every `ObjectData` stored in it is immutable.
	// No `*ObjectData` in this map is nil.
	objectIdentifierToData map[util.ObjectIdentifier]*ObjectData

	// Every Set ever stored here is immutable from the time it is stored here.
	destinations sets.Set[string]

	// ownerReference identifies the bindingpolicy that this resolution is
	// associated with as an owning object.
	// This pointer is never nil (why is it a pointer?).
	// The `metav1.OwnerReference` is immutable.
	ownerReference *metav1.OwnerReference

	// requiresSingletonReportedState indicates whether the bindingpolicy
	// that this resolution is associated with requires singleton status.
	requiresSingletonReportedState bool
}

// ObjectData stores most of the downsync modalities for a given workload object
// with respect to a given Binding.
type ObjectData struct {
	// UID is the UID of the workload object.
	UID string
	// ResourceVersion is the resource version of the workload object.
	ResourceVersion string
	// CreateOnly is a boolean that indicates whether the object was selected
	// by at least one downsync clause with the createOnly bit set.
	// If true, it means that the object should only be created and not
	// maintained
	CreateOnly bool
	// StatusCollectors is a set of status collector names that are associated.
	// Every set ever stored here is immutable.
	StatusCollectors sets.Set[string]
}

// Assert that `*bindingPolicyResolution` implements Resolution
var _ Resolution = &bindingPolicyResolution{}

func (resolution *bindingPolicyResolution) GetPolicyUID() string {
	resolution.RLock()
	defer resolution.RUnlock()
	return string(resolution.ownerReference.UID)
}

func (resolution *bindingPolicyResolution) GetDestinations() sets.Set[string] {
	resolution.RLock()
	defer resolution.RUnlock()
	return resolution.destinations
}

func (resolution *bindingPolicyResolution) GetWorkload() abstract.Map[util.ObjectIdentifier, ObjectData] {
	m1 := abstract.AsPrimitiveMap(resolution.objectIdentifierToData)
	m2 := abstract.NewMapLocker(&resolution.RWMutex, m1)
	m3 := abstract.MapMapValues(m2, NonNilPointerDeference)
	return m3
}

// ensureObjectData ensures that an object identifier exists
// in the resolution and is associated with the given UID, resource version,
// create-only bit, and statuscollectors set.
// The given set is expected not to be mutated during and after this call by
// the caller.
//
// The returned bool indicates whether the resolution was changed.
// This function is thread-safe.
func (resolution *bindingPolicyResolution) ensureObjectData(objIdentifier util.ObjectIdentifier,
	objUID, resourceVersion string, createOnly bool, statusCollectors sets.Set[string]) bool {
	resolution.Lock()
	defer resolution.Unlock()

	objData := resolution.objectIdentifierToData[objIdentifier]
	if objData == nil || objData.UID != objUID || objData.ResourceVersion != resourceVersion ||
		objData.CreateOnly != createOnly || !objData.StatusCollectors.Equal(statusCollectors) {
		resolution.objectIdentifierToData[objIdentifier] = &ObjectData{
			UID:              objUID,
			ResourceVersion:  resourceVersion,
			CreateOnly:       createOnly,
			StatusCollectors: statusCollectors,
		}
		return true
	}

	return false
}

// removeObjectIdentifier removes an object identifier from the resolution if it
// exists. The return bool indicates whether the resolution was changed.
// This function is thread-safe.
func (resolution *bindingPolicyResolution) removeObjectIdentifier(objIdentifier util.ObjectIdentifier) bool {
	resolution.Lock()
	defer resolution.Unlock()

	if _, exists := resolution.objectIdentifierToData[objIdentifier]; !exists {
		return false
	}

	delete(resolution.objectIdentifierToData, objIdentifier)
	return true
}

// toBindingSpec converts the resolution to a binding
// spec. This function is thread-safe.
func (resolution *bindingPolicyResolution) toBindingSpec() *v1alpha1.BindingSpec {
	resolution.RLock()
	defer resolution.RUnlock()

	workload := v1alpha1.DownsyncObjectClauses{}

	// iterate over all objects and build workload efficiently. No (GVR, namespace, name) tuple is
	// duplicated in the objectIdentifierToResourceVersion map, due to the uniqueness of the identifiers.
	// Therefore, whenever an object is about to be appended, we simply append.
	for objIdentifier, objData := range resolution.objectIdentifierToData {
		// check if object is cluster-scoped or namespaced by checking namespace
		if objIdentifier.ObjectName.Namespace == metav1.NamespaceNone {
			clause := v1alpha1.ClusterScopeDownsyncClause{
				ClusterScopeDownsyncObject: v1alpha1.ClusterScopeDownsyncObject{
					GroupVersionResource: metav1.GroupVersionResource(objIdentifier.GVR()),
					Name:                 objIdentifier.ObjectName.Name,
					ResourceVersion:      objData.ResourceVersion,
				},
				CreateOnly: objData.CreateOnly,
			}
			if objData.StatusCollectors.Len() > 0 {
				clause.StatusCollection = &v1alpha1.StatusCollection{
					StatusCollectors: sets.List(objData.StatusCollectors),
				}
			}

			workload.ClusterScope = append(workload.ClusterScope, clause)
			continue
		}

		clause := v1alpha1.NamespaceScopeDownsyncClause{
			NamespaceScopeDownsyncObject: v1alpha1.NamespaceScopeDownsyncObject{
				GroupVersionResource: metav1.GroupVersionResource(objIdentifier.GVR()),
				Name:                 objIdentifier.ObjectName.Name,
				Namespace:            objIdentifier.ObjectName.Namespace,
				ResourceVersion:      objData.ResourceVersion,
			},
			CreateOnly: objData.CreateOnly,
		}
		if objData.StatusCollectors.Len() > 0 {
			clause.StatusCollection = &v1alpha1.StatusCollection{
				StatusCollectors: sets.List(objData.StatusCollectors),
			}
		}

		workload.NamespaceScope = append(workload.NamespaceScope, clause)
	}

	// sort workload objects
	sortBindingWorkloadObjects(&workload)

	return &v1alpha1.BindingSpec{
		Workload:     workload,
		Destinations: destinationsStringSetToSortedDestinations(resolution.destinations),
	}
}

func (resolution *bindingPolicyResolution) matchesBindingSpec(bindingSpec *v1alpha1.BindingSpec) bool {
	resolution.RLock()
	defer resolution.RUnlock()

	// check destinations
	if !destinationsMatch(resolution.destinations, bindingSpec.Destinations) {
		return false
	}

	// check workload
	if len(resolution.objectIdentifierToData) != len(bindingSpec.Workload.ClusterScope)+
		len(bindingSpec.Workload.NamespaceScope) {
		return false
	}

	objRefToDataFromWorkload := bindingObjectRefToDataFromWorkload(&bindingSpec.Workload)

	for objIdentifier, objData := range resolution.objectIdentifierToData {
		// check if object ref exists, then check if the object data matches
		if objDataFromWorkload := objRefToDataFromWorkload[objectRef{
			GroupVersionResource: objIdentifier.GVR(),
			ObjectName:           objIdentifier.ObjectName,
		}]; objDataFromWorkload == nil ||
			objData.ResourceVersion != objDataFromWorkload.ResourceVersion ||
			objData.CreateOnly != objDataFromWorkload.CreateOnly ||
			!objData.StatusCollectors.Equal(objDataFromWorkload.StatusCollectors) {
			return false
		}
	} // this check works because both groups have unique members and are of equal size

	return true
}

// getDestinationsList returns a sorted list of v1alpha1.Destination in the
// resolution.
func (resolution *bindingPolicyResolution) getDestinationsList() []v1alpha1.Destination {
	resolution.RLock()
	defer resolution.RUnlock()

	return destinationsStringSetToSortedDestinations(resolution.destinations)
}

func (resolution *bindingPolicyResolution) getWorkloadReferences() []util.ObjectIdentifier {
	resolution.RLock()
	defer resolution.RUnlock()

	return abstract.PrimitiveMapKeySlice(resolution.objectIdentifierToData)
}

// getOwnerReference returns the owner reference of the resolution.
func (resolution *bindingPolicyResolution) getOwnerReference() metav1.OwnerReference {
	resolution.RLock()
	defer resolution.RUnlock()

	return *resolution.ownerReference
}

func (resolution *bindingPolicyResolution) setRequestsSingletonReportedStateReturn(request bool) {
	resolution.Lock()
	defer resolution.Unlock()

	resolution.requiresSingletonReportedState = request
}

func (resolution *bindingPolicyResolution) getRequestsSingletonReportedStateReturn() bool {
	resolution.RLock()
	defer resolution.RUnlock()

	return resolution.requiresSingletonReportedState
}

// getSingletonReportedStateRequestForObject returns what this resolution requests regarding singleton reported state return.
// The first returned bool reports whether this resolution matches the given workload object ID;
// if not then the other returned values are meaningless.
// The second returned bool indicates whether singleton reported state return is requested.
// The returned set is the names of the matching WECs, and is immutable.
func (resolution *bindingPolicyResolution) getSingletonReportedStateRequestForObject(objId util.ObjectIdentifier) (bool, bool, sets.Set[string]) {
	resolution.RLock()
	defer resolution.RUnlock()
	if _, has := resolution.objectIdentifierToData[objId]; !has {
		return false, false, nil
	}
	return true, resolution.requiresSingletonReportedState, resolution.destinations
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
}

func bindingObjectRefToDataFromWorkload(bindingSpecWorkload *v1alpha1.DownsyncObjectClauses) map[objectRef]*ObjectData {
	bindingObjectRefToData := make(map[objectRef]*ObjectData)

	for _, clusterScopeDownsyncClause := range bindingSpecWorkload.ClusterScope {
		var statusCollectors sets.Set[string]
		if clusterScopeDownsyncClause.StatusCollection != nil {
			statusCollectors = sets.New(clusterScopeDownsyncClause.StatusCollection.StatusCollectors...)
		}

		bindingObjectRefToData[objectRef{
			GroupVersionResource: schema.GroupVersionResource(clusterScopeDownsyncClause.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      clusterScopeDownsyncClause.Name,
				Namespace: metav1.NamespaceNone,
			},
		}] = &ObjectData{
			ResourceVersion:  clusterScopeDownsyncClause.ResourceVersion,
			CreateOnly:       clusterScopeDownsyncClause.CreateOnly,
			StatusCollectors: statusCollectors,
		}
	}

	for _, namespacedScopeDownsyncClause := range bindingSpecWorkload.NamespaceScope {
		var statusCollectors sets.Set[string]
		if namespacedScopeDownsyncClause.StatusCollection != nil {
			statusCollectors = sets.New(namespacedScopeDownsyncClause.StatusCollection.StatusCollectors...)
		}

		bindingObjectRefToData[objectRef{
			GroupVersionResource: schema.GroupVersionResource(namespacedScopeDownsyncClause.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      namespacedScopeDownsyncClause.Name,
				Namespace: namespacedScopeDownsyncClause.Namespace,
			},
		}] = &ObjectData{
			ResourceVersion:  namespacedScopeDownsyncClause.ResourceVersion,
			CreateOnly:       namespacedScopeDownsyncClause.CreateOnly,
			StatusCollectors: statusCollectors,
		}
	}

	return bindingObjectRefToData
}

func destinationsStringSetToSortedDestinations(destinationsStringSet sets.Set[string]) []v1alpha1.Destination {
	sortedDestinations := make([]v1alpha1.Destination, 0, len(destinationsStringSet))

	for _, d := range sets.List(destinationsStringSet) {
		sortedDestinations = append(sortedDestinations, v1alpha1.Destination{ClusterId: d})
	}

	return sortedDestinations
}

func sortBindingWorkloadObjects(bindingWorkload *v1alpha1.DownsyncObjectClauses) {
	// sort clusterScopeDownsyncObjects
	slices.SortFunc(bindingWorkload.ClusterScope, func(a, b v1alpha1.ClusterScopeDownsyncClause) bool {
		if a.GroupVersionResource.String() != b.GroupVersionResource.String() {
			return a.GroupVersionResource.String() < b.GroupVersionResource.String()
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.ResourceVersion < b.ResourceVersion
	})
	// sort namespaceScopeDownsyncObjects
	slices.SortFunc(bindingWorkload.NamespaceScope, func(a, b v1alpha1.NamespaceScopeDownsyncClause) bool {
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
