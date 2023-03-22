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
	if err := syncer.StartSyncer(ctx, syncerConfig, 1); err != nil {
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
