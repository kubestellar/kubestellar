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

package spacemanager

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"

	spacev1alpha1 "github.com/kubestellar/kubestellar/space-framework/pkg/apis/space/v1alpha1"
)

func (c *controller) reconcileSpaceProviderDesc(key string) error {
	providerObj, exists, err := c.spaceProviderInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}
	if !exists {
		c.logger.V(2).Info("SpaceProviderDesc deleted", "resource", key)
		if err = c.handleDelete(key); err != nil {
			return err
		}
	} else {
		provider, ok := providerObj.(*spacev1alpha1.SpaceProviderDesc)
		if !ok {
			return errors.New("unexpected object type. expected SpaceProviderDesc")
		}
		c.logger.V(2).Info("reconcile SpaceProviderDesc", "resource", provider.Name)
		if err = c.handleAdd(provider); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) handleAdd(providerDesc *spacev1alpha1.SpaceProviderDesc) error {
	if providerDesc.Status.Phase == spacev1alpha1.SpaceProviderDescPhaseInitializing {
		return nil
	}
	name := providerDesc.Name
	if providerDesc.Status.Phase == spacev1alpha1.SpaceProviderDescPhaseReady {
		// check if we have this provider in cache
		if _, ok := c.providers[name]; ok {
			return nil
		}
	}
	// set provider status to Initializing
	providerDesc, err := c.setProviderStatus(providerDesc, spacev1alpha1.SpaceProviderDescPhaseInitializing)
	if err != nil {
		return err
	}

	provider, err := CreateProvider(c, providerDesc)
	if err != nil {
		// TODO: Check if the err is because the provider already exists
		return err
	}

	// create namespace for provider spaces
	nsName := ProviderNS(name)
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName,
		},
	}
	_, err = c.k8sClientset.CoreV1().Namespaces().Create(c.ctx, ns, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}
	c.logger.Info("Provider namespace created", "provider", name, "namespace", ns.Name)

	err = provider.StartDiscovery()
	if err != nil {
		return err
	}
	// set provider status to Ready
	_, err = c.setProviderStatus(providerDesc, spacev1alpha1.SpaceProviderDescPhaseReady)
	if err != nil {
		return err
	}

	return nil
}

func (c *controller) handleDelete(providerName string) error {
	ns := ProviderNS(providerName)
	provider, ok := c.providers[providerName]
	if ok {
		err := provider.StopDiscovery()
		if err != nil {
			runtime.HandleError(err)
		}
	}
	// 1. delete unmanaged Space.
	// 2. set managed Space to NotReady.
	// 3. remove provider namespace if it's empty from Space.
	isNsEmpty := true
	spaces := c.spaceInformer.GetIndexer().List()
	for _, spaceObj := range spaces {
		space := spaceObj.(*spacev1alpha1.Space)
		if space.Namespace == ns {
			if space.Spec.Managed {
				space.Status.Phase = spacev1alpha1.SpacePhaseNotReady
				_, err := c.clientset.SpaceV1alpha1().Spaces(ns).Update(c.ctx, space, metav1.UpdateOptions{})
				if err != nil {
					runtime.HandleError(err)
				}
				isNsEmpty = false
			} else {
				//delete unmanaged Space
				err := c.clientset.SpaceV1alpha1().Spaces(ns).Delete(c.ctx, space.Name, metav1.DeleteOptions{})
				if err != nil {
					runtime.HandleError(err)
					isNsEmpty = false
				}
			}
		}
	}
	if isNsEmpty {
		delete(c.providers, providerName)
		err := c.k8sClientset.CoreV1().Namespaces().Delete(c.ctx, ns, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func (c *controller) setProviderStatus(provider *spacev1alpha1.SpaceProviderDesc, status spacev1alpha1.SpaceProviderDescPhaseType) (*spacev1alpha1.SpaceProviderDesc, error) {
	provider.Status.Phase = status
	updated, err := c.clientset.SpaceV1alpha1().SpaceProviderDescs().Update(c.ctx, provider, metav1.UpdateOptions{})
	if err != nil {
		return updated, err
	}
	return updated, nil
}
