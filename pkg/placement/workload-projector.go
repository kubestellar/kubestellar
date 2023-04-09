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

	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"

	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgeclusterclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
)

func NewWorkloadProjector(
	ctx context.Context,
	configConcurrency int,
	mbwsInformer k8scache.SharedIndexInformer,
	mbwsLister tenancyv1a1listers.WorkspaceLister,
	edgeClusterClientset edgeclusterclientset.ClusterInterface,
) *workloadProjector {
	wp := &workloadProjector{
		ctx:                  ctx,
		configConcurrency:    configConcurrency,
		queue:                workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		mbwsLister:           mbwsLister,
		edgeClusterClientset: edgeClusterClientset,

		mbwsNameToCluster: WrapMapWithMutex[string, logicalcluster.Name](NewMapMap[string, logicalcluster.Name](nil)),

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
	mbwsInformer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			ws := obj.(*tenancyv1a1.Workspace)
			wp.mbwsNameToCluster.Put(ws.Name, logicalcluster.Name(ws.Spec.Cluster))
		},
		UpdateFunc: func(oldObj, newObj any) {
			ws := newObj.(*tenancyv1a1.Workspace)
			wp.mbwsNameToCluster.Put(ws.Name, logicalcluster.Name(ws.Spec.Cluster))
		},
		DeleteFunc: func(obj any) {
			innerObj := obj
			switch typed := obj.(type) {
			case k8scache.DeletedFinalStateUnknown:
				innerObj = typed.Obj
			default:
			}
			ws := innerObj.(*tenancyv1a1.Workspace)
			wp.mbwsNameToCluster.Delete(ws.Name)
		},
	})
	return wp
}

var _ WorkloadProjector = &workloadProjector{}
var _ Runnable = &workloadProjector{}

type workloadProjector struct {
	ctx                  context.Context
	configConcurrency    int
	queue                workqueue.RateLimitingInterface
	mbwsLister           tenancyv1a1listers.WorkspaceLister
	edgeClusterClientset edgeclusterclientset.ClusterInterface
	mbwsNameToCluster    MutableMap[string /*mailbox workspace name*/, logicalcluster.Name]

	sync.Mutex

	changedDestinations *MapSet[edgeapi.SinglePlacement]
	nsDistributions     MapRelation3[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name]
	nsrDistributions    MapRelation3[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name]
	nsModes             FactoredMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal]
	nnsDistributions    MapRelation3[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name]
	nnsModes            FactoredMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal]
}

type GroupResourceInstance = Pair[metav1.GroupResource, string /*object name*/]

func (wp *workloadProjector) Run(ctx context.Context) {
	doneCh := ctx.Done()
	for worker := 0; worker < wp.configConcurrency; worker++ {
		go wp.configSyncLoop(ctx, worker)
	}
	<-doneCh
}

func (wp *workloadProjector) configSyncLoop(ctx context.Context, worker int) {
	doneCh := ctx.Done()
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("worker", worker)
	ctx = klog.NewContext(ctx, logger)
	logger.V(4).Info("SyncLoop start")
	for {
		select {
		case <-doneCh:
			logger.V(2).Info("SyncLoop done")
			return
		default:
			ref, shutdown := wp.queue.Get()
			if shutdown {
				logger.V(2).Info("Queue shutdown")
				return
			}
			wp.sync1Config(ctx, ref)
		}
	}
}

func (wp *workloadProjector) sync1Config(ctx context.Context, ref any) {
	destination := ref.(edgeapi.SinglePlacement)
	defer wp.queue.Done(ref)
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Dequeued reference", "ref", ref)
	retry := wp.syncConfig(ctx, destination)
	if retry {
		wp.queue.AddRateLimited(ref)
	} else {
		wp.queue.Forget(ref)
	}
}

func (wp *workloadProjector) syncConfig(ctx context.Context, destination edgeapi.SinglePlacement) bool {
	mbwsName := SPMailboxWorkspaceName(destination)
	logger := klog.FromContext(ctx)
	cluster, have := wp.mbwsNameToCluster.Get(mbwsName)
	if !have {
		logger.V(3).Info("Got reference to unknown mailbox workspace", "destination", destination)
		return true
	}
	wp.edgeClusterClientset.Cluster(cluster.Path())
	// TODO: create/update/delete syncer config object if needed
	return false
}

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
		wp.queue.Add(destination)
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
