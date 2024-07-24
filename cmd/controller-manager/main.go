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

// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
// to ensure that exec-entrypoint and run can make use of them.

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	crmetrics "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	crwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	v1alpha1 "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	clientopts "github.com/kubestellar/kubestellar/options"
	"github.com/kubestellar/kubestellar/pkg/binding"
	"github.com/kubestellar/kubestellar/pkg/status"
	"github.com/kubestellar/kubestellar/pkg/util"
)

var (
	scheme = runtime.NewScheme()
)

const (
	// number of workers to run the reconciliation loop
	workers = 4
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr, pprofAddr, probeAddr string
	var enableLeaderElection bool
	var itsName string
	var wdsName string
	var allowedGroupsString string
	var controllers []string
	pflag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The [host]:port from which /metrics is served.")
	pflag.StringVar(&pprofAddr, "pprof-bind-address", ":8082", "The [host]:port fron which /debug/pprof is served.")
	pflag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	pflag.StringVar(&itsName, "its-name", "", "name of the Inventory and Transport Space to connect to (empty string means to use the only one)")
	pflag.StringVar(&wdsName, "wds-name", "", "name of the workload description space to connect to")
	pflag.StringVar(&allowedGroupsString, "api-groups", "", "list of allowed api groups, comma separated. Empty string means all API groups are allowed")
	pflag.StringSliceVar(&controllers, "controllers", []string{}, "list of controllers to be started by the controller manager, lower case and comma separated, e.g. 'binding,status'. If not specified (or emtpy list specifed), all controllers are started. Currently available controllers are 'binding' and 'status'.")
	pflag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	itsClientLimits := clientopts.NewClientLimits[*pflag.FlagSet]("its", "accessing the ITS")
	wdsClientLimits := clientopts.NewClientLimits[*pflag.FlagSet]("wds", "accessing the WDS")
	itsClientLimits.AddFlags(pflag.CommandLine)
	wdsClientLimits.AddFlags(pflag.CommandLine)
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	ctx := context.Background()
	logger := klog.FromContext(ctx)
	ctrl.SetLogger(logger)
	setupLog := ctrl.Log.WithName("setup")

	pflag.VisitAll(func(flg *pflag.Flag) {
		setupLog.Info("Command line flag", "name", flg.Name, "value", flg.Value)
	})

	// parse allowed resources string
	allowedGroupsSet := util.ParseAPIGroupsString(allowedGroupsString)

	// check controllers flag
	ctlrsToStart := sets.New(controllers...)
	if !sets.New(
		strings.ToLower(binding.ControllerName),
		strings.ToLower(status.ControllerName),
	).IsSuperset(ctlrsToStart) {
		setupLog.Error(fmt.Errorf("unkown controller specified"), "'controllers' flag has incorrect value")
		os.Exit(1)
	}

	// setup manager
	// manager here is mainly used for leader election and health checks
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                crmetrics.Options{BindAddress: metricsAddr},
		PprofBindAddress:       pprofAddr,
		WebhookServer:          crwebhook.NewServer(crwebhook.Options{}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "c6f71c85.kflex.kubestellar.org",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager")
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

	// get the config for WDS
	setupLog.Info("Getting config for WDS", "name", wdsName)
	wdsRestConfig, wdsName, err := util.GetWDSKubeconfig(setupLog, wdsName)
	if err != nil {
		setupLog.Error(err, "unable to get WDS kubeconfig")
		os.Exit(1)
	}
	setupLog.Info("Got config for WDS", "name", wdsName)
	wdsRestConfig = wdsClientLimits.LimitConfig(wdsRestConfig)

	// get the config for ITS
	setupLog.Info("Getting config for ITS")
	itsRestConfig, itsName, err := util.GetITSKubeconfig(setupLog, itsName)
	if err != nil {
		setupLog.Error(err, "unable to get ITS kubeconfig")
		os.Exit(1)
	}
	setupLog.Info("Got config for ITS", "name", itsName)
	itsRestConfig = itsClientLimits.LimitConfig(itsRestConfig)

	// start the binding controller
	bindingController, err := binding.NewController(mgr.GetLogger(), wdsRestConfig, itsRestConfig, wdsName, allowedGroupsSet)
	if err != nil {
		setupLog.Error(err, "unable to create binding controller")
		os.Exit(1)
	}

	if err := bindingController.EnsureCRDs(ctx); err != nil {
		setupLog.Error(err, "error installing the CRDs")
		os.Exit(1)
	}

	if err := bindingController.AppendKSResources(ctx); err != nil {
		setupLog.Error(err, "error appending KubeStellar resources to discovered lists")
		os.Exit(1)
	}

	cListers := make(chan interface{}, 1)

	if len(ctlrsToStart) == 0 || ctlrsToStart.Has(strings.ToLower(binding.ControllerName)) {
		setupLog.Info("Starting controller", "name", binding.ControllerName)
		if err := bindingController.Start(ctx, workers, cListers); err != nil {
			setupLog.Error(err, "error starting the binding controller")
			os.Exit(1)
		}
	}

	// check if status add-on present before starting the status controller
	if util.CheckWorkStatusPresence(itsRestConfig) &&
		(len(ctlrsToStart) == 0 || ctlrsToStart.Has(strings.ToLower(status.ControllerName))) {
		setupLog.Info("Starting controller", "name", status.ControllerName)
		statusController, err := status.NewController(wdsRestConfig, itsRestConfig, wdsName,
			bindingController.GetBindingPolicyResolutionBroker())
		if err != nil {
			setupLog.Error(err, "unable to create status controller")
			os.Exit(1)
		}

		if err := statusController.Start(ctx, workers, cListers); err != nil {
			setupLog.Error(err, "error starting the status controller")
			os.Exit(1)
		}
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
	select {}
}
