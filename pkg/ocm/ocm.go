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
	"os"
	"strings"

	clusterv1 "open-cluster-management.io/api/cluster/v1"
	workv1 "open-cluster-management.io/api/work/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/kubestellar/kubestellar/pkg/util"
)

// wrapObject creates a ManifestWork with a single manifest containing the given object
func WrapObject(obj runtime.Object) *workv1.ManifestWork {
	strippedObj := ZeroFields(obj)
	return &workv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ManifestWork",
			APIVersion: "work.open-cluster-management.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: buildManifestName(strippedObj),
		},
		Spec: workv1.ManifestWorkSpec{
			Workload: workv1.ManifestsTemplate{
				Manifests: []workv1.Manifest{
					{
						RawExtension: runtime.RawExtension{Object: strippedObj},
					},
				},
			},
		},
	}
}

// BuildEmptyManifestFromObject creates an empty ManifestWork which can be used to delete
func BuildEmptyManifestFromObject(obj runtime.Object) *workv1.ManifestWork {
	return &workv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ManifestWork",
			APIVersion: "work.open-cluster-management.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: buildManifestName(obj),
		},
	}
}

func ZeroFields(obj runtime.Object) runtime.Object {
	zeroed := obj.DeepCopyObject()
	mObj := zeroed.(metav1.Object)
	mObj.SetManagedFields(nil)
	mObj.SetCreationTimestamp(metav1.Time{})
	mObj.SetGeneration(0)
	mObj.SetResourceVersion("")
	mObj.SetUID("")
	annotations := mObj.GetAnnotations()
	delete(annotations, "kubectl.kubernetes.io/last-applied-configuration")
	mObj.SetAnnotations(annotations)

	// service needs additional processing (see https://github.com/kubestellar/kubestellar/issues/4
	// and https://github.com/kubestellar/kubestellar/issues/1451)
	if util.IsService(zeroed) {
		util.RemoveRuntimeGeneratedFieldsFromService(zeroed)
	}
	return zeroed
}

func GetOCMClient(kubeconfig *rest.Config) *client.Client {

	scheme := runtime.NewScheme()

	httpClient, err := rest.HTTPClientFor(kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating HTTPClient: %v\n", err)
		os.Exit(1)
	}
	mapper, err := apiutil.NewDiscoveryRESTMapper(kubeconfig, httpClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating NewDiscoveryRESTMapper: %v\n", err)
		os.Exit(1)
	}
	if err := workv1.AddToScheme(scheme); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding to schema: %v\n", err)
		os.Exit(1)
	}
	if err := clusterv1.AddToScheme(scheme); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding to schema: %v\n", err)
		os.Exit(1)
	}
	c, err := client.New(kubeconfig, client.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}
	return &c
}

func buildManifestName(obj any) string {
	mObj := obj.(metav1.Object)
	rObj := obj.(runtime.Object)
	ok := rObj.GetObjectKind()
	gvk := ok.GroupVersionKind()
	gvStr := RemoveChar(gvk.GroupVersion().String(), '/')
	return fmt.Sprintf("%s-%s-%s-%s",
		strings.ToLower(gvStr),
		strings.ToLower(gvk.Kind),
		mObj.GetNamespace(),
		mObj.GetName(),
	)
}

func RemoveChar(s string, c rune) string {
	i := strings.IndexRune(s, c)
	if i == -1 {
		return s
	}
	return s[:i] + s[i+1:]
}

func GetClusterByName(ocmClient client.Client, clusterName string) (clusterv1.ManagedCluster, error) {
	cluster := clusterv1.ManagedCluster{}
	nn := types.NamespacedName{
		Namespace: "",
		Name:      clusterName,
	}
	if err := ocmClient.Get(context.TODO(), nn, &cluster); err != nil {
		return cluster, err
	}
	return cluster, nil
}

func FindClustersBySelectors(ocmClient client.Client, selectors []metav1.LabelSelector) (sets.Set[string], error) {
	clusters := &clusterv1.ManagedClusterList{}
	labelSelectors := []labels.Selector{}
	for _, s := range selectors {
		selector, err := metav1.LabelSelectorAsSelector(&s)
		if err != nil {
			return nil, err
		}
		labelSelectors = append(labelSelectors, selector)
	}
	labelsMap, err := convertLabelSelectorsToMap(labelSelectors)
	if err != nil {
		return nil, err
	}
	if err := ocmClient.List(context.TODO(), clusters, client.MatchingLabels(labelsMap)); err != nil {
		return nil, err
	}
	if len(clusters.Items) == 0 {
		return nil, nil
	}

	clusterNames := sets.New[string]()
	for _, cluster := range clusters.Items {
		clusterNames.Insert(cluster.GetName())
	}

	return clusterNames, nil
}

func convertLabelSelectorsToMap(labelSelectors []labels.Selector) (map[string]string, error) {
	labelsMap := make(map[string]string)
	for _, selector := range labelSelectors {
		strSelector := selector.String()
		labelSet, err := labels.ConvertSelectorToLabelsMap(strSelector)
		if err != nil {
			return nil, err
		}
		for key, value := range labelSet {
			labelsMap[key] = value
		}
	}
	return labelsMap, nil
}
