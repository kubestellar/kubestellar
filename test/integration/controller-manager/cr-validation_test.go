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
	"testing"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2/ktesting"
	kastesting "k8s.io/kubernetes/cmd/kube-apiserver/app/testing"
	"k8s.io/kubernetes/test/integration/framework"
	"sigs.k8s.io/yaml"

	ksapi "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	kscrd "github.com/kubestellar/kubestellar/pkg/crd"
)

func TestBindingPolicyValidation(t *testing.T) {
	testWriter := framework.NewTBWriter(t)
	logger, ctx := ktesting.NewTestContext(t)
	logger.Info("Starting etcd server")
	framework.StartEtcd(t, testWriter)
	logger.Info("Starting TestController")
	t.Log("Beginning TestController")
	ctx, cancel := context.WithCancel(ctx)
	testServer, err := kastesting.StartTestServer(t, kastesting.NewDefaultTestServerOptions(), []string{}, framework.SharedEtcd())
	if err != nil {
		t.Fatalf("Failed to kastesting.StartTestServer: %s", err)
	}
	fullTeardwon := func() {
		cancel()
		testServer.TearDownFn()
	}
	t.Cleanup(fullTeardwon)
	config := testServer.ClientConfig
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %s", err)
	}
	logger.Info("Started test server", "config", config)
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
	converter := runtime.DefaultUnstructuredConverter
	for _, testCase := range []struct {
		name     string
		specJSON string
		expectOK bool
	}{
		{name: "junk-field-in-spec", specJSON: `{"junk": 1}`},
		{name: "junk-field-in-policy", specJSON: `{"downsync": [{"junq": true}]}`},
		// Test: empty ObjectTest (no selector fields, valid fields only)
		{name: "empty-object-test", specJSON: `{"downsync": [{"createOnly": true}]}`},
		// Test: use of now-invalid 'statusCollection' field (should fail OpenAPI validation)
		{name: "invalid-statusCollection-field", specJSON: `{"downsync": [{"statusCollection": {"statusCollectors": ["phred"]}}]}`},
		{name: "match-all-resources", specJSON: `{"downsync": [{"resources": ["*"]}]}`, expectOK: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			blankBP := ksapi.BindingPolicy{
				TypeMeta:   metav1.TypeMeta{APIVersion: ksapi.SchemeGroupVersion.String(), Kind: "BindingPolicy"},
				ObjectMeta: metav1.ObjectMeta{Name: testCase.name},
			}
			srcM, err := converter.ToUnstructured(&blankBP)
			if err != nil {
				t.Fatalf("Failed to convert to Unstructured: %s", err)
			}
			spec := map[string]any{}
			err = json.Unmarshal([]byte(testCase.specJSON), &spec)
			if err != nil {
				t.Fatalf("Failed to unmarshal spec: %s", err)
			}
			unstructured.SetNestedMap(srcM, spec, "spec")
			srcU := &unstructured.Unstructured{Object: srcM}
			echo, err := dynIfc.Create(ctx, srcU, metav1.CreateOptions{FieldValidation: metav1.FieldValidationStrict, FieldManager: "test"})
			if testCase.expectOK == (err == nil) {
				t.Logf("Success; err=%s", err)
			} else if err == nil {
				t.Errorf("Failed to get expected error; echo=%#v", echo.Object)
			} else {
				t.Errorf("Got unexpected error: %s; srcM=%#v", err, srcM)
			}
			if err == nil {
				dynIfc.Delete(ctx, srcU.GetName(), metav1.DeleteOptions{})
			}
		})
	}
}
