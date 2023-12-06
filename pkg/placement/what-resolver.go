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
	"strings"
	"sync"
	"time"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextinfactory "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	k8ssets "k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	kubedynamicinformer "k8s.io/client-go/dynamic/dynamicinformer"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	ksmetav1a1 "github.com/kubestellar/kubestellar/pkg/apis/meta/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/apiwatch"
	edgev2alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev2alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
	msclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
)

type whatResolver struct {
	ctx        context.Context
	logger     klog.Logger
	numThreads int
	queue      workqueue.RateLimitingInterface
	receiver   MappingReceiver[ExternalName, ResolvedWhat]

	edgePlacementInformer upstreamcache.SharedIndexInformer
	edgePlacementLister   edgev2alpha1listers.EdgePlacementLister

	spaceclient     msclient.KubestellarSpaceInterface
	spaceProviderNs string
	kbSpaceRelation kbuser.KubeBindSpaceRelation

	// Hold this while accessing data listed below
	sync.Mutex

	// workspaceDetails maps lc.Name of a workload LC (WDS) to all the relevant information for that LC.
	workspaceDetails map[string]*workspaceDetails
}

// workspaceDetails holds the data for a given WDS
type workspaceDetails struct {
	ctx context.Context
	// placements maps name of relevant EdgePlacement object to that object
	placements             map[ObjectName]*edgeapi.EdgePlacement
	stop                   func()
	apiInformer            upstreamcache.SharedInformer
	apiLister              apiwatch.APIResourceLister
	dynamicInformerFactory kubedynamicinformer.DynamicSharedInformerFactory
	// resources maps APIResource.Name to data for that resource,
	// and formerly only contains entries for non-namespaced resources
	resources map[string]*resourceResolver

	// maps GroupKind to Name of ksmetav1a1.APIResource
	gkToARName map[schema.GroupKind]string
}

// resourceResolver holds the data for a given resource in a given WDS
type resourceResolver struct {
	gvr      schema.GroupVersionResource
	informer upstreamcache.SharedInformer
	lister   upstreamcache.GenericLister
	stop     func()

	// byObjName maps object namespace (if namespaced) and name to relevant details
	byObjName map[NamespacedName]*objectDetails

	definers Set[ksmetav1a1.Definer]
}

type NamespacedName = Pair[NamespaceName, ObjectName]

// holds the data for a given object (necessarily in a particular WDS)
type objectDetails struct {
	// PlacementBits holds an entry for each EdgePlacement whose what predicate
	// matches the object.
	PlacementBits MutableMap[ObjectName, DistributionBits]
}

// NewWhatResolver returns a WhatResolver;
// invoke that function after the namespace informer has synced.
func NewWhatResolver(
	ctx context.Context,
	edgePlacementPreInformer edgev2alpha1informers.EdgePlacementInformer,
	spaceclient msclient.KubestellarSpaceInterface,
	spaceProviderNs string,
	kbSpaceRelation kbuser.KubeBindSpaceRelation,
	numThreads int,
) WhatResolver {
	controllerName := "what-resolver"
	logger := klog.FromContext(ctx).WithValues("part", controllerName)
	ctx = klog.NewContext(ctx, logger)
	wr := &whatResolver{
		ctx:                   ctx,
		logger:                logger,
		numThreads:            numThreads,
		queue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		edgePlacementInformer: edgePlacementPreInformer.Informer(),
		edgePlacementLister:   edgePlacementPreInformer.Lister(),
		spaceclient:           spaceclient,
		spaceProviderNs:       spaceProviderNs,
		kbSpaceRelation:       kbSpaceRelation,
		workspaceDetails:      map[string]*workspaceDetails{},
	}
	return func(receiver MappingReceiver[ExternalName, ResolvedWhat]) Runnable {
		wr.receiver = receiver
		wr.edgePlacementInformer.AddEventHandler(WhatResolverClusterHandler{wr, mkgk(edgeapi.SchemeGroupVersion.Group, "EdgePlacement")})
		if !upstreamcache.WaitForNamedCacheSync(controllerName, ctx.Done(), wr.edgePlacementInformer.HasSynced) {
			logger.Info("Failed to sync EdgePlacements in time")
		}
		return wr
	}
}

func (wr *whatResolver) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(wr.numThreads)
	for i := 0; i < wr.numThreads; i++ {
		go func() {
			wait.Until(wr.runWorker, time.Second, ctx.Done())
			wg.Done()
		}()
	}
	wg.Wait()
}

type namespacedQueueItem struct {
	GK      schema.GroupKind
	Cluster string
	NN      NamespacedName
}

