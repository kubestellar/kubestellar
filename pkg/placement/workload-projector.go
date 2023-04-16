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
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	k8scorev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	machruntime "k8s.io/apimachinery/pkg/runtime"
	k8sdynamic "k8s.io/client-go/dynamic"
	k8sdynamicinformer "k8s.io/client-go/dynamic/dynamicinformer"
	upstreaminformers "k8s.io/client-go/informers"
	k8scorev1informers "k8s.io/client-go/informers/core/v1"
	k8scorev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	clusterdynamic "github.com/kcp-dev/client-go/dynamic"
	kcpkubecorev1informers "github.com/kcp-dev/client-go/informers/core/v1"
	kcpkubecorev1client "github.com/kcp-dev/client-go/kubernetes/typed/core/v1"
	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgeclusterclientset "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/cluster"
	edgev1a1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
)

const SyncerConfigName = "the-one"
const FieldManager = "placement-translator"

// This workload projector:
// (a) maintains SyncerConfig objects in mailbox workspaces, and
// (b) propagates changes from source workspaces to mailbox workspaces.
//
// This workload projector currently does not react to changes in workload objects
// in mailbox workspaces. There is currently no designed need for that, except
// perhaps the general principal that a controller should overwriting competing writes.
//
// This workload projector maintains a dynamic informer for each relevant combination
// of source cluster, API group, and resource.  Further filtering is done here,
// not in the apiserver.
//
// This workload projector currently does not attempt to resolve version conflicts,
// and will complain and pick one arbitrarily if such a conflict is detected.

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
	dynamicClusterClient clusterdynamic.ClusterInterface,
	nsClusterPreInformer kcpkubecorev1informers.NamespaceClusterInformer,
	nsClusterClient kcpkubecorev1client.NamespaceClusterInterface,
) *workloadProjector {
	wp := &workloadProjector{
		// delay:                 2 * time.Second,
		ctx:                   ctx,
		configConcurrency:     configConcurrency,
		queue:                 workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		mbwsLister:            mbwsLister,
		syncfgClusterInformer: syncfgClusterInformer,
		syncfgClusterLister:   syncfgClusterLister,
		edgeClusterClientset:  edgeClusterClientset,
		dynamicClusterClient:  dynamicClusterClient,
		nsClusterPreInformer:  nsClusterPreInformer,
		nsClusterClient:       nsClusterClient,

		mbwsNameToCluster: WrapMapWithMutex[string, logicalcluster.Name](NewMapMap[string, logicalcluster.Name](nil)),
		clusterToMBWSName: WrapMapWithMutex[logicalcluster.Name, string](NewMapMap[logicalcluster.Name, string](nil)),
		mbwsNameToSP:      WrapMapWithMutex[string, SinglePlacement](NewMapMap[string, SinglePlacement](nil)),

		perSource:      NewMapMap[logicalcluster.Name, *wpPerSource](nil),
		perDestination: NewMapMap[SinglePlacement, *wpPerDestination](nil),

		upsyncs: NewHashRelation2[SinglePlacement, edgeapi.UpsyncSet](
			HashSinglePlacement{}, HashUpsyncSet{}),
	}
	wp.nsDistributionsForProj = NewGenericIndexedSet[NamespaceDistributionTuple, logicalcluster.Name, Pair[NamespaceName, SinglePlacement],
		wpPerSourceNSDistributions, wpPerSourceNSDistributions](
		TripleFactorerTo1and23[logicalcluster.Name, NamespaceName, SinglePlacement](),
		func(source logicalcluster.Name) wpPerSourceNSDistributions {
			wps := MapGetAdd(wp.perSource, source, true, wp.newPerSourceLocked)
			return wpPerSourceNSDistributions{wps}
		},
		func(nsd wpPerSourceNSDistributions) MutableSet[Pair[NamespaceName, SinglePlacement]] {
			return nsd.wps.nsDistributions
		},
		Identity1[wpPerSourceNSDistributions],
		NewMapMap[logicalcluster.Name, wpPerSourceNSDistributions](nil),
	)
	wp.nsrDistributionsForProj = NewGenericIndexedSet[NamespacedResourceDistributionTuple,
		logicalcluster.Name, Pair[metav1.GroupResource, SinglePlacement],
		wpPerSourceNSRDistributions, wpPerSourceNSRDistributions](
		factorNamespacedResourceDistributionTupleForProj1,
		func(source logicalcluster.Name) wpPerSourceNSRDistributions {
			wps := MapGetAdd(wp.perSource, source, true, wp.newPerSourceLocked)
			return wpPerSourceNSRDistributions{wps}
		},
		func(nsd wpPerSourceNSRDistributions) MutableSet[Pair[metav1.GroupResource, SinglePlacement]] {
			return nsd.wps.nsrDistributions
		},
		Identity1[wpPerSourceNSRDistributions],
		NewMapMap[logicalcluster.Name, wpPerSourceNSRDistributions](nil),
	)
	wp.nnsDistributionsForProj = NewGenericIndexedSet[NonNamespacedDistributionTuple,
		logicalcluster.Name, Triple[metav1.GroupResource, string /*obj name*/, SinglePlacement],
		wpPerSourceNNSDistributions, wpPerSourceNNSDistributions](
		factorNonNamespacedDistributionTupleForProj1and234,
		func(source logicalcluster.Name) wpPerSourceNNSDistributions {
			wps := MapGetAdd(wp.perSource, source, true, wp.newPerSourceLocked)
			return wpPerSourceNNSDistributions{wps}
		},
		func(nsd wpPerSourceNNSDistributions) MutableSet[Triple[metav1.GroupResource, string /*obj name*/, SinglePlacement]] {
			return nsd.wps.nnsDistributions
		},
		Identity1[wpPerSourceNNSDistributions],
		NewMapMap[logicalcluster.Name, wpPerSourceNNSDistributions](nil),
	)
	wp.nsDistributionsForSync = NewGenericIndexedSet[NamespaceDistributionTuple, SinglePlacement, Pair[NamespaceName, logicalcluster.Name],
		wpPerDestinationNSDistributions, wpPerDestinationNSDistributions](
		TripleFactorerTo3and21[logicalcluster.Name, NamespaceName, SinglePlacement](),
		func(destination SinglePlacement) wpPerDestinationNSDistributions {
			wpd := MapGetAdd(wp.perDestination, destination, true, wp.newPerDestinationLocked)
			return wpPerDestinationNSDistributions{wpd}
		},
		func(nsd wpPerDestinationNSDistributions) MutableSet[Pair[NamespaceName, logicalcluster.Name]] {
			return nsd.wpd.nsDistributions
		},
		Identity1[wpPerDestinationNSDistributions],
		NewMapMap[SinglePlacement, wpPerDestinationNSDistributions](nil),
	)
	wp.nsrDistributionsForSync = NewGenericIndexedSet[NamespacedResourceDistributionTuple, SinglePlacement, Pair[metav1.GroupResource, logicalcluster.Name],
		wpPerDestinationNSRDistributions, wpPerDestinationNSRDistributions](
		factorNamespacedResourceDistributionTupleForSync1,
		func(destination SinglePlacement) wpPerDestinationNSRDistributions {
			wpd := MapGetAdd(wp.perDestination, destination, true, wp.newPerDestinationLocked)
			return wpPerDestinationNSRDistributions{wpd}
		},
		func(nsd wpPerDestinationNSRDistributions) MutableSet[Pair[metav1.GroupResource, logicalcluster.Name]] {
			return nsd.wpd.nsrDistributions
		},
		Identity1[wpPerDestinationNSRDistributions],
		NewMapMap[SinglePlacement, wpPerDestinationNSRDistributions](nil),
	)
	wp.nnsDistributionsForSync = NewGenericIndexedSet[NonNamespacedDistributionTuple, SinglePlacement, Pair[GroupResourceInstance, logicalcluster.Name],
		wpPerDestinationNNSDistributions, wpPerDestinationNNSDistributions](
		factorNonNamespacedDistributionTupleForSync1,
		func(destination SinglePlacement) wpPerDestinationNNSDistributions {
			wpd := MapGetAdd(wp.perDestination, destination, true, wp.newPerDestinationLocked)
			return wpPerDestinationNNSDistributions{wpd}
		},
		func(nsd wpPerDestinationNNSDistributions) MutableSet[Pair[GroupResourceInstance, logicalcluster.Name]] {
			return nsd.wpd.nnsDistributions
		},
		Identity1[wpPerDestinationNNSDistributions],
		NewMapMap[SinglePlacement, wpPerDestinationNNSDistributions](nil),
	)
	noteModeWrite := func(add bool, destination SinglePlacement) {
		if add {
			(*wp.changedDestinations).Add(destination)
		} else {
			(*wp.changedDestinations).Remove(destination)
		}
	}
	wp.nsModesForSync = NewFactoredMapMap[ProjectionModeKey, SinglePlacement, metav1.GroupResource, ProjectionModeVal](factorProjectionModeKeyForSyncer, nil, noteModeWrite, nil)
	wp.nnsModesForSync = NewFactoredMapMap[ProjectionModeKey, SinglePlacement, metav1.GroupResource, ProjectionModeVal](factorProjectionModeKeyForSyncer, nil, noteModeWrite, nil)
	wp.nsModesForProj = NewFactoredMapMap[ProjectionModeKey, metav1.GroupResource, SinglePlacement, ProjectionModeVal](factorProjectionModeKeyForProj, nil, nil, nil)
	wp.nnsModesForProj = NewFactoredMapMap[ProjectionModeKey, metav1.GroupResource, SinglePlacement, ProjectionModeVal](factorProjectionModeKeyForProj, nil, nil, nil)
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
			scRef := syncerConfigRef{cluster, SyncerConfigName}
			logger.V(4).Info("Enqueuing reference to SyncerConfig in new workspace", "wsName", ws.Name, "cluster", cluster)
			wp.queue.Add(scRef)
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
				scRef := syncerConfigRef{cluster, SyncerConfigName}
				logger.V(4).Info("Enqueuing reference to SyncerConfig of modified workspace", "wsName", ws.Name, "cluster", cluster, "oldCluster", oldCluster)
				wp.queue.Add(scRef)
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
	enqueueSCRef := func(obj any, event string) {
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
		scRef := syncerConfigRef{cluster, syncfg.Name}
		logger.V(4).Info("Enqueuing reference to SyncerConfig from informer", "scRef", scRef, "event", event)
		wp.queue.Add(scRef)
	}
	syncfgClusterInformer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { enqueueSCRef(obj, "add") },
		UpdateFunc: func(oldObj, newObj any) { enqueueSCRef(newObj, "update") },
		DeleteFunc: func(obj any) { enqueueSCRef(obj, "delete") },
	})
	return wp
}

