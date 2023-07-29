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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	spacev1alpha1 "github.com/kubestellar/kubestellar/pkg/apis/space/v1alpha1"
	pclient "github.com/kubestellar/kubestellar/pkg/space-manager/providerclient"
)

const finalizerName = "SpaceFinalizer"

// containsFinalizer: returns true if the finalizer list contains the space finalizer
func containsFinalizer(space *spacev1alpha1.Space, finalizer string) bool {
	finalizersList := space.ObjectMeta.Finalizers
	for _, f := range finalizersList {
		if f == finalizer {
			return true
		}
	}
	return false
}

func (c *controller) reconcileSpace(key string) error {
	spaceObj, exists, err := c.spaceInformer.GetIndexer().GetByKey(key)
	if err != nil {
		return err
	}

	if !exists {
		c.logger.V(2).Info("Space deleted", "space", key)
		return nil
	}

	space, ok := spaceObj.(*spacev1alpha1.Space)
	if !ok {
		return errors.New("unexpected object type. expected Space")
	}

	if space.ObjectMeta.DeletionTimestamp != nil {
		c.logger.V(2).Info("reconcile Space delete", "resource", space.Name)
		err := c.processDeleteSpace(space)
		if err != nil {
			return err
		}
	} else {
		c.logger.V(2).Info("reconcile Space", "resource", space.Name)
		err := c.processAddOrUpdateSpace(space)
		if err != nil {
			return err
		}
	}
	return nil
}

// processAddOrUpdateSpace: process an added or updated space object
func (c *controller) processAddOrUpdateSpace(space *spacev1alpha1.Space) error {
	if space.Status.Phase == spacev1alpha1.SpacePhaseNotReady && !space.Spec.Managed {
		// Discovery noticed that a physical space has disappeared.
		// If the space is managed, do nothing and let the user handle.
		// If the psace is unmanaged, then delete the corresponding object.
		return c.clientset.
			SpaceV1alpha1().
			Spaces(ProviderNS(space.Spec.SpaceProviderDescName)).
			Delete(c.ctx, space.Name, v1.DeleteOptions{})
	}
	if space.Status.Phase == "" && space.Spec.Managed {
		// The client created a new space object and we need to
		// create the corresponding physical space.
		providerInfo, err := c.clientset.SpaceV1alpha1().SpaceProviderDescs().Get(
			c.ctx, space.Spec.SpaceProviderDescName, v1.GetOptions{})
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
		if !containsFinalizer(space, finalizerName) {
			space.ObjectMeta.Finalizers = append(space.ObjectMeta.Finalizers, finalizerName)
		}
		space.Status.Phase = spacev1alpha1.SpacePhaseInitializing
		_, err = c.clientset.
			SpaceV1alpha1().
			Spaces(ProviderNS(providerInfo.Name)).
			Update(c.ctx, space, v1.UpdateOptions{})
		if err != nil {
			c.logger.Error(err, "failed to update space status.")
			return err
		}

		// Async call the provider to create the space. Once created, discovery
		// will set the space in the Ready state.
		go providerClient.Create(space.Name, pclient.Options{})
		return nil
	}
	// case spacev1alpha1.SpacePhaseInitializing:
	// A managed space is being created by the provider. Need to wait for
	// the space to be created at which point discovery will change the
	// state to READY and update the space config.
	//
	// case spacev1alpha1.SpacePhaseReady:
	// The space is ready - we have nothing to do but celebrate!
	// Likely we got here after the provider finished creating a managed
	// space and dicovery moved it into the Ready state.
	//
	// If a space has been created for an unmanaged physical
	// space, then wait for discovery to move the phase to Ready.
	return nil
}

// processDeleteSpace: process a space object delete event.
// If the space is managed, then async delete the physical space.
func (c *controller) processDeleteSpace(delSpace *spacev1alpha1.Space) error {
	if delSpace.Spec.Managed {
		provider, ok := c.providers[delSpace.Spec.SpaceProviderDescName]
		if !ok {
			return errors.New("failed to get provider client")
		}

		client := provider.providerClient
		if client == nil {
			return errors.New("failed to get provider client")
		}
		go client.Delete(delSpace.Name, pclient.Options{})
	}
	return nil
}
