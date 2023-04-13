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
	"fmt"
	"strings"
	"sync"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgeclusterclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgev1a1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
)

const SyncerConfigName = "the-one"

// NewWorkloadProjector constructs a WorkloadProjector that also implements Runnable.
// Run it after starting the informer factories.
func NewWorkloadProjector(
	ctx context.Context,
	configConcurrency int,
	mbwsInformer k8scache.SharedIndexInformer,
	mbwsLister tenancyv1a1listers.WorkspaceLister,
	syncfgClusterInformer kcpcache.ScopeableSharedIndexInformer,
	syncfgClusterLister edgev1a1listers.SyncerConfigClusterLister,
	edgeClusterClientset edgeclusterclientset.ClusterInterface,
) *workloadProjector {
	wp := &workloadProjector{
		ctx:                   ctx,
		configConcurrency:     configConcurrency,
		queue:                 workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		mbwsLister:            mbwsLister,
		syncfgClusterInformer: syncfgClusterInformer,
		syncfgClusterLister:   syncfgClusterLister,
		edgeClusterClientset:  edgeClusterClientset,

		mbwsNameToCluster: WrapMapWithMutex[string, logicalcluster.Name](NewMapMap[string, logicalcluster.Name](nil)),
		clusterToMBWSName: WrapMapWithMutex[logicalcluster.Name, string](NewMapMap[logicalcluster.Name, string](nil)),
		mbwsNameToSP:      WrapMapWithMutex[string, edgeapi.SinglePlacement](NewMapMap[string, edgeapi.SinglePlacement](nil)),

		nsDistributionsForProj:  NewMapRelation3[logicalcluster.Name, NamespaceName, edgeapi.SinglePlacement](),
		nsrDistributionsForProj: NewMapRelation3[logicalcluster.Name, metav1.GroupResource, edgeapi.SinglePlacement](),
		nnsDistributionsForProj: NewMapRelation4[logicalcluster.Name, metav1.GroupResource, string /*obj name*/, edgeapi.SinglePlacement](),

		nsDistributionsForSync:  NewMapRelation3[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name](),
		nsrDistributionsForSync: NewMapRelation3[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name](),
		nnsDistributionsForSync: NewMapRelation3[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name](),

		upsyncs: NewHashRelation2[edgeapi.SinglePlacement, edgeapi.UpsyncSet](
			HashSinglePlacement{}, HashUpsyncSet{}),
	}
	noteModeWrite := func(add bool, destination edgeapi.SinglePlacement) {
		if add {
			(*wp.changedDestinations).Add(destination)
		} else {
			(*wp.changedDestinations).Remove(destination)
		}
	}
	wp.nsModesForSync = NewFactoredMapMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal](factorProjectionModeKeyForSyncer, nil, noteModeWrite, nil)
	wp.nnsModesForSync = NewFactoredMapMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal](factorProjectionModeKeyForSyncer, nil, noteModeWrite, nil)
	logger := klog.FromContext(ctx)
	mbwsInformer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			ws := obj.(*tenancyv1a1.Workspace)
			cluster := logicalcluster.Name(ws.Spec.Cluster)
			if !looksLikeMBWSName(ws.Name) {
				logger.V(4).Info("Ignoring non-mailbox workspace", "wsName", ws.Name, "cluster", cluster)
				return
			}
			wp.mbwsNameToCluster.Put(ws.Name, cluster)
			wp.clusterToMBWSName.Put(cluster, ws.Name)
			logger.V(4).Info("Enqueuing reference to new workspace", "wsName", ws.Name, "cluster", cluster)
			wp.queue.Add(ExternalName{cluster, SyncerConfigName})
		},
		UpdateFunc: func(oldObj, newObj any) {
			ws := newObj.(*tenancyv1a1.Workspace)
			cluster := logicalcluster.Name(ws.Spec.Cluster)
			if !looksLikeMBWSName(ws.Name) {
				logger.V(4).Info("Ignoring non-mailbox workspace", "wsName", ws.Name, "cluster", cluster)
				return
			}
			oldCluster, has := wp.mbwsNameToCluster.Get(ws.Name)
			if !has || cluster != oldCluster {
				wp.mbwsNameToCluster.Put(ws.Name, cluster)
				wp.clusterToMBWSName.Put(cluster, ws.Name)
				logger.V(4).Info("Enqueuing reference to modified workspace", "wsName", ws.Name, "cluster", cluster, "oldCluster", oldCluster)
				wp.queue.Add(ExternalName{cluster, SyncerConfigName})
			}
		},
		DeleteFunc: func(obj any) {
			innerObj := obj
			switch typed := obj.(type) {
			case k8scache.DeletedFinalStateUnknown:
				innerObj = typed.Obj
			default:
			}
			ws := innerObj.(*tenancyv1a1.Workspace)
			cluster := logicalcluster.Name(ws.Spec.Cluster)
			if !looksLikeMBWSName(ws.Name) {
				logger.V(4).Info("Ignoring non-mailbox workspace", "wsName", ws.Name, "cluster", cluster)
				return
			}
			wp.mbwsNameToCluster.Delete(ws.Name)
			wp.clusterToMBWSName.Delete(cluster)
		},
	})
	enqueueObjRef := func(obj any, event string) {
		dfu, ok := obj.(k8scache.DeletedFinalStateUnknown)
		if ok {
			obj = dfu.Obj
		}
		syncfg := obj.(*edgeapi.SyncerConfig)
		cluster := logicalcluster.From(syncfg)
		if syncfg.Name != SyncerConfigName {
			logger.V(4).Info("Ignoring SyncerConfig with non-standard name", "cluster", cluster, "name", syncfg.Name, "standardName", SyncerConfigName)
			return
		}
		ref := ExternalName{cluster, syncfg.Name}
		logger.V(4).Info("Enqueuing reference to written SyncerConfig", "ref", ref, "event", event)
		wp.queue.Add(ref)
	}
	syncfgClusterInformer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { enqueueObjRef(obj, "add") },
		UpdateFunc: func(oldObj, newObj any) { enqueueObjRef(newObj, "update") },
		DeleteFunc: func(obj any) { enqueueObjRef(obj, "delete") },
	})
	return wp
}

