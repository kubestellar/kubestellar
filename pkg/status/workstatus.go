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

package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/pkg/util"
)

type workStatus struct {
	wecName                string
	sourceObjectIdentifier util.ObjectIdentifier
	status                 map[string]interface{}
}

func convertToWorkStatusesList(objList []runtime.Object) ([]*workStatus, error) {
	var workStatuses []*workStatus
	for _, ws := range objList {
		wStatus, err := convertToWorkStatus(ws)
		if err != nil {
			return workStatuses, err
		}
		workStatuses = append(workStatuses, wStatus)
	}

	return workStatuses, nil
}

func convertToWorkStatus(obj runtime.Object) (*workStatus, error) {
	// wecName is the WorkStatus namespace
	wecName := obj.(metav1.Object).GetNamespace()

	status, err := util.GetWorkStatusStatus(obj)
	if err != nil {
		return nil, err
	}
	sourceRef, err := util.GetWorkStatusSourceRef(obj)
	if err != nil {
		return nil, err
	}

	objIdentifier := util.ObjectIdentifier{
		GVK:        schema.GroupVersionKind{Group: sourceRef.Group, Version: sourceRef.Version, Kind: sourceRef.Kind},
		Resource:   sourceRef.Resource,
		ObjectName: cache.NewObjectName(sourceRef.Namespace, sourceRef.Name),
	}

	return &workStatus{
		wecName:                wecName,
		sourceObjectIdentifier: objIdentifier,
		status:                 status}, nil
}
