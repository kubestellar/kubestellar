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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type GVKObjRef struct {
	GK schema.GroupKind
	OR klog.ObjectRef
}

func (ref GVKObjRef) String() string {
	return ref.GK.String() + "(" + ref.OR.String() + ")"
}

// RefToRuntimeObj creates a GVKObjRef to a runtime.Object.
func RefToRuntimeObj(obj runtime.Object) GVKObjRef {
	gvk := obj.GetObjectKind().GroupVersionKind()
	mObj := obj.(metav1.Object)
	ans := GVKObjRef{
		GK: gvk.GroupKind(),
		OR: klog.ObjectRef{
			Namespace: mObj.GetNamespace(),
			Name:      mObj.GetName(),
		},
	}
	return ans
}
