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

package ocm

import (
	workv1 "open-cluster-management.io/api/work/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubestellar/kubestellar/pkg/transport"
)

const (
	wrappedObjectKind       = "ManifestWork"
	wrappedObjectAPIVersion = "work.open-cluster-management.io/v1"
)

func NewOCMTransport() transport.Transport {
	return &ocm{}
}

type ocm struct {
}

var createOnlyStrategy = workv1.UpdateStrategy{Type: workv1.UpdateStrategyTypeCreateOnly}

func (ocm *ocm) WrapObjects(wrapees []transport.Wrapee) runtime.Object {
	manifests := make([]workv1.Manifest, len(wrapees))
	var configs []workv1.ManifestConfigOption
	for i, wrapee := range wrapees {
		manifests[i].RawExtension = runtime.RawExtension{Object: wrapee.Object}
		if wrapee.CreateOnly {
			gvk := wrapee.Object.GroupVersionKind()
			configs = append(configs, workv1.ManifestConfigOption{
				ResourceIdentifier: workv1.ResourceIdentifier{
					Group:     gvk.Group,
					Resource:  wrapee.Resource,
					Namespace: wrapee.Object.GetNamespace(),
					Name:      wrapee.Object.GetName(),
				},
				UpdateStrategy: &createOnlyStrategy,
			})
		}
	}
	return &workv1.ManifestWork{
		TypeMeta: metav1.TypeMeta{
			Kind:       wrappedObjectKind,
			APIVersion: wrappedObjectAPIVersion,
		},
		Spec: workv1.ManifestWorkSpec{
			Workload: workv1.ManifestsTemplate{
				Manifests: manifests,
			},
			ManifestConfigs: configs,
		},
	}
}
