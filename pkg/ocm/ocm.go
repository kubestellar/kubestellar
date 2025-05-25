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

package ocm

import (
	"context"
	"fmt"

	managedclusterapi "open-cluster-management.io/api/cluster/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	ksmetrics "github.com/kubestellar/kubestellar/pkg/metrics"
)

func FindClustersBySelectors(ctx context.Context, client ksmetrics.ClientModNamespace[*managedclusterapi.ManagedCluster, *managedclusterapi.ManagedClusterList], selectors []metav1.LabelSelector) (sets.Set[string], error) {
	// in order to support OR between label selectors in a straightforward manner, we perform List for each selector.
	// additionally, to support complex selectors (such as set selectors), we avoid conversion to maps.
	clusterNames := sets.New[string]()
	for _, s := range selectors {
		ls, err := metav1.LabelSelectorAsSelector(&s)
		if err != nil {
			return clusterNames, fmt.Errorf("failed to convert metav1.LabelSelector to labels.Selector: %w", err)
		}
		clusters, err := client.List(ctx, metav1.ListOptions{LabelSelector: ls.String()})
		if err != nil {
			return nil, fmt.Errorf("error listing clusters with selector %s: %w", ls, err)
		}

		for _, cluster := range clusters.Items {
			clusterNames.Insert(cluster.GetName())
		}
	}

	return clusterNames, nil
}
