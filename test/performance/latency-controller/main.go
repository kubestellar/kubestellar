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

	"github.com/kubestellar/kubestellar/test/performance/latency-controller/internal/controller"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

var excludedGroups = map[string]bool{
	"flowcontrol.apiserver.k8s.io": true,
	"discovery.k8s.io":             true,
	"apiregistration.k8s.io":       true,
	"coordination.k8s.io":          true,
	"control.kubestellar.io":       true,
}

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
		wdsKubeconfigPath  string
		itsKubeconfigPath  string
		wecKubeconfigPaths string
		monitoredNamespace string
		bindingPolicyname  string
		excludedResources  string
		includedGroups     string
	)

	// Flags
	pflag.StringVar(&wdsContext, "wds-context", "wds1", "Context name for WDS cluster in kubeconfig")
	pflag.StringVar(&itsContext, "its-context", "its1", "Context name for ITS cluster in kubeconfig")
	pflag.StringVar(&wecContexts, "wec-contexts", "cluster1", "Comma-separated WEC contexts")
	pflag.StringVar(&kubeconfigPath, "kubeconfig", "~/.kube/config", "Path to kubeconfig file for external use")
	pflag.StringVar(&wdsKubeconfigPath, "wds-kubeconfig", "", "Path to WDS kubeconfig (empty = detect automatically)")
	pflag.StringVar(&itsKubeconfigPath, "its-kubeconfig", "", "Path to ITS kubeconfig (empty = detect automatically)")
	pflag.StringVar(&wecKubeconfigPaths, "wec-kubeconfigs", "", "Comma-separated paths to WEC kubeconfigs")
	pflag.StringVar(&monitoredNamespace, "monitored-namespace", "default", "Namespace to monitor")
	pflag.StringVar(&bindingPolicyname, "binding-policy-name", "nginx-singleton-bpolicy", "Binding policy name")
	pflag.StringVar(&excludedResources, "excluded-resources", "events,nodes,componentstatuses,endpoints,persistentvolumes,clusterroles,clusterrolebindings", "Resources to exclude")
	pflag.StringVar(&includedGroups, "included-groups", "", "API groups to include (empty=all)")

	pflag.StringVar(&metricsAddr, "metrics-bind-address", ":2222", "Address for metrics endpoint")
	pflag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "Address for health probes")
	pflag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election")
	pflag.BoolVar(&secureMetrics, "metrics-secure", false, "Serve metrics securely")
	pflag.BoolVar(&enableHTTP2, "enable-http2", false, "Enable HTTP2 for servers")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	pflag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Build manager config from WDS cluster
	// Automatic detection: use explicit path > in-cluster > default kubeconfig
	mgrCfg := resolveConfig(wdsKubeconfigPath, kubeconfigPath, wdsContext)
	mgr, err := ctrl.NewManager(mgrCfg, ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsserver.Options{BindAddress: metricsAddr, SecureServing: secureMetrics},
		WebhookServer:          webhook.NewServer(webhook.Options{}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "34020a28.kubestellar.io",
		Client:                 client.Options{Cache: &client.CacheOptions{Unstructured: true}},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Discover resources
	discovered, err := discoverResources(mgrCfg, excludedResources, includedGroups)
	if err != nil {
		setupLog.Error(err, "unable to discover resources")
		os.Exit(1)
	}

	// Prepare per-cluster configs
	wdsCfg := resolveConfig(wdsKubeconfigPath, kubeconfigPath, wdsContext)
	itsCfg := resolveConfig(itsKubeconfigPath, kubeconfigPath, itsContext)

	// Build WEC clients
	wecContextList := strings.Split(wecContexts, ",")
	wecClients := make(map[string]kubernetes.Interface)
	wecDynamics := make(map[string]dynamic.Interface)
	if wecKubeconfigPaths != "" {
		paths := strings.Split(wecKubeconfigPaths, ",")
		for i, ctx := range wecContextList {
			path := kubeconfigPath
			if i < len(paths) && paths[i] != "" {
				path = paths[i]
			}
			cfg := resolveConfig(path, kubeconfigPath, ctx)
			cs, err := kubernetes.NewForConfig(cfg)
			utilruntime.Must(err)
			dc, err := dynamic.NewForConfig(cfg)
			utilruntime.Must(err)
			wecClients[ctx], wecDynamics[ctx] = cs, dc
		}
	} else {
		for _, ctx := range wecContextList {
			cfg := resolveConfig(kubeconfigPath, kubeconfigPath, ctx)
			cs, err := kubernetes.NewForConfig(cfg)
			utilruntime.Must(err)
			dc, err := dynamic.NewForConfig(cfg)
			utilruntime.Must(err)
			wecClients[ctx], wecDynamics[ctx] = cs, dc
		}
	}

	// Build core clients
	wdsClientset, err := kubernetes.NewForConfig(wdsCfg)
	utilruntime.Must(err)
	wdsClient := wdsClientset
	wdsDynamic, err := dynamic.NewForConfig(wdsCfg)
	utilruntime.Must(err)
	itsDynamic, err := dynamic.NewForConfig(itsCfg)
	utilruntime.Must(err)

	// Setup controller
	r := &controller.GenericLatencyCollectorReconciler{
		Client:              mgr.GetClient(),
		Scheme:              mgr.GetScheme(),
		WdsClient:           wdsClient,
		WecClients:          wecClients,
		WdsDynamic:          wdsDynamic,
		ItsDynamic:          itsDynamic,
		WecDynamics:         wecDynamics,
		MonitoredNamespace:  monitoredNamespace,
		BindingPolicyName:   bindingPolicyname,
		DiscoveredResources: discovered,
	}
	r.RegisterMetrics()
	if err := r.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	// Health checks
	mgr.AddHealthzCheck("healthz", healthz.Ping)
	mgr.AddReadyzCheck("readyz", healthz.Ping)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// resolveConfig picks explicit path > in-cluster > default kubeconfig
func resolveConfig(explicit, defaultPath, context string) *rest.Config {
	// explicit flag
	if explicit != "" {
		return buildClusterConfig(explicit, context)
	}
	// in-cluster if env present
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		cfg, err := rest.InClusterConfig()
		if err == nil {
			return cfg
		}
	}
	// fallback to default kubeconfig
	return buildClusterConfig(defaultPath, context)
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
