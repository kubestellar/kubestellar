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
	"strings"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

// ParseResourcesString takes a comma separated string list of resources in the form of
// <api-group1>, <api-group2> .. and returns a map[string]bool
func ParseResourceGroupsString(resourceGroups string) map[string]bool {
	if resourceGroups == "" {
		return nil
	}

	// trim single and double quotes to make this safe from
	// user passing the resource option quoted in the container args
	resourceGroups = strings.Trim(resourceGroups, `'"`)

	groupsMap := map[string]bool{}
	for _, g := range strings.Split(resourceGroups, ",") {
		groupsMap[strings.Trim(g, " ")] = true
	}

	return groupsMap
}

// IsResourceGroupAllowed checks if an infomer for the resource is allowed to start.
// an empty or nil allowedResources slice is equivalent to allow all,
func IsResourceGroupAllowed(resourceGroup string, allowedResourceGroups map[string]bool) bool {
	if allowedResourceGroups == nil || allowedResourceGroups != nil && len(allowedResourceGroups) == 0 {
		return true
	}
	addRequiredResourceGroups(allowedResourceGroups)

	return allowedResourceGroups[resourceGroup]
}

// append the minimal set of resources that are required to operate
func addRequiredResourceGroups(allowedResourceGroups map[string]bool) {
	// if resources are provided, we need to ensure that at least CRD and KS API
	// resources are watched

	allowedResourceGroups[v1alpha1.GroupVersion.Group] = true

	// disabled until https://github.com/kubestellar/kubestellar/issues/1705 is resolved
	// to avoid client-side throttling
	// allowedResourceGroups[CRDGroup] = true
}