var _ WorkloadProjector = &workloadProjector{}
var _ Runnable = &workloadProjector{}

type workloadProjector struct {
	ctx                   context.Context
	configConcurrency     int
	queue                 workqueue.RateLimitingInterface
	mbwsLister            tenancyv1a1listers.WorkspaceLister
	syncfgClusterInformer kcpcache.ScopeableSharedIndexInformer
	syncfgClusterLister   edgev1a1listers.SyncerConfigClusterLister
	edgeClusterClientset  edgeclusterclientset.ClusterInterface
	mbwsNameToCluster     MutableMap[string /*mailbox workspace name*/, logicalcluster.Name]
	clusterToMBWSName     MutableMap[logicalcluster.Name, string /*mailbox workspace name*/]
	mbwsNameToSP          MutableMap[string /*mailbox workspace name*/, edgeapi.SinglePlacement]

	sync.Mutex

	changedDestinations     *MutableSet[edgeapi.SinglePlacement]
	nsDistributionsForProj  SingleIndexedRelation3[logicalcluster.Name, NamespaceName, edgeapi.SinglePlacement]
	nsrDistributionsForProj SingleIndexedRelation3[logicalcluster.Name, metav1.GroupResource, edgeapi.SinglePlacement]
	nnsDistributionsForProj SingleIndexedRelation4[logicalcluster.Name, metav1.GroupResource, string /*object name*/, edgeapi.SinglePlacement]
	nsModesForProj          FactoredMap[ProjectionModeKey, metav1.GroupResource, edgeapi.SinglePlacement, ProjectionModeVal]
	nnsModesForProj         FactoredMap[ProjectionModeKey, metav1.GroupResource, edgeapi.SinglePlacement, ProjectionModeVal]
	nsDistributionsForSync  SingleIndexedRelation3[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name]
	nsrDistributionsForSync SingleIndexedRelation3[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name]
	nnsDistributionsForSync SingleIndexedRelation3[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name]
	nsModesForSync          FactoredMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal]
	nnsModesForSync         FactoredMap[ProjectionModeKey, edgeapi.SinglePlacement, metav1.GroupResource, ProjectionModeVal]
	upsyncs                 SingleIndexedRelation2[edgeapi.SinglePlacement, edgeapi.UpsyncSet]
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
	defer wp.queue.Done(ref)
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Dequeued reference", "ref", ref)
	var retry bool
	switch typed := ref.(type) {
	case edgeapi.SinglePlacement:
		retry = wp.syncConifgDestination(ctx, typed)
	case ExternalName:
		retry = wp.syncConfigObject(ctx, typed)
	default:
		logger.Error(nil, "Dequeued unexpected type of reference", "type", fmt.Sprintf("%T", ref), "val", ref)
	}
	if retry {
		wp.queue.AddRateLimited(ref)
	} else {
		wp.queue.Forget(ref)
	}
}

