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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	lcv1alpha1 "github.com/kubestellar/kubestellar/pkg/apis/logicalcluster/v1alpha1"
)

func (c *controller) reconcileClusterProviderDesc(key string) error {
	providerObj, exists, err := c.clusterProviderInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		c.logger.V(2).Info("ClusterProviderDesc deleted", "resource", key)
		if err = c.handleDelete(key); err != nil {
			return err
		}
	} else {
		provider, ok := providerObj.(*lcv1alpha1.ClusterProviderDesc)
		if !ok {
			return errors.New("unexpected object type. expected ClusterProviderDesc")
		}
		c.logger.V(2).Info("reconcile ClusterProviderDesc", "resource", provider.Name)
		if err = c.handleAdd(provider); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) handleAdd(provider *lcv1alpha1.ClusterProviderDesc) error {
	if provider.Status.Phase == lcv1alpha1.ClusterProviderDescPhaseInitializing {
		return nil
	}
	name := provider.Name
	if provider.Status.Phase == lcv1alpha1.ClusterProviderDescPhaseReady {
		// check if we have this provider in cache
		if _, ok := c.providers[name]; ok {
			return nil
		}
	}
	// set provider status to Initializing
	provider, err := c.setProviderStatus(provider, lcv1alpha1.ClusterProviderDescPhaseInitializing)
	if err != nil {
		return err
	}

	_, err = c.CreateProvider(name, provider.Spec.ProviderType)
	if err != nil {
		// TODO: Check if the err is because the provider already exists
		return err
	}

	// create namespace for provider clusters
	nsName := GetNamespace(name)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}
	_, err = c.k8sClientset.CoreV1().Namespaces().Create(c.context, ns, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	err = c.StartDiscovery(name)
	if err != nil {
		return err
	}
	// set provider status to Ready
	_, err = c.setProviderStatus(provider, lcv1alpha1.ClusterProviderDescPhaseReady)
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) handleDelete(providerName string) error {
	err := c.StopDiscovery(providerName)
	if err != nil {
		runtime.HandleError(err)
	}
	// 1. delete unmanaged Logicalclusters.
	// 2. set managed Logicalclusters to NotReady.
	// 3. remove provider namespace if it's empty from Logicalclusters.
	ns := GetNamespace(providerName)
	isNsEmpty := true
	lclusters := c.logicalClusterInformer.GetIndexer().List()
	for _, lcObj := range lclusters {
		lc := lcObj.(*lcv1alpha1.LogicalCluster)
		if lc.Namespace == ns {
			if lc.Spec.Managed {
				//set managed Logicalcluster to NotReady
				lc.Status.Phase = lcv1alpha1.LogicalClusterPhaseNotReady
				_, err := c.clientset.LogicalclusterV1alpha1().LogicalClusters(ns).Update(c.context, lc, metav1.UpdateOptions{})
				if err != nil {
					runtime.HandleError(err)
				}
				isNsEmpty = false
			} else {
				//delete unmanaged Logicalcluster
				err := c.clientset.LogicalclusterV1alpha1().LogicalClusters(ns).Delete(c.context, lc.Name, metav1.DeleteOptions{})
				if err != nil {
					runtime.HandleError(err)
					isNsEmpty = false
				}
			}
		}
	}
	if isNsEmpty {
		err := c.k8sClientset.CoreV1().Namespaces().Delete(c.context, ns, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *controller) setProviderStatus(provider *lcv1alpha1.ClusterProviderDesc, status lcv1alpha1.ClusterProviderDescPhaseType) (*lcv1alpha1.ClusterProviderDesc, error) {
	provider.Status.Phase = status
	updated, err := c.clientset.LogicalclusterV1alpha1().ClusterProviderDescs().Update(c.context, provider, metav1.UpdateOptions{})
	if err != nil {
		return updated, err
	}
	return updated, nil
}
