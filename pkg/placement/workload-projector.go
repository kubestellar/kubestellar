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
	"k8s.io/apimachinery/pkg/labels"
	machruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
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
	schedulingv1a1 "github.com/kcp-dev/kcp/pkg/apis/scheduling/v1alpha1"
	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	schedulingv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/scheduling/v1alpha1"
	tenancyv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v1alpha1"
	edgeclusterclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
	edgev1a1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/customize"
)

const SyncerConfigName = "the-one"
const FieldManager = "placement-translator"

// This workload projector:
// (a) maintains SyncerConfig objects in mailbox workspaces, and
// (b) propagates changes from source workspaces to mailbox workspaces.
//
// For a given mailbox workspace, for every resource that the WP is
// projecting to that MBWS, the WP has an informer on that resource
// and reacts to the presence of objects previously projected that
// should not now be projected by deleting that object.
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
	resourceModes ResourceModes,
	mbwsInformer k8scache.SharedIndexInformer,
	mbwsLister tenancyv1a1listers.WorkspaceLister,
	locationClusterInformer kcpcache.ScopeableSharedIndexInformer,
	locationClusterLister schedulingv1a1listers.LocationClusterLister,
	syncfgClusterInformer kcpcache.ScopeableSharedIndexInformer,
	syncfgClusterLister edgev1a1listers.SyncerConfigClusterLister,
	customizerClusterInformer kcpcache.ScopeableSharedIndexInformer,
	customizerClusterLister edgev1a1listers.CustomizerClusterLister,
	edgeClusterClientset edgeclusterclientset.ClusterInterface,
	dynamicClusterClient clusterdynamic.ClusterInterface,
	nsClusterPreInformer kcpkubecorev1informers.NamespaceClusterInformer,
	nsClusterClient kcpkubecorev1client.NamespaceClusterInterface,
) *workloadProjector {
	wp := &workloadProjector{
		// delay:                 2 * time.Second,
		ctx:                       ctx,
		configConcurrency:         configConcurrency,
		resourceModes:             resourceModes,
		queue:                     workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		mbwsLister:                mbwsLister,
		locationClusterInformer:   locationClusterInformer,
		locationClusterLister:     locationClusterLister,
		syncfgClusterInformer:     syncfgClusterInformer,
		syncfgClusterLister:       syncfgClusterLister,
		customizerClusterInformer: customizerClusterInformer,
		customizerClusterLister:   customizerClusterLister,
		edgeClusterClientset:      edgeClusterClientset,
		dynamicClusterClient:      dynamicClusterClient,
		nsClusterPreInformer:      nsClusterPreInformer,
		nsClusterClient:           nsClusterClient,

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
	ctx                       context.Context
	configConcurrency         int
	resourceModes             ResourceModes
	delay                     time.Duration // to slow down for debugging
	queue                     workqueue.RateLimitingInterface
	mbwsLister                tenancyv1a1listers.WorkspaceLister
	locationClusterInformer   kcpcache.ScopeableSharedIndexInformer
	locationClusterLister     schedulingv1a1listers.LocationClusterLister
	syncfgClusterInformer     kcpcache.ScopeableSharedIndexInformer
	syncfgClusterLister       edgev1a1listers.SyncerConfigClusterLister
	customizerClusterInformer kcpcache.ScopeableSharedIndexInformer
	customizerClusterLister   edgev1a1listers.CustomizerClusterLister
	edgeClusterClientset      edgeclusterclientset.ClusterInterface
	dynamicClusterClient      clusterdynamic.ClusterInterface
	nsClusterPreInformer      kcpkubecorev1informers.NamespaceClusterInformer
	nsClusterClient           kcpkubecorev1client.NamespaceClusterInterface

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
		logger:           klog.FromContext(wp.ctx).WithValues("destination", destination),
		nsDistributions:  NewMapRelation2[NamespaceName, logicalcluster.Name](),
		nsrDistributions: NewMapRelation2[metav1.GroupResource, logicalcluster.Name](),
		nnsDistributions: NewMapRelation2[GroupResourceInstance, logicalcluster.Name](),
		preInformers:     NewMapMap[metav1.GroupResource, dynamicDuo](nil),
	}
	return wpd
}