type WhatResolverClusterHandler struct {
	*whatResolver
	gk schema.GroupKind
}

func (wrh WhatResolverClusterHandler) OnAdd(obj any) {
	wrh.enqueue(wrh.gk, obj)
}

func (wrh WhatResolverClusterHandler) OnUpdate(oldObj, newObj any) {
	wrh.enqueue(wrh.gk, newObj)
}

func (wrh WhatResolverClusterHandler) OnDelete(obj any) {
	wrh.enqueue(wrh.gk, obj)
}

func (wr *whatResolver) enqueue(gk schema.GroupKind, objAny any) {
	key, err := upstreamcache.DeletionHandlingMetaNamespaceKeyFunc(objAny)
	if err != nil {
		wr.logger.Error(err, "Failed to extract object reference", "object", objAny)
		return
	}
	namespace, name, err := upstreamcache.SplitMetaNamespaceKey(key)
	if err != nil {
		wr.logger.Error(err, "Impossible! SplitMetaClusterNamespaceKey failed", "key", key)
	}
	if namespace != "" {
		panic("Namespace must be empty here")
	}
	// Enqueuewith empty cluster. Set it later
	item := namespacedQueueItem{GK: gk, Cluster: "", NN: NewPair(NamespaceName(metav1.NamespaceNone), ObjectName(name))}
	wr.logger.V(4).Info("Enqueuing", "item", item)
	wr.queue.Add(item)
}

type WhatResolverScopedHandler struct {
	*whatResolver
	gk      schema.GroupKind
	cluster string
}

func (wrh WhatResolverScopedHandler) OnAdd(obj any) {
	wrh.enqueueScoped(wrh.gk, wrh.cluster, obj)
}

func (wrh WhatResolverScopedHandler) OnUpdate(oldObj, newObj any) {
	wrh.enqueueScoped(wrh.gk, wrh.cluster, newObj)
}

func (wrh WhatResolverScopedHandler) OnDelete(obj any) {
	wrh.enqueueScoped(wrh.gk, wrh.cluster, obj)
}

func (wr *whatResolver) enqueueScoped(gk schema.GroupKind, cluster string, objAny any) {
	key, err := upstreamcache.DeletionHandlingMetaNamespaceKeyFunc(objAny)
	if err != nil {
		wr.logger.Error(err, "Failed to extract object reference", "object", objAny)
		return
	}
	namespace, name, err := upstreamcache.SplitMetaNamespaceKey(key)
	if err != nil {
		wr.logger.Error(err, "Impossible! SplitMetaNamespaceKey failed", "key", key)
	}
	nn := NewPair(NamespaceName(namespace), ObjectName(name))
	item := namespacedQueueItem{GK: gk, Cluster: cluster, NN: nn}
	wr.logger.V(4).Info("Enqueuing", "item", item)
	wr.queue.Add(item)
}

func (wr *whatResolver) getPartsLocked(wldCluster string, epName ObjectName) ResolvedWhat {
	parts := WorkloadParts{}
	var upsyncs []edgeapi.UpsyncSet
	wsDetails, found := wr.workspaceDetails[wldCluster]
	var definers MutableSet[ksmetav1a1.Definer] = NewEmptyMapSet[ksmetav1a1.Definer]()
	if !found {
		return ResolvedWhat{parts, upsyncs}
	}
	if ep, found := wsDetails.placements[epName]; found {
		upsyncs = ep.Spec.Upsync
	}
	for _, rr := range wsDetails.resources {
		var gotSome bool
		for objName, objDetails := range rr.byObjName {
			// TODO: add index by EdgePlacement name to make this faster
			distrBits, found := objDetails.PlacementBits.Get(epName)
			if !found {
				continue
			}
			partID := NewTriple(SchemaGroupResourceToMeta(rr.gvr.GroupResource()), objName.First, objName.Second)
			partDetails := WorkloadPartDetails{APIVersion: rr.gvr.Version}.setDistributionBits(distrBits)
			wr.logger.V(4).Info("Returning part", "partID", partID, "partDetails", partDetails, "definers", rr.definers)
			parts[partID] = partDetails
			gotSome = true
		}
		if gotSome {
			SetAddAll[ksmetav1a1.Definer](definers, rr.definers)
		}
	}
	definers.Visit(func(definer ksmetav1a1.Definer) error {
		if definer.Kind == "APIBinding" && definer.Name == "espw" {
			// not neceesary to add, already in every mailbox workspace
			return nil
		}
		nsName := NamespaceName("")
		objName := ObjectName(definer.Name)
		pieces := definerKindToPieces[definer.Kind]
		definerID := NewTriple(pieces.First, nsName, objName)
		otherDetails, found := parts[definerID]
		if found {
			return nil
		}
		pieces.Second.ReturnSingletonState = pieces.Second.ReturnSingletonState || otherDetails.ReturnSingletonState
		parts[definerID] = pieces.Second
		wr.logger.V(4).Info("Implicitly adding definer", "definerID", definerID, "details", pieces.Second)
		return nil
	})
	return ResolvedWhat{parts, upsyncs}
}

