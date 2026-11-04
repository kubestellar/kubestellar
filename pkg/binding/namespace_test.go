/*
Copyright 2026 The KubeStellar Authors.

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
	"testing"

	k8scoreapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2/ktesting"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	controllisters "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

var deploymentGVR = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

func policyWithClauses(name string, clauses ...v1alpha1.DownsyncPolicyClause) *v1alpha1.BindingPolicy {
	return &v1alpha1.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       v1alpha1.BindingPolicySpec{Downsync: clauses},
	}
}

func nsSelectorClause(selectors ...metav1.LabelSelector) v1alpha1.DownsyncPolicyClause {
	return v1alpha1.DownsyncPolicyClause{
		DownsyncObjectTest: v1alpha1.DownsyncObjectTest{NamespaceSelectors: selectors},
	}
}

func TestNamespaceSelectionChanged(t *testing.T) {
	logger, _ := ktesting.NewTestContext(t)
	deployYes := metav1.LabelSelector{MatchLabels: map[string]string{"deploy": "yes"}}
	envProd := metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
		{Key: "env", Operator: metav1.LabelSelectorOpIn, Values: []string{"prod"}}}}

	for _, tc := range []struct {
		name      string
		policies  []*v1alpha1.BindingPolicy
		oldLabels map[string]string
		newLabels map[string]string
		expected  bool
	}{
		{
			name:      "label added makes selector match",
			policies:  []*v1alpha1.BindingPolicy{policyWithClauses("p", nsSelectorClause(deployYes))},
			newLabels: map[string]string{"deploy": "yes"},
			expected:  true,
		},
		{
			name:      "label removed makes selector stop matching",
			policies:  []*v1alpha1.BindingPolicy{policyWithClauses("p", nsSelectorClause(deployYes))},
			oldLabels: map[string]string{"deploy": "yes", "other": "x"},
			newLabels: map[string]string{"other": "x"},
			expected:  true,
		},
		{
			name:      "label value change flips match expression",
			policies:  []*v1alpha1.BindingPolicy{policyWithClauses("p", nsSelectorClause(envProd))},
			oldLabels: map[string]string{"env": "staging"},
			newLabels: map[string]string{"env": "prod"},
			expected:  true,
		},
		{
			name:      "unrelated label change",
			policies:  []*v1alpha1.BindingPolicy{policyWithClauses("p", nsSelectorClause(deployYes))},
			oldLabels: map[string]string{"deploy": "yes"},
			newLabels: map[string]string{"deploy": "yes", "other": "x"},
			expected:  false,
		},
		{
			name:      "clause without namespaceSelectors",
			policies:  []*v1alpha1.BindingPolicy{policyWithClauses("p", v1alpha1.DownsyncPolicyClause{})},
			newLabels: map[string]string{"deploy": "yes"},
			expected:  false,
		},
		{
			name:      "empty selector matches every namespace",
			policies:  []*v1alpha1.BindingPolicy{policyWithClauses("p", nsSelectorClause(deployYes, metav1.LabelSelector{}))},
			newLabels: map[string]string{"deploy": "yes"},
			expected:  false,
		},
		{
			name: "second policy affected",
			policies: []*v1alpha1.BindingPolicy{
				policyWithClauses("p1", nsSelectorClause(envProd)),
				policyWithClauses("p2", v1alpha1.DownsyncPolicyClause{}, nsSelectorClause(deployYes)),
			},
			newLabels: map[string]string{"deploy": "yes"},
			expected:  true,
		},
		{
			name:      "no policies",
			newLabels: map[string]string{"deploy": "yes"},
			expected:  false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if actual := namespaceSelectionChanged(logger, tc.policies, tc.oldLabels, tc.newLabels); actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func newUnstructured(apiVersion, kind, namespace, name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(apiVersion)
	obj.SetKind(kind)
	obj.SetNamespace(namespace)
	obj.SetName(name)
	return obj
}

func newTestLister(t *testing.T, gvr schema.GroupVersionResource, objs ...*unstructured.Unstructured) cache.GenericLister {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, obj := range objs {
		if err := indexer.Add(obj); err != nil {
			t.Fatalf("failed to add object to indexer: %v", err)
		}
	}
	return cache.NewGenericLister(indexer, gvr.GroupResource())
}

func newNamespace(name string, labels map[string]string) *k8scoreapi.Namespace {
	return &k8scoreapi.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}}
}

// drainQueue returns the set of items currently in the queue.
func drainQueue(queue workqueue.TypedRateLimitingInterface[any]) sets.Set[any] {
	items := sets.New[any]()
	for queue.Len() > 0 {
		item, _ := queue.Get()
		items.Insert(item)
		queue.Done(item)
	}
	return items
}

func TestHandleNamespaceUpdate(t *testing.T) {
	_, ctx := ktesting.NewTestContext(t)

	nginx := newUnstructured("apps/v1", "Deployment", "demo", "nginx")
	cfg := newUnstructured("v1", "ConfigMap", "demo", "cfg")
	otherNSCfg := newUnstructured("v1", "ConfigMap", "other", "cfg")
	nginxID := util.IdentifierForObject(nginx, deploymentGVR.Resource)
	cfgID := util.IdentifierForObject(cfg, "configmaps")

	policyIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	deployYes := metav1.LabelSelector{MatchLabels: map[string]string{"deploy": "yes"}}
	if err := policyIndexer.Add(policyWithClauses("ns-selected", nsSelectorClause(deployYes))); err != nil {
		t.Fatalf("failed to add BindingPolicy to indexer: %v", err)
	}

	newController := func() *Controller {
		listers := util.NewConcurrentMap[schema.GroupVersionResource, cache.GenericLister]()
		listers.Set(deploymentGVR, newTestLister(t, deploymentGVR, nginx))
		configMapGVR := k8scoreapi.SchemeGroupVersion.WithResource("configmaps")
		listers.Set(configMapGVR, newTestLister(t, configMapGVR, cfg, otherNSCfg))
		listers.Set(util.GetBindingPolicyGVR(), newTestLister(t, util.GetBindingPolicyGVR(),
			newUnstructured(v1alpha1.GroupVersion.String(), "BindingPolicy", "demo", "should-be-skipped")))
		return &Controller{
			bindingPolicyLister: controllisters.NewBindingPolicyLister(policyIndexer),
			listers:             listers,
			workqueue: workqueue.NewTypedRateLimitingQueue(
				workqueue.DefaultTypedControllerRateLimiter[any]()),
		}
	}

	for _, tc := range []struct {
		name      string
		oldLabels map[string]string
		newLabels map[string]string
		expected  sets.Set[any]
	}{
		{
			name:      "namespace starts matching",
			newLabels: map[string]string{"deploy": "yes"},
			expected:  sets.New[any](nginxID, cfgID),
		},
		{
			name:      "namespace stops matching",
			oldLabels: map[string]string{"deploy": "yes"},
			expected:  sets.New[any](nginxID, cfgID),
		},
		{
			name:      "unrelated label change",
			oldLabels: map[string]string{"deploy": "yes"},
			newLabels: map[string]string{"deploy": "yes", "team": "a"},
			expected:  sets.New[any](),
		},
		{
			name:      "labels unchanged",
			oldLabels: map[string]string{"deploy": "yes"},
			newLabels: map[string]string{"deploy": "yes"},
			expected:  sets.New[any](),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newController()
			defer c.workqueue.ShutDown()
			c.handleNamespaceUpdate(ctx, newNamespace("demo", tc.oldLabels), newNamespace("demo", tc.newLabels))
			if actual := drainQueue(c.workqueue); !actual.Equal(tc.expected) {
				t.Errorf("expected queued items %v, got %v", tc.expected.UnsortedList(), actual.UnsortedList())
			}
		})
	}
}
