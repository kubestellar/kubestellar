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

package controller

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const workStatusSuffixFmt = "appsv1-deployment-%s-%s"

// ClusterData holds timestamps for a specific cluster
type ClusterData struct {
	manifestWorkName           string
	manifestWorkCreated        time.Time
	appliedManifestWorkCreated time.Time
	wecDeploymentCreated       time.Time
	wecDeploymentStatusTime    time.Time
	workStatusTime             time.Time
}

// PerWorkloadCache holds all observed timestamps for one workload
type PerWorkloadCache struct {
	wdsDeploymentCreated    time.Time
	wdsDeploymentStatusTime time.Time
	clusterData             map[string]*ClusterData
}

// LatencyCollectorReconciler collects end-to-end latencies across all workloads in a namespace
type LatencyCollectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Clients for each cluster
	WdsClient   kubernetes.Interface
	WecClients  map[string]kubernetes.Interface
	WdsDynamic  dynamic.Interface
	ItsDynamic  dynamic.Interface
	WecDynamics map[string]dynamic.Interface

	// Configuration
	MonitoredNamespace string
	BindingName        string
	bindingCreated     time.Time

	// Cache mapping workload name -> timestamps
	cache    map[string]*PerWorkloadCache
	cacheMux sync.Mutex

	// Histogram metrics for each stage
	totalPackagingHistogram    *prometheus.HistogramVec
	totalDeliveryHistogram     *prometheus.HistogramVec
	totalActivationHistogram   *prometheus.HistogramVec
	totalDownsyncHistogram     *prometheus.HistogramVec
	totalUpsyncReportHistogram *prometheus.HistogramVec
	totalUpsyncFinalHistogram  *prometheus.HistogramVec
	totalUpsyncHistogram       *prometheus.HistogramVec
	totalE2EHistogram          *prometheus.HistogramVec
	workloadCountGauge         *prometheus.GaugeVec
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
//+kubebuilder:rbac:groups=control.kubestellar.io,resources=bindingpolicies;workstatuses,verbs=get;list
//+kubebuilder:rbac:groups=work.open-cluster-management.io,resources=manifestworks;appliedmanifestworks,verbs=get;list

func (r *LatencyCollectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.cache = make(map[string]*PerWorkloadCache)
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}

func (r *LatencyCollectorReconciler) RegisterMetrics() {
	// Initialize HistogramVecs for each lifecycle stage
	r.totalPackagingHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_packaging_duration_seconds",
		Help:    "Histogram of WDS deployment → ManifestWork creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster"})

	r.totalDeliveryHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_delivery_duration_seconds",
		Help:    "Histogram of ManifestWork → AppliedManifestWork creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster"})

	r.totalActivationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_activation_duration_seconds",
		Help:    "Histogram of AppliedManifestWork → WEC deployment creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster"})

	r.totalDownsyncHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_duration_seconds",
		Help:    "Histogram of WDS deployment → WEC deployment creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster"})

	r.totalUpsyncReportHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_upsync_report_duration_seconds",
		Help:    "Histogram of WEC deployment → WorkStatus report durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster"})

	r.totalUpsyncFinalHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_upsync_finalization_duration_seconds",
		Help:    "Histogram of WorkStatus → WDS Deployment status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster"})

	r.totalUpsyncHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_upsync_duration_seconds",
		Help:    "Histogram of WEC deployment → WDS Deployment status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster"})

	r.totalE2EHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_e2e_latency_duration_seconds",
		Help:    "Histogram of total binding → WDS status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster"})

	r.workloadCountGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kubestellar_workload_count",
		Help: "Number of workload objects deployed in clusters",
	}, []string{"cluster"})

	// Register all histograms
	metrics.Registry.MustRegister(
		r.totalPackagingHistogram,
		r.totalDeliveryHistogram,
		r.totalActivationHistogram,
		r.totalDownsyncHistogram,
		r.totalUpsyncReportHistogram,
		r.totalUpsyncFinalHistogram,
		r.totalUpsyncHistogram,
		r.totalE2EHistogram,
		r.workloadCountGauge,
	)
}

