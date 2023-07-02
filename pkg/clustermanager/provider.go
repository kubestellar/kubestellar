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

	"github.com/go-logr/logr"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	kindprovider "github.com/kubestellar/kubestellar/clusterprovider/kind"
	lcv1alpha1apis "github.com/kubestellar/kubestellar/pkg/apis/logicalcluster/v1alpha1"
	clusterprovider "github.com/kubestellar/kubestellar/pkg/clustermanager/providerclient"
)

// Each provider gets its own namespace named prefixNamespace+providerName
const prefixNamespace = "lcprovider-"

type provider struct {
	name            string
	providerClient  clusterprovider.ProviderClient
	c               *controller
	providerWatcher clusterprovider.Watcher
	nameSpace       string
}

// TODO: this is termporary for stage 1. For stage 2 we expect to have a uniform interface for all informers.
func newProviderClient(ctx context.Context, providerName string,
	providerType lcv1alpha1apis.ClusterProviderType) clusterprovider.ProviderClient {
	var pClient clusterprovider.ProviderClient = nil
	switch providerType {
	case lcv1alpha1apis.KindProviderType:
		pClient = kindprovider.New(providerName)
	default:
		return nil
	}
	return pClient
}

// CreateProvider returns new provider client
func CreateProvider(c *controller, providerName string,
	providerType lcv1alpha1apis.ClusterProviderType) (*provider, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, exists := c.providers[providerName]
	if exists {
		err := errors.New("provider already in the list")
		return nil, err
	}

	newProviderClient := newProviderClient(c.context, providerName, providerType)
	if newProviderClient == nil {
		return nil, errors.New("unknown provider type")
	}

	p := &provider{
		name:           providerName,
		providerClient: newProviderClient,
		c:              c,
		nameSpace:      prefixNamespace + providerName,
	}

	c.providers[providerName] = p
	return p, nil
}

// StartDiscovery will start watching provider clusters for changes
func (p *provider) StartDiscovery() error {
	watcher, err := p.providerClient.Watch()
	if err != nil {
		return err
	}
	p.providerWatcher = watcher
	go p.processProviderWatchEvents()

	return nil
}

// StopDiscovery will stop watching provider clusters for changes
func (p *provider) StopDiscovery() error {
	p.c.lock.Lock()
	defer p.c.lock.Unlock()

	if p.providerWatcher == nil {
		return errors.New("failed to stop provider discovery. provider watcher does not exist")
	}
	p.providerWatcher.Stop()
	return nil
}

func (p *provider) processProviderWatchEvents() {
	logger := p.c.logger
	ctx := p.c.context
	for {
		event, ok := <-p.providerWatcher.ResultChan()
		if !ok {
			logger.Info("Cluster provider watch was stopped", "provider", p.name)
			return
		}
		lcName := event.Name
		reflcluster, err := p.c.clientset.LogicalclusterV1alpha1().LogicalClusters(p.nameSpace).Get(ctx, lcName, v1.GetOptions{})
		found := reflcluster != nil && err == nil

		switch event.Type {
		case watch.Added:
			logger.Info("New cluster was detected", "cluster", event.Name)
			// A new cluster was detected either create it or change the status to READY
			if !found {
				// No corresponding Logicalcluster, let's create it
				logger.Info("Creating new LogicalCluster object", "cluster", event.Name)
				lcluster := lcv1alpha1apis.LogicalCluster{}
				lcluster.Name = lcName
				lcluster.Spec.ClusterProviderDescName = p.name
				lcluster.Spec.Managed = false
				lcluster.Status.ClusterConfig = event.LCInfo.Config
				lcluster.Status.Phase = lcv1alpha1apis.LogicalClusterPhaseReady
				_, err = p.c.clientset.LogicalclusterV1alpha1().LogicalClusters(p.nameSpace).Create(ctx, &lcluster, v1.CreateOptions{})
				chkErrAndReturn(logger, err, "Detected New cluster. Couldn't create the corresponding LogicalCluster", "cluster name", lcName)
			} else {
				// TODO: when finalizers added - cheeck the logicalcluster delete timestamp
				// Logical cluster exists , just update its status
				reflcluster.Status.Phase = lcv1alpha1apis.LogicalClusterPhaseReady
				// TODO: Should we really update the config ?
				reflcluster.Status.ClusterConfig = event.LCInfo.Config
				_, err = p.c.clientset.LogicalclusterV1alpha1().LogicalClusters(p.nameSpace).Update(ctx, reflcluster, v1.UpdateOptions{})
				chkErrAndReturn(logger, err, "Detected New cluster. Couldn't update the corresponding LogicalCluster status", "cluster name", lcName)
			}

		case watch.Deleted:
			logger.Info("A cluster was removed", "cluster", event.Name)
			if !found {
				// There is no LC object so there is nothing we should do
				return
			}
			if !reflcluster.DeletionTimestamp.IsZero() {
				//TODO: When using finalizers check if LC was deleted and if so remove the finalizer.
				return
			}
			reflcluster.Status.Phase = lcv1alpha1apis.LogicalClusterPhaseNotReady
			_, err = p.c.clientset.LogicalclusterV1alpha1().LogicalClusters(p.nameSpace).Update(ctx, reflcluster, v1.UpdateOptions{})
			chkErrAndReturn(logger, err, "Cluster was removed, Couldn't update the LogicalCluster status")

		default:
			logger.Info("unknown event type", "type", event.Type)
		}
	}
}

func chkErrAndReturn(logger logr.Logger, err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		logger.Error(err, msg)
	}
}
