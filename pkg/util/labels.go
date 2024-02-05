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
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

const (
	// placement key has the form managed-by.kubestellar.io/<wds-name>.<placement-name>
	// this is because key must be unique per wds and we need to identify the wds for
	// multi-wds environments.
	PlacementLabelKeyBase         = "managed-by.kubestellar.io"
	PlacementLabelValueEnabled    = "true"
	PlacementLabelSingletonStatus = "managed-by.kubestellar.io/singletonstatus"
)

func GetPlacementListerKey() string {
	return KeyForGroupVersionKind(v1alpha1.GroupVersion.Group,
		v1alpha1.GroupVersion.Version, PlacementKind)
}

func GetPlacementDecisionListerKey() string {
	return KeyForGroupVersionKind(v1alpha1.GroupVersion.Group,
		v1alpha1.GroupVersion.Version, PlacementDecisionKind)
}

func SetManagedByPlacementLabels(obj metav1.Object, wdsName string, managedByPlacements []string, singletonStatus bool) {
	objLabels := obj.GetLabels()
	if objLabels == nil {
		objLabels = make(labels.Set)
	}
	for _, placement := range managedByPlacements {
		objLabels = mergeManagedByPlacementLabel(objLabels, wdsName, placement)
	}
	// label manifest requiring simgleton status so that status controller can evaluate
	if singletonStatus {
		objLabels[PlacementLabelSingletonStatus] = PlacementLabelValueEnabled
	}
	obj.SetLabels(objLabels)
}

func mergeManagedByPlacementLabel(l labels.Set, wdsName, placementName string) labels.Set {
	plLabel := make(labels.Set)
	key := GenerateManagedByPlacementLabelKey(wdsName, placementName)
	plLabel[key] = PlacementLabelValueEnabled
	return labels.Merge(l, plLabel)
}

func GenerateManagedByPlacementLabelKey(wdsName, placementName string) string {
	return fmt.Sprintf("%s/%s.%s", PlacementLabelKeyBase, wdsName, placementName)
}

func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

type Label struct {
	Key   string
	Value string
}

func SplitLabelKeyAndValue(keyvalue string) (Label, error) {
	label := Label{}
	parts := strings.Split(keyvalue, "=")
	if len(parts) != 2 {
		return label, fmt.Errorf("invalid key=value label: %s", keyvalue)
	}
	label.Key = parts[0]
	label.Value = parts[1]
	return label, nil
}