var definerKindToPieces = map[string]Pair[metav1.GroupResource, WorkloadPartDetails]{
	"CustomResourceDefinition": NewPair(metav1.GroupResource{Group: apiext.GroupName, Resource: "customresourcedefinitions"}, WorkloadPartDetails{APIVersion: apiext.SchemeGroupVersion.Version}),
	"APIBinding":               NewPair(metav1.GroupResource{Group: apisv1alpha1.SchemeGroupVersion.Group, Resource: "apibindings"}, WorkloadPartDetails{APIVersion: apisv1alpha1.SchemeGroupVersion.Version}),
}

func (wr *whatResolver) notifyReceiversOfPlacements(cluster string, placements Set[ObjectName]) {
	placements.Visit(func(epName ObjectName) error {
		wr.notifyReceivers(cluster, epName)
		return nil
	})
}

func (wr *whatResolver) notifyReceivers(wldCluster string, epName ObjectName) {
	resolvedWhat := wr.getPartsLocked(wldCluster, epName)
	epRef := ExternalName{Cluster: wldCluster, Name: epName}
	wr.receiver.Put(epRef, resolvedWhat)
}

func (wr *whatResolver) runWorker() {
	for wr.processNextWorkItem() {
	}
}

func (wr *whatResolver) processNextWorkItem() bool {
	// Wait until there is a new item in the working queue
	itemAny, quit := wr.queue.Get()
	if quit {
		return false
	}
	defer wr.queue.Done(itemAny)
	item := itemAny.(namespacedQueueItem)

	logger := klog.FromContext(wr.ctx).WithValues("queueItem", item)
	ctx := klog.NewContext(wr.ctx, logger)
	logger.V(4).Info("Started processing queueItem")

	if wr.process(ctx, item) {
		logger.V(4).Info("Finished processing queueItem")
		wr.queue.Forget(itemAny)
	} else {
		logger.V(4).Info("Will retry processing queueItem")
		wr.queue.AddRateLimited(itemAny)
	}
	return true
}

// process returns true on success or unrecoverable error, false to retry
func (wr *whatResolver) process(ctx context.Context, item namespacedQueueItem) bool {
	if item.GK.Group == edgeapi.SchemeGroupVersion.Group && item.GK.Kind == "EdgePlacement" {
		return wr.processEdgePlacement(ctx, item.NN.Second)
	} else if item.GK.Group == ksmetav1a1.SchemeGroupVersion.Group && item.GK.Kind == "APIResource" {
		return wr.processResource(ctx, item.Cluster, string(item.NN.Second))
	} else {
		return wr.processCenterObject(ctx, item.Cluster, item.GK, item.NN)
	}
}

// process returns true on success or unrecoverable error, false to retry
func (wr *whatResolver) processCenterObject(ctx context.Context, cluster string, gk schema.GroupKind, objName NamespacedName) bool {
	logger := klog.FromContext(ctx)
	wr.Lock()
	defer wr.Unlock()
	wsDetails, detailsFound := wr.workspaceDetails[cluster]
	if !detailsFound {
		logger.V(3).Info("Ignoring notification about object in uninteresting cluster")
		return true
	}
	arName, found := wsDetails.gkToARName[gk]
	if !found {
		logger.Error(nil, "Impossible, processing object of unknown kind")
		return true
	}
	rr, found := wsDetails.resources[arName]
	if !found {
		logger.V(3).Info("Ignoring extinct kind of object")
		return true
	}
	var lister upstreamcache.GenericNamespaceLister = rr.lister
	if len(objName.First) > 0 {
		lister = rr.lister.ByNamespace(string(objName.First))
	}
	rObj, err := lister.Get(string(objName.Second))
	if err != nil {
		if !k8sapierrors.IsNotFound(err) {
			logger.Error(err, "Failed to fetch generic object from lister")
			return true
		}
		rObj = nil
	}
	oldDetails := rr.byObjName[objName]
	if oldDetails == nil {
		oldDetails = newObjectDetails()
	}
	var newDetails *objectDetails
	if rObj == nil {
		delete(rr.byObjName, objName)
		newDetails = newObjectDetails()
	} else {
		mrObj := rObj.(mrObject)
		newDetails = whatMatchingPlacements(logger, wsDetails, wsDetails.placements, rr.gvr.Resource, mrObj)
		if newDetails == nil {
			return false
		}
	}
	changedPlacements := NewMapMap[ObjectName, DistributionBits](nil)
	// TODO: make a way to not enumerate two when the bits change
	MapEnumerateDifferences[ObjectName, DistributionBits](newDetails.PlacementBits, oldDetails.PlacementBits, MappingReceiverDiscardsPrevious[ObjectName, DistributionBits](changedPlacements))
	MapEnumerateDifferences[ObjectName, DistributionBits](oldDetails.PlacementBits, newDetails.PlacementBits, MappingReceiverDiscardsPrevious[ObjectName, DistributionBits](changedPlacements))
	logger.V(4).Info("Processed object", "newDetails", newDetails, "changedPlacements", changedPlacements)
	if changedPlacements.IsEmpty() {
		return true
	}
	if rObj != nil {
		rr.byObjName[objName] = newDetails
	}
	wr.notifyReceiversOfPlacements(cluster, MapKeySet[ObjectName, DistributionBits](changedPlacements))
	return true
}

