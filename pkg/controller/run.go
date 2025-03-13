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
	"time"

	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/component-base/metrics/legacyregistry"
	"k8s.io/klog/v2"
	"github.com/prometheus/client_golang/prometheus"

	ksopts "github.com/kubestellar/kubestellar/options"
)

// Define Prometheus metrics for API server requests/responses
var (
	apiServerRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "generic_transport_apiserver_requests_total",
			Help: "Total number of API server requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	apiServerRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "generic_transport_apiserver_request_duration_seconds",
			Help:    "Duration of API server requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	// Register Prometheus metrics
	prometheus.MustRegister(apiServerRequestsTotal)
	prometheus.MustRegister(apiServerRequestDuration)
}

// InstrumentedTransport wraps the HTTP client to track metrics
type instrumentedTransport struct {
	transport http.RoundTripper
}

func (t *instrumentedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Make the API server request
	resp, err := t.transport.RoundTrip(req)

	// Record metrics
	duration := time.Since(start).Seconds()
	apiServerRequestDuration.WithLabelValues(req.Method, req.URL.Path).Observe(duration)
	if err != nil {
		apiServerRequestsTotal.WithLabelValues(req.Method, req.URL.Path, "error").Inc()
	} else {
		apiServerRequestsTotal.WithLabelValues(req.Method, req.URL.Path, resp.Status).Inc()
	}

	return resp, err
}

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

	// Replace the default HTTP client with the instrumented one
	http.DefaultTransport = &instrumentedTransport{transport: http.DefaultTransport}

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
