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
	"strings"

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

func (ocm *ocm) WrapObjects(wrapees []transport.Wrapee, kindToResource func(schema.GroupKind) string) runtime.Object {
	manifests := make([]workv1.Manifest, len(wrapees))
	var configs []workv1.ManifestConfigOption
	for i, wrapee := range wrapees {
		manifests[i].RawExtension = runtime.RawExtension{Object: wrapee.Object}
		gvk := wrapee.Object.GroupVersionKind()
		rsc := kindToResource(gvk.GroupKind())

		var ignorePaths []string
		var conditionRules []workv1.ConditionRule
		if gvk.Group == "batch" && gvk.Kind == "Job" {
			ignorePaths = append(ignorePaths,
				".spec.selector",
				".spec.template.metadata.labels.controller-uid",
				".spec.template.metadata.labels[\"batch.kubernetes.io/controller-uid\"]",
			)
			conditionRules = append(conditionRules, workv1.ConditionRule{
				Condition: "Complete",
				Type:      workv1.WellKnownConditionsType,
			})
		}

		if annotationVal := wrapee.Object.GetAnnotations()["kubestellar.io/ignore-fields"]; annotationVal != "" {
			parts := strings.Split(annotationVal, ",")
			for _, p := range parts {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					ignorePaths = append(ignorePaths, trimmed)
				}
			}
		}

		if wrapee.CreateOnly {
			configs = append(configs, workv1.ManifestConfigOption{
				ResourceIdentifier: workv1.ResourceIdentifier{
					Group:     gvk.Group,
					Resource:  rsc,
					Namespace: wrapee.Object.GetNamespace(),
					Name:      wrapee.Object.GetName(),
				},
				UpdateStrategy: &createOnlyStrategy,
			})
		} else if len(ignorePaths) > 0 || len(conditionRules) > 0 {
			var updateStrategy *workv1.UpdateStrategy
			if len(ignorePaths) > 0 {
				updateStrategy = &workv1.UpdateStrategy{
					Type: workv1.UpdateStrategyTypeServerSideApply,
					ServerSideApply: &workv1.ServerSideApplyConfig{
						FieldManager: workv1.DefaultFieldManager,
						IgnoreFields: []workv1.IgnoreField{
							{
								Condition: workv1.IgnoreFieldsConditionOnSpokePresent,
								JSONPaths: ignorePaths,
							},
						},
					},
				}
			}

			configs = append(configs, workv1.ManifestConfigOption{
				ResourceIdentifier: workv1.ResourceIdentifier{
					Group:     gvk.Group,
					Resource:  rsc,
					Namespace: wrapee.Object.GetNamespace(),
					Name:      wrapee.Object.GetName(),
				},
				UpdateStrategy: updateStrategy,
				ConditionRules: conditionRules,
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
