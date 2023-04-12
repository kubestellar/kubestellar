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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	. "github.com/kcp-dev/edge-mc/pkg/syncer/clientfactory"
)

type DownSyncer struct {
	sync.Mutex
	logger                  klog.Logger
	upstreamClientFactory   ClientFactory
	downstreamClientFactory ClientFactory
	upstreamClients         map[schema.GroupKind]*Client
	downstreamClients       map[schema.GroupKind]*Client
}

func NewDownSyncer(logger klog.Logger, upstreamClientFactory ClientFactory, downstreamClientFactory ClientFactory, syncedResources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) (*DownSyncer, error) {

	downSyncer := DownSyncer{
		logger:                  logger,
		upstreamClientFactory:   upstreamClientFactory,
		downstreamClientFactory: downstreamClientFactory,
	}

	if err := downSyncer.initializeClients(syncedResources, conversions); err != nil {
		logger.Error(err, "failed to initialize clients")
		return &downSyncer, err
	}

	return &downSyncer, nil
}

func (ds *DownSyncer) initializeClients(syncedResources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	ds.upstreamClients = map[schema.GroupKind]*Client{}
	ds.downstreamClients = map[schema.GroupKind]*Client{}
	return initializeClients(ds.logger, syncedResources, ds.upstreamClientFactory, ds.downstreamClientFactory, ds.upstreamClients, ds.downstreamClients, conversions)
}

func (ds *DownSyncer) ReInitializeClients(syncedResources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	ds.Lock()
	defer ds.Unlock()
	return initializeClients(ds.logger, syncedResources, ds.upstreamClientFactory, ds.downstreamClientFactory, ds.upstreamClients, ds.downstreamClients, conversions)
}

func (ds *DownSyncer) getClients(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) (*Client, *Client, error) {
	ds.Lock()
	defer ds.Unlock()
	return getClients(resource, ds.upstreamClients, ds.downstreamClients, conversions)
}

func (ds *DownSyncer) SyncOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	ds.logger.V(3).Info(fmt.Sprintf("downsync %q", resourceToString(resource)))
	upstreamClient, downstreamClient, err := ds.getClients(resource, conversions)
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to get client %q", resourceToString(resource)))
		return err
	}
	resourceForUp := convertToUpstream(resource, conversions)
	ds.logger.V(3).Info(fmt.Sprintf("  get %q from upstream", resourceToString(resourceForUp)))
	upstreamResource, err := upstreamClient.Get(resourceForUp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			ds.logger.V(3).Info(fmt.Sprintf("  not found %q in upstream", resourceToString(resourceForUp)))
			ds.logger.V(3).Info(fmt.Sprintf("  skip downsync %q", resourceToString(resourceForUp)))
			return nil
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from upstream %q", resourceToString(resourceForUp)))
			return err
		}
	}

	resourceForDown := convertToDownstream(resource, conversions)
	ds.logger.V(3).Info(fmt.Sprintf("  get %q from downstream", resourceToString(resourceForDown)))
	downstreamResource, err := downstreamClient.Get(resourceForDown)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// create
			ds.logger.V(3).Info(fmt.Sprintf("  create %q in downstream since it's not found", resourceToString(resourceForDown)))
			upstreamResource.SetResourceVersion("")
			upstreamResource.SetUID("")
			applyConversion(upstreamResource, resourceForDown)
			if _, err := downstreamClient.Create(resourceForDown, upstreamResource); err != nil {
				ds.logger.Error(err, fmt.Sprintf("failed to create resource to downstream %q", resourceToString(resourceForDown)))
				return err
			}
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from downstream %q", resourceToString(resourceForDown)))
			return err
		}
	} else {
		if downstreamResource != nil {
			// update
			ds.logger.V(3).Info(fmt.Sprintf("  update %q in downstream since it's found", resourceToString(resourceForDown)))
			upstreamResource.SetResourceVersion("")
			upstreamResource.SetUID("")
			upstreamResource.SetManagedFields(nil)
			applyConversion(upstreamResource, resourceForDown)
			if _, err := downstreamClient.Update(resourceForDown, upstreamResource); err != nil {
				ds.logger.Error(err, fmt.Sprintf("failed to update resource on downstream %q", resourceToString(resourceForDown)))
				return err
			}
		} else {
			msg := fmt.Sprintf("downstream resource is nil even if there is no error %q", resourceToString(resourceForDown))
			return errors.New(msg)
		}
	}
	return nil
}

func (ds *DownSyncer) BackStatusOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	upstreamClient, downstreamClient, err := ds.getClients(resource, conversions)
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to get client %q", resourceToString(resource)))
		return err
	}
	resourceForDown := convertToDownstream(resource, conversions)
	downstreamResource, err := downstreamClient.Get(resourceForDown)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			ds.logger.V(3).Info(fmt.Sprintf("  not found %q in downstream", resourceToString(resourceForDown)))
			ds.logger.V(3).Info(fmt.Sprintf("  skip status upsync %q", resourceToString(resourceForDown)))
			return nil
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from downstream %q", resourceToString(resourceForDown)))
			return err
		}
	}
	status, found, err := unstructured.NestedMap(downstreamResource.Object, "status")
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to extract status from downstream object %q", resourceToString(resourceForDown)))
		return err
	} else if !found {
		ds.logger.V(3).Info(fmt.Sprintf("  skip status upsync %q since no status field in it", resourceToString(resourceForDown)))
		return nil
	}
	resourceForUp := convertToUpstream(resource, conversions)
	upstreamResource, err := upstreamClient.Get(resourceForUp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			ds.logger.V(3).Info(fmt.Sprintf("  not found %q in upstream", resourceToString(resourceForUp)))
			ds.logger.V(3).Info(fmt.Sprintf("  skip status upsync %q", resourceToString(resourceForUp)))
			return nil
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from upstream %q", resourceToString(resourceForUp)))
			return err
		}
	}
	_, found, err = unstructured.NestedMap(upstreamResource.Object, "status")
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to extract status from upstream object %q", resourceToString(resourceForUp)))
		return err
	}
	upstreamResource.Object["status"] = status
	applyConversion(upstreamResource, resourceForUp)
	if _, err := upstreamClient.UpdateStatus(resourceForUp, upstreamResource); err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to update resource on upstream %q", resourceToString(resourceForUp)))
		return err
	}
	return nil
}
