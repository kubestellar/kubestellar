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

package clusterproviderclient

import (
	"context"
	"errors"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"

	clusterprovider "github.com/kcp-dev/edge-mc/cluster-provider-client/cluster"
	kindprovider "github.com/kcp-dev/edge-mc/cluster-provider-client/kind"
	v1alpha1apis "github.com/kcp-dev/edge-mc/pkg/apis/logicalcluster/v1alpha1"
	edgeclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned"
)

// Each provider gets its own namespace named prefixNamespace+providerName
const prefixNamespace = "lcprovider-"

func GetNamespace(providerName string) string {
	return prefixNamespace + providerName
}

func ProcessProviderWatchEvents(ctx context.Context, w clusterprovider.Watcher, clientset edgeclient.Interface, providerName string) {
	logger := klog.FromContext(ctx)
	for {
		event, ok := <-w.ResultChan()
		if !ok {
			w.Stop()
			logger.Info("stopping")
			// TODO: return an error
			return
		}
		listLogicalClusters, err := clientset.LogicalclusterV1alpha1().LogicalClusters(providerName).List(ctx, v1.ListOptions{})
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
				logger.Info("Creating new LogicalCluster object", event.Name)
				var eventLogicalCluster v1alpha1apis.LogicalCluster
				eventLogicalCluster.Name = event.Name
				eventLogicalCluster.Spec.ClusterProviderDesc = providerName
				eventLogicalCluster.Spec.Managed = false
				eventLogicalCluster.Status.Phase = "Initializing"
				_, err = clientset.LogicalclusterV1alpha1().LogicalClusters(providerName).Create(ctx, &eventLogicalCluster, v1.CreateOptions{})
				if err != nil {
					logger.Error(err, "")
					// TODO: how do we handle failure?
					return
				}
			}
		case watch.Deleted:
			logger.Info("Deleting LogicalCluster object", event.Name)
			err := clientset.LogicalclusterV1alpha1().LogicalClusters(providerName).Delete(ctx, event.Name, v1.DeleteOptions{})
			if err != nil {
				// TODO: If the logical cluster object does not exist, ignore the error.
				logger.Error(err, "")
				// TODO: how do we handle failure?
				return
			}

		default:
			// Unknown!
			logger.Info("unknown event")
			// TODO return an error or panic?
		}
	}
}

// ES: return and error and don't panic, move push to outside the case
func NewProvider(ctx context.Context,
	clientset edgeclient.Interface,
	providerName string,
	providerType v1alpha1apis.ClusterProviderType) (ProviderClient, error) {
	var newProvider ProviderClient = nil
	switch providerType {
	case v1alpha1apis.KindProviderType:
		newProvider = kindprovider.New(providerName)
	default:
		err := errors.New("unknown provider type")
		return nil, err
	}
	return newProvider, nil
}

// Provider defines methods to retrieve, list, and watch fleet of clusters.
// The provider is responsible for discovering and managing the lifecycle of each
// cluster.
//
// Example:
// ES: add example, remove list() add Get()
type ProviderClient interface {
	Create(ctx context.Context, name string, opts clusterprovider.Options) (clusterprovider.LogicalClusterInfo, error)
	Delete(ctx context.Context, name string, opts clusterprovider.Options) error

	// List returns a list of logical clusters.
	// This method is used to discover the initial set of logical clusters
	// and to refresh the list of logical clusters periodically.
	List() ([]string, error)

	// Watch returns a Watcher that watches for changes to a list of logical clusters
	// and react to potential changes.
	// TODO: I have yet to implement it for the kind provider type, so commenting it out for now
	Watch() (clusterprovider.Watcher, error)
}
