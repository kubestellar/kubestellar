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
	"k8s.io/client-go/util/flowcontrol"
)

type FlagSet interface {
	Float64Var(p *float64, name string, value float64, usage string)
	IntVar(p *int, name string, value int, usage string)
	StringVar(p *string, name string, value string, usage string)
}

type ClientLimits struct {
	name        string
	description string
	QPS         float64
	Burst       int
}

type ClientOptions struct {
	ClientLimits
	loadingRules *clientcmd.ClientConfigLoadingRules
	overrides    clientcmd.ConfigOverrides
}

func NewClientLimits[FS FlagSet](name, description string) ClientLimits {
	return ClientLimits{
		name:        name,
		description: description,
		QPS:         float64(rest.DefaultQPS),
		Burst:       rest.DefaultBurst,
	}
}

func NewClientOptions[FS FlagSet](name string, description string) *ClientOptions {
	return &ClientOptions{
		ClientLimits: NewClientLimits[FS](name, description),
		loadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
		overrides:    clientcmd.ConfigOverrides{},
	}
}

func (opts *ClientLimits) AddFlags(flags *pflag.FlagSet) {
	flags.Float64Var(&opts.QPS, opts.name+"-qps", opts.QPS, "Max average requests/sec for "+opts.description)
	flags.IntVar(&opts.Burst, opts.name+"-burst", opts.Burst, "Allowed burst in requests/sec for "+opts.description)
}

func (opts *ClientLimits) AddFlagsSansName(flags *pflag.FlagSet) {
	flags.Float64Var(&opts.QPS, "qps", opts.QPS, "Max average requests/sec for "+opts.description)
	flags.IntVar(&opts.Burst, "burst", opts.Burst, "Allowed burst in requests/sec for "+opts.description)
}

func (opts *ClientOptions) AddFlags(flags *pflag.FlagSet) {
	opts.ClientLimits.AddFlags(flags)
	flags.StringVar(&opts.loadingRules.ExplicitPath, opts.name+"-kubeconfig", opts.loadingRules.ExplicitPath, "Path to the kubeconfig file to use for "+opts.description)
	flags.StringVar(&opts.overrides.CurrentContext, opts.name+"-context", opts.overrides.CurrentContext, "The name of the kubeconfig context to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.AuthInfo, opts.name+"-user", opts.overrides.Context.AuthInfo, "The name of the kubeconfig user to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.Cluster, opts.name+"-cluster", opts.overrides.Context.Cluster, "The name of the kubeconfig cluster to use for "+opts.description)

}

func (opts *ClientOptions) AddFlagsSansName(flags *pflag.FlagSet) {
	opts.ClientLimits.AddFlags(flags)
	flags.StringVar(&opts.loadingRules.ExplicitPath, "kubeconfig", opts.loadingRules.ExplicitPath, "Path to the kubeconfig file to use for "+opts.description)
	flags.StringVar(&opts.overrides.CurrentContext, "context", opts.overrides.CurrentContext, "The name of the kubeconfig context to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.AuthInfo, "user", opts.overrides.Context.AuthInfo, "The name of the kubeconfig user to use for "+opts.description)
	flags.StringVar(&opts.overrides.Context.Cluster, "cluster", opts.overrides.Context.Cluster, "The name of the kubeconfig cluster to use for "+opts.description)
}

func (opts *ClientOptions) ToRESTConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(opts.loadingRules, &opts.overrides)
	base, err := clientConfig.ClientConfig()
	if err != nil {
		return base, err
	}
	return opts.ClientLimits.LimitConfig(base), nil
}

func (opts *ClientLimits) LimitConfig(base *rest.Config) *rest.Config {
	ans := *base
	ans.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(float32(opts.QPS), opts.Burst)
	return &ans
}
