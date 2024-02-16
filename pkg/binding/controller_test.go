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

package binding

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	ocmfake "open-cluster-management.io/api/client/cluster/clientset/versioned/fake"
	clusterapi "open-cluster-management.io/api/cluster/v1"
	workapi "open-cluster-management.io/api/work/v1"

	k8score "k8s.io/api/core/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	k8snetv1b1 "k8s.io/api/networking/v1beta1"
	apiextensionsfakeclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2/ktesting"
	ctlrunfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	ksfakeclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned/fake"
	"github.com/kubestellar/kubestellar/pkg/util"
)

func TestMatching(t *testing.T) {
	rg := rand.New(rand.NewSource(42))
	rg.Uint64()
	rg.Uint64()
	rg.Uint64()
	logger, ctx := ktesting.NewTestContext(t)
	nObj := 3
	for trial := 1; trial <= 2; trial++ {
		ctx, cancel := context.WithCancel(ctx)
		scheme := k8sruntime.NewScheme()
		k8score.AddToScheme(scheme)
		k8snetv1.AddToScheme(scheme)
		k8snetv1b1.AddToScheme(scheme)
		clusterapi.AddToScheme(scheme)
		workapi.AddToScheme(scheme)
		ksapi.AddToScheme(scheme)

		namespaces := []*k8score.Namespace{
			generateNamespace(t, ctx, rg, "ns1"),
			generateNamespace(t, ctx, rg, "ns2"),
			generateNamespace(t, ctx, rg, "ns3"),
		}
		objs := make([]mrObjRsc, nObj)
		expectedObjRefs := sets.New[util.Key]()
		initialObjects := []k8sruntime.Object{}
		for _, ns := range namespaces {
			initialObjects = append(initialObjects, ns)
		}
		for i := 0; i < nObj; i++ {
			objs[i] = generateObject(t, ctx, rg, 0, namespaces)
			if i*3 < nObj*2 {
				initialObjects = append(initialObjects, objs[i].mrObject)
			}
		}
		tests := []ksapi.DownsyncObjectTest{}
		for i := nObj / 3; i < nObj; i++ {
			tests = append(tests, extractTest(rg, objs[i]))
		}
		for i := 0; i < nObj; i++ {
			if objs[i].MatchesAny(t, tests) {
				key, err := util.KeyForGroupVersionKindNamespaceName(objs[i].mrObject)
				if err != nil {
					t.Fatalf("Failed to extract Key from %#v: %s", objs[i].mrObject, err)
				}
				expectedObjRefs.Insert(key)
			}
		}
		bp := &ksapi.BindingPolicy{
			TypeMeta: typeMeta("BindingPolicy", ksapi.GroupVersion),
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("trial%d", trial),
			},
			Spec: ksapi.BindingPolicySpec{
				Downsync: tests,
			},
		}
		// initialObjects = append(initialObjects, bp)
		logger.Info("Initial objects", "objects", initialObjects)
		k8sFakeClient := k8sfake.NewSimpleClientset(initialObjects...)
		logger.Info("Initial fake Resources", "resources", k8sFakeClient.Fake.Resources)
		disco := k8sFakeClient.Discovery()
		preferredResources, err := disco.ServerPreferredResources()
		logger.Info("Discover result", "ServerPreferredResources", preferredResources, "err", err)
		initialObjects = append(initialObjects, bp)
		dynamicFakeClient := dynamicfake.NewSimpleDynamicClient(scheme, initialObjects...)
		extFakeClient := apiextensionsfakeclientset.NewSimpleClientset()
		_ = ksfakeclient.NewSimpleClientset(bp)
		ocmFakeClient := ocmfake.NewSimpleClientset()
		cb := ctlrunfake.NewClientBuilder()
		cb = cb.WithScheme(scheme)
		for _, obj := range initialObjects {
			cb = cb.WithObjects(obj.(mrObject))
		}
		ctlrunFakeClient := cb.Build()

		ctlr, err := makeController(logger, dynamicFakeClient, k8sFakeClient, extFakeClient, ocmFakeClient, ctlrunFakeClient, "wds1", nil)
		if err != nil {
			t.Fatalf("Failed to create controller: %s", err)
		}
		ctlr.Start(ctx, 4)
		// TODO: wait to settle, evaluate
		cancel()
	}
}

