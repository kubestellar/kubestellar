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

package transport

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	clusterclientfake "open-cluster-management.io/api/client/cluster/clientset/versioned/fake"
	clusterinformers "open-cluster-management.io/api/client/cluster/informers/externalversions"
	clusterapi "open-cluster-management.io/api/cluster/v1"
	workapi "open-cluster-management.io/api/work/v1"

	k8sautoscalingapiv2 "k8s.io/api/autoscaling/v2"
	k8score "k8s.io/api/core/v1"
	k8snetv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	ksclientfake "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned/fake"
	ksinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions"
	ksmetrics "github.com/kubestellar/kubestellar/pkg/metrics"
	"github.com/kubestellar/kubestellar/pkg/transport"
	"github.com/kubestellar/kubestellar/pkg/util"
)

func TestMain(m *testing.M) {
	klog.InitFlags(nil)
	os.Exit(m.Run())
}

type generator struct {
	t   *testing.T
	ctx context.Context
	*rand.Rand
	counts     [4]int64
	namespaces []*k8score.Namespace
}

type mrObject interface {
	metav1.Object
	runtime.Object
}

type mrObjRsc struct {
	MRObject mrObject
	Resource string
}

func (gen *generator) generateLabels() map[string]string {
	ans := map[string]string{}
	n := 1 + gen.Intn(2)
	for i := 1; i <= n; i++ {
		ans[fmt.Sprintf("l%d", i*10+gen.Intn(2))] = fmt.Sprintf("v%d", i*10+gen.Intn(2))
	}
	switch gen.Intn(4) {
	case 0:
		ans["test.kubestellar.io/dont-delete-me"] = "thank"
	case 1:
		ans["test.kubestellar.io/delete-me"] = "you"
	default:
	}
	return ans
}

func generateResourceVersion() string {
	// using a timestamp to simulate a unique resource version.
	// this is nowhere as complex as the real resourceVersion generation,
	// but suffices for testing.
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func (gen *generator) generateObjectMeta(name string, namespace *k8score.Namespace) metav1.ObjectMeta {
	ans := metav1.ObjectMeta{
		Name:            name,
		Labels:          gen.generateLabels(),
		ResourceVersion: generateResourceVersion(),
	}
	if namespace != nil {
		ans.Namespace = namespace.Name
	}
	return ans
}

func (gen *generator) generateNamespace(name string) *k8score.Namespace {
	return &k8score.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: gen.generateObjectMeta(name, nil),
	}
}

func (rg *generator) generateObject() mrObjRsc {
	namespace := rg.namespaces[rg.Intn(len(rg.namespaces))]
	x := rg.Intn(4)
	switch {
	case x < 1:
		rg.counts[0]++
		name := fmt.Sprintf("o%d", rg.counts[0])
		return mrObjRsc{&k8score.ConfigMap{
			TypeMeta:   typeMeta("ConfigMap", k8score.SchemeGroupVersion),
			ObjectMeta: rg.generateObjectMeta(name, namespace),
		}, "configmaps"}
	case x < 2:
		rg.counts[1]++
		name := fmt.Sprintf("o%d", rg.counts[1])
		return mrObjRsc{&rbacv1.ClusterRole{
			TypeMeta:   typeMeta("ClusterRole", rbacv1.SchemeGroupVersion),
			ObjectMeta: rg.generateObjectMeta(name, nil),
		}, "clusterroles"}
	case x < 3:
		rg.counts[2]++
		name := fmt.Sprintf("o%d", rg.counts[2])
		return mrObjRsc{&k8snetv1.NetworkPolicy{
			TypeMeta:   typeMeta("NetworkPolicy", k8snetv1.SchemeGroupVersion),
			ObjectMeta: rg.generateObjectMeta(name, namespace),
		}, "networkpolicies"}
	default:
		rg.counts[3]++
		name := fmt.Sprintf("o%d", rg.counts[3])
		return mrObjRsc{&k8sautoscalingapiv2.HorizontalPodAutoscaler{
			TypeMeta:   typeMeta("HorizontalPodAutoscaler", k8sautoscalingapiv2.SchemeGroupVersion),
			ObjectMeta: rg.generateObjectMeta(name, namespace),
			Spec: k8sautoscalingapiv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: k8sautoscalingapiv2.CrossVersionObjectReference{
					Kind:       "StatefulSet",
					Name:       "mu",
					APIVersion: "apps/v1",
				},
				MaxReplicas: 2,
			},
		}, "horizontalpodautoscalers"}
	}
}