var _ WorkloadProjector = &workloadProjector{}
var _ Runnable = &workloadProjector{}

// workloadProjector is the main data structure here.
// For each of the downsync sets in its WorkloadProjectionSections,
// this projector maintains two map-based representations: one is an index
// convenient to use in making SyncerConfig objects, and the other is
// an index convenient to use in copyint the workload objects from WMW to mailbox.
// These indices are not the plain tuple-based ones (e.g., SingleIndexedRelation3)
// but rather take advantage of the
// generic index ability to use arbitrary types for both the whole tuple
// and the value set.
// The value set types are thin wrappers (wpPerSourceNSDistributions et al.) around
// the relevant per-cluster data structures, each of which has the various
// second-level indices in it and the wrapper exposes the relevant second-level index.
// We currently do not bother with distinct types for readonly indices.
//
// The fields following the Mutex should only be accessed with the Mutex locked.
type workloadProjector struct {
	ctx                   context.Context
	configConcurrency     int
	delay                 time.Duration // to slow down for debugging
	queue                 workqueue.RateLimitingInterface
	mbwsLister            tenancyv1a1listers.WorkspaceLister
	syncfgClusterInformer kcpcache.ScopeableSharedIndexInformer
	syncfgClusterLister   edgev1a1listers.SyncerConfigClusterLister
	edgeClusterClientset  edgeclusterclientset.ClusterInterface
	dynamicClusterClient  clusterdynamic.ClusterInterface
	nsClusterPreInformer  kcpkubecorev1informers.NamespaceClusterInformer
	nsClusterClient       kcpkubecorev1client.NamespaceClusterInterface

	mbwsNameToCluster MutableMap[string /*mailbox workspace name*/, logicalcluster.Name]
	clusterToMBWSName MutableMap[logicalcluster.Name, string /*mailbox workspace name*/]
	mbwsNameToSP      MutableMap[string /*mailbox workspace name*/, SinglePlacement]

	sync.Mutex

	perSource      MutableMap[logicalcluster.Name, *wpPerSource]
	perDestination MutableMap[SinglePlacement, *wpPerDestination]

	// changedDestinations is the destinations affected during a transaction
	changedDestinations *MutableSet[SinglePlacement]

	// NamespaceDistributions indexed for projection from source to mailbox
	nsDistributionsForProj GenericMutableIndexedSet[NamespaceDistributionTuple, logicalcluster.Name,
		Pair[NamespaceName, SinglePlacement], wpPerSourceNSDistributions]

	// NamespacedResourceDistributions indexed for projection
	nsrDistributionsForProj GenericMutableIndexedSet[NamespacedResourceDistributionTuple, logicalcluster.Name,
		Pair[metav1.GroupResource, SinglePlacement], wpPerSourceNSRDistributions]

	// NonNamespacedDistributions indexed for projection
	nnsDistributionsForProj GenericMutableIndexedSet[NonNamespacedDistributionTuple, logicalcluster.Name,
		Triple[metav1.GroupResource, string /*obj name*/, SinglePlacement], wpPerSourceNNSDistributions]

	nsModesForProj  FactoredMap[ProjectionModeKey, metav1.GroupResource, SinglePlacement, ProjectionModeVal]
	nnsModesForProj FactoredMap[ProjectionModeKey, metav1.GroupResource, SinglePlacement, ProjectionModeVal]

	// NamespaceDistributions indexed for SyncerConfig maintenance
	nsDistributionsForSync GenericMutableIndexedSet[NamespaceDistributionTuple, SinglePlacement,
		Pair[NamespaceName, logicalcluster.Name], wpPerDestinationNSDistributions]

	// NamespacedResourceDistributions indexed for SyncerConfig maintenance
	nsrDistributionsForSync GenericMutableIndexedSet[NamespacedResourceDistributionTuple, SinglePlacement,
		Pair[metav1.GroupResource, logicalcluster.Name], wpPerDestinationNSRDistributions]

	// NonNamespacedDistributions indexed for SyncerConfig maintenance
	nnsDistributionsForSync GenericMutableIndexedSet[NonNamespacedDistributionTuple, SinglePlacement,
		Pair[GroupResourceInstance, logicalcluster.Name], wpPerDestinationNNSDistributions]

	nsModesForSync  FactoredMap[ProjectionModeKey, SinglePlacement, metav1.GroupResource, ProjectionModeVal]
	nnsModesForSync FactoredMap[ProjectionModeKey, SinglePlacement, metav1.GroupResource, ProjectionModeVal]

	upsyncs SingleIndexedRelation2[SinglePlacement, edgeapi.UpsyncSet]
}