// The data structure specific to a mailbox/edge-cluster.
// All the variable fields must be accessed with the wp mutex locked.
// The readyChan is closed once the namespace informer has synced.
type wpPerDestination struct {
	wp               *workloadProjector
	destination      SinglePlacement
	logger           klog.Logger
	nsDistributions  SingleIndexedRelation2[NamespaceName, logicalcluster.Name]
	nsrDistributions SingleIndexedRelation2[metav1.GroupResource, logicalcluster.Name]
	nnsDistributions SingleIndexedRelation2[GroupResourceInstance, logicalcluster.Name]

	namespaceClient      k8scorev1client.NamespaceInterface
	namespacePreInformer k8scorev1informers.NamespaceInformer
	nsReadyChan          <-chan struct{}

	dynamicClient          k8sdynamic.Interface
	dynamicInformerFactory k8sdynamicinformer.DynamicSharedInformerFactory
	preInformers           MutableMap[metav1.GroupResource, dynamicDuo]
}

type dynamicDuo struct {
	apiVersion  string
	namespaced  bool
	preInformer upstreaminformers.GenericInformer // nil iff resource is namespaces
	client      k8sdynamic.NamespaceableResourceInterface
}

func (wpd *wpPerDestination) getDynamicDuoLocked(gr metav1.GroupResource, apiVersion string, namespaced bool) (dynamicDuo, <-chan struct{}, error) {
	if wpd.dynamicClient == nil {
		mbwsName := SPMailboxWorkspaceName(wpd.destination)
		mbwsCluster, have := wpd.wp.mbwsNameToCluster.Get(mbwsName)
		if !have {
			return dynamicDuo{}, nil, errors.New("unable to map mailbox workspace name to cluster")
		}
		wpd.dynamicClient = wpd.wp.dynamicClusterClient.Cluster(mbwsCluster.Path())
		justMine, err := labels.NewRequirement(ProjectedLabelKey, selection.Equals, []string{ProjectedLabelVal})
		if err != nil {
			panic(err)
		}
		justMineStr := justMine.String()
		wpd.dynamicInformerFactory = k8sdynamicinformer.NewFilteredDynamicSharedInformerFactory(wpd.dynamicClient, 0,
			metav1.NamespaceAll, func(opts *metav1.ListOptions) {
				if opts.LabelSelector == "" {
					opts.LabelSelector = justMineStr
				} else {
					opts.LabelSelector = opts.LabelSelector + "," + justMineStr
				}
			})
		wpd.namespaceClient = wpd.wp.nsClusterClient.Cluster(mbwsCluster.Path())
		wpd.namespacePreInformer = wpd.wp.nsClusterPreInformer.Cluster(mbwsCluster)
		nsInformer := wpd.namespacePreInformer.Informer()
		nsReadyChan := make(chan struct{})
		wpd.nsReadyChan = nsReadyChan
		go func() {
			k8scache.WaitForNamedCacheSync("workload-projector", wpd.wp.ctx.Done(), nsInformer.HasSynced)
			close(nsReadyChan)
		}()
		go nsInformer.Run(wpd.wp.ctx.Done())
		wpd.dynamicInformerFactory.Start(wpd.wp.ctx.Done())
	}
	duo, have := wpd.preInformers.Get(gr)
	if have {
		if apiVersion != duo.apiVersion || namespaced != duo.namespaced {
			wpd.logger.Error(nil, "Not implemented yet: changing version or namespaced of GroupResource", "groupResource", gr,
				"oldVersion", duo.apiVersion, "newVersion", apiVersion,
				"oldNamespaced", duo.namespaced, "newNamespaced", namespaced)
			// TODO: implement
		}
	} else {
		sgvr := MetaGroupResourceToSchema(gr).WithVersion(apiVersion)
		wpd.logger.V(4).Info("Creating informer at destination", "groupResource", gr, "apiVersion", apiVersion, "namespaced", namespaced)
		duo = dynamicDuo{
			apiVersion: apiVersion,
			namespaced: namespaced,
			client:     wpd.dynamicClient.Resource(sgvr)}
		if mgrIsNamespace(gr) {
			// No way to know if a namespace is needed for other reasons,
			// so no point in reacting to them.
		} else {
			duo.preInformer = wpd.dynamicInformerFactory.ForResource(sgvr)
			duo.preInformer.Informer().AddEventHandler(k8scache.ResourceEventHandlerFuncs{
				AddFunc:    func(obj any) { wpd.enqueueDestinationObject(gr, namespaced, obj, "add") },
				UpdateFunc: func(oldObj, newObj any) { wpd.enqueueDestinationObject(gr, namespaced, newObj, "update") },
				DeleteFunc: func(obj any) { wpd.enqueueDestinationObject(gr, namespaced, obj, "delete") }})
			go duo.preInformer.Informer().Run(wpd.wp.ctx.Done())

		}
		wpd.preInformers.Put(gr, duo)
	}
	return duo, wpd.nsReadyChan, nil
}

