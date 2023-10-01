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

	"k8s.io/component-base/logs"
	logsapi "k8s.io/component-base/logs/api/v1"

	clientoptions "github.com/kubestellar/kubestellar/pkg/client-options"
)

type Options struct {
	EspwClientOpts clientoptions.ClientOpts
	BaseClientOpts clientoptions.ClientOpts
	Logs           *logs.Options
}

func NewOptions() *Options {
	// Default to -v=2
	logs := logs.NewOptions()
	logs.Verbosity = 2

	return &Options{
		EspwClientOpts: *clientoptions.NewClientOpts("espw", "access to the edge service provider workspace"),
		BaseClientOpts: *clientoptions.NewClientOpts("base", "access to all logical clusters as kcp-admin"),
		Logs:           logs,
	}
}

func (options *Options) AddFlags(fs *pflag.FlagSet) {
	options.EspwClientOpts.AddFlags(fs)
	options.BaseClientOpts.SetDefaultUserAndCluster("kcp-admin", "base")
	options.BaseClientOpts.AddFlags(fs)
	logsapi.AddFlags(options.Logs, fs)
}

func (options *Options) Complete() error {
	return nil
}

func (options *Options) Validate() error {
	return nil
}
