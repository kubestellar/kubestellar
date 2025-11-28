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
	"slices"
	"strings"
	"sync"

	"github.com/go-logr/logr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/abstract"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// bindingPolicyResolution stores the selected object identifiers and
// destinations and modulations for a single bindingpolicy. The mutex should be read locked
// before reading, and write locked before writing to any field.
type bindingPolicyResolution struct {
	// One immutable function that gets called synchronously whenever there is a change
	// in the requiresSingletonReportedState or requiresMultiWECReportedState setting for an object.
	reportedStateRequestChangeConsumer func(util.ObjectIdentifier)

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
}

// ObjectData stores most of the downsync modalities for a given workload object
// with respect to a given Binding.
type ObjectData struct {
	// UID is the UID of the workload object.
	UID string
	// ResourceVersion is the resource version of the workload object.
	ResourceVersion string

	Modulation DownsyncModulation
}

// Assert that `*bindingPolicyResolution` implements Resolution
var _ Resolution = &bindingPolicyResolution{}

var _ logr.Marshaler = &bindingPolicyResolution{}

func (resolution *bindingPolicyResolution) MarshalLog() any {
	return map[string]any{
		"objectIdentifierToData": util.PrimitiveMap4Log(resolution.objectIdentifierToData),
		"destinations":           resolution.destinations,
		"ownerReference":         resolution.ownerReference,
	}
}

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
	objUID, resourceVersion string, modulation DownsyncModulation) bool {
	resolution.Lock()
	defer resolution.Unlock()

	objData := resolution.objectIdentifierToData[objIdentifier]
	if objData == nil || objData.UID != objUID || objData.ResourceVersion != resourceVersion ||
		!objData.Modulation.Equal(modulation) {
		resolution.objectIdentifierToData[objIdentifier] = &ObjectData{
			UID:             objUID,
			ResourceVersion: resourceVersion,
			Modulation:      modulation,
		}
		// Notify when singleton or multi-WEC status flags change
		singletonChanged := objData == nil && modulation.WantSingletonReportedState ||
			objData != nil && objData.Modulation.WantSingletonReportedState != modulation.WantSingletonReportedState
		multiWECChanged := objData == nil && modulation.WantMultiWECReportedState ||
			objData != nil && objData.Modulation.WantMultiWECReportedState != modulation.WantMultiWECReportedState

		if singletonChanged || multiWECChanged {
			klog.InfoS("Noting addition/change of object to resolution", "resolution", fmt.Sprintf("%p", resolution), "objId", objIdentifier)
			resolution.reportedStateRequestChangeConsumer(objIdentifier)
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

	objData, exists := resolution.objectIdentifierToData[objIdentifier]
	if !exists {
		return false
	}

	delete(resolution.objectIdentifierToData, objIdentifier)
	if objData.Modulation.WantSingletonReportedState || objData.Modulation.WantMultiWECReportedState {
		klog.InfoS("Noting removal of object from resolution", "resolution", fmt.Sprintf("%p", resolution), "objId", objIdentifier)
		resolution.reportedStateRequestChangeConsumer(objIdentifier)
	}
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
				DownsyncModulation: objData.Modulation.ToExternal(),
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
			DownsyncModulation: objData.Modulation.ToExternal(),
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
			!objData.Modulation.Equal(objDataFromWorkload.Modulation) {
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

// getSingletonReportedStateRequestForObject returns what this resolution requests regarding singleton reported state return.
// The first returned bool reports whether this resolution matches the given workload object ID;
// if not then the other returned values are meaningless.
// The second returned bool indicates whether singleton reported state return is requested.
// The returned set is the names of the matching WECs, and is immutable.
func (resolution *bindingPolicyResolution) getSingletonReportedStateRequestForObject(objId util.ObjectIdentifier) (bool, bool, sets.Set[string]) {
	resolution.RLock()
	defer resolution.RUnlock()
	if objData, has := resolution.objectIdentifierToData[objId]; has {
		return true, objData.Modulation.WantSingletonReportedState, resolution.destinations
	}
	return false, false, nil
}

// getMultiWECReportedStateRequestForObject returns what this resolution requests regarding multi-wec reported state return.
// The first returned bool reports whether this resolution matches the given workload object ID;
// if not then the other returned values are meaningless.
// The second returned bool indicates whether multi-wec reported state return is requested.
// The returned set is the names of the matching WECs, and is immutable.
func (resolution *bindingPolicyResolution) getMultiWECReportedStateRequestForObject(objId util.ObjectIdentifier) (bool, bool, sets.Set[string]) {
	resolution.RLock()
	defer resolution.RUnlock()
	if objData, has := resolution.objectIdentifierToData[objId]; has {
		return true, objData.Modulation.WantMultiWECReportedState, resolution.destinations
	}
	return false, false, nil
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
		bindingObjectRefToData[objectRef{
			GroupVersionResource: schema.GroupVersionResource(clusterScopeDownsyncClause.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      clusterScopeDownsyncClause.Name,
				Namespace: metav1.NamespaceNone,
			},
		}] = &ObjectData{
			ResourceVersion: clusterScopeDownsyncClause.ResourceVersion,
			Modulation:      DownsyncModulationFromExternal(clusterScopeDownsyncClause.DownsyncModulation),
		}
	}

	for _, namespacedScopeDownsyncClause := range bindingSpecWorkload.NamespaceScope {
		bindingObjectRefToData[objectRef{
			GroupVersionResource: schema.GroupVersionResource(namespacedScopeDownsyncClause.GroupVersionResource),
			ObjectName: cache.ObjectName{
				Name:      namespacedScopeDownsyncClause.Name,
				Namespace: namespacedScopeDownsyncClause.Namespace,
			},
		}] = &ObjectData{
			ResourceVersion: namespacedScopeDownsyncClause.ResourceVersion,
			Modulation:      DownsyncModulationFromExternal(namespacedScopeDownsyncClause.DownsyncModulation),
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
	slices.SortFunc(bindingWorkload.ClusterScope, func(a, b v1alpha1.ClusterScopeDownsyncClause) int {
		if cmp := strings.Compare(a.GroupVersionResource.String(), b.GroupVersionResource.String()); cmp != 0 {
			return cmp
		}
		if cmp := strings.Compare(a.Name, b.Name); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.ResourceVersion, b.ResourceVersion)
	})
	// sort namespaceScopeDownsyncObjects
	slices.SortFunc(bindingWorkload.NamespaceScope, func(a, b v1alpha1.NamespaceScopeDownsyncClause) int {
		if cmp := strings.Compare(a.GroupVersionResource.String(), b.GroupVersionResource.String()); cmp != 0 {
			return cmp
		}
		objectNameA := cache.NewObjectName(a.Namespace, a.Name).String()
		objectNameB := cache.NewObjectName(b.Namespace, b.Name).String()
		if cmp := strings.Compare(objectNameA, objectNameB); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.ResourceVersion, b.ResourceVersion)
	})
}
