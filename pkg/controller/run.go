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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

func Start(ctx context.Context, processOpts ksopts.ProcessOptions) error {
	logger := klog.FromContext(ctx)
	errChan := make(chan error, 3)
	var servers []*http.Server

	// Start health probe server if address is configured
	if processOpts.HealthProbeBindAddr != "" {
		healthServer := &http.Server{
			Addr:    processOpts.HealthProbeBindAddr,
			Handler: http.HandlerFunc(HappyDumbHandler),
		}
		servers = append(servers, healthServer)

		go func() {
			logger.Info("Starting health probe server", "bindAddress", processOpts.HealthProbeBindAddr)
			if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errChan <- fmt.Errorf("health probe server failed: %w", err)
			}
		}()
	}

	// Start Prometheus metrics server
	metricsServer := &http.Server{
		Addr:    processOpts.MetricsBindAddr,
		Handler: legacyregistry.Handler(),
	}
	servers = append(servers, metricsServer)

	go func() {
		logger.Info("Starting Prometheus metrics server", "bindAddress", processOpts.MetricsBindAddr)
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("Prometheus metrics server failed: %w", err)
		}
	}()

	// Start pprof debug server
	mymux := mux.NewPathRecorderMux("debug")
	routes.Profiling{}.Install(mymux)
	pprofServer := &http.Server{
		Addr:    processOpts.PProfBindAddr,
		Handler: mymux,
	}
	servers = append(servers, pprofServer)

	go func() {
		logger.Info("Starting pprof debug server", "bindAddress", processOpts.PProfBindAddr)
		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("pprof debug server failed: %w", err)
		}
	}()

	// Handle graceful shutdown on context cancellation
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down HTTP servers gracefully")
		for _, server := range servers {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				logger.Error(err, "Error shutting down server", "address", server.Addr)
			}
		}
	}()

	// Wait for any server error or context cancellation
	select {
	case err := <-errChan:
		logger.Error(err, "HTTP server failed, initiating shutdown")
		// Cancel context to trigger graceful shutdown
		return err
	case <-ctx.Done():
		logger.Info("Context cancelled, HTTP servers shutting down")
		return ctx.Err()
	}
}

func HappyDumbHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Write([]byte("ok\r\n"))
}
