/*
Copyright 2022 The KubeStellar Authors.

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
	upstreamClient, downstreamClient, err := us.getClients(resource, conversions)
	if err != nil {
		us.logger.Error(err, fmt.Sprintf("failed to get client %q", resourceToString(resource)))
		return err
	}
	resourceForDown := convertToDownstream(resource, conversions)
	us.logger.V(3).Info(fmt.Sprintf("  get %q from downstream", resourceToString(resourceForDown)))
	downstreamResource, err := downstreamClient.Get(resourceForDown)
	isDeleted := false
	if err != nil {
		if k8serrors.IsNotFound(err) {
			us.logger.V(3).Info(fmt.Sprintf("  not found %q in downstream", resourceToString(resourceForDown)))
			us.logger.V(3).Info(fmt.Sprintf("  delete %q from upstream", resourceToString(resourceForDown)))
			isDeleted = true
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
			if !isDeleted {
				// create
				us.logger.V(3).Info(fmt.Sprintf("  create %q in upstream since it's not found", resourceToString(resourceForUp)))
				downstreamResource.SetResourceVersion("")
				downstreamResource.SetUID("")
				setUpsyncAnnotation(downstreamResource)
				applyConversion(downstreamResource, resourceForUp)
				if _, err := upstreamClient.Create(resourceForUp, downstreamResource); err != nil {
					us.logger.Error(err, fmt.Sprintf("failed to create resource to upstream %q", resourceToString(resourceForUp)))
					return err
				}
			} else {
				us.logger.V(3).Info(fmt.Sprintf("  %q has already been deleted from upstream", resourceToString(resourceForUp)))
			}
		} else {
			us.logger.Error(err, fmt.Sprintf("failed to get resource from upstream %q", resourceToString(resourceForUp)))
			return err
		}
	} else {
		if upstreamResource != nil {
			if !isDeleted {
				// update
				us.logger.V(3).Info(fmt.Sprintf("  update %q in upstream since it's found", resourceToString(resourceForUp)))
				if hasUpsyncAnnotation(upstreamResource) {
					downstreamResource.SetResourceVersion(upstreamResource.GetResourceVersion())
					downstreamResource.SetUID(upstreamResource.GetUID())
					setUpsyncAnnotation(downstreamResource)
					applyConversion(downstreamResource, resourceForUp)
					if _, err := upstreamClient.Update(resourceForUp, downstreamResource); err != nil {
						us.logger.Error(err, fmt.Sprintf("failed to update resource on upstream %q", resourceToString(resourceForUp)))
						return err
					}
				} else {
					us.logger.V(2).Info(fmt.Sprintf("  ignore updating %q in upstream since upstream annotation is not set", resourceToString(resourceForUp)))
				}
			} else {
				// Upsyncer should not delete upstream resource objects that are not created by Upsyncer
				if hasUpsyncAnnotation(upstreamResource) {
					us.logger.V(3).Info(fmt.Sprintf("  delete %q from upstream since it's found", resourceToString(resourceForUp)))
					if err := upstreamClient.Delete(resourceForUp, resourceForUp.Name); err != nil {
						us.logger.Error(err, fmt.Sprintf("failed to delete resource from upstream %q", resourceToString(resourceForUp)))
						return err
					}
				} else {
					us.logger.V(2).Info(fmt.Sprintf("  ignore deleting %q from upstream since downsync annotation is not set", resourceToString(resourceForUp)))
				}
			}
		} else {
			msg := fmt.Sprintf("upstream resource is nil even if there is no error %q", resourceToString(resource))
			return errors.New(msg)
		}
	}
	return nil
}

func (us *UpSyncer) UnsyncOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	// It's OK to use same logic as SyncOne unless we execute specific actions for unsynced resources
	return us.SyncOne(resource, conversions)
}

func (us *UpSyncer) SyncMany(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	if resource.Name == "*" && resource.Namespace != "*" {
		return us.syncMany(resource, conversions)
	} else if resource.Name != "*" && resource.Namespace == "*" {
		return us.syncAllNamespaces(resource, conversions, us.SyncOne)
	} else if resource.Name == "*" && resource.Namespace == "*" {
		return us.syncAllNamespaces(resource, conversions, us.SyncMany)
	}
	return us.syncMany(resource, conversions)
}

func (us *UpSyncer) syncMany(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
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
		if k8serrors.IsNotFound(err) {
			downstreamResourceList = &unstructured.UnstructuredList{}
		} else {
			logger.Error(err, "failed to list resource from downstream")
			return err
		}
	}

	logger.V(3).Info("  list resources from upstream")
	resourceForUp := convertToUpstream(resource, conversions)
	upstreamResourceList, err := upstreamClient.List(resourceForUp)
	if err != nil {
		logger.Error(err, "failed to list resource from upstream")
		return err
	}

	logger.V(3).Info("  compute diff between downstream and upstream")
	newResources, updatedResources, deletedResources := diff(logger, downstreamResourceList, upstreamResourceList, setUpsyncAnnotation, hasUpsyncAnnotation)

	logger.V(3).Info("  create resources in upstream")
	for _, resource := range newResources {
		applyConversion(&resource, resourceForUp)
		logger.V(3).Info("  create " + resource.GetName())
		if _, err := upstreamClient.Create(resourceForUp, &resource); err != nil {
			logger.Error(err, "failed to create resource in upstream")
			return err
		}
	}
	logger.V(3).Info("  update resources in upstream")
	for _, resource := range updatedResources {
		applyConversion(&resource, resourceForUp)
		logger.V(3).Info("  update " + resource.GetName())
		if _, err := upstreamClient.Update(resourceForUp, &resource); err != nil {
			logger.Error(err, "failed to update resource in upstream")
			return err
		}
	}
	logger.V(3).Info("  delete resources from upstream")
	for _, resource := range deletedResources {
		applyConversion(&resource, resourceForUp)
		logger.V(3).Info("  delete " + resource.GetName())
		if err := upstreamClient.Delete(resourceForUp, resource.GetName()); err != nil {
			logger.Error(err, "failed to delete resource from upstream")
			return err
		}
	}
	return nil
}

func (us *UpSyncer) UnsyncMany(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	// It's OK to use same logic as SyncMany unless we execute specific actions for unsynced resources
	return us.SyncMany(resource, conversions)
}

func (us *UpSyncer) syncAllNamespaces(
	resource edgev1alpha1.EdgeSyncConfigResource,
	conversions []edgev1alpha1.EdgeSynConversion,
	syncFunc func(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error,
) error {
	namespaces, err := us.getNamespaces()
	if err != nil {
		us.logger.Error(err, fmt.Sprintf("failed to get namespaces %q", resourceToString(resource)))
		return err
	}
	for _, namespace := range namespaces {
		_resource := resource.DeepCopy()
		_resource.Namespace = namespace
		err := syncFunc(*_resource, conversions)
		if err != nil {
			us.logger.Error(err, fmt.Sprintf("failed to upsync %q for namespace %s", resourceToString(resource), namespace))
			return err
		}
	}
	return nil
}

func (us *UpSyncer) getNamespaces() ([]string, error) {
	namespaces := []string{}
	nsResource := edgev1alpha1.EdgeSyncConfigResource{
		Kind: "Namespace", Group: "", Version: "v1",
	}
	_, downstreamClient, err := us.getClients(nsResource, []edgev1alpha1.EdgeSynConversion{})
	if err != nil {
		us.logger.Error(err, "failed to get namespace client")
		return nil, err
	}
	nsUnstList, err := downstreamClient.List(nsResource)
	if err != nil {
		us.logger.Error(err, "failed to get namespaces")
		return nil, err
	}
	for _, nsUnst := range nsUnstList.Items {
		namespaces = append(namespaces, nsUnst.GetName())
	}
	return namespaces, nil
}

func setUpsyncAnnotation(resource *unstructured.Unstructured) {
	setAnnotation(resource, "edge.kcp.io/upsynced", "true")
}

func hasUpsyncAnnotation(resource *unstructured.Unstructured) bool {
	return hasAnnotation(resource, "edge.kcp.io/upsynced")
}

func (us *UpSyncer) BackStatusOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	return nil
}
