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

package binding

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	clusterclientset "open-cluster-management.io/api/client/cluster/clientset/versioned"
	clusterpkginformers "open-cluster-management.io/api/client/cluster/informers/externalversions"
	clusterinformers "open-cluster-management.io/api/client/cluster/informers/externalversions/cluster/v1"
	clusterlisters "open-cluster-management.io/api/client/cluster/listers/cluster/v1"
	managedclusterapi "open-cluster-management.io/api/cluster/v1"

	k8scoreapi "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/crd"
	ksclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	controlclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned/typed/control/v1alpha1"
	ksinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions"
	controlinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions/control/v1alpha1"
	controllisters "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
	ksmetrics "github.com/kubestellar/kubestellar/pkg/metrics"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const (
	ControllerName      = "Binding"
	defaultResyncPeriod = time.Duration(0)
)

// Resource groups to exclude for watchers as they should not be delivered to other clusters
var excludedGroups = map[string]bool{
	"flowcontrol.apiserver.k8s.io": true,
	"discovery.k8s.io":             true,
	"apiregistration.k8s.io":       true,
	"coordination.k8s.io":          true,
	"control.kubestellar.io":       true,
}

// Resource names to exclude for watchers as they should not delivered to other clusters
// TODO - add also group version to qualify and avoid filtering when same names used on
// user-supplied CRDs
var excludedResourceNames = map[string]bool{
	"events":               true,
	"nodes":                true,
	"csistoragecapacities": true,
	"csinodes":             true,
	"endpoints":            true,
	"workstatuses":         true,
}

const (
	bindingQueueingDelay = 2 * time.Second
	// https://github.com/kubernetes/kubernetes/blob/5d527dcf1265d7fcd0e6c8ec511ce16cc6a40699/staging/src/k8s.io/cli-runtime/pkg/genericclioptions/config_flags.go#L477
	referenceBurstUpperBound = 300
	// https://github.com/kubernetes/kubernetes/pull/105520/files
	referenceQPSUpperBound = 50.0
)

// Controller watches all objects, finds associated bindingpolicies, when matched a bindingpolicy wraps and
// places objects into mailboxes
type Controller struct {
	logger                      logr.Logger
	bindingPolicyClient         ksmetrics.ClientModNamespace[*v1alpha1.BindingPolicy, *v1alpha1.BindingPolicyList]
	bindingClient               ksmetrics.ClientModNamespace[*v1alpha1.Binding, *v1alpha1.BindingList]
	ksInformerFactoryStart      func(stopCh <-chan struct{})
	bindingInformer             cache.SharedIndexInformer
	bindingLister               controllisters.BindingLister
	bindingPolicyInformer       cache.SharedIndexInformer
	bindingPolicyLister         controllisters.BindingPolicyLister
	managedClusterClient        ksmetrics.ClientModNamespace[*managedclusterapi.ManagedCluster, *managedclusterapi.ManagedClusterList]
	clusterInformerFactoryStart func(stopCh <-chan struct{})
	clusterInformer             cache.SharedIndexInformer // used for ManagedCluster in ITS
	clusterLister               clusterlisters.ManagedClusterLister
	dynamicClient               dynamic.Interface // used for workload

	discoveryClient discovery.DiscoveryInterface                                                   // for WDS
	namespaceClient ksmetrics.ClientModNamespace[*k8scoreapi.Namespace, *k8scoreapi.NamespaceList] // for WDS

	extClient ksmetrics.ClientModNamespace[*apiextensionsv1.CustomResourceDefinition, *apiextensionsv1.CustomResourceDefinitionList] // for CRDs in WDS

	apiResourceLists []*metav1.APIResourceList
	listers          util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister]
	informers        util.ConcurrentMap[schema.GroupVersionResource, cache.SharedIndexInformer]
	stoppers         util.ConcurrentMap[schema.GroupVersionResource, chan struct{}]

	bindingPolicyResolver BindingPolicyResolver

	// Contains bindingPolicyRef, bindingRef, util.ObjectIdentifier
	workqueue        workqueue.RateLimitingInterface
	initializedTs    time.Time
	wdsName          string
	allowedGroupsSet sets.Set[string]
}

