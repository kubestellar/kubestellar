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

package cmd

import (
	"github.com/spf13/pflag"

	clientopts "github.com/kubestellar/kubestellar/options"
)

const (
	defaultConcurrency = 4
)

type TransportOptions struct {
	Concurrency            int
	WdsClientOptions       *clientopts.ClientOptions[*pflag.FlagSet]
	TransportClientOptions *clientopts.ClientOptions[*pflag.FlagSet]
	MaxSizeWrappedObject   int
	WdsName                string
	metricsBindAddr        string
	pprofBindAddr          string
}

func NewTransportOptions() *TransportOptions {
	return &TransportOptions{
		Concurrency:            defaultConcurrency,
		WdsClientOptions:       clientopts.NewClientOptions[*pflag.FlagSet]("wds", "accessing the WDS"),
		TransportClientOptions: clientopts.NewClientOptions[*pflag.FlagSet]("transport", "accessing the ITS"),
		MaxSizeWrappedObject:   500 * 1024,
		metricsBindAddr:        ":8090",
		pprofBindAddr:          ":8092",
	}
}

func (options *TransportOptions) AddFlags(fs *pflag.FlagSet) {
	fs.IntVar(&options.Concurrency, "concurrency", options.Concurrency, "number of concurrent workers to run in parallel")
	options.WdsClientOptions.AddFlags(fs)
	options.TransportClientOptions.AddFlags(fs)
	fs.IntVar(&options.MaxSizeWrappedObject, "max-size-wrapped-object", options.MaxSizeWrappedObject, "Max size of the wrapped object")
	fs.StringVar(&options.WdsName, "wds-name", options.WdsName, "name of the wds to connect to. name should be unique")
	fs.StringVar(&options.metricsBindAddr, "metrics-bind-addr", options.metricsBindAddr, "the [host]:port from which to serve /metrics")
	fs.StringVar(&options.pprofBindAddr, "pprof-bind-addr", options.pprofBindAddr, "the [host]:port from which to serve /debug/pprof")
}
