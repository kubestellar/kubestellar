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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	k8sautoscalingapiv2 "k8s.io/api/autoscaling/v2"
	k8score "k8s.io/api/core/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
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
	a "github.com/kubestellar/kubestellar/pkg/abstract"
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
// The test maintains a set of workload objects, which varies throughout the test.
// The test maintains a set of DownsyncObjectTests, which varies throughout the test.
// The DownsyncObjectTests are derived from, on average, 2/3 of the workload objects,
// chosen uniformly at random. The derived test looks for the base object but with tweaks
// to make it match more or fewer.
// An average of 2/3 of the workload objects, chosen randomly, have been created in the apiserver
// at any given moment; the rest have not been created or have also been deleted.
//
// The test proceeds through a series of rounds.
// There are 7 types of rounds, identified by the numbers 1 through 7.
// Each round's type is chosen uniformly at random, except the first round's type is 7.
// Each round has three phases.
// The first phase thrashes half of the workload objects randomly,
// if the type number is odd, otherwise does nothing.
// Thrashing means randomly deleting a workload object from the apiserver, or
// creating a workload object in the apiserver, or replacing an object that is
// neither in the apiserver nor the source of a test.
// The second phase randomly generates a new set of tests from the workload objects
// if the type number's 2's bit is set, otherwise does nothing.
// The third phase randomly thrashes half of the worklod objects
// if the type number is greater than 3, otherwise does nothing.
// At the end of a round the expected "what resolution" is computed and the test
// waits for a limited amount of time to observe that result, failing if the
// expected result is not observed in time.
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
	err = ctlr.AppendKSResources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Appended KS resources to discovered lists")
	err = ctlr.Start(ctx, 4, make(chan interface{}, 1))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)
	namespaces := []*k8score.Namespace{}
	nsORs := []mrObjRsc{}
	for i := 1; i <= 3; i++ {
		ns := generateNamespace(t, ctx, rg, fmt.Sprintf("ns%d", i), k8sClient)
		namespaces = append(namespaces, ns)
		nsORs = append(nsORs, mrObjRsc{ns, "namespaces", nil, nil, nil})
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
	counts := counters{}
	allIdxs := make([]int, nObj)
	for i := 0; i < nObj; i++ {
		objs[i] = generateObject(ctx, rg, &counts, namespaces, k8sClient)
		allIdxs[i] = i
		logger.V(1).Info("Generated", "idx", i, "obj", objs[i])
	}
	idxsCreated := []int{}
	idxsNotCreated := append([]int{}, allIdxs...)
	bp := &ksapi.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "matching-tester",
		},
		Spec: ksapi.BindingPolicySpec{},
	}
	createPolicy := true
	minObjs := (nObj * 2) / 5
	maxObjs := (nObj * 3) / 5
	serialThrash := func() {
		logger.V(1).Info("Thrashing objects serially")
		for i := 0; i < nObj/2; i++ {
			nCreated := len(idxsCreated)
			if minObjs <= nCreated && nCreated <= maxObjs && rg.Intn(3) == 0 { // replace
				j := rg.Intn(len(idxsNotCreated))
				idx := idxsNotCreated[j]
				oldObj := objs[idx]
				newObj := generateObject(ctx, rg, &counts, namespaces, k8sClient)
				objs[idx] = newObj
				logger.V(1).Info("Replaced test object", "idx", idx, "oldObj", oldObj, "newObj", newObj)
			} else if nCreated <= minObjs || nCreated < maxObjs && rg.Intn(2) == 0 { // create
				j := rg.Intn(len(idxsNotCreated))
				idx := idxsNotCreated[j]
				obj := objs[idx]
				err := obj.create()
				if err != nil {
					t.Fatalf("Failed to create object %#v: %s", obj, err)
				}
				idxsCreated = append(idxsCreated, idx)
				a.SliceDelete(&idxsNotCreated, j)
				logger.V(1).Info("Created test object", "idx", idx, "obj", obj)
			} else { // delete
				j := rg.Intn(len(idxsCreated))
				idx := idxsCreated[j]
				obj := objs[idx]
				err := obj.delete()
				if err != nil {
					t.Fatalf("Failed to delete object %#v: %s", obj, err)
				}
				a.SliceDelete(&idxsCreated, j)
				idxsNotCreated = append(idxsNotCreated, idx)
				logger.V(1).Info("Deleted test object", "idx", idx, "obj", obj)
			}
		}
	}
	parallelThrash := func() {
		logger.V(1).Info("Thrashing objects in parallel")
		var wg sync.WaitGroup
		var listsMutex sync.Mutex
		var failed int32 = 0
		newIdxsCreated := []int{}
		newIdxsNotCreated := []int{}
		thrashObj := func(j int) {
			defer wg.Add(-1)
			var idx int
			var exists bool
			if j < len(idxsCreated) {
				idx = idxsCreated[j]
				exists = true
			} else {
				idx = idxsNotCreated[j-len(idxsCreated)]
			}
			obj := objs[idx]
			for k := 0; k <= rg.Intn(4); k++ {
				if exists {
					err := obj.delete()
					if err != nil {
						t.Errorf("Failed to delete; idx=%d, obj=%#v, err=%s", idx, obj, err.Error())
						atomic.StoreInt32(&failed, 1)
					}
					exists = false
				} else {
					err := obj.create()
					if err != nil {
						t.Errorf("Failed to create; idx=%d, obj=%#v, err=%s", idx, obj, err.Error())
						atomic.StoreInt32(&failed, 1)
					}
					exists = true
				}
			}
			listsMutex.Lock()
			defer listsMutex.Unlock()
			if exists {
				newIdxsCreated = append(newIdxsCreated, idx)
			} else {
				newIdxsNotCreated = append(newIdxsNotCreated, idx)
			}
		}
		for j := 0; j < len(objs); j++ {
			wg.Add(1)
			go thrashObj(j)
		}
		wg.Wait()
		if atomic.LoadInt32(&failed) != 0 {
			t.FailNow()
		}
		idxsCreated = newIdxsCreated
		idxsNotCreated = newIdxsNotCreated
	}
	tests := []ksapi.DownsyncObjectTest{}
	bindingClient := ksClient.Bindings()
	const thrashBeforeBit int = 1
	const changeTestsBit int = 2
	const thrashAfterbit int = 4
	var roundType int = thrashBeforeBit | changeTestsBit | thrashAfterbit
	nRounds := 10
	for round := 1; round <= nRounds; round++ {
		logger.Info("Starting round", "round", round, "roundType", roundType)
		if roundType&thrashBeforeBit != 0 {
			if rg.Intn(2) == 0 {
				serialThrash()
			} else {
				parallelThrash()
			}
		}
		if roundType&changeTestsBit != 0 {
			idxsNotInTest := append([]int{}, allIdxs...)
			tests = []ksapi.DownsyncObjectTest{}
			for i := 0; i*3 < nObj*2; i++ {
				j := rg.Intn(len(idxsNotInTest))
				idx := idxsNotInTest[j]
				a.SliceDelete(&idxsNotInTest, j)
				obj := objs[idx]
				test := extractTest(rg, obj)
				logger.Info("Adding test", "test", test)
				tests = append(tests, test)
			}
			if createPolicy {
				bp.Spec.Downsync = tests
				bp, err = ksClient.BindingPolicies().Create(ctx, bp, metav1.CreateOptions{})
				if err != nil {
					t.Fatalf("Failed to create BidingPolicy: %s", err)
				}
				logger.Info("Created BindingPolicy", "name", bp.Name)
				createPolicy = false
			} else {
				bp.Spec.Downsync = tests
				bp, err = ksClient.BindingPolicies().Update(ctx, bp, metav1.UpdateOptions{})
				if err != nil && errors.IsConflict(err) {
					bp, err = ksClient.BindingPolicies().Get(ctx, "matching-tester", metav1.GetOptions{})
					if err != nil {
						t.Fatalf("Failed to Get updated BidingPolicy: %s", err)
					}
					bp.Spec.Downsync = tests
					bp, err = ksClient.BindingPolicies().Update(ctx, bp, metav1.UpdateOptions{})
				}
				if err != nil {
					t.Fatalf("Failed to update BidingPolicy: %s", err)
				}
				logger.Info("Updated BindingPolicy", "name", bp.Name, "round", round)
			}
		}
		if roundType&thrashAfterbit != 0 {
			if rg.Intn(2) == 0 {
				serialThrash()
			} else {
				parallelThrash()
			}
		}
		expectation := map[gvrnn]any{}
		consider := func(obj mrObjRsc, idx int) {
			identifier := util.IdentifierForObject(obj.MRObject, obj.Resource)
			id2 := id2r(identifier)
			if test := obj.MatchesAny(t, tests); test != nil {
				expectation[id2] = test
				logger.Info("Added expectation", "idx", idx, "identifier", id2, "test", *test)
			} else {
				logger.Info("Not expected", "idx", idx, "identifier", id2)
			}
		}
		for idx, obj := range nsORs {
			consider(obj, idx)
		}
		for _, idx := range idxsCreated {
			obj := objs[idx]
			consider(obj, idx)
		}
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
			t.Fatalf("Round %d type %d never got expected matches; numTests=%d, numObjects=%d, numExpected=%d", round, roundType, len(tests), len(idxsCreated), len(expectation))
		}
		logger.Info("Round success", "round", round, "roundType", roundType, "numTests", len(tests), "numObjects", len(idxsCreated), "numExpected", len(expectation))

		roundType = 1 + rg.Intn(7)
	}

	logger.Info("Success", "rounds", nRounds, "nObj", nObj)
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
	create    func() error
	delete    func() error
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

