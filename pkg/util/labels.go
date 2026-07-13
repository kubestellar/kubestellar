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
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

const (
	// BindingPolicyLabelSingletonStatusKey is the key for the singleton status reporting requirement.
	BindingPolicyLabelSingletonStatusKey = "managed-by.kubestellar.io/singletonstatus"

	// BindingPolicyLabelMultiWECStatusKey is the key for the multi-WEC aggregated status reporting.
	BindingPolicyLabelMultiWECStatusKey = "managed-by.kubestellar.io/multiwecstatus"

	// WorkStatusSourceRefKey is the key for the source reference in the WorkStatus.
	WorkStatusSourceRefKey = "managed-by.kubestellar.io/sourceRef"
)

// GetBindingPolicyGVR returns the GroupVersionResource for BindingPolicy.
func GetBindingPolicyGVR() schema.GroupVersionResource {
	return v1alpha1.GroupVersion.WithResource(BindingPolicyResource)
}

// GetBindingGVR returns the GroupVersionResource for Binding.
func GetBindingGVR() schema.GroupVersionResource {
	return v1alpha1.GroupVersion.WithResource(BindingResource)
}

// Label represents a key-value pair used for labeling resources.
type Label struct {
	Key   string
	Value string
}

// SplitLabelKeyAndValue splits a "key=value" string into a Label struct.
// It uses "=" as the delimiter to separate the label key from its value.
// It returns an error if the string is not in the correct format.
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

// SelectorsMatchLabels checks if any of the provided label selectors match the given label set.
// It returns true if at least one selector matches, otherwise false.
func SelectorsMatchLabels(selectors []metav1.LabelSelector, labelsSet labels.Set) (bool, error) {
	for _, selectorApi := range selectors {
		selector, err := metav1.LabelSelectorAsSelector(&selectorApi)
		if err != nil {
			return false, err
		}
		if selector.Matches(labelsSet) {
			return true, nil
		}
	}
	return false, nil
}