// bindingPolicyRef is a workqueue item that references a BindingPolicy
type bindingPolicyRef string

// bindingRef is a workqueue item that references a Binding
type bindingRef string

// Create a new binding controller
func NewController(parentLogger logr.Logger,
	wdsClientMetrics, itsClientMetrics ksmetrics.ClientMetrics,
	wdsRestConfig *rest.Config, itsRestConfig *rest.Config,
	wdsName string, allowedGroupsSet sets.Set[string]) (*Controller, error) {
	logger := parentLogger.WithName(ControllerName)

	kubernetesClient, err := kubernetes.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	extClient, err := apiextensionsclientset.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	// do the discovery, save the results, then discard the 'depleted' client
	disposableConfig := rest.CopyConfig(wdsRestConfig)
	disposableConfig.Burst = referenceBurstUpperBound
	disposableConfig.QPS = referenceQPSUpperBound
	disposableConfig.RateLimiter = nil
	disposableClient, err := kubernetes.NewForConfig(disposableConfig)
	if err != nil {
		return nil, err
	}
	apiResourceLists, nGVRs, err := doDiscovery(logger, disposableClient)
	if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
		return nil, err
	}

	// tuning the rate limiter based on the number of GVRs is tested to be working well
	wdsRestConfigTuned := rest.CopyConfig(wdsRestConfig)
	wdsRestConfigTuned.Burst = computeBurstFromNumGVRs(nGVRs)
	wdsRestConfigTuned.QPS = computeQPSFromNumGVRs(nGVRs)
	wdsRestConfigTuned.RateLimiter = nil
	logger.V(1).Info("Parameters of the tuned client's token bucket rate limiter", "burst", wdsRestConfigTuned.Burst, "qps", wdsRestConfigTuned.QPS)

	// baseDynamicClient needs higher rate than its default because dynamicClient is repeatedly used by the
	// reflectors for each of the GVRs, all at the beginning of the controller run
	baseDynamicClient, err := dynamic.NewForConfig(wdsRestConfigTuned)
	if err != nil {
		return nil, err
	}
	dynamicClient := ksmetrics.NewWrappedDynamicClient(wdsClientMetrics, baseDynamicClient)

	ksClient, err := ksclient.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}
	ksInformerFactory := ksinformers.NewSharedInformerFactory(ksClient, defaultResyncPeriod)

	clusterClient, err := clusterclientset.NewForConfig(itsRestConfig)
	if err != nil {
		return nil, err
	}
	clusterInformerFactory := clusterpkginformers.NewSharedInformerFactory(clusterClient, defaultResyncPeriod)

	return makeController(logger, wdsClientMetrics, itsClientMetrics, ksClient.ControlV1alpha1(), ksInformerFactory.Start, ksInformerFactory.Control().V1alpha1(), dynamicClient, kubernetesClient, extClient, clusterClient, clusterInformerFactory.Start, clusterInformerFactory.Cluster().V1().ManagedClusters(), apiResourceLists, wdsName, allowedGroupsSet)
}

// doDiscovery contains the exact one occurence of ServerPreferredResources() in this repository.
// doDiscovery is supposed to be invoked exactly one time during the lifecycle of a binding controller.
// That is, full API discovery against the WDS is done exactly one time during the lifecycle of a binding controller.
// doDiscovery also returns the number of successfully discovered GVRs.
// We do these to optimize the performance of the binding controller, especially when it runs against a WDS which
// (a) uses legacy discovery;
// (b) has large number of api groups (e.g. because of large number of CRDs).
func doDiscovery(logger logr.Logger, client kubernetes.Interface) ([]*metav1.APIResourceList, int, error) {
	dc := client.Discovery().(*discovery.DiscoveryClient)
	if dc.UseLegacyDiscovery { // by default it should be false already, just double check
		dc.UseLegacyDiscovery = false
	}
	apiResourceLists, err := dc.ServerPreferredResources()
	logger.Info("Discovery", "numAPIResourceLists", len(apiResourceLists), "err", err)
	n := 0
	for _, list := range apiResourceLists {
		n += len(list.APIResources)
	}
	return apiResourceLists, n, err
}

