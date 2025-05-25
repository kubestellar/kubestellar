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

package controller

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"

	ksopts "github.com/kubestellar/kubestellar/options"
)

func InitialContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 2)
	go func() {
		<-sigChan
		cancel()
		<-sigChan
		os.Exit(2)
	}()
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	return ctx, cancel
}

func Start(ctx context.Context, processOpts ksopts.ProcessOptions) {
	logger := klog.FromContext(ctx)
	if processOpts.HealthProbeBindAddr != "" {
		go func() {
			err := http.ListenAndServe(processOpts.HealthProbeBindAddr, http.HandlerFunc(HappyDumbHandler))
			if err != nil {
				logger.Error(err, "Failed to serve health probes", "bindAddress", processOpts.HealthProbeBindAddr)
				panic(err)
			}
		}()

	}
	go func() {
		err := http.ListenAndServe(processOpts.MetricsBindAddr, legacyregistry.Handler())
		if err != nil {
			logger.Error(err, "Failed to serve Prometheus metrics", "bindAddress", processOpts.MetricsBindAddr)
			panic(err)
		}
	}()

	mymux := mux.NewPathRecorderMux("debug")
	routes.Profiling{}.Install(mymux)
	go func() {
		err := http.ListenAndServe(processOpts.PProfBindAddr, mymux)
		if err != nil {
			logger.Error(err, "Failure in serving /debug/pprof", "bindAddress", processOpts.PProfBindAddr)
			panic(err)
		}
	}()
}

func HappyDumbHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("ok\r\n"))
}