func (wp *workloadProjector) syncConifgDestination(ctx context.Context, destination edgeapi.SinglePlacement) bool {
	mbwsName := SPMailboxWorkspaceName(destination)
	mbwsCluster, ok := wp.mbwsNameToCluster.Get(mbwsName)
	logger := klog.FromContext(ctx)
	if ok {
		ref := ExternalName{mbwsCluster, SyncerConfigName}
		logger.V(3).Info("Finally able to enqueue SyncerConfig ref", "ref", ref)
		wp.queue.Add(ref)
		return false
	}
	return true
}

func (wp *workloadProjector) syncConfigObject(ctx context.Context, scRef ExternalName) bool {
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("syncerConfig", scRef)
	mbwsName, ok := wp.clusterToMBWSName.Get(scRef.Cluster)
	if !ok {
		logger.Error(nil, "Failed to map mailbox cluster Name to mailbox WS name")
		return true
	}
	sp, ok := wp.mbwsNameToSP.Get(mbwsName)
	if !ok {
		logger.Error(nil, "Failed to map mailbox workspace name to SinglePlacement", "mbwsName", mbwsName)
		return true
	}
	logger = logger.WithValues("destination", sp)
	syncfg, err := wp.syncfgClusterLister.Cluster(scRef.Cluster).Get(scRef.Name)
	if err != nil {
		if k8sapierrors.IsNotFound(err) {
			goodConfigSpecRelations := wp.syncerConfigRelations(sp)
			syncfg = &edgeapi.SyncerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: scRef.Name,
				},
				Spec: wp.syncerConfigSpecFromRelations(goodConfigSpecRelations)}
			client := wp.edgeClusterClientset.EdgeV1alpha1().Cluster(scRef.Cluster.Path()).SyncerConfigs()
			syncfg2, err := client.Create(ctx, syncfg, metav1.CreateOptions{FieldManager: "placement-translator"})
			if logger.V(4).Enabled() {
				logger = logger.WithValues("specNamespaces", syncfg.Spec.NamespaceScope.Namespaces,
					"specResources", syncfg.Spec.NamespaceScope.Resources,
					"specClusterObjects", syncfg.Spec.ClusterScope)
			}
			if err == nil {
				logger.V(2).Info("Created SyncerConfig", "resourceVersion", syncfg2.ResourceVersion)
				return false
			}
			logger.Error(err, "Failed to create SyncerConfig")
			return true
		}
		logger.Error(err, "Unexpected failure reading local cache")
	}
	goodConfigSpecRelations := wp.syncerConfigRelations(sp)
	if wp.syncerConfigIsGood(sp, scRef, syncfg, goodConfigSpecRelations) {
		logger.V(4).Info("SyncerConfig is already good", "resourceVersion", syncfg.ResourceVersion)
		return false
	}
	syncfg.Spec = wp.syncerConfigSpecFromRelations(goodConfigSpecRelations)
	client := wp.edgeClusterClientset.EdgeV1alpha1().Cluster(scRef.Cluster.Path()).SyncerConfigs()
	syncfg2, err := client.Update(ctx, syncfg, metav1.UpdateOptions{FieldManager: "placement-translator"})
	if logger.V(4).Enabled() {
		logger = logger.WithValues("specNamespaces", syncfg.Spec.NamespaceScope.Namespaces,
			"specResources", syncfg.Spec.NamespaceScope.Resources,
			"specClusterObjects", syncfg.Spec.ClusterScope)
	}
	if err != nil {
		logger.Error(err, "SyncerConfig update failed", "resourceVersion", syncfg.ResourceVersion)
		return true
	}
	logger.V(2).Info("Updated SyncerConfig", "resourceVersionOld", syncfg.ResourceVersion, "resourceVersionNew", syncfg2.ResourceVersion)
	return false
}

