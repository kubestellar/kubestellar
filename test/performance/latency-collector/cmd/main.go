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
	"flag"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// Resource groups to exclude for watchers as they should not be delivered to other clusters
var excludedGroups = map[string]bool{
	"flowcontrol.apiserver.k8s.io": true,
	"discovery.k8s.io":             true,
	"apiregistration.k8s.io":       true,
	"coordination.k8s.io":          true,
	"control.kubestellar.io":       true,
}

// Resource names to exclude for watchers as they should not be delivered to other clusters
var excludedResourceNames = map[string]bool{
	"events":               true,
	"nodes":                true,
	"csistoragecapacities": true,
	"csinodes":             true,
	"endpoints":            true,
	"workstatuses":         true,
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// Allow Unstructured (and lists) in v1 core
	gv := schema.GroupVersion{Group: "", Version: "v1"}
	scheme.AddKnownTypes(gv,
		&unstructured.Unstructured{},
		&unstructured.UnstructuredList{},
	)
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
		excludedResources  string
		includedGroups     string
	)

	// Add flags
	pflag.StringVar(&wdsContext, "wds-context", "wds1", "Context name for WDS cluster in kubeconfig")
	pflag.StringVar(&itsContext, "its-context", "its1", "Context name for ITS cluster in kubeconfig")
	pflag.StringVar(&wecContexts, "wec-contexts", "cluster1", "Comma-separated context names for WEC clusters in kubeconfig")
	pflag.StringVar(&kubeconfigPath, "kubeconfig", "~/.kube/config", "Path to kubeconfig file")
	pflag.StringVar(&monitoredNamespace, "monitored-namespace", "default", "Namespace of the resources to monitor")
	pflag.StringVar(&bindingName, "binding-name", "nginx-singleton-bpolicy", "Name of the binding policy for the monitored resources")
	pflag.StringVar(&excludedResources, "excluded-resources", "events,nodes,componentstatuses,endpoints,persistentvolumes,clusterroles,clusterrolebindings", "Comma-separated list of resource types to exclude from monitoring")
	pflag.StringVar(&includedGroups, "included-groups", "", "Comma-separated list of API groups to include (empty means all groups)")

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

	// Create manager with unstructured cache support
	mgrCfg := buildClusterConfig(kubeconfigPath, wdsContext)
	mgr, err := ctrl.NewManager(mgrCfg, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:   metricsAddr,
			SecureServing: secureMetrics,
		},
		WebhookServer:          webhook.NewServer(webhook.Options{}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "34020a28.kubestellar.io",

		// FIXED: Properly configure client with unstructured cache support
		Client: client.Options{
			Cache: &client.CacheOptions{
				DisableFor:   []client.Object{},
				Unstructured: true, // Enable unstructured caching
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Discover available resources
	discoveredResources, err := discoverResources(mgrCfg, excludedResources, includedGroups)
	if err != nil {
		setupLog.Error(err, "unable to discover resources")
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

	// Initialize Generic Reconciler with dependencies
	r := &controller.GenericLatencyCollectorReconciler{
		Client:              mgr.GetClient(),
		Scheme:              mgr.GetScheme(),
		WdsClient:           wdsClient,
		WecClients:          wecClients,
		WdsDynamic:          wdsDynamic,
		ItsDynamic:          itsDynamic,
		WecDynamics:         wecDynamics,
		MonitoredNamespace:  monitoredNamespace,
		BindingName:         bindingName,
		DiscoveredResources: discoveredResources,
	}

	r.RegisterMetrics()
	if err := r.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GenericLatencyCollector")
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

// Update the discoverResources function to return GVK instead of GVR
func discoverResources(config *rest.Config, excludedResources, includedGroups string) ([]schema.GroupVersionKind, error) {
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	discoveryClient := clientset.Discovery()

	// Get all API resources
	serverResources, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	var resources []schema.GroupVersionKind
	excludedSet := make(map[string]bool)
	includedGroupsSet := make(map[string]bool)

	// Parse excluded resources
	if excludedResources != "" {
		for _, res := range strings.Split(excludedResources, ",") {
			excludedSet[strings.TrimSpace(res)] = true
		}
	}

	// Parse included groups
	if includedGroups != "" {
		for _, group := range strings.Split(includedGroups, ",") {
			includedGroupsSet[strings.TrimSpace(group)] = true
		}
	}

	for _, group := range serverResources {
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			continue
		}
		if excludedGroups[gv.Group] {
			continue
		}
		for _, resource := range group.APIResources {
			// Skip subresources
			if strings.Contains(resource.Name, "/") {
				continue
			}

			// Skip by resource-name exclusion
			if excludedResourceNames[resource.Name] {
				continue
			}

			// Skip excluded resources
			if excludedSet[resource.Name] {
				continue
			}

			// Parse group and version
			gv, err := schema.ParseGroupVersion(group.GroupVersion)
			if err != nil {
				continue
			}

			// Filter by included groups if specified
			if len(includedGroupsSet) > 0 && !includedGroupsSet[gv.Group] {
				continue
			}

			// Only include namespaced resources that support list and watch
			if resource.Namespaced && containsVerb(resource.Verbs, "list") && containsVerb(resource.Verbs, "watch") {
				kind := resource.Kind
				gvk := schema.GroupVersionKind{
					Group:   gv.Group,
					Version: gv.Version,
					Kind:    kind,
				}
				resources = append(resources, gvk)
			}
		}
	}

	setupLog.Info("Discovered resources", "count", len(resources))
	return resources, nil
}

func containsVerb(verbs []string, verb string) bool {
	for _, v := range verbs {
		if v == verb {
			return true
		}
	}
	return false
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