type GroupResourceInstance = Pair[metav1.GroupResource, string /*object name*/]

// Constructs the data structure specific to a mailbox/edge-cluster
func (wp *workloadProjector) newPerDestinationLocked(destination SinglePlacement) *wpPerDestination {
	wpd := &wpPerDestination{wp: wp, destination: destination,
		nsDistributions:  NewMapRelation2[NamespaceName, logicalcluster.Name](),
		nsrDistributions: NewMapRelation2[metav1.GroupResource, logicalcluster.Name](),
		nnsDistributions: NewMapRelation2[GroupResourceInstance, logicalcluster.Name](),
	}
	return wpd
}

// The data structure specific to a mailbox/edge-cluster.
// All the variable fields must be accessed with the wp mutex locked.
// The readyChan is closed once the namespace informer has synced.
type wpPerDestination struct {
	wp                     *workloadProjector
	destination            SinglePlacement
	nsDistributions        SingleIndexedRelation2[NamespaceName, logicalcluster.Name]
	nsrDistributions       SingleIndexedRelation2[metav1.GroupResource, logicalcluster.Name]
	nnsDistributions       SingleIndexedRelation2[GroupResourceInstance, logicalcluster.Name]
	dynamicClient          k8sdynamic.Interface
	dynamicInformerFactory k8sdynamicinformer.DynamicSharedInformerFactory
	namespaceClient        k8scorev1client.NamespaceInterface
	namespacePreInformer   k8scorev1informers.NamespaceInformer
	readyChan              <-chan struct{}
}

func (wpd *wpPerDestination) getDynamicClientLocked() (k8sdynamic.Interface, <-chan struct{}, error) {
	if wpd.dynamicClient == nil {
		mbwsName := SPMailboxWorkspaceName(wpd.destination)
		mbwsCluster, have := wpd.wp.mbwsNameToCluster.Get(mbwsName)
		if !have {
			return nil, nil, errors.New("unable to map mailbox workspace name to cluster")
		}
		wpd.dynamicClient = wpd.wp.dynamicClusterClient.Cluster(mbwsCluster.Path())
		wpd.dynamicInformerFactory = k8sdynamicinformer.NewDynamicSharedInformerFactory(wpd.dynamicClient, 0)
		wpd.namespaceClient = wpd.wp.nsClusterClient.Cluster(mbwsCluster.Path())
		wpd.namespacePreInformer = wpd.wp.nsClusterPreInformer.Cluster(mbwsCluster)
		nsInformer := wpd.namespacePreInformer.Informer()
		readyChan := make(chan struct{})
		wpd.readyChan = readyChan
		go func() {
			k8scache.WaitForNamedCacheSync("workload-projector", wpd.wp.ctx.Done(), nsInformer.HasSynced)
			close(readyChan)
		}()
		go nsInformer.Run(wpd.wp.ctx.Done())
		wpd.dynamicInformerFactory.Start(wpd.wp.ctx.Done())
	}
	return wpd.dynamicClient, wpd.readyChan, nil

}

type wpPerDestinationNSDistributions struct {
	wpd *wpPerDestination
}

func (nsd wpPerDestinationNSDistributions) GetIndex1to2() Map[NamespaceName, Set[logicalcluster.Name]] {
	return nsd.wpd.nsDistributions.GetIndex1to2()
}

type wpPerDestinationNSRDistributions struct {
	wpd *wpPerDestination
}

func (nsrd wpPerDestinationNSRDistributions) GetIndex1to2() Map[metav1.GroupResource, Set[logicalcluster.Name]] {
	return nsrd.wpd.nsrDistributions.GetIndex1to2()
}

type wpPerDestinationNNSDistributions struct {
	wpd *wpPerDestination
}

func (nsd wpPerDestinationNNSDistributions) GetIndex1to2() Map[GroupResourceInstance, Set[logicalcluster.Name]] {
	return nsd.wpd.nnsDistributions.GetIndex1to2()
}

