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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kubestellar/kubestellar/api/edge/v1alpha1"
)

func namespacedNameFromObjectMeta(meta metav1.ObjectMeta) types.NamespacedName {
	return types.NamespacedName{
		Namespace: meta.Namespace,
		Name:      meta.Name,
	}
}

func destinationsStringSliceToDestinations(destinations []string) []v1alpha1.Destination {
	dests := make([]v1alpha1.Destination, len(destinations))
	for i, d := range destinations {
		dests[i] = v1alpha1.Destination{ClusterId: d}
	}

	return dests
}
