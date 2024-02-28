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

package cmtest

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	k8sautoscalingapiv2 "k8s.io/api/autoscaling/v2"
	k8score "k8s.io/api/core/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	runtime "k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/wait"
	k8sclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"
	kastesting "k8s.io/kubernetes/cmd/kube-apiserver/app/testing"
	"k8s.io/kubernetes/test/integration/framework"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/binding"
	ksclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned/typed/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const managedClusterCRDURL = "https://raw.githubusercontent.com/open-cluster-management-io/api/v0.12.0/cluster/v1/0000_00_clusters.open-cluster-management.io_managedclusters.crd.yaml"
const manifestWorkCRDURL = "https://raw.githubusercontent.com/open-cluster-management-io/api/v0.12.0/work/v1/0000_00_work.open-cluster-management.io_manifestworks.crd.yaml"

// NumObjEnvar is the name of the environment variable that can be used to specify the number of objects
const NumObjEnvar = "CONTROLLER_TEST_NUM_OBJECTS"

// An integration test for the binding controller.
// This test uses an in-process kube-apiserver created by k8s.io/kubernetes/cmd/kube-apiserver/app/testing.
// That code launches an external insecure etcd server.
// YOU MUST HAVE THE ETCD BINARY ON YOUR `$PATH`.
//
// This test exercises the workload matching functionality.
// This test generates a given number of objects randomly.
// 2/3 of them are created in the apiserver.
// DownsyncObjectTests are extracted from 2/3 of them, with random tweaks.
// This test creates a BindingPolicy object with those DownsyncObjectTests.
// This test then waits up to 1 minute for a Binding with expected matches to appear.
func TestMatching(t *testing.T) {
	rg := rand.New(rand.NewSource(42))
	rg.Uint64()
	rg.Uint64()
	rg.Uint64()
	testWriter := framework.NewTBWriter(t)
	logger, ctx := ktesting.NewTestContext(t)
	logger.Info("Starting etcd server")
	framework.StartEtcd(t, testWriter)
	logger.Info("Starting TestController")
	t.Log("Beginning TestController")
	ctx, cancel := context.WithCancel(ctx)
	testServer, err := kastesting.StartTestServer(t, kastesting.NewDefaultTestServerOptions(), []string{}, framework.SharedEtcd())
	if err != nil {
		t.Fatalf("Failed to kastesting.StartTestServer: %s", err)
	}
	fullTeardwon := func() {
		cancel()
		testServer.TearDownFn()
	}
	t.Cleanup(fullTeardwon)
	config := testServer.ClientConfig
	k8sClient, err := k8sclient.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %s", err)
	}
	logger.Info("Started test server", "config", config)
	configCopy := *config
	config4json := &configCopy
	config4json.ContentType = "application/json"
	logger.Info("REST config for JSON marshaling", "config", config4json)
	ksClient, err := ksclient.NewForConfig(config4json)
	if err != nil {
		t.Fatalf("Failed to create KubeStellar client: %s", err)
	}
	apiextClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create apiextensions client: %s", err)
	}
	scheme := runtime.NewScheme()
	err = apiextensionsapi.AddToScheme(scheme)
	if err != nil {
		t.Fatalf("Failed to apiextensionsapi.AddToScheme(scheme): %s", err)
	}
	serializer := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, scheme, scheme)
	createCRD(t, ctx, "ManagedCluster", managedClusterCRDURL, serializer, apiextClient)
	createCRD(t, ctx, "ManifestWork", manifestWorkCRDURL, serializer, apiextClient)
	time.Sleep(5 * time.Second)
	ctlr, err := binding.NewController(logger, config4json, config, "test-wds", nil)
	if err != nil {
		t.Fatalf("Failed to create controller: %s", err)
	}
	logger.Info("About to EnsureCRDs")
	err = ctlr.EnsureCRDs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("CRDs ensured")
	err = ctlr.Start(ctx, 4)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	namespaces := []*k8score.Namespace{}
	nsORs := []mrObjRsc{}
	for i := 1; i <= 3; i++ {
		ns := generateNamespace(t, ctx, rg, fmt.Sprintf("ns%d", i), k8sClient)
		namespaces = append(namespaces, ns)
		nsORs = append(nsORs, mrObjRsc{ns, "namespaces", nil})
	}
	nObjStr := os.Getenv(NumObjEnvar)
	nObj := 18
	if len(nObjStr) > 0 {
		nObj64, err := strconv.ParseInt(nObjStr, 10, 64)
		if err != nil {
			t.Fatalf("Failed to parse value of environment variable CONTROLLER_TEST_NUM_OBJECTS %q as an int64: %s", nObjStr, err)
		}
		nObj = int(nObj64)
	}
	logger.Info("Generating mrObjRscs", "count", nObj)
	objs := make([]mrObjRsc, nObj)
	objsCreated := make([]mrObjRsc, 0, nObj)
	counts := counters{}
	for i := 0; i < nObj; i++ {
		var thisClient k8sclient.Interface = nil
		if i*3 < nObj*2 { // create only the first 2/3 of the objects
			thisClient = k8sClient
		}
		objs[i] = generateObject(t, ctx, rg, &counts, namespaces, thisClient)
		if i*3 < nObj*2 {
			objsCreated = append(objsCreated, objs[i])
		}
	}
	objsToTests := objs[nObj/3:] // extract tests from the last 2/3
	tests := []ksapi.DownsyncObjectTest{}
	for _, obj := range objsToTests {
		test := extractTest(rg, obj)
		logger.Info("Adding test", "test", test)
		tests = append(tests, test)
	}
	expectation := map[gvrnn]any{}
	for _, obj := range append(nsORs, objsCreated...) {
		identifier, err := util.IdentifierForObject(obj.MRObject, ctlr.GetGvkToGvrMapper())
		if err != nil {
			t.Fatalf("Failed to extract Key from %#v: %s", obj.MRObject, err)
		}
		id2 := id2r(identifier)
		if test := obj.MatchesAny(t, tests); test != nil {
			expectation[id2] = test
			logger.Info("Added expectation", "identifier", id2, "test", *test)
		} else {
			logger.Info("Not expected", "identifier", id2)
		}
	}
	bp := &ksapi.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "matching-tester",
		},
		Spec: ksapi.BindingPolicySpec{
			Downsync: tests,
		},
	}
	_, err = ksClient.BindingPolicies().Create(ctx, bp, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create BidingPolicy: %s", err)
	}
	logger.Info("Created BindingPolicy", "name", bp.Name)
	bindingClient := ksClient.Bindings()
	err = wait.PollUntilContextTimeout(ctx, 5*time.Second, time.Minute, false, func(ctx context.Context) (done bool, err error) {
		binding, err := bindingClient.Get(ctx, bp.Name, metav1.GetOptions{})
		if err != nil {
			logger.Info("Failed to GET Binding", "err", err)
			return false, nil
		}
		if excess, missed := workloadIsExpected(binding.Spec.Workload, expectation); len(excess)+len(missed) > 0 {
			logger.Info("Wrong stuff matched", "excess", excess, "numMissed", len(missed), "missed", missed, "matched", binding.Spec.Workload)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("Never got expected matches; numTests=%d, numObjects=%d, numExpected=%d", len(tests), (nObj*2)/3, len(expectation))
	}
	logger.Info("Success", "numTests", len(tests), "numObjects", (nObj*2)/3, "numExpected", len(expectation))
}

var crdGVK = apiextensionsapi.SchemeGroupVersion.WithKind("CustomResourceDefinition")

func createCRD(t *testing.T, ctx context.Context, kind, url string, serializer *k8sjson.Serializer, apiextClient apiextensionsclientset.Interface) error {
	crdYAML, err := urlGet(url)
	if err != nil {
		t.Fatalf("Failed to read %s CRD from %s: %s", kind, url, err)
	}
	crdAny, _, err := serializer.Decode([]byte(crdYAML), &crdGVK, &apiextensionsapi.CustomResourceDefinition{})
	if err != nil {
		t.Fatalf("Failed to Decode %s CRD: %s", kind, err)
	}
	crd := crdAny.(*apiextensionsapi.CustomResourceDefinition)
	_, err = apiextClient.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create %s CRD: %s", kind, err)
	}
	return nil
}

