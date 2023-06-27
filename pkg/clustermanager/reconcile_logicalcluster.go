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
	"errors"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	lcv1alpha1 "github.com/kubestellar/kubestellar/pkg/apis/logicalcluster/v1alpha1"
	pclient "github.com/kubestellar/kubestellar/pkg/clustermanager/providerclient"
)

func (c *controller) reconcileLogicalCluster(key string) error {
	clusterObj, exists, err := c.logicalClusterInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	cluster, ok := clusterObj.(*lcv1alpha1.LogicalCluster)
	if !ok {
		return errors.New("unexpected object type. expected LogicalCluster")
	}

	if !exists {
		c.logger.V(2).Info("LogicalCluster deleted", "resource", cluster.Name)
		err := c.processDeleteLC(cluster)
		if err != nil {
			return err
		}
	} else {
		c.logger.V(2).Info("reconcile LogicalCluster", "resource", cluster.Name)
		err := c.processAddOrUpdateLC(cluster)
		if err != nil {
			return err
		}
	}
	return nil
}

// processAddOrUpdateLC: process an added LC object
func (c *controller) processAddOrUpdateLC(logicalCluster *lcv1alpha1.LogicalCluster) error {
	switch logicalCluster.Status.Phase {
	case lcv1alpha1.LogicalClusterPhaseInitializing:
		// A managed cluster is being created by the provider. Need to wait for
		// the cluster to be created at which point discovery will change the
		// state to READY and update the cluster config.
		return nil
	case lcv1alpha1.LogicalClusterPhaseNotReady:
		// Discovery noticed that an cluster has disappeared
		if logicalCluster.Spec.Managed {
			err := errors.New("a managed physical cluster has been removed")
			return err
		}
		return c.clientset.
			LogicalclusterV1alpha1().
			LogicalClusters(GetNamespace(logicalCluster.Spec.ClusterProviderDescName)).
			Delete(c.context, logicalCluster.Name, v1.DeleteOptions{})
	case lcv1alpha1.LogicalClusterPhaseReady:
		// The cluster is ready - we have nothing to do but celebrate!
		// Likely we got here after the provider finished creating a managed
		// cluster and dicovery moved it into the Ready state.
		return nil
	default:
		if logicalCluster.Status.Phase != "" {
			err := errors.New("unknown status phase")
			return err
		}

		// The client created a new logical cluster object and we need to
		// create a physical cluster.
		return c.createNewLC(logicalCluster)
	}
}

// createNewLC: creates a new managed logical cluster
func (c *controller) createNewLC(newCluster *lcv1alpha1.LogicalCluster) error {
	logger := klog.FromContext(c.context)
	var err error

	providerInfo, err := c.clientset.LogicalclusterV1alpha1().ClusterProviderDescs().Get(
		c.context, newCluster.Spec.ClusterProviderDescName, v1.GetOptions{})
	if err != nil {
		logger.Error(err, "failed to get the provider resource")
		return err
	}

	provider, err := c.GetProvider(providerInfo.Name)
	if err != nil {
		logger.Error(err, "failed to get provider client")
		return err
	}

	// Update status Initializing
	newCluster.Status.Phase = lcv1alpha1.LogicalClusterPhaseInitializing
	newCluster.Spec.Managed = true
	_, err = c.clientset.
		LogicalclusterV1alpha1().
		LogicalClusters(GetNamespace(newCluster.Spec.ClusterProviderDescName)).
		Update(c.context, newCluster, v1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to update cluster status.")
		return err
	}

	// Async call the provider to create the cluster. Once created, discovery
	// will set the logical cluster in the Ready state.
	go provider.Create(c.context, newCluster.Name, pclient.Options{})
	return err
}

// processDeleteLC: process an LC object deleted event
// If the cluster is not managed, then the provider will be asked to delete it.
// TODO: add a finalizer to the logical cluster object if the cluster is managed.
func (c *controller) processDeleteLC(delCluster *lcv1alpha1.LogicalCluster) error {
	logger := klog.FromContext(c.context)
	var err error

	var opts pclient.Options

	provider, err := c.GetProvider(delCluster.Spec.ClusterProviderDescName)
	if err != nil {
		logger.Error(err, "failed to get provider client")
		return err
	}

	err = provider.Delete(c.context, delCluster.Name, opts)
	if err != nil {
		logger.Error(err, "failed to delete cluster")
		return err
	}

	return err
}
