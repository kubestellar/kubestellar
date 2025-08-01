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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	workStatusSuffixFmt = "%s-%s-%s-%s" // Format: <apiVersion>-<kind>-<namespace>-<name>
)

// ClusterData holds timestamps for a specific cluster
type ClusterData struct {
	manifestWorkName           string
	manifestWorkCreated        time.Time
	appliedManifestWorkCreated time.Time
	wecObjectCreated           time.Time
	wecObjectStatusTime        time.Time
	workStatusTime             time.Time
}

// PerWorkloadCache holds all observed timestamps for one workload
type PerWorkloadCache struct {
	wdsObjectCreated    time.Time
	wdsObjectStatusTime time.Time
	clusterData         map[string]*ClusterData
	gvr                 schema.GroupVersionResource
	gvk                 schema.GroupVersionKind
	name                string
	namespace           string
}

// GenericLatencyCollectorReconciler collects end-to-end latencies across all workloads in a namespace
type GenericLatencyCollectorReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	// Clients for each cluster
	WdsClient   kubernetes.Interface
	WecClients  map[string]kubernetes.Interface
	WdsDynamic  dynamic.Interface
	ItsDynamic  dynamic.Interface
	WecDynamics map[string]dynamic.Interface

	// Configuration
	MonitoredNamespace  string
	BindingName         string
	DiscoveredResources []schema.GroupVersionKind
	gvkToGVR            map[schema.GroupVersionKind]schema.GroupVersionResource // Precomputed GVK to GVR mapping
	bindingCreated      time.Time

	// Cache mapping workload key (namespace/name/kind) -> timestamps
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

//+kubebuilder:rbac:groups=*,resources=*,verbs=get;list;watch
//+kubebuilder:rbac:groups=control.kubestellar.io,resources=bindingpolicies;workstatuses,verbs=get;list
//+kubebuilder:rbac:groups=work.open-cluster-management.io,resources=manifestworks;appliedmanifestworks,verbs=get;list

func (r *GenericLatencyCollectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.cache = make(map[string]*PerWorkloadCache)

	// Precompute GVK to GVR mappings
	mapper := mgr.GetRESTMapper()
	r.gvkToGVR = make(map[schema.GroupVersionKind]schema.GroupVersionResource)
	for _, gvk := range r.DiscoveredResources {
		mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
		if err == nil {
			r.gvkToGVR[gvk] = mapping.Resource
		}
	}

	// Create a controller builder with proper unstructured handling
	preds := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if obj, ok := e.Object.(*unstructured.Unstructured); ok {
				gvk := obj.GetObjectKind().GroupVersionKind()
				if gvk.Kind == "" {
					for _, discoveredGVK := range r.DiscoveredResources {
						if discoveredGVK.Kind == obj.GetKind() {
							obj.SetGroupVersionKind(discoveredGVK)
							break
						}
					}
				}
				return obj.GetObjectKind().GroupVersionKind().Kind != ""
			}
			return e.Object.GetObjectKind().GroupVersionKind().Kind != ""
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if obj, ok := e.ObjectNew.(*unstructured.Unstructured); ok {
				gvk := obj.GetObjectKind().GroupVersionKind()
				if gvk.Kind == "" {
					for _, discoveredGVK := range r.DiscoveredResources {
						if discoveredGVK.Kind == obj.GetKind() {
							obj.SetGroupVersionKind(discoveredGVK)
							break
						}
					}
				}
				return obj.GetObjectKind().GroupVersionKind().Kind != ""
			}
			return e.ObjectNew.GetObjectKind().GroupVersionKind().Kind != ""
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
	}

	excludedGroups := map[string]bool{
		"flowcontrol.apiserver.k8s.io": true,
		"discovery.k8s.io":             true,
		"apiregistration.k8s.io":       true,
		"coordination.k8s.io":          true,
		"control.kubestellar.io":       true,
	}
	excludedResourceNames := map[string]bool{
		"events":               true,
		"nodes":                true,
		"csistoragecapacities": true,
		"csinodes":             true,
		"endpoints":            true,
		"workstatuses":         true,
	}

	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		Named("generic-latency-collector").
		WithEventFilter(preds)

	for _, gvk := range r.DiscoveredResources {
		// Skip entire group
		if excludedGroups[gvk.Group] {
			continue
		}
		// Skip resource name
		kindLower := strings.ToLower(gvk.Kind)
		if excludedResourceNames[kindLower] {
			continue
		}
		// Skip if no GVR mapping
		if _, found := r.gvkToGVR[gvk]; !found {
			continue
		}

		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		controllerBuilder = controllerBuilder.Watches(
			obj,
			&handler.EnqueueRequestForObject{},
		)
	}

	return controllerBuilder.Complete(r)
}