func (wp *workloadProjector) Transact(xn func(WorkloadProjectionSections)) {
	logger := klog.FromContext(wp.ctx)
	wp.Lock()
	defer wp.Unlock()
	logger.V(3).Info("Begin transaction")
	var s1 SetWriter[Triple[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name]] = wp.nsDistributionsForSync
	var s2 SetWriter[Triple[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name]] = wp.nsrDistributionsForSync
	var s3 SetWriter[Triple[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name]] = wp.nnsDistributionsForSync
	changedDestinations := func() *MutableSet[edgeapi.SinglePlacement] {
		var ms MutableSet[edgeapi.SinglePlacement] = NewMapSet[edgeapi.SinglePlacement]()
		ms = WrapSetWithMutex(ms)
		return &ms
	}()
	wp.changedDestinations = changedDestinations
	recordLogger := logger.V(4)
	changedSources := WrapSetWithMutex[logicalcluster.Name](NewMapSet[logicalcluster.Name]())
	s1 = SetWriterFork(false, s1, recordFirst[edgeapi.SinglePlacement, NamespaceName, logicalcluster.Name](recordLogger, changedDestinations))
	s2 = SetWriterFork(false, s2, recordFirst[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name](recordLogger, changedDestinations))
	s3 = SetWriterFork(false, s3, recordFirst[edgeapi.SinglePlacement, GroupResourceInstance, logicalcluster.Name](recordLogger, changedDestinations))
	j1 := SetWriterFork[Triple[logicalcluster.Name, NamespaceName, edgeapi.SinglePlacement]](false, wp.nsDistributionsForProj,
		recordFirst[logicalcluster.Name, NamespaceName, edgeapi.SinglePlacement](recordLogger, &changedSources),
	)
	j2 := SetWriterFork[Triple[logicalcluster.Name, metav1.GroupResource, edgeapi.SinglePlacement]](false, wp.nsrDistributionsForProj,
		recordFirst[logicalcluster.Name, metav1.GroupResource, edgeapi.SinglePlacement](recordLogger, &changedSources),
	)
	xn(WorkloadProjectionSections{
		SetWriterFork(false, TransformSetWriter(NamespaceDistributionTuple.Reverse, s1), j1),
		SetWriterFork(false, TransformSetWriter(factorNamespacedResourceDistributionTupleForSyncer, s2),
			TransformSetWriter(factorNamespacedResourceDistributionTupleForProj, j2)),
		wp.nsModesForSync,
		TransformSetWriter(factorNonNamespacedDistributionTupleForSyncer, s3),
		wp.nnsModesForSync,
		wp.upsyncs})
	logger.V(3).Info("Transaction response", "changedDestinations", *changedDestinations)
	(*changedDestinations).Visit(func(destination edgeapi.SinglePlacement) error {
		nsds, have := wp.nsDistributionsForSync.GetIndex1to2().Get(destination)
		if have {
			nses := MapKeySet[NamespaceName, Set[logicalcluster.Name]](nsds.GetIndex1to2())
			logger.V(4).Info("Namespaces after transaction", "destination", destination, "namespaces", MapSetCopy[NamespaceName](nses))
		} else {
			logger.V(4).Info("No Namespaces after transaction", "destination", destination)
		}
		nsrds, have := wp.nsrDistributionsForSync.GetIndex1to2().Get(destination)
		if have {
			nsrs := MapKeySet[metav1.GroupResource, Set[logicalcluster.Name]](nsrds.GetIndex1to2())
			logger.V(4).Info("NamespacedResources after transaction", "destination", destination, "resources", MapSetCopy[metav1.GroupResource](nsrs))
		} else {
			logger.V(4).Info("No NamespacedResources after transaction", "destination", destination)
		}
		nsms, have := wp.nsModesForSync.GetIndex().Get(destination)
		if have {
			logger.V(4).Info("Namespaced modes after transaction", "destination", destination, "modes", MapMapCopy[metav1.GroupResource, ProjectionModeVal](nil, nsms))
		} else {
			logger.V(4).Info("No Namespaced modes after transaction", "destination", destination)
		}
		nnsds, have := wp.nnsDistributionsForSync.GetIndex1to2().Get(destination)
		if have {
			objs := MapKeySet[GroupResourceInstance, Set[logicalcluster.Name]](nnsds.GetIndex1to2())
			logger.V(4).Info("NamespacedResources after transaction", "destination", destination, "objs", MapSetCopy[GroupResourceInstance](objs))
		} else {
			logger.V(4).Info("No NamespacedResources after transaction", "destination", destination)
		}
		nnsms, have := wp.nnsModesForSync.GetIndex().Get(destination)
		if have {
			logger.V(4).Info("NonNamespaced modes after transaction", "destination", destination, "modes", MapMapCopy[metav1.GroupResource, ProjectionModeVal](nil, nnsms))
		} else {
			logger.V(4).Info("No NonNamespaced modes after transaction", "destination", destination)
		}
		upsyncs, have := wp.upsyncs.GetIndex1to2().Get(destination)
		if have {
			logger.V(4).Info("Upsyncs after transaction", "destination", destination, "upsyncs", VisitableToSlice[edgeapi.UpsyncSet](upsyncs))
		} else {
			logger.V(4).Info("No Upsyncs after transaction", "destination", destination)
		}
		mbwsName := SPMailboxWorkspaceName(destination)
		wp.mbwsNameToSP.Put(mbwsName, destination)
		mbwsCluster, ok := wp.mbwsNameToCluster.Get(mbwsName)
		if !ok {
			logger.Error(nil, "Mailbox workspace not known yet", "destination", destination)
			wp.queue.Add(destination)
		} else {
			ref := ExternalName{mbwsCluster, SyncerConfigName}
			logger.V(4).Info("Enqueuing reference to SyncerConfig affected by transaction", "destination", destination, "mbwsName", mbwsName, "ref", ref)
			wp.queue.Add(ref)
		}
		return nil
	})
	logger.V(3).Info("End transaction")
	wp.changedDestinations = nil
}

