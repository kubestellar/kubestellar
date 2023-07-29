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

package kflexprovider

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"

	clusterprovider "github.com/kubestellar/kubestellar/pkg/space-manager/providerclient"
)

// KflexClusterProvider is a kubeflex cluster provider
type KflexClusterProvider struct {
	providerName string
	watch        clusterprovider.Watcher
}

// New creates a new KflexClusterProvider
func New(providerName string) KflexClusterProvider {
	return KflexClusterProvider{
		providerName: providerName,
	}
}

// TODO: switch from CLI to getting this info via kube directives
func (k KflexClusterProvider) Create(name string, opts clusterprovider.Options) error {
	logger := klog.Background()
	logger.V(2).Info("Creating KubeFlex cluster", "name", name)
	cmd := exec.Command("kflex", "create", name)
	err := cmd.Run()

	if err != nil {
		// TODO:  Need to differentiate between "already exists" and an error
		logger.Error(err, "Failed to create cluster", "name", name)
	}

	return err
}

// TODO: switch from CLI to getting this info via kube directives
func (k KflexClusterProvider) Delete(name string, opts clusterprovider.Options) error {
	logger := klog.Background()
	logger.V(2).Info("Deleting KubeFlex cluster", "name", name)
	cmd := exec.Command("kflex", "delete", name)
	err := cmd.Run()
	if err != nil {
		logger.V(2).Error(err, "Deleting KubeFlex cluster", "name", name)
	}
	return err
}

// ListSpacesNames: returns a list of clusters in KubeFlex that are
// in the Ready condition.
// TODO: switch from CLI to getting this info via kube directives
func (k KflexClusterProvider) ListSpacesNames() ([]string, error) {
	var out bytes.Buffer
	var listClusterNames []string
	cmd := exec.Command("kubectl", "--context", "kind-kubeflex", "get", "controlplanes", "--no-headers")
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	b := out.Bytes()
	list := strings.Split(string(b), "\n")
	for _, v := range list {
		words := strings.Fields(v)
		if len(words) > 2 && words[2] == "True" {
			listClusterNames = append(listClusterNames, words[0])
		}
	}
	return listClusterNames, nil
}

// Get: obtains the kubeconfig for the given lcName cluster.
// TODO: switch from cli to kube directives
func (k KflexClusterProvider) Get(lcName string) (clusterprovider.SpaceInfo, error) {
	cmd := "kubectl --context kind-kubeflex get secrets -n lc3-system admin-kubeconfig -o jsonpath='{.data.*}' | base64 -d"
	cfg, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return clusterprovider.SpaceInfo{}, err
	}
	lcInfo := clusterprovider.SpaceInfo{
		Name:   lcName,
		Config: string(cfg),
	}
	return lcInfo, nil
}

func (k KflexClusterProvider) ListSpaces() ([]clusterprovider.SpaceInfo, error) {
	logger := klog.Background()
	lcNames, _ := k.ListSpacesNames()

	lcInfoList := make([]clusterprovider.SpaceInfo, 0, len(lcNames))

	for _, lcName := range lcNames {
		// TODO: switch from cli to kube directives
		cmd := "kubectl --context kind-kubeflex get secrets -n lc3-system admin-kubeconfig -o jsonpath='{.data.*}' | base64 -d"
		cfg, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			logger.Error(err, "Failed to fetch config for cluster", "name", lcName)
		}

		lcInfoList = append(lcInfoList, clusterprovider.SpaceInfo{
			Name:   lcName,
			Config: string(cfg),
		})
	}

	return lcInfoList, nil
}

func (k KflexClusterProvider) Watch() (clusterprovider.Watcher, error) {
	w := &KflexWatcher{
		ch:       make(chan clusterprovider.WatchEvent),
		provider: &k}
	k.watch = w
	return w, nil
}

type KflexWatcher struct {
	init     sync.Once
	wg       sync.WaitGroup
	ch       chan clusterprovider.WatchEvent
	cancel   context.CancelFunc
	provider *KflexClusterProvider
}

func (k *KflexWatcher) Stop() {
	if k.cancel != nil {
		k.cancel()
	}
	k.wg.Wait()
	close(k.ch)
}

func (k *KflexWatcher) ResultChan() <-chan clusterprovider.WatchEvent {
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
						logger.Error(err, "Failed to list KubeFlex clusters")
						continue
					}
					newSetClusters := sets.NewString(list...)
					// Check for new clusters.
					for _, name := range newSetClusters.Difference(setClusters).UnsortedList() {
						logger.V(2).Info("Processing KubeFlex cluster", "name", name)
						spaceInfo, err := k.provider.Get(name)
						if err != nil {
							logger.V(2).Info("KubeFlex cluster is not ready. Retrying", "cluster", name)
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
						logger.V(2).Info("Processing KubeFlex cluster delete", "name", cl)
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