func (r *GenericLatencyCollectorReconciler) RegisterMetrics() {
	// Initialize HistogramVecs for each lifecycle stage
	r.totalPackagingHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_packaging_duration_seconds",
		Help:    "Histogram of WDS object → ManifestWork creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion"})

	r.totalDeliveryHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_delivery_duration_seconds",
		Help:    "Histogram of ManifestWork → AppliedManifestWork creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion"})

	r.totalActivationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_activation_duration_seconds",
		Help:    "Histogram of AppliedManifestWork → WEC object creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion"})

	r.totalDownsyncHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_duration_seconds",
		Help:    "Histogram of WDS object → WEC object creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion"})

	r.totalUpsyncReportHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_upsync_report_duration_seconds",
		Help:    "Histogram of WEC object → WorkStatus report durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion"})

	r.totalUpsyncFinalHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_upsync_finalization_duration_seconds",
		Help:    "Histogram of WorkStatus → WDS object status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion"})

	r.totalUpsyncHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_upsync_duration_seconds",
		Help:    "Histogram of WEC object → WDS object status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion"})

	r.totalE2EHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_e2e_latency_duration_seconds",
		Help:    "Histogram of total binding → WDS status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion"})

	r.workloadCountGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kubestellar_workload_count",
		Help: "Number of workload objects deployed in clusters",
	}, []string{"cluster", "kind", "apiVersion"})

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

func (r *GenericLatencyCollectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	if req.NamespacedName.Namespace != r.MonitoredNamespace {
		return ctrl.Result{}, nil
	}

	obj := &unstructured.Unstructured{}
	var targetGVK schema.GroupVersionKind

	// Try all discovered GVKs to find the object
	for _, gvk := range r.DiscoveredResources {
		obj.SetGroupVersionKind(gvk)
		if err := r.Get(ctx, req.NamespacedName, obj); err == nil {
			targetGVK = gvk
			break
		}
	}

	if targetGVK.Kind == "" {
		return ctrl.Result{}, nil
	}

	// Process the generic object
	r.processGenericObject(ctx, obj)

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *GenericLatencyCollectorReconciler) processGenericObject(ctx context.Context, obj *unstructured.Unstructured) {
	logger := log.FromContext(ctx).WithValues("object", obj.GetName())

	gvk := obj.GetObjectKind().GroupVersionKind()
	if gvk.Kind == "" {
		// Try to extract from object metadata
		if kind, found, _ := unstructured.NestedString(obj.Object, "kind"); found {
			if apiVersion, found, _ := unstructured.NestedString(obj.Object, "apiVersion"); found {
				gv, err := schema.ParseGroupVersion(apiVersion)
				if err == nil {
					gvk = schema.GroupVersionKind{
						Group:   gv.Group,
						Version: gv.Version,
						Kind:    kind,
					}
					obj.SetGroupVersionKind(gvk)
				}
			}
		}
	}

	if gvk.Kind == "" {
		logger.Error(nil, "Unable to determine GVK for object", "object", obj.Object)
		return
	}

	gvr, found := r.gvkToGVR[gvk]
	if !found {
		logger.Info("Skipping object: no GVR mapping found", "gvk", gvk)
		return
	}

	logger.Info("Processing generic object",
		"namespace", obj.GetNamespace(),
		"name", obj.GetName(),
		"kind", gvk.Kind,
		"gvr", gvr)

	key := fmt.Sprintf("%s/%s/%s", strings.ToLower(gvk.Kind), obj.GetNamespace(), obj.GetName())

	r.cacheMux.Lock()
	defer r.cacheMux.Unlock()

	entry, exists := r.cache[key]
	if !exists {
		entry = &PerWorkloadCache{
			gvr:         gvr,
			gvk:         gvk,
			name:        obj.GetName(),
			namespace:   obj.GetNamespace(),
			clusterData: make(map[string]*ClusterData),
		}
		r.cache[key] = entry
	}

	// Update object timestamps
	entry.wdsObjectCreated = obj.GetCreationTimestamp().Time

	// Extract status time using improved method
	if statusTime := r.getGenericStatusTime(obj); !statusTime.IsZero() {
		entry.wdsObjectStatusTime = statusTime
	}

	r.processClusters(ctx, entry, key)
}

func (r *GenericLatencyCollectorReconciler) getGenericStatusTime(obj *unstructured.Unstructured) time.Time {
	// First, try to get from conditions
	var latest time.Time
	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err == nil && found {
		for _, cond := range conditions {
			if condMap, ok := cond.(map[string]interface{}); ok {
				timeFields := []string{"lastUpdateTime", "lastTransitionTime", "lastProbeTime"}
				for _, field := range timeFields {
					timeStr, found, _ := unstructured.NestedString(condMap, field)
					if !found {
						continue
					}
					if ts, err := time.Parse(time.RFC3339, timeStr); err == nil {
						if ts.After(latest) {
							latest = ts
						}
					}
				}
			}
		}
	}

	// Fallback to managed fields for status updates
	if managedFieldsTime := getStatusTime(obj); !managedFieldsTime.IsZero() && managedFieldsTime.After(latest) {
		latest = managedFieldsTime
	}

	return latest
}