func computeBurstFromNumGVRs(nGVRs int) int {
	burst := nGVRs
	// in case too small, fall back to default
	if burst < rest.DefaultBurst {
		return rest.DefaultBurst
	}
	// in case too large, look at some value for reference
	if burst > referenceBurstUpperBound {
		return referenceBurstUpperBound
	}
	return burst
}

func computeQPSFromNumGVRs(nGVRs int) float32 {
	qps := float32(nGVRs) / 4
	// in case too small, fall back to default
	if qps < rest.DefaultQPS {
		return rest.DefaultQPS
	}
	// in case too large, look at some value for reference
	if qps > referenceQPSUpperBound {
		return referenceQPSUpperBound
	}
	return qps
}

func makeController(logger logr.Logger,
	wdsClientMetrics, itsClientMetrics ksmetrics.ClientMetrics,
	controlClient controlclient.ControlV1alpha1Interface,
	ksInformerFactoryStart func(stopCh <-chan struct{}),
	controlInformers controlinformers.Interface,
	dynamicClient dynamic.Interface, // used for CRD, Binding[Policy], workload
	kubernetesClient kubernetes.Interface, // used for Namespaces, and Discovery
	extClient apiextensionsclientset.Interface, // used for CRD
	clusterClient clusterclientset.Interface, // used for ManagedCluster in ITS
	clusterInformerFactoryStart func(<-chan struct{}),
	clusterPreInformer clusterinformers.ManagedClusterInformer, // used for ManagedCluster in ITS
	apiResourceLists []*metav1.APIResourceList,
	wdsName string, allowedGroupsSet sets.Set[string]) (*Controller, error) {

	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	clusterInformer := clusterPreInformer.Informer()
	controller := &Controller{
		wdsName:                     wdsName,
		logger:                      logger,
		bindingPolicyClient:         ksmetrics.NewWrappedClusterScopedClient(wdsClientMetrics, util.GetBindingPolicyGVR(), controlClient.BindingPolicies()),
		bindingClient:               ksmetrics.NewWrappedClusterScopedClient(wdsClientMetrics, util.GetBindingGVR(), controlClient.Bindings()),
		ksInformerFactoryStart:      ksInformerFactoryStart,
		bindingInformer:             controlInformers.Bindings().Informer(),
		bindingLister:               controlInformers.Bindings().Lister(),
		bindingPolicyInformer:       controlInformers.BindingPolicies().Informer(),
		bindingPolicyLister:         controlInformers.BindingPolicies().Lister(),
		managedClusterClient:        ksmetrics.NewWrappedClusterScopedClient[*managedclusterapi.ManagedCluster, *managedclusterapi.ManagedClusterList](itsClientMetrics, managedclusterapi.SchemeGroupVersion.WithResource("manageclusters"), clusterClient.ClusterV1().ManagedClusters()),
		clusterInformerFactoryStart: clusterInformerFactoryStart,
		clusterInformer:             clusterInformer,
		clusterLister:               clusterPreInformer.Lister(),
		dynamicClient:               dynamicClient,
		discoveryClient:             kubernetesClient.Discovery(),
		namespaceClient:             ksmetrics.NewWrappedClusterScopedClient(wdsClientMetrics, k8scoreapi.SchemeGroupVersion.WithResource("namespaces"), kubernetesClient.CoreV1().Namespaces()),
		extClient:                   ksmetrics.NewWrappedClusterScopedClient(wdsClientMetrics, apiextensionsv1.SchemeGroupVersion.WithResource("customresourcedefinitions"), extClient.ApiextensionsV1().CustomResourceDefinitions()),
		apiResourceLists:            apiResourceLists,
		listers:                     util.NewConcurrentMap[schema.GroupVersionResource, cache.GenericLister](),
		informers:                   util.NewConcurrentMap[schema.GroupVersionResource, cache.SharedIndexInformer](),
		stoppers:                    util.NewConcurrentMap[schema.GroupVersionResource, chan struct{}](),
		bindingPolicyResolver:       NewBindingPolicyResolver(),
		workqueue:                   workqueue.NewRateLimitingQueueWithConfig(ratelimiter, workqueue.RateLimitingQueueConfig{Name: ControllerName + "-" + wdsName}),
		allowedGroupsSet:            allowedGroupsSet,
	}

	return controller, nil
}