type mrObjRsc struct {
	mrObject
	resource  string
	namespace *k8score.Namespace
}

func (mor mrObjRsc) MatchesAny(t *testing.T, tests []ksapi.DownsyncObjectTest) bool {
	for _, test := range tests {
		gvk := mor.GetObjectKind().GroupVersionKind()
		if test.APIGroup != nil && gvk.Group != *test.APIGroup {
			continue
		}
		if len(test.Resources) > 0 && !(SliceContains(test.Resources, mor.resource) || SliceContains(test.Resources, "*")) {
			continue
		}
		if len(test.Namespaces) > 0 && !(SliceContains(test.Namespaces, mor.GetNamespace()) || SliceContains(test.Namespaces, "*")) {
			continue
		}
		if len(test.ObjectNames) > 0 && !(SliceContains(test.ObjectNames, mor.GetName()) || SliceContains(test.ObjectNames, "*")) {
			continue
		}
		if len(test.NamespaceSelectors) > 0 && !LabelsMatchAny(t, mor.namespace.Labels, test.NamespaceSelectors) {
			continue
		}
		if len(test.ObjectSelectors) > 0 && !LabelsMatchAny(t, mor.GetLabels(), test.ObjectSelectors) {
			continue
		}
	}
	return false
}

func LabelsMatchAny(t *testing.T, labels map[string]string, selectors []metav1.LabelSelector) bool {
	for _, ls := range selectors {
		sel, err := metav1.LabelSelectorAsSelector(&ls)
		if err != nil {
			t.Fatalf("Failed to convert LabelSelector %#v to labels.Selector: %s", ls, err)
			continue
		}
		if sel.Matches(k8slabels.Set(labels)) {
			return true
		}
	}
	return false

}

func extractTest(rg *rand.Rand, object mrObjRsc) ksapi.DownsyncObjectTest {
	ans := ksapi.DownsyncObjectTest{}
	if rg.Intn(10) < 7 {
		group := object.GetObjectKind().GroupVersionKind().Group
		ans.APIGroup = &group
	}
	ans.Resources = extractStringTest(rg, object.resource)
	if object.namespace != nil {
		ans.Namespaces = extractStringTest(rg, object.GetNamespace())
		ans.NamespaceSelectors = extractLabelsTest(rg, object.namespace.Labels)
	}
	ans.ObjectNames = extractStringTest(rg, object.GetName())
	ans.ObjectSelectors = extractLabelsTest(rg, object.GetLabels())
	return ans
}

func extractStringTest(rg *rand.Rand, good string) []string {
	ans := []string{}
	if rg.Intn(10) < 2 {
		ans = append(ans, "foo")
	}
	if rg.Intn(10) < 7 {
		ans = append(ans, good)
	}
	if rg.Intn(10) < 2 {
		ans = append(ans, "bar")
	}
	return ans
}

func extractLabelsTest(rg *rand.Rand, goodLabels map[string]string) []metav1.LabelSelector {
	testLabels := map[string]string{}
	if rg.Intn(10) < 2 {
		testLabels["foo"] = "bar"
	}
	for key, val := range goodLabels {
		if rg.Intn(10) < 5 {
			continue
		}
		testVal := val
		if rg.Intn(10) < 2 {
			testVal = val + "not"
		}
		testLabels[key] = testVal
	}
	return []metav1.LabelSelector{{MatchLabels: testLabels}}
}

