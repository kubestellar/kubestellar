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

	clusterprovider "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager/providerclient"
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

func (k KindClusterProvider) Create(name string, opts clusterprovider.Options) error {
	logger := klog.Background()
	logger.V(2).Info("Creating Kind cluster", "name", name)
	err := k.kindProvider.Create(name)

	if err != nil {
		// TODO:  Need to differentiate between "already exists" and an error
		logger.Error(err, "Failed to create cluster", "name", name)
	}

	return err
}

func (k KindClusterProvider) Delete(name string, opts clusterprovider.Options) error {
	logger := klog.Background()
	logger.V(2).Info("Deleting kind cluster", "name", name)
	return k.kindProvider.Delete(name, "")
}

func (k KindClusterProvider) ListSpacesNames() ([]string, error) {
	list, err := k.kindProvider.List()
	if err != nil {
		return nil, err
	}
	logicalNameList := make([]string, 0, len(list))
	logicalNameList = append(logicalNameList, list...)
	return logicalNameList, err
}

func (k KindClusterProvider) Get(spaceName string) (clusterprovider.SpaceInfo, error) {
	cfg, err := k.kindProvider.KubeConfig(spaceName, false)
	if err != nil {
		return clusterprovider.SpaceInfo{}, err
	}

	spaceInfo := clusterprovider.SpaceInfo{
		Name:   spaceName,
		Config: cfg,
	}
	return spaceInfo, err
}

func (k KindClusterProvider) ListSpaces() ([]clusterprovider.SpaceInfo, error) {
	logger := klog.Background()
	spaceNames, err := k.ListSpacesNames()
	if err != nil {
		return nil, err
	}

	spaceInfoList := make([]clusterprovider.SpaceInfo, 0, len(spaceNames))

	for _, spaceName := range spaceNames {
		cfg, err := k.kindProvider.KubeConfig(spaceName, false)
		if err != nil {
			logger.Error(err, "Failed to fetch config for cluster", "name", spaceName)
		}

		spaceInfoList = append(spaceInfoList, clusterprovider.SpaceInfo{
			Name:   spaceName,
			Config: cfg,
		})
	}

	return spaceInfoList, err
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
				case <-time.After(2 * time.Second):
					list, err := k.provider.ListSpacesNames()
					if err != nil {
						logger.Error(err, "Failed to list Kind clusters")
						continue
					}
					newSetClusters := sets.NewString(list...)
					// Check for new clusters.
					for _, name := range newSetClusters.Difference(setClusters).UnsortedList() {
						logger.V(2).Info("Processing Kind cluster", "name", name)
						spaceInfo, err := k.provider.Get(name)
						if err != nil {
							logger.V(2).Info("Kind cluster is not ready. Retrying", "cluster", name)
							// Can't get the cluster info, so let's discover it again
							newSetClusters.Delete(name)
							continue
						}
						k.ch <- clusterprovider.WatchEvent{
							Type:      watch.Added,
							Name:      name,
							SpaceInfo: spaceInfo,
						}
					}
					// Check for deleted clusters.
					for _, cl := range setClusters.Difference(newSetClusters).UnsortedList() {
						logger.V(2).Info("Processing Kind cluster delete", "name", cl)
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
