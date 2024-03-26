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

package util

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	CRDKind                       = "CustomResourceDefinition"
	AnyVersion                    = "*"
	ServiceVersion                = "v1"
	ServiceKind                   = "Service"
	BindingPolicyKind             = "BindingPolicy"
	BindingPolicyResource         = "bindingpolicies"
	BindingKind                   = "Binding"
	BindingResource               = "bindings"
	WorkStatusGroup               = "control.kubestellar.io"
	WorkStatusVersion             = "v1alpha1"
	WorkStatusResource            = "workstatuses"
	AnnotationToPreserveValuesKey = "annotations.kubestellar.io/preserve"
	PreserveNodePortValue         = "nodeport"
)

// this type is used in status-addon, which we cannot import due to conflicting versions
// of packages used in Open Cluster Management AddOn framework (otel - open telemetry)
// TODO - consider separating APIs in a different git repo to become independent of
// libs deps in different projects
type SourceRef struct {
	Group     string `json:"group,omitempty"`
	Version   string `json:"version,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Kind      string `json:"kind,omitempty"`
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

func IsCRD(o interface{}) bool { // CRDs might have different versions. therefore, using "any" in CRD version
	return objectMatchesGVK(o, apiextensions.GroupName, AnyVersion, CRDKind)
}

func objectMatchesGVK(o interface{}, group, version, kind string) bool {
	gvk, err := getObjectGVK(o)
	if err != nil {
		return false
	}

	return gvkMatches(gvk, group, version, kind)
}

func getObjectGVK(o interface{}) (schema.GroupVersionKind, error) {
	switch obj := o.(type) {
	case runtime.Object:
		return obj.GetObjectKind().GroupVersionKind(), nil
	}

	return schema.GroupVersionKind{}, fmt.Errorf("object is of wrong type: %#v", o)
}

func GetWorkStatusSourceRef(workStatus runtime.Object) (*SourceRef, error) {
	obj, ok := workStatus.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("object cannot be cast to *unstructured.Unstructured")
	}

	spec := obj.Object["spec"].(map[string]interface{})
	if spec == nil {
		return nil, fmt.Errorf("could not find spec in object")
	}
	sourceRef := spec["sourceRef"].(map[string]interface{})
	if sourceRef == nil {
		return nil, fmt.Errorf("could not find sourceRef in object spec")
	}

	group, ok := sourceRef["group"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find group in sourceRef or it's not a string")
	}

	version, ok := sourceRef["version"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find version in sourceRef or it's not a string")
	}

	resource, ok := sourceRef["resource"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find resource in sourceRef or it's not a string")
	}

	kind, ok := sourceRef["kind"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find kind in sourceRef or it's not a string")
	}

	name, ok := sourceRef["name"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find name in sourceRef or it's not a string")
	}

	namespace, ok := sourceRef["namespace"].(string)
	if !ok {
		return nil, fmt.Errorf("could not find namespace in sourceRef or it's not a string")
	}

	return &SourceRef{
		Group:     group,
		Version:   version,
		Resource:  resource,
		Kind:      kind,
		Name:      name,
		Namespace: namespace,
	}, nil
}

func GetWorkStatusStatus(workStatus runtime.Object) (map[string]interface{}, error) {
	obj, ok := workStatus.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("object cannot be cast to *unstructured.Unstructured")
	}

	statusObj := obj.Object["status"]
	if statusObj == nil {
		return nil, fmt.Errorf("could not find status in object %s", obj.GetName())
	}

	status, ok := statusObj.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("object cannot be cast to map[string]interface{}")
	}

	return status, nil
}

func CheckWorkStatusPresence(config *rest.Config) bool {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return false
	}

	gvr := schema.GroupVersionResource{
		Group:    "control.kubestellar.io",
		Version:  "v1alpha1",
		Resource: "workstatuses",
	}

	return CheckAPIisPresent(discoveryClient, gvr)
}

func CheckAPIisPresent(dc *discovery.DiscoveryClient, gvr schema.GroupVersionResource) bool {
	resourceList, err := dc.ServerResourcesForGroupVersion(gvr.GroupVersion().String())
	if err != nil {
		return false
	}

	for _, resource := range resourceList.APIResources {
		if resource.Name == gvr.Resource {
			return true
		}
	}

	return false
}

// CreateStatusPatch creates a status patch for unstructured object.
func CreateStatusPatch(unstrObj *unstructured.Unstructured, status map[string]interface{}) *unstructured.Unstructured {
	patchedObj := &unstructured.Unstructured{}
	patchedObj.SetAPIVersion(unstrObj.GetAPIVersion())
	patchedObj.SetKind(unstrObj.GetKind())
	patchedObj.SetName(unstrObj.GetName())
	patchedObj.SetNamespace(unstrObj.GetNamespace())
	patchedObj.Object["status"] = status
	return patchedObj
}

// PatchStatus updates the object status with Patch.
func PatchStatus(ctx context.Context, unstrObj *unstructured.Unstructured, status map[string]interface{},
	namespace string, gvr schema.GroupVersionResource, dynamicClient dynamic.Interface) error {

	patchBytes, err := json.Marshal(CreateStatusPatch(unstrObj, status))
	if err != nil {
		return err
	}

	if namespace != "" {
		_, err = dynamicClient.Resource(gvr).Namespace(namespace).Patch(ctx, unstrObj.GetName(), types.MergePatchType, patchBytes, metav1.PatchOptions{})
	} else {
		_, err = dynamicClient.Resource(gvr).Patch(ctx, unstrObj.GetName(), types.MergePatchType, patchBytes, metav1.PatchOptions{})
	}

	return err
}
