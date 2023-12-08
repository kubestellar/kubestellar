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

package msc

import (
	"time"

	rest "k8s.io/client-go/rest"

	spacemsc "github.com/kubestellar/kubestellar/space-framework/pkg/msclientlib"
)

// MultiSpaceClientGen is something that can create client stubs of a given kind
// for any given space.
type MultiSpaceClientGen[ClientInterface any] interface {
	NewForSpace(name, providerNS string) (ClientInterface, error)
}

// MultiSpaceInformerGen is something that can create both client stubs and a shared informer factory.
// IDK why FactoryOption is generated rather than imported.
// Bundling the NewInformerFactoryWithOptions method in here is not really providing any value.
type MultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory any] interface {
	MultiSpaceClientGen[ClientInterface]
	NewInformerFactoryWithOptions(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory
}

// NewMSC wraps a KubestellarSpaceInterface with the ability to make typed clients of a given kind.
func NewMSC[ClientInterface, FactoryOption, InformerFactory any](
	ksi spacemsc.KubestellarSpaceInterface,
	newForConfig func(c *rest.Config) (ClientInterface, error),
	newFactory func(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory,
) MultiSpaceInformerGen[ClientInterface, FactoryOption, InformerFactory] {
	return multiSpaceClientGen[ClientInterface, FactoryOption, InformerFactory]{ksi, newForConfig, newFactory}
}

type multiSpaceClientGen[ClientInterface, FactoryOption, InformerFactory any] struct {
	ksi          spacemsc.KubestellarSpaceInterface
	newForConfig func(c *rest.Config) (ClientInterface, error)
	newFactory   func(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory
}

func (msc multiSpaceClientGen[ClientInterface, FactoryOption, InformerFactory]) NewForSpace(name, providerNS string) (ClientInterface, error) {
	config, err := msc.ksi.ConfigForSpace(name, providerNS)
	if err != nil {
		var zero ClientInterface
		return zero, err
	}
	return msc.newForConfig(config)
}

func (msc multiSpaceClientGen[ClientInterface, FactoryOption, InformerFactory]) NewInformerFactoryWithOptions(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory {
	return msc.newFactory(client, defaultResync, options...)
}
