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
	"context"
	"testing"
	"time"

	k8scorev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	machruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	upstreamdiscovery "k8s.io/client-go/discovery"

	clusterdynamicinformer "github.com/kcp-dev/client-go/dynamic/dynamicinformer"
	fakekube "github.com/kcp-dev/client-go/kubernetes/fake"
	fakekubediscovery "github.com/kcp-dev/client-go/third_party/k8s.io/client-go/discovery/fake"
	fakeclusterdynamic "github.com/kcp-dev/client-go/third_party/k8s.io/client-go/dynamic/fake"
	kcptesting "github.com/kcp-dev/client-go/third_party/k8s.io/client-go/testing"
	kcpapisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v1alpha1"
	fakeedge "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster/fake"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
)

func TestWhatResolver(t *testing.T) {
	orDie := func(err error) {
		if err != nil {
			t.Fatalf("Eek! %v", err)
		}
	}
	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		var cancel func()
		ctx, cancel = context.WithDeadline(ctx, deadline)
		t.Cleanup(cancel)
	}
	wds1N := logicalcluster.Name("wds1clusterid")
	ns1 := &k8scorev1.Namespace{
		TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "default",
			Annotations: map[string]string{logicalcluster.AnnotationKey: wds1N.String()},
			Labels:      map[string]string{"kubernetes.io/metadata.name": "default"},
		}}
	ns2 := &k8scorev1.Namespace{
		TypeMeta: metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "ns2",
			Annotations: map[string]string{logicalcluster.AnnotationKey: wds1N.String()},
			Labels:      map[string]string{"kubernetes.io/metadata.name": "ns2"},
		}}
	cm1 := &k8scorev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   "default",
			Name:        "cm1",
			Annotations: map[string]string{logicalcluster.AnnotationKey: wds1N.String()},
			Labels:      map[string]string{"foo": "bar"},
		}}
	cm3 := &k8scorev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{Kind: "ConfigMap", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   "default",
			Name:        "cm3",
			Annotations: map[string]string{logicalcluster.AnnotationKey: wds1N.String()},
			Labels:      map[string]string{"foo": "baz"},
		}}
	ep1 := &edgeapi.EdgePlacement{
		TypeMeta: metav1.TypeMeta{Kind: "EdgePlacement", APIVersion: edgeapi.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{logicalcluster.AnnotationKey: wds1N.String()},
			Name:        "ep1",
		},
		Spec: edgeapi.EdgePlacementSpec{
			Downsync: []edgeapi.DownsyncObjectTest{
				{Resources: []string{"configmaps"},
					NamespaceSelectors: []metav1.LabelSelector{{MatchLabels: map[string]string{"kubernetes.io/metadata.name": "default"}}},
					ObjectNames:        []string{"cm1"},
					LabelSelectors:     []metav1.LabelSelector{{MatchLabels: map[string]string{"foo": "baz"}}},
				},
				{Resources: []string{"namespaces"},
					ObjectNames: []string{"ns2"},
				},
			},
			WantSingletonReportedState: true,
		}}
	ep1EN := ExternalName{wds1N, ObjectName(ep1.Name)}
	edgeViewClusterClientset := fakeedge.NewSimpleClientset(ep1)
	edgeClusterInformerFactory := edgeinformers.NewSharedInformerFactory(edgeViewClusterClientset, 0)
	epClusterPreInformer := edgeClusterInformerFactory.Edge().V1alpha1().EdgePlacements()
	fakeKubeClusterClientset := fakekube.NewSimpleClientset(ns1, cm1)
	k8sCoreGroupVersion := metav1.GroupVersion{Version: "v1"}
	usualVerbs := []string{"get", "list", "watch"}
	nsResource := metav1.APIResource{Name: "namespaces", SingularName: "namespace", Namespaced: false, Version: "v1", Kind: "Namespace", Verbs: usualVerbs}
	cmResource := metav1.APIResource{Name: "configmaps", SingularName: "configmap", Namespaced: true, Version: "v1", Kind: "ConfigMap", Verbs: usualVerbs}
	setFakeClusterAPIResources(fakeKubeClusterClientset.Fake, wds1N.Path(), []*metav1.APIResourceList{
		{GroupVersion: k8sCoreGroupVersion.String(),
			APIResources: []metav1.APIResource{nsResource, cmResource},
		}})
	fcd := FakeClusterDisco{fakeKubeClusterClientset.Discovery(), fakeKubeClusterClientset}
	scheme := machruntime.NewScheme()
	orDie(apiextensionsv1.AddToScheme(scheme))
	orDie(kcpapisv1alpha1.AddToScheme(scheme))
	orDie(k8scorev1.AddToScheme(scheme))
	fakeDynamicClusterClientset := fakeclusterdynamic.NewSimpleDynamicClient(scheme, ns1, cm1, ns2, cm3)
	dynamicClusterInformerFactory := clusterdynamicinformer.NewDynamicSharedInformerFactory(fakeDynamicClusterClientset, 0)
	crdClusterPreInformer := dynamicClusterInformerFactory.ForResource(apiextensionsv1.SchemeGroupVersion.WithResource("customresourcedefinitions"))
	bindingClusterPreInformer := dynamicClusterInformerFactory.ForResource(kcpapisv1alpha1.SchemeGroupVersion.WithResource("apibindings"))

	whatResolver := NewWhatResolver(ctx, epClusterPreInformer, fcd, crdClusterPreInformer, bindingClusterPreInformer, fakeDynamicClusterClientset, 3)
	edgeClusterInformerFactory.Start(ctx.Done())
	dynamicClusterInformerFactory.Start(ctx.Done())
	rcvr := NewMapMap[ExternalName, ResolvedWhat](nil)
	runnable := whatResolver(rcvr)
	go runnable.Run(ctx)
	partid1 := WorkloadPartID{metav1.GroupResource{Resource: "configmaps"}, "default", "cm1"}
	partdt1 := WorkloadPartDetails{APIVersion: "v1", ReturnSingletonState: true}
	partid2 := WorkloadPartID{metav1.GroupResource{Resource: "namespaces"}, "", "ns2"}
	partid3 := WorkloadPartID{metav1.GroupResource{Resource: "configmaps"}, "default", "cm3"}
	expectedWhat := ResolvedWhat{Downsync: WorkloadParts{partid1: partdt1, partid2: partdt1, partid3: partdt1}}
	err := wait.PollWithContext(ctx, time.Second, 5*time.Second, func(context.Context) (bool, error) {
		gotWhat, found := rcvr.Get(ep1EN)
		t.Logf("gotWhat=%v, found=%v", gotWhat, found)
		return found && apiequality.Semantic.DeepEqual(expectedWhat, gotWhat), nil
	})
	if err != nil {
		t.Fatalf("Failed to get expected ResolvedWhat in time: %v", err)
	}
}

func setFakeClusterAPIResources(fake *kcptesting.Fake, cluster logicalcluster.Path, resources []*metav1.APIResourceList) {
	fake.Lock()
	defer fake.Unlock()
	if fake.Resources == nil {
		fake.Resources = map[logicalcluster.Path][]*metav1.APIResourceList{}
	}
	fake.Resources[cluster] = resources
}

type FakeClusterDisco struct {
	upstreamdiscovery.DiscoveryInterface
	kubeClusterClientset *fakekube.ClusterClientset
}

func (fcd FakeClusterDisco) Cluster(lc logicalcluster.Path) upstreamdiscovery.DiscoveryInterface {
	return &fakekubediscovery.FakeDiscovery{Fake: fcd.kubeClusterClientset.Fake, ClusterPath: lc}
}
