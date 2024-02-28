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
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// bindingPolicyResolution stores the selected objects and destinations for a single bindingpolicy.
// The mutex should be read locked before reading, and write locked before writing to any field.
type bindingPolicyResolution struct {
	sync.RWMutex

	// map key is GVK/namespace/name as outputted by util.key.GvkNamespacedNameKey()
	objectIdentifierToKey     map[string]*util.Key
	objectIdentifierToVersion map[string]string // hold resource version for each entry in the above map
	// TODO: unify maps through reworking the key

	destinations sets.Set[string]

	// ownerReference identifies the bindingpolicy that this resolution is
	// associated with as an owning object.
	ownerReference *metav1.OwnerReference

	// requiresSingletonReportedState indicates whether the bindingpolicy
	// that this resolution is associated with requires singleton status.
	requiresSingletonReportedState bool
}

// noteObject adds/deletes an object to/from the resolution.
// The return bool indicates whether the bindingpolicy resolution was changed.
// This function is thread-safe.
func (resolution *bindingPolicyResolution) noteObject(obj runtime.Object) (bool, error) {
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
		delete(resolution.objectIdentifierToVersion, formattedKey)

		return exists, nil
	}

	// add object to map
	resolution.objectIdentifierToKey[formattedKey] = &key
	resolution.objectIdentifierToVersion[formattedKey] = obj.(metav1.Object).GetResourceVersion()

	// Internal changes to noted objects are also changes to the resolution.
	return true, nil
}

// removeObject deletes an object from the resolution if it exists.
// The return bool indicates whether the bindingpolicy resolution was changed.
// This function is thread-safe.
func (resolution *bindingPolicyResolution) removeObject(obj runtime.Object) bool {
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
	delete(resolution.objectIdentifierToVersion, formattedKey)

	return exists
}

// setDestinations updates the destinations list in the resolution.
// The given destinations set is expected not to be mutated after this call.
func (resolution *bindingPolicyResolution) setDestinations(destinations sets.Set[string]) {
	resolution.Lock()
	defer resolution.Unlock()

	resolution.destinations = destinations
}

// getObjectKeys returns the Keys of the objects in the resolution.
func (resolution *bindingPolicyResolution) getObjectKeys() []*util.Key {
	resolution.RLock()
	defer resolution.RUnlock()

	keys := make([]*util.Key, 0, len(resolution.objectIdentifierToKey))
	for _, key := range resolution.objectIdentifierToKey {
		keys = append(keys, key)
	}

	return keys
}

// toBindingSpec converts the resolution to a binding
// spec. This function is thread-safe.
func (resolution *bindingPolicyResolution) toBindingSpec(gvkGvrMapper util.GvkGvrMapper) (*v1alpha1.BindingSpec, error) {
	resolution.RLock()
	defer resolution.RUnlock()

	workload := v1alpha1.DownsyncObjectReferences{}

	// iterate over all objects and build workload efficiently. No (GVR, namespace, name) tuple is
	// duplicated in the objectIdentifierToKey map, due to the uniqueness of the Key.
	// Therefore, whenever an object is about to be appended to an objects slice, we simply append.
	for identifier, key := range resolution.objectIdentifierToKey {
		gvr, found := gvkGvrMapper.GetGvr(key.GVK)
		if !found {
			return nil, fmt.Errorf("failed to get GVR for GVK %s", key.GvkKey())
		}

		// check if object is cluster-scoped or namespaced by checking namespace
		if key.NamespacedName.Namespace == metav1.NamespaceNone {
			workload.ClusterScope = append(workload.ClusterScope, v1alpha1.ClusterScopeDownsyncObject{
				GroupVersionResource: metav1.GroupVersionResource{
					Group:    gvr.Group,
					Version:  gvr.Version,
					Resource: gvr.Resource,
				},
				Name:            key.NamespacedName.Name,
				ResourceVersion: resolution.objectIdentifierToVersion[identifier], // necessarily exists
			})
			continue
		}

		workload.NamespaceScope = append(workload.NamespaceScope, v1alpha1.NamespaceScopeDownsyncObject{
			GroupVersionResource: metav1.GroupVersionResource{
				Group:    gvr.Group,
				Version:  gvr.Version,
				Resource: gvr.Resource,
			},
			Namespace:       key.NamespacedName.Namespace,
			Name:            key.NamespacedName.Name,
			ResourceVersion: resolution.objectIdentifierToVersion[identifier], // necessarily exists
		})
	}

	return &v1alpha1.BindingSpec{
		Workload:     workload,
		Destinations: destinationsStringSetToDestinations(resolution.destinations),
	}, nil
}

