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

package clientfactory

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
	discoveryClient discovery.DiscoveryInterface
	dyClient        dynamic.Interface
}

func NewClientFactory(logger klog.Logger, dyClient dynamic.Interface, discoveryClient discovery.DiscoveryInterface) (ClientFactory, error) {
	clientFactory := ClientFactory{
		logger:          logger,
		discoveryClient: discoveryClient,
		dyClient:        dyClient,
	}
	return clientFactory, nil
}

func (cf *ClientFactory) GetAPIGroupResources() ([]*restmapper.APIGroupResources, error) {
	return restmapper.GetAPIGroupResources(cf.discoveryClient)
}

func (cf *ClientFactory) GetResourceClient(group string, kind string) (Client, error) {
	var resourceClient Client
	var client dynamic.NamespaceableResourceInterface
	gk := schema.GroupKind{
		Group: group,
		Kind:  kind,
	}
	groupResources, err := cf.GetAPIGroupResources()
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