type counters = [4]int64

func generateObject(ctx context.Context, rg *rand.Rand, counts *counters, namespaces []*k8score.Namespace, client k8sclient.Interface) mrObjRsc {
	x := rg.Intn(40)
	namespace := namespaces[rg.Intn(len(namespaces))]
	var ans mrObjRsc
	switch {
	case x < 10:
		counts[0]++
		name := fmt.Sprintf("o%d", counts[0])
		obj := &k8score.ConfigMap{
			TypeMeta:   typeMeta("ConfigMap", k8score.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, name, namespace),
		}
		ans = mrObjRsc{obj, "configmaps", namespace,
			func() error {
				_, err := client.CoreV1().ConfigMaps(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
				return err
			},
			func() error {
				return client.CoreV1().ConfigMaps(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
			}}
	case x < 20:
		counts[1]++
		name := fmt.Sprintf("o%d", counts[1])
		obj := &rbacv1.ClusterRole{
			TypeMeta:   typeMeta("ClusterRole", rbacv1.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, name, nil),
		}
		ans = mrObjRsc{obj, "clusterroles", nil,
			func() error {
				_, err := client.RbacV1().ClusterRoles().Create(ctx, obj, metav1.CreateOptions{})
				return err
			},
			func() error {
				return client.RbacV1().ClusterRoles().Delete(ctx, obj.Name, metav1.DeleteOptions{})
			}}
	case x < 30:
		counts[2]++
		name := fmt.Sprintf("o%d", counts[2])
		obj := &k8snetv1.NetworkPolicy{
			TypeMeta:   typeMeta("NetworkPolicy", k8snetv1.SchemeGroupVersion),
			ObjectMeta: generateObjectMeta(rg, name, namespace),
		}
		ans = mrObjRsc{obj, "networkpolicies", namespace,
			func() error {
				_, err := client.NetworkingV1().NetworkPolicies(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
				return err
			},
			func() error {
				return client.NetworkingV1().NetworkPolicies(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
			}}

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
		ans = mrObjRsc{obj, "horizontalpodautoscalers", namespace,
			func() error {
				_, err := client.AutoscalingV2().HorizontalPodAutoscalers(obj.Namespace).Create(ctx, obj, metav1.CreateOptions{})
				return err
			},
			func() error {
				return client.AutoscalingV2().HorizontalPodAutoscalers(obj.Namespace).Delete(ctx, obj.Name, metav1.DeleteOptions{})
			}}
	}
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
