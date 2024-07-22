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

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	//"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

// ParseAPIGroupsString takes a comma separated string list of api groups in the form of
// <api-group1>, <api-group2> .. and returns a sets.Set[string]
func ParseAPIGroupsString(apiGroups string) sets.Set[string] {
	if apiGroups == "" {
		return nil
	}

	groupsSet := sets.New(strings.Split(apiGroups, ",")...)

	addRequiredResourceGroups(groupsSet)

	return groupsSet
}

// IsResourceGroupAllowed checks if a API group is explicitly allowed by user,
// an empty or nil allowedResources slice is equivalent to allow all.
func IsAPIGroupAllowed(apiGroup string, allowedAPIGroups sets.Set[string]) bool {
	if len(allowedAPIGroups) == 0 {
		return true
	}
	return allowedAPIGroups.Has(apiGroup)
}

// append the minimal set of resources that are required to operate
func addRequiredResourceGroups(allowedResourceGroups sets.Set[string]) {
	// if groups are provided, we need to ensure that at least CRD and KS API
	// groups are watched

	allowedResourceGroups.Insert(v1alpha1.GroupVersion.Group)
	allowedResourceGroups.Insert(apiextensions.GroupName)
}