func (r *GenericLatencyCollectorReconciler) lookupManifestWorkForCluster(ctx context.Context, key, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues(
		"workload", key,
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

	// match on the originOwnerReferenceBindingKey label
	labelKey := "transport.kubestellar.io/originOwnerReferenceBindingKey"
	selector := fmt.Sprintf("%s=%s", labelKey, r.BindingName)
	list, err := r.ItsDynamic.
		Resource(gvr).
		Namespace(clusterName).
		List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		logger.Error(err, "Failed to list ManifestWorks in cluster namespace")
		return
	}

	logger.Info("Processing ManifestWorks matching binding key", "selector", selector, "count", len(list.Items))
	for _, mw := range list.Items {
		ts := mw.GetCreationTimestamp().Time
		if clusterData.manifestWorkCreated.IsZero() {
			clusterData.manifestWorkName = mw.GetName()
			clusterData.manifestWorkCreated = ts
			logger.Info("📦 ManifestWork creation timestamp recorded",
				"manifestWork", mw.GetName(), "timestamp", ts)
		}
		return
	}

	logger.Info("No ManifestWork found with label", "label", labelKey, "value", r.BindingName)
}

func (r *GenericLatencyCollectorReconciler) lookupAppliedManifestWork(ctx context.Context, key string, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues(
		"workload", key,
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
	for _, aw := range list.Items {
		parts := strings.SplitN(aw.GetName(), "-", 2)
		if len(parts) == 2 && parts[1] == clusterData.manifestWorkName {
			ts := aw.GetCreationTimestamp().Time
			if clusterData.appliedManifestWorkCreated.IsZero() {
				clusterData.appliedManifestWorkCreated = ts
				logger.Info("📬 AppliedManifestWork creation timestamp recorded",
					"appliedManifestWork", aw.GetName(), "timestamp", ts)
			}
			return
		}
	}

	logger.Info("No matching AppliedManifestWork found",
		"manifestWork", clusterData.manifestWorkName)
}

func (r *GenericLatencyCollectorReconciler) lookupWorkStatus(ctx context.Context, key string, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues(
		"workload", key,
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
	// Format: <apiVersion>-<kind>-<namespace>-<name>
	suffix := fmt.Sprintf(workStatusSuffixFmt,
		normalizedGroupVersion(entry.gvk),
		strings.ToLower(entry.gvk.Kind),
		entry.namespace,
		entry.name)

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
					"workStatus", ws.GetName(), "timestamp", ts, "suffix", suffix)
			}
			found = true
			break
		}
	}

	if !found {
		logger.Info("No matching WorkStatus found", "suffix", suffix)
		// Log first 5 WorkStatus names for debugging
		statusNames := make([]string, 0, 5)
		for i, ws := range list.Items {
			if i >= 5 {
				break
			}
			statusNames = append(statusNames, ws.GetName())
		}
		logger.Info("Sample WorkStatus names", "names", statusNames)
	}
}

func (r *GenericLatencyCollectorReconciler) lookupWECObject(ctx context.Context, key string, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues(
		"workload", key,
		"cluster", clusterName,
		"function", "lookupWECObject",
	)

	dynClient, exists := r.WecDynamics[clusterName]
	if !exists {
		logger.Error(nil, "WEC dynamic client not found for cluster")
		return
	}

	// Get the object from WEC using dynamic client
	obj, err := dynClient.Resource(entry.gvr).Namespace(entry.namespace).Get(ctx, entry.name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("WEC object not yet created", "kind", entry.gvk.Kind)
		} else {
			logger.Error(err, "Error fetching WEC object", "kind", entry.gvk.Kind)
		}
		return
	}

	clusterData := entry.clusterData[clusterName]
	if clusterData == nil {
		clusterData = &ClusterData{}
		entry.clusterData[clusterName] = clusterData
	}

	// Record creation timestamp
	createdTime := obj.GetCreationTimestamp().Time
	if clusterData.wecObjectCreated.IsZero() {
		clusterData.wecObjectCreated = createdTime
		logger.Info("🏭 WEC object creation timestamp recorded",
			"kind", entry.gvk.Kind, "timestamp", createdTime)
	}

	// Record status timestamp
	statusTime := r.getGenericStatusTime(obj)
	if !statusTime.IsZero() {
		if clusterData.wecObjectStatusTime.IsZero() || statusTime.After(clusterData.wecObjectStatusTime) {
			clusterData.wecObjectStatusTime = statusTime
			logger.Info("✅ WEC object status timestamp recorded",
				"kind", entry.gvk.Kind, "timestamp", statusTime)
		}
	} else {
		logger.Info("⚠️ No valid status conditions found for WEC object", "kind", entry.gvk.Kind)
	}
}

