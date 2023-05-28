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

package client

import (
	"github.com/spf13/pflag"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type ClientOpts struct {
	which        string
	description  string
	loadingRules *clientcmd.ClientConfigLoadingRules
	overrides    clientcmd.ConfigOverrides
}

func NewClientOpts(which, description string) *ClientOpts {
	return &ClientOpts{
		which:        which,
		description:  description,
		loadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
		overrides:    clientcmd.ConfigOverrides{},
	}
}

func (opts *ClientOpts) SetDefaultCurrentContext(currCtx string) *ClientOpts {
	opts.overrides.CurrentContext = currCtx
	return opts
}

func (opts *ClientOpts) SetDefaultUserAndCluster(user, cluster string) *ClientOpts {
	opts.overrides.Context = clientcmdapi.Context{
		AuthInfo: user,
		Cluster:  cluster,
	}
	return opts
}

func (opts *ClientOpts) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.loadingRules.ExplicitPath, opts.which+"-kubeconfig", opts.loadingRules.ExplicitPath, "Path to the kubeconfig file to use for "+opts.description)
	flags.StringVar(&opts.overrides.CurrentContext, opts.which+"-context", opts.overrides.CurrentContext, "The name of the kubeconfig context to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.AuthInfo, opts.which+"-user", opts.overrides.Context.AuthInfo, "The name of the kubeconfig user to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.Cluster, opts.which+"-cluster", opts.overrides.Context.Cluster, "The name of the kubeconfig cluster to use for "+opts.description)

}

func (opts *ClientOpts) ToRESTConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(opts.loadingRules, &opts.overrides)
	return clientConfig.ClientConfig()
}
