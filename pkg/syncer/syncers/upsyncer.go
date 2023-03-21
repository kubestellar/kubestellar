package syncers

import (
	"errors"
	"fmt"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type UpSyncer struct {
	logger                  klog.Logger
	upstreamClientFactory   ClientFactory
	downstreamClientFactory ClientFactory
	upstreamClients         map[schema.GroupKind]*Client
	downstreamClients       map[schema.GroupKind]*Client
}

func NewUpSyncer(logger klog.Logger, upstreamClientFactory ClientFactory, downstreamClientFactory ClientFactory, syncedResources []edgev1alpha1.EdgeSyncConfigResource) (*UpSyncer, error) {

	upSyncer := UpSyncer{
		logger:                  logger,
		upstreamClientFactory:   upstreamClientFactory,
		downstreamClientFactory: downstreamClientFactory,
	}
	if err := upSyncer.initializeClients(syncedResources); err != nil {
		logger.Error(err, "failed to initialize clients")
		return &upSyncer, err
	}

	return &upSyncer, nil
}

func (us *UpSyncer) initializeClients(syncedResources []edgev1alpha1.EdgeSyncConfigResource) error {
	us.upstreamClients = map[schema.GroupKind]*Client{}
	us.downstreamClients = map[schema.GroupKind]*Client{}

	return initializeClients(us.logger, syncedResources, us.upstreamClientFactory, us.downstreamClientFactory, us.upstreamClients, us.downstreamClients)
}

func (us *UpSyncer) ReInitializeClients(syncedResources []edgev1alpha1.EdgeSyncConfigResource) error {
	return initializeClients(us.logger, syncedResources, us.upstreamClientFactory, us.downstreamClientFactory, us.upstreamClients, us.downstreamClients)
}

func (us *UpSyncer) getClients(resource edgev1alpha1.EdgeSyncConfigResource) (*Client, *Client, error) {
	return getClients(resource, us.upstreamClients, us.downstreamClients)
}

func (us *UpSyncer) SyncOne(resource edgev1alpha1.EdgeSyncConfigResource) error {
	us.logger.V(3).Info(fmt.Sprintf("upsync '%s'", resourceToString(resource)))

	upstreamClient, k8sClient, err := us.getClients(resource)
	if err != nil {
		us.logger.Error(err, fmt.Sprintf("failed to get client '%s'", resourceToString(resource)))
		return err
	}
	us.logger.V(3).Info(fmt.Sprintf("  get '%s' from downstream", resourceToString(resource)))
	downstreamResource, err := k8sClient.Get(resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			us.logger.V(3).Info(fmt.Sprintf("  not found '%s' in downstream", resourceToString(resource)))
			us.logger.V(3).Info(fmt.Sprintf("  skip upsync '%s'", resourceToString(resource)))
			return nil
		} else {
			us.logger.Error(err, fmt.Sprintf("failed to get resource from upstream '%s'", resourceToString(resource)))
			return err
		}
	}

	us.logger.V(3).Info(fmt.Sprintf("  get '%s' from upstream", resourceToString(resource)))
	upstreamResource, err := upstreamClient.Get(resource)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// create
			us.logger.V(3).Info(fmt.Sprintf("  create '%s' in upstream since it's not found", resourceToString(resource)))
			downstreamResource.SetResourceVersion("")
			downstreamResource.SetUID("")
			if _, err := upstreamClient.Create(resource, downstreamResource); err != nil {
				us.logger.Error(err, fmt.Sprintf("failed to create resource to upstream '%s'", resourceToString(resource)))
				return err
			}
		} else {
			us.logger.Error(err, fmt.Sprintf("failed to get resource from upstream '%s'", resourceToString(resource)))
			return err
		}
	} else {
		if upstreamResource != nil {
			// update
			downstreamResource.SetResourceVersion(upstreamResource.GetResourceVersion())
			downstreamResource.SetUID(upstreamResource.GetUID())
			us.logger.V(3).Info(fmt.Sprintf("  update '%s' in upstream since it's found", resourceToString(resource)))
			if _, err := upstreamClient.Update(resource, downstreamResource); err != nil {
				us.logger.Error(err, fmt.Sprintf("failed to update resource on upstream '%s'", resourceToString(resource)))
				return err
			}
		} else {
			msg := fmt.Sprintf("upstream resource is nil even if there is no error '%s'", resourceToString(resource))
			return errors.New(msg)
		}
	}
	return nil
}

func (us *UpSyncer) BackStatusOne(resource edgev1alpha1.EdgeSyncConfigResource) error {
	return nil
}
