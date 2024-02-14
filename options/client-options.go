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

package clientsopts

import (
	"github.com/spf13/pflag"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type ClientOptions struct {
	name         string
	description  string
	loadingRules *clientcmd.ClientConfigLoadingRules
	overrides    clientcmd.ConfigOverrides
}

func NewClientOptions(name string, description string) *ClientOptions {
	return &ClientOptions{
		name:         name,
		description:  description,
		loadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
		overrides:    clientcmd.ConfigOverrides{},
	}
}

func (opts *ClientOptions) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.loadingRules.ExplicitPath, opts.name+"-kubeconfig", opts.loadingRules.ExplicitPath, "Path to the kubeconfig file to use for "+opts.description)
	flags.StringVar(&opts.overrides.CurrentContext, opts.name+"-context", opts.overrides.CurrentContext, "The name of the kubeconfig context to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.AuthInfo, opts.name+"-user", opts.overrides.Context.AuthInfo, "The name of the kubeconfig user to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.Cluster, opts.name+"-cluster", opts.overrides.Context.Cluster, "The name of the kubeconfig cluster to use for "+opts.description)

}

func (opts *ClientOptions) ToRESTConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(opts.loadingRules, &opts.overrides)
	return clientConfig.ClientConfig()
}