func (wpd *wpPerDestination) resyncGroupResource(gr metav1.GroupResource, duo dynamicDuo) {
	objs := duo.preInformer.Informer().GetStore().List()
	for _, obj := range objs {
		wpd.enqueueDestinationObject(gr, duo.namespaced, obj, "resync")
	}
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
		preInformers:           NewMapMap[metav1.GroupResource, nsdPreInformer](nil),
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
	preInformers           MutableMap[metav1.GroupResource, nsdPreInformer]
}

type nsdPreInformer struct {
	namespaced  bool
	preInformer upstreaminformers.GenericInformer
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

// destinationObjectRef refers to an namespaced object in a mailbox workspace
type destinationObjectRef struct {
	destination   edgeapi.SinglePlacement
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
	case destinationObjectRef:
		retry = wp.syncDestinationObject(ctx, typed)
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

func (wp *workloadProjector) syncDestinationObject(ctx context.Context, doRef destinationObjectRef) bool {
	namespaced := doRef.namespace != noNamespace
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("objectRef", doRef)
	finish := func() func() bool {
		wp.Lock()
		defer wp.Unlock()
		wpd, have := wp.perDestination.Get(doRef.destination)
		if !have {
			logger.V(4).Info("wp.perDestination.Get said no")
			return returnFalse
		}
		duo, have := wpd.preInformers.Get(doRef.groupResource)
		if !have {
			logger.V(4).Info("No local informer")
			return returnFalse
		}
		lister := duo.preInformer.Lister()
		var nsl k8scache.GenericNamespaceLister = lister
		if namespaced {
			nsl = lister.ByNamespace(doRef.namespace)
		}
		obj, err := nsl.Get(doRef.name)
		present := err == nil && obj != nil
		var objM metav1.Object
		if present {
			objM = obj.(metav1.Object)
			if objM.GetDeletionTimestamp() != nil {
				present = false
			}
		}
		var sources Set[logicalcluster.Name]
		var haveSources bool
		if namespaced {
			sourcesForGR, foundGR := wpd.nsrDistributions.GetIndex1to2().Get(doRef.groupResource)
			sourcesForNS, foundNS := wpd.nsDistributions.GetIndex1to2().Get(NamespaceName(doRef.namespace))
			var sources Set[logicalcluster.Name] = NewEmptyMapSet[logicalcluster.Name]()
			if foundGR && foundNS {
				sources = SetIntersection(sourcesForGR, sourcesForNS)
			}
			if !sources.IsEmpty() {
				logger.V(4).Info("Retaining namespaced destination object", "sourcesForGR", VisitableToSlice[logicalcluster.Name](sourcesForGR), "sourcesForNS", VisitableToSlice[logicalcluster.Name](sourcesForNS))
				return returnFalse
			}
		} else {
			sources, haveSources = wpd.nnsDistributions.GetIndex1to2().Get(NewPair(doRef.groupResource, doRef.name))
			if haveSources && !sources.IsEmpty() {
				logger.V(4).Info("Retaining cluster-scoped destination object", "sources", VisitableToSlice[logicalcluster.Name](sources))
				return returnFalse
			}
		}
		if !present {
			logger.V(4).Info("Undesired destination object is already absent", "err", err, "obj", obj)
			return returnFalse
		}
		resourceVersion := objM.GetResourceVersion()
		var rscClient k8sdynamic.ResourceInterface = duo.client
		if namespaced {
			rscClient = duo.client.Namespace(doRef.namespace)
		}
		return func() bool {
			err := rscClient.Delete(ctx, doRef.name,
				metav1.DeleteOptions{Preconditions: &metav1.Preconditions{ResourceVersion: &resourceVersion}})
			if err == nil {
				logger.V(3).Info("Deleted undesired object in mailbox workspace", "resourceVersion", resourceVersion)
			} else if k8sapierrors.IsNotFound(err) {
				logger.V(4).Info("Undesired object in mailbox workspace was deleted concurrently", "resourceVersion", resourceVersion)
			} else {
				logger.Error(err, "Failed to delete unwanted object in mailbox workspace", "resourceVersion", resourceVersion)
				return true
			}
			return false
		}
	}()
	return finish()
}

// Returns `retry bool`.
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
		npi, have := wps.preInformers.Get(soRef.groupResource)
		if !have {
			logger.Error(nil, "Impossible: handling source object of unknown resource")
			return []func() bool{returnFalse}
		}
		var srcObj machruntime.Object
		var err error
		if namespaced {
			srcObj, err = npi.preInformer.Lister().ByNamespace(soRef.namespace).Get(soRef.name)
		} else {
			srcObj, err = npi.preInformer.Lister().Get(soRef.name)
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
				logger.V(4).Info("No objects of this source and cluster-sccoped kind are going anywhere")
				return []func() bool{returnFalse}
			}
			destinations, haveDestinations = byName.GetIndex1to2().Get(soRef.name)
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
			retryThis, rem := wp.syncSourceToDestLocked(ctx, logger, soRef, srcMRObject, namespaced, deleted, modesForSync, destination)
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

// Sync a source object to one MBWS.
// Returns `(retry bool, unlocked func() (retry bool))`
func (wp *workloadProjector) syncSourceToDestLocked(ctx context.Context, logger klog.Logger,
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
	duo, clientReadyChan, err := wpd.getDynamicDuoLocked(soRef.groupResource, pmv.APIVersion, namespaced)
	if err != nil {
		logger.Error(err, "Failed to wpd.getDynamicDuoLocked")
		return true, nil
	}
	return false, func() bool {
		// sgvr := MetaGroupResourceToSchema(soRef.groupResource).WithVersion(pmv.APIVersion)
		var rscClient k8sdynamic.ResourceInterface = duo.client
		if namespaced {
			rscClient = duo.client.Namespace(soRef.namespace)
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
			if err != nil {
				if !k8sapierrors.IsNotFound(err) {
					logger.Error(err, "Failed to lookup namespace in local cache")
					return true
				}
				nsObj = nil
			}
			if nsObj == nil {
				nsObj = &k8scorev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name:   soRef.namespace,
						Labels: map[string]string{ProjectedLabelKey: ProjectedLabelVal},
					}}
				_, err := wpd.namespaceClient.Create(ctx, nsObj, metav1.CreateOptions{FieldManager: FieldManager})
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
			asUpdated, err := rscClient.Update(ctx, revisedDestObj, metav1.UpdateOptions{FieldManager: FieldManager})
			if err != nil {
				logger.Error(err, "Failed to update object in mailbox workspace", "resourceVersion", asUpdated.GetResourceVersion())
				return true
			}
			if logger.V(5).Enabled() {
				logger = logger.WithValues("oldVal", destObj, "newVal", revisedDestObj)
			}
			logger.V(3).Info("Updated object in mailbox workspace",
				"oldResourceVersion", revisedDestObj.GetResourceVersion(),
				"newResourceVersion", asUpdated.GetResourceVersion())
			return false
		}
		destObj = wpd.wp.xformForDestination(soRef.cluster, destination, srcMRObject)
		time.Sleep(time.Second)
		asCreated, err := rscClient.Create(ctx, destObj, metav1.CreateOptions{FieldManager: FieldManager})
		if err != nil {
			logger.Error(err, "Failed to create object in mailbox workspace")
			return true
		}
		logger.V(3).Info("Created object in mailbox workspace", "resourceVersion", asCreated.GetResourceVersion())
		return false
	}
}

const ProjectedLabelKey string = "edge.kcp.io/projected"
const ProjectedLabelVal string = "yes"

func (wp *workloadProjector) xformForDestination(sourceCluster logicalcluster.Name, destSP SinglePlacement, srcObj mrObject) *unstructured.Unstructured {
	srcObjU := srcObj.(*unstructured.Unstructured)
	logger := klog.FromContext(wp.ctx).WithValues(
		"sourceCluster", sourceCluster,
		"destSP", destSP,
		"destGVK", srcObjU.GroupVersionKind,
		"namespace", srcObj.GetNamespace(),
		"name", srcObj.GetName())
	srcObjU = wp.customizeOrCopy(logger, sourceCluster, srcObjU, destSP, true)
	destObjR := srcObjU.NewEmptyInstance()
	destObj := destObjR.(*unstructured.Unstructured)
	// customize.Customize(wp.ctx, srcObjU.UnstructuredContent(), customizer, log)
	destObj.SetUnstructuredContent(srcObjU.UnstructuredContent())
	destObj.SetManagedFields([]metav1.ManagedFieldsEntry{})
	destObj.SetOwnerReferences([]metav1.OwnerReference{}) // we do not transport owner UIDs
	destObj.SetResourceVersion("")
	destObj.SetSelfLink("")
	destObj.SetUID("")
	destObj.SetZZZ_DeprecatedClusterName("")
	labels := destObj.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[ProjectedLabelKey] = ProjectedLabelVal
	destObj.SetLabels(labels)
	return destObj
}

func (wp *workloadProjector) genericObjectMerge(sourceCluster logicalcluster.Name, destSP SinglePlacement,
	srcObj mrObject, inputDest *unstructured.Unstructured) *unstructured.Unstructured {
	srcObjU := srcObj.(*unstructured.Unstructured)
	logger := klog.FromContext(wp.ctx).WithValues(
		"sourceCluster", sourceCluster,
		"destSP", destSP,
		"destGVK", srcObjU.GroupVersionKind,
		"namespace", srcObj.GetNamespace(),
		"name", srcObj.GetName())
	srcObjU = wp.customizeOrCopy(logger, sourceCluster, srcObjU, destSP, false)
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
	if len(srcObjU.GetAnnotations()) != 0 { // If nothing to merge then do not gratuitously change absent to empty map.
		outputDestU.SetAnnotations(kvMerge("annotations", srcObjU.GetAnnotations(), inputDest.GetAnnotations()))
	}
	mergedLabels := kvMerge("labels", srcObjU.GetLabels(), inputDest.GetLabels())
	mergedLabels[ProjectedLabelKey] = ProjectedLabelVal
	outputDestU.SetLabels(mergedLabels)
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

func (wp *workloadProjector) customizeOrCopy(logger klog.Logger, srcCluster logicalcluster.Name, srcObjU *unstructured.Unstructured, destSP edgeapi.SinglePlacement, insistCopy bool) *unstructured.Unstructured {
	srcAnnotations := srcObjU.GetAnnotations()
	expandParameters := srcAnnotations[edgeapi.ParameterExpansionAnnotationKey] == "true"
	customizerRef := srcAnnotations[edgeapi.CustomizerAnnotationKey]
	var customizer *edgeapi.Customizer
	var err error
	if len(customizerRef) != 0 {
		refParts := strings.SplitN(customizerRef, "/", 2)
		custNS := refParts[0]
		custName := refParts[len(refParts)-1]
		if len(refParts) == 1 {
			custNS = srcObjU.GetNamespace()
		}
		customizer, err = wp.customizerClusterLister.Cluster(logicalcluster.Name(srcCluster)).Customizers(custNS).Get(custName)
		if err != nil {
			logger.Error(err, "Failed to find referenced Customizer")
		} else {
			expandParameters = expandParameters || customizer.Annotations[edgeapi.ParameterExpansionAnnotationKey] == "true"
		}
	}
	var location *schedulingv1a1.Location
	if expandParameters {
		location, err = wp.locationClusterLister.Cluster(logicalcluster.Name(destSP.Cluster)).Get(destSP.LocationName)
		if err != nil {
			logger.Error(err, "Failed to find referenced Location")
		}
	}
	if (len(customizerRef) != 0 || expandParameters) &&
		(customizer != nil || len(customizerRef) == 0) &&
		(location != nil || !expandParameters) {
		return customize.Customize(logger, srcObjU, customizer, location)
	}
	if !insistCopy {
		return srcObjU
	}
	return srcObjU.DeepCopy()
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
		wps.preInformers.Visit(func(tup Pair[metav1.GroupResource, nsdPreInformer]) error {
			logger.V(4).Info("Resyncing old informer for resource in source", "groupResource", tup.First, "namespaced", tup.Second.namespaced)
			wps.resyncGroupResource(tup.First, tup.Second.namespaced, tup.Second.preInformer.Informer())
			return nil
		})
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
			npi, have := wps.preInformers.Get(gr)
			if !have {
				logger.V(4).Info("Instantiating new informer for namespaced resource")
				sgvr := MetaGroupResourceToSchema(gr).WithVersion(pmv.APIVersion)
				npi = nsdPreInformer{namespaced: true, preInformer: wps.dynamicInformerFactory.ForResource(sgvr)}
				wps.preInformers.Put(gr, npi)
				npi.preInformer.Informer().AddEventHandler(k8scache.ResourceEventHandlerFuncs{
					AddFunc:    func(obj any) { wps.enqueueSourceObject(gr, true, obj, "add") },
					UpdateFunc: func(oldObj, newObj any) { wps.enqueueSourceObject(gr, true, newObj, "update") },
					DeleteFunc: func(obj any) { wps.enqueueSourceObject(gr, true, obj, "delete") },
				})
				go npi.preInformer.Informer().Run(wp.ctx.Done())
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
			npi, have := wps.preInformers.Get(gr)
			if !have {
				logger.V(4).Info("Instantiating new informer for cluster-scoped resource")
				sgvr := MetaGroupResourceToSchema(gr).WithVersion(pmv.APIVersion)
				npi = nsdPreInformer{namespaced: false, preInformer: wps.dynamicInformerFactory.ForResource(sgvr)}
				wps.preInformers.Put(gr, npi)
				npi.preInformer.Informer().AddEventHandler(k8scache.ResourceEventHandlerFuncs{
					AddFunc:    func(obj any) { wps.enqueueSourceObject(gr, false, obj, "add") },
					UpdateFunc: func(oldObj, newObj any) { wps.enqueueSourceObject(gr, false, newObj, "update") },
					DeleteFunc: func(obj any) { wps.enqueueSourceObject(gr, false, obj, "delete") },
				})
				go npi.preInformer.Informer().Run(wp.ctx.Done())
				time.Sleep(wp.delay)
			}

			return nil
		})
		return nil
	})
	(*changedDestinations).Visit(func(destination SinglePlacement) error {
		mbwsName := SPMailboxWorkspaceName(destination)
		wp.mbwsNameToSP.Put(mbwsName, destination)
		logger := logger.WithValues("destination", destination)
		wpd := MapGetAdd(wp.perDestination, destination, false, wp.newPerDestinationLocked)
		if wpd == nil {
			logger.Error(nil, "Impossible: no per-destination record for affected destination")
			return nil
		}
		logger.V(4).Info("NamespaceDistributions after transaction", "them", VisitableToSlice[Pair[NamespaceName, Set[logicalcluster.Name]]](wpd.nsDistributions.GetIndex1to2()))
		logger.V(4).Info("NamespacedResourceDistributions after transaction", "them", VisitableToSlice[Pair[metav1.GroupResource, Set[logicalcluster.Name]]](wpd.nsrDistributions.GetIndex1to2()))
		logger.V(4).Info("NonNamespacedDistributions after transaction", "them", VisitableToSlice[Pair[GroupResourceInstance, Set[logicalcluster.Name]]](wpd.nnsDistributions.GetIndex1to2()))
		nsms, have := wp.nsModesForSync.GetIndex().Get(destination)
		if have {
			logger.V(4).Info("Namespaced modes after transaction", "destination", destination, "modes", MapMapCopy[metav1.GroupResource, ProjectionModeVal](nil, nsms))
		} else {
			logger.V(4).Info("No Namespaced modes after transaction", "destination", destination)
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
		wpd.preInformers.Visit(func(tup Pair[metav1.GroupResource, dynamicDuo]) error {
			logger.V(4).Info("Resyncing GroupResource at destination", "groupResource", tup.First, "namespaced", tup.Second.namespaced)
			wpd.resyncGroupResource(tup.First, tup.Second)
			return nil
		})
		mbwsCluster, ok := wp.mbwsNameToCluster.Get(mbwsName)
		if !ok {
			logger.Error(nil, "Mailbox workspace not known yet", "destination", destination)
			wp.queue.Add(destination)
		} else {
			scRef := syncerConfigRef{mbwsCluster, SyncerConfigName}
			logger.V(4).Info("Enqueuing reference to SyncerConfig affected by transaction", "destination", destination, "mbwsName", mbwsName, "scRef", scRef)
			wp.queue.Add(scRef)
		}
		wpd.nsrDistributions.GetIndex1to2().Visit(func(tup Pair[metav1.GroupResource, Set[logicalcluster.Name]]) error {
			if tup.Second.IsEmpty() {
				return nil
			}
			pmv, have := nsms.Get(tup.First)
			if !have {
				logger.Error(nil, "Missing API version", "groupResource", tup.First)
				return nil
			}
			wpd.getDynamicDuoLocked(tup.First, pmv.APIVersion, true)
			return nil
		})
		wpd.nnsDistributions.GetIndex1to2().Visit(func(tup Pair[GroupResourceInstance, Set[logicalcluster.Name]]) error {
			if tup.Second.IsEmpty() {
				return nil
			}
			pmv, have := nnsms.Get(tup.First.First)
			if !have {
				logger.Error(nil, "Missing API version", "groupResourceInstance", tup.First)
				return nil
			}
			wpd.getDynamicDuoLocked(tup.First.First, pmv.APIVersion, false)
			return nil
		})
		return nil
	})
	logger.V(3).Info("End transaction")
	wp.changedDestinations = nil
}