// EnsureCRDs will ensure that the CRDs are installed.
// Call this before Start.
func (c *Controller) EnsureCRDs(ctx context.Context) error {
	return crd.ApplyCRDs(ctx, ControllerName, c.extClient, c.logger)
}

// AppendKSResources lets the controller know about the KS resources.
// Call this after EnsureCRDs and before Start.
func (c *Controller) AppendKSResources(ctx context.Context) error {
	gv := v1alpha1.GroupVersion.String()
	list, err := c.discoveryClient.ServerResourcesForGroupVersion(gv)
	if err != nil {
		return err
	}
	c.apiResourceLists = append(c.apiResourceLists, list)
	return nil
}

// Start the controller
func (c *Controller) Start(parentCtx context.Context, workers int, cListers chan interface{}) error {
	logger := klog.FromContext(parentCtx).WithName(ControllerName)
	ctx := klog.NewContext(parentCtx, logger)

	// Create informer on managedclusters so we can re-evaluate BindingPolicies.
	// This informer differs from the other informers in that it listens on the ocm hub.
	if err := c.setupManagedClustersInformer(ctx); err != nil {
		return err
	}

	if err := c.setupBindingPolicyInformer(ctx); err != nil {
		return err
	}
	if err := c.setupBindingInformer(ctx); err != nil {
		return err
	}
	c.ksInformerFactoryStart(ctx.Done())
	if ok := cache.WaitForCacheSync(ctx.Done(), c.bindingPolicyInformer.HasSynced, c.bindingInformer.HasSynced); !ok {
		return fmt.Errorf("failed to wait for KubeStellar informers to sync")
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- c.run(ctx, workers, cListers)
	}()

	// check for errors at startup, after all started we let it continue
	// so we can start the controller-runtime manager
	select {
	case err := <-errChan:
		return err
	case <-time.After(3 * time.Second):
		return nil
	}
}

// GetBindingPolicyResolutionBroker returns the ResolutionBroker for the
// BindingPolicyResolver managed by the controller.
func (c *Controller) GetBindingPolicyResolutionBroker() ResolutionBroker {
	return c.bindingPolicyResolver.Broker()
}

func (c *Controller) GetBindingPolicyResolver() BindingPolicyResolver {
	return c.bindingPolicyResolver
}