type gvrnn struct {
	metav1.GroupVersionResource
	cache.ObjectName
}

func (x gvrnn) String() string {
	return x.GroupVersionResource.String() + "," + x.ObjectName.String()
}

func id2r(objId util.ObjectIdentifier) gvrnn {
	return gvrnn{
		GroupVersionResource: metav1.GroupVersionResource(objId.GVR()),
		ObjectName:           objId.ObjectName,
	}
}

// workloadIsExpected returns the symmetric difference, as JSON data
func workloadIsExpected(workload ksapi.DownsyncObjectReferences, expectation map[gvrnn]any) (excess []gvrnn, missed map[string]any) {
	excess = []gvrnn{}
	missed = map[string]any{}
	for key, val := range expectation {
		missed[key.String()] = val
	}
	for _, clusterScopeDownsyncObject := range workload.ClusterScope {
		key := gvrnn{GroupVersionResource: clusterScopeDownsyncObject.GroupVersionResource,
			ObjectName: cache.ObjectName{Name: clusterScopeDownsyncObject.Name}}
		if missed[key.String()] != nil {
			delete(missed, key.String())
		} else {
			excess = append(excess, key)
		}
	}
	for _, namespaceScopeDownsyncObject := range workload.NamespaceScope {
		key := gvrnn{GroupVersionResource: namespaceScopeDownsyncObject.GroupVersionResource,
			ObjectName: cache.ObjectName{Namespace: namespaceScopeDownsyncObject.Namespace,
				Name: namespaceScopeDownsyncObject.Name}}
		if missed[key.String()] != nil {
			delete(missed, key.String())
		} else {
			excess = append(excess, key)
		}
	}
	return
}

