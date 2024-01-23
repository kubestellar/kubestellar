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

	"github.com/kubestellar/kubestellar/api/edge/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const clusterScopedObjectNamespace = ""

// decisionData stores the placement resolution (selected objects and destinations) for a single placement.
// The mutex should be read locked before reading, and write locked before writing to any field.
// TODO: consider "protecting" fields through interfacing.
type decisionData struct {
	sync.RWMutex
	placementKey types.NamespacedName

	mappedObjects map[string]*util.Key // map key is GVK/namespace/name as outputted by util.key.GvkNamespacedNameKey()

	destinations []string
}

// updateObject adds/deletes an object to/from the decision data.
// The return bool indicates whether the placement decision resolution was changed.
// This function is thread-safe.
func (dd *decisionData) updateObject(obj runtime.Object) (bool, error) {
	dd.Lock()
	defer dd.Unlock()

	key, err := util.KeyForGroupVersionKindNamespaceName(obj)
	if err != nil {
		return false, fmt.Errorf("failed to get key for object: %v", err)
	}

	// formatted-key to use for mapping
	formattedKey := key.GvkNamespacedNameKey()
	_, exists := dd.mappedObjects[formattedKey]

	// avoid further processing for keys of objects being deleted that do not have a deleted object
	if isBeingDeleted(obj) || key.DeletedObject != nil {
		delete(dd.mappedObjects, formattedKey)

		return exists, nil
	}

	// add object to map
	dd.mappedObjects[key.GvkNamespacedNameKey()] = nil
	// if an object is reconciled all the way into here, it means something internal changed, so we need to trigger
	// a placement-decision update event in order for transport watchers to re-evaluate it
	return true, nil
}

// updateDestination updates the destinations list in the decision data.
// This function is thread-safe.
func (dd *decisionData) updateDestinations(destinations []string) {
	dd.Lock()
	defer dd.Unlock()

	dd.destinations = destinations
}

// toPlacementDecisionSpec converts the decision data to a placement decision
// spec. This function is thread-safe.
func (dd *decisionData) toPlacementDecisionSpec(gvkGvrMapper GvkGvrMapper) (*v1alpha1.PlacementDecisionSpec, error) {
	dd.RLock()
	defer dd.RUnlock()

	workload := v1alpha1.DownsyncObjectsReferences{}

	// the following optimize the building of the workload by maintaining pointers to the object-wrapper structs
	// that are key'd by GVR
	clusterScopeDownsyncResources := map[schema.GroupVersionResource]*v1alpha1.ClusterScopeDownsyncObject{}
	// Since namespaceScopeDownsyncObjects groups objects by NS, these maps are used to efficiently locate the
	// * struct by GVR
	// * objects slice by namespace (for GVR)
	namespaceScopeDownsyncObjects := map[schema.GroupVersionResource]*v1alpha1.NamespaceScopeDownsyncObjects{}
	nsObjectsLocationInSlice := map[string]int{} // key is GVR/ns to avoid 2D maps

	// iterate over all objects and build workload efficiently. We assume that no (GVR, namespace, name) tuple is
	// duplicated in the mappedObjects map, due to the uniqueness of the Key. Therefore, whenever an object is about to
	// be appended to an objects slice, we simply append.
	for _, key := range dd.mappedObjects {
		gvr, found := gvkGvrMapper.GetGvr(key.GVK)
		if !found {
			return nil, fmt.Errorf("failed to get GVR for GVK %s", key.GvkKey())
		}

		// check if object is cluster-scoped or namespaced by checking namespace
		if key.NamespacedName.Namespace == clusterScopedObjectNamespace {
			dd.handleClusterScopedObject(*gvr, key, &workload, clusterScopeDownsyncResources)
			continue
		}

		// object is namespaced, check if obj GVR is already mapped
		dd.handleNamespacedObject(*gvr, key, &workload, namespaceScopeDownsyncObjects, nsObjectsLocationInSlice)
	}

	return &v1alpha1.PlacementDecisionSpec{
		Workload:     workload,
		Destinations: destinationsStringSliceToDestinations(dd.destinations),
	}, nil
}

