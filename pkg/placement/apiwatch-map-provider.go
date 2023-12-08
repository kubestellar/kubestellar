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
	"fmt"
	"sync"

	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextinfactory "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/discovery"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	ksmetav1a1 "github.com/kubestellar/kubestellar/pkg/apis/meta/v1alpha1"
	apiwatch "github.com/kubestellar/kubestellar/pkg/apiwatch"
	msclient "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
)

// NewAPIWatchMapProvider constructs an APIMapProvider that gets its information
// from apiwatch.
func NewAPIWatchMapProvider(ctx context.Context,
	numThreads int,
	spaceclient msclient.KubestellarSpaceInterface,
	spaceProviderNs string,
) APIWatchMapProvider {
	awp := &apiWatchProvider{
		context:         ctx,
		numThreads:      numThreads,
		queue:           workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		perCluster:      NewMapMap[string, *apiWatchProviderPerCluster](nil),
		spaceclient:     spaceclient,
		spaceProviderNs: spaceProviderNs,
	}
	return awp
}

type APIWatchMapProvider interface {
	APIMapProvider
	Runnable
}

type apiWatchProvider struct {
	context    context.Context
	numThreads int
	queue      workqueue.RateLimitingInterface

	sync.Mutex

	perCluster      MutableMap[string, *apiWatchProviderPerCluster]
	spaceclient     msclient.KubestellarSpaceInterface
	spaceProviderNs string
}

func (awp *apiWatchProvider) AddReceivers(clusterName string,
	groupReceiver *MappingReceiverHolder[string /*group name*/, APIGroupInfo],
	resourceReceiver *MappingReceiverHolder[metav1.GroupResource, ResourceDetails],
) {
	awp.Lock()
	defer awp.Unlock()
	wpc := MapGetAdd(awp.perCluster, clusterName, true, func(clusterName string) *apiWatchProviderPerCluster {
		ctx := context.Background()
		logger := klog.FromContext(ctx).WithValues("part", "apiwatch-map-provider")
		ctx = klog.NewContext(ctx, logger)
		wpc := &apiWatchProviderPerCluster{
			awp:     awp,
			cluster: clusterName,
			// TODO: feed the groupReceivers
			groupReceivers:    MappingReceiverHolderFork[string /*group name*/, APIGroupInfo]{},
			resourceReceivers: MappingReceiverHolderFork[metav1.GroupResource, ResourceDetails]{},
		}

		config, err := awp.spaceclient.ConfigForSpace(clusterName, awp.spaceProviderNs)
		if err != nil {
			logger.Error(err, "Failed to get space config", "space", clusterName)
		}
		discoveryScopedClient, err := discovery.NewDiscoveryClientForConfig(config)
		if err != nil {
			logger.Error(err, "Failed to create discovery client", "space", clusterName)
		}

		apiextClient, err := apiextclient.NewForConfig(config)
		if err != nil {
			logger.Error(err, "Failed to create clientset for CustomResourceDefinitions")
		}
		apiextFactory := apiextinfactory.NewSharedInformerFactory(apiextClient, 0)
		crdInformer := apiextFactory.Apiextensions().V1().CustomResourceDefinitions().Informer()
		apiextFactory.Start(ctx.Done())

		wpc.informer, wpc.lister, _ = apiwatch.NewAPIResourceInformer(ctx, clusterName, discoveryScopedClient, false, crdInformer)
		wpc.informer.AddEventHandler(wpc)
		go wpc.informer.Run(ctx.Done())
		return wpc
	})
	wpc.groupReceivers = append(MappingReceiverHolderFork[string /*group name*/, APIGroupInfo]{groupReceiver}, wpc.groupReceivers...)
	wpc.resourceReceivers = append(MappingReceiverHolderFork[metav1.GroupResource, ResourceDetails]{resourceReceiver}, wpc.resourceReceivers...)
	// The following make sure that the new receiver is notified about already-known resources
	awp.queue.Add(receiverForCluster[string /*group name*/, APIGroupInfo]{groupReceiver, clusterName})
	awp.queue.Add(receiverForCluster[metav1.GroupResource, ResourceDetails]{resourceReceiver, clusterName})
}

func (awp *apiWatchProvider) RemoveReceivers(clusterName string,
	groupReceiver *MappingReceiverHolder[string /*group name*/, APIGroupInfo],
	resourceReceiver *MappingReceiverHolder[metav1.GroupResource, ResourceDetails]) {
	awp.Lock()
	defer awp.Unlock()
	wpc, have := awp.perCluster.Get(clusterName)
	if !have {
		return
	}
	wpc.groupReceivers = SliceRemoveFunctional(wpc.groupReceivers, groupReceiver)
	wpc.resourceReceivers = SliceRemoveFunctional(wpc.resourceReceivers, resourceReceiver)
	// TODO: shut it down if there are no remaining receivers
}

type apiWatchProviderPerCluster struct {
	awp      *apiWatchProvider
	cluster  string
	informer upstreamcache.SharedInformer
	lister   apiwatch.APIResourceLister

	// The following fields may be read or written only with the provider locked,
	// but every value ever held in these fields is immutable.

	groupReceivers    MappingReceiverHolderFork[string /*group name*/, APIGroupInfo]
	resourceReceivers MappingReceiverHolderFork[metav1.GroupResource, ResourceDetails]
}

func (wpc *apiWatchProviderPerCluster) OnAdd(obj any) {
	wpc.enqueueResourceRef(obj, "add")
}

func (wpc *apiWatchProviderPerCluster) OnUpdate(oldObj, newObj any) {
	wpc.enqueueResourceRef(newObj, "update")
}