// Constructs the data structure specific to a workload management workspace
func (wp *workloadProjector) newPerSourceLocked(source logicalcluster.Name) *wpPerSource {
	dynamicClient := wp.dynamicClusterClient.Cluster(source.Path())
	dynamicInformerFactory := k8sdynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0)
	wps := &wpPerSource{wp: wp, source: source,
		logger:                 klog.FromContext(wp.ctx).WithValues("source", source),
		nsDistributions:        NewMapRelation2[NamespaceName, SinglePlacement](),
		nsrDistributions:       NewMapRelation2[metav1.GroupResource, SinglePlacement](),
		nnsDistributions:       NewMapRelation3[metav1.GroupResource, string /*obj name*/, SinglePlacement](),
		dynamicClient:          dynamicClient,
		dynamicInformerFactory: dynamicInformerFactory,
		preInformers:           NewMapMap[metav1.GroupResource, upstreaminformers.GenericInformer](nil),
	}
	dynamicInformerFactory.Start(wp.ctx.Done())
	return wps
}

// The data structure specific to a workload management workspace.
// All the variable fields must be accessed with the wp mutex locked.
type wpPerSource struct {
	wp                     *workloadProjector
	source                 logicalcluster.Name
	logger                 klog.Logger
	nsDistributions        SingleIndexedRelation2[NamespaceName, SinglePlacement]
	nsrDistributions       SingleIndexedRelation2[metav1.GroupResource, SinglePlacement]
	nnsDistributions       SingleIndexedRelation3[metav1.GroupResource, string /*obj name*/, SinglePlacement]
	dynamicClient          k8sdynamic.Interface
	dynamicInformerFactory k8sdynamicinformer.DynamicSharedInformerFactory
	preInformers           MutableMap[metav1.GroupResource, upstreaminformers.GenericInformer]
}

type wpPerSourceNSDistributions struct {
	wps *wpPerSource
}

func (nsd wpPerSourceNSDistributions) GetIndex1to2() Map[NamespaceName, Set[SinglePlacement]] {
	return nsd.wps.nsDistributions.GetIndex1to2()
}

type wpPerSourceNSRDistributions struct {
	wps *wpPerSource
}

func (nsrd wpPerSourceNSRDistributions) GetIndex1to2() Map[metav1.GroupResource, Set[SinglePlacement]] {
	return nsrd.wps.nsrDistributions.GetIndex1to2()
}

type wpPerSourceNNSDistributions struct {
	wps *wpPerSource
}

func (nsd wpPerSourceNNSDistributions) GetIndex1to2() Index2[metav1.GroupResource,
	Pair[string /*obj name*/, SinglePlacement], ObjectNameToDestinations] {
	return nsd.wps.nnsDistributions.GetIndex1to2()
}

type ObjectNameToDestinations = GenericIndexedSet[Pair[string /*obj name*/, SinglePlacement],
	string /*obj name*/, SinglePlacement, Set[SinglePlacement]]

// The workqueue contains the following types of object references.
// - SinglePlacementSlice
// - syncerConfigRef
// - sourceObjectRef

// syncerConfigRef is a workqueue item that refers to a SyncerConfig in a mailbox workspace
type syncerConfigRef ExternalName

// sourceObjectRef refers to an namespaced object in a workload management workspace
type sourceObjectRef struct {
	cluster       logicalcluster.Name
	groupResource metav1.GroupResource
	namespace     string // == noNamespace iff not namespaced
	name          string
}

const noNamespace = "no NS"

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
	case SinglePlacement:
		retry = wp.syncConifgDestination(ctx, typed)
	case syncerConfigRef:
		retry = wp.syncConfigObject(ctx, typed)
	case sourceObjectRef:
		retry = wp.syncSourceObject(ctx, typed)
	default:
		logger.Error(nil, "Dequeued unexpected type of reference", "type", fmt.Sprintf("%T", ref), "val", ref)
	}
	if retry {
		wp.queue.AddRateLimited(ref)
	} else {
		wp.queue.Forget(ref)
	}
}

func (wp *workloadProjector) syncConifgDestination(ctx context.Context, destination SinglePlacement) bool {
	mbwsName := SPMailboxWorkspaceName(destination)
	mbwsCluster, ok := wp.mbwsNameToCluster.Get(mbwsName)
	logger := klog.FromContext(ctx)
	if ok {
		scRef := syncerConfigRef{mbwsCluster, SyncerConfigName}
		logger.V(3).Info("Finally able to enqueue SyncerConfig ref", "scRef", scRef)
		wp.queue.Add(scRef)
		return false
	}
	return true
}