func newObjectDetails() *objectDetails {
	return &objectDetails{PlacementBits: NewMapMap[ObjectName, DistributionBits](nil)}
}

// process returns true on success or unrecoverable error, false to retry
func (wr *whatResolver) processResource(ctx context.Context, cluster string, arName string) bool {
	logger := klog.FromContext(ctx)
	wr.Lock()
	defer wr.Unlock()
	wsDetails, detailsFound := wr.workspaceDetails[cluster]
	if !detailsFound {
		logger.V(4).Info("Ignoring notification about resource in uninteresting cluster", "cluster", cluster, "arName", arName)
		return true
	}
	ar, err := wsDetails.apiLister.Get(arName)
	if err != nil {
		if !k8sapierrors.IsNotFound(err) {
			logger.Error(err, "Failed to lookup APIResource")
			return true
		}
		ar = nil
	}
	rr := wsDetails.resources[arName]
	// TODO: handle the case where ar.Spec changed
	if ar == nil {
		// APIResource does not exist
		if rr == nil { // no data for the resource
			logger.V(4).Info("Nothing to do for resource", "isNil", ar == nil, "isNamespaced", ar != nil && ar.Spec.Namespaced)
			return true
		}
		rr.stop()
		delete(wsDetails.resources, arName)
		changedPlacements := NewEmptyMapSet[ObjectName]()
		for _, objDetails := range rr.byObjName {
			SetAddAll[ObjectName](changedPlacements, MapKeySet[ObjectName, DistributionBits](objDetails.PlacementBits))
		}
		logger.V(4).Info("Removing resource", "changedPlacements", changedPlacements)
		wr.notifyReceiversOfPlacements(cluster, changedPlacements)
		return true
	}
	gvr := schema.GroupVersionResource{
		Group:    ar.Spec.Group,
		Version:  ar.Spec.Version,
		Resource: ar.Spec.Name,
	}
	gr := schema.GroupResource{
		Group:    ar.Spec.Group,
		Resource: ar.Spec.Name,
	}
	logger = logger.WithValues("gvr", gvr, "arName", arName, "definers", ar.Spec.Definers)
	if _, found := GRsNotSupported[gr]; found {
		logger.V(4).Info("Ignoring unsupported resource")
		return true
	}
	if _, found := GRsNotForEdge[gr]; found || GroupsNotForEdge.Has(gr.Group) {
		logger.V(4).Info("Ignoring resource that does not downsync")
		return true
	}
	gk := schema.GroupKind{Group: ar.Spec.Group, Kind: ar.Spec.Kind}
	logger = logger.WithValues("gk", gk)
	if rr == nil {
		informerCtx, stopInformer := context.WithCancel(wsDetails.ctx)
		preInformer := wsDetails.dynamicInformerFactory.ForResource(gvr)
		objInformer := preInformer.Informer()
		objInformer.AddEventHandler(WhatResolverScopedHandler{wr, gk, cluster})
		rr = &resourceResolver{
			gvr:       gvr,
			informer:  objInformer,
			lister:    preInformer.Lister(),
			stop:      stopInformer,
			byObjName: map[NamespacedName]*objectDetails{},
			definers:  NewSliceSet(ar.Spec.Definers...),
		}
		go rr.informer.Run(informerCtx.Done())
		logger.V(3).Info("Started to watch resource")
		wsDetails.resources[arName] = rr
		wsDetails.gkToARName[gk] = arName
	} else {
		var newDefiners Set[ksmetav1a1.Definer] = NewSliceSet(ar.Spec.Definers...)
		changed := !SetEqual(rr.definers, newDefiners)
		logger.V(4).Info("Continuing to watch resource", "changed", changed)
		if changed {
			rr.definers = newDefiners
			changedPlacements := NewEmptyMapSet[ObjectName]()
			for _, objDetails := range rr.byObjName {
				SetAddAll[ObjectName](changedPlacements, MapKeySet[ObjectName, DistributionBits](objDetails.PlacementBits))
			}
			wr.notifyReceiversOfPlacements(cluster, changedPlacements)
		}
	}
	return true
}

