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
	"testing"

	workv1 "open-cluster-management.io/api/work/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubestellar/kubestellar/pkg/transport"
)

func testObject(apiVersion, kind, ns, name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata":   map[string]any{"name": name},
	}}
	if ns != "" {
		obj.Object["metadata"].(map[string]any)["namespace"] = ns
	}
	if kind == "Pod" {
		_ = unstructured.SetNestedSlice(obj.Object, []any{map[string]any{
			"name":  "c",
			"image": "nginx:1.27.1",
		}}, "spec", "containers")
	}
	return obj
}

func testKindToResource(kind string) func(schema.GroupKind) string {
	return func(gk schema.GroupKind) string {
		switch gk.Kind {
		case "Pod":
			return "pods"
		case "Deployment":
			return "deployments"
		default:
			return kind + "s"
		}
	}
}

func manifestConfig(wrapped interface{}, name string) (workv1.ManifestConfigOption, bool) {
	mw := wrapped.(*workv1.ManifestWork)
	for _, cfg := range mw.Spec.ManifestConfigs {
		if cfg.ResourceIdentifier.Name == name {
			return cfg, true
		}
	}
	return workv1.ManifestConfigOption{}, false
}

func TestWrapObjectsUpdateStrategy(t *testing.T) {
	kindToResource := testKindToResource("")

	t.Run("pod gets server-side apply", func(t *testing.T) {
		wrapees := []transport.Wrapee{transport.NewWrapee(testObject("v1", "Pod", "ns", "mypod"), false)}
		wrapped := NewOCMTransport().WrapObjects(wrapees, kindToResource)
		cfg, ok := manifestConfig(wrapped, "mypod")
		if !ok {
			t.Fatalf("expected a ManifestConfigOption for pod mypod")
		}
		if cfg.UpdateStrategy == nil || cfg.UpdateStrategy.Type != workv1.UpdateStrategyTypeServerSideApply {
			t.Errorf("expected ServerSideApply strategy for pod, got %v", cfg.UpdateStrategy)
		}
		if cfg.ResourceIdentifier.Group != "" || cfg.ResourceIdentifier.Resource != "pods" ||
			cfg.ResourceIdentifier.Namespace != "ns" {
			t.Errorf("unexpected resource identifier: %+v", cfg.ResourceIdentifier)
		}
	})

	t.Run("createOnly pod takes precedence over server-side apply", func(t *testing.T) {
		wrapees := []transport.Wrapee{transport.NewWrapee(testObject("v1", "Pod", "ns", "mypod"), true)}
		wrapped := NewOCMTransport().WrapObjects(wrapees, kindToResource)
		cfg, ok := manifestConfig(wrapped, "mypod")
		if !ok {
			t.Fatalf("expected a ManifestConfigOption for pod mypod")
		}
		if cfg.UpdateStrategy == nil || cfg.UpdateStrategy.Type != workv1.UpdateStrategyTypeCreateOnly {
			t.Errorf("expected CreateOnly strategy for createOnly pod, got %v", cfg.UpdateStrategy)
		}
	})

	t.Run("deployment gets no config option", func(t *testing.T) {
		wrapees := []transport.Wrapee{transport.NewWrapee(testObject("apps/v1", "Deployment", "ns", "mydeploy"), false)}
		wrapped := NewOCMTransport().WrapObjects(wrapees, kindToResource)
		mw := wrapped.(*workv1.ManifestWork)
		if len(mw.Spec.ManifestConfigs) != 0 {
			t.Errorf("expected no ManifestConfigOption for deployment, got %d", len(mw.Spec.ManifestConfigs))
		}
	})

	t.Run("metav1 wrapper not required", func(t *testing.T) {
		// ensure the object stored in the manifest keeps its spec intact (e.g. image)
		wrapees := []transport.Wrapee{transport.NewWrapee(testObject("v1", "Pod", "ns", "mypod"), false)}
		wrapped := NewOCMTransport().WrapObjects(wrapees, kindToResource)
		mw := wrapped.(*workv1.ManifestWork)
		obj := mw.Spec.Workload.Manifests[0].Object.(*unstructured.Unstructured)
		containers, found, err := unstructured.NestedSlice(obj.Object, "spec", "containers")
		if err != nil || !found || len(containers) == 0 {
			t.Fatalf("expected containers to be present, got %v (err=%v)", containers, err)
		}
		c0, ok := containers[0].(map[string]any)
		if !ok || c0["image"] != "nginx:1.27.1" {
			t.Errorf("expected container image to be preserved, got %v (ok=%v)", c0, ok)
		}
	})
}
