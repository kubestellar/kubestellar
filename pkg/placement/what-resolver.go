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
	"strings"
	"sync"
	"time"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	k8ssets "k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	kubedynamicinformer "k8s.io/client-go/dynamic/dynamicinformer"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	clusterdiscovery "github.com/kcp-dev/client-go/discovery"
	clusterdynamic "github.com/kcp-dev/client-go/dynamic"
	kcpinformers "github.com/kcp-dev/client-go/informers"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	urmetav1a1 "github.com/kcp-dev/edge-mc/pkg/apis/meta/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/apiwatch"
	edgev1alpha1informers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kcp-dev/edge-mc/pkg/client/listers/edge/v1alpha1"
)

type whatResolver struct {
	ctx        context.Context
	logger     klog.Logger
	numThreads int
	queue      workqueue.RateLimitingInterface
	receiver   MappingReceiver[ExternalName, WorkloadParts]

	sync.Mutex

	edgePlacementInformer kcpcache.ScopeableSharedIndexInformer
	edgePlacementLister   edgev1alpha1listers.EdgePlacementClusterLister

	discoveryClusterClient    clusterdiscovery.DiscoveryClusterInterface
	crdClusterPreInformer     kcpinformers.GenericClusterInformer
	bindingClusterPreInformer kcpinformers.GenericClusterInformer
	dynamicClusterClient      clusterdynamic.ClusterInterface

	// workspaceDetails maps lc.Name of a workload LC to all the relevant information for that LC.
	workspaceDetails map[logicalcluster.Name]*workspaceDetails
}

type workspaceDetails struct {
	ctx context.Context
	// placements maps name of relevant EdgePlacement object to that object
	placements             map[string]*edgeapi.EdgePlacement
	stop                   func()
	apiInformer            upstreamcache.SharedInformer
	apiLister              apiwatch.APIResourceLister
	dynamicInformerFactory kubedynamicinformer.DynamicSharedInformerFactory
	// resources maps APIResource.Name to data for that resource,
	// and only contains entries for non-namespaced resources
	resources  map[string]*resourceResolver
	gkToARName map[schema.GroupKind]string
}

type resourceResolver struct {
	gvr      schema.GroupVersionResource
	informer upstreamcache.SharedInformer
	lister   upstreamcache.GenericLister
	stop     func()

	// byObjName maps object name to relevant details
	byObjName map[string]*objectDetails
}

type objectDetails struct {
	placements              k8ssets.String
	placementsWantNamespace k8ssets.String // non-nil only for Namespace objects
}

