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

package clustermanager

import (
	"context"
	"errors"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"

	lcv1alpha1apis "github.com/kubestellar/kubestellar/pkg/apis/logicalcluster/v1alpha1"
	edgeclient "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"

	//	clusterproviderclient "github.com/kubestellar/kubestellar/pkg/clustermanager/provider-client-interface"
	clusterproviderclient "github.com/kubestellar/kubestellar/pkg/clustermanager/provider-client-interface"
	clusterprovider "github.com/kubestellar/kubestellar/pkg/clustermanager/provider-client-interface/cluster"
)

// GetProvider returns provider client interface for provider
func (c *controller) GetProvider(providerName string) (clusterprovider.ProviderClient, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	provider, exists := c.providers[providerName]
	if !exists {
		// If the provider does not exists in the list, then likely the provider object was
		// recentely added, and the provider is in the process of being added.  Return an
		// error. The caller, which we expect is the logical cluster reconciler, will requeue
		// the logical cluster request that triggered the GetProvider call.
		err := errors.New("provider does not exist in the provider list")
		return nil, err
	}
	return provider.providerClient, nil
}

// CreateProvider returns new provider client
func (c *controller) CreateProvider(
	providerName string, providerType lcv1alpha1apis.ClusterProviderType) (clusterprovider.ProviderClient, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, exists := c.providers[providerName]
	if exists {
		err := errors.New("provider already in the list")
		return nil, err
	}
	newProvider, err := clusterproviderclient.NewProvider(c.context, providerName, providerType)
	if err != nil {
		return nil, err
	}
	c.providers[providerName] = providerInfo{providerClient: newProvider}
	return newProvider, nil
}

// StartDiscovery will start watching provider clusters for changes
func (c *controller) StartDiscovery(providerName string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	providerInfo, ok := c.providers[providerName]
	if !ok {
		return errors.New("failed to start provider discovery. provider info not exists")
	}

	watcher, err := providerInfo.providerClient.Watch()
	if err != nil {
		return err
	}
	go processProviderWatchEvents(c.context, watcher, c.clientset, providerName)

	providerInfo.providerWatcher = watcher
	c.providers[providerName] = providerInfo
	return nil
}

// StopDiscovery will stop watching provider clusters for changes
func (c *controller) StopDiscovery(providerName string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	providerInfo, ok := c.providers[providerName]
	if !ok {
		return errors.New("failed to stop provider discovery. provider info does not exist")
	}

	if providerInfo.providerWatcher == nil {
		return errors.New("failed to stop provider discovery. provider watcher does not exist")
	}
	providerInfo.providerWatcher.Stop()
	return nil
}

func processProviderWatchEvents(ctx context.Context, w clusterprovider.Watcher, clientset edgeclient.Interface, providerName string) {
	logger := klog.FromContext(ctx)
	for {
		event, ok := <-w.ResultChan()
		if !ok {
			w.Stop()
			logger.Info("stopping cluster provider watch", "provider", providerName)
			// TODO: return an error
			return
		}
		listLogicalClusters, err := clientset.LogicalclusterV1alpha1().LogicalClusters(clusterproviderclient.GetNamespace(providerName)).List(ctx, v1.ListOptions{})
		if err != nil {
			logger.Error(err, "")
			// TODO: how do we handle failure?
			return
		}
		switch event.Type {
		case watch.Added:
			// TODO: I am currently ignoring the possibility of the logical cluster object already existing
			var found bool = false
			for _, logicalCluster := range listLogicalClusters.Items {
				if logicalCluster.Name == event.Name {
					found = true
					break
				}
			}
			if !found {
				logger.Info("Creating new LogicalCluster object", "cluster", event.Name)
				var eventLogicalCluster lcv1alpha1apis.LogicalCluster
				eventLogicalCluster.Name = event.Name
				eventLogicalCluster.Spec.ClusterProviderDescName = providerName
				eventLogicalCluster.Spec.Managed = false
				eventLogicalCluster.Status.Phase = "Initializing"
				_, err = clientset.LogicalclusterV1alpha1().LogicalClusters(clusterproviderclient.GetNamespace(providerName)).Create(ctx, &eventLogicalCluster, v1.CreateOptions{})
				if err != nil {
					logger.Error(err, "")
					// TODO: how do we handle failure?
					return
				}
			}
		case watch.Deleted:
			logger.Info("Deleting LogicalCluster object", "cluster", event.Name)
			err := clientset.LogicalclusterV1alpha1().LogicalClusters(clusterproviderclient.GetNamespace(providerName)).Delete(ctx, event.Name, v1.DeleteOptions{})
			if err != nil {
				// TODO: If the logical cluster object does not exist, ignore the error.
				logger.Error(err, "")
				// TODO: how do we handle failure?
				return
			}

		default:
			logger.Info("unknown event type", "type", event.Type)
		}
	}
}
