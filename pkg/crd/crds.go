/*
Copyright 2023 The KubeStellar Authors.

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
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"time"

	"github.com/go-logr/logr"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"

	ksmetrics "github.com/kubestellar/kubestellar/pkg/metrics"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// CRDs to apply
var crdNames = sets.New(
	"bindings.control.kubestellar.io",
	"bindingpolicies.control.kubestellar.io",
	"customtransforms.control.kubestellar.io",
	"statuscollectors.control.kubestellar.io",
	"combinedstatuses.control.kubestellar.io",
)

//go:embed files/*
var embeddedFiles embed.FS

const (
	FieldManager = "kubestellar"
)

// ApplyCRDs ensures that the KubeStellar control CRDs exist and are "established".
func ApplyCRDs(ctx context.Context, controllerName string,
	clientsetExt ksmetrics.ClientModNamespace[*apiextensionsv1.CustomResourceDefinition, *apiextensionsv1.CustomResourceDefinitionList],
	logger logr.Logger) error {
	ctxLimited, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	crdUnstructureds, err := readCRDs()
	if err != nil {
		return err
	}

	crdUnstructureds = filterCRDsByNames(crdUnstructureds, crdNames)

	for _, crdU := range crdUnstructureds {
		logger.V(1).Info("Applying CRD", "name", crdU.GetName())

		desiredCRD := &apiextensionsv1.CustomResourceDefinition{}
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(crdU.UnstructuredContent(), desiredCRD)
		if err != nil {
			return fmt.Errorf("unable to convert from Unstructured to CRD, name=%s: %w", crdU.GetName(), err)
		}
		_, err = clientsetExt.Create(ctx, desiredCRD, metav1.CreateOptions{FieldValidation: metav1.FieldValidationStrict, FieldManager: controllerName})
		if err == nil {
			logger.Info("Created CRD", "name", desiredCRD.Name)
		} else if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("unable to create CRD named %s: %w", crdU.GetName(), err)
		} else {
			existingCRD, err := clientsetExt.Get(ctx, desiredCRD.Name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to fetch an existing CRD, name=%s: %w", desiredCRD.Name, err)
			}
			if apiequality.Semantic.DeepEqual(existingCRD.Spec, desiredCRD.Spec) {
				logger.Info("Existing CRD is acceptable, name=%s", desiredCRD.Name)
				continue
			}
			desiredCRD.ResourceVersion = existingCRD.ResourceVersion
			_, err = clientsetExt.Update(ctx, desiredCRD, metav1.UpdateOptions{FieldManager: controllerName})
			if err != nil {
				return fmt.Errorf("unable to update existing CRD named %s: %w", crdU.GetName(), err)
			}
			logger.Info("Updated CRD", "name", desiredCRD.Name)
		}

		// wait until name accepted
		err = waitForCRDAccepted(ctxLimited, clientsetExt, desiredCRD.Name)
		if err != nil {
			return err
		}
		logger.Info("CRD is established", "name", crdU.GetName())
	}
	return nil
}

func filterCRDsByNames(crds []*unstructured.Unstructured, names sets.Set[string]) []*unstructured.Unstructured {
	out := make([]*unstructured.Unstructured, 0)

	for _, o := range crds {
		if names.Has(o.GetName()) {
			out = append(out, o)
		}
	}

	return out
}

func readCRDs() ([]*unstructured.Unstructured, error) {
	crds := make([]*unstructured.Unstructured, 0)

	dirEntries, _ := fs.ReadDir(embeddedFiles, "files")
	for _, entry := range dirEntries {
		file, err := embeddedFiles.Open("files/" + entry.Name())
		if err != nil {
			return nil, err
		}

		content, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		objs, err := DecodeYAML(content)
		if err != nil {
			return nil, err
		}

		for _, obj := range objs {
			if util.IsCRD(obj) {
				crds = append(crds, obj)
			}
		}
	}
	return crds, nil
}

// DecodeYAML decodes the content of a yaml file into a slice of unstructured objects.
func DecodeYAML(yamlBytes []byte) ([]*unstructured.Unstructured, error) {
	var objects []*unstructured.Unstructured
	yamlReader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(yamlBytes)))
	for {
		yamlDoc, err := yamlReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// Decode the YAML document into an unstructured object
		obj := &unstructured.Unstructured{}
		err = yaml.Unmarshal(yamlDoc, obj)
		if err != nil {
			return nil, err
		}

		objects = append(objects, obj)
	}
	return objects, nil
}

func waitForCRDAccepted(ctx context.Context,
	clientset ksmetrics.ClientModNamespace[*apiextensionsv1.CustomResourceDefinition, *apiextensionsv1.CustomResourceDefinitionList],
	crdName string) error {
	return wait.PollUntilContextCancel(ctx, 1*time.Second, true, func(ctx context.Context) (bool, error) {
		crd, err := clientset.Get(ctx, crdName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, condition := range crd.Status.Conditions {
			if condition.Type == apiextensionsv1.Established && condition.Status == apiextensionsv1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	})
}
