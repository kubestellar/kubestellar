/*
Copyright 2025.

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
	"crypto/tls"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/kubestellar/latency-collector/internal/controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

func main() {
	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
		secureMetrics        bool
		enableHTTP2          bool

		// Context flags
		wdsContext         string
		itsContext         string
		wecContexts        string
		kubeconfigPath     string
		monitoredNamespace string
		bindingName        string
	)

	// Add WDS context flag
	pflag.StringVar(&wdsContext, "wds-context", "wds1", "Context name for WDS cluster in kubeconfig")
	pflag.StringVar(&itsContext, "its-context", "its1", "Context name for ITS cluster in kubeconfig")
	pflag.StringVar(&wecContexts, "wec-contexts", "cluster1", "Comma-separated context names for WEC clusters in kubeconfig")
	pflag.StringVar(&kubeconfigPath, "kubeconfig", "~/.kube/config", "Path to kubeconfig file")
	pflag.StringVar(&monitoredNamespace, "monitored-namespace", "default", "Namespace of the deployment to monitor")
	pflag.StringVar(&bindingName, "binding-name", "nginx-singleton-bpolicy", "Name of the binding policy for the monitored deployment")

	// Existing flags
	pflag.StringVar(&metricsAddr, "metrics-bind-address", ":2222", "The address the metric endpoint binds to.")
	pflag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	pflag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	pflag.BoolVar(&secureMetrics, "metrics-secure", false,
		"If set the metrics endpoint is served securely")
	pflag.BoolVar(&enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	pflag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Disable HTTP/2 if not enabled
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	tlsOpts := []func(*tls.Config){}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	// Create manager
	mgrCfg := buildClusterConfig(kubeconfigPath, wdsContext)

	mgr, err := ctrl.NewManager(mgrCfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
			TLSOpts:       tlsOpts,
		},
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "34020a28.kubestellar.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Split WEC contexts
	wecContextList := []string{}
	if wecContexts != "" {
		wecContextList = strings.Split(wecContexts, ",")
	}

	// Build clients for all clusters
	wdsCfg := buildClusterConfig(kubeconfigPath, wdsContext)
	itsCfg := buildClusterConfig(kubeconfigPath, itsContext)

	// Create WEC clients map
	wecClients := make(map[string]kubernetes.Interface)
	wecDynamics := make(map[string]dynamic.Interface)

	for _, ctx := range wecContextList {
		wecCfg := buildClusterConfig(kubeconfigPath, ctx)
		clientSet, err := kubernetes.NewForConfig(wecCfg)
		if err != nil {
			setupLog.Error(err, "unable to create WEC client", "context", ctx)
			os.Exit(1)
		}
		dynClient, err := dynamic.NewForConfig(wecCfg)
		if err != nil {
			setupLog.Error(err, "unable to create WEC dynamic client", "context", ctx)
			os.Exit(1)
		}
		wecClients[ctx] = clientSet
		wecDynamics[ctx] = dynClient
	}

	wdsClient, err := kubernetes.NewForConfig(wdsCfg)
	if err != nil {
		setupLog.Error(err, "unable to create WDS client")
		os.Exit(1)
	}

	wdsDynamic, err := dynamic.NewForConfig(wdsCfg)
	if err != nil {
		setupLog.Error(err, "unable to create WDS dynamic client")
		os.Exit(1)
	}

	itsDynamic, err := dynamic.NewForConfig(itsCfg)
	if err != nil {
		setupLog.Error(err, "unable to create ITS dynamic client")
		os.Exit(1)
	}

	// Initialize Reconciler with dependencies
	r := &controller.LatencyCollectorReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		WdsClient:          wdsClient,
		WecClients:         wecClients,
		WdsDynamic:         wdsDynamic,
		ItsDynamic:         itsDynamic,
		WecDynamics:        wecDynamics,
		MonitoredNamespace: monitoredNamespace,
		BindingName:        bindingName,
	}
	r.RegisterMetrics()
	if err := r.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Deployment")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func buildClusterConfig(kubeconfigPath, context string) *rest.Config {
	if kubeconfigPath == "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			setupLog.Error(err, "unable to load in-cluster config")
			os.Exit(1)
		}
		return config
	}

	if strings.HasPrefix(kubeconfigPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			kubeconfigPath = strings.Replace(kubeconfigPath, "~", home, 1)
		}
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfigPath

	overrides := &clientcmd.ConfigOverrides{}
	if context != "" {
		overrides.CurrentContext = context
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules, overrides).ClientConfig()
	if err != nil {
		setupLog.Error(err, "unable to load kubeconfig", "context", context)
		os.Exit(1)
	}

	config.Timeout = 15 * time.Second
	return config
}