// process returns true on success or unrecoverable error, false to retry
func (wr *whatResolver) processEdgePlacement(ctx context.Context, epName ObjectName) bool {
	logger := klog.FromContext(ctx)
	ep, err := wr.edgePlacementLister.Get(string(epName))
	if err != nil {
		if !k8sapierrors.IsNotFound(err) {
			logger.Error(err, "Failed to fetch EdgePlacement from local cache")
			return true // I think these errors are not transient
		}
		ep = nil
	}
	epFound := err == nil
	//change to consumer SpaceID
	_, epOriginalName, kbSpaceID, err := kbuser.AnalyzeObjectID(ep)
	if err != nil {
		logger.Error(err, "Object does not appear to be a provider's copy of a consumer's object")
		return true
	}
	spaceID := wr.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if spaceID == "" {
		logger.Error(nil, "Failed to get consumer space ID from a provider's copy")
		return false
	}
	epName = ObjectName(epOriginalName)

	wr.Lock()
	defer wr.Unlock()
	wsDetails, wsDetailsFound := wr.workspaceDetails[spaceID]
	if !wsDetailsFound {
		if !epFound {
			logger.V(4).Info(`Both workspaceDetails and EdgePlacement were not found`)
			return true
		}
		config, err := wr.spaceclient.ConfigForSpace(spaceID, wr.spaceProviderNs)
		if err != nil {
			logger.Error(err, "Failed to get space config", "space", spaceID)
			return true
		}
		discoveryScopedClient, err := discovery.NewDiscoveryClientForConfig(config)
		if err != nil {
			logger.Error(err, "Failed to create discovery client", "space", spaceID)
			return true
		}
		scopedDynamic, err := dynamic.NewForConfig(config)
		if err != nil {
			logger.Error(err, "Failed to create dynamic client", "space", spaceID)
			return true
		}
		apiextClient, err := apiextclient.NewForConfig(config)
		if err != nil {
			logger.Error(err, "Failed to create clientset for CustomResourceDefinitions")
		}
		apiextFactory := apiextinfactory.NewSharedInformerFactory(apiextClient, 0)
		crdInformer := apiextFactory.Apiextensions().V1().CustomResourceDefinitions().Informer()
		kcpClientset, err := kcpclientset.NewForConfig(config)
		if err != nil {
			logger.Error(err, "Failed to create clientset for kcp APIs")
		}
		kcpInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(kcpClientset, 0)
		bindingInformer := kcpInformerFactory.Apis().V1alpha1().APIBindings().Informer()

		wsCtx, stopWS := context.WithCancel(wr.ctx)
		doneCh := wsCtx.Done()
		apiextFactory.Start(doneCh)
		kcpInformerFactory.Start(doneCh)

		apiInformer, apiLister, _ := apiwatch.NewAPIResourceInformer(wsCtx, spaceID, discoveryScopedClient, false,
			apiwatch.CRDAnalyzer{ObjectNotifier: crdInformer}, apiwatch.APIBindingAnalyzer{ObjectNotifier: bindingInformer})
		dynamicInformerFactory := kubedynamicinformer.NewDynamicSharedInformerFactory(scopedDynamic, 0)
		wsDetails = &workspaceDetails{
			ctx:                    wsCtx,
			placements:             map[ObjectName]*edgeapi.EdgePlacement{},
			stop:                   stopWS,
			apiInformer:            apiInformer,
			apiLister:              apiLister,
			dynamicInformerFactory: dynamicInformerFactory,
			resources:              map[string]*resourceResolver{},
			gkToARName:             map[schema.GroupKind]string{},
		}
		wr.workspaceDetails[spaceID] = wsDetails
		apiInformer.AddEventHandler(WhatResolverScopedHandler{wr, mkgk(ksmetav1a1.SchemeGroupVersion.Group, "APIResource"), spaceID})
		logger.V(2).Info("Started watching space")
		go apiInformer.Run(doneCh)
		dynamicInformerFactory.Start(doneCh)
		if !upstreamcache.WaitForCacheSync(doneCh, apiInformer.HasSynced) {
			logger.Error(nil, "Failed to sync API informer in time")
			return true
		}
	}
	if wsDetailsFound && !epFound {
		_, wasIncluded := wsDetails.placements[epName]
		if !wasIncluded {
			logger.V(4).Info(`Absent EdgePlacement is already irrelevant`)
			return true
		}
		delete(wsDetails.placements, epName)
		for _, rr := range wsDetails.resources {
			for objName, objDetails := range rr.byObjName {
				objDetails.PlacementBits.Delete(epName)
				if objDetails.PlacementBits.IsEmpty() {
					delete(rr.byObjName, objName)
				}
			}
		}
		logger.V(3).Info("Stopped watching EdgePlacement")
		if len(wsDetails.placements) == 0 {
			wsDetails.stop()
			logger.V(2).Info("Stopped watching space")
			delete(wr.workspaceDetails, spaceID)
		} else {
			wr.notifyReceivers(spaceID, epName)
		}
		wr.notifyReceivers(spaceID, epName)
		return true
	}
	// Now we know that ep != nil
	prevEp := wsDetails.placements[epName]
	wsDetails.placements[epName] = ep
	if prevEp == nil {
		logger.V(3).Info("Starting watching EdgePlacement")
	} else {
		whatPredicateUnChanged := apiequality.Semantic.DeepEqual(prevEp.Spec.Downsync, ep.Spec.Downsync)
		if whatPredicateUnChanged {
			logger.V(4).Info(`No change in "what" predicate`)
			return true
		}
	}
	anyChange := false
	completeSuccess := true
	for _, rr := range wsDetails.resources {
		logger := logger.WithValues("gvr", rr.gvr)
		rObjs, err := rr.lister.List(labels.Everything())
		if err != nil {
			logger.Error(err, "Failed to list objects", "gvr", rr.gvr)
		}
		for _, rObj := range rObjs {
			mrObj := rObj.(mrObject)
			objNS := NamespaceName(mrObj.GetNamespace())
			objName := ObjectName(mrObj.GetName())
			objNN := NewPair(objNS, objName)
			objDetails, found := rr.byObjName[objNN]
			if objDetails == nil {
				objDetails = newObjectDetails()
			}
			objChange, success := objDetails.setByMatch(logger, wsDetails, &ep.Spec, epName, rr.gvr.Resource, mrObj)
			logger.V(5).Info("From objDetails.setByMatch", "objNN", objNN, "found", found, "objChange", objChange, "success", success)
			if !success {
				completeSuccess = false
				continue
			}
			if objChange && !found {
				rr.byObjName[objNN] = objDetails
			}
			anyChange = anyChange || objChange
		}
	}
	logger.V(5).Info("Finished looping over resources", "numResources", len(wsDetails.resources))
	if anyChange {
		wr.notifyReceivers(spaceID, epName)
	}
	return completeSuccess
}