func typeMeta(kind string, groupVersion k8sschema.GroupVersion) metav1.TypeMeta {
	return metav1.TypeMeta{Kind: kind, APIVersion: groupVersion.String()}
}

type bindingCase struct {
	Binding      *ksapi.Binding
	expect       map[util.GVKObjRef]jsonMapToWrap
	ExpectedKeys []any // JSON equivalent of keys of expect, for logging
}

func newClusterScope(gvr metav1.GroupVersionResource, name, resourceVersion string, createOnly bool) ksapi.ClusterScopeDownsyncClause {
	return ksapi.ClusterScopeDownsyncClause{
		ClusterScopeDownsyncObject: ksapi.ClusterScopeDownsyncObject{
			GroupVersionResource: gvr,
			Name:                 name,
			ResourceVersion:      resourceVersion,
		},
		CreateOnly: createOnly}
}

func newNamespaceScope(gvr metav1.GroupVersionResource, namespace, name, resourceVersion string, createOnly bool) ksapi.NamespaceScopeDownsyncClause {
	return ksapi.NamespaceScopeDownsyncClause{
		NamespaceScopeDownsyncObject: ksapi.NamespaceScopeDownsyncObject{
			GroupVersionResource: gvr,
			Namespace:            namespace,
			Name:                 name,
			ResourceVersion:      resourceVersion,
		},
		CreateOnly: createOnly}
}

func (bc *bindingCase) Add(obj mrObjRsc, createOnly bool) {
	key := util.RefToRuntimeObj(obj.MRObject)
	gvr := metav1.GroupVersionResource{
		Group:    key.GK.Group,
		Version:  obj.MRObject.GetObjectKind().GroupVersionKind().Version,
		Resource: obj.Resource,
	}
	objNS := obj.MRObject.GetNamespace()
	objName := obj.MRObject.GetName()
	objRV := obj.MRObject.GetResourceVersion()
	jm, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj.MRObject)
	if err != nil {
		panic(err)
	}

	if objNS == "" {
		clusterObj := newClusterScope(gvr, objName, objRV, createOnly)
		bc.Binding.Spec.Workload.ClusterScope = append(bc.Binding.Spec.Workload.ClusterScope, clusterObj)
	} else {
		namespaceObj := newNamespaceScope(gvr, objNS, objName, objRV, createOnly)
		bc.Binding.Spec.Workload.NamespaceScope = append(bc.Binding.Spec.Workload.NamespaceScope, namespaceObj)
	}

	bc.expect[key] = jsonMapToWrap{jm, createOnly}
	bc.ExpectedKeys = append(bc.ExpectedKeys, key.String())
}

func (rg *generator) generateBindingCase(name string, objs []mrObjRsc) bindingCase {
	bc := bindingCase{
		Binding: &ksapi.Binding{
			TypeMeta:   typeMeta("Binding", ksapi.GroupVersion),
			ObjectMeta: rg.generateObjectMeta(name, nil),
			Spec:       ksapi.BindingSpec{},
		},
		expect: map[util.GVKObjRef]jsonMapToWrap{},
	}
	for _, obj := range objs {
		if rg.Intn(10) < 7 {
			createOnly := rg.Intn(2) == 0
			klog.FromContext(rg.ctx).V(3).Info("Adding to bindingCase", "case", name, "obj", util.RefToRuntimeObj(obj.MRObject), "createOnly", createOnly)
			bc.Add(obj, createOnly)
		}
	}
	return bc
}

type jsonMap = map[string]any
type jsonMapToWrap struct {
	jm         jsonMap
	createOnly bool
}

type testTransport struct {
	t              *testing.T
	ctx            context.Context
	bindingName    string
	ctc            customTransformCollection
	kindToResource map[metav1.GroupKind]string

	expect map[util.GVKObjRef]jsonMapToWrap
	sync.Mutex
	wrapped bool
	missed  map[string]any
	wrong   map[string]any
	extra   []any
}