func (wp *workloadProjector) syncConfigObject(ctx context.Context, scRef syncerConfigRef) bool {
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
			syncfg2, err := client.Create(ctx, syncfg, metav1.CreateOptions{FieldManager: FieldManager})
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
	if wp.syncerConfigIsGood(sp, ExternalName(scRef), syncfg, goodConfigSpecRelations) {
		logger.V(4).Info("SyncerConfig is already good", "resourceVersion", syncfg.ResourceVersion)
		return false
	}
	syncfg.Spec = wp.syncerConfigSpecFromRelations(goodConfigSpecRelations)
	client := wp.edgeClusterClientset.EdgeV1alpha1().Cluster(scRef.Cluster.Path()).SyncerConfigs()
	syncfg2, err := client.Update(ctx, syncfg, metav1.UpdateOptions{FieldManager: FieldManager})
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

func (wp *workloadProjector) syncSourceObject(ctx context.Context, soRef sourceObjectRef) bool {
	namespaced := soRef.namespace != noNamespace
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("objectRef", soRef)
	finish := func() []func() bool { // produce the work to do after releasing the mutex
		wp.Lock()
		defer wp.Unlock()
		wps, have := wp.perSource.Get(soRef.cluster)
		if !have {
			logger.Error(nil, "Impossible: handing object from unknown source")
			return []func() bool{returnFalse}
		}
		preInformer, have := wps.preInformers.Get(soRef.groupResource)
		if !have {
			logger.Error(nil, "Impossible: handling source object of unknown resource")
			return []func() bool{returnFalse}
		}
		var srcObj machruntime.Object
		var err error
		if namespaced {
			srcObj, err = preInformer.Lister().ByNamespace(soRef.namespace).Get(soRef.name)
		} else {
			srcObj, err = preInformer.Lister().Get(soRef.name)
		}
		if err != nil && !k8sapierrors.IsNotFound(err) {
			logger.Error(nil, "Impossible: failed to lookup source object in local cache")
			return []func() bool{returnFalse}
		}
		var srcMRObject mrObject
		deleted := srcObj == nil || k8sapierrors.IsNotFound(err)
		if !deleted {
			srcMRObject = srcObj.(mrObject)
			deleted = srcMRObject.GetDeletionTimestamp() != nil
		}
		var destinations Set[SinglePlacement]
		var haveDestinations bool
		if namespaced {
			destinations, haveDestinations = wps.nsDistributions.GetIndex1to2().Get(NamespaceName(soRef.namespace))
		} else {
			byName, have := wps.nnsDistributions.GetIndex1to2().Get(soRef.groupResource)
			if !have {
				logger.Error(nil, "Missing index for cluster-scoped objects")
			} else {
				destinations, haveDestinations = byName.GetIndex1to2().Get(soRef.name)
			}
		}
		if !haveDestinations {
			logger.V(4).Info("Object is not going anywhere")
			return []func() bool{returnFalse}
		} else {
			logger.V(4).Info("Object is going places", "num", destinations.Len())
		}
		modesForSync := wps.wp.nnsModesForSync
		if namespaced {
			modesForSync = wps.wp.nsModesForSync
		}
		var tryAgain bool
		remWork := []func() bool{}
		destinations.Visit(func(destination SinglePlacement) error {
			retryThis, rem := wp.syncSourceToDest(ctx, logger, soRef, srcMRObject, namespaced, deleted, modesForSync, destination)
			tryAgain = tryAgain || retryThis
			if rem != nil {
				remWork = append(remWork, rem)
			}
			return nil
		})
		if tryAgain {
			remWork = append(remWork, returnTrue)
		}
		return remWork
	}()
	hadBad := false
	for _, work := range finish {
		if work() {
			hadBad = true
		}
	}
	return hadBad
}

var returnFalse = func() bool { return false }
var returnTrue = func() bool { return true }

func (wp *workloadProjector) syncSourceToDest(ctx context.Context, logger klog.Logger,
	soRef sourceObjectRef, srcMRObject mrObject, namespaced, deleted bool,
	modesForSync FactoredMap[ProjectionModeKey, SinglePlacement, metav1.GroupResource, ProjectionModeVal],
	destination SinglePlacement) (bool, func() bool) {
	logger = logger.WithValues("destination", destination)
	wpd, have := wp.perDestination.Get(destination)
	if !have {
		logger.Error(nil, "Impossible: object going to unknown destination")
		return true, nil
	}
	pmv, have := modesForSync.Get(ProjectionModeKey{soRef.groupResource, destination})
	if !have {
		logger.Error(nil, "Missing version")
		return true, nil
	}
	dynamicClient, clientReadyChan, err := wpd.getDynamicClientLocked()
	if err != nil {
		logger.Error(err, "Failed to wpd.getDynamicClientLocked")
		return true, nil
	}
	return false, func() bool {
		sgvr := MetaGroupResourceToSchema(soRef.groupResource).WithVersion(pmv.APIVersion)
		nsblClient := dynamicClient.Resource(sgvr)
		var rscClient k8sdynamic.ResourceInterface = nsblClient
		if namespaced {
			rscClient = nsblClient.Namespace(soRef.namespace)
		}
		if deleted { // propagate deletion
			time.Sleep(wp.delay)
			err := rscClient.Delete(ctx, soRef.name, metav1.DeleteOptions{})
			if err == nil {
				logger.V(3).Info("Deleted object in mailbox workspace")
			} else if !k8sapierrors.IsNotFound(err) {
				logger.Error(err, "Failed to delete object in mailbox workspace")
				return true
			} else {
				logger.V(3).Info("Deletion already propagated")
			}
			return false
		}
		if namespaced {
			<-clientReadyChan
			nsObj, err := wpd.namespacePreInformer.Lister().Get(soRef.namespace)
			if err != nil && !k8sapierrors.IsNotFound(err) {
				logger.Error(err, "Failed to lookup namespace in local cache")
				return true
			}
			if nsObj == nil {
				nsObj = &k8scorev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: soRef.namespace,
					},
				}
				_, err := wpd.namespaceClient.Create(ctx, nsObj, metav1.CreateOptions{})
				if err == nil {
					logger.V(3).Info("Created namespace in mailbox workspace")
				} else if k8sapierrors.IsAlreadyExists(err) {
					logger.V(4).Info("Something else created needed namespace concurrently")
				} else {
					logger.Error(err, "Failed to create needed namespace in mailbox workspace")
					return true
				}
			}
		}
		destObj, err := rscClient.Get(ctx, soRef.name, metav1.GetOptions{})
		if err != nil && !k8sapierrors.IsNotFound(err) {
			logger.Error(err, "Failed to fetch object from mailbox workspace")
			return true
		} else if err == nil {
			revisedDestObj := wpd.wp.genericObjectMerge(soRef.cluster, destination, srcMRObject, destObj)
			if apiequality.Semantic.DeepEqual(destObj, revisedDestObj) {
				logger.V(4).Info("No need to update object in mailbox workspace")
				return false
			}
			time.Sleep(wp.delay)
			_, err = rscClient.Update(ctx, revisedDestObj, metav1.UpdateOptions{FieldManager: FieldManager})
			if err != nil {
				logger.Error(err, "Failed to update object in mailbox workspace")
				return true
			}
			if logger.V(5).Enabled() {
				logger = logger.WithValues("oldVal", destObj, "newVal", revisedDestObj)
			}
			logger.V(3).Info("Updated object in mailbox workspace")
			return false
		}
		destObj = wpd.wp.trimForDestination(srcMRObject)
		time.Sleep(time.Second)
		_, err = rscClient.Create(ctx, destObj, metav1.CreateOptions{FieldManager: FieldManager})
		if err != nil {
			logger.Error(err, "Failed to create object in mailbox workspace")
			return true
		}
		logger.V(3).Info("Created object in mailbox workspace")
		return false
	}
}

