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

package filtering

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// cleanObjectSpecificsFunction is a function for cleaning fields from a specific object.
// The function cleans the specific fields in place (object is modified).
// If the object was retrieved using a lister, it's the caller responsibility
// to do a DeepCopy before calling this function.
type cleanObjectSpecificsFunction func(object *unstructured.Unstructured)

func NewObjectFilteringMap() *ObjectFilteringMap {
	filteringMap := map[schema.GroupVersionKind]cleanObjectSpecificsFunction{
		corev1.SchemeGroupVersion.WithKind("Service"): cleanService,
		batchv1.SchemeGroupVersion.WithKind("Job"):    cleanJob,
	}

	return &ObjectFilteringMap{
		gvkToFilteringFunc: filteringMap,
	}
}

type ObjectFilteringMap struct {
	gvkToFilteringFunc map[schema.GroupVersionKind]cleanObjectSpecificsFunction // map from GVK to clean object function
}

func (filteringMap *ObjectFilteringMap) CleanObjectSpecifics(object *unstructured.Unstructured) {
	gvk := object.GetObjectKind().GroupVersionKind()
	filteringFunction, found := filteringMap.gvkToFilteringFunc[gvk]
	if !found {
		return // if no filtering function was defined for this gvk, do not clean any field
	}
	// otherwise, need to clean specific fields from this object
	filteringFunction(object)
}
