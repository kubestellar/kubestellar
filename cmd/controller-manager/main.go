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
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/pflag"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
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
	// getUniqueIdentity returns a unique identifier for leader election
	lockNamespace    = "kubestellar-system"
	lockName         = "kubestellar-controller-manager"
	fieldManagerName = "kubestellar-controller-manager"
)

// readyFlag indicates whether the controller-manager is ready to serve requests.
// It is set to true when:
// 1. Leadership has been acquired (if leader election is enabled), OR
// 2. Controllers have been started directly (if leader election is disabled)
var readyFlag atomic.Bool
var itsConnectedFlag atomic.Bool

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// getUniqueIdentity returns a unique identifier for leader election
func getUniqueIdentity() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	return fmt.Sprintf("%s-%d-%d", hostname, os.Getpid(), time.Now().UnixNano())
}

// isTransientError determines if an error is transient and should be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common transient error patterns
	switch {
	case errors.IsServerTimeout(err):
		return true
	case errors.IsTooManyRequests(err):
		return true
	case errors.IsServiceUnavailable(err):
		return true
	case errors.IsInternalError(err):
		return true
	case errors.IsTimeout(err):
		return true
	}

	// Check for network-related errors
	errStr := err.Error()
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no route to host",
		"i/o timeout",
	}

	for _, pattern := range transientPatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	return false
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

	// Start /readyz endpoint - be resilient to ITS connection issues
	go func() {
		http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
			// Controller-manager is ready if:
			// 1. It has become leader (if leader election is enabled), OR
			// 2. It's running without leader election (if leader election is disabled)
			// 3. AND it has successfully connected to WDS at least once
			if readyFlag.Load() {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte("not ready - waiting for leadership or controller startup"))
			}
		})
		http.ListenAndServe(":8081", nil)
	}()

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

// startControllersWithLeaderElection uses client-go leader election instead of controller-runtime
func startControllersWithLeaderElection(ctx context.Context, setupLog klog.Logger, wdsRestConfig, itsRestConfig *rest.Config, wdsName, itsName string, allowedGroupsSet sets.Set[string], ctlrsToStart sets.Set[string], wdsClientMetrics, itsClientMetrics ksmetrics.ClientMetrics) {
	// Create WDS client for leader election with retry for transient errors
	var wdsClient *kubernetes.Clientset
	var err error
	for i := 0; i < 3; i++ {
		wdsClient, err = kubernetes.NewForConfig(wdsRestConfig)
		if err == nil {
			break
		}
		if i < 2 && isTransientError(err) {
			setupLog.Info("Failed to create WDS client due to transient error, retrying", "attempt", i+1, "error", err)
			time.Sleep(time.Duration(i+1) * time.Second)
		} else if i < 2 {
			setupLog.Info("Failed to create WDS client, retrying", "attempt", i+1, "error", err)
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	if err != nil {
		setupLog.Error(err, "unable to create WDS client for leader election after retries")
		os.Exit(1)
	}

	// Check/create namespace for leader election with retry for transient errors
	var namespaceErr error
	for i := 0; i < 3; i++ {
		_, namespaceErr = wdsClient.CoreV1().Namespaces().Get(ctx, lockNamespace, metav1.GetOptions{})
		if namespaceErr == nil {
			break
		}
		if errors.IsNotFound(namespaceErr) {
			setupLog.Info("Creating namespace for leader election", "namespace", lockNamespace)
			_, createErr := wdsClient.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: lockNamespace,
				},
			}, metav1.CreateOptions{FieldManager: fieldManagerName})
			if createErr == nil || errors.IsAlreadyExists(createErr) {
				// AlreadyExists is treated as success - another replica or previous run created it
				namespaceErr = nil
				break
			}
			if i < 2 {
				setupLog.Info("Failed to create namespace, retrying", "attempt", i+1, "error", createErr)
				time.Sleep(time.Duration(i+1) * time.Second)
			} else {
				namespaceErr = createErr
			}
		} else {
			if i < 2 && isTransientError(namespaceErr) {
				setupLog.Info("Failed to check namespace due to transient error, retrying", "attempt", i+1, "error", namespaceErr)
				time.Sleep(time.Duration(i+1) * time.Second)
			} else if i < 2 {
				setupLog.Info("Failed to check namespace, retrying", "attempt", i+1, "error", namespaceErr)
				time.Sleep(time.Duration(i+1) * time.Second)
			}
		}
	}
	if namespaceErr != nil {
		setupLog.Error(namespaceErr, "unable to check/create namespace for leader election after retries")
		os.Exit(1)
	}

	// Create resource lock for leader election
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      lockName,
			Namespace: lockNamespace,
		},
		Client: wdsClient.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: getUniqueIdentity(),
		},
	}

	// Configure leader election
	// These durations follow the standard Kubernetes leader election patterns:
	// - LeaseDuration: 60s provides more stability and reduces API server load
	// - RenewDeadline: 40s allows for timely renewal (2/3 of lease duration)
	// - RetryPeriod: 5s reduces retry frequency for better performance
	lec := leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: false, // Set to false as the shutdown behavior constraint is not met
		LeaseDuration:   60 * time.Second,
		RenewDeadline:   40 * time.Second,
		RetryPeriod:     5 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				setupLog.Info("Acquired leadership, starting controllers and workers")
				startControllersDirectly(ctx, setupLog, wdsRestConfig, itsRestConfig, wdsName, itsName, allowedGroupsSet, ctlrsToStart, wdsClientMetrics, itsClientMetrics)
			},
			OnStoppedLeading: func() {
				setupLog.Info("Lost leadership, stopping controllers")
				// Controllers will be stopped when the context is cancelled
			},
			OnNewLeader: func(identity string) {
				setupLog.Info("New leader elected", "identity", identity)
			},
		},
	}

	// Start leader election and block until leadership is acquired
	leaderelection.RunOrDie(ctx, lec)
}

// startControllersDirectly starts controllers immediately without leader election
func startControllersDirectly(ctx context.Context, setupLog klog.Logger, wdsRestConfig, itsRestConfig *rest.Config, wdsName, itsName string, allowedGroupsSet sets.Set[string], ctlrsToStart sets.Set[string], wdsClientMetrics, itsClientMetrics ksmetrics.ClientMetrics) {
	logger := klog.FromContext(ctx) // Get base logger from context for controllers
	workloadEventRelay := &workloadEventRelay{}

	// create the binding controller
	bindingController, err := binding.NewController(logger, wdsClientMetrics, itsClientMetrics, wdsRestConfig, itsRestConfig, wdsName, allowedGroupsSet, workloadEventRelay)
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
		statusController, err = status.NewController(logger, wdsClientMetrics, itsClientMetrics, wdsRestConfig, itsRestConfig, wdsName,
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

	// Mark as ready after controllers are started
	readyFlag.Store(true)
	setupLog.Info("Controllers started successfully, marking as ready")
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