// Invoked by Start() to run the controller
func (c *Controller) run(ctx context.Context, workers int, cListers chan interface{}) error {
	defer c.workqueue.ShutDown()

	logger := klog.FromContext(ctx)

	// Create a dynamic shared informer factory
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(c.dynamicClient, 0*time.Minute)

	// Loop through the api resources and create informers and listers for each of them
	for _, list := range c.apiResourceLists {
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			c.logger.Error(err, "Failed to parse a GroupVersion", "groupVersion", list.GroupVersion)
			continue
		}
		if _, excluded := excludedGroups[gv.Group]; excluded {
			logger.V(1).Info("Ignoring APIResourceList", "groupVersion", list.GroupVersion)
			continue
		}
		if !util.IsAPIGroupAllowed(gv.Group, c.allowedGroupsSet) {
			logger.V(1).Info("No need to watch per user input", "groupVersion", list.GroupVersion)
			continue
		}
		logger.V(1).Info("Working on APIResourceList", "groupVersion", list.GroupVersion, "numResources", len(list.APIResources))
		for _, resource := range list.APIResources {
			if _, excluded := excludedResourceNames[resource.Name]; excluded {
				continue
			}
			informable := verbsSupportInformers(resource.Verbs)
			if informable {
				gvr := gv.WithResource(resource.Name)
				informer := informerFactory.ForResource(gvr).Informer()
				c.informers.Set(gvr, informer)

				// add the event handler functions
				informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
					AddFunc: func(obj interface{}) {
						// IMPORTANT: the resource loop variable cannot be used in these closures.
						c.handleObject(obj, gvr.Resource, "add")
					},
					UpdateFunc: func(old, new interface{}) {
						if shouldSkipUpdate(old, new) {
							return
						}
						c.handleObject(new, gvr.Resource, "update")
					},
					DeleteFunc: func(obj interface{}) {
						c.handleObject(obj, gvr.Resource, "delete")
					},
				})

				// create and index the lister
				c.listers.Set(gvr, cache.NewGenericLister(informer.GetIndexer(), gvr.GroupResource()))

				// run the informer
				// we need to be able to stop informers for APIs (CRDs) that are removed
				// after startup, therefore we use a stopper channel for each informer
				// instead than informerFactory.Start(ctx.Done())
				stopper := make(chan struct{})
				defer close(stopper)
				c.stoppers.Set(gvr, stopper)
				go informer.Run(stopper)
			}
		}
	}

	// wait for all informers caches to be synced
	// then send listers for the status controller to use
	if err := c.informers.Iterator(func(_ schema.GroupVersionResource, informer cache.SharedIndexInformer) error {
		if ok := cache.WaitForCacheSync(ctx.Done(), informer.HasSynced); !ok {
			return fmt.Errorf("failed to wait for caches to sync")
		}

		return nil // continue iterating
	}); err != nil {
		return err // no need to wrap because it is already clear
	}

	c.logger.Info("All workload caches synced")
	cListers <- c.listers
	c.logger.Info("Sent listers")

	// populate the BindingPolicyResolver with entries for existing bindingpolicies
	if err := c.populateBindingPolicyResolverWithExistingBindingPolicies(); err != nil {
		return fmt.Errorf("failed to populate the BindingPolicyResolver for the existing bindingpolicies: %w", err)
	}

	c.logger.Info("Starting workers", "count", workers)
	for i := 0; i < workers; i++ {
		logger := c.logger.WithName(fmt.Sprintf("worker-%d", i))
		workerCtx := klog.NewContext(ctx, logger)
		go wait.UntilWithContext(workerCtx, c.runWorker, time.Second)
	}

	c.logger.Info("Started workers")
	c.initializedTs = time.Now()

	<-ctx.Done()
	c.logger.Info("Shutting down workers")

	return nil
}

func (c *Controller) setupManagedClustersInformer(ctx context.Context) error {
	_, err := c.clusterInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			objM := obj.(metav1.Object)
			c.evaluateBindingPolicies(ctx, objM.GetName(), objM.GetLabels())
		},
		UpdateFunc: func(old, new interface{}) {
			oldM := old.(metav1.Object)
			newM := new.(metav1.Object)
			// Re-evaluateBindingPolicies iff labels have changed.
			oldLabels := oldM.GetLabels()
			newLabels := newM.GetLabels()
			if !reflect.DeepEqual(oldLabels, newLabels) {
				c.logger.V(5).Info("Handling labels change", "old", old, "new", new)
				c.evaluateBindingPoliciesForUpdate(ctx, newM.GetName(), oldLabels, newLabels)
			}
		},
		DeleteFunc: func(obj interface{}) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			objM := obj.(metav1.Object)
			c.evaluateBindingPolicies(ctx, objM.GetName(), objM.GetLabels())
		},
	})
	if err != nil {
		c.logger.Error(err, "failed to add managedclusters informer event handler")
		return err
	}
	c.clusterInformerFactoryStart(ctx.Done())
	if ok := cache.WaitForCacheSync(ctx.Done(), c.clusterInformer.HasSynced); !ok {
		return fmt.Errorf("failed to wait for managedclusters informer to sync")
	}
	return nil
}