func (wp *workloadProjector) trimForDestination(srcObj mrObject) *unstructured.Unstructured {
	srcObjU := srcObj.(*unstructured.Unstructured)
	srcObjU = srcObjU.DeepCopy()
	destObjR := srcObjU.NewEmptyInstance()
	destObj := destObjR.(*unstructured.Unstructured)
	destObj.SetUnstructuredContent(srcObjU.UnstructuredContent())
	destObj.SetManagedFields([]metav1.ManagedFieldsEntry{})
	destObj.SetOwnerReferences([]metav1.OwnerReference{}) // we do not transport owner UIDs
	destObj.SetResourceVersion("")
	destObj.SetSelfLink("")
	destObj.SetUID("")
	destObj.SetZZZ_DeprecatedClusterName("")
	return destObj
}

func (wp *workloadProjector) genericObjectMerge(sourceCluster logicalcluster.Name, destSP SinglePlacement,
	srcObj mrObject, inputDest *unstructured.Unstructured) *unstructured.Unstructured {
	logger := klog.FromContext(wp.ctx).WithValues(
		"sourceCluster", sourceCluster,
		"destSP", destSP,
		"destGVK", inputDest.GroupVersionKind,
		"namespace", srcObj.GetNamespace(),
		"name", srcObj.GetName())
	srcObjU := srcObj.(*unstructured.Unstructured)
	outputDestR := inputDest.NewEmptyInstance()
	outputDestU := outputDestR.(*unstructured.Unstructured)
	inputDest = inputDest.DeepCopy() // because the following only swings the top-level pointer
	outputDestU.SetUnstructuredContent(inputDest.UnstructuredContent())
	kvMerge := func(which string, src, inputDest map[string]string) map[string]string {
		outputDest := map[string]string{}
		for key, val := range inputDest {
			outputDest[key] = val
		}
		for key, val := range src {
			if kvIsSystem(which, key) {
				continue
			}
			if oval, have := outputDest[key]; have && oval != val {
				logger.Info("Overwriting key/val", "which", which, "key", key, "oldVal", oval, "newVal", val)
			}
			outputDest[key] = val
		}
		return outputDest
	}
	if len(srcObj.GetAnnotations()) != 0 { // If nothing to merge then do not gratuitously change absent to empty map.
		outputDestU.SetAnnotations(kvMerge("annotations", srcObj.GetAnnotations(), inputDest.GetAnnotations()))
	}
	if len(srcObj.GetLabels()) != 0 { // If nothing to merge then do not gratuitously change absent to empty map.
		outputDestU.SetLabels(kvMerge("labels", srcObj.GetLabels(), inputDest.GetLabels()))
	}
	destContent := outputDestU.UnstructuredContent()
	srcContent := srcObjU.UnstructuredContent()
	for topKey, srcTopVal := range srcContent {
		switch topKey {
		case "apiVersion", "kind", "metadata", "status":
			continue
		default:
		}
		destContent[topKey] = srcTopVal
	}
	outputDestU.SetUnstructuredContent(destContent)
	return outputDestU
}

func kvIsSystem(which, key string) bool {
	return (strings.Contains(key, ".kcp.io/") || strings.HasPrefix(key, "kcp.io/")) && !strings.Contains(key, "edge.kcp.io/")
}

