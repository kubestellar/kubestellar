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

package transport

import (
	"context"
	"slices"
	"testing"

	workapi "open-cluster-management.io/api/work/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/ktesting"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	ksclientfake "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned/fake"
)

func newTestWrappedObject(namespace, name, bindingName, wdsName string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(workapi.GroupVersion.String())
	obj.SetKind("ManifestWork")
	obj.SetNamespace(namespace)
	obj.SetName(name)
	obj.SetLabels(map[string]string{
		originOwnerReferenceLabel: bindingName,
		originWdsLabel:            wdsName,
	})
	return obj
}

// TestDeleteWrappedObjectsOfDroppedDestination checks that deleting a Binding removes
// every wrapped object it owns, including one in the mailbox namespace of a destination
// that is no longer listed in the Binding's spec.
func TestDeleteWrappedObjectsOfDroppedDestination(t *testing.T) {
	_, ctx := ktesting.NewTestContext(t)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	const (
		bindingName = "b1"
		wdsName     = "wds1"
	)
	gvr := workapi.GroupVersion.WithResource("manifestworks")
	itsClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[k8sschema.GroupVersionResource]string{gvr: "ManifestWorkList"})

	type ref struct{ namespace, name string }
	owned := []ref{
		{"wec1", bindingName + "-" + wdsName},        // destination still listed
		{"wec2", bindingName + "-" + wdsName},        // destination that was dropped
		{"wec2", bindingName + "-" + wdsName + "-1"}, // another shard for the dropped destination
	}
	notOwned := []*unstructured.Unstructured{
		newTestWrappedObject("wec1", "b2-"+wdsName, "b2", wdsName),             // other Binding
		newTestWrappedObject("wec2", bindingName+"-wds2", bindingName, "wds2"), // same Binding name, other WDS
	}
	toCreate := slices.Clone(notOwned)
	for _, r := range owned {
		toCreate = append(toCreate, newTestWrappedObject(r.namespace, r.name, bindingName, wdsName))
	}
	for _, obj := range toCreate {
		if _, err := itsClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(ctx, obj, metav1.CreateOptions{}); err != nil {
			t.Fatalf("failed to create wrapped object %s/%s: %v", obj.GetNamespace(), obj.GetName(), err)
		}
	}

	now := metav1.Now()
	binding := &ksapi.Binding{
		TypeMeta: typeMeta("Binding", ksapi.GroupVersion),
		ObjectMeta: metav1.ObjectMeta{
			Name:              bindingName,
			Finalizers:        []string{transportFinalizer},
			DeletionTimestamp: &now,
		},
		Spec: ksapi.BindingSpec{
			Destinations: []ksapi.Destination{{ClusterId: "wec1"}},
		},
	}
	ksClient := ksclientfake.NewSimpleClientset(binding)

	ctlr := &genericTransportController{
		logger:           klog.FromContext(ctx),
		bindingClient:    ksClient.ControlV1alpha1().Bindings(),
		transportClient:  itsClient,
		wrappedObjectGVR: gvr,
		wdsName:          wdsName,
		customTransformCollection: newCustomTransformCollection(nil, func(string, string) ([]any, error) {
			return nil, nil
		}, func(any) {}),
	}

	if err := ctlr.deleteWrappedObjectsAndFinalizer(ctx, binding.DeepCopy()); err != nil {
		t.Fatalf("deleteWrappedObjectsAndFinalizer failed: %v", err)
	}

	for _, r := range owned {
		_, err := itsClient.Resource(gvr).Namespace(r.namespace).Get(ctx, r.name, metav1.GetOptions{})
		if !apierrors.IsNotFound(err) {
			t.Errorf("expected wrapped object %s/%s to be deleted, Get returned err=%v", r.namespace, r.name, err)
		}
	}
	for _, obj := range notOwned {
		if _, err := itsClient.Resource(gvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), metav1.GetOptions{}); err != nil {
			t.Errorf("wrapped object %s/%s not owned by the Binding should be kept, Get returned err=%v", obj.GetNamespace(), obj.GetName(), err)
		}
	}

	got, err := ksClient.ControlV1alpha1().Bindings().Get(ctx, bindingName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get Binding: %v", err)
	}
	if slices.Contains(got.Finalizers, transportFinalizer) {
		t.Errorf("expected finalizer %q to be removed, finalizers are %v", transportFinalizer, got.Finalizers)
	}
}
