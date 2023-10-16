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

package providerkcp

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-logr/logr"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	kcpcore "github.com/kcp-dev/kcp/pkg/apis/core"
	kcpcorev1alpha1 "github.com/kcp-dev/kcp/pkg/apis/core/v1alpha1"
	kcptenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	kcptenancyclusteredv1alpha1 "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster/typed/tenancy/v1alpha1"
	extkcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	"github.com/kcp-dev/logicalcluster/v3"

	clusterprovider "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager/providerclient"
)

// KcpClusterProvider is a cluster provider for KCP workspaces.
type KcpClusterProvider struct {
	ctx            context.Context
	logger         logr.Logger
	kcpConfig      string
	kcpWsClientset kcptenancyclusteredv1alpha1.WorkspaceClusterInterface
	adminClientset *kcpclientset.ClusterClientset
	workspaces     map[string]string
	watch          clusterprovider.Watcher
	lock           sync.Mutex
}

// New returns a new KcpClusterProvider
func New(kcpConfig string) (*KcpClusterProvider, error) {
	ctx := context.Background()
	logger := klog.FromContext(ctx)

	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(kcpConfig))
	if err != nil {
		return nil, err
	}
	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return nil, err
	}
	baseConfig, err := clientcmd.NewNonInteractiveClientConfig(rawConfig, "base", nil, nil).ClientConfig()
	if err != nil {
		return nil, err
	}
	baseClientset, err := kcpclientset.NewForConfig(baseConfig)
	if err != nil {
		return nil, err
	}
	adminConfig, err := clientcmd.NewNonInteractiveClientConfig(rawConfig, "system:admin", nil, nil).ClientConfig()
	if err != nil {
		return nil, err
	}
	adminClientset, err := kcpclientset.NewForConfig(adminConfig)
	if err != nil {
		return nil, err
	}
	c := &KcpClusterProvider{
		ctx:            ctx,
		logger:         logger,
		kcpConfig:      kcpConfig,
		kcpWsClientset: baseClientset.TenancyV1alpha1().Workspaces(),
		adminClientset: adminClientset,
		workspaces:     make(map[string]string),
	}
	return c, nil
}

func (k *KcpClusterProvider) Create(name string, opts clusterprovider.Options) error {
	k.logger.V(2).Info("Creating KCP workspace", "name", name)
	ws := &kcptenancyv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: kcptenancyv1alpha1.WorkspaceSpec{},
	}
	_, err := k.kcpWsClientset.Cluster(getClusterPath(opts.Parent)).Create(k.ctx, ws, metav1.CreateOptions{})
	if err != nil && !k8sapierrors.IsAlreadyExists(err) {
		k.logger.Error(err, "Failed to create KCP workspace", "space", name)
		return err
	}
	k.logger.V(2).Info("Created KCP workspace for space", "name", name)
	return nil
}

func (k *KcpClusterProvider) Delete(name string, opts clusterprovider.Options) error {
	k.logger.V(2).Info("Deleting KCP workspace", "name", name)
	return k.kcpWsClientset.Cluster(getClusterPath(opts.Parent)).Delete(k.ctx, name, metav1.DeleteOptions{})
}

// ListSpacesNames is N/A for KCP
// TODO remove it from ProviderClient interface
func (k *KcpClusterProvider) ListSpacesNames() ([]string, error) {
	k.logger.V(2).Info("Not implemented")
	return []string{}, nil
}

func (k *KcpClusterProvider) Get(spaceName string) (clusterprovider.SpaceInfo, error) {

	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(k.kcpConfig))
	if err != nil {
		k.logger.Error(err, "Failed to get client config for workspace", "workspace", spaceName)
	}
	cfg, err := clientConfig.RawConfig()
	if err != nil {
		k.logger.Error(err, "Failed to get raw config for workspace", "workspace", spaceName)
	}
	cfgBytes, err := clientcmd.Write(buildRawConfig(cfg, spaceName))
	if err != nil {
		k.logger.Error(err, "Failed to write workspace config", "workspace", spaceName)
	}

	spaceInfo := clusterprovider.SpaceInfo{
		Name:   spaceName,
		Config: string(cfgBytes[:]),
	}
	return spaceInfo, err
}

