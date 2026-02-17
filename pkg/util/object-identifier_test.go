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
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
)

func TestObjectIdentifierString(t *testing.T) {
	testCases := []struct {
		name     string
		input    ObjectIdentifier
		expected string
	}{
		{
			name: "namespaced object with api group",
			input: ObjectIdentifier{
				GVK:        schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
				Resource:   "deployments",
				ObjectName: cache.ObjectName{Namespace: "default", Name: "nginx"},
			},
			expected: "apps/v1/deployments(default/nginx)",
		},
		{
			name: "cluster-scoped object",
			input: ObjectIdentifier{
				GVK:        schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
				Resource:   "clusterroles",
				ObjectName: cache.ObjectName{Name: "admin"},
			},
			expected: "rbac.authorization.k8s.io/v1/clusterroles(admin)",
		},
		{
			name: "core group resource with empty group",
			input: ObjectIdentifier{
				GVK:        schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
				Resource:   "pods",
				ObjectName: cache.ObjectName{Namespace: "kube-system", Name: "coredns-abc123"},
			},
			expected: "v1/pods(kube-system/coredns-abc123)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := tc.input.String()
			if actual != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}
