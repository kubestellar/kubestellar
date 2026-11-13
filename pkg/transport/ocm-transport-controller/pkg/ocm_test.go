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

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	workv1 "open-cluster-management.io/api/work/v1"

	"github.com/kubestellar/kubestellar/pkg/transport"
)

func TestWrapObjectsUpdateStrategies(t *testing.T) {
	newConfigMap := func(name string) *unstructured.Unstructured {
		obj := &unstructured.Unstructured{}
		obj.SetAPIVersion("v1")
		obj.SetKind("ConfigMap")
		obj.SetNamespace("test")
		obj.SetName(name)
		return obj
	}

	wrapped := NewOCMTransport().WrapObjects([]transport.Wrapee{
		{Object: newConfigMap("updated")},
		{Object: newConfigMap("create-only"), CreateOnly: true},
	}, func(gk schema.GroupKind) string {
		if gk != (schema.GroupKind{Kind: "ConfigMap"}) {
			t.Fatalf("unexpected GroupKind: %v", gk)
		}
		return "configmaps"
	}).(*workv1.ManifestWork)

	if len(wrapped.Spec.Workload.Manifests) != 2 || len(wrapped.Spec.ManifestConfigs) != 2 {
		t.Fatalf("expected two manifests and two configs, got %d manifests and %d configs",
			len(wrapped.Spec.Workload.Manifests), len(wrapped.Spec.ManifestConfigs))
	}

	for i, want := range []struct {
		name     string
		strategy workv1.UpdateStrategyType
	}{
		{name: "updated", strategy: workv1.UpdateStrategyTypeServerSideApply},
		{name: "create-only", strategy: workv1.UpdateStrategyTypeCreateOnly},
	} {
		config := wrapped.Spec.ManifestConfigs[i]
		if config.ResourceIdentifier != (workv1.ResourceIdentifier{
			Resource: "configmaps", Namespace: "test", Name: want.name,
		}) {
			t.Errorf("config %d has wrong resource identifier: %+v", i, config.ResourceIdentifier)
		}
		if config.UpdateStrategy == nil || config.UpdateStrategy.Type != want.strategy {
			t.Errorf("config %d has wrong update strategy: %+v", i, config.UpdateStrategy)
		}
	}
}
