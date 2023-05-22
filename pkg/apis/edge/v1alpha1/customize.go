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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ParameterExpansionAnnotationKey, when paired with the value "true" in an annotation of
// an object subject to edge management, indicates that parameter expansion should be
// bundled with propagation from center to edge.
//
// Parameter expansion applies to every leaf string in the object, and involves
// two substring replacements.
// One is replacing "%%" with "%".
// The other replaces every substring of the form "%(parameter_name)" with the destination's
// value for the named parameter.  A parameter_name can be any label or annotation key.
//
// A destination is a [kcp Location](https://github.com/kcp-dev/kcp/blob/v0.8.2/pkg/apis/scheduling/v1alpha1/types_location.go#L50)
// and its labels and annotations provide parameter values (with labels taking priority over annotations).
//
// Note that this sort of customization has limited applicability.  It can only be used where
// the un-expanded string passes the validation conditions of the relevant object type.
// For more broadly applicable customization, see Customizer objects.

const ParameterExpansionAnnotationKey string = "edge.kcp.io/expand-parameters"

// CustomizerAnnotationKey is the key of an annotation that identifies the Customizer
// that applies to the annotated object as it propagates from center to edge.
// When the annotated object is namespaced, the Customizer object to apply is in
// the same namespace and the value of this annotation is the name of the Customizer object;
// otherwise the value of this annotation uses the form "namespace/name" to reference
// the desired Customizer.
const CustomizerAnnotationKey string = "edge.kcp.io/customizer"

// +crd
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Customizer defines modifications to make to the relevant objects as they propagate
// from center to edge.
//
// The relevant objects are those with an annotation whose key is
// "edge.kcp.io/customizer" and whose value refers to this object as explained above.
//
// If this object is marked as being subject to parameter expansion then
// the parameter-expanded version of this object is what gets applied to a relevant
// object as it propagates to a destination.
type Customizer struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// `replacements` defines modifications to do to an object.
	// +listType=map
	// +listMapKey=path
	// +patchMergeKey=path
	// +patchStrategy=merge
	// +optional
	Replacements []Replacement `patchStrategy:"merge" patchMergeKey:"path" json:"replacements,omitempty"`
}

// Replacement represents one modification to an object.
// Such a replacement is conceptually done on the JSON representation of that object.
type Replacement struct {
	// `path` is a JSON Path identifying the part of the object to replace/inject.
	Path string `json:"path"`

	// `value` supplies the new value to put where the path points, in JSON.
	Value string `json:"value"`
}

// CustomizerList is the API type for a list of Customizer
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CustomizerList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Customizer `json:"items"`
}
