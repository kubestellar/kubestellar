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
	"strconv"
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

	clusterdynamic "github.com/kcp-dev/client-go/dynamic"
	kcpkubecorev1informers "github.com/kcp-dev/client-go/informers/core/v1"
	kcpkubecorev1client "github.com/kcp-dev/client-go/kubernetes/typed/core/v1"
	kcpcorev1a1 "github.com/kcp-dev/kcp/pkg/apis/core/v1alpha1"
	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyv1a1listers "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgeclusterclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
	edgev2a1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
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
// and reacts to objects reported by that informer in two ways:
// (1) remove objects that the WP projected earlier but should no longer be
// projected and (2) return the object's reported state to the source object in
// the WDS if requested by that object and the number of corresponding WECs is 1.
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
	locationInformer k8scache.SharedIndexInformer,
	locationLister edgev2a1listers.LocationLister,
	syncfgInformer k8scache.SharedIndexInformer,
	syncfgLister edgev2a1listers.SyncerConfigLister,
	customizerInformer k8scache.SharedIndexInformer,
	customizerLister edgev2a1listers.CustomizerLister,
	edgeClusterClientset edgeclusterclientset.ClusterInterface,
	dynamicClusterClient clusterdynamic.ClusterInterface,
	nsClusterPreInformer kcpkubecorev1informers.NamespaceClusterInformer,
	nsClusterClient kcpkubecorev1client.NamespaceClusterInterface,
) *workloadProjector {
	wp := &workloadProjector{
		// delay:                 2 * time.Second,
		ctx:                  ctx,
		configConcurrency:    configConcurrency,
		resourceModes:        resourceModes,
		queue:                workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		mbwsLister:           mbwsLister,
		locationInformer:     locationInformer,
		locationLister:       locationLister,
		syncfgInformer:       syncfgInformer,
		syncfgLister:         syncfgLister,
		customizerInformer:   customizerInformer,
		customizerLister:     customizerLister,
		edgeClusterClientset: edgeClusterClientset,
		dynamicClusterClient: dynamicClusterClient,
		nsClusterPreInformer: nsClusterPreInformer,
		nsClusterClient:      nsClusterClient,

		mbwsNameToCluster: WrapMapWithMutex[string, logicalcluster.Name](NewMapMap[string, logicalcluster.Name](nil)),
		clusterToMBWSName: WrapMapWithMutex[logicalcluster.Name, string](NewMapMap[logicalcluster.Name, string](nil)),
		mbwsNameToSP:      WrapMapWithMutex[string, SinglePlacement](NewMapMap[string, SinglePlacement](nil)),

		perSource:      NewMapMap[logicalcluster.Name, *wpPerSource](nil),
		perDestination: NewMapMap[SinglePlacement, *wpPerDestination](nil),

		upsyncs: NewHashRelation2[SinglePlacement, edgeapi.UpsyncSet](
			HashSinglePlacement{}, HashUpsyncSet{}),
	}
	wp.nsdDistributionsForProj = NewGenericFactoredMap[NamespacedDistributionTuple,
		logicalcluster.Name, Triple[metav1.GroupResource, NamespacedName, SinglePlacement], bool,
		wpPerSourceNSDistributions, wpPerSourceNSDistributions](
		factorNamespacedDistributionTupleForProj1and234,
		func(source logicalcluster.Name) wpPerSourceNSDistributions {
			wps := MapGetAdd(wp.perSource, source, true, wp.newPerSourceLocked)
			return wpPerSourceNSDistributions{wps}
		},
		func(nsd wpPerSourceNSDistributions) MutableMap[Triple[metav1.GroupResource, NamespacedName, SinglePlacement], bool] {
			return nsd.wps.nsdDistributions
		},
		Identity1[wpPerSourceNSDistributions],
		NewMapMap[logicalcluster.Name, wpPerSourceNSDistributions](nil),
		nil, nil)
	wp.nnsDistributionsForProj = NewGenericFactoredMap[NonNamespacedDistributionTuple,
		logicalcluster.Name, Triple[metav1.GroupResource, ObjectName, SinglePlacement], bool,
		wpPerSourceNNSDistributions, wpPerSourceNNSDistributions](
		factorNonNamespacedDistributionTupleForProj1and234,
		func(source logicalcluster.Name) wpPerSourceNNSDistributions {
			wps := MapGetAdd(wp.perSource, source, true, wp.newPerSourceLocked)
			return wpPerSourceNNSDistributions{wps}
		},
		func(nsd wpPerSourceNNSDistributions) MutableMap[Triple[metav1.GroupResource, ObjectName, SinglePlacement], bool] {
			return nsd.wps.nnsDistributions
		},
		Identity1[wpPerSourceNNSDistributions],
		NewMapMap[logicalcluster.Name, wpPerSourceNNSDistributions](nil),
		nil, nil)
	wp.nsdDistributionsForSync = NewGenericFactoredMap[NamespacedDistributionTuple, SinglePlacement, Pair[GroupResourceNamespacedName, logicalcluster.Name],
		bool, wpPerDestinationNSDistributions, wpPerDestinationNSDistributions](
		factorNamespacedDistributionTupleForSync1,
		func(destination SinglePlacement) wpPerDestinationNSDistributions {
			wpd := MapGetAdd(wp.perDestination, destination, true, wp.newPerDestinationLocked)
			return wpPerDestinationNSDistributions{wpd}
		},
		func(nsd wpPerDestinationNSDistributions) MutableMap[Pair[GroupResourceNamespacedName, logicalcluster.Name], bool] {
			return nsd.wpd.nsdDistributions
		},
		Identity1[wpPerDestinationNSDistributions],
		NewMapMap[SinglePlacement, wpPerDestinationNSDistributions](nil), nil, nil)
	wp.nnsDistributionsForSync = NewGenericFactoredMap[NonNamespacedDistributionTuple, SinglePlacement, Pair[GroupResourceObjectName, logicalcluster.Name],
		bool, wpPerDestinationNNSDistributions, wpPerDestinationNNSDistributions](
		factorNonNamespacedDistributionTupleForSync1,
		func(destination SinglePlacement) wpPerDestinationNNSDistributions {
			wpd := MapGetAdd(wp.perDestination, destination, true, wp.newPerDestinationLocked)
			return wpPerDestinationNNSDistributions{wpd}
		},
		func(nsd wpPerDestinationNNSDistributions) MutableMap[Pair[GroupResourceObjectName, logicalcluster.Name], bool] {
			return nsd.wpd.nnsDistributions
		},
		Identity1[wpPerDestinationNNSDistributions],
		NewMapMap[SinglePlacement, wpPerDestinationNNSDistributions](nil), nil, nil)
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
			switch ws.Status.Phase {
			case kcpcorev1a1.LogicalClusterPhaseReady: // the happy case
			case kcpcorev1a1.LogicalClusterPhaseInitializing, kcpcorev1a1.LogicalClusterPhaseScheduling:
				logger.V(4).Info("Ignoring mailbox workspace that is not ready yet", "wsName", ws.Name, "phase", ws.Status.Phase)
				return
			default:
				logger.V(2).Info("Ignoring mailbox workspace with unexpected phase", "wsName", ws.Name, "phase", ws.Status.Phase)
				return
			}
			if cluster == "" {
				logger.Error(nil, "Impossible: mailbox workspace with empty cluster name", "wsName", ws.Name)
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
			switch ws.Status.Phase {
			case kcpcorev1a1.LogicalClusterPhaseReady: // the happy case
			case kcpcorev1a1.LogicalClusterPhaseInitializing, kcpcorev1a1.LogicalClusterPhaseScheduling:
				logger.V(4).Info("Ignoring mailbox workspace that is not ready yet", "wsName", ws.Name, "phase", ws.Status.Phase)
				return
			default:
				logger.V(2).Info("Ignoring mailbox workspace with unexpected phase", "wsName", ws.Name, "phase", ws.Status.Phase)
				return
			}
			oldCluster, has := wp.mbwsNameToCluster.Get(ws.Name)
			if cluster == "" {
				logger.Error(nil, "Impossible: mailbox workspace with empty cluster name", "wsName", ws.Name)
				return
			}
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
		if cluster == "" {
			logger.Error(nil, "Impossible: SyncerConfig with empty cluster name", "syncfg", syncfg)
			return
		}
		scRef := syncerConfigRef{cluster, ObjectName(syncfg.Name)}
		logger.V(4).Info("Enqueuing reference to SyncerConfig from informer", "scRef", scRef, "event", event)
		wp.queue.Add(scRef)
	}
	syncfgInformer.AddEventHandler(k8scache.ResourceEventHandlerFuncs{
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
// The value set types are thin wrappers (wpPerSourceNSDistributions_Gone et al.) around
// the relevant per-cluster data structures, each of which has the various
// second-level indices in it and the wrapper exposes the relevant second-level index.
// We currently do not bother with distinct types for readonly indices.
//
// The fields following the Mutex should only be accessed with the Mutex locked.
type workloadProjector struct {
	ctx                  context.Context
	configConcurrency    int
	resourceModes        ResourceModes
	delay                time.Duration // to slow down for debugging
	queue                workqueue.RateLimitingInterface
	mbwsLister           tenancyv1a1listers.WorkspaceLister
	locationInformer     k8scache.SharedIndexInformer
	locationLister       edgev2a1listers.LocationLister
	syncfgInformer       k8scache.SharedIndexInformer
	syncfgLister         edgev2a1listers.SyncerConfigLister
	customizerInformer   k8scache.SharedIndexInformer
	customizerLister     edgev2a1listers.CustomizerLister
	edgeClusterClientset edgeclusterclientset.ClusterInterface
	dynamicClusterClient clusterdynamic.ClusterInterface
	nsClusterPreInformer kcpkubecorev1informers.NamespaceClusterInformer
	nsClusterClient      kcpkubecorev1client.NamespaceClusterInterface

	mbwsNameToCluster MutableMap[string /*mailbox workspace name*/, logicalcluster.Name]
	clusterToMBWSName MutableMap[logicalcluster.Name, string /*mailbox workspace name*/]
	mbwsNameToSP      MutableMap[string /*mailbox workspace name*/, SinglePlacement]

	sync.Mutex

	perSource      MutableMap[logicalcluster.Name, *wpPerSource]
	perDestination MutableMap[SinglePlacement, *wpPerDestination]

	// changedDestinations is the destinations affected during a transaction
	changedDestinations *MutableSet[SinglePlacement]

	// NonNamespacedDistributions indexed for projection
	nsdDistributionsForProj GenericFactoredMap[NamespacedDistributionTuple, logicalcluster.Name,
		Triple[metav1.GroupResource, NamespacedName, SinglePlacement], bool, wpPerSourceNSDistributions]

	// NonNamespacedDistributions indexed for projection
	nnsDistributionsForProj GenericFactoredMap[NonNamespacedDistributionTuple, logicalcluster.Name,
		Triple[metav1.GroupResource, ObjectName, SinglePlacement], bool, wpPerSourceNNSDistributions]

	nsModesForProj  FactoredMap[ProjectionModeKey, metav1.GroupResource, SinglePlacement, ProjectionModeVal]
	nnsModesForProj FactoredMap[ProjectionModeKey, metav1.GroupResource, SinglePlacement, ProjectionModeVal]

	// NamespacedDistributions indexed for SyncerConfig maintenance
	nsdDistributionsForSync GenericFactoredMap[NamespacedDistributionTuple, SinglePlacement,
		Pair[GroupResourceNamespacedName, logicalcluster.Name], bool, wpPerDestinationNSDistributions]

	// NonNamespacedDistributions indexed for SyncerConfig maintenance
	nnsDistributionsForSync GenericFactoredMap[NonNamespacedDistributionTuple, SinglePlacement,
		Pair[GroupResourceObjectName, logicalcluster.Name], bool, wpPerDestinationNNSDistributions]

	nsModesForSync  FactoredMap[ProjectionModeKey, SinglePlacement, metav1.GroupResource, ProjectionModeVal]
	nnsModesForSync FactoredMap[ProjectionModeKey, SinglePlacement, metav1.GroupResource, ProjectionModeVal]

	upsyncs SingleIndexedRelation2[SinglePlacement, edgeapi.UpsyncSet]
}

type GroupResourceNamespacedName = Pair[metav1.GroupResource, NamespacedName]
type GroupResourceObjectName = Pair[metav1.GroupResource, ObjectName]

// Constructs the data structure specific to a mailbox/edge-cluster
func (wp *workloadProjector) newPerDestinationLocked(destination SinglePlacement) *wpPerDestination {
	wpd := &wpPerDestination{wp: wp, destination: destination,
		logger:           klog.FromContext(wp.ctx).WithValues("destination", destination),
		nsdDistributions: NewSingleIndexedMapMap2[GroupResourceNamespacedName, logicalcluster.Name, bool](),
		nnsDistributions: NewSingleIndexedMapMap2[GroupResourceObjectName, logicalcluster.Name, bool](),
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
	nsdDistributions SingleIndexedMap2[GroupResourceNamespacedName, logicalcluster.Name, bool]
	nnsDistributions SingleIndexedMap2[GroupResourceObjectName, logicalcluster.Name, bool]

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

func (duo *dynamicDuo) clientForMaybeNamespace(namespaced bool, namespace string) k8sdynamic.ResourceInterface {
	if namespaced {
		return duo.client.Namespace(namespace)
	}
	return duo.client
}

func (duo *dynamicDuo) clientAndGetterForMaybeNamespace(namespaced bool, namespace string) (k8sdynamic.ResourceInterface, getter) {
	if namespaced {
		return duo.client.Namespace(namespace), duo.preInformer.Lister().ByNamespace(namespace)
	}
	return duo.client, duo.preInformer.Lister()
}

type getter interface {
	Get(string) (machruntime.Object, error)
}

// getDynamicDuoLocked ensures that wpd's dynamic client and informerFactory have been created,
// and ensures that gr's instances are being monitored (unless gr is namespaces) (in case they need to be deleted).
func (wpd *wpPerDestination) getDynamicDuoLocked(gr metav1.GroupResource, apiVersion string, namespaced bool) (dynamicDuo, <-chan struct{}, error) {
	if wpd.dynamicClient == nil {
		wpd.logger.V(4).Info("Creating dynamicClient")
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
			k8scache.WaitForNamedCacheSync("workload-projector("+gr.String()+","+apiVersion+")", wpd.wp.ctx.Done(), nsInformer.HasSynced)
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

func (nsd wpPerDestinationNSDistributions) GetIndex() GenericFactoredMapIndex[GroupResourceNamespacedName, logicalcluster.Name, bool, sourcesWantReturns] {
	return nsd.wpd.nsdDistributions.GetIndex()
}

type wpPerDestinationNNSDistributions struct {
	wpd *wpPerDestination
}

func (nsd wpPerDestinationNNSDistributions) GetIndex() GenericFactoredMapIndex[GroupResourceObjectName, logicalcluster.Name, bool, sourcesWantReturns] {
	return nsd.wpd.nnsDistributions.GetIndex()
}

type sourcesWantReturns = Map[logicalcluster.Name, bool]

// Constructs the data structure specific to a workload management workspace
func (wp *workloadProjector) newPerSourceLocked(source logicalcluster.Name) *wpPerSource {
	dynamicClient := wp.dynamicClusterClient.Cluster(source.Path())
	dynamicInformerFactory := k8sdynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0)
	wps := &wpPerSource{wp: wp, source: source,
		logger:                 klog.FromContext(wp.ctx).WithValues("source", source),
		nsdDistributions:       NewSingleIndexedMapMap3[metav1.GroupResource, NamespacedName, SinglePlacement, bool](),
		nnsDistributions:       NewSingleIndexedMapMap3[metav1.GroupResource, ObjectName, SinglePlacement, bool](),
		dynamicClient:          dynamicClient,
		dynamicInformerFactory: dynamicInformerFactory,
		preInformers:           NewMapMap[metav1.GroupResource, dynamicDuo](nil),
	}
	dynamicInformerFactory.Start(wp.ctx.Done())
	return wps
}

// The data structure specific to a WDS (formerly workload management workspace).
// All the variable fields must be accessed with the wp mutex locked.
// All the clients and informers are about accessing the contents of the WDS.
type wpPerSource struct {
	wp                     *workloadProjector
	source                 logicalcluster.Name
	logger                 klog.Logger
	nsdDistributions       SingleIndexedMap3[metav1.GroupResource, NamespacedName, SinglePlacement, bool]
	nnsDistributions       SingleIndexedMap3[metav1.GroupResource, ObjectName, SinglePlacement, bool]
	dynamicClient          k8sdynamic.Interface
	dynamicInformerFactory k8sdynamicinformer.DynamicSharedInformerFactory
	preInformers           MutableMap[metav1.GroupResource, dynamicDuo]
}

type wpPerSourceNSDistributions struct {
	wps *wpPerSource
}

func (nsd wpPerSourceNSDistributions) GetIndex() GenericFactoredMapIndex[metav1.GroupResource,
	Pair[NamespacedName, SinglePlacement], bool, NamspacedNameToObjectDestinations] {
	return nsd.wps.nsdDistributions.GetIndex()
}

type NamspacedNameToDestinations = GenericIndexedSet[Pair[NamespacedName, SinglePlacement],
	NamespacedName, SinglePlacement, Set[SinglePlacement]]

type NamspacedNameToObjectDestinations = GenericFactoredMap[Pair[NamespacedName, SinglePlacement],
	NamespacedName, SinglePlacement, bool, Map[SinglePlacement, bool]]

type wpPerSourceNNSDistributions struct {
	wps *wpPerSource
}

func (nsd wpPerSourceNNSDistributions) GetIndex() GenericFactoredMapIndex[metav1.GroupResource,
	Pair[ObjectName, SinglePlacement], bool, ObjectNameToObjectDestinations] {
	return nsd.wps.nnsDistributions.GetIndex()
}

type ObjectNameToDestinations = GenericIndexedSet[Pair[ObjectName, SinglePlacement],
	ObjectName, SinglePlacement, Set[SinglePlacement]]

type ObjectNameToObjectDestinations = GenericFactoredMap[Pair[ObjectName, SinglePlacement],
	ObjectName, SinglePlacement, bool, Map[SinglePlacement, bool]]

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
	name          ObjectName
}

const noNamespace = "no NS"

func (doRef *destinationObjectRef) namespacedName() NamespacedName {
	return NewPair(NamespaceName(doRef.namespace), doRef.name)
}

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
	logger.V(4).Info("Dequeued reference", "ref", ref, "type", fmt.Sprintf("%T", ref))
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

// Returns `retry bool`.
func (wp *workloadProjector) syncConifgDestination(ctx context.Context, destination SinglePlacement) bool {
	mbwsName := SPMailboxWorkspaceName(destination)
	mbwsCluster, ok := wp.mbwsNameToCluster.Get(mbwsName)
	logger := klog.FromContext(ctx)
	if !ok {
		return true
	}
	if mbwsCluster == "" {
		logger.Error(nil, "Impossible: mailbox workspace corresponds with empty cluster ID", "mbwsName", mbwsName)
		return true
	}
	scRef := syncerConfigRef{mbwsCluster, SyncerConfigName}
	logger.V(3).Info("Finally able to enqueue SyncerConfig ref", "scRef", scRef)
	wp.queue.Add(scRef)
	return false
}

// Returns `retry bool`.
func (wp *workloadProjector) syncConfigObject(ctx context.Context, scRef syncerConfigRef) bool {
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("syncerConfig", scRef)
	if scRef.Cluster == "" {
		logger.Error(nil, "Impossible: syncerConfigRef with empty string for Cluster")
		return false
	}
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
	syncfg, err := wp.syncfgLister.Get(string(scRef.Name))
	if err != nil {
		if k8sapierrors.IsNotFound(err) {
			goodConfigSpecRelations := wp.syncerConfigRelations(sp)
			syncfg = &edgeapi.SyncerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: string(scRef.Name),
				},
				Spec: wp.syncerConfigSpecFromRelations(goodConfigSpecRelations)}
			client := wp.edgeClusterClientset.EdgeV2alpha1().Cluster(scRef.Cluster.Path()).SyncerConfigs()
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
	client := wp.edgeClusterClientset.EdgeV2alpha1().Cluster(scRef.Cluster.Path()).SyncerConfigs()
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

// Returns `retry bool`.
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
		_, getter := duo.clientAndGetterForMaybeNamespace(namespaced, doRef.namespace)
		obj, err := getter.Get(string(doRef.name))
		present := err == nil && obj != nil
		var objM metav1.Object
		if present {
			objM = obj.(metav1.Object)
			if objM.GetDeletionTimestamp() != nil {
				present = false
			}
		}
		var sourcesWants sourcesWantReturns
		var haveSources bool
		if namespaced {
			sourcesWants, haveSources = wpd.nsdDistributions.GetIndex().Get(NewPair(doRef.groupResource, doRef.namespacedName()))
		} else {
			sourcesWants, haveSources = wpd.nnsDistributions.GetIndex().Get(NewPair(doRef.groupResource, doRef.name))
		}
		if haveSources && !sourcesWants.IsEmpty() {
			if !present {
				logger.V(4).Info("Ignoring destination object that is being deleted", "namespaced", namespaced)
				return returnFalse
			}
			logger.V(4).Info("Retaining destination object", "namespaced", namespaced, "sources", VisitableToSlice[Pair[logicalcluster.Name, bool]](sourcesWants))
			tryem := triers{}
			addRetry := false
			// accumulate reported state return tasks
			sourcesWants.Visit(func(sourceWant Pair[logicalcluster.Name, bool]) error {
				if !sourceWant.Second {
					return nil
				}
				wps, haveWPS := wp.perSource.Get(sourceWant.First)
				if !haveWPS {
					logger.Error(nil, "Impossible: no wps found for source that wants singleton reported state", "source", sourceWant.First)
					return nil
				}
				var destWants Map[SinglePlacement, bool]
				var haveDestWants bool
				if namespaced {
					rem, haveRem := wps.nsdDistributions.GetIndex().Get(doRef.groupResource)
					if haveRem {
						destWants, haveDestWants = rem.GetIndex().Get(doRef.namespacedName())
					}
				} else {
					rem, haveRem := wps.nnsDistributions.GetIndex().Get(doRef.groupResource)
					if haveRem {
						destWants, haveDestWants = rem.GetIndex().Get(doRef.name)
					}
				}
				if !haveDestWants {
					logger.Error(nil, "Impossible: up-then-down look got nothing", "source", sourceWant.First)
					return nil
				}
				if destCount := destWants.Len(); destCount != 1 {
					logger.Error(nil, "Reported state not returned because not singleton", "source", sourceWant.First, "destCount", destCount)
					return nil
				}
				srcDuo, haveSrcDuo := wps.preInformers.Get(doRef.groupResource)
				if !haveSrcDuo {
					logger.Error(nil, "Impossible: no dynamicDuo found for source that wants singleton reported state", "source", sourceWant.First)
					return nil
				}
				srcClient, srcGetter := srcDuo.clientAndGetterForMaybeNamespace(namespaced, doRef.namespace)
				srcObj, srcErr := srcGetter.Get(string(doRef.name))
				if srcObj == nil || srcErr != nil && k8sapierrors.IsNotFound(srcErr) {
					logger.V(3).Info("Retrying later because source object not found for source that wants singleton reported state", "source", sourceWant.First)
					addRetry = true
					return nil
				}
				if srcErr != nil {
					logger.V(3).Info("Impossible: failed to Get from cache", "source", sourceWant.First, "err", err)
					return nil
				}
				dstU := objM.(*unstructured.Unstructured)
				srcU := srcObj.(*unstructured.Unstructured)
				tryem = append(tryem, func() bool {
					return wps.reportSingletonState(ctx, logger, srcClient, srcU, dstU)
				})
				return nil
			})
			if addRetry {
				tryem = append(tryem, returnTrue)
			}
			return tryem.try
		}
		if !present {
			logger.V(4).Info("Undesired destination object is already absent", "err", err, "obj", obj)
			return returnFalse
		}
		resourceVersion := objM.GetResourceVersion()
		rscClient := duo.clientForMaybeNamespace(namespaced, doRef.namespace)
		return func() bool {
			err := rscClient.Delete(ctx, string(doRef.name),
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
	finish := func() triers { // produce the work to do after releasing the mutex
		wp.Lock()
		defer wp.Unlock()
		wps, have := wp.perSource.Get(soRef.cluster)
		if !have {
			logger.Error(nil, "Impossible: handing object from unknown source")
			return triers{returnFalse}
		}
		srcDuo, have := wps.preInformers.Get(soRef.groupResource)
		if !have {
			logger.Error(nil, "Impossible: handling source object of unknown resource")
			return triers{returnFalse}
		}
		srcClient, srcGetter := srcDuo.clientAndGetterForMaybeNamespace(namespaced, soRef.namespace)
		srcObj, err := srcGetter.Get(soRef.name)
		if err != nil && !k8sapierrors.IsNotFound(err) {
			logger.Error(nil, "Impossible: failed to lookup source object in local cache")
			return triers{returnFalse}
		}
		var srcMRObject mrObject
		deleted := srcObj == nil || k8sapierrors.IsNotFound(err)
		if !deleted {
			srcMRObject = srcObj.(mrObject)
			deleted = srcMRObject.GetDeletionTimestamp() != nil
		}
		var objectDestinations Map[SinglePlacement, bool]
		var haveDestinations bool
		if namespaced {
			byNN, have := wps.nsdDistributions.GetIndex().Get(soRef.groupResource)
			if !have {
				logger.V(4).Info("No objects of this source and namespaced kind are going anywhere")
				return triers{returnFalse}
			}
			objectDestinations, haveDestinations = byNN.GetIndex().Get(NamespacedName{NamespaceName(soRef.namespace), ObjectName(soRef.name)})
		} else {
			byName, have := wps.nnsDistributions.GetIndex().Get(soRef.groupResource)
			if !have {
				logger.V(4).Info("No objects of this source and cluster-sccoped kind are going anywhere")
				return triers{returnFalse}
			}
			objectDestinations, haveDestinations = byName.GetIndex().Get(ObjectName(soRef.name))
		}
		if !haveDestinations {
			logger.V(4).Info("Object is not going anywhere")
			return triers{returnFalse}
		}
		numDestinations := objectDestinations.Len()
		logger.V(4).Info("Object is going places", "num", numDestinations)
		modesForSync := wps.wp.nnsModesForSync
		if namespaced {
			modesForSync = wps.wp.nsModesForSync
		}
		var tryAgain bool
		remWork := triers{}
		objectDestinations.Visit(func(tup Pair[SinglePlacement, bool]) error {
			retryThis, rem := wps.syncSourceToDestLocked(ctx, logger, srcClient, soRef, srcMRObject, namespaced, deleted, modesForSync, tup.First, tup.Second, numDestinations)
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
	return finish.try()
}

var returnFalse = func() bool { return false }
var returnTrue = func() bool { return true }

// a trier is a func that is executed outside the mutex and returns `retry bool`
type triers []func() bool

func (ts triers) try() bool {
	retry := false
	for _, trier := range ts {
		if trier() {
			retry = true
		}
	}
	return retry
}

// Sync a source object to one MBWS.
// Returns `(retry bool, unlocked func() (retry bool))`
func (wps *wpPerSource) syncSourceToDestLocked(ctx context.Context, logger klog.Logger,
	srcClient k8sdynamic.ResourceInterface,
	soRef sourceObjectRef, srcMRObject mrObject, namespaced, deleted bool,
	modesForSync FactoredMap[ProjectionModeKey, SinglePlacement, metav1.GroupResource, ProjectionModeVal],
	destination SinglePlacement, returnSingletonReport bool, numDestinations int) (bool, func() bool) {
	wp := wps.wp
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
	destDuo, clientReadyChan, err := wpd.getDynamicDuoLocked(soRef.groupResource, pmv.APIVersion, namespaced)
	if err != nil {
		logger.Error(err, "Failed to wpd.getDynamicDuoLocked")
		return true, nil
	}
	return false, func() bool {
		// sgvr := MetaGroupResourceToSchema(soRef.groupResource).WithVersion(pmv.APIVersion)
		rscClient := destDuo.clientForMaybeNamespace(namespaced, soRef.namespace)
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
		if returnSingletonReport {
			if !wp.ensureDestCount(ctx, logger, srcClient, srcMRObject, numDestinations) {
				return false
			}
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
			if returnSingletonReport && numDestinations == 1 {
				srcU := srcMRObject.(*unstructured.Unstructured)
				if !wps.reportSingletonState(ctx, logger, srcClient, srcU, destObj) {
					return false
				}
			}
			revisedDestObj := wpd.wp.genericObjectMerge(soRef.cluster, destination, srcMRObject, destObj)
			if apiequality.Semantic.DeepEqual(destObj, revisedDestObj) {
				logger.V(4).Info("No need to update object in mailbox workspace")
				return false
			}
			time.Sleep(wp.delay)
			asUpdated, err := rscClient.Update(ctx, revisedDestObj, metav1.UpdateOptions{FieldManager: FieldManager})
			if err != nil {
				logger.V(2).Info("Failed to update object in mailbox workspace", "resourceVersion", revisedDestObj.GetResourceVersion(), "err", err)
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

func (wp *workloadProjector) ensureDestCount(ctx context.Context, logger klog.Logger,
	srcClient k8sdynamic.ResourceInterface, srcMRObject mrObject, numDestinations int,
) bool /* OK */ {
	annotations := srcMRObject.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	have := annotations[edgeapi.ExecutingCountKey]
	want := strconv.FormatInt(int64(numDestinations), 10)
	if have == want {
		return true
	}
	srcUnstr := srcMRObject.DeepCopyObject().(*unstructured.Unstructured)
	annotations = srcUnstr.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[edgeapi.ExecutingCountKey] = want
	srcUnstr.SetAnnotations(annotations)
	// TODO: try patch (conditional on ResourceVersion) instead
	_, err := srcClient.Update(ctx, srcUnstr, metav1.UpdateOptions{FieldManager: FieldManager})
	if err != nil {
		logger.Info("Update attempt did not succeed", "err", err)
		return false
	}
	return true
}

func (wps *wpPerSource) reportSingletonState(ctx context.Context, logger klog.Logger,
	srcClient k8sdynamic.ResourceInterface, srcU, dstU *unstructured.Unstructured,
) bool /* OK */ {
	srcUU := srcU.UnstructuredContent()
	dstUU := dstU.UnstructuredContent()
	srcStatusU := srcUU["status"]
	dstStatusU := dstUU["status"]
	if apiequality.Semantic.DeepEqual(srcStatusU, dstStatusU) {
		logger.V(4).Info("Singleton reported state did not change", "statusIsNil", srcStatusU == nil)
		return true
	}
	srcCopyU := srcU.DeepCopy()
	srcCopyUU := srcCopyU.UnstructuredContent()
	dstCopyU := dstU.DeepCopy()
	dstCopyUU := dstCopyU.UnstructuredContent()
	srcCopyUU["status"] = dstCopyUU["status"]
	// TODO: use API metadata to decide whether to use Update or UpdateStatus
	_, err := srcClient.UpdateStatus(ctx, srcCopyU, metav1.UpdateOptions{FieldManager: FieldManager})
	if err != nil {
		logger.V(2).Info("Return of singleton reported state did not happen", "err", err)
		return false
	}
	logger.V(2).Info("Return of singleton reported state happened")
	return true
}

func LabelsGet[Val any](labels map[string]Val, key string) Val {
	if labels == nil {
		var zero Val
		return zero
	}
	return labels[key]
}

const ProjectedLabelKey string = "edge.kubestellar.io/projected"
const ProjectedLabelVal string = "yes"

func (wp *workloadProjector) xformForDestination(sourceCluster logicalcluster.Name, destSP SinglePlacement, srcObj mrObject) *unstructured.Unstructured {
	srcObjU := srcObj.(*unstructured.Unstructured)
	logger := klog.FromContext(wp.ctx).WithValues(
		"sourceCluster", sourceCluster,
		"destSP", destSP,
		"destGVK", srcObjU.GroupVersionKind(),
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
		"destGVK", srcObjU.GroupVersionKind(),
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
		customizer, err = wp.customizerLister.Customizers(custNS).Get(custName)
		if err != nil {
			logger.Error(err, "Failed to find referenced Customizer")
		} else {
			expandParameters = expandParameters || customizer.Annotations[edgeapi.ParameterExpansionAnnotationKey] == "true"
		}
	}
	var location *edgeapi.Location
	if expandParameters {
		location, err = wp.locationLister.Get(destSP.LocationName)
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
	return (strings.Contains(key, ".kcp.io/") || strings.HasPrefix(key, "kcp.io/"))
	// return (strings.Contains(key, ".kcp.io/") || strings.HasPrefix(key, "kcp.io/")) && !strings.Contains(key, "edge.kubestellar.io/")
}

func (wp *workloadProjector) Transact(xn func(WorkloadProjectionSections)) {
	logger := klog.FromContext(wp.ctx)
	wp.Lock()
	defer wp.Unlock()
	logger.V(3).Info("Begin transaction")
	changedDestinations := WrapSetWithMutex[SinglePlacement](NewMapSet[SinglePlacement]())
	wp.changedDestinations = &changedDestinations
	recordLogger := logger.V(4)
	changedSources := WrapSetWithMutex[logicalcluster.Name](NewMapSet[logicalcluster.Name]())
	nsod := MappingReceiverFork[NamespacedDistributionTuple, bool]{
		MapKeySetReceiverLossy[NamespacedDistributionTuple, bool](SetWriterFork[NamespacedDistributionTuple](false,
			recordPart(recordLogger, "nsd.src", &changedDestinations, factorNamespacedDistributionTupleForSync1),
			recordPart(recordLogger, "nsd.dest", &changedSources, factorNamespacedDistributionTupleForProj1))),
		wp.nsdDistributionsForSync, wp.nsdDistributionsForProj}
	nnsod := MappingReceiverFork[NonNamespacedDistributionTuple, bool]{
		MapKeySetReceiverLossy[NonNamespacedDistributionTuple, bool](SetWriterFork[NonNamespacedDistributionTuple](false,
			recordPart(recordLogger, "nns.src", &changedDestinations, factorNonNamespacedDistributionTupleForSync1),
			recordPart(recordLogger, "nns.dest", &changedSources, factorNonNamespacedDistributionTupleForProj1))),
		wp.nnsDistributionsForSync, wp.nnsDistributionsForProj}
	xn(WorkloadProjectionSections{
		nsod,
		NewMappingReceiverFork[ProjectionModeKey, ProjectionModeVal](wp.nsModesForSync, wp.nsModesForProj,
			NewLoggingMappingReceiver[ProjectionModeKey, ProjectionModeVal]("wp.nsModes", recordLogger)),
		nnsod,
		NewMappingReceiverFork[ProjectionModeKey, ProjectionModeVal](wp.nnsModesForSync, wp.nnsModesForProj,
			NewLoggingMappingReceiver[ProjectionModeKey, ProjectionModeVal]("wp.nnsModes", recordLogger)),
		wp.upsyncs})
	logger.V(3).Info("Transaction response",
		"changedDestinations", VisitableToSlice[SinglePlacement](changedDestinations),
		"changedSources", VisitableToSlice[logicalcluster.Name](changedSources))
	changedSources.Visit(func(source logicalcluster.Name) error {
		wps, have := wp.perSource.Get(source)
		logger := logger.WithValues("source", source)
		if !have {
			logger.Error(nil, "No per-source data for changed source")
			return nil
		}
		logger.V(4).Info("Finishing transaction wrt source",
			"nsdDistributions", VisitableToSlice[Pair[Triple[metav1.GroupResource, NamespacedName, SinglePlacement], bool]](wps.nsdDistributions),
			"nnsDistributions", VisitableToSlice[Pair[Triple[metav1.GroupResource, ObjectName, SinglePlacement], bool]](wps.nnsDistributions),
		)
		wps.preInformers.Visit(func(tup Pair[metav1.GroupResource, dynamicDuo]) error {
			logger.V(4).Info("Resyncing old informer for resource in source", "groupResource", tup.First, "namespaced", tup.Second.namespaced)
			wps.resyncGroupResource(tup.First, tup.Second.namespaced, tup.Second.preInformer.Informer())
			return nil
		})
		wps.nsdDistributions.GetIndex().Visit(func(tup Pair[metav1.GroupResource, NamspacedNameToObjectDestinations]) error {
			gr := tup.First
			logger := logger.WithValues("groupResource", gr)
			problem, have := wp.nsModesForProj.GetIndex().Get(gr)
			if !have {
				logger.Error(nil, "No projection mode")
				return nil
			}
			solve := pickThe1[metav1.GroupResource, SinglePlacement](logger, "eeek")
			pmv := solve(gr, problem)
			wps.getDynamicDuoLocked(logger, gr, pmv.APIVersion, true)
			return nil
		})

		wps.nnsDistributions.GetIndex().Visit(func(tup Pair[metav1.GroupResource, ObjectNameToObjectDestinations]) error {
			gr := tup.First
			logger := logger.WithValues("groupResource", gr)
			problem, have := wp.nnsModesForProj.GetIndex().Get(gr)
			if !have {
				logger.Error(nil, "No projection mode")
				return nil
			}
			solve := pickThe1[metav1.GroupResource, SinglePlacement](logger, "eeek")
			pmv := solve(gr, problem)
			wps.getDynamicDuoLocked(logger, gr, pmv.APIVersion, false)
			return nil
		})

		return nil
	})
	changedDestinations.Visit(func(destination SinglePlacement) error {
		mbwsName := SPMailboxWorkspaceName(destination)
		wp.mbwsNameToSP.Put(mbwsName, destination)
		logger := logger.WithValues("destination", destination)
		wpd := MapGetAdd(wp.perDestination, destination, false, wp.newPerDestinationLocked)
		if wpd == nil {
			logger.Error(nil, "Impossible: no per-destination record for affected destination")
			return nil
		}
		logger.V(4).Info("NamespacedDistributions after transaction", "them", VisitableToSlice[Pair[Pair[GroupResourceNamespacedName, logicalcluster.Name], bool]](wpd.nsdDistributions))
		logger.V(4).Info("NonNamespacedDistributions after transaction", "them", VisitableToSlice[Pair[Pair[GroupResourceObjectName, logicalcluster.Name], bool]](wpd.nnsDistributions))
		nsms, have := wp.nsModesForSync.GetIndex().Get(destination)
		if have {
			logger.V(4).Info("Namespaced modes after transaction", "modes", MapMapCopy[metav1.GroupResource, ProjectionModeVal](nil, nsms))
		} else {
			logger.V(4).Info("No Namespaced modes after transaction")
		}
		nnsms, have := wp.nnsModesForSync.GetIndex().Get(destination)
		if have {
			logger.V(4).Info("NonNamespaced modes after transaction", "modes", MapMapCopy[metav1.GroupResource, ProjectionModeVal](nil, nnsms))
		} else {
			logger.V(4).Info("No NonNamespaced modes after transaction")
		}
		if upsyncs, have := wp.upsyncs.GetIndex1to2().Get(destination); have {
			logger.V(4).Info("Upsyncs after transaction", "upsyncs", VisitableToSlice[edgeapi.UpsyncSet](upsyncs))
		} else {
			logger.V(4).Info("No Upsyncs after transaction")
		}
		wpd.preInformers.Visit(func(tup Pair[metav1.GroupResource, dynamicDuo]) error {
			// Reconsider every instance of this resource in case it should stop being projected.
			logger.V(4).Info("Resyncing GroupResource at destination", "groupResource", tup.First, "namespaced", tup.Second.namespaced)
			wpd.resyncGroupResource(tup.First, tup.Second)
			return nil
		})
		mbwsCluster, ok := wp.mbwsNameToCluster.Get(mbwsName)
		if !ok {
			logger.Error(nil, "Mailbox workspace not known yet")
			wp.queue.Add(destination)
		} else if mbwsCluster == "" {
			logger.Error(nil, "Mailbox workspace has empty clustername")
			wp.queue.Add(destination)
		} else {
			scRef := syncerConfigRef{mbwsCluster, SyncerConfigName}
			logger.V(4).Info("Enqueuing reference to SyncerConfig affected by transaction", "mbwsName", mbwsName, "scRef", scRef)
			wp.queue.Add(scRef)
		}
		wpd.nsdDistributions.GetIndex().Visit(func(tup Pair[GroupResourceNamespacedName, sourcesWantReturns]) error {
			if tup.Second.IsEmpty() {
				return nil
			}
			pmv, have := nsms.Get(tup.First.First)
			if !have {
				logger.Error(nil, "Missing API version", "groupResourceInstance", tup.First)
				return nil
			}
			wpd.getDynamicDuoLocked(tup.First.First, pmv.APIVersion, true)
			return nil
		})

		wpd.nnsDistributions.GetIndex().Visit(func(tup Pair[GroupResourceObjectName, sourcesWantReturns]) error {
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

func (wps *wpPerSource) getDynamicDuoLocked(logger klog.Logger, gr metav1.GroupResource, apiVersion string, namespaced bool) dynamicDuo {
	logger = logger.WithValues("apiVersion", apiVersion)
	duo, have := wps.preInformers.Get(gr)
	if !have {
		logger.V(4).Info("Instantiating new informer at source for resource", "namespaced", namespaced)
		sgvr := MetaGroupResourceToSchema(gr).WithVersion(apiVersion)
		duo = dynamicDuo{apiVersion: apiVersion, namespaced: namespaced,
			preInformer: wps.dynamicInformerFactory.ForResource(sgvr),
			client:      wps.dynamicClient.Resource(sgvr)}
		wps.preInformers.Put(gr, duo)
		duo.preInformer.Informer().AddEventHandler(k8scache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj any) { wps.enqueueSourceObject(gr, namespaced, obj, "add") },
			UpdateFunc: func(oldObj, newObj any) { wps.enqueueSourceObject(gr, namespaced, newObj, "update") },
			DeleteFunc: func(obj any) { wps.enqueueSourceObject(gr, namespaced, obj, "delete") },
		})
		go duo.preInformer.Informer().Run(wps.wp.ctx.Done())
		time.Sleep(wps.wp.delay)
	}
	return duo
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
	if false /* filtering now done in what resolver */ && ObjectIsSystem(objm) {
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
	if false /* filtering now done in what resolver */ && ObjectIsSystem(objm) {
		wpd.logger.V(4).Info("Ignoring system object", "groupResource", gr, "namespace", namespace, "name", objm.GetName(), "action", action)
		return
	}
	ref := destinationObjectRef{wpd.destination, gr, namespace, ObjectName(objm.GetName())}
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

var factorNamespacedDistributionTupleForSync1 = NewFactorer(
	func(whole NamespacedDistributionTuple) Pair[SinglePlacement, Pair[GroupResourceNamespacedName, logicalcluster.Name]] {
		grni := NewPair(whole.First.GroupResource, NamespacedName{whole.Second.Second, whole.Second.Third})
		return NewPair(whole.First.Destination, NewPair(grni, whole.Second.First))
	},
	func(parts Pair[SinglePlacement, Pair[GroupResourceNamespacedName, logicalcluster.Name]]) NamespacedDistributionTuple {
		return NewPair(ProjectionModeKey{parts.Second.First.First, parts.First},
			ExternalNamespacedName{parts.Second.Second, parts.Second.First.Second.First, parts.Second.First.Second.Second})
	})

var factorNamespacedDistributionTupleForProj1 = NewFactorer(
	func(whole NamespacedDistributionTuple) Pair[logicalcluster.Name, Pair[GroupResourceNamespacedName, SinglePlacement]] {
		grni := NewPair(whole.First.GroupResource, NamespacedName{whole.Second.Second, whole.Second.Third})
		return NewPair(whole.Second.First, NewPair(grni, whole.First.Destination))
	},
	func(parts Pair[logicalcluster.Name, Pair[GroupResourceNamespacedName, SinglePlacement]]) NamespacedDistributionTuple {
		return NewPair(ProjectionModeKey{parts.Second.First.First, parts.Second.Second},
			ExternalNamespacedName{parts.First, parts.Second.First.Second.First, parts.Second.First.Second.Second})
	})

var factorNamespacedDistributionTupleForProj1and234 = NewFactorer(
	func(whole NamespacedDistributionTuple) Pair[logicalcluster.Name, Triple[metav1.GroupResource, NamespacedName, SinglePlacement]] {
		return NewPair(whole.Second.First, NewTriple(whole.First.GroupResource, NamespacedName{whole.Second.Second, whole.Second.Third}, whole.First.Destination))
	},
	func(parts Pair[logicalcluster.Name, Triple[metav1.GroupResource, NamespacedName, SinglePlacement]]) NamespacedDistributionTuple {
		return NewPair(ProjectionModeKey{parts.Second.First, parts.Second.Third},
			ExternalNamespacedName{parts.First, parts.Second.Second.First, parts.Second.Second.Second})
	})

var factorNonNamespacedDistributionTupleForSync1 = NewFactorer(
	func(whole NonNamespacedDistributionTuple) Pair[SinglePlacement, Pair[GroupResourceObjectName, logicalcluster.Name]] {
		return NewPair(whole.First.Destination, NewPair(NewPair(whole.First.GroupResource, whole.Second.Name), whole.Second.Cluster))
	},
	func(parts Pair[SinglePlacement, Pair[GroupResourceObjectName, logicalcluster.Name]]) NonNamespacedDistributionTuple {
		return NewPair(ProjectionModeKey{parts.Second.First.First, parts.First},
			ExternalName{parts.Second.Second, parts.Second.First.Second})
	})

var factorNonNamespacedDistributionTupleForProj1 = NewFactorer(
	func(whole NonNamespacedDistributionTuple) Pair[logicalcluster.Name, Pair[GroupResourceObjectName, SinglePlacement]] {
		return NewPair(whole.Second.Cluster, NewPair(NewPair(whole.First.GroupResource, whole.Second.Name), whole.First.Destination))
	},
	func(parts Pair[logicalcluster.Name, Pair[GroupResourceObjectName, SinglePlacement]]) NonNamespacedDistributionTuple {
		return NewPair(ProjectionModeKey{parts.Second.First.First, parts.Second.Second},
			ExternalName{parts.First, parts.Second.First.Second})
	})

var factorNonNamespacedDistributionTupleForProj1and234 = NewFactorer(
	func(whole NonNamespacedDistributionTuple) Pair[logicalcluster.Name, Triple[metav1.GroupResource, ObjectName, SinglePlacement]] {
		return NewPair(whole.Second.Cluster, NewTriple(whole.First.GroupResource, whole.Second.Name, whole.First.Destination))
	},
	func(parts Pair[logicalcluster.Name, Triple[metav1.GroupResource, ObjectName, SinglePlacement]]) NonNamespacedDistributionTuple {
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
	namespacedObjects    MutableMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[NamespacedName]]]
	clusterScopedObjects MutableMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[ObjectName]]]
	upsyncs              Set[edgeapi.UpsyncSet]
}

func (wp *workloadProjector) syncerConfigRelations(destination SinglePlacement) syncerConfigSpecRelations {
	logger := klog.FromContext(wp.ctx).WithValues("destination", destination)
	wp.Lock()
	defer wp.Unlock()
	ans := syncerConfigSpecRelations{
		namespacedObjects:    NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[NamespacedName]]](nil),
		clusterScopedObjects: NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[ObjectName]]](nil),
	}
	nsds, haveDists := wp.nsdDistributionsForSync.GetIndex().Get(destination)
	if haveDists {
		nsms, haveModes := wp.nsModesForSync.GetIndex().Get(destination)
		if !haveModes {
			logger.Error(nil, "No ProjectionModeVals for namespace-scoped resources")
			nsms = NewMapMap[metav1.GroupResource, ProjectionModeVal](nil)
		}
		objs := MapKeySet[GroupResourceNamespacedName, sourcesWantReturns](nsds.GetIndex())
		objs.Visit(func(gri GroupResourceNamespacedName) error {
			gr := gri.First
			rscMode := wp.resourceModes(gr)
			if !rscMode.GoesToEdge() {
				logger.V(5).Info("Omitting namespaced resource from SyncerConfig because it does not go to edge clusters", "groupResource", gr)
				return nil
			}
			pmv, ok := nsms.Get(gr)
			if !ok {
				logger.Error(nil, "Missing API version", "obj", gri)
			}
			nso := MapGetAdd(ans.namespacedObjects, gr,
				true, func(metav1.GroupResource) Pair[ProjectionModeVal, MutableSet[NamespacedName]] {
					return NewPair[ProjectionModeVal, MutableSet[NamespacedName]](pmv, NewEmptyMapSet[NamespacedName]())
				})
			nso.Second.Add(gri.Second)
			return nil
		})
	}
	nnsds, haveDists := wp.nnsDistributionsForSync.GetIndex().Get(destination)
	if haveDists {
		nnsms, haveModes := wp.nnsModesForSync.GetIndex().Get(destination)
		if !haveModes {
			logger.Error(nil, "No ProjectionModeVals for cluster-scoped resources")
			nnsms = NewMapMap[metav1.GroupResource, ProjectionModeVal](nil)
		}
		objs := MapKeySet[GroupResourceObjectName, sourcesWantReturns](nnsds.GetIndex())
		objs.Visit(func(gri GroupResourceObjectName) error {
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
				true, func(metav1.GroupResource) Pair[ProjectionModeVal, MutableSet[ObjectName]] {
					return NewPair[ProjectionModeVal, MutableSet[ObjectName]](pmv, NewEmptyMapSet[ObjectName]())
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
		NamespacedObjects: MapTransformToSlice[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[NamespacedName]], edgeapi.NamespaceScopeDownsyncObjects](specRelations.namespacedObjects,
			func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[NamespacedName]]) edgeapi.NamespaceScopeDownsyncObjects {
				indexedNNs := NewMapRelation2[NamespaceName, ObjectName]()
				SetAddAll[NamespacedName](indexedNNs, val.Second)
				nans := VisitableTransformToSlice[Pair[NamespaceName, Set[ObjectName]], edgeapi.NamespaceAndNames](indexedNNs.GetIndex1to2(), func(perNS Pair[NamespaceName, Set[ObjectName]]) edgeapi.NamespaceAndNames {
					return edgeapi.NamespaceAndNames{Namespace: string(perNS.First), Names: VisitableTransformToSlice[ObjectName, string](perNS.Second, ObjectName.String)}
				})
				return edgeapi.NamespaceScopeDownsyncObjects{
					GroupResource:      key,
					APIVersion:         val.First.APIVersion,
					ObjectsByNamespace: nans,
				}
			}),
		ClusterScope: MapTransformToSlice[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[ObjectName]], edgeapi.ClusterScopeDownsyncResource](specRelations.clusterScopedObjects,
			func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[ObjectName]]) edgeapi.ClusterScopeDownsyncResource {
				return edgeapi.ClusterScopeDownsyncResource{
					GroupResource: key,
					APIVersion:    val.First.APIVersion,
					Objects:       VisitableToSlice(TransformVisitable[ObjectName, string](val.Second, ObjectName.String)),
				}
			}),
		Upsync: VisitableToSlice[edgeapi.UpsyncSet](specRelations.upsyncs),
	}
	return ans
}

func StringToObjectName(name string) ObjectName { return ObjectName(name) }

func NANToSet(nan edgeapi.NamespaceAndNames) Set[NamespacedName] {
	nsName := NamespaceName(nan.Namespace)
	nns := VisitableMap[string, NamespacedName](Slice[string](nan.Names), func(objName string) NamespacedName { return NamespacedName{nsName, ObjectName(objName)} })
	reducer := StatefulReducer(
		func() MutableSet[NamespacedName] { return NewEmptyMapSet[NamespacedName]() },
		MutableSetUnion1Elt[NamespacedName], Identity1[MutableSet[NamespacedName]])
	return reducer(nns)
}

func (wp *workloadProjector) syncerConfigIsGood(destination SinglePlacement, configRef ExternalName, syncfg *edgeapi.SyncerConfig, goodSpecRelations syncerConfigSpecRelations) bool {
	spec := syncfg.Spec
	logger := klog.FromContext(wp.ctx)
	logger = logger.WithValues("destination", destination, "syncerConfig", configRef, "resourceVersion", syncfg.ResourceVersion)
	good := true
	haveNamespacedOjbects := NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[NamespacedName]]](nil)
	for _, cr := range spec.NamespacedObjects { // union all that stuff
		nansList := VisitableMap[edgeapi.NamespaceAndNames, Set[NamespacedName]](Slice[edgeapi.NamespaceAndNames](cr.ObjectsByNamespace), NANToSet)
		reducer := StatefulReducer(
			func() MutableSet[NamespacedName] { return NewEmptyMapSet[NamespacedName]() },
			MutableSetUnion1Set[NamespacedName], Identity1[MutableSet[NamespacedName]])
		nans := reducer(nansList)
		mapping, had := haveNamespacedOjbects.Get(cr.GroupResource)
		if had {
			// TODO: Can mapping.First mismatch?
			SetAddAll[NamespacedName](mapping.Second, nans)
		} else {
			mapping = NewPair(ProjectionModeVal{cr.APIVersion}, nans)
		}
		haveNamespacedOjbects.Put(cr.GroupResource, mapping)
	}
	MapEnumerateDifferencesParametric[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[NamespacedName]]](nsoEqual, goodSpecRelations.namespacedObjects, haveNamespacedOjbects, MapChangeReceiverFuncs[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[NamespacedName]]]{
		OnCreate: func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[NamespacedName]]) {
			logger.V(4).Info("SyncerConfig has excess ClusterScopeDownsyncResource", "groupResource", key, "apiVersion", val.First.APIVersion, "objects", val.Second)
			good = false
		},
		OnUpdate: func(key metav1.GroupResource, goodVal, haveVal Pair[ProjectionModeVal, MutableSet[NamespacedName]]) {
			logger.V(4).Info("SyncerConfig wrong ClusterScopeDownsyncResource", "groupResource", key, "apiVersionGood", goodVal.First.APIVersion, "apiVersionHave", haveVal.First.APIVersion, "objectsGood", goodVal.Second, "objectsHave", haveVal.Second)
			good = false
		},
		OnDelete: func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[NamespacedName]]) {
			logger.V(4).Info("SyncerConfig lacks ClusterScopeDownsyncResource", "groupResource", key, "apiVersion", val.First.APIVersion, "objects", val.Second)
			good = false
		},
	})
	haveClusterScopedResources := NewMapMap[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[ObjectName]]](nil)
	for _, cr := range spec.ClusterScope {
		// var objects MutableSet[ObjectName] = NewMapSet(cr.Objects...)
		var objects MutableSet[ObjectName] = MapSetCopy(TransformVisitable[string, ObjectName](NewSlice(cr.Objects...), NewObjectName))
		haveClusterScopedResources.Put(cr.GroupResource, NewPair(ProjectionModeVal{cr.APIVersion}, objects))
	}
	MapEnumerateDifferencesParametric[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[ObjectName]]](csrEqual, goodSpecRelations.clusterScopedObjects, haveClusterScopedResources, MapChangeReceiverFuncs[metav1.GroupResource, Pair[ProjectionModeVal, MutableSet[ObjectName]]]{
		OnCreate: func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[ObjectName]]) {
			logger.V(4).Info("SyncerConfig has excess ClusterScopeDownsyncResource", "groupResource", key, "apiVersion", val.First.APIVersion, "objects", val.Second)
			good = false
		},
		OnUpdate: func(key metav1.GroupResource, goodVal, haveVal Pair[ProjectionModeVal, MutableSet[ObjectName]]) {
			logger.V(4).Info("SyncerConfig wrong ClusterScopeDownsyncResource", "groupResource", key, "apiVersionGood", goodVal.First.APIVersion, "apiVersionHave", haveVal.First.APIVersion, "objectsGood", goodVal.Second, "objectsHave", haveVal.Second)
			good = false
		},
		OnDelete: func(key metav1.GroupResource, val Pair[ProjectionModeVal, MutableSet[ObjectName]]) {
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

func nsoEqual(a, b Pair[ProjectionModeVal, MutableSet[NamespacedName]]) bool {
	return a.First == b.First && SetEqual[NamespacedName](a.Second, b.Second)
}
func csrEqual(a, b Pair[ProjectionModeVal, MutableSet[ObjectName]]) bool {
	return a.First == b.First && SetEqual[ObjectName](a.Second, b.Second)
}

func looksLikeMBWSName(wsName string) bool {
	mbwsNameParts := strings.Split(wsName, WSNameSep)
	return len(mbwsNameParts) == 2
}
