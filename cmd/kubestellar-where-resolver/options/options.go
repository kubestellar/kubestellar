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

package options

import (
	"github.com/spf13/pflag"

	"k8s.io/component-base/config"
	"k8s.io/component-base/logs"

	clientoptions "github.com/kubestellar/kubestellar/pkg/client-options"
)

const (
	defaultProviderName string = "default"
	defaultKcsSpaceName string = "espw"
)

type Options struct {
	SpaceMgtOpts   clientoptions.ClientOpts
	EspwClientOpts clientoptions.ClientOpts
	BaseClientOpts clientoptions.ClientOpts
	Logs           *logs.Options
	Provider       string
	KcsName        string
}

func NewOptions() *Options {
	// Default to -v=2
	logs := logs.NewOptions()
	logs.Config.Verbosity = config.VerbosityLevel(2)

	return &Options{
		SpaceMgtOpts:   *clientoptions.NewClientOpts("space-mgt", "access to space management workspace"),
		EspwClientOpts: *clientoptions.NewClientOpts("espw", "access to the edge service provider workspace"),
		BaseClientOpts: *clientoptions.NewClientOpts("base", "access to all logical clusters as kcp-admin"),
		Logs:           logs,
		Provider:       defaultProviderName,
		KcsName:        defaultKcsSpaceName,
	}
}

func (options *Options) AddFlags(fs *pflag.FlagSet) {
	options.SpaceMgtOpts.SetDefaultUserAndCluster("kcp-admin", "root:space-mgt")
	options.SpaceMgtOpts.AddFlags(fs)
	options.EspwClientOpts.AddFlags(fs)
	options.BaseClientOpts.SetDefaultUserAndCluster("kcp-admin", "base")
	options.BaseClientOpts.AddFlags(fs)
	options.Logs.AddFlags(fs)
}

func (options *Options) Complete() error {
	return nil
}

func (options *Options) Validate() error {
	return nil
}
