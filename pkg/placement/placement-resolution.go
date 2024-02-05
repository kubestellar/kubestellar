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

package placement

import (
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// placementResolution stores the selected objects and destinations for a single placement.
// The mutex should be read locked before reading, and write locked before writing to any field.
type placementResolution struct {
	sync.RWMutex

	// map key is GVK/namespace/name as outputted by util.key.GvkNamespacedNameKey()
	objectIdentifierToKey map[string]*util.Key
	destinations          sets.Set[string]

	// workloadGeneration is a local counter that reflects that internal
	// changes occurred to the objects referenced in the objectIdentifierToKey
	// map. That means, this field is incremented whenever an already noted
	// object is called for noting again.
	workloadGeneration int64

	// ownerReference identifies the placement that this resolution is
	// associated with as an owning object.
	ownerReference *metav1.OwnerReference
}

// noteObject adds/deletes an object to/from the resolution.
// The return bool indicates whether the placement resolution was changed.
// This function is thread-safe.
func (resolution *placementResolution) noteObject(obj runtime.Object) (bool, error) {
	resolution.Lock()
	defer resolution.Unlock()

	key, err := util.KeyForGroupVersionKindNamespaceName(obj)
	if err != nil {
		return false, fmt.Errorf("failed to get key for object: %w", err)
	}

	// formatted-key to use for mapping
	formattedKey := key.GvkNamespacedNameKey()
	_, exists := resolution.objectIdentifierToKey[formattedKey]

	// avoid further processing for keys of objects being deleted that do not have a deleted object
	if isBeingDeleted(obj) {
		delete(resolution.objectIdentifierToKey, formattedKey)

		return exists, nil
	}

	if exists {
		// the object is already in the resolution but the object content changed
		resolution.workloadGeneration++
	}

	// add object to map
	resolution.objectIdentifierToKey[formattedKey] = &key
	// Internal changes to noted objects are also changes to the resolution.
	return true, nil
}

// removeObject deletes an object from the resolution if it exists.
// The return bool indicates whether the placement resolution was changed.
// This function is thread-safe.
func (resolution *placementResolution) removeObject(obj runtime.Object) bool {
	resolution.Lock()
	defer resolution.Unlock()

	key, err := util.KeyForGroupVersionKindNamespaceName(obj)
	if err != nil {
		return false // assume object was never added
	}

	// formatted-key to use for mapping
	formattedKey := key.GvkNamespacedNameKey()
	_, exists := resolution.objectIdentifierToKey[formattedKey]

	delete(resolution.objectIdentifierToKey, formattedKey)
	return exists
}

// setDestinations updates the destinations list in the resolution.
// The given destinations set is expected not to be mutated after this call.
func (resolution *placementResolution) setDestinations(destinations sets.Set[string]) {
	resolution.Lock()
	defer resolution.Unlock()

	resolution.destinations = destinations
}

// toPlacementDecisionSpec converts the resolution to a placement decision
// spec. This function is thread-safe.
func (resolution *placementResolution) toPlacementDecisionSpec(gvkGvrMapper util.GvkGvrMapper) (*v1alpha1.PlacementDecisionSpec, error) {
	resolution.RLock()
	defer resolution.RUnlock()

	workload := v1alpha1.DownsyncObjectReferences{
		WorkloadGeneration: resolution.workloadGeneration,
		// rest of fields calculated below
	}

	// the following optimize the building of the workload by maintaining pointers to the object-wrapper structs
	// that are keyed by GVR
	clusterScopeDownsyncObjectsMap := map[schema.GroupVersionResource]*v1alpha1.ClusterScopeDownsyncObjects{}
	// Since namespaceScopeDownsyncObjectsMap groups objects by NS, these maps are used to efficiently locate the
	// * struct by GVR
	// * objects slice by namespace (for GVR)
	namespaceScopeDownsyncObjectsMap := map[schema.GroupVersionResource]*v1alpha1.NamespaceScopeDownsyncObjects{}
	nsObjectsLocationInSlice := map[string]int{} // key is GVR/ns to avoid 2D maps

	// iterate over all objects and build workload efficiently. No (GVR, namespace, name) tuple is
	// duplicated in the objectIdentifierToKey map, due to the uniqueness of the Key. Therefore, whenever an object is about to
	// be appended to an objects slice, we simply append.
	for _, key := range resolution.objectIdentifierToKey {
		gvr, found := gvkGvrMapper.GetGvr(key.GVK)
		if !found {
			return nil, fmt.Errorf("failed to get GVR for GVK %s", key.GvkKey())
		}

		// check if object is cluster-scoped or namespaced by checking namespace
		if key.NamespacedName.Namespace == metav1.NamespaceNone {
			resolution.handleClusterScopedObject(gvr, key, &workload, clusterScopeDownsyncObjectsMap)
			continue
		}

		resolution.handleNamespacedObject(gvr, key, &workload, namespaceScopeDownsyncObjectsMap, nsObjectsLocationInSlice)
	}

	return &v1alpha1.PlacementDecisionSpec{
		Workload:     workload,
		Destinations: destinationsStringSetToDestinations(resolution.destinations),
	}, nil
}

// handleClusterScopedObject handles a cluster-scoped object by adding it to the workload.
// ClusterScopeDownsyncObjectsMap is used to efficiently locate the object in the workload.
func (resolution *placementResolution) handleClusterScopedObject(gvr schema.GroupVersionResource,
	key *util.Key,
	workload *v1alpha1.DownsyncObjectReferences,
	clusterScopeDownsyncObjectsMap map[schema.GroupVersionResource]*v1alpha1.ClusterScopeDownsyncObjects) {
	// check if obj GVR already exists in map
	if csdObjects, found := clusterScopeDownsyncObjectsMap[gvr]; found {
		// GVR exists, append cluster-scope object
		csdObjects.ObjectNames = append(csdObjects.ObjectNames, key.NamespacedName.Name)
		return
	}
	// GVR doesn't exist, this is the first time
	// add ClusterScopeDownsyncResource to the workload
	workload.ClusterScope = append(workload.ClusterScope, v1alpha1.ClusterScopeDownsyncObjects{
		GroupVersionResource: metav1.GroupVersionResource{
			Group:    gvr.Group,
			Version:  gvr.Version,
			Resource: gvr.Resource,
		},
		ObjectNames: []string{key.NamespacedName.Name},
	})

	// retain a pointer to the added ClusterScopeDownsyncResource for efficiency
	clusterScopeDownsyncObjectsMap[gvr] = &workload.ClusterScope[len(workload.ClusterScope)-1]
}

// handleNamespacedObject handles a namespaced object by adding it to the workload.
// namespaceScopeDownsyncObjectsMap and nsObjectsLocationInSlice are used to efficiently locate the object in the workload.
func (resolution *placementResolution) handleNamespacedObject(gvr schema.GroupVersionResource,
	key *util.Key,
	workload *v1alpha1.DownsyncObjectReferences,
	namespaceScopeDownsyncObjectsMap map[schema.GroupVersionResource]*v1alpha1.NamespaceScopeDownsyncObjects,
	nsObjectsLocationInSlice map[string]int) {
	gvrAndNSKey := util.KeyFromGVRandNS(gvr, key.NamespacedName.Namespace)
	if nsdObjects, found := namespaceScopeDownsyncObjectsMap[gvr]; found {
		// GVR mapping is found, check NS mapping
		if nsIdx, found := nsObjectsLocationInSlice[gvrAndNSKey]; found {
			// NS mapping is found, append name
			nsdObjects.ObjectsByNamespace[nsIdx].Names = append(nsdObjects.ObjectsByNamespace[nsIdx].Names,
				key.NamespacedName.Name)
			return
		}

		// namespace mapping is not found, create new (ns, names) entry for object
		nsdObjects.ObjectsByNamespace = append(nsdObjects.ObjectsByNamespace, v1alpha1.NamespaceAndNames{
			Namespace: key.NamespacedName.Namespace,
			Names:     []string{key.NamespacedName.Name},
		})

		// update index mapping
		nsObjectsLocationInSlice[gvrAndNSKey] = len(nsdObjects.ObjectsByNamespace) - 1
		return
	}

	// GVR mapping isn't found, create new entry for this obj (and NS map)
	// add namespaceScopeDownsyncObjectsMap to the workload
	workload.NamespaceScope = append(workload.NamespaceScope, v1alpha1.NamespaceScopeDownsyncObjects{
		GroupVersionResource: metav1.GroupVersionResource{
			Group:    gvr.Group,
			Version:  gvr.Version,
			Resource: gvr.Resource,
		},
		ObjectsByNamespace: []v1alpha1.NamespaceAndNames{
			{
				Namespace: key.NamespacedName.Namespace,
				Names:     []string{key.NamespacedName.Name},
			},
		},
	})

	// retain a pointer to the added namespaceScopeDownsyncObjectsMap for efficiency
	namespaceScopeDownsyncObjectsMap[gvr] = &workload.NamespaceScope[len(workload.NamespaceScope)-1]
	// update ns mapping
	nsObjectsLocationInSlice[gvrAndNSKey] = 0 // first entry
}

func (resolution *placementResolution) matchesPlacementDecisionSpec(placementDecisionSpec *v1alpha1.PlacementDecisionSpec,
	gvkGvrMapper util.GvkGvrMapper) bool {
	resolution.RLock()
	defer resolution.RUnlock()

	// check workloadGeneration
	if resolution.workloadGeneration != placementDecisionSpec.Workload.WorkloadGeneration {
		return false
	}

	// check destinations
	if !destinationsMatch(resolution.destinations, placementDecisionSpec.Destinations) {
		return false
	}

	// check workload
	return workloadMatchesPlacementDecisionSpec(&placementDecisionSpec.Workload, resolution.objectIdentifierToKey, gvkGvrMapper)
}

// destinationsMatch returns true if the destinations in the resolution
// match the destinations in the placement decision spec.
func destinationsMatch(resolvedDestinations sets.Set[string], placementDecisionDestinations []v1alpha1.Destination) bool {
	if len(resolvedDestinations) != len(placementDecisionDestinations) {
		return false
	}

	for _, destination := range placementDecisionDestinations {
		if !resolvedDestinations.Has(destination.ClusterId) {
			return false
		}
	}

	return true
}

// workloadMatchesPlacementDecisionSpec returns true if the workload in the
// resolution matches the workload in the placement decision spec.
func workloadMatchesPlacementDecisionSpec(placementDecisionSpecWorkload *v1alpha1.DownsyncObjectReferences,
	objectIdentifierToKeyMap map[string]*util.Key, gvkGvrMapper util.GvkGvrMapper) bool {
	// check lengths
	clusterScopedObjectsCount := 0
	for _, clusterScopeDownsyncObjects := range placementDecisionSpecWorkload.ClusterScope {
		clusterScopedObjectsCount += len(clusterScopeDownsyncObjects.ObjectNames)
	}

	namespacedObjectsCount := 0
	for _, namespaceScopeDownsyncObjects := range placementDecisionSpecWorkload.NamespaceScope {
		for _, objectsByNamespace := range namespaceScopeDownsyncObjects.ObjectsByNamespace {
			namespacedObjectsCount += len(objectsByNamespace.Names)
		}
	}

	if len(objectIdentifierToKeyMap) != clusterScopedObjectsCount+namespacedObjectsCount {
		return false
	}

	// again we can check match by making sure all objects are mapped, since entries are unique and the length is equal.
	// check cluster-scoped all exist
	if !placementDecisionClusterScopeIsMapped(placementDecisionSpecWorkload.ClusterScope, objectIdentifierToKeyMap,
		gvkGvrMapper) {
		return false
	}

	// check namespace-scoped all exist
	return namespaceScopeMatchesPlacementDecisionSpec(placementDecisionSpecWorkload.NamespaceScope, objectIdentifierToKeyMap,
		gvkGvrMapper)
}

// clusterScopeMatchesPlacementDecisionSpec returns true if the cluster-scope
// section in the placement decision spec all exist in the resolution.
func placementDecisionClusterScopeIsMapped(placementDecisionSpecClusterScope []v1alpha1.ClusterScopeDownsyncObjects,
	objectIdentifierToKeyMap map[string]*util.Key, gvkGvrMapper util.GvkGvrMapper) bool {
	for _, clusterScopeDownsyncObjects := range placementDecisionSpecClusterScope {
		gvr := schema.GroupVersionResource(clusterScopeDownsyncObjects.GroupVersionResource)
		gvk, found := gvkGvrMapper.GetGvk(gvr)

		if !found {
			return false // if not found then not mapped and not in resolution
		}

		for _, objName := range clusterScopeDownsyncObjects.ObjectNames {
			formattedKey := util.KeyFromGVKandNamespacedName(gvk, types.NamespacedName{
				Namespace: metav1.NamespaceNone,
				Name:      objName,
			})

			if _, found := objectIdentifierToKeyMap[formattedKey]; !found {
				return false
			}
		}
	}

	return true
}

// namespaceScopeMatchesPlacementDecisionSpec returns true if the namespace-scope
// section in the placement decision spec all exist in the resolution.
func namespaceScopeMatchesPlacementDecisionSpec(placementDecisionSpecNamespaceScope []v1alpha1.NamespaceScopeDownsyncObjects,
	objectIdentifierToKeyMap map[string]*util.Key, gvkGvrMapper util.GvkGvrMapper) bool {
	for _, namespaceScopeDownsyncObjects := range placementDecisionSpecNamespaceScope {
		gvr := schema.GroupVersionResource(namespaceScopeDownsyncObjects.GroupVersionResource)
		gvk, found := gvkGvrMapper.GetGvk(gvr)

		if !found {
			return false // if GVK mapping is not found then we cant know if this object is
			// mapped or not, therefore returning false is the safe (and correct) option
		}

		// iterate over namespaced objects and check if they are mapped
		for _, nsAndNames := range namespaceScopeDownsyncObjects.ObjectsByNamespace {
			for _, objName := range nsAndNames.Names {
				formattedKey := util.KeyFromGVKandNamespacedName(gvk, types.NamespacedName{
					Namespace: nsAndNames.Namespace,
					Name:      objName,
				})

				if _, found := objectIdentifierToKeyMap[formattedKey]; !found {
					return false
				}
			}
		}
	}

	return true
}

func destinationsStringSetToDestinations(destinations sets.Set[string]) []v1alpha1.Destination {
	dests := make([]v1alpha1.Destination, 0, len(destinations))
	for d := range destinations {
		dests = append(dests, v1alpha1.Destination{ClusterId: d})
	}

	return dests
}