// NewWhatResolver returns a WhatResolver;
// invoke that function after the namespace informer has synced.
func NewWhatResolver(
	ctx context.Context,
	edgePlacementPreInformer edgev1alpha1informers.EdgePlacementClusterInformer,
	discoveryClusterClient clusterdiscovery.DiscoveryClusterInterface,
	crdClusterPreInformer kcpinformers.GenericClusterInformer,
	bindingClusterPreInformer kcpinformers.GenericClusterInformer,
	dynamicClusterClient clusterdynamic.ClusterInterface,
	numThreads int,
) WhatResolver {
	controllerName := "what-resolver"
	logger := klog.FromContext(ctx).WithValues("part", controllerName)
	ctx = klog.NewContext(ctx, logger)
	wr := &whatResolver{
		ctx:                       ctx,
		logger:                    logger,
		numThreads:                numThreads,
		queue:                     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		edgePlacementInformer:     edgePlacementPreInformer.Informer(),
		edgePlacementLister:       edgePlacementPreInformer.Lister(),
		discoveryClusterClient:    discoveryClusterClient,
		crdClusterPreInformer:     crdClusterPreInformer,
		bindingClusterPreInformer: bindingClusterPreInformer,
		dynamicClusterClient:      dynamicClusterClient,
		workspaceDetails:          map[logicalcluster.Name]*workspaceDetails{},
	}
	return func(receiver MappingReceiver[ExternalName, WorkloadParts]) Runnable {
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

type queueItem struct {
	gk      schema.GroupKind
	cluster logicalcluster.Name
	name    string
}

func (qi queueItem) toExternalName() ExternalName {
	return ExternalName{Cluster: qi.cluster, Name: qi.name}
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
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(objAny)
	if err != nil {
		wr.logger.Error(err, "Failed to extract object reference", "object", objAny)
		return
	}
	cluster, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(key)
	if err != nil {
		wr.logger.Error(err, "Impossible! SplitMetaClusterNamespaceKey failed", "key", key)
	}
	item := queueItem{gk: gk, cluster: cluster, name: name}
	wr.logger.V(4).Info("Enqueuing", "item", item)
	wr.queue.Add(item)
}

type WhatResolverScopedHandler struct {
	*whatResolver
	gk      schema.GroupKind
	cluster logicalcluster.Name
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

func (wr *whatResolver) enqueueScoped(gk schema.GroupKind, cluster logicalcluster.Name, objAny any) {
	key, err := upstreamcache.DeletionHandlingMetaNamespaceKeyFunc(objAny)
	if err != nil {
		wr.logger.Error(err, "Failed to extract object reference", "object", objAny)
		return
	}
	_, name, err := upstreamcache.SplitMetaNamespaceKey(key)
	if err != nil {
		wr.logger.Error(err, "Impossible! SplitMetaNamespaceKey failed", "key", key)
	}
	item := queueItem{gk: gk, cluster: cluster, name: name}
	wr.logger.V(4).Info("Enqueuing", "item", item)
	wr.queue.Add(item)
}

func (wr *whatResolver) Get(placement ExternalName, kont func(WorkloadParts)) {
	wr.Lock()
	defer wr.Unlock()
	parts := wr.getPartsLocked(placement.Cluster, placement.Name)
	kont(parts)
}

func (wr *whatResolver) getPartsLocked(wldCluster logicalcluster.Name, epName string) WorkloadParts {
	parts := WorkloadParts{}
	wsDetails, found := wr.workspaceDetails[wldCluster]
	if !found {
		return parts
	}
	for _, rr := range wsDetails.resources {
		for objName, objDetails := range rr.byObjName {
			// TODO: add index by EdgePlacement name to make this faster
			if _, found := objDetails.placements[epName]; !found {
				continue
			}
			_, wantNamespace := objDetails.placementsWantNamespace[epName]
			partID := WorkloadPartID{APIGroup: rr.gvr.Group, Resource: rr.gvr.Resource, Name: objName}
			partDetails := WorkloadPartDetails{APIVersion: rr.gvr.Version, IncludeNamespaceObject: wantNamespace}
			parts[partID] = partDetails
		}
	}
	return parts
}

func (wr *whatResolver) notifyReceiversOfPlacements(cluster logicalcluster.Name, placements k8ssets.String) {
	for epName := range placements {
		wr.notifyReceivers(cluster, epName)
	}
}

func (wr *whatResolver) notifyReceivers(wldCluster logicalcluster.Name, epName string) {
	parts := wr.getPartsLocked(wldCluster, epName)
	epRef := ExternalName{Cluster: wldCluster, Name: epName}
	wr.receiver.Put(epRef, parts)
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
	item := itemAny.(queueItem)

	logger := klog.FromContext(wr.ctx).WithValues("group", item.gk.Group, "kind", item.gk.Kind, "cluster", item.cluster, "name", item.name)
	ctx := klog.NewContext(wr.ctx, logger)
	logger.V(4).Info("processing queueItem")

	if wr.process(ctx, item) {
		wr.queue.Forget(itemAny)
	} else {
		wr.queue.AddRateLimited(itemAny)
	}
	return true
}

// process returns true on success or unrecoverable error, false to retry
func (wr *whatResolver) process(ctx context.Context, item queueItem) bool {
	if item.gk.Group == edgeapi.SchemeGroupVersion.Group && item.gk.Kind == "EdgePlacement" {
		return wr.processEdgePlacement(ctx, item.cluster, item.name)
	} else if item.gk.Group == urmetav1a1.SchemeGroupVersion.Group && item.gk.Kind == "APIResource" {
		return wr.processResource(ctx, item.cluster, item.name)
	} else {
		return wr.processObject(ctx, item.cluster, item.gk, item.name)
	}
}

func (wr *whatResolver) processObject(ctx context.Context, cluster logicalcluster.Name, gk schema.GroupKind, objName string) bool {
	logger := klog.FromContext(ctx)
	isNamespace := gkIsNamespace(gk)
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
	rObj, err := rr.lister.Get(objName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Failed to fetch generic object from lister")
		return true
	}
	oldDetails := rr.byObjName[objName]
	if oldDetails == nil {
		oldDetails = newObjectDetails(isNamespace)
	}
	var newDetails *objectDetails
	if rObj == nil {
		delete(rr.byObjName, objName)
		newDetails = newObjectDetails(isNamespace)
	} else {
		mrObj := rObj.(mrObject)
		newDetails = whatMatchingPlacements(logger, wsDetails.placements, rr.gvr.Resource, mrObj)
	}
	changedPlacements := newDetails.placements.Difference(oldDetails.placements)
	if isNamespace {
		changedPlacements = changedPlacements.Union(newDetails.placementsWantNamespace.Difference(oldDetails.placementsWantNamespace))
	}
	logger.V(4).Info("Processed object", "newDetails", newDetails, "changedPlacements", changedPlacements)
	if len(changedPlacements) == 0 {
		return true
	}
	if rObj != nil {
		rr.byObjName[objName] = newDetails
	}
	wr.notifyReceiversOfPlacements(cluster, changedPlacements)
	return true
}

func newObjectDetails(isNamespace bool) *objectDetails {
	ans := &objectDetails{placements: k8ssets.NewString()}
	if isNamespace {
		ans.placementsWantNamespace = k8ssets.NewString()
	}
	return ans
}

func (wr *whatResolver) processResource(ctx context.Context, cluster logicalcluster.Name, arName string) bool {
	logger := klog.FromContext(ctx)
	wr.Lock()
	defer wr.Unlock()
	wsDetails, detailsFound := wr.workspaceDetails[cluster]
	if !detailsFound {
		logger.V(4).Info("Ignoring notification about resource in uninteresting cluster", "cluster", cluster, "arName", arName)
		return true
	}
	ar, err := wsDetails.apiLister.Get(arName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Failed to lookup APIResource")
		return true
	}
	rr := wsDetails.resources[arName]
	// TODO: handle the case where ar.Spec changed
	if ar == nil || ar.Spec.Namespaced {
		// APIResource does not exist or is uninteresting
		if rr == nil { // no data for the resource
			logger.V(4).Info("Nothing to do for resource", "isNil", ar == nil, "isNamespaced", ar != nil && ar.Spec.Namespaced)
			return true
		}
		rr.stop()
		delete(wsDetails.resources, arName)
		changedPlacements := k8ssets.NewString()
		for _, objDetails := range rr.byObjName {
			changedPlacements = changedPlacements.Union(objDetails.placements)
			if gvrIsNamespace(rr.gvr) {
				changedPlacements = changedPlacements.Union(objDetails.placementsWantNamespace)
			}
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
	if _, found := GRsNotSupported[gr]; found {
		logger.V(4).Info("Ignoring unsupported resource")
		return true
	}
	if _, found := GRsNotForEdge[gr]; found || GroupsNotForEdge.Has(gr.Group) {
		logger.V(4).Info("Ignoring resource that does not downsync")
		return true
	}
	gk := schema.GroupKind{Group: ar.Spec.Group, Kind: ar.Spec.Kind}
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
			byObjName: map[string]*objectDetails{},
		}
		go rr.informer.Run(informerCtx.Done())
		logger.V(3).Info("Started watching resource")
		wsDetails.resources[arName] = rr
		wsDetails.gkToARName[gk] = arName
	} else {
		logger.V(4).Info("Continuing to watch resource")
	}
	return true
}

func (wr *whatResolver) processEdgePlacement(ctx context.Context, cluster logicalcluster.Name, epName string) bool {
	logger := klog.FromContext(ctx)
	ep, err := wr.edgePlacementLister.Cluster(cluster).Get(epName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Failed to fetch EdgePlacement from local cache", "cluster", cluster, "epName", epName)
		return true // I think these errors are not transient
	}
	epFound := err == nil
	wr.Lock()
	defer wr.Unlock()
	wsDetails, wsDetailsFound := wr.workspaceDetails[cluster]
	if !wsDetailsFound {
		if !epFound {
			logger.V(4).Info(`Both workspaceDetails and EdgePlacement were not found`)
			return true
		}
		wsCtx, stopWS := context.WithCancel(wr.ctx)
		discoveryScopedClient := wr.discoveryClusterClient.Cluster(cluster.Path())
		crdInformer := wr.crdClusterPreInformer.Cluster(cluster).Informer()
		bindingInformer := wr.bindingClusterPreInformer.Cluster(cluster).Informer()
		apiInformer, apiLister, _ := apiwatch.NewAPIResourceInformer(wsCtx, cluster.String(), discoveryScopedClient, crdInformer, bindingInformer)
		scopedDynamic := wr.dynamicClusterClient.Cluster(cluster.Path())
		dynamicInformerFactory := kubedynamicinformer.NewDynamicSharedInformerFactory(scopedDynamic, 0)
		wsDetails = &workspaceDetails{
			ctx:                    wsCtx,
			placements:             map[string]*edgeapi.EdgePlacement{},
			stop:                   stopWS,
			apiInformer:            apiInformer,
			apiLister:              apiLister,
			dynamicInformerFactory: dynamicInformerFactory,
			resources:              map[string]*resourceResolver{},
			gkToARName:             map[schema.GroupKind]string{},
		}
		wr.workspaceDetails[cluster] = wsDetails
		apiInformer.AddEventHandler(WhatResolverScopedHandler{wr, mkgk(urmetav1a1.SchemeGroupVersion.Group, "APIResource"), cluster})
		logger.V(2).Info("Started watching logical cluster")
		go apiInformer.Run(wsCtx.Done())
		dynamicInformerFactory.Start(wsCtx.Done())
		if !upstreamcache.WaitForCacheSync(wsCtx.Done(), apiInformer.HasSynced) {
			logger.Error(nil, "Failed to sync API informer in time", "cluster", cluster)
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
				delete(objDetails.placements, epName)
				if objDetails.placementsWantNamespace != nil {
					delete(objDetails.placementsWantNamespace, epName)
				}
				if len(objDetails.placements) == 0 {
					delete(rr.byObjName, objName)
				}
			}
		}
		logger.V(3).Info("Stopped watching EdgePlacement")
		if len(wsDetails.placements) == 0 {
			wsDetails.stop()
			logger.V(2).Info("Stopped watching logical cluster")
			delete(wr.workspaceDetails, cluster)
		} else {
			wr.notifyReceivers(cluster, epName)
		}
		wr.notifyReceivers(cluster, epName)
		return true
	}
	// Now we know that ep != nil
	prevEp := wsDetails.placements[epName]
	wsDetails.placements[epName] = ep
	if prevEp == nil {
		logger.V(3).Info("Starting watching EdgePlacement")
	} else {
		whatPredicateUnChanged := (apiequality.Semantic.DeepEqual(prevEp.Spec.NamespaceSelector, ep.Spec.NamespaceSelector) &&
			apiequality.Semantic.DeepEqual(prevEp.Spec.NonNamespacedObjects, ep.Spec.NonNamespacedObjects))
		if whatPredicateUnChanged {
			logger.V(4).Info(`No change in "what" predicate`)
			return true
		}
	}
	anyChange := false
	for _, rr := range wsDetails.resources {
		isNamespace := gvrIsNamespace(rr.gvr)
		rObjs, err := rr.lister.List(labels.Everything())
		if err != nil {
			logger.Error(err, "Failed to list objects", "gvr", rr.gvr)
		}
		for _, rObj := range rObjs {
			mrObj := rObj.(mrObject)
			objName := mrObj.GetName()
			objDetails, found := rr.byObjName[objName]
			if objDetails == nil {
				objDetails = newObjectDetails(isNamespace)
			}
			objChange := objDetails.setByMatch(logger, &ep.Spec, epName, isNamespace, rr.gvr.Resource, mrObj)
			if objChange && !found {
				rr.byObjName[objName] = objDetails
			}
			anyChange = anyChange || objChange
		}
	}
	if anyChange {
		wr.notifyReceivers(cluster, epName)
	}
	return true
}

type mrObject interface {
	metav1.Object
	k8sruntime.Object
}

func whatMatchingPlacements(logger klog.Logger, candidates map[string]*edgeapi.EdgePlacement, whatResource string, whatObj mrObject) *objectDetails {
	gvk := whatObj.GetObjectKind().GroupVersionKind()
	isNamespace := gkIsNamespace(gvk.GroupKind())
	ans := newObjectDetails(isNamespace)
	for epName, ep := range candidates {
		ans.setByMatch(logger, &ep.Spec, epName, isNamespace, whatResource, whatObj)
	}
	return ans
}

func (od *objectDetails) setByMatch(logger klog.Logger, spec *edgeapi.EdgePlacementSpec, epName string, isNamespace bool, whatResource string, whatObj mrObject) bool {
	_, found := od.placements[epName]
	if isNamespace {
		objMatch, nsMatch := whatMatches(logger, spec, whatResource, whatObj)
		var found2 bool
		if isNamespace {
			_, found2 = od.placementsWantNamespace[epName]
		}
		if !nsMatch {
			if !(found || found2) {
				return false
			}
			delete(od.placements, epName)
			if isNamespace {
				delete(od.placementsWantNamespace, epName)
			}
			return true
		}
		od.placements.Insert(epName)
		if objMatch {
			od.placementsWantNamespace.Insert(epName)
		}
		return (!found) || objMatch && !found2
	}
	objMatch, _ := whatMatches(logger, spec, whatResource, whatObj)
	if objMatch {
		if !found {
			od.placements.Insert(epName)
			return true
		}
	} else if found {
		delete(od.placements, epName)
		return true
	}
	return false
}

// whatMatches tests the given object against the "what predicate" of an EdgePlacementSpec.
// The first returned bool indicates whether the given object matches the NonNamespacedObjects part.
// The first returned bool indicates whether the given object is a Namespace and matches the NamespaceSelector part.
func whatMatches(logger klog.Logger, spec *edgeapi.EdgePlacementSpec, whatResource string, whatObj mrObject) (bool, bool) {
	gvk := whatObj.GetObjectKind().GroupVersionKind()
	objName := whatObj.GetName()
	labelSet := labels.Set(whatObj.GetLabels())
	matches := false
outer:
	for _, objSet := range spec.NonNamespacedObjects {
		if objSet.APIGroup != gvk.Group {
			continue
		}
		if !(len(objSet.Resources) == 1 && objSet.Resources[0] == "*" || SliceContains(objSet.Resources, whatResource)) {
			continue
		}
		if len(objSet.ResourceNames) == 1 && objSet.ResourceNames[9] == "*" || SliceContains(objSet.ResourceNames, objName) {
			matches = true
			break outer
		}
		for _, ls := range objSet.LabelSelectors {
			sel, err := metav1.LabelSelectorAsSelector(&ls)
			if err != nil {
				logger.Info("Failed to convert LabelSelector to labels.Selector", "ls", ls, "err", err)
				continue
			}
			if sel.Matches(labelSet) {
				matches = true
				break outer
			}
		}
	}
	if gkIsNamespace(gvk.GroupKind()) {
		sel, err := metav1.LabelSelectorAsSelector(&spec.NamespaceSelector)
		if err != nil {
			logger.Info("Failed to convert spec.NamespaceSelector to labels.Selector", "nsSelector", spec.NamespaceSelector, "err", err)
			return false, false
		}
		return matches, sel.Matches(labelSet)
	} else {
		return matches, false
	}
}

func gkIsNamespace(gk schema.GroupKind) bool {
	return gk.Group == "" && gk.Kind == "Namespace"
}

func gvrIsNamespace(gr schema.GroupVersionResource) bool {
	return gr.Group == "" && gr.Resource == "namespaces"
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

var NaturedInCenterNoGo = NewMapSet(
	mkgr("apis.kcp.io", "apibindings"),
)

var GRsNotSupported = NewMapSet(
	mkgr("apiregistration.k8s.io", "apiservices"),
	mkgr("apiresource.kcp.io", "apiresourceimports"),
	mkgr("apiresource.kcp.io", "negotiatedapiresources"),
	mkgr("apis.kcp.io", "apiconversions"),
)

var GroupsNotForEdge = k8ssets.NewString(
	"edge.kcp.io",
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
	case NaturedInCenterNoGo.Has(sgr):
		return ResourceMode{TolerateInCenter, NaturallyNatured, builtin}
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