type mrObjRsc struct {
	MRObject  util.MRObject
	Resource  string
	Namespace *k8score.Namespace
}

func (mor mrObjRsc) MatchesAny(t *testing.T, tests []ksapi.DownsyncObjectTest) *ksapi.DownsyncObjectTest {
	for _, test := range tests {
		gvk := mor.MRObject.GetObjectKind().GroupVersionKind()
		if test.APIGroup != nil && gvk.Group != *test.APIGroup {
			continue
		}
		if len(test.Resources) > 0 && !(binding.SliceContains(test.Resources, mor.Resource) || binding.SliceContains(test.Resources, "*")) {
			continue
		}
		if len(test.Namespaces) > 0 && !(binding.SliceContains(test.Namespaces, mor.MRObject.GetNamespace()) || binding.SliceContains(test.Namespaces, "*")) {
			continue
		}
		if len(test.ObjectNames) > 0 && !(binding.SliceContains(test.ObjectNames, mor.MRObject.GetName()) || binding.SliceContains(test.ObjectNames, "*")) {
			continue
		}
		if len(test.NamespaceSelectors) > 0 && !(mor.Namespace == nil && binding.ALabelSelectorIsEmpty(test.NamespaceSelectors...) ||
			mor.Namespace != nil && LabelsMatchAny(t, mor.Namespace.Labels, test.NamespaceSelectors)) {
			continue
		}
		if len(test.ObjectSelectors) > 0 && !LabelsMatchAny(t, mor.MRObject.GetLabels(), test.ObjectSelectors) {
			continue
		}
		thisTest := test // don't trust those golang loop vars
		return &thisTest
	}
	return nil
}

func LabelsMatchAny(t *testing.T, labels map[string]string, selectors []metav1.LabelSelector) bool {
	for _, ls := range selectors {
		sel, err := metav1.LabelSelectorAsSelector(&ls)
		if err != nil {
			t.Fatalf("Failed to convert LabelSelector %#v to labels.Selector: %s", ls, err)
		}
		if sel.Matches(k8slabels.Set(labels)) {
			return true
		}
	}
	return false

}

func extractTest(rg *rand.Rand, object mrObjRsc) ksapi.DownsyncObjectTest {
	ans := ksapi.DownsyncObjectTest{}
	if rg.Intn(100) < 25 {
		group := object.MRObject.GetObjectKind().GroupVersionKind().Group
		ans.APIGroup = &group
	}
	ans.Resources = extractStringTest(rg, object.Resource, false)
	if object.Namespace != nil {
		ans.Namespaces = extractStringTest(rg, object.MRObject.GetNamespace(), false)
		ans.NamespaceSelectors = extractLabelsTest(rg, object.Namespace.Labels, false)
	}
	// Ensure there is a name test that does not accept non-test objects
	discBySel := rg.Intn(2) < 1
	ans.ObjectNames = extractStringTest(rg, object.MRObject.GetName(), !discBySel)
	ans.ObjectSelectors = extractLabelsTest(rg, object.MRObject.GetLabels(), discBySel)
	return ans
}

func extractStringTest(rg *rand.Rand, good string, avoidAny bool) []string {
	ans := []string{}
	tail := []string{}
	if rg.Intn(20) < 1 {
		tail = append(tail, "bar")
	}
	if rg.Intn(20) < 1 {
		ans = append(ans, "foo")
	}
	if rg.Intn(10) < 7 || avoidAny && len(ans) == 0 && len(tail) == 0 {
		ans = append(ans, good)
	}
	return append(ans, tail...)
}

