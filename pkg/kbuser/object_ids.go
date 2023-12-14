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

// AnalyzeClusterScopedObject examines a cluster-scoped object in a kube-bind service provider cluster
// and returns the object's name as known in the service consumer cluster and the kube-bind space id for
// the consumer, or an error if the object does not appear to be a provider's copy of a consumer's object.
func AnalyzeClusterScopedObject(obj metav1.Object) (name, kbSpaceID string, err error) {
	if annotations := obj.GetAnnotations(); annotations != nil {
		kbSpaceID = annotations["kube-bind.io/cluster-namespace"]
	}
	if kbSpaceID == "" {
		return "", "", errors.New("no 'kube-bind.io/cluster-namespace' annotation found")
	}
	if obj.GetNamespace() != "" {
		return "", "", errors.New("object is namespaced-scoped and not cluster-scoped")
	}
	if name, ok := cutPrefix(obj.GetName(), kbSpaceID+"-"); !ok {
		err = fmt.Errorf("name %q does not have prefix for comsumer %q", name, kbSpaceID)
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

// If in the future a need arises to analyze both cluster-scoped and namespace-scoped objects, switch to using
// the following AnalyzeObjectID function instead of the existing AnalyzeClusterScopedObject function.

// AnalyzeObjectID examines an object in a kube-bind service provider
// cluster and returns the object's ID as known in the service consumer
// cluster and the kube-bind ID for the consumer, or an error if
// the object does not appear to be a provider's copy of a consumer's object.
// func AnalyzeObjectID(obj metav1.Object) (namespace, name, kbSpaceID string, err error) {
// 	if annotations := obj.GetAnnotations(); annotations != nil {
// 		kbSpaceID = annotations["kube-bind.io/cluster-namespace"]
// 	}
// 	if kbSpaceID == "" {
// 		return "", "", "", errors.New("no 'kube-bind.io/cluster-namespace' annotation found")
// 	}
// 	namespace = obj.GetNamespace()
// 	name = obj.GetName()
// 	var ok bool
// 	if namespace == "" { // no namespace, cluster scoped object
// 		name, ok = cutPrefix(name, kbSpaceID+"-")
// 		if !ok {
// 			err = fmt.Errorf("name %q does not have prefix for comsumer %q", name, kbSpaceID)
// 		}
// 		return
// 	}
// 	// otherwise, it's a namespace scoped object
// 	if namespace, ok = cutPrefix(namespace, kbSpaceID+"-"); !ok {
// 		err = fmt.Errorf("namespace %q does not have prefix for comsumer %q", namespace, kbSpaceID)
// 	}

// 	return
// }