func getObjectTest(rg *rand.Rand, apiGroups []string, resources []string, namespaces []*k8score.Namespace, objects []mrObject) ksapi.DownsyncObjectTest {
	ans := ksapi.DownsyncObjectTest{}
	if rg.Intn(10) < 7 {
		ans.APIGroup = &apiGroups[rg.Intn(len(apiGroups))]
	}
	ans.Resources = make([]string, rg.Intn(3))
	baseJ := 0
	for i := range ans.Resources {
		// Leave room for len(ans.Resources) - (i+1) more
		// pick an index in [baseJ, len(resources) + i+1 - len(ans.Resources))
		j := baseJ + rg.Intn(len(resources)+i+1-len(ans.Resources)-baseJ)
		ans.Resources[i] = resources[j]
		baseJ = j + 1
	}
	ans.Namespaces = make([]string, rg.Intn(3))
	baseJ = 0
	for i := range ans.Namespaces {
		j := baseJ + rg.Intn(len(namespaces)+i+1-len(ans.Namespaces)-baseJ)
		ans.Namespaces[i] = namespaces[j].Name
		baseJ = j + 1
	}
	if rg.Intn(2) == 0 {
		i := rg.Intn(len(namespaces))
		ans.NamespaceSelectors = []metav1.LabelSelector{{MatchLabels: namespaces[i].Labels}}
	}
	ans.ObjectNames = make([]string, rg.Intn(3))
	baseJ = 0
	for i := range ans.ObjectNames {
		j := baseJ + rg.Intn(len(objects)+i+1-len(ans.ObjectNames)-baseJ)
		ans.ObjectNames[i] = objects[j].GetName()
		baseJ = j + 1
	}
	return ans
}

func generateLabels(rg *rand.Rand) map[string]string {
	ans := map[string]string{}
	n := 1 + rg.Intn(2)
	for i := 1; i <= n; i++ {
		ans[fmt.Sprintf("l%d", i*10+rg.Intn(2))] = fmt.Sprintf("v%d", i*10+rg.Intn(2))
	}
	return ans
}

func generateObjectMeta(rg *rand.Rand, name string, namespace *k8score.Namespace) metav1.ObjectMeta {
	ans := metav1.ObjectMeta{
		Name:   name,
		Labels: generateLabels(rg),
	}
	if namespace != nil {
		ans.Namespace = namespace.Name
	}
	return ans
}

func generateNamespace(t *testing.T, ctx context.Context, rg *rand.Rand, name string) *k8score.Namespace {
	ans := &k8score.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: generateObjectMeta(rg, name, nil),
	}
	return ans
}

func generateObject(t *testing.T, ctx context.Context, rg *rand.Rand, index int, namespaces []*k8score.Namespace) mrObjRsc {
	x := rg.Intn(40)
	namespace := namespaces[rg.Intn(len(namespaces))]
	switch {
	case x < 10:
		obj := &k8score.ConfigMap{
			TypeMeta:   typeMeta("ConfigMap", k8score.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, fmt.Sprintf("o%d", index), namespace),
		}
		return mrObjRsc{obj, "configmaps", namespace}
	case x < 20:
		obj := &k8score.Node{
			TypeMeta:   typeMeta("Node", k8score.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, fmt.Sprintf("o%d", index), nil),
		}
		return mrObjRsc{obj, "nodes", nil}
	case x < 30:
		obj := &k8snetv1.NetworkPolicy{
			TypeMeta:   typeMeta("NetworkPolicy", k8snetv1.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, fmt.Sprintf("o%d", index), namespace),
		}
		return mrObjRsc{obj, "networkpolicies", namespace}
	default:
		obj := &k8snetv1b1.IngressClass{
			TypeMeta:   typeMeta("IngressClass", k8snetv1b1.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, fmt.Sprintf("o%d", index), nil),
		}
		return mrObjRsc{obj, "networkpolicies", nil}
	}
}

func typeMeta(kind string, groupVersion k8sschema.GroupVersion) metav1.TypeMeta {
	return metav1.TypeMeta{Kind: kind, APIVersion: groupVersion.String()}
}
