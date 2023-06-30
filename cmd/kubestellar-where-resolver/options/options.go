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

type Options struct {
	EspwClientOpts   clientoptions.ClientOpts
	RootClientOpts   clientoptions.ClientOpts
	SysAdmClientOpts clientoptions.ClientOpts
	Logs             *logs.Options
}

func NewOptions() *Options {
	// Default to -v=2
	logs := logs.NewOptions()
	logs.Config.Verbosity = config.VerbosityLevel(2)

	return &Options{
		EspwClientOpts:   *clientoptions.NewClientOpts("espw", "access to the edge service provider workspace"),
		RootClientOpts:   *clientoptions.NewClientOpts("root", "access to all clusters"),
		SysAdmClientOpts: *clientoptions.NewClientOpts("sysadm", "access to all clusters as system:admin"),
		Logs:             logs,
	}
}

func (options *Options) AddFlags(fs *pflag.FlagSet) {
	options.EspwClientOpts.AddFlags(fs)
	options.RootClientOpts.SetDefaultUserAndCluster("kcp-admin", "root")
	options.RootClientOpts.AddFlags(fs)
	options.SysAdmClientOpts.SetDefaultCurrentContext("system:admin")
	options.SysAdmClientOpts.AddFlags(fs)
	options.Logs.AddFlags(fs)
}

func (options *Options) Complete() error {
	return nil
}

func (options *Options) Validate() error {
	return nil
}
