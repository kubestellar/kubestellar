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
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

const (
	CRDKind                              = "CustomResourceDefinition"
	AnyVersion                           = "*"
	ServiceVersion                       = "v1"
	ServiceKind                          = "Service"
	BindingPolicyKind                    = "BindingPolicy"
	BindingPolicyResource                = "bindingpolicies"
	BindingKind                          = "Binding"
	BindingResource                      = "bindings"
	WorkStatusGroup                      = "control.kubestellar.io"
	WorkStatusVersion                    = "v1alpha1"
	WorkStatusResource                   = "workstatuses"
	AnnotationToPreserveValuesKey        = "annotations.kubestellar.io/preserve"
	PreserveNodePortValue                = "nodeport"
	UnableToRetrieveCompleteAPIListError = "unable to retrieve the complete list of server APIs"
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

func IsBindingPolicy(o interface{}) bool {
	return objectMatchesGVK(o, v1alpha1.GroupVersion.Group, v1alpha1.GroupVersion.Version, BindingPolicyKind)
}

func IsService(o interface{}) bool {
	return objectMatchesGVK(o, "", ServiceVersion, ServiceKind)
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

func RemoveRuntimeGeneratedFieldsFromService(obj interface{}) error {
	uObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("object could not be cast to unstructured.Unstructured %#v", obj)
	}
	// Fields to remove
	fieldsToDelete := []string{"clusterIP", "clusterIPs", "ipFamilies",
		"externalTrafficPolicy", "internalTrafficPolicy"}

	for _, field := range fieldsToDelete {
		unstructured.RemoveNestedField(uObj.Object, "spec", field)
	}

	// Set the nodePort to an empty string unelss the annotation "kubestellar.io/annotations/preserve=nodeport" is present
	if !(uObj.GetAnnotations() != nil && uObj.GetAnnotations()[AnnotationToPreserveValuesKey] == PreserveNodePortValue) {
		if ports, found, _ := unstructured.NestedSlice(uObj.Object, "spec", "ports"); found {
			for i, port := range ports {
				if portMap, ok := port.(map[string]interface{}); ok {
					portMap["nodePort"] = nil
					ports[i] = portMap
				}
			}
			unstructured.SetNestedSlice(uObj.Object, ports, "spec", "ports")
		}
	}
	return nil
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

func CheckWorkStatusIPresent(config *rest.Config) bool {
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
