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
	"sync"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgeclusterclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgev1a1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
)

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
			cluster := logicalcluster.Name(ws.Spec.Cluster)
			wp.mbwsNameToCluster.Put(ws.Name, cluster)
			wp.clusterToMBWSName.Put(cluster, ws.Name)
		},
		UpdateFunc: func(oldObj, newObj any) {
			ws := newObj.(*tenancyv1a1.Workspace)
			cluster := logicalcluster.Name(ws.Spec.Cluster)
			wp.mbwsNameToCluster.Put(ws.Name, cluster)
			wp.clusterToMBWSName.Put(cluster, ws.Name)
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
			wp.mbwsNameToCluster.Delete(ws.Name)
			wp.clusterToMBWSName.Delete(cluster)
		},
	})
	logger := klog.FromContext(ctx)
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
		logger.V(4).Info("Enqueuing SyncerConfig reference", "ref", ref, "event", event)
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
		logger.V(3).Info("Finally able to enqueue syncer config ref", "ref", ref)
		wp.queue.Add(ref)
		return false
	}
	return true
}

func (wp *workloadProjector) syncConfigObject(ctx context.Context, scRef ExternalName) bool {
	logger := klog.FromContext(ctx)
	mbwsName, ok := wp.clusterToMBWSName.Get(scRef.Cluster)
	if !ok {
		logger.Error(nil, "Failed to map mailbox cluster Name to mailbox WS name", "cluster", scRef.Cluster)
		return true
	}
	sp, ok := wp.mbwsNameToSP.Get(mbwsName)
	if !ok {
		logger.Error(nil, "Failed to map mailbox workspace name to SinglePlacement", "mbwsName", mbwsName, "cluster", scRef.Cluster)
		return true
	}
	syncfg, err := wp.syncfgClusterLister.Cluster(scRef.Cluster).Get(scRef.Name)
	if err != nil && k8sapierrors.IsNotFound(err) {
		syncfg = &edgeapi.SyncerConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name: scRef.Name,
			},
			Spec: wp.syncerConfigSpecForMailbox(sp)}
		client := wp.edgeClusterClientset.EdgeV1alpha1().Cluster(scRef.Cluster.Path()).SyncerConfigs()
		_, err := client.Create(ctx, syncfg, metav1.CreateOptions{FieldManager: "placement-translator"})
		if err == nil {
			logger.V(2).Info("Created SyncerConfig", "ref", scRef)
			return false
		}
		logger.Error(err, "Failed to created SyncerConfig", "ref", scRef)
		return true
	}
	if wp.syncerConfigIsGood(sp, syncfg.Spec) {
		logger.V(4).Info("SyncerConfig is already good", "ref", scRef)
		return false
	}
	syncfg.Spec = wp.syncerConfigSpecForMailbox(sp)
	client := wp.edgeClusterClientset.EdgeV1alpha1().Cluster(scRef.Cluster.Path()).SyncerConfigs()
	_, err = client.Update(ctx, syncfg, metav1.UpdateOptions{FieldManager: "placement-translator"})
	if err != nil {
		logger.Error(err, "SyncerConfig update failed", "ref", scRef)
		return true
	}
	logger.V(2).Info("Updated SyncerConfig", "ref", scRef)
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
		mbwsName := SPMailboxWorkspaceName(destination)
		wp.mbwsNameToSP.Put(mbwsName, destination)
		mbwsCluster, ok := wp.mbwsNameToCluster.Get(mbwsName)
		if !ok {
			logger.Error(nil, "Mailbox workspace not known yet", "destination", destination)
			wp.queue.Add(destination)
		} else {
			ref := ExternalName{mbwsCluster, SyncerConfigName}
			logger.V(3).Info("Enqueuing syncer config reference", "ref", ref)
			wp.queue.Add(ref)
		}
		wp.queue.Add(destination)
		return nil
	})
	wp.changedDestinations = nil
}

const SyncerConfigName = "config"

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

