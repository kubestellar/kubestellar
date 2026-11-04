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

package ocm

import (
	"reflect"
	"testing"

	workv1 "open-cluster-management.io/api/work/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/transport"
	"github.com/kubestellar/kubestellar/pkg/util"
)

func TestWrapObjects(t *testing.T) {
	ocmTransport := NewOCMTransport()

	// 1. Regular namespaced object without CreateOnly (ConfigMap)
	configMap := &unstructured.Unstructured{}
	configMap.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})
	configMap.SetName("test-cm")
	configMap.SetNamespace("default")
	wrapee1 := transport.NewWrapee(configMap, false)

	// 2. Regular namespaced object with CreateOnly (Deployment)
	deploy := &unstructured.Unstructured{}
	deploy.SetGroupVersionKind(schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"})
	deploy.SetName("test-deploy")
	deploy.SetNamespace("default")
	wrapee2 := transport.NewWrapee(deploy, true)

	// 3. CSIDriver without CreateOnly (cluster-scoped, should use ServerSideApply)
	csiDriver := &unstructured.Unstructured{}
	csiDriver.SetGroupVersionKind(schema.GroupVersionKind{Group: "storage.k8s.io", Version: "v1", Kind: "CSIDriver"})
	csiDriver.SetName("test-csidriver")
	wrapee3 := transport.NewWrapee(csiDriver, false)

	// 4. CSIDriver with CreateOnly (cluster-scoped, CreateOnly takes precedence)
	csiDriverCreateOnly := &unstructured.Unstructured{}
	csiDriverCreateOnly.SetGroupVersionKind(schema.GroupVersionKind{Group: "storage.k8s.io", Version: "v1", Kind: "CSIDriver"})
	csiDriverCreateOnly.SetName("test-csidriver-create-only")
	wrapee4 := transport.NewWrapee(csiDriverCreateOnly, true)

	wrapees := []transport.Wrapee{wrapee1, wrapee2, wrapee3, wrapee4}

	kindToResource := func(gk schema.GroupKind) string {
		switch gk {
		case schema.GroupKind{Group: "", Kind: "ConfigMap"}:
			return "configmaps"
		case schema.GroupKind{Group: "apps", Kind: "Deployment"}:
			return "deployments"
		case schema.GroupKind{Group: "storage.k8s.io", Kind: "CSIDriver"}:
			return "csidrivers"
		default:
			return ""
		}
	}

	result := ocmTransport.WrapObjects(wrapees, kindToResource)
	manifestWork, ok := result.(*workv1.ManifestWork)
	if !ok {
		t.Fatalf("expected *workv1.ManifestWork, got %T", result)
	}

	if manifestWork.Kind != wrappedObjectKind || manifestWork.APIVersion != wrappedObjectAPIVersion {
		t.Errorf("unexpected TypeMeta: got %s / %s, expected %s / %s",
			manifestWork.Kind, manifestWork.APIVersion, wrappedObjectKind, wrappedObjectAPIVersion)
	}

	if len(manifestWork.Spec.Workload.Manifests) != len(wrapees) {
		t.Fatalf("expected %d manifests, got %d", len(wrapees), len(manifestWork.Spec.Workload.Manifests))
	}

	// Verify ManifestConfigs:
	// wrapee1: no config option
	// wrapee2: CreateOnly on Deployment
	// wrapee3: ServerSideApply on CSIDriver
	// wrapee4: CreateOnly on CSIDriver
	expectedConfigs := 3
	if len(manifestWork.Spec.ManifestConfigs) != expectedConfigs {
		t.Fatalf("expected %d manifest configs, got %d", expectedConfigs, len(manifestWork.Spec.ManifestConfigs))
	}

	// Check wrapee2 config (Deployment, CreateOnly)
	cfg0 := manifestWork.Spec.ManifestConfigs[0]
	expectedIdentifier0 := workv1.ResourceIdentifier{
		Group:     "apps",
		Resource:  "deployments",
		Namespace: "default",
		Name:      "test-deploy",
	}
	if !reflect.DeepEqual(cfg0.ResourceIdentifier, expectedIdentifier0) {
		t.Errorf("config[0] unexpected identifier: got %+v, expected %+v", cfg0.ResourceIdentifier, expectedIdentifier0)
	}
	if cfg0.UpdateStrategy == nil || cfg0.UpdateStrategy.Type != workv1.UpdateStrategyTypeCreateOnly {
		t.Errorf("config[0] expected CreateOnly update strategy, got %+v", cfg0.UpdateStrategy)
	}

	// Check wrapee3 config (CSIDriver, ServerSideApply)
	cfg1 := manifestWork.Spec.ManifestConfigs[1]
	expectedIdentifier1 := workv1.ResourceIdentifier{
		Group:     "storage.k8s.io",
		Resource:  "csidrivers",
		Namespace: "",
		Name:      "test-csidriver",
	}
	if !reflect.DeepEqual(cfg1.ResourceIdentifier, expectedIdentifier1) {
		t.Errorf("config[1] unexpected identifier: got %+v, expected %+v", cfg1.ResourceIdentifier, expectedIdentifier1)
	}
	if cfg1.UpdateStrategy == nil || cfg1.UpdateStrategy.Type != workv1.UpdateStrategyTypeServerSideApply {
		t.Errorf("config[1] expected ServerSideApply update strategy, got %+v", cfg1.UpdateStrategy)
	}

	// Check wrapee4 config (CSIDriver, CreateOnly)
	cfg2 := manifestWork.Spec.ManifestConfigs[2]
	expectedIdentifier2 := workv1.ResourceIdentifier{
		Group:     "storage.k8s.io",
		Resource:  "csidrivers",
		Namespace: "",
		Name:      "test-csidriver-create-only",
	}
	if !reflect.DeepEqual(cfg2.ResourceIdentifier, expectedIdentifier2) {
		t.Errorf("config[2] unexpected identifier: got %+v, expected %+v", cfg2.ResourceIdentifier, expectedIdentifier2)
	}
	if cfg2.UpdateStrategy == nil || cfg2.UpdateStrategy.Type != workv1.UpdateStrategyTypeCreateOnly {
		t.Errorf("config[2] expected CreateOnly update strategy, got %+v", cfg2.UpdateStrategy)
	}

	// Test with empty wrapees slice and nil kindToResource
	emptyResult := ocmTransport.WrapObjects(nil, nil)
	emptyWork, ok := emptyResult.(*workv1.ManifestWork)
	if !ok {
		t.Fatalf("expected *workv1.ManifestWork for empty wrapees, got %T", emptyResult)
	}
	if len(emptyWork.Spec.Workload.Manifests) != 0 {
		t.Errorf("expected 0 manifests for empty wrapees, got %d", len(emptyWork.Spec.Workload.Manifests))
	}
	if len(emptyWork.Spec.ManifestConfigs) != 0 {
		t.Errorf("expected 0 manifest configs for empty wrapees, got %d", len(emptyWork.Spec.ManifestConfigs))
	}
}