// handleClusterScopedObject handles a cluster-scoped object by adding it to the workload.
// clusterScopeDownsyncResources is used to efficiently locate the object in the workload.
func (dd *decisionData) handleClusterScopedObject(gvr schema.GroupVersionResource,
	key *util.Key,
	workload *v1alpha1.DownsyncObjectsReferences,
	clusterScopeDownsyncResources map[schema.GroupVersionResource]*v1alpha1.ClusterScopeDownsyncObject) {
	// check if obj GVR already exists in map
	if csdResource, found := clusterScopeDownsyncResources[gvr]; found {
		// GVR exists, append cluster-scope object
		csdResource.ObjectNames = append(csdResource.ObjectNames, key.NamespacedName.Name)
		return
	}
	// GVR doesn't exist, this is the first time
	// add ClusterScopeDownsyncResource to the workload
	workload.ClusterScope = append(workload.ClusterScope, v1alpha1.ClusterScopeDownsyncObject{
		GroupVersionResource: metav1.GroupVersionResource{
			Group:    gvr.Group,
			Version:  gvr.Version,
			Resource: gvr.Resource,
		},
		ObjectNames: []string{key.NamespacedName.Name},
	})

	// retain a pointer to the added ClusterScopeDownsyncResource for efficiency
	clusterScopeDownsyncResources[gvr] = &workload.ClusterScope[len(workload.ClusterScope)-1]
}

// handleNamespacedObject handles a namespaced object by adding it to the workload.
// namespaceScopeDownsyncObjects and nsObjectsLocationInSlice are used to efficiently locate the object in the workload.
func (dd *decisionData) handleNamespacedObject(gvr schema.GroupVersionResource,
	key *util.Key,
	workload *v1alpha1.DownsyncObjectsReferences,
	namespaceScopeDownsyncObjects map[schema.GroupVersionResource]*v1alpha1.NamespaceScopeDownsyncObjects,
	nsObjectsLocationInSlice map[string]int) {
	if nsdResource, found := namespaceScopeDownsyncObjects[gvr]; found {
		// GVR mapping is found, check NS mapping
		gvrAndNSKey := util.KeyFromGVRandNS(gvr, key.NamespacedName.Namespace) // key for ns mapping inside

		if nsIdx, found := nsObjectsLocationInSlice[gvrAndNSKey]; found {
			// NS mapping is found, append name
			nsdResource.ObjectsByNamespace[nsIdx].Names = append(nsdResource.ObjectsByNamespace[nsIdx].Names,
				key.NamespacedName.Name)
			return
		}

		// namespace mapping is not found, create new (ns, names) entry for object
		nsdResource.ObjectsByNamespace = append(nsdResource.ObjectsByNamespace, v1alpha1.NamespaceAndNames{
			Namespace: key.NamespacedName.Namespace,
			Names:     []string{key.NamespacedName.Name},
		})

		// update index mapping
		nsObjectsLocationInSlice[gvrAndNSKey] = len(nsdResource.ObjectsByNamespace) - 1
	}

	// GVR mapping isn't found, create new entry for this obj (and NS map)
	// add NamespaceScopeDownsyncObjects to the workload
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

	// retain a pointer to the added NamespaceScopeDownsyncObjects for efficiency
	namespaceScopeDownsyncObjects[gvr] = &workload.NamespaceScope[len(workload.NamespaceScope)-1]
	// update ns mapping
	nsObjectsLocationInSlice[util.KeyFromGVRandNS(gvr, key.NamespacedName.Namespace)] = 0 // first entry
}

func (dd *decisionData) matchesPlacementDecisionSpec(placementDecisionSpec *v1alpha1.PlacementDecisionSpec,
	gvkGvrMapper GvkGvrMapper) bool {
	dd.RLock()
	defer dd.RUnlock()

	// check destinations
	if !destinationsMatch(dd.destinations, placementDecisionSpec.Destinations) {
		return false
	}

	// check workload
	return workloadMatchesPlacementDecisionSpec(&placementDecisionSpec.Workload, dd.mappedObjects, gvkGvrMapper)
}

