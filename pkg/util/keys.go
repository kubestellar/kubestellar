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

package util

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

// Key struct used to add items to the workqueue. The key is used to identify
// the group/version/Kind of an object, used to index the listers for all
// objects, and the namespace/name key for the object. For deleted objects,
// since they are no longer in the cache, the key stores a shallow copy of the
// deleted object.
type Key struct {
	GVK            schema.GroupVersionKind
	NamespacedName cache.ObjectName
	DeletedObject  *runtime.Object
}

func (k *Key) GvkKey() string {
	if k.GVK.Group == "" {
		return fmt.Sprintf("%s/%s", k.GVK.Version, k.GVK.Kind)
	}

	return fmt.Sprintf("%s/%s/%s", k.GVK.Group, k.GVK.Version, k.GVK.Kind)
}

func (k *Key) GvkNamespacedNameKey() string {
	return fmt.Sprintf("%s/%s", k.GvkKey(), k.NamespacedName.String())
}

// Given an object that implements runtime.Object, create a key of type Key
// that contains the groupVersionKind key and name/namespace
func KeyForGroupVersionKindNamespaceName(obj any) (Key, error) {
	rObj := obj.(runtime.Object)
	ok := rObj.GetObjectKind()
	gvk := ok.GroupVersionKind()

	namespacedName, err := cache.ObjectToName(obj)
	if err != nil {
		return Key{}, err
	}
	key := Key{
		GVK:            gvk,
		NamespacedName: namespacedName,
	}

	return key, nil
}

// Create a string key in the form group/version/Kind or version/Kind if the group is empty
func KeyForGroupVersionKind(group, version, kind string) string {
	if group == "" {
		return fmt.Sprintf("%s/%s", version, kind)
	}

	return fmt.Sprintf("%s/%s/%s", group, version, kind)
}

// Create a string key in the form group/version/resource or version/resource if the group is empty
func KeyForGroupVersionResource(group, version, resource string) string {
	return KeyForGroupVersionKind(group, version, resource)
}

// KeyFromGVRandNS Creates a string key in the form
// group/version/Kind/namespace or version/Kind/namespace if the group is empty.
func KeyFromGVRandNS(gvr schema.GroupVersionResource, ns string) string {
	if gvr.Group == "" {
		return fmt.Sprintf("%s/%s/%s", gvr.Version, gvr.Resource, ns)
	}

	return fmt.Sprintf("%s/%s/%s/%s", gvr.Group, gvr.Version, gvr.Resource, ns)
}

// KeyFromGVKandNamespacedName Creates a string key in the form
// group/version/Kind/{namespaced-name} or version/Kind/{namespaced-name} if the group is empty.
func KeyFromGVKandNamespacedName(gvk schema.GroupVersionKind, name types.NamespacedName) string {
	if gvk.Group == "" {
		return fmt.Sprintf("%s/%s/%s", gvk.Version, gvk.Kind, name.String())
	}

	return fmt.Sprintf("%s/%s/%s/%s", gvk.Group, gvk.Version, gvk.Kind, name.String())
}

// Used for generating a single string unique representation of the object for logging info
func GenerateObjectInfoString(obj runtime.Object) string {
	group := obj.GetObjectKind().GroupVersionKind().Group
	kind := strings.ToLower(obj.GetObjectKind().GroupVersionKind().Kind)
	mObj := obj.(metav1.Object)

	prefix := kind
	if group != "" {
		prefix = fmt.Sprintf("%s.%s", kind, group)

	}

	return fmt.Sprintf("[%s] %s/%s", mObj.GetNamespace(), prefix, mObj.GetName())
}

func EmptyUnstructuredObjectFromKey(key Key) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}

	obj.SetGroupVersionKind(key.GVK)
	obj.SetNamespace(key.NamespacedName.Namespace)
	obj.SetName(key.NamespacedName.Name)

	return obj
}

func KeyIsForPlacementDecision(key Key) bool {
	if key.GVK.Group == v1alpha1.SchemeGroupVersion.Group &&
		key.GVK.Kind == PlacementDecisionKind {
		return true
	}

	return false
}
