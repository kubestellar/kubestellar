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
	ocmclientset "open-cluster-management.io/api/client/cluster/clientset/versioned"
	ocminformers "open-cluster-management.io/api/client/cluster/informers/externalversions"

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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/crd"
	"github.com/kubestellar/kubestellar/pkg/ocm"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const (
	controllerName      = "Binding"
	defaultResyncPeriod = time.Duration(0)
)

// Resource groups to exclude for watchers as they should not be delivered to other clusters
var excludedGroups = map[string]bool{
	"flowcontrol.apiserver.k8s.io": true,
	"scheduling.k8s.io":            true,
	"discovery.k8s.io":             true,
	"apiregistration.k8s.io":       true,
	"coordination.k8s.io":          true,
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
	logger           logr.Logger
	ocmClientset     ocmclientset.Interface // used for ManagedCluster in ITS
	ocmClient        client.Client          // used for ManagedCluster, ManifestWork in ITS
	dynamicClient    dynamic.Interface      // used for CRD, Binding[Policy], workload
	kubernetesClient kubernetes.Interface   // used for Namespaces, and Discovery

	extClient apiextensionsclientset.Interface // used for CRD

	apiResourceLists []*metav1.APIResourceList
	listers          util.ConcurrentMap[schema.GroupVersionResource, cache.GenericLister]
	informers        util.ConcurrentMap[schema.GroupVersionResource, cache.SharedIndexInformer]
	stoppers         util.ConcurrentMap[schema.GroupVersionResource, chan struct{}]

	bindingPolicyResolver BindingPolicyResolver
	workqueue             workqueue.RateLimitingInterface
	initializedTs         time.Time
	wdsName               string
	allowedGroupsSet      sets.Set[string]
}

// Create a new binding controller
func NewController(parentLogger logr.Logger, wdsRestConfig *rest.Config, imbsRestConfig *rest.Config,
	wdsName string, allowedGroupsSet sets.Set[string]) (*Controller, error) {
	logger := parentLogger.WithName(controllerName)

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
	logger.V(1).Info("Parameters of the tuned client's token bucket rate limiter", "burst", wdsRestConfigTuned.Burst, "qps", wdsRestConfigTuned.QPS)

	// dynamicClient needs higher rate than its default because dynamicClient is repeatedly used by the
	// reflectors for each of the GVRs, all at the beginning of the controller run
	dynamicClient, err := dynamic.NewForConfig(wdsRestConfigTuned)
	if err != nil {
		return nil, err
	}

	ocmClientset, err := ocmclientset.NewForConfig(imbsRestConfig)
	if err != nil {
		return nil, err
	}

	ocmClient := ocm.GetOCMClient(imbsRestConfig)

	return makeController(logger, dynamicClient, kubernetesClient, extClient, ocmClientset, ocmClient, apiResourceLists, wdsName, allowedGroupsSet)
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
	dynamicClient dynamic.Interface, // used for CRD, Binding[Policy], workload
	kubernetesClient kubernetes.Interface, // used for Namespaces, and Discovery
	extClient apiextensionsclientset.Interface, // used for CRD
	ocmClientset ocmclientset.Interface, // used for ManagedCluster in ITS
	ocmClient client.Client, // used for ManagedCluster, ManifestWork in ITS
	apiResourceLists []*metav1.APIResourceList,
	wdsName string, allowedGroupsSet sets.Set[string]) (*Controller, error) {

	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	controller := &Controller{
		wdsName:               wdsName,
		logger:                logger,
		ocmClientset:          ocmClientset,
		ocmClient:             ocmClient,
		dynamicClient:         dynamicClient,
		kubernetesClient:      kubernetesClient,
		extClient:             extClient,
		apiResourceLists:      apiResourceLists,
		listers:               util.NewConcurrentMap[schema.GroupVersionResource, cache.GenericLister](),
		informers:             util.NewConcurrentMap[schema.GroupVersionResource, cache.SharedIndexInformer](),
		stoppers:              util.NewConcurrentMap[schema.GroupVersionResource, chan struct{}](),
		bindingPolicyResolver: NewBindingPolicyResolver(),
		workqueue:             workqueue.NewRateLimitingQueue(ratelimiter),
		allowedGroupsSet:      allowedGroupsSet,
	}

	return controller, nil
}

// EnsureCRDs will ensure that the CRDs are installed.
// Call this before Start.
func (c *Controller) EnsureCRDs(ctx context.Context) error {
	return crd.ApplyCRDs(ctx, c.dynamicClient, c.kubernetesClient, c.extClient, c.logger)
}

// AppendKSResources lets the controller know about the KS resources.
// Call this after EnsureCRDs and before Start.
func (c *Controller) AppendKSResources(ctx context.Context) error {
	gv := v1alpha1.GroupVersion.String()
	list, err := c.kubernetesClient.Discovery().ServerResourcesForGroupVersion(gv)
	if err != nil {
		return err
	}
	c.apiResourceLists = append(c.apiResourceLists, list)
	return nil
}