func (tt *testTransport) WrapObjects(wrapees []transport.Wrapee) runtime.Object {
	tt.Lock()
	defer tt.Unlock()
	tt.wrapped = true
	tt.missed = map[string]any{}
	for key, val := range tt.expect {
		tt.missed[key.String()] = fmt.Sprintf("%#v", val)
	}
	tt.wrong = map[string]any{}
	tt.extra = []any{}
	for _, wrapee := range wrapees {
		obj := wrapee.Object
		// TODO: test wrapee.CreateOnly
		key := util.RefToRuntimeObj(obj)
		delete(tt.missed, key.String())
		if expectedJMTW, found := tt.expect[key]; found {
			if wrapee.CreateOnly != expectedJMTW.createOnly {
				tt.t.Errorf("Expected createOnly=%v, got %v obj=%v", expectedJMTW.createOnly, wrapee.CreateOnly, key)
			}
			objM := obj.UnstructuredContent()
			apiVersion := obj.GetAPIVersion()
			groupVersion, err := k8sschema.ParseGroupVersion(apiVersion)
			if err != nil {
				panic(err)
			}
			groupKind := metav1.GroupKind{Group: groupVersion.Group, Kind: obj.GetKind()}
			resource, ok := tt.kindToResource[groupKind]
			if !ok {
				panic(fmt.Errorf("No mapping for %v", groupKind))
			}
			groupResource := metav1.GroupResource{Group: groupKind.Group, Resource: resource}
			// clean expected object since transport objects are cleaned
			uncleanedExpectedObj := &unstructured.Unstructured{Object: expectedJMTW.jm}
			cleanedExpectedObjU := TransformObject(tt.ctx, tt.ctc, groupResource, uncleanedExpectedObj, tt.bindingName)
			cleanedExpectedObj := cleanedExpectedObjU.Object
			cleanable := obj.GetKind() == "ClusterRole"
			hadLabel := uncleanedExpectedObj.GetLabels()["test.kubestellar.io/delete-me"] != ""
			hasLabel := cleanedExpectedObjU.GetLabels()["test.kubestellar.io/delete-me"] != ""
			expectRemoval := cleanable && hadLabel
			actualRemoval := hadLabel && !hasLabel
			if expectRemoval != actualRemoval {
				tt.t.Errorf("Expected removal=%v, actual removal=%v, obj=%v", expectRemoval, actualRemoval, key)
			} else if expectRemoval {
				tt.t.Logf("Saw a removal, obj=%v", key)
			}
			equal := apiequality.Semantic.DeepEqual(objM, cleanedExpectedObj)
			if !equal {
				tt.wrong[key.String()] = obj
			}
		} else {
			tt.extra = append(tt.extra, obj)
		}
	}
	return &workapi.ManifestWork{
		TypeMeta: typeMeta("ManifestWork", workapi.GroupVersion),
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo",
			Name:      "bar",
		},
	}
}

