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

package crd

import (
	"fmt"
	"testing"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

func TestBindingPolicyAdditionalValidations(t *testing.T) {
	crds, err := readCRDs()
	if err != nil {
		t.Fatal("Failed reading CRDs from file system")
	}

	bpCRDAsList := filterCRDsByNames(crds, sets.New(
		"bindingpolicies.control.kubestellar.io",
	))
	if len(bpCRDAsList) != 1 {
		t.Fatal("Failed to get the bindingpolicies CRD out of content read from file")
	}
	bpCRDAsUnstructured := bpCRDAsList[0]

	bpCRDv1 := apiextensionsv1.CustomResourceDefinition{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(bpCRDAsUnstructured.UnstructuredContent(), &bpCRDv1)
	if err != nil {
		t.Fatal("Failed to convert CRD to apiextensionsv1.CustomResourceDefinition")
	}

	bpCRD := apiextensions.CustomResourceDefinition{}
	err = apiextensionsv1.Convert_v1_CustomResourceDefinition_To_apiextensions_CustomResourceDefinition(&bpCRDv1, &bpCRD, nil)
	if err != nil {
		t.Fatal("Failed to convert CRD to apiextensions.CustomResourceDefinition")
	}

	// We assume there is only one version in the CRD
	ver := bpCRD.Spec.Versions[0]
	crdSchema := ver.Schema
	if crdSchema == nil {
		crdSchema = bpCRD.Spec.Validation
	}

	schemaValidator, _, err := validation.NewSchemaValidator(crdSchema.OpenAPIV3Schema)
	if err != nil {
		t.Fatal("Failed to create a validator")
	}

	testCases := []struct {
		name     string
		manifest string
		wantErr  bool
	}{
		{
			name:     "BindingPolicy with invalid matchLabels value",
			manifest: bp_bad_matchLabels_value,
			wantErr:  true,
		},
		{
			name:     "BindingPolicy with valid matchLabels",
			manifest: bp_good_matchLabels,
			wantErr:  false,
		},
		{
			name:     "BindingPolicy with invalid matchExpressions key prefix",
			manifest: bp_bad_matchExpressions_key_prefix,
			wantErr:  true,
		},
		{
			name:     "BindingPolicy with invalid matchExpressions key name",
			manifest: bp_bad_matchExpressions_key_name,
			wantErr:  true,
		},
		{
			name:     "BindingPolicy with invalid matchExpressions operator",
			manifest: bp_bad_matchExpressions_op,
			wantErr:  true,
		},
		{
			name:     "BindingPolicy with invalid matchExpressions value",
			manifest: bp_bad_matchExpressions_value,
			wantErr:  true,
		},
		{
			name:     "BindingPolicy with valid matchExpressions",
			manifest: bp_good_matchExpressions,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCustomResourceYaml(tc.manifest, schemaValidator)
			if tc.wantErr != (err != nil) {
				t.Errorf("wantErr: %v, error, %s", tc.wantErr, err)
			} else if err != nil {
				t.Logf("Got expected validation error, %s", err)
			}
		})
	}
}

func validateCustomResourceYaml(data string, vd validation.SchemaValidator) error {
	cr := &unstructured.Unstructured{}
	if err := yaml.Unmarshal([]byte(data), cr); err != nil {
		return err
	}
	if err := validation.ValidateCustomResource(nil, cr.Object, vd).ToAggregate(); err != nil {
		return fmt.Errorf("%v %v is invalid, err: %w", cr.GetKind(), cr.GetName(), err)
	}
	return nil
}

var bp_bad_matchLabels_value = `
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nvidia-gpu-placement
spec:
  clusterSelectors:
  - matchLabels:
      name: edgeplatform-test-wec1, edgeplatform-test-wec2
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/part-of: wec-gpu
  wantSingletonReportedState: true
`

var bp_good_matchLabels = `
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nvidia-gpu-placement
spec:
  clusterSelectors:
  - matchLabels: {"sub.do-main.test/0.9-5_Xy": "This-is_1.test"}
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/part-of: wec-gpu
  wantSingletonReportedState: true
`

var bp_bad_matchExpressions_key_name = `
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nvidia-gpu-placement
spec:
  clusterSelectors:
  - matchExpressions:
    - key: -model
      operator: In
      values:
      - H200
      - H100
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/part-of: wec-gpu
  wantSingletonReportedState: true
`

var bp_bad_matchExpressions_key_prefix = `
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nvidia-gpu-placement
spec:
  clusterSelectors:
  - matchExpressions:
    - key: -kubestellar.io/model
      operator: In
      values:
      - H200
      - H100
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/part-of: wec-gpu
  wantSingletonReportedState: true
`
var bp_bad_matchExpressions_op = `
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nvidia-gpu-placement
spec:
  clusterSelectors:
  - matchExpressions:
    - key: model
      operator: OutOf
      values:
      - H200
      - H100
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/part-of: wec-gpu
  wantSingletonReportedState: true
`
var bp_bad_matchExpressions_value = `
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nvidia-gpu-placement
spec:
  clusterSelectors:
  - matchExpressions:
    - key: model
      operator: In
      values:
      - H200
      - H100,A100
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/part-of: wec-gpu
  wantSingletonReportedState: true
`

var bp_good_matchExpressions = `
apiVersion: control.kubestellar.io/v1alpha1
kind: BindingPolicy
metadata:
  name: nvidia-gpu-placement
spec:
  clusterSelectors:
  - matchExpressions:
    - key: available.nvidia.models/EST_26-Apr.2024
      operator: In
      values:
      - H200
      - H100
  downsync:
  - objectSelectors:
    - matchLabels:
        app.kubernetes.io/part-of: wec-gpu
  wantSingletonReportedState: true
`
