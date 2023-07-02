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

// processAddOrUpdateLC: process an added or updated LC object
func (c *controller) processAddOrUpdateLC(logicalCluster *lcv1alpha1.LogicalCluster) error {
	if logicalCluster.Status.Phase == lcv1alpha1.LogicalClusterPhaseNotReady && !logicalCluster.Spec.Managed {
		// Discovery noticed that a physical cluster has disappeared.
		// If the cluster is managed, do nothing and let the user handle.
		// If the cluster is unmanaged, then delete the corresponding object.
		return c.clientset.
			LogicalclusterV1alpha1().
			LogicalClusters(GetNamespace(logicalCluster.Spec.ClusterProviderDescName)).
			Delete(c.context, logicalCluster.Name, v1.DeleteOptions{})
	}
	providerInfo, err := c.clientset.LogicalclusterV1alpha1().ClusterProviderDescs().Get(
		c.context, logicalCluster.Spec.ClusterProviderDescName, v1.GetOptions{})
	if err != nil {
		c.logger.Error(err, "failed to get the provider resource")
		return err
	}

	provider, err := c.GetProvider(providerInfo.Name)
	if err != nil {
		c.logger.Error(err, "failed to get provider client")
		return err
	}

	if logicalCluster.Status.Phase == "" && logicalCluster.Spec.Managed {
		// The user created a new logical cluster object and we need to
		// create the corresponding physical cluster.

		// Update status Initializing
		logicalCluster.Status.Phase = lcv1alpha1.LogicalClusterPhaseInitializing
		_, err = c.clientset.
			LogicalclusterV1alpha1().
			LogicalClusters(GetNamespace(logicalCluster.Spec.ClusterProviderDescName)).
			Update(c.context, logicalCluster, v1.UpdateOptions{})
		if err != nil {
			c.logger.Error(err, "failed to update cluster status.")
			return err
		}

		// Async call the provider to create the cluster. Once created, discovery
		// will set the logical cluster in the Ready state.
		go provider.Create(c.context, logicalCluster.Name, pclient.Options{})
		return nil
	}

	// The code from here deals with the case of provider deleted cluster events that might be missed
	_, err = provider.Get(c.context, logicalCluster.Name)

	if err != nil {
		// Cluster is missing on the provider, update the status
		logicalCluster.Status.Phase = lcv1alpha1.LogicalClusterPhaseNotReady
		_, err = c.clientset.LogicalclusterV1alpha1().
			LogicalClusters(GetNamespace(logicalCluster.Spec.ClusterProviderDescName)).
			UpdateStatus(c.context, logicalCluster, v1.UpdateOptions{})
		if err != nil {
			c.logger.Error(err, "failed to update cluster status.")
			return err
		}
	}

	return nil
}

// processDeleteLC: process an LC object deleted event
// If the cluster is managed, then async delete the physical cluster.
// TODO: add a finalizer to the logical cluster object
func (c *controller) processDeleteLC(delCluster *lcv1alpha1.LogicalCluster) error {
	if delCluster.Spec.Managed {
		logger := klog.FromContext(c.context)
		provider, err := c.GetProvider(delCluster.Spec.ClusterProviderDescName)
		if err != nil {
			logger.Error(err, "failed to get provider client")
			return err
		}
		go provider.Delete(c.context, delCluster.Name, pclient.Options{})
	}
	return nil
}