func (r *LatencyCollectorReconciler) lookupManifestWorkForCluster(ctx context.Context, workloadName, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues(
		"workload", workloadName,
		"cluster", clusterName,
		"function", "lookupManifestWorkForCluster",
	)

	clusterData := entry.clusterData[clusterName]
	if clusterData == nil {
		clusterData = &ClusterData{}
		entry.clusterData[clusterName] = clusterData
	}

	gvr := schema.GroupVersionResource{
		Group:    "work.open-cluster-management.io",
		Version:  "v1",
		Resource: "manifestworks",
	}

	// List ManifestWorks in the cluster namespace
	list, err := r.ItsDynamic.Resource(gvr).Namespace(clusterName).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "Failed to list ManifestWorks in cluster namespace")
		return
	}

	logger.Info("Processing ManifestWorks", "count", len(list.Items))
	found := false
	for _, mw := range list.Items {
		manifestSlice, _, _ := unstructured.NestedSlice(mw.Object, "spec", "workload", "manifests")
		for _, m := range manifestSlice {
			if mMap, ok := m.(map[string]interface{}); ok {
				kind, _, _ := unstructured.NestedString(mMap, "kind")
				metaName, _, _ := unstructured.NestedString(mMap, "metadata", "name")
				metaNamespace, _, _ := unstructured.NestedString(mMap, "metadata", "namespace")
				if kind == "Deployment" && metaName == workloadName && metaNamespace == r.MonitoredNamespace {
					ts := mw.GetCreationTimestamp().Time
					if clusterData.manifestWorkCreated.IsZero() {
						clusterData.manifestWorkName = mw.GetName()
						clusterData.manifestWorkCreated = ts
						logger.Info("📦 ManifestWork creation timestamp recorded",
							"manifestWork", mw.GetName(), "timestamp", ts)
					} else {
						logger.Info("📦 ManifestWork already recorded",
							"manifestWork", mw.GetName(), "timestamp", ts)
					}
					found = true
					break
				}
			}
		}
		if found {
			break
		}
	}

	if !found {
		logger.Info("No matching ManifestWork found for workload in cluster namespace")
	}
}

func (r *LatencyCollectorReconciler) lookupAppliedManifestWork(ctx context.Context, workloadName, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues(
		"workload", workloadName,
		"cluster", clusterName,
		"function", "lookupAppliedManifestWork",
	)

	clusterData := entry.clusterData[clusterName]
	if clusterData == nil || clusterData.manifestWorkName == "" {
		logger.Info("Skipping AppliedManifestWork lookup - ManifestWork name unknown")
		return
	}

	dynClient, exists := r.WecDynamics[clusterName]
	if !exists {
		logger.Error(nil, "WEC dynamic client not found for cluster")
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    "work.open-cluster-management.io",
		Version:  "v1",
		Resource: "appliedmanifestworks",
	}

	list, err := dynClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "Failed to list AppliedManifestWorks")
		return
	}

	logger.Info("Processing AppliedManifestWorks", "count", len(list.Items))
	found := false
	for _, aw := range list.Items {
		parts := strings.SplitN(aw.GetName(), "-", 2)
		if len(parts) == 2 && parts[1] == clusterData.manifestWorkName {
			ts := aw.GetCreationTimestamp().Time
			if clusterData.appliedManifestWorkCreated.IsZero() {
				clusterData.appliedManifestWorkCreated = ts
				logger.Info("📬 AppliedManifestWork creation timestamp recorded",
					"appliedManifestWork", aw.GetName(), "timestamp", ts)
			} else {
				logger.Info("📬 AppliedManifestWork already recorded",
					"appliedManifestWork", aw.GetName(), "timestamp", ts)
			}
			found = true
			break
		}
	}

	if !found {
		logger.Info("No matching AppliedManifestWork found",
			"manifestWork", clusterData.manifestWorkName)
	}
}