func TestGenericController(t *testing.T) {
	rg := rand.New(rand.NewSource(42))
	rg.Uint64()
	rg.Uint64()
	rg.Uint64()
	scheme := runtime.NewScheme()
	k8score.AddToScheme(scheme)
	k8snetv1.AddToScheme(scheme)
	k8sautoscalingapiv2.AddToScheme(scheme)
	rbacv1.AddToScheme(scheme)
	clusterapi.AddToScheme(scheme)
	workapi.AddToScheme(scheme)
	ksapi.AddToScheme(scheme)
	logger, ctx := ktesting.NewTestContext(t)
	gen := &generator{t: t, ctx: ctx, Rand: rg}
	ct := &ksapi.CustomTransform{
		TypeMeta:   metav1.TypeMeta{APIVersion: ksapi.GroupVersion.String(), Kind: "CustomTransform"},
		ObjectMeta: metav1.ObjectMeta{Name: "test-ct"},
		Spec: ksapi.CustomTransformSpec{
			APIGroup: rbacv1.SchemeGroupVersion.Group,
			Resource: "clusterroles",
			Remove:   []string{`$.metadata.labels["test.kubestellar.io/delete-me"]`},
		}}
	wdsK8sObjs := []runtime.Object{}
	for i := 0; i < 3; i++ {
		ns := gen.generateNamespace(fmt.Sprintf("ns%d", i))
		logger.V(3).Info("Generated namespace", "ns", ns)
		gen.namespaces = append(gen.namespaces, ns)
		wdsK8sObjs = append(wdsK8sObjs, ns)
	}
	objs := []mrObjRsc{}
	nObj := 100
	for i := 0; i < nObj; i++ {
		obj := gen.generateObject()
		logger.V(3).Info("Generated object", "obj", obj)
		objs = append(objs, obj)
		wdsK8sObjs = append(wdsK8sObjs, obj.MRObject)
	}
	bindingCase := gen.generateBindingCase("b1", objs)
	logger.V(3).Info("Generated bindingCase", "case", bindingCase)
	wdsKsObjs := []runtime.Object{bindingCase.Binding, ct}
	wdsKsClientFake := ksclientfake.NewSimpleClientset(wdsKsObjs...)
	wdsKsInformerFactory := ksinformers.NewSharedInformerFactory(wdsKsClientFake, 0*time.Minute)
	wdsDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, wdsK8sObjs...)
	itsDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	wdsControlInformers := wdsKsInformerFactory.Control().V1alpha1()
	ctIndexer := wdsControlInformers.CustomTransforms().Informer().GetIndexer()
	transport := &testTransport{
		t:           t,
		bindingName: bindingCase.Binding.Name,
		ctx:         ctx,
		ctc: newCustomTransformCollection(wdsKsClientFake.ControlV1alpha1().CustomTransforms(),
			ctIndexer.ByIndex,
			func(any) {}),
		kindToResource: map[metav1.GroupKind]string{
			{Group: "", Kind: "ConfigMap"}:                                          "configmaps",
			{Group: rbacv1.GroupName, Kind: "ClusterRole"}:                          "clusterroles",
			{Group: k8snetv1.GroupName, Kind: "NetworkPolicy"}:                      "networkpolicies",
			{Group: k8sautoscalingapiv2.GroupName, Kind: "HorizontalPodAutoscaler"}: "horizontalpodautoscalers",
		},
		expect: bindingCase.expect}
	wrapperGVR := workapi.GroupVersion.WithResource("manifestworks")
	inventoryClientFake := clusterclientfake.NewSimpleClientset()
	inventoryInformerFactory := clusterinformers.NewSharedInformerFactory(inventoryClientFake, 0*time.Second)
	inventoryPreInformer := inventoryInformerFactory.Cluster().V1().ManagedClusters()
	itsK8sClientFake := k8sfake.NewSimpleClientset()
	itsK8sInformerFactory := k8sinformers.NewSharedInformerFactory(itsK8sClientFake, 0*time.Minute)
	parmCfgMapPreInformer := itsK8sInformerFactory.Core().V1().ConfigMaps()
	spacesClientMetrics := ksmetrics.NewMultiSpaceClientMetrics()
	ksmetrics.MustRegister(legacyregistry.Register, spacesClientMetrics)
	wdsClientMetrics := spacesClientMetrics.MetricsForSpace("wds")
	itsClientMetrics := spacesClientMetrics.MetricsForSpace("its")
	ctlr := NewTransportControllerForWrappedObjectGVR(ctx, wdsClientMetrics, itsClientMetrics,
		inventoryPreInformer, wdsKsClientFake.ControlV1alpha1().Bindings(),
		wdsControlInformers.Bindings(), wdsControlInformers.CustomTransforms(),
		transport,
		wdsKsClientFake,
		wdsDynamicClient,
		itsK8sClientFake.CoreV1().Namespaces(), parmCfgMapPreInformer,
		itsDynamicClient, 500*1024, 500*1024, "test-wds", wrapperGVR)
	ctlr.RegisterMetrics(legacyregistry.Register)
	inventoryInformerFactory.Start(ctx.Done())
	wdsKsInformerFactory.Start(ctx.Done())
	itsK8sInformerFactory.Start(ctx.Done())

	go ctlr.Run(ctx, 4)
	err := wait.PollUntilContextTimeout(ctx, 5*time.Second, time.Minute, false, func(ctx context.Context) (done bool, err error) {
		transport.Lock()
		defer transport.Unlock()
		if transport.wrapped && len(transport.missed) == 0 && len(transport.wrong) == 0 && len(transport.extra) == 0 {
			return true, nil
		}
		if !transport.wrapped {
			logger.Info("No wrapping done yet")
		} else {
			logger.Info("Last wrapping was bad", "missed", transport.missed, "wrong", transport.wrong, "extra", transport.extra)
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("Never got right call to WrapObjects")
	} else {
		logger.Info("Success", "objects", len(objs), "numExpected", len(transport.expect))
	}
}
