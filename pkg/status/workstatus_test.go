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

package status

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	appsv1 "k8s.io/api/apps/v1"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/pkg/util"
)

var testGVR = appsv1.SchemeGroupVersion.WithResource("deployments")

var testGVK = appsv1.SchemeGroupVersion.WithKind("Deployment")

func newTestUnstructured(namespace, name string, status map[string]interface{}) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":            name,
				"namespace":       namespace,
				"resourceVersion": "999",
			},
		},
	}
	if status != nil {
		obj.Object["status"] = status
	}
	return obj
}

type fakeGenericLister struct {
	obj runtime.Object
}

func (f *fakeGenericLister) List(selector labels.Selector) ([]runtime.Object, error) {
	return []runtime.Object{f.obj}, nil
}

func (f *fakeGenericLister) Get(name string) (runtime.Object, error) {
	return f.obj, nil
}

func (f *fakeGenericLister) ByNamespace(namespace string) cache.GenericNamespaceLister {
	return f
}

func newFakeScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		&unstructured.Unstructured{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "DeploymentList"},
		&unstructured.UnstructuredList{},
	)
	return scheme
}

func newTestObjectIdentifier(namespace, name string) util.ObjectIdentifier {
	return util.ObjectIdentifier{
		GVK:        testGVK,
		Resource:   "deployments",
		ObjectName: cache.NewObjectName(namespace, name),
	}
}

func newTestListers(obj runtime.Object) util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister] {
	listers := util.NewConcurrentMap[schema.GroupVersionResource, cache.GenericLister]()
	listers.Set(testGVR, &fakeGenericLister{obj: obj})
	return listers
}

func TestUpdateObjectStatus_DoesNotMutateInformerCacheOnUpdateFailure(t *testing.T) {
	ctx := context.Background()

	originalStatus := map[string]interface{}{
		"replicas": int64(1),
	}

	cachedObj := newTestUnstructured("test-ns", "test-deploy", originalStatus)

	fakeDynClient := dynamicfake.NewSimpleDynamicClient(newFakeScheme(), cachedObj)

	fakeDynClient.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, fmt.Errorf("simulated API server failure")
	})

	c := &Controller{wdsDynClient: fakeDynClient}
	listers := newTestListers(cachedObj)
	objectIdentifier := newTestObjectIdentifier("test-ns", "test-deploy")

	newStatus := map[string]interface{}{
		"replicas": int64(5),
	}

	err := c.updateObjectStatus(ctx, objectIdentifier, newStatus, listers, false)
	if err == nil {
		t.Fatal("expected an error when UpdateStatus fails, got nil")
	}

	actualStatus, _, _ := unstructured.NestedMap(cachedObj.Object, "status")
	if !reflect.DeepEqual(actualStatus, originalStatus) {
		t.Errorf(
			"BUG: informer cache was mutated after UpdateStatus failure.\ngot:  %v\nwant: %v",
			actualStatus,
			originalStatus,
		)
	}
}

func TestUpdateObjectStatus_UpdatesStatusWhenDifferent(t *testing.T) {
	ctx := context.Background()

	cachedObj := newTestUnstructured("test-ns", "test-deploy", map[string]interface{}{
		"replicas": int64(1),
	})

	fakeDynClient := dynamicfake.NewSimpleDynamicClient(newFakeScheme(), cachedObj)

	updateCalled := false
	fakeDynClient.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		if updateAction, ok := action.(k8stesting.UpdateAction); ok && updateAction.GetSubresource() == "status" {
			updateCalled = true
		}
		return false, nil, nil
	})

	c := &Controller{wdsDynClient: fakeDynClient}
	listers := newTestListers(cachedObj)
	objectIdentifier := newTestObjectIdentifier("test-ns", "test-deploy")

	err := c.updateObjectStatus(ctx, objectIdentifier, map[string]interface{}{"replicas": int64(5)}, listers, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !updateCalled {
		t.Error("expected UpdateStatus to be called when status differs, but it was not")
	}
}

func TestUpdateObjectStatus_NoOpWhenStatusAlreadyMatches(t *testing.T) {
	ctx := context.Background()

	existingStatus := map[string]interface{}{
		"replicas": int64(3),
	}

	cachedObj := newTestUnstructured("test-ns", "test-deploy", existingStatus)

	fakeDynClient := dynamicfake.NewSimpleDynamicClient(newFakeScheme(), cachedObj)

	statusUpdateCalled := false
	fakeDynClient.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		// only flag if this is a status subresource update, not a label update
		if action.(k8stesting.UpdateAction).GetSubresource() == "status" {
			statusUpdateCalled = true
		}
		return false, nil, nil
	})

	c := &Controller{wdsDynClient: fakeDynClient}
	listers := newTestListers(cachedObj)
	objectIdentifier := newTestObjectIdentifier("test-ns", "test-deploy")

	err := c.updateObjectStatus(ctx, objectIdentifier, existingStatus, listers, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if statusUpdateCalled {
		t.Error("UpdateStatus subresource was called even though status had not changed")
	}
}

func TestUpdateObjectStatus_NoOpWhenNilStatusAndNoLabel(t *testing.T) {
	ctx := context.Background()

	cachedObj := newTestUnstructured("test-ns", "test-deploy", nil)

	fakeDynClient := dynamicfake.NewSimpleDynamicClient(newFakeScheme(), cachedObj)

	updateCalled := false
	fakeDynClient.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		updateCalled = true
		return true, nil, nil
	})

	c := &Controller{wdsDynClient: fakeDynClient}
	listers := newTestListers(cachedObj)
	objectIdentifier := newTestObjectIdentifier("test-ns", "test-deploy")

	err := c.updateObjectStatus(ctx, objectIdentifier, nil, listers, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updateCalled {
		t.Error("UpdateStatus was called when status was nil and no label present")
	}
}
