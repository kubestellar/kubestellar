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

package options

import (
	"github.com/spf13/pflag"

	clientoptions "github.com/kubestellar/kubestellar/pkg/client-options"
)

const (
	defaultConcurrency              = 4
	defaultKubestellarCoreSpaceName = "espw"
	defaultTransportSpaceName       = "transport"
	defaultSpaceProviderObjectName  = "default"
	defaultExternalAccess           = false
)

type Options struct {
	Concurrency              int
	KubestellarCoreSpaceName string
	TransportSpaceName       string
	SpaceProviderObjectName  string
	ExternalAccess           bool
	SpaceMgtClientOpts       clientoptions.ClientOpts
}

func NewOptions() *Options {
	return &Options{
		Concurrency:              defaultConcurrency,
		KubestellarCoreSpaceName: defaultKubestellarCoreSpaceName,
		TransportSpaceName:       defaultTransportSpaceName,
		SpaceProviderObjectName:  defaultSpaceProviderObjectName,
		ExternalAccess:           defaultExternalAccess,
		SpaceMgtClientOpts:       *clientoptions.NewClientOpts("space-mgt", "access to the space reference space"),
	}
}

func (options *Options) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&options.Concurrency, "concurrency", options.Concurrency, "number of concurrent workers to run in parallel")
	fs.StringVar(&options.KubestellarCoreSpaceName, "core-space", options.KubestellarCoreSpaceName, "the name of the KubeStellar core space")
	fs.StringVar(&options.TransportSpaceName, "transport-space", options.TransportSpaceName, "the name of the transport space")
	fs.StringVar(&options.SpaceProviderObjectName, "space-provider", options.SpaceProviderObjectName, "the name of the KubeStellar space provider")
	fs.BoolVar(&options.ExternalAccess, "external-access", options.ExternalAccess, "the access to the spaces. True when the space-provider is hosted in a space while the controller is running outside of that space")

	options.SpaceMgtClientOpts.AddFlags(fs)
}

func (options *Options) Complete() error {
	return nil
}

func (options *Options) Validate() error {
	return nil
}