func (k *KcpClusterProvider) ListSpaces() ([]clusterprovider.SpaceInfo, error) {
	list, err := k.kcpWsClientset.Cluster(kcpcore.RootCluster.Path()).List(k.ctx, metav1.ListOptions{})
	if err != nil {
		return []clusterprovider.SpaceInfo{}, err
	}
	spaceInfoList := make([]clusterprovider.SpaceInfo, 0, len(list.Items))

	for _, ws := range list.Items {
		spaceInfo, err := k.Get(ws.Name)
		if err != nil {
			k.logger.Error(err, "Failed to fetch config for workspace", "name", ws.Name)
		}

		spaceInfoList = append(spaceInfoList, spaceInfo)
	}
	return spaceInfoList, nil
}

func (k *KcpClusterProvider) Watch() (clusterprovider.Watcher, error) {
	w := &KcpWatcher{
		ch:       make(chan clusterprovider.WatchEvent),
		provider: k}
	k.watch = w
	return w, nil
}

type KcpWatcher struct {
	init     sync.Once
	wg       sync.WaitGroup
	ch       chan clusterprovider.WatchEvent
	cancel   context.CancelFunc
	provider *KcpClusterProvider
}

func (k *KcpWatcher) Stop() {
	if k.cancel != nil {
		k.cancel()
	}
	k.wg.Wait()
	close(k.ch)
}

func (k *KcpWatcher) ResultChan() <-chan clusterprovider.WatchEvent {
	k.init.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		k.cancel = cancel

		k.wg.Add(1)
		go func() {
			defer k.wg.Done()
			numThreads := 2
			resyncPeriod := 30 * time.Second
			kcpInformerFactory := extkcpinformers.NewSharedInformerFactory(k.provider.adminClientset, resyncPeriod)
			lcInformer := kcpInformerFactory.Core().V1alpha1().LogicalClusters().Informer()

			kcpInformerFactory.Start(ctx.Done())

			c := NewController(ctx, k, lcInformer)
			c.Run(numThreads)
		}()
	})
	return k.ch
}

type queueItem struct {
	key   string
	path  string
	owner *kcpcorev1alpha1.LogicalClusterOwner
}

type controller struct {
	ctx        context.Context
	logger     logr.Logger
	watcher    *KcpWatcher
	queue      workqueue.RateLimitingInterface
	lcInformer kcpcache.ScopeableSharedIndexInformer
}

// NewController returns new KCP-provider controller
func NewController(
	ctx context.Context,
	watcher *KcpWatcher,
	lcInformer kcpcache.ScopeableSharedIndexInformer,
) *controller {
	controllerName := "KCP-provider"
	ctx = klog.NewContext(ctx, klog.FromContext(ctx).WithValues("controller", controllerName))

	c := &controller{
		ctx:        ctx,
		logger:     klog.FromContext(ctx),
		watcher:    watcher,
		queue:      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
		lcInformer: lcInformer,
	}

	lcInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.enqueue,
		UpdateFunc: func(oldObj, newObj interface{}) {
			old := oldObj.(*kcpcorev1alpha1.LogicalCluster)
			new := newObj.(*kcpcorev1alpha1.LogicalCluster)
			if !reflect.DeepEqual(old, new) {
				c.enqueue(newObj)
			}
		},
		DeleteFunc: c.enqueue,
	})
	return c
}

func (c *controller) enqueue(obj interface{}) {
	key, err := kcpcache.DeletionHandlingMetaClusterNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(err)
		return
	}
	lc, ok := obj.(*kcpcorev1alpha1.LogicalCluster)
	if !ok {
		runtime.HandleError(errors.New("unexpected object type. expected LogicalCluster"))
		return
	}
	path, ok := lc.Annotations[kcpcore.LogicalClusterPathAnnotationKey]
	if ok && lc.Spec.Owner != nil {
		c.logger.V(4).Info("queueing LogicalCluster", "key", key)
		c.queue.Add(
			queueItem{
				key:   key,
				path:  path,
				owner: lc.Spec.Owner,
			},
		)
	}

}

// Run starts the controller, which stops when c.context.Done() is closed.
func (c *controller) Run(numThreads int) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.logger.Info("starting KCP provider controller")
	defer c.logger.Info("shutting down KCP provider controller")

	for i := 0; i < numThreads; i++ {
		go wait.UntilWithContext(c.ctx, c.runWorker, time.Second)
	}

	<-c.ctx.Done()
}

