package status

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func makeObj(labels map[string]string) metav1.Object {
	obj := &unstructured.Unstructured{}
	obj.SetLabels(labels)
	return obj
}

func TestObjNotInThisWDS(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		thisWDS  string
		expected bool
	}{
		{
			name:     "label missing, should not process",
			labels:   map[string]string{},
			thisWDS:  "wds1",
			expected: true,
		},
		{
			name:     "label matches this WDS",
			labels:   map[string]string{originWdsLabelKey: "wds1"},
			thisWDS:  "wds1",
			expected: false,
		},
		{
			name:     "label belongs to different WDS",
			labels:   map[string]string{originWdsLabelKey: "wds2"},
			thisWDS:  "wds1",
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := makeObj(tt.labels)
			result := objNotInThisWDS(obj, tt.thisWDS)
			if result != tt.expected {
				t.Errorf("got %v, want %v", result, tt.expected)
			}
		})
	}
}
