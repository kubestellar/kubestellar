package syncers

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
)

var BURST = 1024

type ClientFactory struct {
	logger          klog.Logger
	discoveryClient *discovery.DiscoveryClient
	dyClient        dynamic.Interface
}

func NewClientFactory(logger klog.Logger, dyClient dynamic.Interface, discoveryClient *discovery.DiscoveryClient) (ClientFactory, error) {
	clientFactory := ClientFactory{
		logger:          logger,
		discoveryClient: discoveryClient,
		dyClient:        dyClient,
	}
	return clientFactory, nil
}

func (cf *ClientFactory) GetResourceClient(group string, kind string) (Client, error) {
	var resourceClient Client
	var client dynamic.NamespaceableResourceInterface
	gk := schema.GroupKind{
		Group: group,
		Kind:  kind,
	}
	groupResources, err := restmapper.GetAPIGroupResources(cf.discoveryClient)
	if err != nil {
		cf.logger.Error(err, "failed to get APIGroupResource")
		return resourceClient, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	mappings, err := restMapper.RESTMappings(gk)
	if err != nil {
		cf.logger.Error(err, fmt.Sprintf("failed to get restMapping %s", gk.String()))
		return resourceClient, err
	}
	if len(mappings) == 0 {
		cf.logger.Error(err, fmt.Sprintf("no restMapping %s", gk.String()))
		return resourceClient, err
	}
	mapping := mappings[0]
	client = cf.dyClient.Resource(mapping.Resource)
	resourceClient = Client{
		ResourceClient: client,
		scope:          mapping.Scope,
	}
	return resourceClient, nil
}
