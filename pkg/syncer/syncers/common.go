package syncers

import (
	"errors"
	"fmt"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

func resourceToString(resource edgev1alpha1.EdgeSyncConfigResource) string {
	return fmt.Sprintf("%s.%s/%s in %s", resource.Kind, resource.Group, resource.Name, resource.Namespace)
}

func initializeClients(
	logger klog.Logger,
	syncResources []edgev1alpha1.EdgeSyncConfigResource,
	upstreamClientFactory ClientFactory,
	downstreamClientFactory ClientFactory,
	upstreamClients map[schema.GroupKind]*Client,
	downstreamClients map[schema.GroupKind]*Client,
) error {
	logger.V(3).Info("initialize clients")
	for _, syncResource := range syncResources {
		logger.V(3).Info(fmt.Sprintf("  setup ResourceClient for '%s'", resourceToString(syncResource)))

		group := syncResource.Group
		kind := syncResource.Kind
		gk := schema.GroupKind{
			Group: group,
			Kind:  kind,
		}
		_, ok := upstreamClients[gk]
		if ok {
			logger.V(3).Info(fmt.Sprintf("  skip since upstreamClientFactory is already setup for '%s'", resourceToString(syncResource)))
		} else {
			logger.V(3).Info(fmt.Sprintf("  create upstreamClientFactory for '%s'", resourceToString(syncResource)))
			upstreamClient, err := upstreamClientFactory.GetResourceClient(group, kind)
			if err != nil {
				logger.Error(err, fmt.Sprintf("failed to create kcpResourceClient '%s.%s'", group, kind))
				return err
			}
			upstreamClients[gk] = &upstreamClient
		}

		_, ok = downstreamClients[gk]
		if ok {
			logger.V(3).Info(fmt.Sprintf("  skip since downstreamClientFactory is already setup for '%s'", resourceToString(syncResource)))
		} else {
			logger.V(3).Info(fmt.Sprintf("  create downstreamClientFactory for '%s'", resourceToString(syncResource)))
			k8sClient, err := downstreamClientFactory.GetResourceClient(group, kind)
			if err != nil {
				logger.Error(err, fmt.Sprintf("failed to create k8sResourceClient '%s.%s'", group, kind))
				return err
			}
			downstreamClients[gk] = &k8sClient
		}
	}
	return nil
}

func getClients(resource edgev1alpha1.EdgeSyncConfigResource, upstreamClients map[schema.GroupKind]*Client, downstreamClients map[schema.GroupKind]*Client) (*Client, *Client, error) {
	gk := schema.GroupKind{
		Group: resource.Group,
		Kind:  resource.Kind,
	}
	upstreamClient, ok := upstreamClients[gk]
	if !ok {
		msg := fmt.Sprintf("upstreamClient for '%s.%s' is not registered", resource.Group, resource.Kind)
		return nil, nil, errors.New(msg)
	}
	downstreamClient, ok := downstreamClients[gk]
	if !ok {
		msg := fmt.Sprintf("downstreamClient for '%s.%s' is not registered", resource.Group, resource.Kind)
		return nil, nil, errors.New(msg)
	}
	return upstreamClient, downstreamClient, nil
}
