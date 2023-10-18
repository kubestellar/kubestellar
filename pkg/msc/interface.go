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
type MultiSpaceClientGen[ClientStubs any] interface {
	NewForSpace(name, providerNS string) (ClientStubs, error)
}

// MultiSpaceInformerGen is something that can create both client stubs and a shared informer factory.
// IDK why FactoryOption is generated rather than imported.
// The go type system is stupid about function subtyping and does not allow
// one type parameter to constrain another, so this has to take two independent parameters for the client stubs
// type: one for what NewForConfig returns, and one for what the informer factory consumes.
// Bundling the NewInformerFactoryWithOptions method in here is not really providing any value.
type MultiSpaceInformerGen[ClientStubs, ClientInterface, FactoryOption, InformerFactory any] interface {
	MultiSpaceClientGen[ClientStubs]
	NewInformerFactoryWithOptions(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory
}

// NewMSC wraps a KubestellarSpaceInterface with the ability to make typed clients of a given kind.
func NewMSC[ClientStubs, ClientInterface, FactoryOption, InformerFactory any](
	ksi spacemsc.KubestellarSpaceInterface,
	newForConfig func(c *rest.Config) (ClientStubs, error),
	newFactory func(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory,
) MultiSpaceInformerGen[ClientStubs, ClientInterface, FactoryOption, InformerFactory] {
	return multiSpaceClientGen[ClientStubs, ClientInterface, FactoryOption, InformerFactory]{ksi, newForConfig, newFactory}
}

type multiSpaceClientGen[ClientStubs, ClientInterface, FactoryOption, InformerFactory any] struct {
	ksi          spacemsc.KubestellarSpaceInterface
	newForConfig func(c *rest.Config) (ClientStubs, error)
	newFactory   func(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory
}

func (msc multiSpaceClientGen[ClientStubs, ClientInterface, FactoryOption, InformerFactory]) NewForSpace(name, providerNS string) (ClientStubs, error) {
	config, err := msc.ksi.ConfigForSpace(name, providerNS)
	if err != nil {
		var zero ClientStubs
		return zero, err
	}
	return msc.newForConfig(config)
}

func (msc multiSpaceClientGen[ClientStubs, ClientInterface, FactoryOption, InformerFactory]) NewInformerFactoryWithOptions(client ClientInterface, defaultResync time.Duration, options ...FactoryOption) InformerFactory {
	return msc.newFactory(client, defaultResync, options...)
}