func recordFirst[First, Second, Third comparable](logger klog.Logger, record *MutableSet[First]) SetWriter[Triple[First, Second, Third]] {
	return SetWriterFuncs[Triple[First, Second, Third]]{
		OnAdd: func(tup Triple[First, Second, Third]) bool {
			news := (*record).Add(tup.First)
			logger.Info("Recorded subject of Add", "news", news, "first", tup.First, "second", tup.Second, "third", tup.Third, "revisedSet", *record)
			return true
		},
		OnRemove: func(tup Triple[First, Second, Third]) bool {
			news := (*record).Add(tup.First)
			logger.Info("Recorded subject of Remove", "news", news, "first", tup.First, "second", tup.Second, "third", tup.Third, "revisedSet", *record)
			return true
		}}
}

func factorNamespacedResourceDistributionTupleForProj(nrdt NamespacedResourceDistributionTuple) Triple[logicalcluster.Name, metav1.GroupResource, edgeapi.SinglePlacement] {
	return NewTriple(nrdt.SourceCluster, nrdt.GroupResource, nrdt.Destination)
}

func factorNamespacedResourceDistributionTupleForSyncer(nrdt NamespacedResourceDistributionTuple) Triple[edgeapi.SinglePlacement, metav1.GroupResource, logicalcluster.Name] {
	return NewTriple(nrdt.Destination, nrdt.GroupResource, nrdt.SourceCluster)
}

func factorNonNamespacedDistributionTupleForProj(nndt NonNamespacedDistributionTuple) Quad[logicalcluster.Name, metav1.GroupResource, string /*object name*/, edgeapi.SinglePlacement] {
	return NewQuad(nndt.Second.Cluster, nndt.First.GroupResource, nndt.Second.Name, nndt.First.Destination)
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

// syncerConfigSpecRelations is a relational represetntation of SyncerConfigSpec.
// It takes O(N) to construct and O(N) to compare.
type syncerConfigSpecRelations struct {
	namespaces           Set[string]
	namespacedResources  Set[edgeapi.NamespaceScopeDownsyncResource]
	clusterScopedObjects MutableMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string /*object name*/]]]
	upsyncs              Set[edgeapi.UpsyncSet]
}

