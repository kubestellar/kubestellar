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
	"fmt"

	workv1 "open-cluster-management.io/api/work/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/transport"
	"github.com/kubestellar/kubestellar/pkg/util"
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

// serverSideApplyStrategy is used for Pods. Pods are mostly immutable, but a few
// fields (container image, activeDeadlineSeconds, toleration additions) can be
// updated in place and are desired state that must be synced in both directions.
// Server-side apply gives per-field ownership: the hub declares image etc., so
// hub updates reach the edge and edge drift is reverted, while cluster-owned
// fields such as spec.nodeName (stripped in TransformObject) are left to the
// cluster. createOnly still takes precedence when set.
var serverSideApplyStrategy = workv1.UpdateStrategy{Type: workv1.UpdateStrategyTypeServerSideApply}

func (ocm *ocm) WrapObjects(wrapees []transport.Wrapee, kindToResource func(schema.GroupKind) string) runtime.Object {
	manifests := make([]workv1.Manifest, len(wrapees))
	var configs []workv1.ManifestConfigOption
	for i, wrapee := range wrapees {
		manifests[i].RawExtension = runtime.RawExtension{Object: wrapee.Object}
		if wrapee.CreateOnly {
			// createOnly always takes precedence.
			gvk := wrapee.Object.GroupVersionKind()
			rsc := kindToResource(gvk.GroupKind())
			configs = append(configs, workv1.ManifestConfigOption{
				ResourceIdentifier: workv1.ResourceIdentifier{
					Group:     gvk.Group,
					Resource:  rsc,
					Namespace: wrapee.Object.GetNamespace(),
					Name:      wrapee.Object.GetName(),
				},
				UpdateStrategy: &createOnlyStrategy,
			})
			continue
		}
		if isPod(wrapee.Object) {
			gvk := wrapee.Object.GroupVersionKind()
			rsc := kindToResource(gvk.GroupKind())
			configs = append(configs, workv1.ManifestConfigOption{
				ResourceIdentifier: workv1.ResourceIdentifier{
					Group:     gvk.Group,
					Resource:  rsc,
					Namespace: wrapee.Object.GetNamespace(),
					Name:      wrapee.Object.GetName(),
				},
				UpdateStrategy: &serverSideApplyStrategy,
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

func (ocm *ocm) UnwrapObjects(wrapped runtime.Object, kindToResource func(schema.GroupKind) (string, bool)) (transport.Gloss, error) {
	gloss := transport.Gloss{}
	switch typed := wrapped.(type) {
	case *workv1.ManifestWork:
		for idx, manifest := range typed.Spec.Workload.Manifests {
			if manifest.Object == nil {
				return nil, fmt.Errorf("manifests[%d] has nil Object", idx)
			}
			obj := manifest.Object.(util.MRObject)
			gvk := obj.GetObjectKind().GroupVersionKind()
			gloss.Insert(util.GKObjRef{GK: gvk.GroupKind(), OR: klog.KObj(obj)})
		}
	case *unstructured.Unstructured:
		objData := typed.UnstructuredContent()
		manifests, found, err := unstructured.NestedSlice(objData, "spec", "workload", "manifests")
		if err != nil {
			return nil, fmt.Errorf("failed to extract manifests from ManifestWork: found=%v, err=%w", found, err)
		}
		for idx, manifest := range manifests {
			if manifestM, ok := manifest.(map[string]any); ok {
				obj := &unstructured.Unstructured{Object: manifestM}
				gvk := obj.GetObjectKind().GroupVersionKind()
				gloss.Insert(util.GKObjRef{GK: gvk.GroupKind(), OR: klog.KObj(obj)})
			} else {
				return nil, fmt.Errorf("manifests[%d] is a %T but expected a map[string]any", idx, manifest)
			}
		}
	}
	return gloss, nil
}

func ManifestConfigOptionResourceIdentifier(mc workv1.ManifestConfigOption) workv1.ResourceIdentifier {
	return mc.ResourceIdentifier
}

func ManifestConfigOptionUpdateStrategy(mc workv1.ManifestConfigOption) *workv1.UpdateStrategy {
	return mc.UpdateStrategy
}

// isPod reports whether the object is a core/v1 Pod.
func isPod(object *unstructured.Unstructured) bool {
	gvk := object.GroupVersionKind()
	return gvk.Group == "" && gvk.Kind == "Pod"
}
