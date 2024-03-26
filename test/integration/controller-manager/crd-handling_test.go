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
	"testing"
	"time"

	apiextensionsapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2/ktesting"
	kastesting "k8s.io/kubernetes/cmd/kube-apiserver/app/testing"
	"k8s.io/kubernetes/test/integration/framework"

	"github.com/kubestellar/kubestellar/pkg/binding"
)

// An integration test for the binding controller.
// This test uses an in-process kube-apiserver created by k8s.io/kubernetes/cmd/kube-apiserver/app/testing.
// That code launches an external insecure etcd server.
// YOU MUST HAVE THE ETCD BINARY ON YOUR `$PATH`.
//
// This test exercises the crd handling functionality.
func TestCRDHandling(t *testing.T) {
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
	if err != nil {
		t.Fatalf("Failed to create KubeStellar client: %s", err)
	}
	apiextClient, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create apiextensions client: %s", err)
	}
	scheme := runtime.NewScheme()
	err = apiextensionsapi.AddToScheme(scheme)
	if err != nil {
		t.Fatalf("Failed to apiextensionsapi.AddToScheme(scheme): %s", err)
	}
	serializer := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, scheme, scheme)
	createCRD(t, ctx, "ManagedCluster", managedClusterCRDURL, serializer, apiextClient)
	createCRD(t, ctx, "ManifestWork", manifestWorkCRDURL, serializer, apiextClient)
	time.Sleep(5 * time.Second)
	ctlr, err := binding.NewController(logger, config4json, config, "test-wds", nil)
	if err != nil {
		t.Fatalf("Failed to create controller: %s", err)
	}
	logger.Info("About to EnsureCRDs")
	err = ctlr.EnsureCRDs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("CRDs ensured")
	err = ctlr.AppendKSResources(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Appended KS resources to discovered lists")
	err = ctlr.Start(ctx, 4, make(chan interface{}, 1))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Second)

	initNumInformers := ctlr.GetInformers().Len()
	initNumListers := ctlr.GetListers().Len()
	logger.Info("Check controller's initial watch", "initNumInformers", initNumInformers, "initNumListers", initNumListers)
	if initNumInformers != initNumListers {
		t.Fatalf("Mismatch, initNumInformers=%d, initNumListers=%d", initNumInformers, initNumListers)
	}

	crd, _ := createCRDFromLiteral(t, ctx, "CR1", crd1Literal, serializer, apiextClient)
	watched := schema.GroupVersionResource{Group: "synthetic-crd.com", Version: "v2alpha1", Resource: "cr1s"}
	notWatched := schema.GroupVersionResource{Group: "synthetic-crd.com", Version: "v3beta1", Resource: "cr1s"}

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, time.Minute, false, func(ctx context.Context) (done bool, err error) {
		informers, listers := ctlr.GetInformers(), ctlr.GetListers()
		numInformers, numListers := informers.Len(), listers.Len()
		if numInformers != initNumInformers+2 {
			logger.Info("Doesn't increase", "numInformers", numInformers, "initNumInformers", initNumInformers)
			return false, nil
		}
		if numListers != initNumListers+2 {
			logger.Info("Doesn't increase", "numListers", numListers, "initNumListers", initNumListers)
			return false, nil
		}
		if numInformers != numListers {
			logger.Info("Mismatch", "numInformers", numInformers, "numListers", numListers)
			return false, nil
		}
		if _, found := informers.Get(watched); !found {
			logger.Info("Informer is missing", "gvk", watched)
			return false, nil
		}
		if _, found := listers.Get(watched); !found {
			logger.Info("Lister is missing", "gvk", watched)
			return false, nil
		}
		if _, found := informers.Get(notWatched); found {
			logger.Info("Informer unexpectedly appears", "gvk", notWatched)
			return false, nil
		}
		if _, found := listers.Get(notWatched); found {
			logger.Info("Lister unexpectedly appears", "gvk", notWatched)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("Incorrect informers/listers for %s and/or %s", watched, notWatched)
	}

	apiextClient.ApiextensionsV1().CustomResourceDefinitions().Delete(ctx, crd.Name, metav1.DeleteOptions{})

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, time.Minute, false, func(ctx context.Context) (done bool, err error) {
		informers, listers := ctlr.GetInformers(), ctlr.GetListers()
		numInformers, numListers := informers.Len(), listers.Len()
		if numInformers != initNumInformers {
			logger.Info("Doesn't reset", "numInformers", numInformers, "initNumInformers", initNumInformers)
			return false, nil
		}
		if numListers != initNumListers {
			logger.Info("Doesn't reset", "numListers", numListers, "initNumListers", initNumListers)
			return false, nil
		}
		if numInformers != numListers {
			logger.Info("Mismatch", "numInformers", numInformers, "numListers", numListers)
			return false, nil
		}
		if _, found := informers.Get(watched); found {
			logger.Info("Informer still exists", "gvk", watched)
			return false, nil
		}
		if _, found := listers.Get(watched); found {
			logger.Info("Lister still exists", "gvk", watched)
			return false, nil
		}
		if _, found := informers.Get(notWatched); found {
			logger.Info("Informer still unexpectedly appears", "gvk", notWatched)
			return false, nil
		}
		if _, found := listers.Get(notWatched); found {
			logger.Info("Lister still unexpectedly appears", "gvk", notWatched)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		t.Fatalf("Incorrect informers/listers for %s and/or %s", watched, notWatched)
	}

	logger.Info("Success")
}

func createCRDFromLiteral(t *testing.T, ctx context.Context, kind, literal string, serializer *k8sjson.Serializer,
	apiextClient apiextensionsclientset.Interface) (*apiextensionsapi.CustomResourceDefinition, error) {
	crdAny, _, err := serializer.Decode([]byte(literal), &crdGVK, &apiextensionsapi.CustomResourceDefinition{})
	if err != nil {
		t.Fatalf("Failed to Decode %s CRD: %s", kind, err)
	}
	crd := crdAny.(*apiextensionsapi.CustomResourceDefinition)
	created, err := apiextClient.ApiextensionsV1().CustomResourceDefinitions().Create(ctx, crd, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create %s CRD: %s", kind, err)
	}
	return created, nil
}

const crd1Literal string = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: cr1s.synthetic-crd.com
spec:
  group: synthetic-crd.com
  names:
    kind: CR1
    listKind: CR1List
    plural: cr1s
    singular: cr1
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              tier:
                type: string
                enum:
                - Dedicated
                - Shared
                default: Shared
          status:
            type: object
            properties:
              phase:
                type: string
        required:
        - spec
    subresources:
      status: {}
  - name: v2alpha1
    served: true
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              ownership:
                type: string
                enum:
                - Dedicated
                - Shared
                default: Shared
          status:
            type: object
            properties:
              phase:
                type: string
        required:
        - spec
    subresources:
      status: {}
  - name: v3beta1
    served: false
    storage: false
    schema:
      openAPIV3Schema:
        type: object
        properties:
          spec:
            type: object
            properties:
              possession:
                type: string
                enum:
                - Dedicated
                - Shared
                default: Shared
          status:
            type: object
            properties:
              phase:
                type: string
        required:
        - spec
    subresources:
      status: {}
`
