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

	kcpv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/core/v1alpha1"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/typed/tenancy/v1alpha1"

	clusterprovider "github.com/kubestellar/kubestellar/pkg/clustermanager/providerclient"
)

// KcpClusterProvider is a cluster provider for KCP workspaces.
type KcpClusterProvider struct {
	ctx        context.Context
	logger     logr.Logger
	kcpConfig  string
	clientset  tenancyclient.WorkspaceInterface
	workspaces map[string]string
	watch      clusterprovider.Watcher
}

// New returns a new KcpClusterProvider
func New(kcpConfig string) (*KcpClusterProvider, error) {
	ctx := context.Background()
	logger := klog.FromContext(ctx)
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kcpConfig))
	if err != nil {
		return nil, err
	}
	clientset, err := tenancyclient.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	c := &KcpClusterProvider{
		ctx:        ctx,
		logger:     logger,
		kcpConfig:  kcpConfig,
		clientset:  clientset.Workspaces(),
		workspaces: make(map[string]string),
	}
	return c, nil
}

func (k *KcpClusterProvider) Create(name string, opts clusterprovider.Options) error {
	k.logger.V(2).Info("Creating KCP workspace", "name", name)
	ws := &tenancyv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: tenancyv1alpha1.WorkspaceSpec{},
	}
	_, err := k.clientset.Create(k.ctx, ws, metav1.CreateOptions{})
	if err != nil && !k8sapierrors.IsAlreadyExists(err) {
		return err
	}
	k.logger.V(2).Info("Created KCP workspace for logical cluster", "name", name)
	return nil
}

func (k *KcpClusterProvider) Delete(name string, opts clusterprovider.Options) error {
	k.logger.V(2).Info("Deleting KCP workspace", "name", name)
	return k.clientset.Delete(k.ctx, name, metav1.DeleteOptions{})
}

// ListClustersNames is N/A for KCP
// TODO remove it from ProviderClient interface
func (k KcpClusterProvider) ListClustersNames() ([]string, error) {
	k.logger.V(2).Info("Not implemented")
	return []string{}, nil
}

func (k *KcpClusterProvider) Get(lcName string) (clusterprovider.LogicalClusterInfo, error) {

	clientConfig, err := clientcmd.NewClientConfigFromBytes([]byte(k.kcpConfig))
	if err != nil {
		k.logger.Error(err, "Failed to get client config for workspace", "workspace", lcName)
	}
	cfg, err := clientConfig.RawConfig()
	if err != nil {
		k.logger.Error(err, "Failed to get raw config for workspace", "workspace", lcName)
	}
	cfgBytes, err := clientcmd.Write(buildRawConfig(cfg, lcName))
	if err != nil {
		k.logger.Error(err, "Failed to write workspace config", "workspace", lcName)
	}

	lcInfo := clusterprovider.LogicalClusterInfo{
		Name:   lcName,
		Config: string(cfgBytes[:]),
	}
	return lcInfo, err
}

func (k *KcpClusterProvider) ListClusters() ([]clusterprovider.LogicalClusterInfo, error) {
	list, err := k.clientset.List(k.ctx, metav1.ListOptions{})
	if err != nil {
		return []clusterprovider.LogicalClusterInfo{}, err
	}
	lcInfoList := make([]clusterprovider.LogicalClusterInfo, 0, len(list.Items))

	for _, ws := range list.Items {
		lcInfo, err := k.Get(ws.Name)
		if err != nil {
			k.logger.Error(err, "Failed to fetch config for workspace", "name", ws.Name)
		}

		lcInfoList = append(lcInfoList, lcInfo)
	}
	return lcInfoList, nil
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
			watcher, err := k.provider.clientset.Watch(k.provider.ctx, metav1.ListOptions{})
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
						ws := event.Object.(*tenancyv1alpha1.Workspace)
						name := ws.Name
						if event.Type == "MODIFIED" {
							k.provider.logger.V(2).Info("KCP workspace modify event", "ws", ws.Name)

							_, ok := k.provider.workspaces[name]
							if !ok && ws.Status.Phase == kcpv1alpha1.LogicalClusterPhaseReady {
								lcInfo, err := k.provider.Get(ws.Name)
								if err != nil {
									continue
								}
								k.provider.logger.V(2).Info("New KCP workspace is ready", "ws", ws.Name)
								// add ready WS to cache and send an event
								k.provider.workspaces[name] = string(ws.Status.Phase)
								k.ch <- clusterprovider.WatchEvent{
									Type:   watch.Added,
									Name:   ws.Name,
									LCInfo: lcInfo,
								}
							}
							if ok && ws.Status.Phase != kcpv1alpha1.LogicalClusterPhaseReady {
								k.provider.logger.V(2).Info("KCP workspace is not ready")
								if ok {
									delete(k.provider.workspaces, name)
									k.ch <- clusterprovider.WatchEvent{
										Type: watch.Deleted,
										Name: name,
									}
								}
							}
						}
						if event.Type == "DELETED" {
							k.provider.logger.V(2).Info("KCP workspace deleted", "ws", ws.Name)
							delete(k.provider.workspaces, name)
							k.ch <- clusterprovider.WatchEvent{
								Type: watch.Deleted,
								Name: ws.Name,
							}
						}
					}
				}
			}
		}()
	})
	return k.ch
}

func buildRawConfig(baseRaw api.Config, lcName string) api.Config {
	main := "root"
	// remove all clusters and contexts exept main cluster/context
	clusters := make(map[string]*api.Cluster)
	contexts := make(map[string]*api.Context)
	contexts[main] = baseRaw.Contexts[main]
	// modify server path
	baseRaw.Clusters[main].Server = strings.ReplaceAll(baseRaw.Clusters[main].Server, main, main+":"+lcName)
	clusters[main] = baseRaw.Clusters[main]
	baseRaw.Clusters = clusters
	baseRaw.Contexts = contexts
	baseRaw.CurrentContext = main
	return baseRaw
}
