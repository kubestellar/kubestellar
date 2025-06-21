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
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	_ "k8s.io/component-base/metrics/prometheus/version"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"

	v1alpha1 "github.com/kubestellar/kubestellar/api/control/v1alpha1"
	clientopts "github.com/kubestellar/kubestellar/options"
	"github.com/kubestellar/kubestellar/pkg/binding"
	ksctlr "github.com/kubestellar/kubestellar/pkg/controller"
	"github.com/kubestellar/kubestellar/pkg/ctrlutil"
	ksmetrics "github.com/kubestellar/kubestellar/pkg/metrics"
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
	processOpts := clientopts.ProcessOptions{
		MetricsBindAddr:     ":8080",
		HealthProbeBindAddr: ":8081",
		PProfBindAddr:       ":8082",
	}
	var enableLeaderElection bool
	var itsName string
	var wdsName string
	var allowedGroupsString string
	var controllers []string
	pflag.StringVar(&itsName, "its-name", "", "name of the Inventory and Transport Space to connect to (empty string means to use the only one)")
	pflag.StringVar(&wdsName, "wds-name", "", "name of the workload description space to connect to")
	pflag.StringVar(&allowedGroupsString, "api-groups", "", "list of allowed api groups, comma separated. Empty string means all API groups are allowed")
	pflag.StringSliceVar(&controllers, "controllers", []string{}, "list of controllers to be started by the controller manager, lower case and comma separated, e.g. 'binding,status'. If not specified (or emtpy list specifed), all controllers are started. Currently available controllers are 'binding' and 'status'.")
	pflag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	itsClientLimits := clientopts.NewClientLimits[*pflag.FlagSet]("its", "accessing the ITS")
	wdsClientLimits := clientopts.NewClientLimits[*pflag.FlagSet]("wds", "accessing the WDS")
	processOpts.AddToFlags(pflag.CommandLine)
	itsClientLimits.AddFlags(pflag.CommandLine)
	wdsClientLimits.AddFlags(pflag.CommandLine)
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	ctx, _ := ksctlr.InitialContext()
	logger := klog.FromContext(ctx)
	ctrl.SetLogger(logger)
	setupLog := logger.WithName("setup")

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

	ksctlr.Start(ctx, processOpts)

	spacesClientMetrics := ksmetrics.NewMultiSpaceClientMetrics()
	ksmetrics.MustRegister(legacyregistry.Register, spacesClientMetrics)
	wdsClientMetrics := spacesClientMetrics.MetricsForSpace("wds")
	itsClientMetrics := spacesClientMetrics.MetricsForSpace("its")

	// get the config for WDS
	setupLog.Info("Getting config for WDS", "name", wdsName)
	wdsRestConfig, wdsName, err := ctrlutil.GetWDSKubeconfig(setupLog, wdsName)
	if err != nil {
		setupLog.Error(err, "unable to get WDS kubeconfig")
		os.Exit(1)
	}
	setupLog.Info("Got config for WDS", "name", wdsName)
	wdsRestConfig = wdsClientLimits.LimitConfig(wdsRestConfig)

	// get the config for ITS
	setupLog.Info("Getting config for ITS")
	itsRestConfig, itsName, err := ctrlutil.GetITSKubeconfig(setupLog, itsName)
	if err != nil {
		setupLog.Error(err, "unable to get ITS kubeconfig")
		os.Exit(1)
	}
	setupLog.Info("Got config for ITS", "name", itsName)
	itsRestConfig = itsClientLimits.LimitConfig(itsRestConfig)

	// Start controllers based on leader election setting
	if enableLeaderElection {
		setupLog.Info("Starting with leader election enabled")
		startControllersWithLeaderElection(ctx, setupLog, wdsRestConfig, itsRestConfig, wdsName, itsName, allowedGroupsSet, ctlrsToStart, wdsClientMetrics, itsClientMetrics)
	} else {
		setupLog.Info("Starting without leader election")
		startControllersDirectly(ctx, setupLog, wdsRestConfig, itsRestConfig, wdsName, itsName, allowedGroupsSet, ctlrsToStart, wdsClientMetrics, itsClientMetrics)
	}

	select {}
}