type mrObject interface {
	metav1.Object
	k8sruntime.Object
}

// Returns nil when an accurate answer cannot be computed.
func whatMatchingPlacements(logger klog.Logger, wsd *workspaceDetails, candidates map[ObjectName]*edgeapi.EdgePlacement, whatResource string, whatObj mrObject) *objectDetails {
	ans := newObjectDetails()
	for epName, ep := range candidates {
		_, success := ans.setByMatch(logger, wsd, &ep.Spec, epName, whatResource, whatObj)
		if !success {
			return nil
		}
	}
	return ans
}

// returns `(changed bool, success bool)`
func (od *objectDetails) setByMatch(logger klog.Logger, wsd *workspaceDetails, spec *edgeapi.EdgePlacementSpec, epName ObjectName, whatResource string, whatObj mrObject) (bool, bool) {
	oldDistrBits, found := od.PlacementBits.Get(epName)
	newDistrBits := DistributionBits{ReturnSingletonState: spec.WantSingletonReportedState,
		CreateOnly: whatObj != nil && isCreateOnly(whatObj)}
	objMatch, success := whatMatches(logger, wsd, spec, whatResource, whatObj)
	if !success {
		return false, false
	}
	if objMatch == found && (oldDistrBits == newDistrBits || !found) {
		return false, true
	}
	if objMatch {
		od.PlacementBits.Put(epName, newDistrBits)
	} else {
		od.PlacementBits.Delete(epName)
	}
	return true, true
}

func isCreateOnly(whatObj mrObject) bool {
	annotations := whatObj.GetAnnotations()
	if annotations == nil {
		return false
	}
	return annotations[edgeapi.DownsyncOverwriteKey] == "false"
}