// Start the controller
func (c *Controller) Start(parentCtx context.Context, workers int, cListers chan interface{}) error {
	logger := klog.FromContext(parentCtx).WithName(controllerName)
	ctx := klog.NewContext(parentCtx, logger)
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

	c.logger.Info("All caches synced")
	cListers <- c.listers
	c.logger.Info("Sent listers")

	// Create informer on managedclusters so we can re-evaluate BindingPolicies.
	// This informer differ from the previous informers in that it listens on the ocm hub.
	err := c.createManagedClustersInformer(ctx)
	if err != nil {
		return err
	}

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

func (c *Controller) createManagedClustersInformer(ctx context.Context) error {
	informerFactory := ocminformers.NewSharedInformerFactory(c.ocmClientset, defaultResyncPeriod)
	informer := informerFactory.Cluster().V1().ManagedClusters().Informer()
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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
				c.evaluateBindingPoliciesForUpdate(ctx, newM.GetName(), oldLabels, newLabels)
			}
		},
		DeleteFunc: func(obj interface{}) {
			objM := obj.(metav1.Object)
			c.evaluateBindingPolicies(ctx, objM.GetName(), objM.GetLabels())
		},
	})
	if err != nil {
		c.logger.Error(err, "failed to add managedclusters informer event handler")
		return err
	}
	informerFactory.Start(ctx.Done())
	if ok := cache.WaitForCacheSync(ctx.Done(), informer.HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
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
	c.logger.Info("Got object event", "eventType", eventType,
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
	c.workqueue.AddAfter(util.ObjectIdentifier{
		GVK: schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    util.BindingKind},
		Resource: util.BindingResource,
		ObjectName: cache.ObjectName{
			Namespace: metav1.NamespaceNone,
			Name:      name,
		},
	}, bindingQueueingDelay) // this resource can have bursts of
	// updates due to being updated by multiple workload-objects getting
	// processed concurrently at a high rate.
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
	logger.V(4).Info("Dequeued", "item", item)

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(objIdentifierA any) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(objIdentifierA)
		var objIdentifier util.ObjectIdentifier
		var ok bool
		// We expect util.ObjectIdentifier to come off the workqueue. We do this as the
		// delayed nature of the workqueue means the items in the informer
		// cache may actually be more up to date that when the item was
		// initially put onto the workqueue.
		if objIdentifier, ok = objIdentifierA.(util.ObjectIdentifier); !ok {
			// if the item in the workqueue is invalid, we call
			// Forget here to avoid process a work item that is invalid.
			c.workqueue.Forget(objIdentifierA)
			utilruntime.HandleError(fmt.Errorf("expected util.ObjectIdentifier in workqueue but got %#v",
				objIdentifierA))
			return nil
		}
		// Run the reconciler, passing it the full object identifier
		if err := c.reconcile(ctx, objIdentifier); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(objIdentifier)
			return fmt.Errorf("error reconciling object (identifier: %#v): %s, requeuing", objIdentifier, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(objIdentifier)
		logger.V(4).Info("Successfully reconciled", "objectIdentifier", objIdentifier)
		return nil
	}(item)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) reconcile(ctx context.Context, objIdentifier util.ObjectIdentifier) error {
	logger := klog.FromContext(ctx)

	// special handling for selected API resources
	// note that object is *unstructured.Unstructured so we cannot
	// just use "switch obj.(type)"
	if util.ObjIdentifierIsForBinding(objIdentifier) {
		return c.syncBinding(ctx, objIdentifier) // this function logs through all its exits
	} else if util.ObjIdentifierIsForBindingPolicy(objIdentifier) {
		if err := c.handleBindingPolicy(ctx, objIdentifier); err != nil {
			return fmt.Errorf("failed to handle bindingpolicy: %w", err) // error logging after this call
			// will add name.
		}

		logger.Info("Handled bindingpolicy", "objectIdentifier", objIdentifier)
		return nil
	} else if util.ObjIdentifierIsForCRD(objIdentifier) {
		if err := c.handleCRD(ctx, objIdentifier); err != nil {
			return fmt.Errorf("failed to handle CRD: %w", err) // error logging after this call
			// will add name.
		}
		logger.Info("Handled CRD", "objectIdentifier", objIdentifier)
	}

	return c.updateResolutions(ctx, objIdentifier)
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
func (c *Controller) populateBindingPolicyResolverWithExistingBindingPolicies() error {
	bindingpolicies, err := c.listBindingPolicies()
	if err != nil {
		return fmt.Errorf("failed to list BindingPolicies: %w", err)
	}

	for _, bindingPolicyObject := range bindingpolicies {
		bindingpolicy, err := runtimeObjectToBindingPolicy(bindingPolicyObject)
		if err != nil {
			return fmt.Errorf("failed to convert runtime.Object to BindingPolicy: %w", err)
		}

		c.bindingPolicyResolver.NoteBindingPolicy(bindingpolicy)
	}

	return nil
}
