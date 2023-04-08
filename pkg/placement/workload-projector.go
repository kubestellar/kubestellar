/*
Copyright 2023 The KCP Authors.

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
	"sync"

	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func NewWorkloadProjector(
	ctx context.Context,
) *workloadProjector {
	wp := &workloadProjector{
		ctx:              ctx,
		queue:            workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		nsDistributions:  NewMapRelation3Index[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name](),
		nsrDistributions: NewMapRelation3Index[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name](),
		nsModes:          NewMapMap[ProjectionModeKey, ProjectionModeVal](nil),
		nnsDistributions: NewMapRelation3Index[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name](),
		nnsModes:         NewMapMap[ProjectionModeKey, ProjectionModeVal](nil),
	}
	return wp
}

var _ WorkloadProjector = &workloadProjector{}

type workloadProjector struct {
	ctx   context.Context
	queue workqueue.RateLimitingInterface

	sync.Mutex

	nsDistributions  *MapRelation2[edgeapi.SinglePlacement, Pair[NamespaceName, logicalcluster.Name]]
	nsrDistributions *MapRelation2[edgeapi.SinglePlacement, Pair[metav1.GroupResource, logicalcluster.Name]]
	nsModes          MutableMap[ProjectionModeKey, ProjectionModeVal]
	nnsDistributions *MapRelation2[edgeapi.SinglePlacement, Pair[GroupResourceInstance, logicalcluster.Name]]
	nnsModes         MutableMap[ProjectionModeKey, ProjectionModeVal]
}

type GroupResourceInstance = Pair[metav1.GroupResource, string /*object name*/]

func (wp *workloadProjector) Transact(xn func(WorkloadProjectionSections)) {
	logger := klog.FromContext(wp.ctx)
	var s1 SetChangeReceiver[Pair[edgeapi.SinglePlacement, Pair[NamespaceName, logicalcluster.Name]]] = wp.nsDistributions
	var s2 SetChangeReceiver[Pair[edgeapi.SinglePlacement, Pair[metav1.GroupResource, logicalcluster.Name]]] = wp.nsrDistributions
	var s3 SetChangeReceiver[Pair[edgeapi.SinglePlacement, Pair[GroupResourceInstance, logicalcluster.Name]]] = wp.nnsDistributions
	changedDestinations := NewMapSet[edgeapi.SinglePlacement]()
	s1 = SetChangeReceiverFork(false, s1, recordFirst[edgeapi.SinglePlacement, Pair[NamespaceName, logicalcluster.Name]](changedDestinations))
	s2 = SetChangeReceiverFork(false, s2, recordFirst[edgeapi.SinglePlacement, Pair[metav1.GroupResource, logicalcluster.Name]](changedDestinations))
	s3 = SetChangeReceiverFork(false, s3, recordFirst[edgeapi.SinglePlacement, Pair[GroupResourceInstance, logicalcluster.Name]](changedDestinations))
	xn(WorkloadProjectionSections{
		TransformSetChangeReceiver(factorNamespaceDistributionTupleForSyncer, s1),
		TransformSetChangeReceiver(factorNamespacedResourceDistributionTupleForSyncer, s2),
		wp.nsModes,
		TransformSetChangeReceiver(factorNonNamespacedDistributionTupleForSyncer, s3),
		wp.nnsModes})
	logger.V(2).Info("Transaction response", "changedDestinations", changedDestinations)
	changedDestinations.Visit(func(destination edgeapi.SinglePlacement) error {
		wp.nsDistributions.GetIndex1to2().Get(destination)
		// TODO: finish implementing
		return nil
	})
}

func recordFirst[First, Second comparable](record MutableSet[First]) SetChangeReceiver[Pair[First, Second]] {
	return SetChangeReceiverFuncs[Pair[First, Second]]{
		OnAdd: func(tup Pair[First, Second]) bool {
			record.Add(tup.First)
			return true
		},
		OnRemove: func(tup Pair[First, Second]) bool {
			record.Add(tup.First)
			return true
		}}
}

func factorNamespaceDistributionTupleForSyncer(ndt NamespaceDistributionTuple) Pair[edgeapi.SinglePlacement, Pair[NamespaceName, logicalcluster.Name]] {
	return Pair[edgeapi.SinglePlacement, Pair[NamespaceName, logicalcluster.Name]]{
		First:  ndt.Third,
		Second: Pair[NamespaceName, logicalcluster.Name]{ndt.Second, ndt.First}}
}

func factorNamespacedResourceDistributionTupleForSyncer(nrdt NamespacedResourceDistributionTuple) Pair[edgeapi.SinglePlacement, Pair[metav1.GroupResource, logicalcluster.Name]] {
	return Pair[edgeapi.SinglePlacement, Pair[metav1.GroupResource, logicalcluster.Name]]{
		First:  nrdt.Destination,
		Second: Pair[metav1.GroupResource, logicalcluster.Name]{nrdt.GroupResource, nrdt.SourceCluster}}
}

func factorNonNamespacedDistributionTupleForSyncer(nndt NonNamespacedDistributionTuple) Pair[edgeapi.SinglePlacement, Pair[GroupResourceInstance, logicalcluster.Name]] {
	return Pair[edgeapi.SinglePlacement, Pair[GroupResourceInstance, logicalcluster.Name]]{
		First: nndt.First.Destination,
		Second: Pair[GroupResourceInstance, logicalcluster.Name]{
			GroupResourceInstance{nndt.First.GroupResource, nndt.Second.Name},
			nndt.Second.Cluster}}
}