func (c *controller) runWorker(ctx context.Context) {
	for c.processNextItem() {
	}
}

func (c *controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	i, quit := c.queue.Get()
	if quit {
		return false
	}
	item := i.(queueItem)

	// Done with this key, unblock other workers.
	defer c.queue.Done(i)

	if err := c.process(item); err != nil {
		runtime.HandleError(err)
		c.queue.AddRateLimited(i)
		return true
	}
	c.queue.Forget(i)
	return true
}

func (c *controller) process(item queueItem) error {
	lc, exists, err := c.lcInformer.GetIndexer().GetByKey(item.key)
	if err != nil {
		return err
	}
	if !exists {
		c.handleDelete(item)
	} else {
		c.handleAdd(lc, item)
	}
	return nil
}

func (c *controller) handleAdd(logicalCluster interface{}, item queueItem) {
	lc, ok := logicalCluster.(*kcpcorev1alpha1.LogicalCluster)
	if !ok {
		runtime.HandleError(errors.New("unexpected object type. expected LogicalCluster"))
		return
	}
	path, owner := item.path, item.owner
	c.logger.V(2).Info("KCP workspace modify/add event", "ws", path, "status", lc.Status.Phase)
	c.watcher.provider.lock.Lock()
	defer c.watcher.provider.lock.Unlock()

	_, ok = c.watcher.provider.workspaces[path]
	if !ok && lc.Status.Phase == kcpcorev1alpha1.LogicalClusterPhaseReady {
		spaceInfo, err := c.watcher.provider.Get(path)
		if err != nil {
			c.logger.Info("Failed to get space info")
			return
		}
		c.logger.V(2).Info("New KCP workspace is ready", "ws", path)
		// add ready WS to cache and send an event
		c.watcher.provider.workspaces[path] = string(lc.Status.Phase)
		c.watcher.ch <- clusterprovider.WatchEvent{
			Type:      watch.Added,
			Name:      owner.Name,
			SpaceInfo: spaceInfo,
		}
	}
	if ok && lc.Status.Phase != kcpcorev1alpha1.LogicalClusterPhaseReady {
		c.logger.V(2).Info("KCP workspace is not ready")
		if ok {
			delete(c.watcher.provider.workspaces, path)
			c.watcher.ch <- clusterprovider.WatchEvent{
				Type: watch.Deleted,
				Name: lc.Spec.Owner.Name,
			}
		}
	}

}

func (c *controller) handleDelete(item queueItem) {
	c.logger.V(2).Info("KCP workspace deleted", "ws", item.path)
	c.watcher.provider.lock.Lock()
	defer c.watcher.provider.lock.Unlock()
	delete(c.watcher.provider.workspaces, item.path)
	c.watcher.ch <- clusterprovider.WatchEvent{
		Type: watch.Deleted,
		Name: item.owner.Name,
	}
}

func buildRawConfig(baseRaw api.Config, spaceName string) api.Config {
	main := "root"
	delimiter := ":"
	// remove all clusters and contexts exept main cluster/context
	clusters := make(map[string]*api.Cluster)
	contexts := make(map[string]*api.Context)
	contexts[main] = baseRaw.Contexts[main]
	// modify server path
	if strings.HasPrefix(spaceName, main+delimiter) {
		// spaceName is full path
		baseRaw.Clusters[main].Server = strings.ReplaceAll(baseRaw.Clusters[main].Server, main, spaceName)
	} else {
		baseRaw.Clusters[main].Server = strings.ReplaceAll(baseRaw.Clusters[main].Server, main, main+delimiter+spaceName)
	}

	clusters[main] = baseRaw.Clusters[main]
	baseRaw.Clusters = clusters
	baseRaw.Contexts = contexts
	baseRaw.CurrentContext = main
	return baseRaw
}

func getClusterPath(parent string) logicalcluster.Path {
	clusterPath := kcpcore.RootCluster.Path()
	if parent != "" {
		parentCluster := logicalcluster.Name(parent)
		clusterPath = parentCluster.Path()
	}
	return clusterPath
}