func (r *LatencyCollectorReconciler) lookupWorkStatus(ctx context.Context, workloadName, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues(
		"workload", workloadName,
		"cluster", clusterName,
		"function", "lookupWorkStatus",
	)

	clusterData := entry.clusterData[clusterName]
	if clusterData == nil {
		return
	}

	gvr := schema.GroupVersionResource{
		Group:    "control.kubestellar.io",
		Version:  "v1alpha1",
		Resource: "workstatuses",
	}

	list, err := r.ItsDynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "Failed to list WorkStatuses")
		return
	}

	logger.Info("Processing WorkStatuses", "count", len(list.Items))
	suffix := fmt.Sprintf("appsv1-deployment-%s-%s", r.MonitoredNamespace, workloadName)
	found := false
	for _, ws := range list.Items {
		if strings.HasSuffix(ws.GetName(), suffix) {
			ts := getStatusTime(&ws)
			if ts.IsZero() {
				logger.Info("WorkStatus found but no valid status timestamp", "workStatus", ws.GetName())
				continue
			}

			if clusterData.workStatusTime.IsZero() {
				clusterData.workStatusTime = ts
				logger.Info("📝 WorkStatus timestamp recorded",
					"workStatus", ws.GetName(), "timestamp", ts)
			} else {
				logger.Info("📝 WorkStatus already recorded",
					"workStatus", ws.GetName(), "timestamp", ts)
			}
			found = true
			break
		}
	}

	if !found {
		logger.Info("No matching WorkStatus found", "suffix", suffix)
	}
}

func (r *LatencyCollectorReconciler) lookupWECDeployment(ctx context.Context, workloadName, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues(
		"workload", workloadName,
		"cluster", clusterName,
		"function", "lookupWECDeployment",
	)

	client, exists := r.WecClients[clusterName]
	if !exists {
		logger.Error(nil, "WEC client not found for cluster")
		return
	}

	dep, err := client.AppsV1().Deployments(r.MonitoredNamespace).Get(ctx, workloadName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("WEC Deployment not yet created")
		} else {
			logger.Error(err, "Error fetching WEC Deployment")
		}
		return
	}

	clusterData := entry.clusterData[clusterName]
	if clusterData == nil {
		clusterData = &ClusterData{}
		entry.clusterData[clusterName] = clusterData
	}

	// Record creation timestamp
	if clusterData.wecDeploymentCreated.IsZero() {
		clusterData.wecDeploymentCreated = dep.CreationTimestamp.Time
		logger.Info("🏭 WEC Deployment creation timestamp recorded",
			"timestamp", dep.CreationTimestamp.Time)
	}

	// Record status timestamp
	if st := getDeploymentStatusTime(dep); !st.IsZero() {
		if clusterData.wecDeploymentStatusTime.IsZero() || st.After(clusterData.wecDeploymentStatusTime) {
			clusterData.wecDeploymentStatusTime = st
			logger.Info("✅ WEC Deployment status timestamp recorded",
				"timestamp", st)
		}
	} else {
		logger.Info("⚠️ No valid status conditions found for WEC Deployment")
	}
}