func (wps *wpPerSource) resyncGroupResource(gr metav1.GroupResource, namespaced bool, informer k8scache.SharedIndexInformer) {
	objs := informer.GetStore().List()
	for _, obj := range objs {
		wps.enqueueSourceObject(gr, namespaced, obj, "resync")
	}
}

func MGRWithVersion(gr metav1.GroupResource, version string) metav1.GroupVersionResource {
	return metav1.GroupVersionResource{Group: gr.Group, Version: version, Resource: gr.Resource}
}

func (wps *wpPerSource) enqueueSourceObject(gr metav1.GroupResource, namespaced bool, obj any, action string) {
	dfu, ok := obj.(k8scache.DeletedFinalStateUnknown)
	if ok {
		obj = dfu.Obj
	}
	objm := obj.(metav1.Object)
	var namespace string = noNamespace
	if namespaced {
		namespace = objm.GetNamespace()
	}
	if ObjectIsSystem(objm) {
		wps.logger.V(4).Info("Ignoring system object", "groupResource", gr, "namespace", namespace, "name", objm.GetName(), "action", action)
		return
	}
	ref := sourceObjectRef{wps.source, gr, namespace, objm.GetName()}
	wps.logger.V(4).Info("Enqueuing reference to source object", "ref", ref)
	wps.wp.queue.Add(ref)
}