func (wp *workloadProjector) Transact(xn func(WorkloadProjectionSections)) {
	logger := klog.FromContext(wp.ctx)
	wp.Lock()
	defer wp.Unlock()
	logger.V(3).Info("Begin transaction")
	changedDestinations := func() *MutableSet[SinglePlacement] {
		var ms MutableSet[SinglePlacement] = NewMapSet[SinglePlacement]()
		ms = WrapSetWithMutex(ms)
		return &ms
	}()
	wp.changedDestinations = changedDestinations
	recordLogger := logger.V(4)
	changedSources := WrapSetWithMutex[logicalcluster.Name](NewMapSet[logicalcluster.Name]())
	xn(WorkloadProjectionSections{
		SetWriterFork[NamespaceDistributionTuple](false,
			wp.nsDistributionsForSync,
			wp.nsDistributionsForProj,
			recordPart(recordLogger, "nsd.src", changedDestinations, TripleFactorerTo3and21[logicalcluster.Name, NamespaceName, SinglePlacement]()),
			recordPart(recordLogger, "nsd.dest", &changedSources, TripleFactorerTo1and23[logicalcluster.Name, NamespaceName, SinglePlacement]())),
		SetWriterFork[NamespacedResourceDistributionTuple](false,
			wp.nsrDistributionsForSync, wp.nsrDistributionsForProj,
			recordPart(recordLogger, "nsrd.src", changedDestinations, factorNamespacedResourceDistributionTupleForSync1),
			recordPart(recordLogger, "nsrc.dest", &changedSources, factorNamespacedResourceDistributionTupleForProj1)),
		NewMappingReceiverFork[ProjectionModeKey, ProjectionModeVal](wp.nsModesForSync, wp.nsModesForProj),
		SetWriterFork[NonNamespacedDistributionTuple](false,
			wp.nnsDistributionsForSync, wp.nnsDistributionsForProj,
			recordPart(recordLogger, "nns.src", changedDestinations, factorNonNamespacedDistributionTupleForSync1),
			recordPart(recordLogger, "nns.dest", &changedSources, factorNonNamespacedDistributionTupleForProj1)),
		NewMappingReceiverFork[ProjectionModeKey, ProjectionModeVal](wp.nnsModesForSync, wp.nnsModesForProj),
		wp.upsyncs})
	logger.V(3).Info("Transaction response",
		"changedDestinations", VisitableToSlice[SinglePlacement](*changedDestinations),
		"changedSources", VisitableToSlice[logicalcluster.Name](changedSources))
	changedSources.Visit(func(source logicalcluster.Name) error {
		wps, have := wp.perSource.Get(source)
		logger := logger.WithValues("source", source)
		if !have {
			logger.Error(nil, "No per-source data for changed source")
			return nil
		}
		logger.V(4).Info("Finishing transaction wrt source",
			"nsDistributions", VisitableToSlice[Pair[NamespaceName, SinglePlacement]](wps.nsDistributions),
			"nsrDistributions", VisitableToSlice[Pair[metav1.GroupResource, SinglePlacement]](wps.nsrDistributions),
			"nnsDistributions", VisitableToSlice[Triple[metav1.GroupResource, string, SinglePlacement]](wps.nnsDistributions))
		wps.nsrDistributions.GetIndex1to2().Visit(func(tup Pair[metav1.GroupResource, Set[SinglePlacement]]) error {
			gr := tup.First
			logger := logger.WithValues("groupResource", gr)
			problem, have := wp.nsModesForProj.GetIndex().Get(gr)
			if !have {
				logger.Error(nil, "No projection mode")
				return nil
			}
			solve := pickThe1[metav1.GroupResource, SinglePlacement](logger, "eeek")
			pmv := solve(gr, problem)
			logger = logger.WithValues("apiVersion", pmv.APIVersion)
			preInformer, have := wps.preInformers.Get(gr)
			if !have {
				logger.V(4).Info("Instantiating new informer for namespaced resource")
				sgvr := MetaGroupResourceToSchema(gr).WithVersion(pmv.APIVersion)
				preInformer = wps.dynamicInformerFactory.ForResource(sgvr)
				wps.preInformers.Put(gr, preInformer)
				preInformer.Informer().AddEventHandler(k8scache.ResourceEventHandlerFuncs{
					AddFunc:    func(obj any) { wps.enqueueSourceNamespacedObject(gr, obj, "add") },
					UpdateFunc: func(oldObj, newObj any) { wps.enqueueSourceNamespacedObject(gr, newObj, "update") },
					DeleteFunc: func(obj any) { wps.enqueueSourceNamespacedObject(gr, obj, "delete") },
				})
				go preInformer.Informer().Run(wp.ctx.Done()) // TODO: just once per resource
				time.Sleep(wp.delay)
			}
			return nil
		})
		wps.nnsDistributions.GetIndex1to2().Visit(func(tup Pair[metav1.GroupResource, ObjectNameToDestinations]) error {
			gr := tup.First
			logger := logger.WithValues("groupResource", gr)
			problem, have := wp.nnsModesForProj.GetIndex().Get(gr)
			if !have {
				logger.Error(nil, "No projection mode")
				return nil
			}
			solve := pickThe1[metav1.GroupResource, SinglePlacement](logger, "eeek")
			pmv := solve(gr, problem)
			logger = logger.WithValues("apiVersion", pmv.APIVersion)
			preInformer, have := wps.preInformers.Get(gr)
			if !have {
				logger.V(4).Info("Instantiating new informer for cluster-scoped resource")
				sgvr := MetaGroupResourceToSchema(gr).WithVersion(pmv.APIVersion)
				preInformer = wps.dynamicInformerFactory.ForResource(sgvr)
				wps.preInformers.Put(gr, preInformer)
				preInformer.Informer().AddEventHandler(k8scache.ResourceEventHandlerFuncs{
					AddFunc:    func(obj any) { wps.enqueueSourceNonNamespacedObject(gr, obj, "add") },
					UpdateFunc: func(oldObj, newObj any) { wps.enqueueSourceNonNamespacedObject(gr, newObj, "update") },
					DeleteFunc: func(obj any) { wps.enqueueSourceNonNamespacedObject(gr, obj, "delete") },
				})
				go preInformer.Informer().Run(wp.ctx.Done()) // TODO: just once per resource
				time.Sleep(wp.delay)
			}

			return nil
		})
		return nil
	})
	(*changedDestinations).Visit(func(destination SinglePlacement) error {
		mbwsName := SPMailboxWorkspaceName(destination)
		wp.mbwsNameToSP.Put(mbwsName, destination)
		nsds, have := wp.nsDistributionsForSync.GetIndex1to2().Get(destination)
		if have {
			nses := MapKeySet[NamespaceName, Set[logicalcluster.Name]](nsds.GetIndex1to2())
			logger.V(4).Info("Namespaces after transaction", "destination", destination, "namespaces", VisitableToSlice[NamespaceName](nses))
		} else {
			logger.V(4).Info("No Namespaces after transaction", "destination", destination)
		}
		nsrds, have := wp.nsrDistributionsForSync.GetIndex1to2().Get(destination)
		if have {
			nsrs := MapKeySet[metav1.GroupResource, Set[logicalcluster.Name]](nsrds.GetIndex1to2())
			logger.V(4).Info("NamespacedResources after transaction", "destination", destination, "resources", VisitableToSlice[metav1.GroupResource](nsrs))
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
			logger.V(4).Info("NonNamespacedResources after transaction", "destination", destination, "objs", VisitableToSlice[GroupResourceInstance](objs))
		} else {
			logger.V(4).Info("No NonNamespacedResources after transaction", "destination", destination)
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
		mbwsCluster, ok := wp.mbwsNameToCluster.Get(mbwsName)
		if !ok {
			logger.Error(nil, "Mailbox workspace not known yet", "destination", destination)
			wp.queue.Add(destination)
		} else {
			scRef := syncerConfigRef{mbwsCluster, SyncerConfigName}
			logger.V(4).Info("Enqueuing reference to SyncerConfig affected by transaction", "destination", destination, "mbwsName", mbwsName, "scRef", scRef)
			wp.queue.Add(scRef)
		}
		return nil
	})
	logger.V(3).Info("End transaction")
	wp.changedDestinations = nil
}

func (wps *wpPerSource) enqueueSourceNamespacedObject(gr metav1.GroupResource, obj any, action string) {
	dfu, ok := obj.(k8scache.DeletedFinalStateUnknown)
	if ok {
		obj = dfu.Obj
	}
	objm := obj.(metav1.Object)
	if ObjectIsSystem(objm) {
		wps.logger.V(4).Info("Ignoring system object", "groupResource", gr, "namespace", objm.GetNamespace(), "name", objm.GetName())
		return
	}
	ref := sourceObjectRef{wps.source, gr, objm.GetNamespace(), objm.GetName()}
	wps.logger.V(4).Info("Enqueuing reference to source namespaced object", "ref", ref)
	wps.wp.queue.Add(ref)
}

func (wps *wpPerSource) enqueueSourceNonNamespacedObject(gr metav1.GroupResource, obj any, action string) {
	dfu, ok := obj.(k8scache.DeletedFinalStateUnknown)
	if ok {
		obj = dfu.Obj
	}
	objm := obj.(metav1.Object)
	if ObjectIsSystem(objm) {
		wps.logger.V(4).Info("Ignoring system object", "groupResource", gr, "namespace", objm.GetNamespace(), "name", objm.GetName())
		return
	}
	ref := sourceObjectRef{wps.source, gr, noNamespace, objm.GetName()}
	wps.logger.V(4).Info("Enqueuing reference to source cluster-scoped object", "ref", ref)
	wps.wp.queue.Add(ref)
}

func ObjectIsSystem(objm metav1.Object) bool {
	obju := objm.(*unstructured.Unstructured)
	objt := objm.(metav1.Type)
	apiVersion := objt.GetAPIVersion()
	kind := objt.GetKind()
	if apiVersion != "v1" {
		return false
	}
	switch kind {
	case "Secret":
		secretType := obju.UnstructuredContent()["type"]
		return secretType == "kubernetes.io/service-account-token" ||
			secretType == "bootstrap.kubernetes.io/token"
	case "ConfigMap":
		return objm.GetName() == "kube-root-ca.crt"
	case "ServiceAccount":
		return objm.GetName() == "default"
	default:
		return false
	}
}

func recordPart[Whole, Part, Rest any](logger klog.Logger, partType string, record *MutableSet[Part], factorer Factorer[Whole, Part, Rest]) SetWriter[Whole] {
	return SetWriterFuncs[Whole]{
		OnAdd: func(whole Whole) bool {
			parts := factorer.First(whole)
			news := (*record).Add(parts.First)
			logger.Info("Recorded subject of Add", "news", news, "partType", partType, "partVal", parts.First, "rest", parts.Second, "revisedSet", *record)
			return true
		},
		OnRemove: func(whole Whole) bool {
			parts := factorer.First(whole)
			news := (*record).Add(parts.First)
			logger.Info("Recorded subject of Remove", "news", news, "partType", partType, "partVal", parts.First, "rest", parts.Second, "revisedSet", *record)
			return true
		}}
}

var factorNamespacedResourceDistributionTupleForProj1 = NewFactorer(
	func(whole NamespacedResourceDistributionTuple) Pair[logicalcluster.Name, Pair[metav1.GroupResource, SinglePlacement]] {
		return NewPair(whole.SourceCluster, NewPair(whole.GroupResource, whole.Destination))
	},
	func(parts Pair[logicalcluster.Name, Pair[metav1.GroupResource, SinglePlacement]]) NamespacedResourceDistributionTuple {
		return NamespacedResourceDistributionTuple{parts.First, ProjectionModeKey{parts.Second.First, parts.Second.Second}}
	})

var factorNamespacedResourceDistributionTupleForSync1 = NewFactorer(
	func(whole NamespacedResourceDistributionTuple) Pair[SinglePlacement, Pair[metav1.GroupResource, logicalcluster.Name]] {
		return NewPair(whole.Destination, NewPair(whole.GroupResource, whole.SourceCluster))
	},
	func(parts Pair[SinglePlacement, Pair[metav1.GroupResource, logicalcluster.Name]]) NamespacedResourceDistributionTuple {
		return NamespacedResourceDistributionTuple{parts.Second.Second, ProjectionModeKey{parts.Second.First, parts.First}}
	})

var factorNonNamespacedDistributionTupleForSync1 = NewFactorer(
	func(whole NonNamespacedDistributionTuple) Pair[SinglePlacement, Pair[GroupResourceInstance, logicalcluster.Name]] {
		return NewPair(whole.First.Destination, NewPair(NewPair(whole.First.GroupResource, whole.Second.Name), whole.Second.Cluster))
	},
	func(parts Pair[SinglePlacement, Pair[GroupResourceInstance, logicalcluster.Name]]) NonNamespacedDistributionTuple {
		return NewPair(ProjectionModeKey{parts.Second.First.First, parts.First},
			ExternalName{parts.Second.Second, parts.Second.First.Second})
	})

var factorNonNamespacedDistributionTupleForProj1 = NewFactorer(
	func(whole NonNamespacedDistributionTuple) Pair[logicalcluster.Name, Pair[GroupResourceInstance, SinglePlacement]] {
		return NewPair(whole.Second.Cluster, NewPair(NewPair(whole.First.GroupResource, whole.Second.Name), whole.First.Destination))
	},
	func(parts Pair[logicalcluster.Name, Pair[GroupResourceInstance, SinglePlacement]]) NonNamespacedDistributionTuple {
		return NewPair(ProjectionModeKey{parts.Second.First.First, parts.Second.Second},
			ExternalName{parts.First, parts.Second.First.Second})
	})

var factorNonNamespacedDistributionTupleForProj1and234 = NewFactorer(
	func(whole NonNamespacedDistributionTuple) Pair[logicalcluster.Name, Triple[metav1.GroupResource, string /*obj name*/, SinglePlacement]] {
		return NewPair(whole.Second.Cluster, NewTriple(whole.First.GroupResource, whole.Second.Name, whole.First.Destination))
	},
	func(parts Pair[logicalcluster.Name, Triple[metav1.GroupResource, string /*obj name*/, SinglePlacement]]) NonNamespacedDistributionTuple {
		return NewPair(ProjectionModeKey{parts.Second.First, parts.Second.Third},
			ExternalName{parts.First, parts.Second.Second})
	})

var factorProjectionModeKeyForSyncer = NewFactorer(
	func(pmk ProjectionModeKey) Pair[SinglePlacement, metav1.GroupResource] {
		return NewPair(pmk.Destination, pmk.GroupResource)
	},
	func(tup Pair[SinglePlacement, metav1.GroupResource]) ProjectionModeKey {
		return ProjectionModeKey{Destination: tup.First, GroupResource: tup.Second}
	})

var factorProjectionModeKeyForProj = NewFactorer(
	func(pmk ProjectionModeKey) Pair[metav1.GroupResource, SinglePlacement] {
		return NewPair(pmk.GroupResource, pmk.Destination)
	},
	func(tup Pair[metav1.GroupResource, SinglePlacement]) ProjectionModeKey {
		return ProjectionModeKey{Destination: tup.Second, GroupResource: tup.First}
	})

// syncerConfigSpecRelations is a relational represetntation of SyncerConfigSpec.
// It takes O(N) to construct and O(N) to compare.
type syncerConfigSpecRelations struct {
	namespaces           Set[string]
	namespacedResources  Set[edgeapi.NamespaceScopeDownsyncResource]
	clusterScopedObjects MutableMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string /*object name*/]]]
	upsyncs              Set[edgeapi.UpsyncSet]
}

func (wp *workloadProjector) syncerConfigRelations(destination SinglePlacement) syncerConfigSpecRelations {
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
	ans.upsyncs = HashSetCopy[edgeapi.UpsyncSet](HashUpsyncSet{})(upsyncs)
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

func (wp *workloadProjector) syncerConfigIsGood(destination SinglePlacement, configRef ExternalName, syncfg *edgeapi.SyncerConfig, goodSpecRelations syncerConfigSpecRelations) bool {
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
