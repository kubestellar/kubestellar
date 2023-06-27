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
	clusterprovider "github.com/kubestellar/kubestellar/pkg/clustermanager/providerclient"
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
	newProvider, err := NewProvider(c.context, providerName, providerType)
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
			logger.Info("Stopping cluster provider watch", "provider", providerName)
			// TODO: return an error
			return
		}
		lcName := event.Name
		reflcluster, err := clientset.LogicalclusterV1alpha1().LogicalClusters(GetNamespace(providerName)).Get(ctx, lcName, v1.GetOptions{})
		found := reflcluster != nil && err != nil

		switch event.Type {
		case watch.Added:
			logger.Info("New cluster was detected", "cluster", event.Name)
			// A new cluster was detected either create it or change the status to READY
			if !found {
				// No corresponding Logicalcluster, let's create it
				logger.Info("Creating new LogicalCluster object", "cluster", event.Name)
				lcluster := lcv1alpha1apis.LogicalCluster{}
				lcluster.Name = lcName
				lcluster.Spec.ClusterProviderDescName = providerName
				lcluster.Spec.Managed = false
				lcluster.Status.Phase = lcv1alpha1apis.LogicalClusterPhaseReady
				_, err = clientset.LogicalclusterV1alpha1().LogicalClusters(GetNamespace(providerName)).Create(ctx, &lcluster, v1.CreateOptions{})
				if err != nil {
					// TODO: how do we handle failure?
					logger.Error(err, "New cluster detected,  couldn't create the corresponding LogicalCluster")
					return
				}
			} else {
				// TODO: when finalizers added - cheeck the logicalcluster delete timestamp
				// Logical cluster exists , just update its status
				reflcluster.Status.Phase = lcv1alpha1apis.LogicalClusterPhaseReady
				_, err = clientset.LogicalclusterV1alpha1().LogicalClusters(GetNamespace(providerName)).UpdateStatus(ctx, reflcluster, v1.UpdateOptions{})
				if err != nil {
					// TODO: how do we handle failure?
					logger.Error(err, "New cluster detected, Couldn't update the LogicalCluster status")
					return
				}
			}

		case watch.Deleted:
			logger.Info("A cluster was removed", "cluster", event.Name)
			if found {
				reflcluster.Status.Phase = lcv1alpha1apis.LogicalClusterPhaseNotReady
				_, err = clientset.LogicalclusterV1alpha1().LogicalClusters(GetNamespace(providerName)).UpdateStatus(ctx, reflcluster, v1.UpdateOptions{})
				if err != nil {
					// TODO: how do we handle failure?
					logger.Error(err, "Cluster was removed, Couldn't update the LogicalCluster status")
					return
				}
				//TODO: When using finalizers we need to remove the finalizer in this point.
			}

		default:
			logger.Info("unknown event type", "type", event.Type)
		}
	}
}