// enqueueDestinationObject enqueues a reference to a particular object in a particular MBWS.
// Calls to this happen only after the object's resource was defined (although it might later be undefined).
func (wpd *wpPerDestination) enqueueDestinationObject(gr metav1.GroupResource, namespaced bool, obj any, action string) {
	dfu, ok := obj.(k8scache.DeletedFinalStateUnknown)
	if ok {
		obj = dfu.Obj
	}
	objm := obj.(metav1.Object)
	var namespace string = noNamespace
	if namespaced {
		namespace = objm.GetNamespace()
	}
	if ObjectIsSystem(objm) {
		wpd.logger.V(4).Info("Ignoring system object", "groupResource", gr, "namespace", namespace, "name", objm.GetName(), "action", action)
		return
	}
	ref := destinationObjectRef{wpd.destination, gr, namespace, objm.GetName()}
	wpd.logger.V(4).Info("Enqueuing reference to destination object", "ref", ref)
	wpd.wp.queue.Add(ref)
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
	logger := klog.FromContext(wp.ctx).WithValues("destination", destination)
	wp.Lock()
	defer wp.Unlock()
	nsds, have := wp.nsDistributionsForSync.GetIndex1to2().Get(destination)
	ans := syncerConfigSpecRelations{
		clusterScopedObjects: NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[string /*object name*/]]](nil),
	}
	if have {
		nses := MapKeySet(nsds.GetIndex1to2())
		ans.namespaces = MapSetCopy(TransformVisitable[NamespaceName, string](nses, func(ns NamespaceName) string { return string(ns) }))
	} else {
		ans.namespaces = NewEmptyMapSet[string]()
	}
	nsrds, haveDists := wp.nsrDistributionsForSync.GetIndex1to2().Get(destination)
	if haveDists {
		nsms, haveModes := wp.nsModesForSync.GetIndex().Get(destination)
		if !haveModes {
			logger.Error(nil, "No ProjectionModeVals for namespaced resources")
			nsms = NewMapMap[metav1.GroupResource, ProjectionModeVal](nil)
		}
		nsrs := MapKeySet(nsrds.GetIndex1to2())
		ans.namespacedResources = MapSetCopy(TransformVisitable[metav1.GroupResource, edgeapi.NamespaceScopeDownsyncResource](nsrs, func(gr metav1.GroupResource) edgeapi.NamespaceScopeDownsyncResource {
			pmv, ok := nsms.Get(gr)
			if !ok {
				logger.Error(nil, "Missing API group version info", "groupResource", gr)
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
			logger.Error(nil, "No ProjectionModeVals for cluster-scoped resources")
			nnsms = NewMapMap[metav1.GroupResource, ProjectionModeVal](nil)
		}
		objs := MapKeySet(nnsds.GetIndex1to2())
		objs.Visit(func(gri GroupResourceInstance) error {
			gr := gri.First
			rscMode := wp.resourceModes(gr)
			if !rscMode.GoesToEdge() {
				logger.V(5).Info("Omitting cluster-scoped resource from SyncerConfig because it does not go to edge clusters", "groupResource", gr)
				return nil
			}
			pmv, ok := nnsms.Get(gr)
			if !ok {
				logger.Error(nil, "Missing API version", "obj", gri)
			}
			cso := MapGetAdd(ans.clusterScopedObjects, gr,
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