func (c *Controller) setupBindingPolicyInformer(ctx context.Context) error {
	logger := klog.FromContext(ctx)
	_, err := c.bindingPolicyInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			bp := obj.(*v1alpha1.BindingPolicy)
			logger.V(5).Info("Enqueuing reference to BindingPolicy because of informer add event", "name", bp.Name, "resourceVersion", bp.ResourceVersion)
			c.workqueue.Add(bindingPolicyRef(bp.Name))
		},
		UpdateFunc: func(old, new interface{}) {
			oldBP := old.(*v1alpha1.BindingPolicy)
			newBP := new.(*v1alpha1.BindingPolicy)
			if oldBP.Generation != newBP.Generation {
				logger.V(5).Info("Enqueuing reference to BindingPolicy because of informer update event", "name", newBP.Name, "resourceVersion", newBP.ResourceVersion)
				c.workqueue.Add(bindingPolicyRef(newBP.Name))
			}
		},
		DeleteFunc: func(obj interface{}) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			bp := obj.(*v1alpha1.BindingPolicy)
			logger.V(5).Info("Enqueuing reference to BindingPolicy because of informer delete event", "name", bp.Name)
			c.workqueue.Add(bindingPolicyRef(bp.Name))
		},
	})
	if err != nil {
		c.logger.Error(err, "failed to add bindingpolicies informer event handler")
		return err
	}
	return nil
}

func (c *Controller) setupBindingInformer(ctx context.Context) error {
	logger := klog.FromContext(ctx)
	_, err := c.bindingInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			bdg := obj.(*v1alpha1.Binding)
			logger.V(5).Info("Enqueuing reference to Binding because of informer add event", "name", bdg.Name, "resourceVersion", bdg.ResourceVersion)
			c.workqueue.Add(bindingPolicyRef(bdg.Name))
		},
		UpdateFunc: func(old, new interface{}) {
			oldBdg := old.(*v1alpha1.Binding)
			newBdg := new.(*v1alpha1.Binding)
			if oldBdg.Generation != newBdg.Generation {
				logger.V(5).Info("Enqueuing reference to Binding because of informer update event", "name", newBdg.Name, "resourceVersion", newBdg.ResourceVersion)
				c.workqueue.Add(bindingPolicyRef(newBdg.Name))
			}
		},
		DeleteFunc: func(obj interface{}) {
			if typed, is := obj.(cache.DeletedFinalStateUnknown); is {
				obj = typed.Obj
			}
			bdg := obj.(*v1alpha1.Binding)
			logger.V(5).Info("Enqueuing reference to Binding because of informer delete event", "name", bdg.Name)
			c.workqueue.Add(bindingPolicyRef(bdg.Name))
		},
	})
	if err != nil {
		c.logger.Error(err, "failed to add bindingpolicies informer event handler")
		return err
	}
	return nil
}

func shouldSkipUpdate(old, new interface{}) bool {
	oldMObj := old.(metav1.Object)
	newMObj := new.(metav1.Object)
	// do not enqueue update events for objects that have not changed
	if newMObj.GetResourceVersion() == oldMObj.GetResourceVersion() {
		return true
	}

	return false
}

// We only start informers on resources that support both watch and list
func verbsSupportInformers(verbs []string) bool {
	var hasList, hasWatch bool
	for _, verb := range verbs {
		switch verb {
		case "list":
			hasList = true
		case "watch":
			hasWatch = true
		}
	}
	return hasList && hasWatch
}

// Event handler: enqueues the objects to be processed
// At this time it is very simple, more complex processing might be required
// here.
func (c *Controller) handleObject(obj any, resource string, eventType string) {
	wasDeletedFinalStateUnknown := false
	switch typed := obj.(type) {
	case cache.DeletedFinalStateUnknown:
		obj = typed.Obj
		wasDeletedFinalStateUnknown = true
	}
	c.logger.V(5).Info("Got object event", "eventType", eventType,
		"wasDeletedFinalStateUnknown", wasDeletedFinalStateUnknown, "obj", util.RefToRuntimeObj(obj.(runtime.Object)),
		"resource", resource)

	c.enqueueObject(obj, resource)
}

// enqueueObject converts an object into an ObjectIdentifier struct which is
// then put onto the work queue.
func (c *Controller) enqueueObject(obj interface{}, resource string) {
	objIdentifier := util.IdentifierForObject(obj.(util.MRObject), resource)
	c.enqueueObjectIdentifier(objIdentifier)
}

