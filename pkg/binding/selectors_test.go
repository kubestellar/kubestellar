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
	"errors"
	"testing"

	k8scoreapi "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2/ktesting"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	controllisters "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// failingLister is a cache.GenericLister whose every call fails with the given error.
type failingLister struct{ err error }

func (fl failingLister) List(labels.Selector) ([]runtime.Object, error)  { return nil, fl.err }
func (fl failingLister) Get(string) (runtime.Object, error)              { return nil, fl.err }
func (fl failingLister) ByNamespace(string) cache.GenericNamespaceLister { return fl }

func listerWith(t *testing.T, gr schema.GroupResource, objs ...runtime.Object) cache.GenericLister {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	for _, obj := range objs {
		if err := indexer.Add(obj); err != nil {
			t.Fatalf("failed to add object to indexer: %v", err)
		}
	}
	return cache.NewGenericLister(indexer, gr)
}

func TestUpdateResolutionsNamespaceLookup(t *testing.T) {
	_, ctx := ktesting.NewTestContext(t)

	deploymentsGVR := schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	nginx := &unstructured.Unstructured{}
	nginx.SetAPIVersion("apps/v1")
	nginx.SetKind("Deployment")
	nginx.SetNamespace("demo")
	nginx.SetName("nginx")
	nginxID := util.IdentifierForObject(nginx, deploymentsGVR.Resource)

	matchingNS := &k8scoreapi.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demo", Labels: map[string]string{"deploy": "yes"}}}
	policy := &v1alpha1.BindingPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "ns-selected"},
		Spec: v1alpha1.BindingPolicySpec{Downsync: []v1alpha1.DownsyncPolicyClause{{
			DownsyncObjectTest: v1alpha1.DownsyncObjectTest{NamespaceSelectors: []metav1.LabelSelector{
				{MatchLabels: map[string]string{"deploy": "yes"}}}},
		}}},
	}

	for _, tc := range []struct {
		name            string
		namespaceLister cache.GenericLister
		expectErr       bool
		expectSelected  bool
	}{
		{
			name:            "namespace in cache matches",
			namespaceLister: listerWith(t, namespacesGVR.GroupResource(), matchingNS),
			expectSelected:  true,
		},
		{
			name: "namespace in cache does not match",
			namespaceLister: listerWith(t, namespacesGVR.GroupResource(),
				&k8scoreapi.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demo"}}),
			expectSelected: false,
		},
		{
			name:            "namespace not found",
			namespaceLister: listerWith(t, namespacesGVR.GroupResource()),
			expectSelected:  false,
		},
		{
			name:            "transient error keeps the resolution unchanged",
			namespaceLister: failingLister{err: apierrors.NewInternalError(errors.New("boom"))},
			expectErr:       true,
			expectSelected:  true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			policyIndexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
			if err := policyIndexer.Add(policy); err != nil {
				t.Fatalf("failed to add BindingPolicy to indexer: %v", err)
			}
			listers := util.NewConcurrentMap[schema.GroupVersionResource, cache.GenericLister]()
			listers.Set(deploymentsGVR, listerWith(t, deploymentsGVR.GroupResource(), nginx))
			listers.Set(namespacesGVR, tc.namespaceLister)
			c := &Controller{
				bindingPolicyLister:   controllisters.NewBindingPolicyLister(policyIndexer),
				bindingPolicyResolver: NewBindingPolicyResolver(),
				listers:               listers,
				workqueue: workqueue.NewTypedRateLimitingQueue(
					workqueue.DefaultTypedControllerRateLimiter[any]()),
			}
			defer c.workqueue.ShutDown()

			// start with the object already in the resolution, as if it had been selected before
			c.bindingPolicyResolver.NoteBindingPolicy(policy)
			if _, err := c.bindingPolicyResolver.EnsureObjectData(policy.Name, nginxID, "", "1", ZeroDownsyncModulation()); err != nil {
				t.Fatalf("failed to seed resolution: %v", err)
			}

			err := c.updateResolutions(ctx, nginxID)
			if tc.expectErr != (err != nil) {
				t.Fatalf("expected error=%v, got %v", tc.expectErr, err)
			}
			objIDs, err := c.bindingPolicyResolver.GetObjectIdentifiers(policy.Name)
			if err != nil {
				t.Fatalf("failed to get object identifiers: %v", err)
			}
			if selected := objIDs.Has(nginxID); selected != tc.expectSelected {
				t.Errorf("expected object selected=%v, got %v", tc.expectSelected, selected)
			}
		})
	}
}

func TestTestObjectClusterScopedWithNamespaceSelectors(t *testing.T) {
	_, ctx := ktesting.NewTestContext(t)

	// The Namespace lister must not be consulted for a cluster-scoped object.
	listers := util.NewConcurrentMap[schema.GroupVersionResource, cache.GenericLister]()
	listers.Set(namespacesGVR, failingLister{err: apierrors.NewInternalError(errors.New("boom"))})
	c := &Controller{listers: listers}

	objID := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		Resource:   "clusterroles",
		ObjectName: cache.ObjectName{Name: "viewer"},
	}
	nonEmpty := metav1.LabelSelector{MatchLabels: map[string]string{"deploy": "yes"}}

	for _, tc := range []struct {
		name      string
		selectors []metav1.LabelSelector
		expected  bool
	}{
		{name: "only non-empty selectors", selectors: []metav1.LabelSelector{nonEmpty}, expected: false},
		{name: "includes an empty selector", selectors: []metav1.LabelSelector{nonEmpty, {}}, expected: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clauses := []v1alpha1.DownsyncPolicyClause{{
				DownsyncObjectTest: v1alpha1.DownsyncObjectTest{NamespaceSelectors: tc.selectors},
			}}
			matched, _, err := c.testObject(ctx, "p", objID, nil, clauses)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if matched != tc.expected {
				t.Errorf("expected matched=%v, got %v", tc.expected, matched)
			}
		})
	}
}
