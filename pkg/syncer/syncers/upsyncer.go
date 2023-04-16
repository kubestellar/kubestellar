/*
Copyright 2022 The KCP Authors.

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

package syncers

import (
	"errors"
	"fmt"
	"sync"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	. "github.com/kcp-dev/edge-mc/pkg/syncer/clientfactory"
)

type UpSyncer struct {
	sync.Mutex
	logger                  klog.Logger
	upstreamClientFactory   ClientFactory
	downstreamClientFactory ClientFactory
	upstreamClients         map[schema.GroupKind]*Client
	downstreamClients       map[schema.GroupKind]*Client
}

func NewUpSyncer(logger klog.Logger, upstreamClientFactory ClientFactory, downstreamClientFactory ClientFactory, syncedResources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) (*UpSyncer, error) {

	upSyncer := UpSyncer{
		logger:                  logger,
		upstreamClientFactory:   upstreamClientFactory,
		downstreamClientFactory: downstreamClientFactory,
	}
	if err := upSyncer.initializeClients(syncedResources, conversions); err != nil {
		logger.Error(err, "failed to initialize clients")
		return &upSyncer, err
	}

	return &upSyncer, nil
}

func (us *UpSyncer) initializeClients(syncedResources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	us.upstreamClients = map[schema.GroupKind]*Client{}
	us.downstreamClients = map[schema.GroupKind]*Client{}

	return initializeClients(us.logger, syncedResources, us.upstreamClientFactory, us.downstreamClientFactory, us.upstreamClients, us.downstreamClients, conversions)
}

func (us *UpSyncer) ReInitializeClients(syncedResources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	us.Lock()
	defer us.Unlock()
	return initializeClients(us.logger, syncedResources, us.upstreamClientFactory, us.downstreamClientFactory, us.upstreamClients, us.downstreamClients, conversions)
}

func (us *UpSyncer) getClients(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) (*Client, *Client, error) {
	us.Lock()
	defer us.Unlock()
	return getClients(resource, us.upstreamClients, us.downstreamClients, conversions)
}

func (us *UpSyncer) SyncOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	us.logger.V(3).Info(fmt.Sprintf("upsync %q", resourceToString(resource)))
	upstreamClient, k8sClient, err := us.getClients(resource, conversions)
	if err != nil {
		us.logger.Error(err, fmt.Sprintf("failed to get client %q", resourceToString(resource)))
		return err
	}
	resourceForDown := convertToDownstream(resource, conversions)
	us.logger.V(3).Info(fmt.Sprintf("  get %q from downstream", resourceToString(resourceForDown)))
	downstreamResource, err := k8sClient.Get(resourceForDown)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			us.logger.V(3).Info(fmt.Sprintf("  not found %q in downstream", resourceToString(resourceForDown)))
			us.logger.V(3).Info(fmt.Sprintf("  skip upsync %q", resourceToString(resourceForDown)))
			return nil
		} else {
			us.logger.Error(err, fmt.Sprintf("failed to get resource from upstream %q", resourceToString(resourceForDown)))
			return err
		}
	}

	resourceForUp := convertToUpstream(resource, conversions)
	us.logger.V(3).Info(fmt.Sprintf("  get %q from upstream", resourceToString(resourceForUp)))
	upstreamResource, err := upstreamClient.Get(resourceForUp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// create
			us.logger.V(3).Info(fmt.Sprintf("  create %q in upstream since it's not found", resourceToString(resourceForUp)))
			downstreamResource.SetResourceVersion("")
			downstreamResource.SetUID("")
			applyConversion(downstreamResource, resourceForUp)
			if _, err := upstreamClient.Create(resourceForUp, downstreamResource); err != nil {
				us.logger.Error(err, fmt.Sprintf("failed to create resource to upstream %q", resourceToString(resourceForUp)))
				return err
			}
		} else {
			us.logger.Error(err, fmt.Sprintf("failed to get resource from upstream %q", resourceToString(resourceForUp)))
			return err
		}
	} else {
		if upstreamResource != nil {
			// update
			downstreamResource.SetResourceVersion(upstreamResource.GetResourceVersion())
			downstreamResource.SetUID(upstreamResource.GetUID())
			applyConversion(downstreamResource, resourceForUp)
			us.logger.V(3).Info(fmt.Sprintf("  update %q in upstream since it's found", resourceToString(resourceForUp)))
			if _, err := upstreamClient.Update(resourceForUp, downstreamResource); err != nil {
				us.logger.Error(err, fmt.Sprintf("failed to update resource on upstream %q", resourceToString(resourceForUp)))
				return err
			}
		} else {
			msg := fmt.Sprintf("upstream resource is nil even if there is no error %q", resourceToString(resource))
			return errors.New(msg)
		}
	}
	return nil
}

func (us *UpSyncer) SyncMany(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	logger := us.logger.WithName("SyncMany").WithValues("resource", resourceToString(resource))
	logger.V(3).Info("upsync many")

	upstreamClient, downstreamClient, err := us.getClients(resource, conversions)
	if err != nil {
		us.logger.Error(err, fmt.Sprintf("failed to get client %q", resourceToString(resource)))
		return err
	}

	logger.V(3).Info("  list resources from downstream")
	resourceForDown := convertToDownstream(resource, conversions)
	downstreamResourceList, err := downstreamClient.List(resourceForDown)
	if err != nil {
		logger.Error(err, "failed to list resource from downstream")
		return err
	}

	logger.V(3).Info("  list resources from upstream")
	resourceForUp := convertToUpstream(resource, conversions)
	upstreamResourceList, err := upstreamClient.List(resourceForUp)
	if err != nil {
		logger.Error(err, "failed to list resource from upstream")
		return err
	}

	logger.V(3).Info("  compute diff between downstream and upstream")
	newResources, updatedResources, deletedResources := diff(logger, downstreamResourceList, upstreamResourceList)

	logger.V(3).Info("  create resources in upstream")
	for _, resource := range newResources {
		resource.SetResourceVersion("")
		resource.SetUID("")
		applyConversion(&resource, resourceForUp)
		logger.V(3).Info("  create " + resource.GetName())
		if _, err := upstreamClient.Create(resourceForUp, &resource); err != nil {
			logger.Error(err, "failed to create resource in upstream")
			return err
		}
	}
	logger.V(3).Info("  update resources in upstream")
	for _, resource := range updatedResources {
		resource.SetResourceVersion("")
		resource.SetUID("")
		resource.SetManagedFields(nil)
		applyConversion(&resource, resourceForUp)
		logger.V(3).Info("  update " + resource.GetName())
		if _, err := upstreamClient.Update(resourceForUp, &resource); err != nil {
			logger.Error(err, "failed to create resource in upstream")
			return err
		}
	}
	// nothing to do for deletion for now
	_ = deletedResources
	return nil
}

func (us *UpSyncer) BackStatusOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	return nil
}
