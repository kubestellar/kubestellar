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

package util

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

// ObjectIdentifier struct is used to add items to the workqueue.
// The ObjectIdentifier contains all the necessary information to retrieve an object.
type ObjectIdentifier struct {
	GVK        schema.GroupVersionKind
	Resource   string
	ObjectName cache.ObjectName
}

func (identifier *ObjectIdentifier) GVR() schema.GroupVersionResource {
	return identifier.GVK.GroupVersion().WithResource(identifier.Resource)
}

// IdentifierForObject creates an ObjectIdentifier given an object
// that implements MRObject.
func IdentifierForObject(mrObj MRObject, resource string) ObjectIdentifier {
	gvk := mrObj.GetObjectKind().GroupVersionKind()

	namespacedName := cache.MetaObjectToName(mrObj)

	return ObjectIdentifier{
		GVK:        gvk,
		Resource:   resource,
		ObjectName: namespacedName,
	}
}

func EmptyUnstructuredObjectFromIdentifier(objIdentifier ObjectIdentifier) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}

	obj.SetGroupVersionKind(objIdentifier.GVK)
	obj.SetNamespace(objIdentifier.ObjectName.Namespace)
	obj.SetName(objIdentifier.ObjectName.Name)

	return obj
}

func ObjIdentifierIsForCRD(objIdentifier ObjectIdentifier) bool { // CRDs might have different versions. therefore, using "any" in CRD version
	return gvkMatches(objIdentifier.GVK, apiextensions.GroupName, AnyVersion, CRDKind)
}

func ObjIdentifierIsForBindingPolicy(objIdentifier ObjectIdentifier) bool {
	return gvkMatches(objIdentifier.GVK, v1alpha1.GroupVersion.Group, AnyVersion, BindingPolicyKind)
}

func ObjIdentifierIsForBinding(objIdentifier ObjectIdentifier) bool {
	return gvkMatches(objIdentifier.GVK, v1alpha1.GroupVersion.Group, AnyVersion, BindingKind)
}

func gvkMatches(gvk schema.GroupVersionKind, group, version, kind string) bool {
	if gvk.Group == group && (version == AnyVersion || gvk.Version == version) &&
		gvk.Kind == kind {
		return true
	}
	return false
}