func (wp *workloadProjector) syncerConfigRelations(destination edgeapi.SinglePlacement) syncerConfigSpecRelations {
	logger := klog.FromContext(wp.ctx)
	wp.Lock()
	defer wp.Unlock()
	nsds, have := wp.nsDistributionsForSync.GetIndex1to2().Get(destination)
	ans := syncerConfigSpecRelations{
		clusterScopedObjects: NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string /*object name*/]]](nil),
	}
	if have {
		nses := MapKeySet[NamespaceName, Set[logicalcluster.Name]](nsds.GetIndex1to2())
		ans.namespaces = MapSetCopy(TransformVisitable[NamespaceName, string](nses, func(ns NamespaceName) string { return string(ns) }))
	} else {
		ans.namespaces = NewEmptyMapSet[string]()
	}
	nsrds, haveDists := wp.nsrDistributionsForSync.GetIndex1to2().Get(destination)
	if haveDists {
		nsms, haveModes := wp.nsModesForSync.GetIndex().Get(destination)
		if !haveModes {
			logger.Error(nil, "No ProjectionModeVals for namespaced resources", "destination", destination)
			nsms = NewMapMap[metav1.GroupResource, ProjectionModeVal](nil)
		}
		nsrs := MapKeySet[metav1.GroupResource, Set[logicalcluster.Name]](nsrds.GetIndex1to2())
		ans.namespacedResources = MapSetCopy(TransformVisitable[metav1.GroupResource, edgeapi.NamespaceScopeDownsyncResource](nsrs, func(gr metav1.GroupResource) edgeapi.NamespaceScopeDownsyncResource {
			pmv, ok := nsms.Get(gr)
			if !ok {
				logger.Error(nil, "Missing API group version info", "groupResource", gr, "destination", destination)
			}
			return edgeapi.NamespaceScopeDownsyncResource{GroupResource: gr, APIVersion: pmv.APIVersion}
		}))
	} else {
		ans.namespacedResources = NewEmptyMapSet[edgeapi.NamespaceScopeDownsyncResource]()
	}
	nnsds, haveDists := wp.nnsDistributionsForSync.GetIndex1to2().Get(destination)
	if haveDists {
		nnsms, haveModes := wp.nnsModesForSync.GetIndex().Get(destination)
		if !haveModes {
			logger.Error(nil, "No ProjectionModeVals for cluster-scoped resources", "destination", destination)
			nnsms = NewMapMap[metav1.GroupResource, ProjectionModeVal](nil)
		}
		objs := MapKeySet[GroupResourceInstance, Set[logicalcluster.Name]](nnsds.GetIndex1to2())
		objs.Visit(func(gri GroupResourceInstance) error {
			gr := gri.First
			pmv, ok := nnsms.Get(gr)
			if !ok {
				logger.Error(nil, "Missing API version", "obj", gri, "destination", destination)
			}
			cso := MapGetAdd[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string /*object name*/]]](ans.clusterScopedObjects, gr,
				true, func(metav1.GroupResource) Pair[ProjectionModeVal, MutableSet[string /*object name*/]] {
					return NewPair[ProjectionModeVal, MutableSet[string]](pmv, NewEmptyMapSet[string /*object name*/]())
				})
			cso.Second.Add(gri.Second)
			return nil
		})
	}
	upsyncs, haveUpsyncs := wp.upsyncs.GetIndex1to2().Get(destination)
	if !haveUpsyncs {
		upsyncs = NewHashSet[edgeapi.UpsyncSet](HashUpsyncSet{})
	}
	ans.upsyncs = upsyncs
	return ans
}

func (wp *workloadProjector) syncerConfigSpecFromRelations(specRelations syncerConfigSpecRelations) edgeapi.SyncerConfigSpec {
	ans := edgeapi.SyncerConfigSpec{
		NamespaceScope: edgeapi.NamespaceScopeDownsyncs{
			Namespaces: VisitableToSlice[string](specRelations.namespaces),
			Resources:  VisitableToSlice[edgeapi.NamespaceScopeDownsyncResource](specRelations.namespacedResources),
		},
		ClusterScope: MapTransformToSlice[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string]], edgeapi.ClusterScopeDownsyncResource](specRelations.clusterScopedObjects,
			func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[string]]) edgeapi.ClusterScopeDownsyncResource {
				return edgeapi.ClusterScopeDownsyncResource{
					GroupResource: key,
					APIVersion:    val.First.APIVersion,
					Objects:       VisitableToSlice[string](val.Second),
				}
			}),
		Upsync: VisitableToSlice[edgeapi.UpsyncSet](specRelations.upsyncs),
	}
	return ans
}

