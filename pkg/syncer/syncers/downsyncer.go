package syncers

import (
	"errors"
	"fmt"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type DownSyncer struct {
	logger                  klog.Logger
	upstreamClientFactory   ClientFactory
	downstreamClientFactory ClientFactory
	upstreamClients         map[schema.GroupKind]*Client
	downstreamClients       map[schema.GroupKind]*Client
}

func NewDownSyncer(logger klog.Logger, upstreamClientFactory ClientFactory, downstreamClientFactory ClientFactory, syncedResources []edgev1alpha1.EdgeSyncConfigResource) (*DownSyncer, error) {

	downSyncer := DownSyncer{
		logger:                  logger,
		upstreamClientFactory:   upstreamClientFactory,
		downstreamClientFactory: downstreamClientFactory,
	}

	if err := downSyncer.initializeClients(syncedResources); err != nil {
		logger.Error(err, "failed to initialize clients")
		return &downSyncer, err
	}

	return &downSyncer, nil
}

func (ds *DownSyncer) initializeClients(syncedResources []edgev1alpha1.EdgeSyncConfigResource) error {
	ds.upstreamClients = map[schema.GroupKind]*Client{}
	ds.downstreamClients = map[schema.GroupKind]*Client{}
	return initializeClients(ds.logger, syncedResources, ds.upstreamClientFactory, ds.downstreamClientFactory, ds.upstreamClients, ds.downstreamClients)
}

func (ds *DownSyncer) ReInitializeClients(syncedResources []edgev1alpha1.EdgeSyncConfigResource) error {
	return initializeClients(ds.logger, syncedResources, ds.upstreamClientFactory, ds.downstreamClientFactory, ds.upstreamClients, ds.downstreamClients)
}

func (ds *DownSyncer) getClients(resource edgev1alpha1.EdgeSyncConfigResource) (*Client, *Client, error) {
	return getClients(resource, ds.upstreamClients, ds.downstreamClients)
}

func (ds *DownSyncer) SyncOne(resource edgev1alpha1.EdgeSyncConfigResource) error {
	ds.logger.V(3).Info(fmt.Sprintf("downsync '%s'", resourceToString(resource)))
	upstreamClient, downstreamClient, err := ds.getClients(resource)
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to get client '%s'", resourceToString(resource)))
		return err
	}
	ds.logger.V(3).Info(fmt.Sprintf("  get '%s' from upstream", resourceToString(resource)))
	upstreamResource, err := upstreamClient.Get(resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			ds.logger.V(3).Info(fmt.Sprintf("  not found '%s' in upstream", resourceToString(resource)))
			ds.logger.V(3).Info(fmt.Sprintf("  skip downsync '%s'", resourceToString(resource)))
			return nil
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from upstream '%s'", resourceToString(resource)))
			return err
		}
	}

	ds.logger.V(3).Info(fmt.Sprintf("  get '%s' from downstream", resourceToString(resource)))
	downstreamResource, err := downstreamClient.Get(resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// create
			ds.logger.V(3).Info(fmt.Sprintf("  create '%s' in upstream since it's not found", resourceToString(resource)))
			upstreamResource.SetResourceVersion("")
			upstreamResource.SetUID("")
			if _, err := downstreamClient.Create(resource, upstreamResource); err != nil {
				ds.logger.Error(err, fmt.Sprintf("failed to create resource to downstream '%s'", resourceToString(resource)))
				return err
			}
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from downstream '%s'", resourceToString(resource)))
			return err
		}
	} else {
		if downstreamResource != nil {
			// update
			ds.logger.V(3).Info(fmt.Sprintf("  update '%s' in upstream since it's found", resourceToString(resource)))
			upstreamResource.SetResourceVersion("")
			upstreamResource.SetUID("")
			upstreamResource.SetManagedFields(nil)
			if _, err := downstreamClient.Update(resource, upstreamResource); err != nil {
				ds.logger.Error(err, fmt.Sprintf("failed to update resource on downstream '%s'", resourceToString(resource)))
				return err
			}
		} else {
			msg := fmt.Sprintf("downstream resource is nil even if there is no error '%s'", resourceToString(resource))
			return errors.New(msg)
		}
	}
	return nil
}

func (ds *DownSyncer) BackStatusOne(resource edgev1alpha1.EdgeSyncConfigResource) error {
	upstreamClient, downstreamClient, err := ds.getClients(resource)
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to get client '%s'", resourceToString(resource)))
		return err
	}
	downstreamResource, err := downstreamClient.Get(resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			ds.logger.V(3).Info(fmt.Sprintf("  not found '%s' in downstream", resourceToString(resource)))
			ds.logger.V(3).Info(fmt.Sprintf("  skip status upsync '%s'", resourceToString(resource)))
			return nil
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from upstream '%s'", resourceToString(resource)))
			return err
		}
	}
	status, found, err := unstructured.NestedMap(downstreamResource.Object, "status")
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to extract status from downstream object '%s'", resourceToString(resource)))
		return err
	} else if !found {
		ds.logger.Info(fmt.Sprintf("no status field downstream object '%s'", resourceToString(resource)))
		return nil
	}
	upstreamResource, err := upstreamClient.Get(resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			ds.logger.V(3).Info(fmt.Sprintf("  not found '%s' in upstream", resourceToString(resource)))
			ds.logger.V(3).Info(fmt.Sprintf("  skip status upsync '%s'", resourceToString(resource)))
			return nil
		} else {
			ds.logger.Error(err, fmt.Sprintf("failed to get resource from upstream '%s'", resourceToString(resource)))
			return err
		}
	}
	_, found, err = unstructured.NestedMap(upstreamResource.Object, "status")
	if err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to extract status from upstream object '%s'", resourceToString(resource)))
		return err
	}
	upstreamResource.Object["status"] = status
	updatedResource := unstructured.Unstructured{
		Object: upstreamResource.Object,
	}
	if _, err := upstreamClient.UpdateStatus(resource, &updatedResource); err != nil {
		ds.logger.Error(err, fmt.Sprintf("failed to update resource on upstream '%s'", resourceToString(resource)))
		return err
	}
	return nil
}