func (resolution *bindingPolicyResolution) matchesBindingSpec(bindingSpec *v1alpha1.BindingSpec,
	gvkGvrMapper util.GvkGvrMapper) bool {
	resolution.RLock()
	defer resolution.RUnlock()

	// check destinations
	if !destinationsMatch(resolution.destinations, bindingSpec.Destinations) {
		return false
	}

	// check workload
	return workloadMatchesBindingSpec(&bindingSpec.Workload, resolution.objectIdentifierToKey,
		resolution.objectIdentifierToVersion, gvkGvrMapper)
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

// workloadMatchesBindingSpec returns true if the workload in the
// resolution matches the workload in the binding spec.
func workloadMatchesBindingSpec(bindingSpecWorkload *v1alpha1.DownsyncObjectReferences,
	objectIdentifierToKeyMap map[string]*util.Key, objectIdentifierToVersionMap map[string]string,
	gvkGvrMapper util.GvkGvrMapper) bool {
	// check lengths
	if len(objectIdentifierToKeyMap) != len(bindingSpecWorkload.ClusterScope)+len(bindingSpecWorkload.NamespaceScope) {
		return false
	}

	// again we can check match by making sure all objects are mapped, since entries are unique and the length is equal.
	// check cluster-scoped all exist
	if !bindingClusterScopeIsMapped(bindingSpecWorkload.ClusterScope, objectIdentifierToKeyMap,
		objectIdentifierToVersionMap, gvkGvrMapper) {
		return false
	}

	// check namespace-scoped all exist
	return bindingNamespaceScopeIsMapped(bindingSpecWorkload.NamespaceScope, objectIdentifierToKeyMap,
		objectIdentifierToVersionMap, gvkGvrMapper)
}

// bindingClusterScopeIsMapped returns true if the cluster-scope
// section in the binding spec all exist in the resolution.
func bindingClusterScopeIsMapped(bindingSpecClusterScope []v1alpha1.ClusterScopeDownsyncObject,
	objectIdentifierToKeyMap map[string]*util.Key, objectIdentifierToVersionMap map[string]string,
	gvkGvrMapper util.GvkGvrMapper) bool {
	for _, clusterScopeDownsyncObject := range bindingSpecClusterScope {
		gvr := schema.GroupVersionResource(clusterScopeDownsyncObject.GroupVersionResource)
		gvk, found := gvkGvrMapper.GetGvk(gvr)

		if !found {
			return false // if not found then not mapped and not in resolution
		}

		formattedKey := util.KeyFromGVKandNamespacedName(gvk, types.NamespacedName{
			Namespace: metav1.NamespaceNone,
			Name:      clusterScopeDownsyncObject.Name,
		})

		if _, found := objectIdentifierToKeyMap[formattedKey]; !found {
			return false
		}

		// check if version matches (if the key is mapped above, then the version should also be mapped)
		if objectIdentifierToVersionMap[formattedKey] != clusterScopeDownsyncObject.ResourceVersion {
			return false
		}
	}

	return true
}

// namespaceScopeMatchesBindingSpec returns true if the namespace-scope
// section in the binding spec all exist in the resolution.
func bindingNamespaceScopeIsMapped(bindingSpecNamespaceScope []v1alpha1.NamespaceScopeDownsyncObject,
	objectIdentifierToKeyMap map[string]*util.Key, objectIdentifierToVersionMap map[string]string,
	gvkGvrMapper util.GvkGvrMapper) bool {
	for _, namespaceScopeDownsyncObject := range bindingSpecNamespaceScope {
		gvr := schema.GroupVersionResource(namespaceScopeDownsyncObject.GroupVersionResource)
		gvk, found := gvkGvrMapper.GetGvk(gvr)

		if !found {
			return false // if GVK mapping is not found then we cant know if this object is
			// mapped or not, therefore returning false is the safe (and correct) option
		}

		formattedKey := util.KeyFromGVKandNamespacedName(gvk, types.NamespacedName{
			Namespace: namespaceScopeDownsyncObject.Namespace,
			Name:      namespaceScopeDownsyncObject.Name,
		})

		if _, found := objectIdentifierToKeyMap[formattedKey]; !found {
			return false
		}

		// check if version matches (if the key is mapped above, then the version is also mapped)
		if objectIdentifierToVersionMap[formattedKey] != namespaceScopeDownsyncObject.ResourceVersion {
			return false
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