func extractLabelsTest(rg *rand.Rand, goodLabels map[string]string, avoidAny bool) []metav1.LabelSelector {
	testLabels := map[string]string{}
	if rg.Intn(15) < 1 {
		testLabels["foo"] = "bar"
	}
	numGood := len(goodLabels)
	iter := 0
	for key, val := range goodLabels {
		iter = iter + 1
		if !(avoidAny && len(testLabels) == 0 && iter == numGood) && rg.Intn(10) < 5 {
			continue
		}
		testVal := val
		if rg.Intn(10) < 1 {
			testVal = val + "not"
		}
		testLabels[key] = testVal
	}
	return []metav1.LabelSelector{{MatchLabels: testLabels}}
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

func generateNamespace(t *testing.T, ctx context.Context, rg *rand.Rand, name string, client k8sclient.Interface) *k8score.Namespace {
	ans := &k8score.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: generateObjectMeta(rg, name, nil),
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, ans, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create Namespace %#v: %s", ans, err)
	}
	klog.FromContext(ctx).V(1).Info("Generated", "namespace", ans)
	return ans
}

var noK8sClient *k8sclient.Clientset

type counters = [4]int64

func generateObject(t *testing.T, ctx context.Context, rg *rand.Rand, counts *counters, namespaces []*k8score.Namespace, client k8sclient.Interface) mrObjRsc {
	x := rg.Intn(40)
	namespace := namespaces[rg.Intn(len(namespaces))]
	var err error
	var ans mrObjRsc
	var noK8sIfc k8sclient.Interface = noK8sClient
	switch {
	case x < 10:
		counts[0]++
		name := fmt.Sprintf("o%d", counts[0])
		obj := &k8score.ConfigMap{
			TypeMeta:   typeMeta("ConfigMap", k8score.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, name, namespace),
		}
		if client != nil && client != noK8sIfc {
			_, err = client.CoreV1().ConfigMaps(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
		}
		ans = mrObjRsc{obj, "configmaps", namespace}
	case x < 20:
		counts[1]++
		name := fmt.Sprintf("o%d", counts[1])
		obj := &rbacv1.ClusterRole{
			TypeMeta:   typeMeta("ClusterRole", rbacv1.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, name, nil),
		}
		if client != nil && client != noK8sIfc {
			_, err = client.RbacV1().ClusterRoles().Create(ctx, obj, metav1.CreateOptions{})
		}
		ans = mrObjRsc{obj, "clusterroles", nil}
	case x < 30:
		counts[2]++
		name := fmt.Sprintf("o%d", counts[2])
		obj := &k8snetv1.NetworkPolicy{
			TypeMeta:   typeMeta("NetworkPolicy", k8snetv1.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, name, namespace),
		}
		if client != nil && client != noK8sIfc {
			_, err = client.NetworkingV1().NetworkPolicies(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
		}
		ans = mrObjRsc{obj, "networkpolicies", namespace}
	default:
		counts[3]++
		name := fmt.Sprintf("o%d", counts[3])
		obj := &k8sautoscalingapiv2.HorizontalPodAutoscaler{
			TypeMeta:   typeMeta("HorizontalPodAutoscaler", k8sautoscalingapiv2.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, name, namespace),
			Spec: k8sautoscalingapiv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: k8sautoscalingapiv2.CrossVersionObjectReference{
					Kind:       "StatefulSet",
					Name:       "mu",
					APIVersion: "apps/v1",
				},
				MaxReplicas: 2,
			},
		}
		if client != nil && client != noK8sIfc {
			_, err = client.AutoscalingV2().HorizontalPodAutoscalers(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
		}
		ans = mrObjRsc{obj, "horizontalpodautoscalers", namespace}
	}
	if err != nil {
		t.Fatalf("Failed to create object %#v: %s", ans.MRObject, err)
	}
	klog.FromContext(ctx).V(1).Info("Generated", "obj", ans)
	return ans
}

func typeMeta(kind string, groupVersion k8sschema.GroupVersion) metav1.TypeMeta {
	return metav1.TypeMeta{Kind: kind, APIVersion: groupVersion.String()}
}

func urlGet(urlStr string) (string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get(%s) returned %v", urlStr, err)
	}
	read, err := io.ReadAll(resp.Body)
	readS := string(read)
	if err != nil {
		return readS, fmt.Errorf("failed to ReadAll(get(%s)): %w", readS, err)
	}
	return readS, err
}
