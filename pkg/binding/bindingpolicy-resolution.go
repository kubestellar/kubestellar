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
	"github.com/kubestellar/kubestellar/pkg/util"
)

// bindingPolicyResolution stores the selected object identifiers and
// destinations and modulations for a single bindingpolicy. The mutex should be read locked
// before reading, and write locked before writing to any field.
type bindingPolicyResolution struct {
	sync.RWMutex

	objectIdentifierToData map[util.ObjectIdentifier]*objectData

	destinations sets.Set[string]

	// ownerReference identifies the bindingpolicy that this resolution is
	// associated with as an owning object.
	ownerReference *metav1.OwnerReference

	// requiresSingletonReportedState indicates whether the bindingpolicy
	// that this resolution is associated with requires singleton status.
	requiresSingletonReportedState bool
}

// objectData stores the UID, resource version, create-only bit,
// and statuscollectors for an object.
type objectData struct {
	UID              string
	ResourceVersion  string
	CreateOnly       bool
	StatusCollectors sets.Set[string]
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
	if objData == nil {
		resolution.objectIdentifierToData[objIdentifier] = &objectData{
			UID:              objUID,
			ResourceVersion:  resourceVersion,
			CreateOnly:       createOnly,
			StatusCollectors: statusCollectors,
		}
		return true
	}

	if objData.UID == objUID && objData.ResourceVersion == resourceVersion && objData.CreateOnly == createOnly && objData.StatusCollectors.Equal(statusCollectors) {
		return false
	}

	objData.UID = objUID
	objData.ResourceVersion = resourceVersion
	objData.CreateOnly = createOnly
	objData.StatusCollectors = statusCollectors

	return true
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
			workload.ClusterScope = append(workload.ClusterScope,
				v1alpha1.ClusterScopeDownsyncClause{
					ClusterScopeDownsyncObject: v1alpha1.ClusterScopeDownsyncObject{
						GroupVersionResource: metav1.GroupVersionResource(objIdentifier.GVR()),
						Name:                 objIdentifier.ObjectName.Name,
						ResourceVersion:      objData.ResourceVersion,
					},
					CreateOnly:       objData.CreateOnly,
					StatusCollectors: sets.List(objData.StatusCollectors),
				})

			continue
		}

		workload.NamespaceScope = append(workload.NamespaceScope,
			v1alpha1.NamespaceScopeDownsyncClause{
				NamespaceScopeDownsyncObject: v1alpha1.NamespaceScopeDownsyncObject{
					GroupVersionResource: metav1.GroupVersionResource(objIdentifier.GVR()),
					Name:                 objIdentifier.ObjectName.Name,
					Namespace:            objIdentifier.ObjectName.Namespace,
					ResourceVersion:      objData.ResourceVersion,
				},
				CreateOnly:       objData.CreateOnly,
				StatusCollectors: sets.List(objData.StatusCollectors),
			})
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

// getOwnerReference returns the owner reference of the resolution.
func (resolution *bindingPolicyResolution) getOwnerReference() metav1.OwnerReference {
	resolution.RLock()
	defer resolution.RUnlock()

	return *resolution.ownerReference
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

func bindingObjectRefToDataFromWorkload(bindingSpecWorkload *v1alpha1.DownsyncObjectClauses) map[objectRef]*objectData {
	bindingObjectRefToData := make(map[objectRef]*objectData)

	for _, clusterScopeDownsyncClause := range bindingSpecWorkload.ClusterScope {
		bindingObjectRefToData[objectRef{
			GroupVersionResource: schema.GroupVersionResource(clusterScopeDownsyncClause.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      clusterScopeDownsyncClause.Name,
				Namespace: metav1.NamespaceNone,
			},
		}] = &objectData{
			ResourceVersion:  clusterScopeDownsyncClause.ResourceVersion,
			CreateOnly:       clusterScopeDownsyncClause.CreateOnly,
			StatusCollectors: sets.New(clusterScopeDownsyncClause.StatusCollectors...),
		}
	}

	for _, namespacedScopeDownsyncClause := range bindingSpecWorkload.NamespaceScope {
		bindingObjectRefToData[objectRef{
			GroupVersionResource: schema.GroupVersionResource(namespacedScopeDownsyncClause.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      namespacedScopeDownsyncClause.Name,
				Namespace: namespacedScopeDownsyncClause.Namespace,
			},
		}] = &objectData{
			ResourceVersion:  namespacedScopeDownsyncClause.ResourceVersion,
			CreateOnly:       namespacedScopeDownsyncClause.CreateOnly,
			StatusCollectors: sets.New(namespacedScopeDownsyncClause.StatusCollectors...),
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
