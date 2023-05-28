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

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/pflag"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"

	synceroptions "github.com/kcp-dev/edge-mc/cmd/syncer/options"
	"github.com/kcp-dev/edge-mc/pkg/syncer"
)

func main() {
	options := synceroptions.NewOptions()
	fs := pflag.NewFlagSet("syncer", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	options.AddFlags(fs)
	fs.Parse(os.Args[1:])
	if err := options.Complete(); err != nil {
		panic(err)
	}
	if err := options.Validate(); err != nil {
		panic(err)
	}

	kcpConfigOverrides := &clientcmd.ConfigOverrides{
		CurrentContext: options.FromContext,
	}
	upstreamConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: options.FromKubeconfig},
		kcpConfigOverrides).ClientConfig()
	if err != nil {
		panic(err)
	}

	upstreamConfig.QPS = options.QPS
	upstreamConfig.Burst = options.Burst

	downstreamConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: options.ToKubeconfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: options.ToContext,
		}).ClientConfig()
	if err != nil {
		panic(err)
	}

	downstreamConfig.QPS = options.QPS
	downstreamConfig.Burst = options.Burst

	syncerConfig := &syncer.SyncerConfig{
		UpstreamConfig:   upstreamConfig,
		DownstreamConfig: downstreamConfig,
		SyncTargetPath:   logicalcluster.NewPath(options.FromClusterPath),
		SyncTargetName:   options.SyncTargetName,
		SyncTargetUID:    options.SyncTargetUID,
	}

	ctx := setupSignalContext()
	if err := syncer.RunSyncer(ctx, syncerConfig, 1); err != nil {
		panic(err)
	}

	<-ctx.Done()
}

var onlyOneSignalHandler = make(chan struct{})
var shutdownHandler chan os.Signal
var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

func setupSignalContext() context.Context {
	close(onlyOneSignalHandler) // panics when called twice

	shutdownHandler = make(chan os.Signal, 2)

	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(shutdownHandler, shutdownSignals...)
	go func() {
		<-shutdownHandler
		cancel()
		<-shutdownHandler
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}
