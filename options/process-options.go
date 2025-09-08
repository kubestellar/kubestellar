/*
Copyright 2024 The KubeStellar Authors.

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
)

type ProcessOptions struct {
	MetricsBindAddr     string
	PProfBindAddr       string
	HealthProbeBindAddr string
}

func (po *ProcessOptions) AddToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&po.MetricsBindAddr, "metrics-bind-address", po.MetricsBindAddr, "the [host]:port from which to serve /metrics")
	flags.StringVar(&po.PProfBindAddr, "pprof-bind-address", po.PProfBindAddr, "the [host]:port from which to serve /debug/pprof")
	flags.StringVar(&po.HealthProbeBindAddr, "health-probe-bind-address", po.HealthProbeBindAddr, "the [host]:port from which to serve /healthz,/readyz (empty string to not serve)")
}