func (wp *workloadProjector) syncerConfigRelations(destination edgeapi.SinglePlacement) (
	namespaces Set[string],
	namespacedResources Set[edgeapi.NamespaceScopeDownsyncResource],
	clusterScopedObjects MutableMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string /*object name*/]]],
) {
	logger := klog.FromContext(wp.ctx)
	wp.Lock()
	defer wp.Unlock()
	nsds, have := wp.nsDistributions.GetIndex1to2().Get(destination)
	if have {
		nses := MapKeySet[NamespaceName, Set[logicalcluster.Name]](nsds.GetIndex1to2())
		namespaces = MapSetCopy(TransformVisitable[NamespaceName, string](nses, func(ns NamespaceName) string { return string(ns) }))
	} else {
		namespaces = NewEmptyMapSet[string]()
	}
	nsrds, haveDists := wp.nsrDistributions.GetIndex1to2().Get(destination)
	if haveDists {
		nsms, haveModes := wp.nsModes.GetIndex().Get(destination)
		if !haveModes {
			logger.Error(nil, "Missing projection modes for namespaced resourced", "destination", destination)
			nsms = NewMapMap[metav1.GroupResource, ProjectionModeVal](nil)
		}
		nsrs := MapKeySet[metav1.GroupResource, Set[logicalcluster.Name]](nsrds.GetIndex1to2())
		namespacedResources = MapSetCopy(TransformVisitable[metav1.GroupResource, edgeapi.NamespaceScopeDownsyncResource](nsrs, func(gr metav1.GroupResource) edgeapi.NamespaceScopeDownsyncResource {
			pmv, ok := nsms.Get(gr)
			if !ok {
				logger.Error(nil, "Missing API group version info", "groupResource", gr, "destination", destination)
			}
			return edgeapi.NamespaceScopeDownsyncResource{GroupResource: gr, APIVersion: pmv.APIVersion}
		}))
	} else {
		namespacedResources = NewEmptyMapSet[edgeapi.NamespaceScopeDownsyncResource]()
	}
	clusterScopedObjects = NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string /*object name*/]]](nil)
	nnsds, haveDists := wp.nnsDistributions.GetIndex1to2().Get(destination)
	if haveDists {
		nnsms, haveModes := wp.nnsModes.GetIndex().Get(destination)
		if !haveModes {
			logger.Error(nil, "Missing projection modes for namespaced resourced", "destination", destination)
			nnsms = NewMapMap[metav1.GroupResource, ProjectionModeVal](nil)
		}
		objs := MapKeySet[GroupResourceInstance, Set[logicalcluster.Name]](nnsds.GetIndex1to2())
		logger.Info("TODO", "x", nnsms, "y", objs)
		objs.Visit(func(gri GroupResourceInstance) error {
			gr := gri.First
			pmv, ok := nnsms.Get(gr)
			if !ok {
				logger.Error(nil, "Missing API version", "obj", gri, "destination", destination)
			}
			cso := MapGetAdd[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string /*object name*/]]](clusterScopedObjects, gr,
				true, func(metav1.GroupResource) Pair[ProjectionModeVal, MutableSet[string /*object name*/]] {
					return NewPair[ProjectionModeVal, MutableSet[string]](pmv, NewEmptyMapSet[string /*object name*/]())
				})
			cso.Second.Add(gri.Second)
			return nil
		})
		// TODO: the rest
	}
	return
}

func (wp *workloadProjector) syncerConfigSpecForMailbox(destination edgeapi.SinglePlacement) edgeapi.SyncerConfigSpec {
	namespaces, namespacedResources, clusterScopedResources := wp.syncerConfigRelations(destination)
	ans := edgeapi.SyncerConfigSpec{
		NamespaceScope: edgeapi.NamespaceScopeDownsyncs{
			Namespaces: SetToSlice(namespaces),
			Resources:  SetToSlice(namespacedResources),
		},
		ClusterScope: MapMapToSlice[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string]], edgeapi.ClusterScopeDownsyncResource](clusterScopedResources,
			func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[string]]) edgeapi.ClusterScopeDownsyncResource {
				return edgeapi.ClusterScopeDownsyncResource{
					GroupResource: key,
					APIVersion:    val.First.APIVersion,
					Objects:       SetToSlice[string](val.Second),
				}
			}),
	}
	return ans
}

func (wp *workloadProjector) syncerConfigIsGood(destination edgeapi.SinglePlacement, spec edgeapi.SyncerConfigSpec) bool {
	goodNamespaces, goodNamespacedResources, goodClusterScopedResources := wp.syncerConfigRelations(destination)
	haveNamespaces := NewMapSet(spec.NamespaceScope.Namespaces...)
	if !SetEqual[string](goodNamespaces, haveNamespaces) {
		return false
	}
	haveNamespacedResources := NewMapSet(spec.NamespaceScope.Resources...)
	if !SetEqual[edgeapi.NamespaceScopeDownsyncResource](goodNamespacedResources, haveNamespacedResources) {
		return false
	}
	haveClusterScopedResources := NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string]]](nil)
	for _, cr := range spec.ClusterScope {
		var objects MutableSet[string] = NewMapSet(cr.Objects...)
		haveClusterScopedResources.Put(cr.GroupResource, NewPair(ProjectionModeVal{cr.APIVersion}, objects))
	}
	return MapEqualParametric[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string]]](csrEqual)(goodClusterScopedResources, haveClusterScopedResources)
}

func csrEqual(a, b Pair[ProjectionModeVal, MutableSet[string]]) bool {
	return a.First == b.First && SetEqual[string](a.Second, b.Second)
}

func MapEqualParametric[Key comparable, Val any](isEqual func(Val, Val) bool) func(map1, map2 Map[Key, Val]) bool {
	return func(map1, map2 Map[Key, Val]) bool {
		if map1.Len() != map2.Len() {
			return false
		}
		return map1.Visit(func(tup1 Pair[Key, Val]) error {
			val2, have := map2.Get(tup1.First)
			if !have || !isEqual(tup1.Second, val2) {
				return errStop
			}
			return nil
		}) == nil
	}
}