// whatMatches tests the given object against the "what predicate" of an EdgePlacementSpec.
// The first returned bool indicates whether there is a match.
// The second indicates whether an accurate answer was found.
func whatMatches(logger klog.Logger, wsd *workspaceDetails, spec *edgeapi.EdgePlacementSpec, whatResource string, whatObj mrObject) (bool, bool) {
	if ObjectIsSystem(whatObj) {
		return false, true
	}
	gvk := whatObj.GetObjectKind().GroupVersionKind()
	objNS := whatObj.GetNamespace()
	objName := whatObj.GetName()
	objLabels := whatObj.GetLabels()
	match, ok := downsyncMatches(logger, wsd, spec.Downsync, whatResource, gvk, objNS, objName, objLabels)
	if !(match && ok) {
		return match, ok
	}
	return !upsyncMatches(logger, wsd, spec.Upsync, whatResource, gvk, objNS, objName, objLabels), true
}

func upsyncMatches(logger klog.Logger, wsd *workspaceDetails, upsync []edgeapi.UpsyncSet, whatResource string, gvk schema.GroupVersionKind, objNS, objName string, objLabels map[string]string) bool {
	for _, objTest := range upsync {
		if objTest.APIGroup != gvk.Group {
			continue
		}
		if !(SliceContains(objTest.Resources, "*") || SliceContains(objTest.Resources, whatResource)) {
			continue
		}
		if objNS != "" && !(SliceContains(objTest.Namespaces, "*") || SliceContains(objTest.Resources, objNS)) {
			continue
		}
		if !(SliceContains(objTest.Names, "*") || SliceContains(objTest.Resources, objName)) {
			continue
		}
		return true
	}
	return false
}

func downsyncMatches(logger klog.Logger, wsd *workspaceDetails, downsync []edgeapi.DownsyncObjectTest, whatResource string, gvk schema.GroupVersionKind, objNS, objName string, objLabels map[string]string) (bool, bool) {
	for _, objTest := range downsync {
		if objTest.APIGroup != nil && (*objTest.APIGroup) != gvk.Group {
			continue
		}
		if len(objTest.Resources) > 0 && !(SliceContains(objTest.Resources, "*") || SliceContains(objTest.Resources, whatResource)) {
			continue
		}
		if len(objTest.Namespaces) > 0 && !(SliceContains(objTest.Namespaces, "*") || SliceContains(objTest.Namespaces, objNS)) {
			continue
		}
		nsMatch, ok, retry := wsd.namespaceLabelsMatch(logger, objTest, objNS)
		if !ok {
			return retry, retry
		}
		if !nsMatch {
			continue
		}
		if len(objTest.ObjectNames) > 0 && !(SliceContains(objTest.ObjectNames, "*") || SliceContains(objTest.ObjectNames, objName)) {
			continue
		}
		if len(objTest.LabelSelectors) > 0 && !labelsMatchAny(logger, objLabels, objTest.LabelSelectors) {
			continue
		}
		return true, true
	}
	return false, true
}

// Returns match, ok, retry.
func (wsd *workspaceDetails) namespaceLabelsMatch(logger klog.Logger, objTest edgeapi.DownsyncObjectTest, objNS string) (bool, bool, bool) {
	if len(objTest.NamespaceSelectors) == 0 {
		return true, true, false
	}
	var nsLabels map[string]string
	if objNS == "" {
		nsLabels = map[string]string{}
	} else {
		nsARName := wsd.gkToARName[schema.GroupKind{Kind: "Namespace"}]
		nsRR := wsd.resources[nsARName]
		var objNSR k8sruntime.Object
		if nsRR == nil {
			logger.V(2).Info("Going around again because namespaces are not known yet", "objNSR", objNSR, "nsARName", nsARName)
			return false, false, false
		}
		var err error
		objNSR, err = nsRR.lister.Get(objNS)
		if err != nil && !k8sapierrors.IsNotFound(err) {
			logger.Error(err, "Impossible: failed to fetch namespace from Lister", "objNS", objNS)
			return false, true, false
		}
		if objNSR == nil || err != nil && k8sapierrors.IsNotFound(err) {
			logger.V(2).Info("Going around again because namespace is not known yet", "objNSR", objNSR)
			return false, false, true
		}
		objNSM := objNSR.(metav1.Object)
		nsLabels = objNSM.GetLabels()
	}
	return labelsMatchAny(logger, nsLabels, objTest.NamespaceSelectors), true, false
}

func labelsMatchAny(logger klog.Logger, labelSet map[string]string, selectors []metav1.LabelSelector) bool {
	for _, ls := range selectors {
		sel, err := metav1.LabelSelectorAsSelector(&ls)
		if err != nil {
			logger.Info("Failed to convert LabelSelector to labels.Selector", "ls", ls, "err", err)
			continue
		}
		if sel.Matches(labels.Set(labelSet)) {
			return true
		}
	}
	return false
}