func (c *Controller) enqueueObjectIdentifier(objIdentifier util.ObjectIdentifier) {
	c.workqueue.Add(objIdentifier)
}

func (c *Controller) enqueueBinding(name string) {
	// this resource can have bursts of
	// updates due to being updated by multiple workload-objects getting
	// processed concurrently at a high rate.
	c.workqueue.AddAfter(bindingRef(name), bindingQueueingDelay)
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem reads a single work item off the workqueue and
// attempt to process it by calling the reconcile.
func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	logger := klog.FromContext(ctx)
	item, shutdown := c.workqueue.Get()
	if shutdown {
		logger.V(1).Info("Worker is done")
		return false
	}
	logger.V(4).Info("Dequeued", "item", item, "type", fmt.Sprintf("%T", item))

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func() error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(item)
		// Run the reconciler, passing it the full object identifier
		if err := c.reconcile(ctx, item); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(item)
			return fmt.Errorf("error reconciling object (identifier: %#v, type: %T): %s, requeuing", item, item, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(item)
		logger.V(4).Info("Successfully reconciled", "objectIdentifier", item, "type", fmt.Sprintf("%T", item))
		return nil
	}()

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) reconcile(ctx context.Context, item any) error {
	logger := klog.FromContext(ctx)

	switch objIdentifier := item.(type) {
	case bindingRef:
		return c.syncBinding(ctx, string(objIdentifier)) // this function logs through all its exits
	case bindingPolicyRef:
		if err := c.syncBindingPolicy(ctx, string(objIdentifier)); err != nil {
			return fmt.Errorf("failed to handle bindingpolicy: %w", err) // error logging after this call
			// will add name.
		}

		logger.V(5).Info("Handled bindingpolicy", "objectIdentifier", objIdentifier)
		return nil
	case util.ObjectIdentifier:
		if util.ObjIdentifierIsForCRD(objIdentifier) {
			if err := c.syncCRD(ctx, objIdentifier); err != nil {
				return fmt.Errorf("failed to handle CRD: %w", err) // error logging after this call
				// will add name.
			}
			logger.V(5).Info("Handled CRD", "objectIdentifier", objIdentifier)
		}

		return c.updateResolutions(ctx, objIdentifier)
	}
	logger.Error(nil, "Impossible workqueue entry", "type", fmt.Sprintf("%T", item), "value", item)
	return nil
}

func (c *Controller) getObjectFromIdentifier(objIdentifier util.ObjectIdentifier) (runtime.Object, error) {
	lister, found := c.listers.Get(objIdentifier.GVR())
	if !found {
		return nil, fmt.Errorf("could not get lister for gvr: %s", objIdentifier.GVR())
	}

	return getObject(lister, objIdentifier.ObjectName.Namespace, objIdentifier.ObjectName.Name)
}

func getObject(lister cache.GenericLister, namespace, name string) (runtime.Object, error) {
	if namespace != "" {
		return lister.ByNamespace(namespace).Get(name)
	}
	return lister.Get(name)
}

func isBeingDeleted(obj runtime.Object) bool {
	mObj := obj.(metav1.Object)
	return mObj.GetDeletionTimestamp() != nil
}

func (c *Controller) GetListers() util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister] {
	return c.listers
}

func (c *Controller) GetInformers() util.ConcurrentMap[schema.GroupVersionResource, cache.SharedIndexInformer] {
	return c.informers
}

// populateBindingPolicyResolverWithExistingBindingPolicies fills the BindingPolicyResolver
// with entries for existing BindingPolicy objects. Any bindingpolicy name that is not
// associated with a resolution gets associated to an empty resolution.
// No concurrent calls allowed.
// May not be called concurrently with Controller::reconcile.
func (c *Controller) populateBindingPolicyResolverWithExistingBindingPolicies() error {
	bindingpolicies, err := c.listBindingPolicies()
	if err != nil {
		return fmt.Errorf("failed to list BindingPolicies: %w", err)
	}

	for _, bindingpolicy := range bindingpolicies {
		c.bindingPolicyResolver.NoteBindingPolicy(bindingpolicy)
	}

	return nil
}