func (r *LatencyCollectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.NamespacedName.Namespace != r.MonitoredNamespace {
		return ctrl.Result{}, nil
	}

	logger := log.FromContext(ctx).WithValues("workload", req.Name, "function", "Reconcile")
	logger.Info("🔄 Reconcile called for Deployment", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)

	// Fetch deployment
	var deploy appsv1.Deployment
	if err := r.Get(ctx, req.NamespacedName, &deploy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Populate cache entry
	r.cacheMux.Lock()
	entry, exists := r.cache[deploy.Name]
	if !exists {
		entry = &PerWorkloadCache{
			clusterData: make(map[string]*ClusterData),
		}
		r.cache[deploy.Name] = entry
	}
	r.cacheMux.Unlock()

	// Record WDS timestamps
	if entry.wdsDeploymentCreated.IsZero() {
		entry.wdsDeploymentCreated = deploy.CreationTimestamp.Time
	}
	if st := getDeploymentStatusTime(&deploy); !st.IsZero() {
		if entry.wdsDeploymentStatusTime.IsZero() || st.After(entry.wdsDeploymentStatusTime) {
			entry.wdsDeploymentStatusTime = st
		}
	}

	// Ensure we have entries for all clusters
	for clusterName := range r.WecClients {
		if entry.clusterData[clusterName] == nil {
			entry.clusterData[clusterName] = &ClusterData{}
		}
	}

	// Process each cluster
	for clusterName := range r.WecClients {
		logger := logger.WithValues("cluster", clusterName)
		clusterData := entry.clusterData[clusterName]

		// Lookup ManifestWork in cluster namespace
		r.lookupManifestWorkForCluster(ctx, deploy.Name, clusterName, entry)

		// Only proceed if ManifestWork was found
		if clusterData.manifestWorkName != "" {
			r.lookupAppliedManifestWork(ctx, deploy.Name, clusterName, entry)
			r.lookupWorkStatus(ctx, deploy.Name, clusterName, entry)
			r.lookupWECDeployment(ctx, deploy.Name, clusterName, entry)
		} else {
			logger.Info("Skipping cluster - no ManifestWork found")
		}
	}

	for clusterName, wecClient := range r.WecClients {
		deps, err := wecClient.
			AppsV1().
			Deployments(r.MonitoredNamespace).
			List(ctx, metav1.ListOptions{})
		if err != nil {
			log.FromContext(ctx).
				WithValues("cluster", clusterName).
				Error(err, "failed to list deployments for workload-count gauge")
			continue
		}
		total := len(deps.Items)
		r.workloadCountGauge.
			WithLabelValues(clusterName).
			Set(float64(total))
	}

	// Record metrics
	now := time.Now()
	for clusterName, clusterData := range entry.clusterData {
		if clusterData == nil {
			continue
		}

		// Only record metrics if we have ManifestWork data
		if clusterData.manifestWorkName == "" {
			continue
		}

		d := duration(entry.wdsDeploymentCreated, clusterData.manifestWorkCreated, now)
		r.totalPackagingHistogram.WithLabelValues(deploy.Name, clusterName).Observe(d)

		d = duration(clusterData.manifestWorkCreated, clusterData.appliedManifestWorkCreated, now)
		r.totalDeliveryHistogram.WithLabelValues(deploy.Name, clusterName).Observe(d)

		d = duration(clusterData.appliedManifestWorkCreated, clusterData.wecDeploymentCreated, now)
		r.totalActivationHistogram.WithLabelValues(deploy.Name, clusterName).Observe(d)

		d = duration(entry.wdsDeploymentCreated, clusterData.wecDeploymentCreated, now)
		r.totalDownsyncHistogram.WithLabelValues(deploy.Name, clusterName).Observe(d)

		d = duration(clusterData.wecDeploymentStatusTime, clusterData.workStatusTime, now)
		r.totalUpsyncReportHistogram.WithLabelValues(deploy.Name, clusterName).Observe(d)

		d = duration(clusterData.workStatusTime, entry.wdsDeploymentStatusTime, now)
		r.totalUpsyncFinalHistogram.WithLabelValues(deploy.Name, clusterName).Observe(d)

		d = duration(clusterData.wecDeploymentStatusTime, entry.wdsDeploymentStatusTime, now)
		r.totalUpsyncHistogram.WithLabelValues(deploy.Name, clusterName).Observe(d)

		d = duration(entry.wdsDeploymentCreated, entry.wdsDeploymentStatusTime, now)
		r.totalE2EHistogram.WithLabelValues(deploy.Name, clusterName).Observe(d)
	}

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func duration(start, end, now time.Time) float64 {
	if start.IsZero() {
		return 0
	}
	if end.IsZero() {
		return now.Sub(start).Seconds()
	}
	return end.Sub(start).Seconds()
}

func getDeploymentStatusTime(dep *appsv1.Deployment) time.Time {
	var latest time.Time
	for _, cond := range dep.Status.Conditions {
		if !cond.LastUpdateTime.IsZero() && cond.LastUpdateTime.Time.After(latest) {
			latest = cond.LastUpdateTime.Time
		}
	}
	return latest
}

func getStatusTime(obj metav1.Object) time.Time {
	var latest time.Time
	for _, mf := range obj.GetManagedFields() {
		if mf.Operation == metav1.ManagedFieldsOperationUpdate && mf.Subresource == "status" {
			if mf.Time != nil && mf.Time.Time.After(latest) {
				latest = mf.Time.Time
			}
		}
	}
	return latest
}
