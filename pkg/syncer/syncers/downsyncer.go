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

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	. "github.com/kubestellar/kubestellar/pkg/syncer/clientfactory"
)

type DownSyncer struct {
	sync.Mutex
	logger                  klog.Logger
	upstreamClientFactory   ClientFactory
	downstreamClientFactory ClientFactory
	upstreamClients         map[schema.GroupKind]*Client
	downstreamClients       map[schema.GroupKind]*Client
}

func NewDownSyncer(logger klog.Logger, upstreamClientFactory ClientFactory, downstreamClientFactory ClientFactory, syncedResources []edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) (*DownSyncer, error) {

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

func (ds *DownSyncer) initializeClients(syncedResources []edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error {
	ds.upstreamClients = map[schema.GroupKind]*Client{}
	ds.downstreamClients = map[schema.GroupKind]*Client{}
	return initializeClients(ds.logger, syncedResources, ds.upstreamClientFactory, ds.downstreamClientFactory, ds.upstreamClients, ds.downstreamClients, conversions)
}

func (ds *DownSyncer) ReInitializeClients(syncedResources []edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error {
	ds.Lock()
	defer ds.Unlock()
	return initializeClients(ds.logger, syncedResources, ds.upstreamClientFactory, ds.downstreamClientFactory, ds.upstreamClients, ds.downstreamClients, conversions)
}

func (ds *DownSyncer) getClients(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) (*Client, *Client, error) {
	ds.Lock()
	defer ds.Unlock()
	return getClients(resource, ds.upstreamClients, ds.downstreamClients, conversions)
}

func (ds *DownSyncer) SyncOne(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error {
	ds.logger.V(3).Info(fmt.Sprintf("sync %q from upstream to downstream", resourceToString(resource)))
	upstreamClient, downstreamClient, err := ds.getClients(resource, conversions)
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to get client %q", resourceToString(resource)))
		return err
	}
	resourceForUp := convertToUpstream(resource, conversions)
	ds.logger.V(3).Info(fmt.Sprintf("  get %q from upstream", resourceToString(resourceForUp)))
	upstreamResource, err := upstreamClient.Get(resourceForUp)
	isDeleted := false
	if err != nil {
		if k8serrors.IsNotFound(err) {
			ds.logger.V(3).Info(fmt.Sprintf("  not found %q in upstream", resourceToString(resourceForUp)))
			ds.logger.V(3).Info(fmt.Sprintf("  delete %q from downstream", resourceToString(resourceForUp)))
			isDeleted = true
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
			if !isDeleted {
				ds.logger.V(3).Info(fmt.Sprintf("  create %q in downstream since it's not found", resourceToString(resourceForDown)))
				upstreamResource.SetResourceVersion("")
				upstreamResource.SetUID("")
				setDownsyncAnnotation(upstreamResource)
				applyConversion(upstreamResource, resourceForDown)
				if _, err := downstreamClient.Create(resourceForDown, upstreamResource); err != nil {
					ds.logger.Error(err, fmt.Sprintf("failed to create resource to downstream %q", resourceToString(resourceForDown)))
					return err
				}
			} else {
				ds.logger.V(3).Info(fmt.Sprintf("  %q has already been deleted from downstream", resourceToString(resourceForDown)))
			}
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from downstream %q", resourceToString(resourceForDown)))
			return err
		}
	} else {
		if downstreamResource != nil {
			if !isDeleted {
				// update
				ds.logger.V(3).Info(fmt.Sprintf("  update %q in downstream since it's found", resourceToString(resourceForDown)))
				if true || hasDownsyncAnnotation(downstreamResource) {
					upstreamResource.SetResourceVersion(downstreamResource.GetResourceVersion())
					upstreamResource.SetUID(downstreamResource.GetUID())
					setDownsyncAnnotation(upstreamResource)
					applyConversion(upstreamResource, resourceForDown)
					_updatedResource, noDiff := ds.computeUpdatedResource(upstreamResource, downstreamResource)
					if !noDiff {
						if _, err := downstreamClient.Update(resourceForDown, _updatedResource); err != nil {
							ds.logger.Error(err, fmt.Sprintf("failed to update resource on downstream %q", resourceToString(resourceForDown)))
							return err
						}
					}
				} else {
					ds.logger.V(2).Info(fmt.Sprintf("  ignore updating %q in downstream since downsync annotation is not set", resourceToString(resourceForDown)))
				}
			} else {
				ds.logger.V(3).Info(fmt.Sprintf("  delete %q from downstream since it's found", resourceToString(resourceForDown)))
				if hasDownsyncAnnotation(downstreamResource) {
					if ds.checkDeletable(downstreamResource) {
						if err := downstreamClient.Delete(resourceForDown, resourceForDown.Name); err != nil {
							ds.logger.Error(err, fmt.Sprintf("failed to delete resource from downstream %q", resourceToString(resourceForDown)))
							return err
						}
					}
				} else {
					ds.logger.V(2).Info(fmt.Sprintf("  ignore deleting %q from downstream since downsync annotation is not setn", resourceToString(resourceForDown)))
				}
			}
		} else {
			msg := fmt.Sprintf("downstream resource is nil even if there is no error %q", resourceToString(resourceForDown))
			return errors.New(msg)
		}
	}
	return nil
}

func (ds *DownSyncer) UnsyncOne(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error {
	// It's OK to use same logic as SyncOne unless we execute specific actions for unsynced resources
	return ds.SyncOne(resource, conversions)
}

func (ds *DownSyncer) BackStatusOne(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error {
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
	if _, err := updateStatusByResource(upstreamClient, resourceForUp, upstreamResource); err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to update resource on upstream %q", resourceToString(resourceForUp)))
		return err
	}
	return nil
}

func (ds *DownSyncer) SyncMany(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error {
	logger := ds.logger.WithName("SyncMany").WithValues("resource", resourceToString(resource))
	logger.V(3).Info("downsync many")
	upstreamClient, downstreamClient, err := ds.getClients(resource, conversions)
	if err != nil {
		logger.Error(err, "failed to get client")
		return err
	}
	resourceForUp := convertToUpstream(resource, conversions)
	logger.V(3).Info("  list resources from upstream")
	upstreamResourceList, err := upstreamClient.List(resourceForUp)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			upstreamResourceList = &unstructured.UnstructuredList{}
		} else {
			logger.Error(err, "failed to list resource from upstream")
			return err
		}
	}
	logger.V(4).Info("  listed objects from upstream", "objects", upstreamResourceList)

	resourceForDown := convertToDownstream(resource, conversions)
	logger.V(3).Info("  list resources from downstream")
	downstreamResourceList, err := downstreamClient.List(resourceForDown)
	if err != nil {
		logger.Error(err, "failed to list resource from downstream")
		return err
	}
	logger.V(4).Info("  listed objects from downstream", "objects", downstreamResourceList)

	logger.V(3).Info("  compute diff between upstream and downstream")
	newResources, updatedResources, deletedResources := diff(logger, upstreamResourceList, downstreamResourceList, setDownsyncAnnotation, hasDownsyncAnnotation)

	logger.V(3).Info("  apply filter such as downsync-overwrite condition to updatedResources and deletedResources")
	updatedResources = ds.computeUpdatedResources(downstreamResourceList, updatedResources)
	deletedResources = ds.computeDeletedResources(downstreamResourceList, deletedResources)

	logger.V(3).Info("  final new/updated/deleted resource list")
	logger.V(3).Info(fmt.Sprintf("    new resource names: %v", mapToNames(newResources)))
	logger.V(3).Info(fmt.Sprintf("    updated resource names: %v", mapToNames(updatedResources)))
	logger.V(3).Info(fmt.Sprintf("    deleted resource names: %v", mapToNames(deletedResources)))

	logger.V(3).Info("  create resources in downstream")
	for _, resource := range newResources {
		applyConversion(&resource, resourceForDown)
		logger.V(3).Info("  create " + resource.GetName())
		if _, err := downstreamClient.Create(resourceForDown, &resource); err != nil {
			logger.Error(err, "failed to create resource to downstream")
			return err
		}
	}
	logger.V(3).Info("  update resources in downstream")
	for _, resource := range updatedResources {
		applyConversion(&resource, resourceForDown)
		logger.V(3).Info("  update " + resource.GetName())
		if _, err := downstreamClient.Update(resourceForDown, &resource); err != nil {
			logger.Error(err, "failed to update resource on downstream")
			return err
		}
	}
	logger.V(3).Info("  delete resources from downstream")
	for _, resource := range deletedResources {
		applyConversion(&resource, resourceForDown)
		logger.V(3).Info("  delete " + resource.GetName())
		if err := downstreamClient.Delete(resourceForDown, resource.GetName()); err != nil {
			logger.Error(err, "failed to delete resource from downstream")
			return err
		}
	}
	return nil
}

// Compute resource object to be pushed from upstreamResource and downstreamResource
//   - updatedResource is the computed resource object to be pushed
//   - noDiff is true if the computed updatedResource is no difference from the downstream resource.
func (ds *DownSyncer) computeUpdatedResource(upstreamResource *unstructured.Unstructured, downstreamResource *unstructured.Unstructured) (updatedResource *unstructured.Unstructured, noDiff bool) {
	if !isDownsyncOverwrite(upstreamResource) {
		ds.logger.V(2).Info(fmt.Sprintf("  downsync-overwrite of %q is marked as false", upstreamResource.GetName()))
		annotations := downstreamResource.GetAnnotations()
		value, ok := annotations[edgev2alpha1.DownsyncOverwriteKey]
		if ok && value == "false" {
			ds.logger.V(2).Info(fmt.Sprintf("  ignore updating %q in downstream", upstreamResource.GetName()))
			return downstreamResource, true
		} else {
			ds.logger.V(2).Info(fmt.Sprintf("  update only annnotation %q in downstream since downsync-overwrite in downstream is still marked as true", upstreamResource.GetName()))
			annotations[edgev2alpha1.DownsyncOverwriteKey] = "false"
			_updatedResource := *downstreamResource
			_updatedResource.SetAnnotations(annotations)
			return &_updatedResource, false
		}
	}
	return upstreamResource, false
}

func (ds *DownSyncer) checkDeletable(downstreamResource *unstructured.Unstructured) bool {
	if isDownsyncOverwrite(downstreamResource) {
		return true
	} else {
		ds.logger.V(2).Info(fmt.Sprintf("  downsync-overwrite of %q is marked as false", downstreamResource.GetName()))
		ds.logger.V(2).Info(fmt.Sprintf("  ignore deleting %q from downstream", downstreamResource.GetName()))
		return false
	}
}

func (ds *DownSyncer) computeUpdatedResources(downstreamResourceList *unstructured.UnstructuredList, updatedResources []unstructured.Unstructured) []unstructured.Unstructured {
	filteredUpdatedResources := []unstructured.Unstructured{}
	for _, updatedResource := range updatedResources {
		downstreamResource, _ := findWithObject(updatedResource, downstreamResourceList)
		_updatedResource, noDiff := ds.computeUpdatedResource(&updatedResource, downstreamResource)
		if !noDiff {
			filteredUpdatedResources = append(filteredUpdatedResources, *_updatedResource)
		}
	}
	return filteredUpdatedResources
}

func (ds *DownSyncer) computeDeletedResources(downstreamResourceList *unstructured.UnstructuredList, deletedResources []unstructured.Unstructured) []unstructured.Unstructured {
	filteredDeletedResources := []unstructured.Unstructured{}
	for _, resource := range deletedResources {
		if ds.checkDeletable(&resource) {
			filteredDeletedResources = append(filteredDeletedResources, resource)
		}
	}
	return filteredDeletedResources
}

func (ds *DownSyncer) UnsyncMany(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error {
	// It's OK to use same logic as SyncMany unless we execute specific actions for unsynced resources
	return ds.SyncMany(resource, conversions)
}

func (ds *DownSyncer) BackStatusMany(resource edgev2alpha1.EdgeSyncConfigResource, conversions []edgev2alpha1.EdgeSynConversion) error {
	logger := ds.logger.WithName("BackStatusMany").WithValues("resource", resourceToString(resource))
	upstreamClient, downstreamClient, err := ds.getClients(resource, conversions)
	if err != nil {
		logger.Error(err, "failed to get client")
		return err
	}

	logger.V(3).Info("  list resources from downstream")
	resourceForDown := convertToDownstream(resource, conversions)
	downstreamResourceList, err := downstreamClient.List(resourceForDown)
	if err != nil {
		logger.Error(err, "failed to list resource from downstream")
		return err
	}

	resourceForUp := convertToUpstream(resource, conversions)
	upstreamResourceList, err := upstreamClient.List(resourceForUp)
	if err != nil {
		logger.Error(err, "failed to list resource from upstream")
		return err
	}

	for _, downstreamResource := range downstreamResourceList.Items {
		status, found, err := unstructured.NestedMap(downstreamResource.Object, "status")
		if err != nil {
			logger.Error(err, fmt.Sprintf("failed to extract status from downstream object: %s. Skip", downstreamResource.GetName()))
			continue
		} else if !found {
			logger.V(3).Info(fmt.Sprintf("  skip status upsync for since no status field in it: %s. Skip", downstreamResource.GetName()))
			continue
		}
		upstreamResource, ok := findWithObject(downstreamResource, upstreamResourceList)
		if ok {
			resourceForUp := convertToUpstream(resource, conversions)
			upstreamResource.Object["status"] = status
			applyConversion(upstreamResource, resourceForUp)
			if _, err := updateStatusByResource(upstreamClient, resourceForUp, upstreamResource); err != nil {
				ds.logger.Error(err, fmt.Sprintf("failed to update resource on upstream %q", resourceToString(resourceForUp)))
				return err
			}
		}
	}
	return nil
}

func updateStatusByResource(upstreamClient *Client, resourceForUp edgev2alpha1.EdgeSyncConfigResource, upstreamResource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if upstreamClient.HasStatusInSubresources() {
		return upstreamClient.UpdateStatus(resourceForUp, upstreamResource)
	}
	return upstreamClient.Update(resourceForUp, upstreamResource)
}

func findWithObject(target unstructured.Unstructured, resourceList *unstructured.UnstructuredList) (*unstructured.Unstructured, bool) {
	for _, resource := range resourceList.Items {
		if target.GetName() == resource.GetName() && target.GetNamespace() == resource.GetNamespace() {
			return &resource, true
		}
	}
	return nil, false
}

const downsyncKey = "edge.kubestellar.io/downsynced"

func setDownsyncAnnotation(resource *unstructured.Unstructured) {
	setAnnotation(resource, downsyncKey, makeOwnedValue(resource))
}

// hasDownsyncAnnotation tests whether the given object has an annotation indicating
// that this object is owned by the syncer.
// The Deployment controller, for example, will copy annotations from a Deployment
// object to owned ReplicaSet objects.
// So it is not enough that there is some annotation with the right key, the value
// has to be tested too.
// But the value test can not be sensitive to something that changes between upstream
// and downstream.
// We assume that only the API group changes between upstream and downstream.
func hasDownsyncAnnotation(resource *unstructured.Unstructured) bool {
	ownedValue := makeOwnedValue(resource)
	return getAnnotation(resource, downsyncKey) == ownedValue
}

func makeOwnedValue(object *unstructured.Unstructured) string {
	gvk := object.GroupVersionKind()
	return gvk.Kind + "/" + object.GetNamespace() + "/" + object.GetName()
}

func isDownsyncOverwrite(resource *unstructured.Unstructured) bool {
	value := getAnnotation(resource, edgev2alpha1.DownsyncOverwriteKey)
	return value != "false"
}
