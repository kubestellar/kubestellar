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

package ocm

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	workv1 "open-cluster-management.io/api/work/v1"

	"github.com/kubestellar/kubestellar/pkg/transport"
)

func TestWrapObjects_ServiceAccount(t *testing.T) {
	ocm := NewOCMTransport()

	sa := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ServiceAccount",
			"metadata": map[string]interface{}{
				"name":      "test-sa",
				"namespace": "test-ns",
			},
		},
	}

	pod := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "test-ns",
			},
		},
	}

	wrapees := []transport.Wrapee{
		{Object: sa},
		{Object: pod},
	}

	kindToResource := func(gk schema.GroupKind) string {
		switch gk.Kind {
		case "ServiceAccount":
			return "serviceaccounts"
		case "Pod":
			return "pods"
		default:
			return "unknown"
		}
	}

	manifestWorkObj := ocm.WrapObjects(wrapees, kindToResource)
	manifestWork, ok := manifestWorkObj.(*workv1.ManifestWork)
	if !ok {
		t.Fatalf("expected ManifestWork, got %T", manifestWorkObj)
	}

	configs := manifestWork.Spec.ManifestConfigs
	if len(configs) != 1 {
		t.Fatalf("expected exactly 1 config option for ServiceAccount, got %d", len(configs))
	}

	config := configs[0]
	if config.ResourceIdentifier.Resource != "serviceaccounts" || config.ResourceIdentifier.Name != "test-sa" {
		t.Errorf("unexpected resource identifier: %v", config.ResourceIdentifier)
	}

	if config.UpdateStrategy == nil || config.UpdateStrategy.Type != workv1.UpdateStrategyTypeServerSideApply {
		t.Errorf("expected ServerSideApply update strategy for ServiceAccount, got %v", config.UpdateStrategy)
	}
	if config.UpdateStrategy.ServerSideApply == nil || !config.UpdateStrategy.ServerSideApply.Force {
		t.Errorf("expected Force: true in ServerSideApply strategy, got %v", config.UpdateStrategy.ServerSideApply)
	}
	if config.UpdateStrategy.ServerSideApply.FieldManager != "kubestellar-transport-controller" {
		t.Errorf("expected FieldManager: kubestellar-transport-controller in ServerSideApply strategy, got %v", config.UpdateStrategy.ServerSideApply.FieldManager)
	}
}