func (wpc *apiWatchProviderPerCluster) OnDelete(obj any) {
	underObj := obj
	switch typed := obj.(type) {
	case upstreamcache.DeletedFinalStateUnknown:
		underObj = typed.Obj
	default:
	}
	wpc.enqueueResourceRef(underObj, "delete")
}

func (wpc *apiWatchProviderPerCluster) enqueueResourceRef(obj any, action string) {
	rsc := obj.(*ksmetav1a1.APIResource)
	rr := resourceRef{cluster: wpc.cluster, metaname: rsc.Name}
	logger := klog.FromContext(wpc.awp.context)
	logger.V(4).Info("Enqueuing", "ref", rr)
	wpc.awp.queue.Add(rr)
}

type resourceRef struct {
	cluster  string
	metaname string
}

type receiverForCluster[Key comparable, Val any] struct {
	receiver MappingReceiver[Key, Val]
	cluster  string
}

// Run animates the controller, finishing and returning when the context of
// the controller is done.
// Call this after the informers have been started.
func (ctl *apiWatchProvider) Run(ctx context.Context) {
	doneCh := ctx.Done()
	for worker := 0; worker < ctl.numThreads; worker++ {
		go ctl.syncLoop(ctx, worker)
	}
	<-doneCh
}

func (ctl *apiWatchProvider) syncLoop(ctx context.Context, worker int) {
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
			ref, shutdown := ctl.queue.Get()
			if shutdown {
				logger.V(2).Info("Queue shutdown")
				return
			}
			ctl.sync1(ctx, ref)
		}
	}
}

func (ctl *apiWatchProvider) sync1(ctx context.Context, ref any) {
	defer ctl.queue.Done(ref)
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Dequeued reference", "ref", ref)
	retry := ctl.sync(ctx, ref)
	if retry {
		ctl.queue.AddRateLimited(ref)
	} else {
		ctl.queue.Forget(ref)
	}
}

func (awp *apiWatchProvider) sync(ctx context.Context, refany any) bool {
	logger := klog.FromContext(ctx)
	switch typed := refany.(type) {
	case resourceRef:
		return awp.syncResourceRef(ctx, typed)
	case receiverForCluster[string /*group name*/, APIGroupInfo]:
		return awp.syncGroupReceiver(ctx, typed.cluster, typed.receiver)
	case receiverForCluster[metav1.GroupResource, ResourceDetails]:
		return awp.syncResourceReceiver(ctx, typed.cluster, typed.receiver)
	default:
		logger.Error(nil, "Got impossible type of object from workqueue", "obj", refany, "type", fmt.Sprintf("%T", refany))
		return false
	}
}

func (awp *apiWatchProvider) syncResourceRef(ctx context.Context, rr resourceRef) bool {
	logger := klog.FromContext(ctx)
	metarsc, receivers := func() (*ksmetav1a1.APIResource, MappingReceiverHolderFork[metav1.GroupResource, ResourceDetails]) {
		awp.Lock()
		defer awp.Unlock()
		wpc, has := awp.perCluster.Get(rr.cluster)
		if !has {
			logger.Error(nil, "Impossible: processing reference to unknown cluster", "rr", rr)
			return nil, []*MappingReceiverHolder[metav1.GroupResource, ResourceDetails]{}
		}
		metasrsc, err := wpc.lister.Get(rr.metaname)
		if err != nil && !k8sapierrors.IsNotFound(err) {
			logger.Error(err, "Impossible error fetching from local cache", "rr", rr)
		}
		return metasrsc, wpc.resourceReceivers
	}()
	externalizeReceiver(receivers)(metarsc)
	return false
}

// externalizeReceiver converts a receiver in terms internal to this package
// into a receiver of the external representation from apiwatch.
func externalizeReceiver(receiver MappingReceiver[metav1.GroupResource, ResourceDetails]) func(metarsc *ksmetav1a1.APIResource) {
	return func(metarsc *ksmetav1a1.APIResource) {
		key := metav1.GroupResource{Group: metarsc.Spec.Group, Resource: metarsc.Spec.Name}
		val := ResourceDetails{
			Namespaced:        metarsc.Spec.Namespaced,
			SupportsInformers: ResourceSupportsInformers(metarsc),
			PreferredVersion:  metarsc.Spec.Version}
		receiver.Put(key, val)
	}
}

func (awp *apiWatchProvider) syncGroupReceiver(ctx context.Context, cluster string, receiver MappingReceiver[string /*group name*/, APIGroupInfo]) bool {
	// TODO: implement, once apiwatch supplies this information
	logger := klog.FromContext(ctx)
	logger.V(4).Info("syncGroupReceiver not implemented")
	return false
}

func (awp *apiWatchProvider) syncResourceReceiver(ctx context.Context, cluster string, receiver MappingReceiver[metav1.GroupResource, ResourceDetails]) bool {
	logger := klog.FromContext(ctx)
	wpc := func() *apiWatchProviderPerCluster {
		awp.Lock()
		defer awp.Unlock()
		wpc, _ := awp.perCluster.Get(cluster)
		return wpc
	}()
	if wpc == nil {
		logger.Info("syncResourceReceiver did not find wpc, which may indicate a bug", "cluster", cluster)
		return false
	}
	resources, err := wpc.lister.List(labels.Everything())
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Impossible error listing from local cache", "cluster", wpc.cluster)
	}
	SliceApply(resources, externalizeReceiver(receiver))
	logger.V(4).Info("syncResourceReceiver done", "cluster", cluster, "numResources", len(resources))
	return false
}

func ResourceSupportsInformers(metarsc *ksmetav1a1.APIResource) bool {
	return SliceContains(metarsc.Spec.Verbs, "list") && SliceContains(metarsc.Spec.Verbs, "watch")
}
