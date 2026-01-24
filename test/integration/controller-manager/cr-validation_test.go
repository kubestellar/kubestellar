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

package cmtest

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/yaml"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	kscrd "github.com/kubestellar/kubestellar/pkg/crd"
)

func TestBindingPolicyValidation(t *testing.T) {
	logger := klog.Background()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info("Setting up test environment")
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "..", "config", "crd", "bases"),
		},
		ErrorIfCRDPathMissing: true,
	}

	logger.Info("Starting test environment")
	config, err := testEnv.Start()
	if err != nil {
		t.Fatalf("Failed to start test environment: %s", err)
	}

	t.Cleanup(func() {
		logger.Info("Stopping test environment")
		if err := testEnv.Stop(); err != nil {
			t.Errorf("Failed to stop test environment: %s", err)
		}
	})

	logger.Info("Test environment started successfully")

	configCopy := *config
	config4json := &configCopy
	config4json.ContentType = "application/json"
	logger.Info("REST config for JSON marshaling", "config", config4json)
	apiextClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create apiextensions client: %s", err)
	}
	err = kscrd.ApplyCRDs(ctx, "test", apiextClient.ApiextensionsV1().CustomResourceDefinitions(), logger)
	if err != nil {
		t.Fatalf("Failed to apply KubeStellar CRDs: %s", err)
	}
	bpCRD, err := apiextClient.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, "bindingpolicies.control.kubestellar.io", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to fetch BindingPolicy CRD: %s", err)
	}
	bpYAML, err := yaml.Marshal(bpCRD)
	if err != nil {
		t.Fatalf("Failed to marshal BindingPolicy CRD: %s", err)
	}
	t.Log("Fetched BindingPolicy CRD:\n" + string(bpYAML))
	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create dynamic client: %s", err)
	}
	dynIfc := dynClient.Resource(ksapi.SchemeGroupVersion.WithResource("bindingpolicies"))
	for _, testCase := range []struct {
		name     string
		specJSON string
		expectOK bool
	}{
		{name: "junk-field-in-spec", specJSON: `{"junk": 1}`},
		{name: "junk-field-in-policy", specJSON: `{"downsync": [{"junq": true}]}`},
		{name: "empty-object-test", specJSON: `{"downsync": [{"createOnly": true}]}`},
		{name: "invalid-statusCollection-field", specJSON: `{"downsync": [{"statusCollection": {"statusCollectors": ["phred"]}}]}`},
		{name: "match-all-resources", specJSON: `{"downsync": [{"resources": ["*"]}]}`, expectOK: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			var specMap map[string]interface{}
			err := json.Unmarshal([]byte(testCase.specJSON), &specMap)
			if err != nil {
				t.Fatalf("Failed to unmarshal test case spec: %s", err)
			}
			bpMap := map[string]interface{}{
				"apiVersion": ksapi.SchemeGroupVersion.String(),
				"kind":       "BindingPolicy",
				"metadata": map[string]interface{}{
					"name": testCase.name,
				},
				"spec": specMap,
			}
			bpUns := &unstructured.Unstructured{Object: bpMap}
			created, err := dynIfc.Create(ctx, bpUns, metav1.CreateOptions{})
			if testCase.expectOK {
				if err != nil {
					t.Fatalf("Expected successful create but got error: %s", err)
				}
				logger.Info("Created BindingPolicy", "name", testCase.name)
				err = dynIfc.Delete(ctx, created.GetName(), metav1.DeleteOptions{})
				if err != nil {
					t.Errorf("Failed to delete created BindingPolicy: %s", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Expected create to fail but it succeeded")
				}
				logger.Info("Create failed as expected", "name", testCase.name, "error", err)
			}
		})
	}
	logger.Info("Success")
}
