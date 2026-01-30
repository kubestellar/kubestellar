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
	"fmt"
	"io"
	"net/http"
	"testing"

	apiextensionsapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// Shared constants for all integration tests
const managedClusterCRDURL = "https://raw.githubusercontent.com/open-cluster-management-io/api/v0.12.0/cluster/v1/0000_00_clusters.open-cluster-management.io_managedclusters.crd.yaml"
const manifestWorkCRDURL = "https://raw.githubusercontent.com/open-cluster-management-io/api/v0.12.0/work/v1/0000_00_work.open-cluster-management.io_manifestworks.crd.yaml"

var crdGVK = apiextensionsapi.SchemeGroupVersion.WithKind("CustomResourceDefinition")

// Helper functions shared across integration tests

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

func createCRD(t *testing.T, ctx context.Context, kind, url string, serializer *k8sjson.Serializer, apiextClient apiextensionsclientset.Interface) (*apiextensionsapi.CustomResourceDefinition, error) {
	crdYAML, err := urlGet(url)
	if err != nil {
		t.Fatalf("Failed to read %s CRD from %s: %s", kind, url, err)
	}
	return createCRDFromLiteral(t, ctx, kind, crdYAML, serializer, apiextClient)
}

func urlGet(urlStr string) (string, error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get(%s) returned status %d", urlStr, resp.StatusCode)
	}
	read, err := io.ReadAll(resp.Body)
	readS := string(read)
	if err != nil {
		return "", err
	}
	return readS, nil
}
