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
	"strings"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	ocmclientset "open-cluster-management.io/api/client/cluster/clientset/versioned"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
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

const controllerName = "Binding"

// Resources group versions to exclude for watchers as they should not delivered to other clusters
var excludedGroupVersions = map[string]bool{
	"flowcontrol.apiserver.k8s.io/v1beta3": true,
	"flowcontrol.apiserver.k8s.io/v1beta2": true,
	"scheduling.k8s.io/v1":                 true,
	"discovery.k8s.io/v1":                  true,
	"apiregistration.k8s.io/v1":            true,
	"coordination.k8s.io/v1":               true,
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

const bindingQueueingDelay = 2 * time.Second

// Controller watches all objects, finds associated bindingpolicies, when matched a bindingpolicy wraps and
// places objects into mailboxes
type Controller struct {
	logger           logr.Logger
	ocmClientset     ocmclientset.Interface // used for ManagedCluster in ITS
	ocmClient        client.Client          // used for ManagedCluster, ManifestWork in ITS
	dynamicClient    dynamic.Interface      // used for CRD, Binding[Policy], workload
	kubernetesClient kubernetes.Interface   // used for Namespaces, and Discovery

	extClient             apiextensionsclientset.Interface // used for CRD
	listers               map[string]cache.GenericLister
	gvkGvrMapper          util.GvkGvrMapper
	informers             map[string]cache.SharedIndexInformer
	stoppers              map[string]chan struct{}
	bindingPolicyResolver BindingPolicyResolver
	workqueue             workqueue.RateLimitingInterface
	initializedTs         time.Time
	wdsName               string
	allowedGroupsSet      sets.Set[string]
}

// Create a new binding controller
func NewController(parentLogger logr.Logger, wdsRestConfig *rest.Config, imbsRestConfig *rest.Config,
	wdsName string, allowedGroupsSet sets.Set[string]) (*Controller, error) {
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(50), 300)},
	)

	dynamicClient, err := dynamic.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	kubernetesClient, err := kubernetes.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	extClient, err := apiextensionsclientset.NewForConfig(wdsRestConfig)
	if err != nil {
		return nil, err
	}

	ocmClientset, err := ocmclientset.NewForConfig(imbsRestConfig)
	if err != nil {
		return nil, err
	}

	ocmClient := *ocm.GetOCMClient(imbsRestConfig)

	gvkGvrMapper := util.NewGvkGvrMapper()

	controller := &Controller{
		wdsName:               wdsName,
		logger:                parentLogger.WithName(controllerName),
		ocmClientset:          ocmClientset,
		ocmClient:             ocmClient,
		dynamicClient:         dynamicClient,
		kubernetesClient:      kubernetesClient,
		extClient:             extClient,
		listers:               make(map[string]cache.GenericLister),
		informers:             make(map[string]cache.SharedIndexInformer),
		stoppers:              make(map[string]chan struct{}),
		bindingPolicyResolver: NewBindingPolicyResolver(gvkGvrMapper),
		gvkGvrMapper:          gvkGvrMapper,
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

// Start the controller
func (c *Controller) Start(parentCtx context.Context, workers int) error {
	logger := klog.FromContext(parentCtx).WithName(controllerName)
	ctx := klog.NewContext(parentCtx, logger)
	errChan := make(chan error, 1)
	go func() {
		errChan <- c.run(ctx, workers)
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
func (c *Controller) run(ctx context.Context, workers int) error {
	defer c.workqueue.ShutDown()

	// Get all the api resources in the cluster
	apiResources, err := c.kubernetesClient.Discovery().ServerPreferredResources()
	if err != nil {
		// ignore the error caused by a stale API service
		if !strings.Contains(err.Error(), util.UnableToRetrieveCompleteAPIListError) {
			return err
		}
	}

	// Create a dynamic shared informer factory
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(c.dynamicClient, 0*time.Minute)

	// Loop through the api resources and create informers and listers for each of them
	for _, group := range apiResources {
		if _, excluded := excludedGroupVersions[group.GroupVersion]; excluded {
			continue
		}
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			c.logger.Error(err, "Failed to parse a GroupVersion", "groupVersion", group.GroupVersion)
			continue
		}
		for _, resource := range group.APIResources {
			if _, excluded := excludedResourceNames[resource.Name]; excluded {
				continue
			}
			if !util.IsAPIGroupAllowed(gv.Group, c.allowedGroupsSet) {
				continue
			}
			informable := verbsSupportInformers(resource.Verbs)
			if informable {
				key := util.KeyForGroupVersionKind(gv.Group, gv.Version, resource.Kind)
				informer := informerFactory.ForResource(gv.WithResource(resource.Name)).Informer()
				c.informers[key] = informer

				// add the event handler functions
				informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
					AddFunc: c.handleObject,
					UpdateFunc: func(old, new interface{}) {
						if shouldSkipUpdate(old, new) {
							return
						}
						c.handleObject(new)
					},
					DeleteFunc: func(obj interface{}) {
						if shouldSkipDelete(obj) {
							return
						}
						c.handleObject(obj)
					},
				})

				// create and index the lister
				lister := cache.NewGenericLister(informer.GetIndexer(), schema.GroupResource{Group: resource.Group, Resource: resource.Name})
				c.listers[key] = lister

				c.gvkGvrMapper.Add(schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: resource.Kind},
					schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource.Name})

				// run the informer
				// we need to be able to stop informers for APIs (CRDs) that are removed
				// after startup, therefore we use a stopper channel for each informer
				// instead than informerFactory.Start(ctx.Done())
				stopper := make(chan struct{})
				defer close(stopper)
				c.stoppers[key] = stopper
				go informer.Run(stopper)
			}
		}
	}

	// wait for all informers caches to be synced
	for _, informer := range c.informers {
		if ok := cache.WaitForCacheSync(ctx.Done(), informer.HasSynced); !ok {
			return fmt.Errorf("failed to wait for caches to sync")
		}
	}
	c.logger.Info("All caches synced")

	// Create informer on managedclusters so we can re-evaluate BindingPolicies.
	// This informer differ from the previous informers in that it listens on the ocm hub.
	err = c.createManagedClustersInformer(ctx)
	if err != nil {
		return err
	}

	// populate the BindingPolicyResolver with entries for existing bindingpolicies
	if err := c.populateBindingPolicyResolverWithExistingBindingPolicies(); err != nil {
		return fmt.Errorf("failed to populate the BindingPolicyResolver for the existing bindingpolicies: %w", err)
	}

	c.logger.Info("Starting workers", "count", workers)
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	c.logger.Info("Started workers")
	c.initializedTs = time.Now()

	<-ctx.Done()
	c.logger.Info("Shutting down workers")

	return nil
}

