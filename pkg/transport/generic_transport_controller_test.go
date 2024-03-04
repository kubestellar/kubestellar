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
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	ksclientfake "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned/fake"
	ksinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions"
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
	expect       map[util.GVKObjRef]jsonMap
	ExpectedKeys []any // JSON equivalent of keys of expect, for logging
}

func newClusterScope(gvr metav1.GroupVersionResource, name, resourceVersion string) ksapi.ClusterScopeDownsyncObject {
	return ksapi.ClusterScopeDownsyncObject{
		GroupVersionResource: gvr,
		Name:                 name,
		ResourceVersion:      resourceVersion,
	}
}

func newNamespaceScope(gvr metav1.GroupVersionResource, namespace, name, resourceVersion string) ksapi.NamespaceScopeDownsyncObject {
	return ksapi.NamespaceScopeDownsyncObject{
		GroupVersionResource: gvr,
		Namespace:            namespace,
		Name:                 name,
		ResourceVersion:      resourceVersion,
	}
}

func (bc *bindingCase) Add(obj mrObjRsc) {
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
		clusterObj := newClusterScope(gvr, objName, objRV)
		bc.Binding.Spec.Workload.ClusterScope = append(bc.Binding.Spec.Workload.ClusterScope, clusterObj)
	} else {
		namespaceObj := newNamespaceScope(gvr, objNS, objName, objRV)
		bc.Binding.Spec.Workload.NamespaceScope = append(bc.Binding.Spec.Workload.NamespaceScope, namespaceObj)
	}

	bc.expect[key] = jm
	bc.ExpectedKeys = append(bc.ExpectedKeys, key.String())
}

func (rg *generator) generateBindingCase(name string, objs []mrObjRsc) bindingCase {
	bc := bindingCase{
		Binding: &ksapi.Binding{
			TypeMeta:   typeMeta("Binding", ksapi.GroupVersion),
			ObjectMeta: rg.generateObjectMeta(name, nil),
			Spec:       ksapi.BindingSpec{},
		},
		expect: map[util.GVKObjRef]jsonMap{},
	}
	for _, obj := range objs {
		if rg.Intn(10) < 7 {
			bc.Add(obj)
		}
	}
	return bc
}

type jsonMap = map[string]any

type testTransport struct {
	expect map[util.GVKObjRef]jsonMap
	sync.Mutex
	wrapped bool
	missed  map[string]any
	wrong   map[string]any
	extra   []any
}

func (tt *testTransport) WrapObjects(objs []*unstructured.Unstructured) runtime.Object {
	tt.Lock()
	defer tt.Unlock()
	tt.wrapped = true
	tt.missed = map[string]any{}
	for key, val := range tt.expect {
		tt.missed[key.String()] = fmt.Sprintf("%#v", val)
	}
	tt.wrong = map[string]any{}
	tt.extra = []any{}
	for _, obj := range objs {
		key := util.RefToRuntimeObj(obj)
		delete(tt.missed, key.String())
		if expectedObj, found := tt.expect[key]; found {
			objM := obj.UnstructuredContent()

			// clean expected object since transport objects are cleaned
			cleanedExpectedObj := cleanObject(&unstructured.Unstructured{Object: expectedObj}).Object
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
	var _ ksapi.Binding
	gen := &generator{t: t, ctx: ctx, Rand: rg}
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
	wdsKsObjs := []runtime.Object{bindingCase.Binding}
	wdsKsClientFake := ksclientfake.NewSimpleClientset(wdsKsObjs...)
	wdsKsInformerFactory := ksinformers.NewSharedInformerFactory(wdsKsClientFake, 0*time.Minute)
	wdsDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, wdsK8sObjs...)
	itsDynamicClient := dynamicfake.NewSimpleDynamicClient(scheme)
	transport := &testTransport{expect: bindingCase.expect}
	wrapperGVR := workapi.GroupVersion.WithResource("manifestworks")
	ctlr := NewTransportControllerForWrappedObjectGVR(ctx, wdsKsInformerFactory.Control().V1alpha1().Bindings(), transport, wdsKsClientFake, wdsDynamicClient, itsDynamicClient, "test-wds", wrapperGVR)
	wdsKsInformerFactory.Start(ctx.Done())
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
