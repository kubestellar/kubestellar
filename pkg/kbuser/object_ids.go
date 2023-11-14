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

package kbuser

import (
	"errors"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AnalyzeObjectID examines an object in a kube-bind service provider
// cluster and returns the object's ID as known in the service consumer
// cluster and the kube-bind ID for the consumer, or an error if
// the object does not appear to be a provider's copy of a consumer's object.
func AnalyzeObjectID(obj metav1.Object) (namespace, name, kbSpaceID string, err error) {
	namespace = obj.GetNamespace()
	name = obj.GetName()
	annotations := obj.GetAnnotations()
	if annotations != nil {
		kbSpaceID = annotations["kube-bind.io/cluster-namespace"]
	}
	if kbSpaceID == "" {
		return "", "", "", errors.New("no 'kube-bind.io/cluster-namespace' annotation found")
	}
	var ok bool
	if namespace == "" {
		name, ok = cutPrefix(name, kbSpaceID+"-")
		if !ok {
			err = fmt.Errorf("name %q does not have prefix for comsumer %q", name, kbSpaceID)
		}
		return
	}
	namespace, ok = cutPrefix(namespace, kbSpaceID+"-")
	if !ok {
		err = fmt.Errorf("namespace %q does not have prefix for comsumer %q", namespace, kbSpaceID)
	}
	return
}

func cutPrefix(s, sep string) (string, bool) {
	if strings.HasPrefix(s, sep) {
		return s[len(sep):], true
	}
	return s, false
}

// ComposeClusterScopedName translates the name of a cluster-scoped object
// from what appears in the service consumer cluster to what appears in the
// service provider cluster. A Namespace is an example of a cluster-scoped
// object.
// For namespaced objects, kube-bind translates the name of their namespace.
func ComposeClusterScopedName(kbSpaceID string, name string) string {
	return kbSpaceID + "-" + name
}
