/*
Copyright 2026 The KubeStellar Authors.

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

package binding

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func TestShouldQueueBindingPolicyOnUpdate(t *testing.T) {
	now := metav1.NewTime(time.Now())
	testCases := []struct {
		name string
		old  *v1alpha1.BindingPolicy
		new  *v1alpha1.BindingPolicy
		want bool
	}{
		{
			name: "spec change",
			old:  &v1alpha1.BindingPolicy{ObjectMeta: metav1.ObjectMeta{Generation: 1}},
			new:  &v1alpha1.BindingPolicy{ObjectMeta: metav1.ObjectMeta{Generation: 2}},
			want: true,
		},
		{
			name: "status only",
			old:  &v1alpha1.BindingPolicy{ObjectMeta: metav1.ObjectMeta{Generation: 2}},
			new:  &v1alpha1.BindingPolicy{ObjectMeta: metav1.ObjectMeta{Generation: 2, ResourceVersion: "9"}},
			want: false,
		},
		{
			name: "deletion timestamp set",
			old:  &v1alpha1.BindingPolicy{ObjectMeta: metav1.ObjectMeta{Generation: 2}},
			new:  &v1alpha1.BindingPolicy{ObjectMeta: metav1.ObjectMeta{Generation: 2, DeletionTimestamp: &now}},
			want: true,
		},
		{
			name: "already deleting",
			old:  &v1alpha1.BindingPolicy{ObjectMeta: metav1.ObjectMeta{Generation: 2, DeletionTimestamp: &now}},
			new:  &v1alpha1.BindingPolicy{ObjectMeta: metav1.ObjectMeta{Generation: 2, DeletionTimestamp: &now, ResourceVersion: "10"}},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldQueueBindingPolicyOnUpdate(tc.old, tc.new)
			if got != tc.want {
				t.Errorf("shouldQueueBindingPolicyOnUpdate() = %v, want %v", got, tc.want)
			}
		})
	}
}