func mkgk(group, kind string) schema.GroupKind {
	return schema.GroupKind{Group: group, Kind: kind}
}

func mkgr(group, resource string) schema.GroupResource {
	return schema.GroupResource{Group: group, Resource: resource}
}

var GRsForciblyDenatured = NewMapSet(
	mkgr("admissionregistration.k8s.io", "mutatingwebhookconfigurations"),
	mkgr("admissionregistration.k8s.io", "validatingwebhookconfigurations"),
	mkgr("apiregistration.k8s.io", "apiservices"),
	mkgr("flowcontrol.apiserver.k8s.io", "flowschemas"),
	mkgr("flowcontrol.apiserver.k8s.io", "prioritylevelconfigurations"),
	mkgr("rbac.authorization.k8s.io", "clusterroles"),
	mkgr("rbac.authorization.k8s.io", "clusterrolebindingss"),
	mkgr("rbac.authorization.k8s.io", "roles"),
	mkgr("rbac.authorization.k8s.io", "rolebindings"),
	mkgr("", "limitranges"),
	mkgr("", "resourcequotas"),
	mkgr("", "serviceaccounts"),
)

var GRsNaturedInBoth = NewMapSet(
	mkgr("apiextensions.k8s.io", "customresourcedefinitions"),
	mkgr("", "namespaces"),
)

var NaturedInCenterGoToMailbox = NewMapSet(
	mkgr("apis.kcp.io", "apibindings"),
)

var GRsNotSupported = NewMapSet(
	mkgr("apiresource.kcp.io", "apiresourceimports"),
	mkgr("apiresource.kcp.io", "negotiatedapiresources"),
	mkgr("apis.kcp.io", "apiconversions"),
)

var GroupsNotForEdge = k8ssets.NewString(
	"edge.kubestellar.io",
	"scheduling.kcp.io",
	"tenancy.kcp.io",
	"topology.kcp.io",
	"workload.kcp.io",
)

var GRsNotForEdge = NewMapSet(
	mkgr("apis.kcp.io", "apiexports"),
	mkgr("apis.kcp.io", "apiexportendpointslices"),
	mkgr("apis.kcp.io", "apiresourceschemas"),
	mkgr("apps", "controllerrevisions"),
	mkgr("authentication.k8s.io", "tokenreviews"),
	mkgr("authorization.k8s.io", "localsubjectaccessreviews"),
	mkgr("authorization.k8s.io", "selfsubjectaccessreviews"),
	mkgr("authorization.k8s.io", "selfsubjectrulesreviews"),
	mkgr("authorization.k8s.io", "subjectaccessreviews"),
	mkgr("certificates.k8s.io", "certificatesigningrequests"),
	mkgr("core.kcp.io", "logicalclusters"),
	mkgr("core.kcp.io", "shards"),
	mkgr("events.k8s.io", "events"),
	mkgr("", "bindings"),
	mkgr("", "componentstatuses"),
	mkgr("", "events"),
	mkgr("", "nodes"),
)

func DefaultResourceModes(mgr metav1.GroupResource) ResourceMode {
	sgr := MetaGroupResourceToSchema(mgr)
	builtin := strings.HasSuffix(sgr.Group, ".k8s.io") || !strings.Contains(sgr.Group, ".")
	switch {
	case GRsForciblyDenatured.Has(sgr):
		return ResourceMode{GoesToEdge, ForciblyDenatured, builtin}
	case GRsNaturedInBoth.Has(sgr):
		return ResourceMode{GoesToEdge, NaturallyNatured, builtin}
	case NaturedInCenterGoToMailbox.Has(sgr):
		return ResourceMode{GoesToMailbox, NaturallyNatured, builtin}
	case GRsNotSupported.Has(sgr):
		return ResourceMode{ErrorInCenter, NaturallyNatured, builtin}
	case GroupsNotForEdge.Has(sgr.Group) || GRsNotForEdge.Has(sgr):
		return ResourceMode{TolerateInCenter, NaturallyNatured, builtin}
	default:
		return ResourceMode{GoesToEdge, NaturalyDenatured, builtin}
	}
}

func MetaGroupResourceToSchema(gr metav1.GroupResource) schema.GroupResource {
	return schema.GroupResource{Group: gr.Group, Resource: gr.Resource}
}

func SchemaGroupResourceToMeta(sgr schema.GroupResource) metav1.GroupResource {
	return metav1.GroupResource{Group: sgr.Group, Resource: sgr.Resource}
}
