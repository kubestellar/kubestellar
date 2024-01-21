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
	"bytes"
	"context"
	"embed"
	"io"
	"io/fs"
	"time"

	"github.com/go-logr/logr"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	kfutil "github.com/kubestellar/kubeflex/pkg/util"

	"github.com/kubestellar/kubestellar/pkg/util"
)

// CRDs to apply
var crdNames = map[string]bool{"placements.edge.kubestellar.io": true}

//go:embed files/*
var embeddedFiles embed.FS

const FieldManager = "kubestellar"

func ApplyCRDs(dynamicClient *dynamic.DynamicClient, clientset *kubernetes.Clientset, clientsetExt *apiextensionsclientset.Clientset, logger logr.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	crds, err := readCRDs()
	if err != nil {
		return err
	}

	crds = filterCRDsByNames(crds, crdNames)

	for _, crd := range crds {
		gvk := kfutil.GetGroupVersionKindFromObject(crd)
		gvr, err := kfutil.GroupVersionKindToResource(clientset, gvk)
		if err != nil {
			return err
		}
		logger.Info("applying crd", "name", crd.GetName())
		_, err = dynamicClient.Resource(*gvr).Apply(context.TODO(), crd.GetName(), crd, metav1.ApplyOptions{FieldManager: FieldManager})
		if err != nil {
			return err
		}

		// wait until name accepted
		err = waitForCRDAccepted(ctx, clientsetExt, crd.GetName())
		if err != nil {
			return err
		}
		logger.Info("crd name accepted", "name", crd.GetName())
	}
	return nil
}

func filterCRDsByNames(crds []*unstructured.Unstructured, names map[string]bool) []*unstructured.Unstructured {
	out := make([]*unstructured.Unstructured, 0)
	for _, o := range crds {
		if _, ok := names[o.GetName()]; ok {
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
		obj, err := DecodeYAML(content)
		if err != nil {
			return nil, err
		}

		if util.IsCRD(obj) {
			crds = append(crds, obj)
		}
	}
	return crds, nil
}

// Read the YAML into an unstructured object
func DecodeYAML(yamlBytes []byte) (*unstructured.Unstructured, error) {
	obj := &unstructured.Unstructured{}
	dec := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(yamlBytes), 4096)
	err := dec.Decode(obj)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func waitForCRDAccepted(ctx context.Context, clientset *apiextensionsclientset.Clientset, crdName string) error {
	return wait.PollUntilContextCancel(ctx, 1*time.Second, true, func(ctx context.Context) (bool, error) {
		crd, err := clientset.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, crdName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, condition := range crd.Status.Conditions {
			if condition.Type == apiextensionsv1.NamesAccepted && condition.Status == apiextensionsv1.ConditionTrue {
				return true, nil
			}
		}

		return false, nil
	})
}
