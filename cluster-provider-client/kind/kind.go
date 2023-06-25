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
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
	kind "sigs.k8s.io/kind/pkg/cluster"

	clusterprovider "github.com/kubestellar/kubestellar/cluster-provider-client/cluster"
)

// KindClusterProvider is a cluster provider that works with a local Kind instance.
type KindClusterProvider struct {
	kindProvider *kind.Provider
	providerName string
}

// New creates a new KindClusterProvider
func New(providerName string) KindClusterProvider {
	kindProvider := kind.NewProvider()
	return KindClusterProvider{
		kindProvider: kindProvider,
		providerName: providerName,
	}
}

func (k KindClusterProvider) Create(ctx context.Context,
	name string,
	opts clusterprovider.Options) (clusterprovider.LogicalClusterInfo, error) {
	var resCluster clusterprovider.LogicalClusterInfo

	err := k.kindProvider.Create(string(name), kind.CreateWithKubeconfigPath(opts.KubeconfigPath))
	if err != nil {
		if strings.HasPrefix(err.Error(), "node(s) already exist for a cluster with the name") {
			// TODO: check whether it's the same cluster and return success if true
		} else {
			return resCluster, err
		}
	}

	cfg, err := k.kindProvider.KubeConfig(string(name), false)
	if err != nil {
		return resCluster, err
	}
	resCluster = *clusterprovider.New(cfg, opts)
	return resCluster, err
}

func (k KindClusterProvider) Delete(ctx context.Context,
	name string,
	opts clusterprovider.Options) error {

	return k.kindProvider.Delete(string(name), opts.KubeconfigPath)
}

func (k KindClusterProvider) List() ([]string, error) {
	list, err := k.kindProvider.List()
	if err != nil {
		return nil, err
	}
	logicalNameList := make([]string, 0, len(list))
	logicalNameList = append(logicalNameList, list...)
	return logicalNameList, err
}

func (k KindClusterProvider) Watch() (clusterprovider.Watcher, error) {
	return &KindWatcher{
		ch:       make(chan clusterprovider.WatchEvent),
		provider: &k}, nil
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
					list, err := k.provider.List()
					if err != nil {
						// TODO add logging
						logger.Error(err, "Getting provider list.")
						continue
					}
					newSetClusters := sets.NewString(list...)
					// Check for new clusters.
					for _, cl := range newSetClusters.Difference(setClusters).UnsortedList() {
						logger.Info("Detected a new cluster")
						k.ch <- clusterprovider.WatchEvent{
							Type: watch.Added,
							Name: cl,
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
