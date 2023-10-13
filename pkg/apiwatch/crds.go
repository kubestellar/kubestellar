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

package apiwatch

import (
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CRDAnalyzer struct {
	ObjectNotifier
}

func (crda CRDAnalyzer) EnumerateDefinedResources(obj any) ResourceDefinitionEnumerator {
	crd := obj.(*apiext.CustomResourceDefinition)
	return func(consumer func(metav1.GroupVersionResource)) {
		for _, version := range crd.Spec.Versions {
			gvr := metav1.GroupVersionResource{Group: crd.Spec.Group, Version: version.Name, Resource: crd.Status.AcceptedNames.Plural}
			consumer(gvr)
		}
	}
}
