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

package util

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

// ParseResourcesString takes a comma separated string list of resources in the form of
// <resource>.<api group>/<version> or <resource>/<version> and returns a slice of schema.GroupVersionResource
func ParseResourcesString(resources string) ([]schema.GroupVersionResource, error) {
	if resources == "" {
		return nil, nil
	}

	var result []schema.GroupVersionResource

	// trim single and double quotes to make this safe from
	// user passing the resource option quoted in the container args
	resources = strings.Trim(resources, `'"`)

	parts := strings.Split(resources, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		subparts := strings.SplitN(part, "/", 2)
		if len(subparts) != 2 {
			return nil, fmt.Errorf("invalid resource format: %s", part)
		}
		resourceGroup := subparts[0]
		version := subparts[1]
		if version == "" {
			return nil, fmt.Errorf("no version found: %s", part)
		}

		subparts = strings.SplitN(resourceGroup, ".", 2)
		var resource, group string
		if len(subparts) == 2 {
			// The part has the form of <resource>.<api group>
			resource = subparts[0]
			group = subparts[1]
		} else if len(subparts) == 1 {
			// The part has the form of <resource>
			resource = subparts[0]
			group = "" // The api group is empty
		} else {
			return nil, fmt.Errorf("invalid resource/group format: %s", resourceGroup)
		}

		gvr := schema.GroupVersionResource{
			Group:    group,
			Version:  version,
			Resource: resource,
		}
		result = append(result, gvr)
	}

	return result, nil
}

// IsResourceAllowed checks if an infomer for the resource is allowed to start.
// an empty or nil allowedResources slice is equivalent to allow all,
func IsResourceAllowed(gvr schema.GroupVersionResource, allowedResources []schema.GroupVersionResource) bool {
	if allowedResources == nil || allowedResources != nil && len(allowedResources) == 0 {
		return true
	}

	allowedResources = appendRequiredResources(allowedResources)

	for _, allowedResource := range allowedResources {
		if gvr.Group == allowedResource.Group &&
			(allowedResource.Version == AnyVersion || gvr.Version == allowedResource.Version) &&
			gvr.Resource == allowedResource.Resource {
			return true
		}
	}

	return false
}

// append the minimal set of resources that are required to operate
func appendRequiredResources(allowedResources []schema.GroupVersionResource) []schema.GroupVersionResource {
	// if resources are provided, we need to ensure that at least CRD and Placement
	// resources are watched

	slice := []schema.GroupVersionResource{}
	slice = append(slice, allowedResources...)

	gvr := schema.GroupVersionResource{
		Group:    v1alpha1.GroupVersion.Group,
		Version:  v1alpha1.GroupVersion.Version,
		Resource: PlacementResource,
	}
	slice = append(slice, gvr)

	// disabled until https://github.com/kubestellar/kubestellar/issues/1705 is resolved
	// to avoid client-side throttling
	// gvr = schema.GroupVersionResource{
	// 	Group:    CRDGroup,
	// 	Version:  AnyVersion,
	// 	Resource: CRDResource,
	// }
	// slice = append(slice, gvr)

	return slice
}