func (c *Controller) createManagedClustersInformer(ctx context.Context) error {
	informerFactory := ocm.GetOCMInformerFactory(c.ocmClientset)
	informer := informerFactory.Cluster().V1().ManagedClusters().Informer()
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			labels := obj.(metav1.Object).GetLabels()
			c.evaluateBindingPolicies(ctx, labels)
		},
		UpdateFunc: func(old, new interface{}) {
			// Re-evaluateBindingPolicies iff labels have changed.
			oldLabels := old.(metav1.Object).GetLabels()
			newLabels := new.(metav1.Object).GetLabels()
			if !reflect.DeepEqual(oldLabels, newLabels) {
				c.evaluateBindingPoliciesForUpdate(ctx, oldLabels, newLabels)
			}
		},
		DeleteFunc: func(obj interface{}) {
			labels := obj.(metav1.Object).GetLabels()
			c.evaluateBindingPolicies(ctx, labels)
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
	// avoid enqueing events when adding finalizers to bindingpolicy
	if util.IsBindingPolicy(new) && (len(newMObj.GetFinalizers()) > len(oldMObj.GetFinalizers())) {
		return true
	}
	return false
}

func shouldSkipDelete(obj interface{}) bool {
	// since delete for bindingpolicy is handled by the finalizer logic
	// no need to handle delete for bindingpolicy to minimize events in the
	// delete manifests
	return util.IsBindingPolicy(obj)
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
// At this time it is very simple, more complex processing might be required here
func (c *Controller) handleObject(obj any) {
	rObj := obj.(runtime.Object)
	c.logger.V(2).Info("Got object event", "obj", util.RefToRuntimeObj(rObj))
	c.enqueueObject(obj, false)
}

// enqueueObject converts an object into a key struct which is then put onto the work queue.
func (c *Controller) enqueueObject(obj interface{}, skipCheckIsDeleted bool) {
	var key util.Key
	var err error
	if key, err = util.KeyForGroupVersionKindNamespaceName(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	if !skipCheckIsDeleted {
		// we need to check if object was deleted. If deleted we need to enqueue the runtime.Object
		// so we can still evaluate the label selectors and delete it from the clusters where it
		// was deployed. This does not break the best practice of only storing the keys so that
		// the latest version of the object is retrieved, as we know the object was deleted when
		// we get a copy of the runtime.Object and there is no longer a copy on the API server.
		_, err = c.getObjectFromKey(key)
		if err != nil {
			// The resource no longer exist, which means it has been deleted.
			if errors.IsNotFound(err) {
				deletedObj := copyObjectMetaAndType(obj.(runtime.Object))
				key.DeletedObject = &deletedObj
				c.workqueue.Add(key)
			}
			return
		}
	}
	c.workqueue.Add(key)
}

func (c *Controller) enqueueBinding(name string) {
	c.workqueue.AddAfter(util.Key{
		GVK: schema.GroupVersionKind{
			Group:   v1alpha1.GroupVersion.Group,
			Version: v1alpha1.GroupVersion.Version,
			Kind:    util.BindingKind},
		NamespacedName: cache.ObjectName{
			Namespace: metav1.NamespaceNone,
			Name:      name,
		},
		DeletedObject: nil,
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
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key util.Key
		var ok bool
		// We expect util.Key to come off the workqueue. We do this as the delayed
		// nature of the workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(util.Key); !ok {
			// if the item in the workqueue is invalid, we call
			// Forget here to avoid process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected util.Key in workqueue but got %#v", obj))
			return nil
		}
		// Run the reconciler, passing it the full key or the metav1 Object
		if err := c.reconcile(ctx, key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(obj)
			return fmt.Errorf("error syncing key '%#v': %s, requeuing", obj, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		c.logger.V(2).Info("Successfully synced", "object", obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) reconcile(ctx context.Context, key util.Key) error {
	var obj runtime.Object
	var err error
	// special handling for binding resource as it is the only
	// resource that is queued directly as key, without necessarily first
	// existing as an object.
	if util.KeyIsForBinding(key) {
		return c.syncBinding(ctx, key) // this function logs through all its exits
	}

	if key.DeletedObject == nil {
		obj, err = c.getObjectFromKey(key)
		if err != nil {
			// The resource no longer exist, which means it has been deleted.
			if errors.IsNotFound(err) {
				c.logger.Info("object referenced from work queue no longer exists",
					"object-name", key.NamespacedName, "object-gvk", key.GvkKey())
				return nil
			}
			return err
		}
	} else {
		obj = *key.DeletedObject
	}

	// special handling for selected API resources
	// note that object is *unstructured.Unstructured so we cannot
	// just use "switch obj.(type)"
	if util.IsBindingPolicy(obj) {
		if err := c.handleBindingPolicy(ctx, obj); err != nil {
			return fmt.Errorf("failed to handle bindingpolicy: %w", err) // error logging after this call
			// will add name.
		}

		c.logger.Info("handled bindingpolicy", "object", util.RefToRuntimeObj(obj))
		return nil
	} else if util.IsCRD(obj) {
		if err := c.handleCRD(obj); err != nil {
			return fmt.Errorf("failed to handle CRD: %w", err) // error logging after this call
			// will add name.
		}
		c.logger.Info("handled CRD", "object", util.RefToRuntimeObj(obj))
	}

	// avoid further processing for keys of objects being deleted that do not have a deleted object
	if isBeingDeleted(obj) && key.DeletedObject == nil {
		return nil
	}

	if err := c.updateDecisions(obj); err != nil {
		c.logger.Error(err, "failed to update bindingpolicy resolutions for object",
			"object", util.RefToRuntimeObj(obj))
		// return nil // not changing existing flow before transport is ready
	}

	//TODO (maroon): everything below this line should be deleted when transport is ready
	clusterSet, managedByBindingPolicies, singletonStatus, err := c.matchSelectors(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error matching selectors: %s", err))
		return nil
	}

	// if no clusters
	if len(clusterSet) == 0 {
		return nil
	}

	if singletonStatus {
		clusterSet = pickSingleCluster(clusterSet)
	}

	if key.DeletedObject != nil {
		c.logger.Info("Deleting", "object", util.RefToRuntimeObj(obj),
			"from clusters", clusterSet)
		deleteObjectOnManagedClusters(c.logger, c.ocmClient, *key.DeletedObject, clusterSet)
		return nil
	}

	c.logger.Info("Delivering", "object", util.RefToRuntimeObj(obj),
		"to clusters", clusterSet)
	return c.deliverObjectToManagedClusters(obj, clusterSet, managedByBindingPolicies, singletonStatus)
}

func (c *Controller) getObjectFromKey(key util.Key) (runtime.Object, error) {
	lister := c.listers[key.GvkKey()]
	if lister == nil {
		utilruntime.HandleError(fmt.Errorf("could not get lister for key: %s", key.GvkKey()))
		return nil, nil
	}

	return getObject(lister, key.NamespacedName.Namespace, key.NamespacedName.Name)
}

func getObject(lister cache.GenericLister, namespace, name string) (runtime.Object, error) {
	if namespace != "" {
		return lister.ByNamespace(namespace).Get(name)
	}
	return lister.Get(name)
}

// create a minimal runtime.Object copy with no spec and status for use
// with the delete
func copyObjectMetaAndType(obj runtime.Object) runtime.Object {
	dest := obj.DeepCopyObject()
	dest = ocm.ZeroFields(dest)
	val := reflect.ValueOf(dest).Elem()

	spec := val.FieldByName("Spec")
	if spec.IsValid() {
		spec.Set(reflect.Zero(spec.Type()))
	}

	status := val.FieldByName("Status")
	if status.IsValid() {
		status.Set(reflect.Zero(status.Type()))
	}

	return dest
}

func isBeingDeleted(obj runtime.Object) bool {
	mObj := obj.(metav1.Object)
	return mObj.GetDeletionTimestamp() != nil
}

func (c *Controller) GetListers() map[string]cache.GenericLister {
	return c.listers
}

func (c *Controller) GetInformers() map[string]cache.SharedIndexInformer {
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

// sort by name and pick first cluster so that the choice is deterministic based on names
func pickSingleCluster(clusterSet sets.Set[string]) sets.Set[string] {
	return sets.New(sets.List(clusterSet)[0])
}

func (c *Controller) deliverObjectToManagedClusters(
	obj runtime.Object,
	managedClusters sets.Set[string],
	managedByBindingPolices []string,
	singletonStatus bool) error {
	for clName := range managedClusters {
		// find which bindingpolicies select this managedCluster
		bindingPolicyNames := []string{}
		for _, plName := range managedByBindingPolices {
			plObj, err := c.getBindingPolicyByName(plName)
			if err != nil {
				return err
			}
			pl, err := runtimeObjectToBindingPolicy(plObj)
			if err != nil {
				return err
			}
			cl, err := ocm.GetClusterByName(c.ocmClient, clName)
			if err != nil {
				return err
			}
			// a bindingpolicy selects a managedCluster iff the managedCluster is selected by every single selector
			// i.e. selectors are ANDed
			matches, err := util.SelectorsMatchLabels(pl.Spec.ClusterSelectors, labels.Set(cl.Labels))
			if err != nil {
				return err
			}
			if matches {
				bindingPolicyNames = append(bindingPolicyNames, plName)
			}
		}
		if len(bindingPolicyNames) == 0 {
			return nil
		}
		manifest := ocm.WrapObject(obj)
		util.SetManagedByBindingPolicyLabels(manifest, c.wdsName, bindingPolicyNames, singletonStatus)
		err := reconcileManifest(c.ocmClient, manifest, clName)
		if err != nil {
			c.logger.Error(err, "Error delivering object to mailbox")
		}
	}
	return nil
}
