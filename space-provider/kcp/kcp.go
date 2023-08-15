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
	"strings"
	"sync"

	"github.com/go-logr/logr"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/clientcmd"
	api "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"

	kcpcore "github.com/kcp-dev/kcp/pkg/apis/core"
	kcpcorev1alpha1 "github.com/kcp-dev/kcp/pkg/apis/core/v1alpha1"
	kcptenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	kcpcoreclusteredv1alpha1 "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster/typed/core/v1alpha1"
	kcptenancyclusteredv1alpha1 "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster/typed/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	clusterprovider "github.com/kubestellar/kubestellar/pkg/space-manager/providerclient"
)

// KcpClusterProvider is a cluster provider for KCP workspaces.
type KcpClusterProvider struct {
	ctx            context.Context
	logger         logr.Logger
	kcpConfig      string
	kcpWsClientset kcptenancyclusteredv1alpha1.WorkspaceClusterInterface
	kcpLcClientset kcpcoreclusteredv1alpha1.LogicalClusterClusterInterface
	workspaces     map[string]string
	watch          clusterprovider.Watcher
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
		kcpLcClientset: adminClientset.CoreV1alpha1().LogicalClusters(),
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
func (k KcpClusterProvider) ListSpacesNames() ([]string, error) {
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
			watcher, err := k.provider.kcpLcClientset.Watch(k.provider.ctx, metav1.ListOptions{})
			if err == nil {
				for {
					select {
					case <-ctx.Done():
						watcher.Stop()
						k.Stop()
					case event, ok := <-watcher.ResultChan():
						if !ok {
							k.Stop()
							return
						}
						lc := event.Object.(*kcpcorev1alpha1.LogicalCluster)

						path, ok := lc.Annotations[kcpcore.LogicalClusterPathAnnotationKey]
						if !ok {
							continue
						}
						if event.Type == "MODIFIED" || event.Type == "ADDED" {
							k.provider.logger.V(2).Info("KCP workspace modify/add event", "ws", path, "status", lc.Status.Phase)
							_, ok := k.provider.workspaces[path]
							if !ok && lc.Status.Phase == kcpcorev1alpha1.LogicalClusterPhaseReady {
								spaceInfo, err := k.provider.Get(path)
								if err != nil {
									k.provider.logger.Info("Failed to get space info")
									continue
								}
								k.provider.logger.V(2).Info("New KCP workspace is ready", "ws", path)
								// add ready WS to cache and send an event
								k.provider.workspaces[path] = string(lc.Status.Phase)
								k.ch <- clusterprovider.WatchEvent{
									Type:      watch.Added,
									Name:      lc.Spec.Owner.Name,
									SpaceInfo: spaceInfo,
								}
							}
							if ok && lc.Status.Phase != kcpcorev1alpha1.LogicalClusterPhaseReady {
								k.provider.logger.V(2).Info("KCP workspace is not ready")
								if ok {
									delete(k.provider.workspaces, path)
									k.ch <- clusterprovider.WatchEvent{
										Type: watch.Deleted,
										Name: lc.Spec.Owner.Name,
									}
								}
							}
						}
						if event.Type == "DELETED" {
							k.provider.logger.V(2).Info("KCP workspace deleted", "ws", path)
							delete(k.provider.workspaces, path)
							k.ch <- clusterprovider.WatchEvent{
								Type: watch.Deleted,
								Name: lc.Spec.Owner.Name,
							}
						}
					}
				}
			}
		}()
	})
	return k.ch
}

func buildRawConfig(baseRaw api.Config, spaceName string) api.Config {
	main := "root"
	// remove all clusters and contexts exept main cluster/context
	clusters := make(map[string]*api.Cluster)
	contexts := make(map[string]*api.Context)
	contexts[main] = baseRaw.Contexts[main]
	// modify server path
	baseRaw.Clusters[main].Server = strings.ReplaceAll(baseRaw.Clusters[main].Server, main, spaceName)
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
