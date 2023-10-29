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

package placement

import (
	"context"
	"testing"
	"time"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	fakeedge "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/fake"
	edgeinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
)

func TestWhereResolver(t *testing.T) {
	ctx := context.Background()
	if deadline, ok := t.Deadline(); ok {
		var cancel func()
		ctx, cancel = context.WithDeadline(ctx, deadline)
		t.Cleanup(cancel)
	}
	is1N := logicalcluster.Name("is1clusterid")
	wds1N := logicalcluster.Name("wds1clusterid")
	// wantedLabels := map[string]string{"foo":"bar"}
	sp1 := SinglePlacement{Cluster: is1N.String(), LocationName: "l1", SyncTargetName: "s1", SyncTargetUID: "s1uid"}
	sps1 := &edgeapi.SinglePlacementSlice{
		TypeMeta: metav1.TypeMeta{Kind: "SinglePlacementSlice", APIVersion: edgeapi.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{logicalcluster.AnnotationKey: wds1N.String()},
			Name:        "ep1",
		},
		Destinations: []edgeapi.SinglePlacement{sp1},
	}
	ep1EN := ExternalName{wds1N, ObjectName(sps1.Name)}
	edgeViewClusterClientset := fakeedge.NewSimpleClientset(sps1)
	edgeInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(edgeViewClusterClientset, 0)
	spsClusterPreInformer := edgeInformerFactory.Edge().V2alpha1().SinglePlacementSlices()
	spsClusterPreInformer.Informer() // get the informer created so that it will start
	whereResolver := NewWhereResolver(ctx, spsClusterPreInformer, 3)
	edgeInformerFactory.Start(ctx.Done())
	rcvr := NewMapMap[ExternalName, ResolvedWhere](nil)
	runnable := whereResolver(rcvr)
	go runnable.Run(ctx)
	expectedWhere := ResolvedWhere{sps1}
	err := wait.PollWithContext(ctx, time.Second, 5*time.Second, func(context.Context) (bool, error) {
		gotWhere, found := rcvr.Get(ep1EN)
		t.Logf("gotWhat=%v, found=%v", gotWhere, found)
		return found && apiequality.Semantic.DeepEqual(expectedWhere, gotWhere), nil
	})
	if err != nil {
		t.Fatalf("Failed to get expected ResolvedWhat in time: %v", err)
	}
}
