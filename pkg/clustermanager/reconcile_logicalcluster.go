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
		err := c.processAddLC(cluster)
		if err != nil {
			return err
		}
	}
	return nil
}

// TODO: the processAddLC needs to check whether the LogicalCluster has already
// been added and processed, or whether this is actually an Update.
func (c *controller) processAddLC(newCluster *lcv1alpha1.LogicalCluster) error {
	logger := klog.FromContext(c.context)
	var err error

	clusterName := newCluster.Name

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

	// Update status to NotReady
	newCluster.Status.Phase = lcv1alpha1.LogicalClusterPhaseInitializing
	_, err = c.clientset.
		LogicalclusterV1alpha1().
		LogicalClusters(GetNamespace(newCluster.Spec.ClusterProviderDescName)).
		Update(c.context, newCluster, v1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to update cluster status.")
		return err
	}

	// Create cluster
	var opts pclient.Options
	//ES: what exactly is this kubeconfig
	err = provider.Create(c.context, clusterName, opts)
	if err != nil {
		logger.Error(err, "failed to create cluster")
		return err
	}
	logger.Info("Done creating cluster", "clusterName", clusterName)

	// Update the new cluster's status
	newCluster.Status.Phase = lcv1alpha1.LogicalClusterPhaseReady
	_, err = c.clientset.
		LogicalclusterV1alpha1().
		LogicalClusters(GetNamespace(newCluster.Spec.ClusterProviderDescName)).
		Update(c.context, newCluster, v1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to update cluster status.")
		return err
	}

	return err
}

/***
func (c *controller) processUpdateLC(ctx context.Context, key any) error {
	logger := klog.FromContext(ctx)
	var err error
	clusterConfig := key.(*lcv1alpha1.LogicalCluster)
	_, err = c.clientset.
		LogicalclusterV1alpha1().
		LogicalClusters(clusterproviderclient.GetNamespace(clusterConfig.Spec.ClusterProviderDescName)).
		Update(ctx, clusterConfig, v1.UpdateOptions{})
	if err != nil {
		logger.Error(err, "failed to update cluster status.")
		return err
	}
	return err
}
***/

func (c *controller) processDeleteLC(delCluster *lcv1alpha1.LogicalCluster) error {
	logger := klog.FromContext(c.context)
	var err error

	var opts pclient.Options
	clusterName := delCluster.Name

	providerInfo, err := c.clientset.LogicalclusterV1alpha1().ClusterProviderDescs().Get(
		c.context, delCluster.Spec.ClusterProviderDescName, v1.GetOptions{})
	if err != nil {
		logger.Error(err, "failed to get provider resource.")
		return err
	}

	provider, err := c.GetProvider(providerInfo.Name)
	if err != nil {
		logger.Error(err, "failed to get provider client")
		return err
	}

	err = provider.Delete(c.context, clusterName, opts)
	if err != nil {
		logger.Error(err, "failed to delete cluster")
		return err
	}

	return err
}
