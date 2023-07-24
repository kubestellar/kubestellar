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

	lcv1alpha1 "github.com/kubestellar/kubestellar/pkg/apis/logicalcluster/v1alpha1"
	pclient "github.com/kubestellar/kubestellar/pkg/clustermanager/providerclient"
)

const finalizerName = "LCFinalizer"

// containsFinalizer: returns true if the finalizer list contains the logical cluster finalizer
func containsFinalizer(lcluster *lcv1alpha1.LogicalCluster, finalizer string) bool {
	finalizersList := lcluster.ObjectMeta.Finalizers
	for _, f := range finalizersList {
		if f == finalizer {
			return true
		}
	}
	return false
}

func (c *controller) reconcileLogicalCluster(key string) error {
	clusterObj, exists, err := c.logicalClusterInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		c.logger.V(2).Info("LogicalCluster deleted", "cluster", key)
		return nil
	}

	cluster, ok := clusterObj.(*lcv1alpha1.LogicalCluster)
	if !ok {
		return errors.New("unexpected object type. expected LogicalCluster")
	}

	if cluster.ObjectMeta.DeletionTimestamp != nil {
		c.logger.V(2).Info("reconcile LogicalCluster delete", "resource", cluster.Name)
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

// processAddOrUpdateLC: process an added or updated LC object
func (c *controller) processAddOrUpdateLC(logicalCluster *lcv1alpha1.LogicalCluster) error {
	if logicalCluster.Status.Phase == lcv1alpha1.LogicalClusterPhaseNotReady && !logicalCluster.Spec.Managed {
		// Discovery noticed that a physical cluster has disappeared.
		// If the cluster is managed, do nothing and let the user handle.
		// If the cluster is unmanaged, then delete the corresponding object.
		return c.clientset.
			LogicalclusterV1alpha1().
			LogicalClusters(ProviderNS(logicalCluster.Spec.ClusterProviderDescName)).
			Delete(c.ctx, logicalCluster.Name, v1.DeleteOptions{})
	}
	if logicalCluster.Status.Phase == "" && logicalCluster.Spec.Managed {
		// The client created a new logical cluster object and we need to
		// create the corresponding physical cluster.
		providerInfo, err := c.clientset.LogicalclusterV1alpha1().ClusterProviderDescs().Get(
			c.ctx, logicalCluster.Spec.ClusterProviderDescName, v1.GetOptions{})
		if err != nil {
			c.logger.Error(err, "failed to get the provider resource")
			return err
		}
		provider, ok := c.providers[providerInfo.Name]
		if !ok {
			c.logger.Error(err, "failed to get provider from controller")
			return err
		}

		providerClient := provider.providerClient
		if providerClient == nil {
			c.logger.Error(err, "failed to get provider client")
			return err
		}

		// Update status Initializing
		if !containsFinalizer(logicalCluster, finalizerName) {
			logicalCluster.ObjectMeta.Finalizers = append(logicalCluster.ObjectMeta.Finalizers, finalizerName)
		}
		logicalCluster.Status.Phase = lcv1alpha1.LogicalClusterPhaseInitializing
		_, err = c.clientset.
			LogicalclusterV1alpha1().
			LogicalClusters(ProviderNS(providerInfo.Name)).
			Update(c.ctx, logicalCluster, v1.UpdateOptions{})
		if err != nil {
			c.logger.Error(err, "failed to update cluster status.")
			return err
		}

		// Async call the provider to create the cluster. Once created, discovery
		// will set the logical cluster in the Ready state.
		go providerClient.Create(logicalCluster.Name, pclient.Options{})
		return nil
	}
	// case lcv1alpha1.LogicalClusterPhaseInitializing:
	// A managed cluster is being created by the provider. Need to wait for
	// the cluster to be created at which point discovery will change the
	// state to READY and update the cluster config.
	//
	// case lcv1alpha1.LogicalClusterPhaseReady:
	// The cluster is ready - we have nothing to do but celebrate!
	// Likely we got here after the provider finished creating a managed
	// cluster and dicovery moved it into the Ready state.
	//
	// If a logical cluster has been created for an unmanaged physical
	// cluster, then wait for discovery to move the phase to Ready.
	return nil
}

// processDeleteLC: process an LC object deleted event
// If the cluster is managed, then async delete the physical cluster.
func (c *controller) processDeleteLC(delCluster *lcv1alpha1.LogicalCluster) error {
	if delCluster.Spec.Managed {
		provider, ok := c.providers[delCluster.Spec.ClusterProviderDescName]
		if !ok {
			return errors.New("failed to get provider client")
		}

		client := provider.providerClient
		if client == nil {
			return errors.New("failed to get provider client")
		}
		go client.Delete(delCluster.Name, pclient.Options{})
	}
	return nil
}
