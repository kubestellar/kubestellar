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
		nsDistributions:  NewMapRelation3[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name](),
		nsrDistributions: NewMapRelation3[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name](),
		nnsDistributions: NewMapRelation3[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name](),
	}
	noteModeWrite := MapChangeReceiverFuncs[edgeapi.SinglePlacement, MutableMap[metav1.GroupResource, ProjectionModeVal]]{
		OnCreate: func(destination edgeapi.SinglePlacement, _ MutableMap[metav1.GroupResource, ProjectionModeVal]) {
			wp.changedDestinations.Add(destination)
		},
		OnDelete: func(destination edgeapi.SinglePlacement, _ MutableMap[metav1.GroupResource, ProjectionModeVal]) {
			wp.changedDestinations.Remove(destination)
		},
	}
	wp.nsModes = NewFactoredMapMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal](factorProjectionModeKeyForSyncer, nil, noteModeWrite)
	wp.nnsModes = NewFactoredMapMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal](factorProjectionModeKeyForSyncer, nil, noteModeWrite)
	return wp
}

var _ WorkloadProjector = &workloadProjector{}

type workloadProjector struct {
	ctx   context.Context
	queue workqueue.RateLimitingInterface

	sync.Mutex

	changedDestinations *MapSet[edgeapi.SinglePlacement]
	nsDistributions     MapRelation3[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name]
	nsrDistributions    MapRelation3[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name]
	nsModes             FactoredMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal]
	nnsDistributions    MapRelation3[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name]
	nnsModes            FactoredMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal]
}

type GroupResourceInstance = Pair[metav1.GroupResource, string /*object name*/]

func (wp *workloadProjector) Transact(xn func(WorkloadProjectionSections)) {
	logger := klog.FromContext(wp.ctx)
	wp.Lock()
	defer wp.Unlock()
	var s1 SetChangeReceiver[Triple[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name]] = wp.nsDistributions
	var s2 SetChangeReceiver[Triple[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name]] = wp.nsrDistributions
	var s3 SetChangeReceiver[Triple[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name]] = wp.nnsDistributions
	changedDestinations := NewMapSet[edgeapi.SinglePlacement]()
	wp.changedDestinations = &changedDestinations
	s1 = SetChangeReceiverFork(false, s1, recordFirst[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name](changedDestinations))
	s2 = SetChangeReceiverFork(false, s2, recordFirst[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name](changedDestinations))
	s3 = SetChangeReceiverFork(false, s3, recordFirst[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name](changedDestinations))
	xn(WorkloadProjectionSections{
		TransformSetChangeReceiver(factorNamespaceDistributionTupleForSyncer, s1),
		TransformSetChangeReceiver(factorNamespacedResourceDistributionTupleForSyncer, s2),
		wp.nsModes,
		TransformSetChangeReceiver(factorNonNamespacedDistributionTupleForSyncer, s3),
		wp.nnsModes})
	logger.V(2).Info("Transaction response", "changedDestinations", changedDestinations)
	changedDestinations.Visit(func(destination edgeapi.SinglePlacement) error {
		nsds, have := wp.nsDistributions.GetIndex1to2().Get(destination)
		if have {
			nses := MapKeySet[NamespaceName, Set[logicalcluster.Name]](nsds.GetIndex1to2())
			logger.Info("Namespaces after transaction", "destination", destination, "namespaces", MapSetCopy[NamespaceName](nses))
		}
		nsrds, have := wp.nsrDistributions.GetIndex1to2().Get(destination)
		if have {
			nsrs := MapKeySet[metav1.GroupResource, Set[logicalcluster.Name]](nsrds.GetIndex1to2())
			logger.Info("NamespacedResources after transation", "destination", destination, "resources", MapSetCopy[metav1.GroupResource](nsrs))
		}
		nsms, have := wp.nsModes.GetIndex().Get(destination)
		if have {
			logger.Info("Namespaced modes after transation", "destination", destination, "modes", MapMapCopy[metav1.GroupResource, ProjectionModeVal](nil, nsms))

		}
		nnsds, have := wp.nnsDistributions.GetIndex1to2().Get(destination)
		if have {
			objs := MapKeySet[GroupResourceInstance, Set[logicalcluster.Name]](nnsds.GetIndex1to2())
			logger.Info("NamespacedResources after transation", "destination", destination, "objs", MapSetCopy[GroupResourceInstance](objs))
		}
		nnsms, have := wp.nnsModes.GetIndex().Get(destination)
		if have {
			logger.Info("NonNamespaced modes after transation", "destination", destination, "modes", MapMapCopy[metav1.GroupResource, ProjectionModeVal](nil, nnsms))
		}
		return nil
	})
	wp.changedDestinations = nil
}

func recordFirst[First, Second, Third comparable](record MutableSet[First]) SetChangeReceiver[Triple[First, Second, Third]] {
	return SetChangeReceiverFuncs[Triple[First, Second, Third]]{
		OnAdd: func(tup Triple[First, Second, Third]) bool {
			record.Add(tup.First)
			return true
		},
		OnRemove: func(tup Triple[First, Second, Third]) bool {
			record.Add(tup.First)
			return true
		}}
}

func factorNamespaceDistributionTupleForSyncer(ndt NamespaceDistributionTuple) Triple[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name] {
	return NewTriple(ndt.Third, ndt.Second, ndt.First)
}

func factorNamespacedResourceDistributionTupleForSyncer(nrdt NamespacedResourceDistributionTuple) Triple[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name] {
	return NewTriple(nrdt.Destination, nrdt.GroupResource, nrdt.SourceCluster)
}

func factorNonNamespacedDistributionTupleForSyncer(nndt NonNamespacedDistributionTuple) Triple[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name] {
	return NewTriple(nndt.First.Destination, GroupResourceInstance{nndt.First.GroupResource, nndt.Second.Name}, nndt.Second.Cluster)
}

var factorProjectionModeKeyForSyncer = NewFactorer(
	func(pmk ProjectionModeKey) Pair[edgeapi.SinglePlacement, metav1.GroupResource] {
		return NewPair(pmk.Destination, pmk.GroupResource)
	},
	func(tup Pair[edgeapi.SinglePlacement, metav1.GroupResource]) ProjectionModeKey {
		return ProjectionModeKey{Destination: tup.First, GroupResource: tup.Second}
	})
