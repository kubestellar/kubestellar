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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
)

type APIBindingAnalyzer struct {
	ObjectNotifier
}

var _ ResourceDefinitionSupplier = APIBindingAnalyzer{}

func (aba APIBindingAnalyzer) GetGVK(obj any) schema.GroupVersionKind {
	return apisv1alpha1.SchemeGroupVersion.WithKind("APIBinding")
}

func (aba APIBindingAnalyzer) EnumerateDefinedResources(obj any) ResourceDefinitionEnumerator {
	ab := obj.(*apisv1alpha1.APIBinding)
	return func(consumer func(metav1.GroupVersionResource)) {
		for _, bar := range ab.Status.BoundResources {
			for _, version := range bar.StorageVersions {
				gvr := metav1.GroupVersionResource{Group: bar.Group, Version: version, Resource: bar.Resource}
				consumer(gvr)
			}
		}
	}
}