func (wp *workloadProjector) syncerConfigIsGood(destination edgeapi.SinglePlacement, configRef ExternalName, syncfg *edgeapi.SyncerConfig, goodSpecRelations syncerConfigSpecRelations) bool {
	spec := syncfg.Spec
	haveNamespaces := NewMapSet(spec.NamespaceScope.Namespaces...)
	logger := klog.FromContext(wp.ctx)
	logger = logger.WithValues("destination", destination, "syncerConfig", configRef, "resourceVersion", syncfg.ResourceVersion)
	good := true
	SetEnumerateDifferences[string](goodSpecRelations.namespaces, haveNamespaces, SetWriterFuncs[string]{
		OnAdd: func(namespace string) bool {
			logger.V(4).Info("SyncerConfig has excess namespace", "namespace", namespace)
			good = false
			return false
		},
		OnRemove: func(namespace string) bool {
			logger.V(4).Info("SyncerConfig lacks namespace", "namespace", namespace)
			good = false
			return false
		},
	})
	haveNamespacedResources := NewMapSet(spec.NamespaceScope.Resources...)
	SetEnumerateDifferences[edgeapi.NamespaceScopeDownsyncResource](goodSpecRelations.namespacedResources, haveNamespacedResources, SetWriterFuncs[edgeapi.NamespaceScopeDownsyncResource]{
		OnAdd: func(nsr edgeapi.NamespaceScopeDownsyncResource) bool {
			logger.V(4).Info("SyncerConfig has excess NamespaceScopeDownsyncResource", "resource", nsr)
			good = false
			return false
		},
		OnRemove: func(nsr edgeapi.NamespaceScopeDownsyncResource) bool {
			logger.V(4).Info("SyncerConfig lacks NamespaceScopeDownsyncResource", "resource", nsr)
			good = false
			return false
		},
	})
	haveClusterScopedResources := NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string]]](nil)
	for _, cr := range spec.ClusterScope {
		var objects MutableSet[string] = NewMapSet(cr.Objects...)
		haveClusterScopedResources.Put(cr.GroupResource, NewPair(ProjectionModeVal{cr.APIVersion}, objects))
	}
	MapEnumerateDifferencesParametric[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string]]](csrEqual, goodSpecRelations.clusterScopedObjects, haveClusterScopedResources, MapChangeReceiverFuncs[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string]]]{
		OnCreate: func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[string]]) {
			logger.V(4).Info("SyncerConfig has excess ClusterScopeDownsyncResource", "groupResource", key, "apiVersion", val.First.APIVersion, "objects", val.Second)
			good = false
		},
		OnUpdate: func(key metav1.GroupResource, goodVal, haveVal Pair[ProjectionModeVal, MutableSet[string]]) {
			logger.V(4).Info("SyncerConfig wrong ClusterScopeDownsyncResource", "groupResource", key, "apiVersionGood", goodVal.First.APIVersion, "apiVersionHave", haveVal.First.APIVersion, "objectsGood", goodVal.Second, "objectsHave", haveVal.Second)
			good = false
		},
		OnDelete: func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[string]]) {
			logger.V(4).Info("SyncerConfig lacks ClusterScopeDownsyncResource", "groupResource", key, "apiVersion", val.First.APIVersion, "objects", val.Second)
			good = false
		},
	})
	haveUpsyncs := NewHashSet[edgeapi.UpsyncSet](HashUpsyncSet{}, spec.Upsync...)
	SetEnumerateDifferences[edgeapi.UpsyncSet](goodSpecRelations.upsyncs, haveUpsyncs, SetWriterFuncs[edgeapi.UpsyncSet]{
		OnAdd: func(upsync edgeapi.UpsyncSet) bool {
			logger.V(4).Info("SyncerConfig has excess UpsyncSet", "upsync", upsync)
			good = false
			return false
		},
		OnRemove: func(upsync edgeapi.UpsyncSet) bool {
			logger.V(4).Info("SyncerConfig lacks UpsyncSet", "upsync", upsync)
			good = false
			return false
		},
	})
	return good
}

func csrEqual(a, b Pair[ProjectionModeVal, MutableSet[string]]) bool {
	return a.First == b.First && SetEqual[string](a.Second, b.Second)
}

func looksLikeMBWSName(wsName string) bool {
	mbwsNameParts := strings.Split(wsName, WSNameSep)
	return len(mbwsNameParts) == 2
}
