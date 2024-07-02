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
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/spf13/pflag"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"
	clusterinformers "open-cluster-management.io/api/client/cluster/informers/externalversions"

	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/client-go/dynamic"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version"
	"k8s.io/klog/v2"

	ksclientset "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	ksinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions"
	ksmetrics "github.com/kubestellar/kubestellar/pkg/metrics"
	"github.com/kubestellar/kubestellar/pkg/transport"
)

// The following code is responsible for running a transport controller with a given
// transport plugin and contains the base or generic functionality.
// The GenericMain function gets the transport-specific implementation and uses it to initialize
// the generic transport controller which is responsible for processing the
// Binding added/updated/deleted events.
// In order to use the GenericMain function, one has to call it in the following format:
// GenericMain(YourTransportSpecificImplementation())

// Example for this can be seen here:
// https://github.com/kubestellar/ocm-transport-plugin/blob/main/cmd/main.go

const (
	defaultResyncPeriod = time.Duration(0)
)

func GenericMain(transportImplementation transport.Transport) {
	logger := klog.Background().WithName(transport.ControllerName)
	ctx := klog.NewContext(context.Background(), logger)

	options := NewTransportOptions()
	fs := pflag.NewFlagSet(transport.ControllerName, pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	options.AddFlags(fs)
	fs.Parse(os.Args[1:])

	fs.VisitAll(func(flg *pflag.Flag) {
		logger.Info("Command line flag", "name", flg.Name, "value", flg.Value) // log all arguments
	})

	go func() {
		err := http.ListenAndServe(options.metricsBindAddr, legacyregistry.Handler())
		if err != nil {
			logger.Error(err, "Failed to serve Prometheus metrics", "bindAddress", options.metricsBindAddr)
			panic(err)
		}
	}()

	mymux := mux.NewPathRecorderMux("transport-controller")
	routes.Profiling{}.Install(mymux)
	go func() {
		err := http.ListenAndServe(options.pprofBindAddr, mymux)
		if err != nil {
			logger.Error(err, "Failure in serving /debug/pprof", "bindAddress", options.pprofBindAddr)
			panic(err)
		}
	}()

	// get the config for WDS
	wdsRestConfig, err := options.WdsClientOptions.ToRESTConfig()
	if err != nil {
		logger.Error(err, "unable to build WDS kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	wdsRestConfig.UserAgent = transport.ControllerName

	// get the config for Transport space
	transportRestConfig, err := options.TransportClientOptions.ToRESTConfig()
	if err != nil {
		logger.Error(err, "unable to build transport kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	transportRestConfig.UserAgent = transport.ControllerName
	spacesClientMetrics := ksmetrics.NewMultiSpaceClientMetrics()
	ksmetrics.MustRegister(legacyregistry.Register, spacesClientMetrics)
	// clients for WDS
	wdsClientMetrics := spacesClientMetrics.MetricsForSpace("wds")
	wdsClientset, err := ksclientset.NewForConfig(wdsRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create KubeStellar clientset for Workload Description Space (WDS)")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	wdsDynamicClient, err := dynamic.NewForConfig(wdsRestConfig)
	if err != nil {
		logger.Error(err, "Failed to create dynamic k8s clientset for Workload Description Space (WDS)")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	// clients for transport space
	itsClientMetrics := spacesClientMetrics.MetricsForSpace("its")
	transportClientset, err := kubernetes.NewForConfig(transportRestConfig)
	if err != nil {
		logger.Error(err, "failed to create k8s clientset for transport space")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	transportDynamicClient, err := dynamic.NewForConfig(transportRestConfig)
	if err != nil {
		logger.Error(err, "failed to create dynamic k8s clientset for transport space")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	ocmClientset, err := clusterclient.NewForConfig(transportRestConfig)
	if err != nil {
		logger.Error(err, "failed to create OCM clientset for transport space")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	ocmInformerFactory := clusterinformers.NewSharedInformerFactory(ocmClientset, defaultResyncPeriod)

	inventoryPreInformer := ocmInformerFactory.Cluster().V1().ManagedClusters()

	wdsKsInformerFactory := ksinformers.NewSharedInformerFactoryWithOptions(wdsClientset, defaultResyncPeriod)
	wdsControlInformers := wdsKsInformerFactory.Control().V1alpha1()

	itsK8sInformerFactory := k8sinformers.NewSharedInformerFactory(transportClientset, defaultResyncPeriod)

	transportController, err := transport.NewTransportController(ctx, wdsClientMetrics, itsClientMetrics, inventoryPreInformer,
		wdsClientset.ControlV1alpha1().Bindings(), wdsControlInformers.Bindings(),
		wdsControlInformers.CustomTransforms(),
		transportImplementation, wdsClientset, wdsDynamicClient, transportClientset.CoreV1().Namespaces(), itsK8sInformerFactory.Core().V1().ConfigMaps(),
		transportClientset, transportDynamicClient, options.MaxSizeWrappedObject, options.WdsName)
	if err != nil {
		logger.Error(err, "failed to construct transport controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	transportController.RegisterMetrics(legacyregistry.Register)

	// notice that there is no need to run Start method in a separate goroutine.
	// Start method is non-blocking and runs each of the factory's informers in its own dedicated goroutine.
	ocmInformerFactory.Start(ctx.Done())
	itsK8sInformerFactory.Start(ctx.Done())
	wdsKsInformerFactory.Start(ctx.Done())

	if err := transportController.Run(ctx, options.Concurrency); err != nil {
		logger.Error(err, "failed to run transport controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	logger.Info("transport controller stopped")
}
