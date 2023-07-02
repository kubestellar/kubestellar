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

package kindprovider

import (
	"context"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
	kind "sigs.k8s.io/kind/pkg/cluster"

	clusterprovider "github.com/kubestellar/kubestellar/pkg/clustermanager/providerclient"
)

// KindClusterProvider is a cluster provider that works with a local Kind instance.
type KindClusterProvider struct {
	kindProvider *kind.Provider
	providerName string
	watch        clusterprovider.Watcher
}

// New creates a new KindClusterProvider
func New(providerName string) KindClusterProvider {
	kindProvider := kind.NewProvider()
	return KindClusterProvider{
		kindProvider: kindProvider,
		providerName: providerName,
	}
}

func (k KindClusterProvider) Create(ctx context.Context, name string, opts clusterprovider.Options) error {
	logger := klog.FromContext(ctx)
	err := k.kindProvider.Create(name)

	if err != nil {
		// TODO:  Need to differentiate between "already exists" and an error
		logger.Error(err, "Provider.Create error")
	}

	return err
}

func (k KindClusterProvider) Delete(ctx context.Context,
	name string,
	opts clusterprovider.Options) error {

	return k.kindProvider.Delete(name, opts.KubeconfigPath)
}

func (k KindClusterProvider) ListClustersNames(ctx context.Context) ([]string, error) {
	list, err := k.kindProvider.List()
	if err != nil {
		return nil, err
	}
	logicalNameList := make([]string, 0, len(list))
	logicalNameList = append(logicalNameList, list...)
	return logicalNameList, err
}

func (k KindClusterProvider) Get(ctx context.Context, lcName string) (clusterprovider.LogicalClusterInfo, error) {
	cfg, err := k.kindProvider.KubeConfig(lcName, false)
	if err != nil {
		return clusterprovider.LogicalClusterInfo{}, err
	}

	lcInfo := clusterprovider.LogicalClusterInfo{
		Name:   lcName,
		Config: cfg,
	}
	return lcInfo, err
}

func (k KindClusterProvider) ListClusters(ctx context.Context) ([]clusterprovider.LogicalClusterInfo, error) {
	logger := klog.FromContext(ctx)
	lcNames, err := k.ListClustersNames(ctx)
	if err != nil {
		return nil, err
	}

	lcInfoList := make([]clusterprovider.LogicalClusterInfo, 0, len(lcNames))

	for _, lcName := range lcNames {
		cfg, err := k.kindProvider.KubeConfig(lcName, false)
		if err != nil {
			logger.Error(err, "couldn't fetch config for cluster")
		}

		lcInfoList = append(lcInfoList, clusterprovider.LogicalClusterInfo{
			Name:   lcName,
			Config: cfg,
		})
	}

	return lcInfoList, err
}

func (k KindClusterProvider) Watch() (clusterprovider.Watcher, error) {
	w := &KindWatcher{
		ch:       make(chan clusterprovider.WatchEvent),
		provider: &k}
	k.watch = w
	return w, nil
}

type KindWatcher struct {
	init     sync.Once
	wg       sync.WaitGroup
	ch       chan clusterprovider.WatchEvent
	cancel   context.CancelFunc
	provider *KindClusterProvider
}

func (k *KindWatcher) Stop() {
	if k.cancel != nil {
		k.cancel()
	}
	k.wg.Wait()
	close(k.ch)
}

func (k *KindWatcher) ResultChan() <-chan clusterprovider.WatchEvent {
	k.init.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		logger := klog.FromContext(ctx)
		k.cancel = cancel
		setClusters := sets.NewString()

		k.wg.Add(1)
		go func() {
			defer k.wg.Done()
			for {
				select {
				// TODO replace the 2 with a param at the cluster-provider-client level
				case <-time.After(5 * time.Second):
					list, err := k.provider.ListClustersNames(ctx)
					if err != nil {
						// TODO add logging
						logger.Error(err, "Getting provider list.")
						continue
					}
					newSetClusters := sets.NewString(list...)
					// Check for new clusters.
					for _, name := range newSetClusters.Difference(setClusters).UnsortedList() {
						logger.Info("Detected a new cluster")
						lcInfo, err := k.provider.Get(ctx, name)
						if err != nil {
							logger.Info("Kind cluster is not ready. Retrying.", "cluster", name)
							// Can't get the cluster info, so let's discover it again
							newSetClusters.Delete(name)
							continue
						}
						k.ch <- clusterprovider.WatchEvent{
							Type:   watch.Added,
							Name:   name,
							LCInfo: lcInfo,
						}
					}
					// Check for deleted clusters.
					for _, cl := range setClusters.Difference(newSetClusters).UnsortedList() {
						logger.Info("Detected cluster was deleted.")
						k.ch <- clusterprovider.WatchEvent{
							Type: watch.Deleted,
							Name: cl,
						}
					}
					setClusters = newSetClusters
				case <-ctx.Done():
					return
				}
			}
		}()
	})

	return k.ch
}