func TestUnwrapObjects(t *testing.T) {
	ocmTransport := NewOCMTransport()

	cm := &unstructured.Unstructured{}
	cm.SetGroupVersionKind(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"})
	cm.SetName("my-cm")
	cm.SetNamespace("default")

	driver := &unstructured.Unstructured{}
	driver.SetGroupVersionKind(schema.GroupVersionKind{Group: "storage.k8s.io", Version: "v1", Kind: "CSIDriver"})
	driver.SetName("my-driver")

	manifestWork := &workv1.ManifestWork{
		Spec: workv1.ManifestWorkSpec{
			Workload: workv1.ManifestsTemplate{
				Manifests: []workv1.Manifest{
					{RawExtension: runtime.RawExtension{Object: cm}},
					{RawExtension: runtime.RawExtension{Object: driver}},
				},
			},
		},
	}

	gloss, err := ocmTransport.UnwrapObjects(manifestWork, nil)
	if err != nil {
		t.Fatalf("unexpected error unwrapping ManifestWork: %v", err)
	}

	expectedGKCM := schema.GroupKind{Group: "", Kind: "ConfigMap"}
	expectedGKDriver := schema.GroupKind{Group: "storage.k8s.io", Kind: "CSIDriver"}

	if !gloss.Has(util.GKObjRef{GK: expectedGKCM, OR: klog.KObj(cm)}) {
		t.Errorf("expected gloss to contain ConfigMap %v", cm)
	}
	if !gloss.Has(util.GKObjRef{GK: expectedGKDriver, OR: klog.KObj(driver)}) {
		t.Errorf("expected gloss to contain CSIDriver %v", driver)
	}

	// Test unwrapping unstructured ManifestWork
	unstrWork := &unstructured.Unstructured{
		Object: map[string]any{
			"spec": map[string]any{
				"workload": map[string]any{
					"manifests": []any{
						map[string]any{
							"apiVersion": "v1",
							"kind":       "ConfigMap",
							"metadata": map[string]any{
								"name":      "my-cm",
								"namespace": "default",
							},
						},
					},
				},
			},
		},
	}
	glossUnstr, err := ocmTransport.UnwrapObjects(unstrWork, nil)
	if err != nil {
		t.Fatalf("unexpected error unwrapping unstructured ManifestWork: %v", err)
	}
	if !glossUnstr.Has(util.GKObjRef{GK: expectedGKCM, OR: klog.KObj(cm)}) {
		t.Errorf("expected gloss to contain ConfigMap from unstructured")
	}

	// Test error case: manifest with nil Object
	nilManifestWork := &workv1.ManifestWork{
		Spec: workv1.ManifestWorkSpec{
			Workload: workv1.ManifestsTemplate{
				Manifests: []workv1.Manifest{
					{RawExtension: runtime.RawExtension{Object: nil}},
				},
			},
		},
	}
	_, err = ocmTransport.UnwrapObjects(nilManifestWork, nil)
	if err == nil {
		t.Errorf("expected error when unwrap ManifestWork has nil Object")
	}
}

func TestManifestConfigOptionHelpers(t *testing.T) {
	mc := workv1.ManifestConfigOption{
		ResourceIdentifier: workv1.ResourceIdentifier{
			Group:     "storage.k8s.io",
			Resource:  "csidrivers",
			Namespace: "",
			Name:      "test-csidriver",
		},
		UpdateStrategy: &serverSideApplyStrategy,
	}

	if !reflect.DeepEqual(ManifestConfigOptionResourceIdentifier(mc), mc.ResourceIdentifier) {
		t.Errorf("ManifestConfigOptionResourceIdentifier returned %+v, expected %+v",
			ManifestConfigOptionResourceIdentifier(mc), mc.ResourceIdentifier)
	}

	if ManifestConfigOptionUpdateStrategy(mc) != mc.UpdateStrategy {
		t.Errorf("ManifestConfigOptionUpdateStrategy returned %+v, expected %+v",
			ManifestConfigOptionUpdateStrategy(mc), mc.UpdateStrategy)
	}
}