// destinationsMatch returns true if the destinations in the decision data
// match the destinations in the placement decision spec.
// The function performs one pass over both slices, sequentially.
func destinationsMatch(destinations []string, placementDecisionDestinations []v1alpha1.Destination) bool {
	if len(destinations) != len(placementDecisionDestinations) {
		return false
	}

	// calculate match by counting matching objects then comparing to length.
	// this works since entries are unique.
	matches := map[string]interface{}{}
	matchesCounter := 0

	// fill up matches
	for _, destination := range placementDecisionDestinations {
		matches[destination.ClusterId] = nil
	}

	for _, destination := range destinations {
		if _, found := matches[destination]; !found {
			return false
		}

		matchesCounter++
	}

	return len(placementDecisionDestinations) == matchesCounter
}

// workloadMatchesPlacementDecisionSpec returns true if the workload in the
// decision data matches the workload in the placement decision spec.
func workloadMatchesPlacementDecisionSpec(placementDecisionSpecWorkload *v1alpha1.DownsyncObjectsReferences,
	mappedObjects map[string]*util.Key, gvkGvrMapper GvkGvrMapper) bool {
	// check lengths
	clusterScopedObjectsCount := 0
	for _, clusterScopeDownsyncResource := range placementDecisionSpecWorkload.ClusterScope {
		clusterScopedObjectsCount += len(clusterScopeDownsyncResource.ObjectNames)
	}

	namespacedObjectsCount := 0
	for _, namespaceScopeDownsyncObjects := range placementDecisionSpecWorkload.NamespaceScope {
		namespacedObjectsCount += len(namespaceScopeDownsyncObjects.ObjectsByNamespace)
	}

	if len(mappedObjects) != clusterScopedObjectsCount+namespacedObjectsCount {
		return false
	}

	// again we can check match by making sure all objects are mapped, since entries are unique and the length is equal.
	// check cluster-scoped all exist
	if !placementDecisionClusterScopeIsMapped(placementDecisionSpecWorkload.ClusterScope, mappedObjects,
		gvkGvrMapper) {
		return false
	}

	// check namespace-scoped all exist
	return namespaceScopeMatchesPlacementDecisionSpec(placementDecisionSpecWorkload.NamespaceScope, mappedObjects,
		gvkGvrMapper)
}

// clusterScopeMatchesPlacementDecisionSpec returns true if the cluster-scope
// section in the placement decision spec all exist in the decision data.
func placementDecisionClusterScopeIsMapped(placementDecisionSpecClusterScope []v1alpha1.ClusterScopeDownsyncObject,
	mappedObjects map[string]*util.Key, gvkGvrMapper GvkGvrMapper) bool {
	for _, clusterScopeDownsyncResource := range placementDecisionSpecClusterScope {
		for _, objName := range clusterScopeDownsyncResource.ObjectNames {
			gvk, found := gvkGvrMapper.GetGvk(schema.GroupVersionResource(clusterScopeDownsyncResource.GroupVersionResource))

			if !found {
				return false // assume obj not mapped
			}

			formattedKey := util.KeyFromGVKandNamespacedName(*gvk, types.NamespacedName{
				Namespace: clusterScopedObjectNamespace,
				Name:      objName,
			})

			if _, found := mappedObjects[formattedKey]; !found {
				return false
			}
		}
	}

	return true
}

// namespaceScopeMatchesPlacementDecisionSpec returns true if the namespace-scope
// section in the placement decision spec all exist in the decision data.
func namespaceScopeMatchesPlacementDecisionSpec(placementDecisionSpecNamespaceScope []v1alpha1.NamespaceScopeDownsyncObjects,
	mappedObjects map[string]*util.Key, gvkGvrMapper GvkGvrMapper) bool {
	for _, namespaceScopeDownsyncObjects := range placementDecisionSpecNamespaceScope {
		gvk, found := gvkGvrMapper.GetGvk(schema.GroupVersionResource(namespaceScopeDownsyncObjects.GroupVersionResource))

		if !found {
			return false // assume obj not mapped
		}

		// iterate over namespaced objects and check if they are mapped
		for _, nsAndNames := range namespaceScopeDownsyncObjects.ObjectsByNamespace {
			for _, objName := range nsAndNames.Names {
				formattedKey := util.KeyFromGVKandNamespacedName(*gvk, types.NamespacedName{
					Namespace: nsAndNames.Namespace,
					Name:      objName,
				})

				if _, found := mappedObjects[formattedKey]; !found {
					return false
				}
			}
		}
	}

	return true
}