func (r *GenericLatencyCollectorReconciler) processClusters(ctx context.Context, entry *PerWorkloadCache, key string) {
	logger := log.FromContext(ctx)
	logger.Info("Processing clusters for workload", "name", entry.name, "kind", entry.gvk.Kind)

	// Only record metrics if we have WDS object timestamps
	if entry.wdsObjectCreated.IsZero() {
		logger.Info("Skipping metrics: WDS object creation time not recorded")
		return
	}

	for clusterName := range r.WecClients {
		if entry.clusterData[clusterName] == nil {
			entry.clusterData[clusterName] = &ClusterData{}
		}
		clusterData := entry.clusterData[clusterName]

		r.lookupManifestWorkForCluster(ctx, key, clusterName, entry)
		if clusterData.manifestWorkName != "" {
			r.lookupAppliedManifestWork(ctx, key, clusterName, entry)
			r.lookupWorkStatus(ctx, key, clusterName, entry)
			r.lookupWECObject(ctx, key, clusterName, entry)
		}

		// Record metrics
		now := time.Now()
		apiVersion := entry.gvk.GroupVersion().String()
		labels := prometheus.Labels{
			"workload":   entry.name,
			"cluster":    clusterName,
			"kind":       entry.gvk.Kind,
			"apiVersion": apiVersion,
		}

		// Only record if timestamps are valid
		if !clusterData.manifestWorkCreated.IsZero() {
			r.totalPackagingHistogram.With(labels).Observe(
				duration(entry.wdsObjectCreated, clusterData.manifestWorkCreated, now),
			)
		}

		if !clusterData.manifestWorkCreated.IsZero() && !clusterData.appliedManifestWorkCreated.IsZero() {
			r.totalDeliveryHistogram.With(labels).Observe(
				duration(clusterData.manifestWorkCreated, clusterData.appliedManifestWorkCreated, now),
			)
		}

		if !clusterData.appliedManifestWorkCreated.IsZero() && !clusterData.wecObjectCreated.IsZero() {
			r.totalActivationHistogram.With(labels).Observe(
				duration(clusterData.appliedManifestWorkCreated, clusterData.wecObjectCreated, now),
			)
		}

		if !clusterData.wecObjectCreated.IsZero() {
			r.totalDownsyncHistogram.With(labels).Observe(
				duration(entry.wdsObjectCreated, clusterData.wecObjectCreated, now),
			)
		}

		if !clusterData.wecObjectStatusTime.IsZero() && !clusterData.workStatusTime.IsZero() {
			r.totalUpsyncReportHistogram.With(labels).Observe(
				duration(clusterData.wecObjectStatusTime, clusterData.workStatusTime, now),
			)
		}

		if !clusterData.workStatusTime.IsZero() && !entry.wdsObjectStatusTime.IsZero() {
			r.totalUpsyncFinalHistogram.With(labels).Observe(
				duration(clusterData.workStatusTime, entry.wdsObjectStatusTime, now),
			)
		}

		if !clusterData.wecObjectStatusTime.IsZero() && !entry.wdsObjectStatusTime.IsZero() {
			r.totalUpsyncHistogram.With(labels).Observe(
				duration(clusterData.wecObjectStatusTime, entry.wdsObjectStatusTime, now),
			)
		}

		if !entry.wdsObjectStatusTime.IsZero() {
			r.totalE2EHistogram.With(labels).Observe(
				duration(entry.wdsObjectCreated, entry.wdsObjectStatusTime, now),
			)
		}
	}

	// Update workload count for all clusters
	for clusterName, dynClient := range r.WecDynamics {
		if list, err := dynClient.Resource(entry.gvr).Namespace(r.MonitoredNamespace).List(ctx, metav1.ListOptions{}); err == nil {
			r.workloadCountGauge.WithLabelValues(
				clusterName,
				entry.gvk.Kind,
				entry.gvk.GroupVersion().String(),
			).Set(float64(len(list.Items)))
		}
	}
}

// Helper functions
func duration(start, end, now time.Time) float64 {
	if start.IsZero() {
		return 0
	}
	if end.IsZero() {
		return now.Sub(start).Seconds()
	}
	return end.Sub(start).Seconds()
}

func normalizedGroupVersion(gvk schema.GroupVersionKind) string {
	if gvk.Group == "" && gvk.Version == "v1" {
		return "v1"
	}
	if gvk.Group == "apps" && gvk.Version == "v1" {
		return "appsv1"
	}
	return strings.ToLower(gvk.GroupVersion().String())
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