// startControllersWithLeaderElection starts controllers only after becoming the leader
func startControllersWithLeaderElection(ctx context.Context, setupLog logr.Logger, wdsRestConfig, itsRestConfig *rest.Config, wdsName, itsName string, allowedGroupsSet sets.Set[string], ctlrsToStart sets.Set[string], wdsClientMetrics, itsClientMetrics ksmetrics.ClientMetrics) {
	// Create a manager with leader election enabled
	mgr, err := ctrl.NewManager(wdsRestConfig, ctrl.Options{
		Scheme:                     scheme,
		LeaderElection:             true,
		LeaderElectionID:           "kubestellar-controller-manager",
		LeaderElectionNamespace:    wdsName + "-system", // Use the WDS name to determine namespace
		LeaderElectionResourceLock: "leases",
		LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	// Start controllers in a goroutine that only runs when we become leader
	go func() {
		<-mgr.Elected()
		setupLog.Info("Became leader, starting controllers")
		startControllersDirectly(ctx, setupLog, wdsRestConfig, itsRestConfig, wdsName, itsName, allowedGroupsSet, ctlrsToStart, wdsClientMetrics, itsClientMetrics)
	}()

	// Start the manager (this will handle leader election)
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
}

// startControllersDirectly starts controllers immediately without leader election
func startControllersDirectly(ctx context.Context, setupLog logr.Logger, wdsRestConfig, itsRestConfig *rest.Config, wdsName, itsName string, allowedGroupsSet sets.Set[string], ctlrsToStart sets.Set[string], wdsClientMetrics, itsClientMetrics ksmetrics.ClientMetrics) {
	workloadEventRelay := &workloadEventRelay{}

	// create the binding controller
	bindingController, err := binding.NewController(setupLog, wdsClientMetrics, itsClientMetrics, wdsRestConfig, itsRestConfig, wdsName, allowedGroupsSet, workloadEventRelay)
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

	startBindingController := len(ctlrsToStart) == 0 || ctlrsToStart.Has(strings.ToLower(binding.ControllerName))
	startStatusCtlr := len(ctlrsToStart) == 0 || ctlrsToStart.Has(strings.ToLower(status.ControllerName))
	var statusController *status.Controller

	if startStatusCtlr {
		if !startBindingController {
			setupLog.Error(nil, "Status controller does not work without binding controller")
			os.Exit(1)
		}
		// check if status add-on present before starting the status controller
		for i := 1; true; i++ {
			if util.CheckWorkStatusPresence(itsRestConfig) {
				break
			}
			if (i & (i - 1)) == 0 {
				setupLog.Info("Not creating status controller yet because WorkStatus is not defined in the ITS")
			}
			i++
			time.Sleep(15 * time.Second)
		}
		setupLog.Info("Creating controller", "name", status.ControllerName)
		statusController, err = status.NewController(setupLog, wdsClientMetrics, itsClientMetrics, wdsRestConfig, itsRestConfig, wdsName,
			bindingController.GetBindingPolicyResolver())
		if err != nil {
			setupLog.Error(err, "unable to create status controller")
			os.Exit(1)
		}
		workloadEventRelay.statusController = statusController
	} else {
		setupLog.Info("Not creating status controller")
	}

	cListers := make(chan interface{}, 1)

	if startBindingController {
		setupLog.Info("Starting controller", "name", binding.ControllerName)
		if err := bindingController.Start(ctx, workers, cListers); err != nil {
			setupLog.Error(err, "error starting the binding controller")
			os.Exit(1)
		}
	}

	if startStatusCtlr {
		setupLog.Info("Starting controller", "name", status.ControllerName)
		if err := statusController.Start(ctx, workers, cListers); err != nil {
			setupLog.Error(err, "error starting the status controller")
			os.Exit(1)
		}
	}
}

// workloadEventRelay implements binding.WorkloadEventHandler and relays the notifications
// to the status controller.
type workloadEventRelay struct {
	statusController *status.Controller
}

var _ binding.WorkloadEventHandler = &workloadEventRelay{}

func (wer *workloadEventRelay) HandleWorkloadObjectEvent(gvr schema.GroupVersionResource, oldObj, obj util.MRObject, eventType binding.WorkloadEventType, wasDeletedFinalStateUnknown bool) {
	if wer.statusController != nil {
		wer.statusController.HandleWorkloadObjectEvent(gvr, oldObj, obj, eventType, wasDeletedFinalStateUnknown)
	}
}
