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

package kbuser

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fake "k8s.io/client-go/kubernetes/fake"
	cache "k8s.io/client-go/tools/cache"
)

func TestSpaceIDMapping(t *testing.T) {
	cm1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CMNamePrefix + "underspace1",
			Namespace: "kubestellar",
			Labels:    map[string]string{KubeBindLabelName: "kb-id1"},
		},
	}
	client := fake.NewSimpleClientset(cm1)
	ctx := context.Background()
	cancel := func() {}
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, deadline)
	}
	reln := NewKubeBindSpaceRelation(ctx, client)
	cache.WaitForCacheSync(ctx.Done(), reln.InformerSynced)
	if actualKB, expectedKB := reln.SpaceIDToKubeBind("underspace1"), "kb-id1"; actualKB != expectedKB {
		t.Errorf("reln.SpaceIDToKubeBind(\"underspace1\") returned %q, expected %q", actualKB, expectedKB)
	}
	if actualUnder, expectedUnder := reln.SpaceIDFromKubeBind("kb-id1"), "underspace1"; actualUnder != expectedUnder {
		t.Errorf("reln.SpaceIDFromKubeBind(\"kb-id1\") returned %q, expected %q", actualUnder, expectedUnder)
	}
	cancel()
}
