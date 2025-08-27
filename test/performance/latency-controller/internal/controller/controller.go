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
	workStatusSuffixFmt = "%s-%s-%s-%s"
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

// GenericLatencyCollectorReconciler collects end-to-end latencies across all workloads
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
	BindingPolicyName   string
	DiscoveredResources []schema.GroupVersionKind
	gvkToGVR            map[schema.GroupVersionKind]schema.GroupVersionResource
	bindingCreated      time.Time

	cache    map[string]*PerWorkloadCache
	cacheMux sync.Mutex

	// Histogram metrics for each stage
	totalPackagingHistogram               *prometheus.HistogramVec
	totalDeliveryHistogram                *prometheus.HistogramVec
	totalActivationHistogram              *prometheus.HistogramVec
	totalDownsyncHistogram                *prometheus.HistogramVec
	totalStatusPropagationReportHistogram *prometheus.HistogramVec
	totalStatusPropagationFinalHistogram  *prometheus.HistogramVec
	totalStatusPropagationHistogram       *prometheus.HistogramVec
	totalE2EHistogram                     *prometheus.HistogramVec
	workloadCountGauge                    *prometheus.GaugeVec
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

	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		Named("generic-latency-collector").
		WithEventFilter(preds)

	for _, gvk := range r.DiscoveredResources {
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
	r.totalPackagingHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_packaging_duration_seconds",
		Help:    "Histogram of WDS object → ManifestWork creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	r.totalDeliveryHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_delivery_duration_seconds",
		Help:    "Histogram of ManifestWork → AppliedManifestWork creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	r.totalActivationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_activation_duration_seconds",
		Help:    "Histogram of AppliedManifestWork → WEC object creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	r.totalDownsyncHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_downsync_duration_seconds",
		Help:    "Histogram of WDS object → WEC object creation durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	r.totalStatusPropagationReportHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_statusPropagation_report_duration_seconds",
		Help:    "Histogram of WEC object → WorkStatus report durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	r.totalStatusPropagationFinalHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_statusPropagation_finalization_duration_seconds",
		Help:    "Histogram of WorkStatus → WDS object status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	r.totalStatusPropagationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_statusPropagation_duration_seconds",
		Help:    "Histogram of WEC object → WDS object status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	r.totalE2EHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kubestellar_e2e_latency_duration_seconds",
		Help:    "Histogram of total binding → WDS status durations",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15),
	}, []string{"workload", "cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	r.workloadCountGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "kubestellar_workload_count",
		Help: "Number of workload objects deployed in clusters",
	}, []string{"cluster", "kind", "apiVersion", "namespace", "bindingPolicyname"})

	metrics.Registry.MustRegister(
		r.totalPackagingHistogram,
		r.totalDeliveryHistogram,
		r.totalActivationHistogram,
		r.totalDownsyncHistogram,
		r.totalStatusPropagationReportHistogram,
		r.totalStatusPropagationFinalHistogram,
		r.totalStatusPropagationHistogram,
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

func (r *GenericLatencyCollectorReconciler) getGenericStatusTime(obj *unstructured.Unstructured) time.Time {
	if managedFieldsTime := getStatusTime(obj); !managedFieldsTime.IsZero() {
		return managedFieldsTime
	}
	var latest time.Time
	conds, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if err == nil && found {
		for _, cond := range conds {
			if condMap, ok := cond.(map[string]interface{}); ok {
				fields := []string{"lastUpdateTime", "lastTransitionTime"}
				for _, field := range fields {
					if timeStr, ok, _ := unstructured.NestedString(condMap, field); ok {
						if ts, err := time.Parse(time.RFC3339, timeStr); err == nil && ts.After(latest) {
							latest = ts
						}
					}
				}
			}
		}
	}
	return latest
}

func (r *GenericLatencyCollectorReconciler) lookupManifestWorkForCluster(ctx context.Context, key, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues("workload", key, "cluster", clusterName, "function", "lookupManifestWorkForCluster")
	clusterData := entry.clusterData[clusterName]
	if clusterData == nil {
		clusterData = &ClusterData{}
		entry.clusterData[clusterName] = clusterData
	}

	gvr := schema.GroupVersionResource{Group: "work.open-cluster-management.io", Version: "v1", Resource: "manifestworks"}
	selector := fmt.Sprintf("transport.kubestellar.io/originOwnerReferenceBindingKey=%s", r.BindingPolicyName)
	list, err := r.ItsDynamic.Resource(gvr).Namespace(clusterName).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		logger.Error(err, "Failed to list ManifestWorks in cluster namespace")
		return
	}

	logger.Info("Processing ManifestWorks matching binding key", "selector", selector, "count", len(list.Items))
	var chosenMW *unstructured.Unstructured
	var chosenTime time.Time

	for _, mw := range list.Items {
		manifests, found, err := unstructured.NestedSlice(mw.Object, "spec", "workload", "manifests")
		if err != nil || !found {
			continue
		}
		for _, m := range manifests {
			if mMap, ok := m.(map[string]interface{}); ok {
				name, nFound, _ := unstructured.NestedString(mMap, "metadata", "name")
				namespace, nsFound, _ := unstructured.NestedString(mMap, "metadata", "namespace")
				_, kFound, _ := unstructured.NestedString(mMap, "kind")
				if nFound && nsFound && kFound && name == entry.name && namespace == entry.namespace {
					ts := mw.GetCreationTimestamp().Time
					if chosenMW == nil || ts.Before(chosenTime) {
						chosenMW = &mw
						chosenTime = ts
					}
					break
				}
			}
		}
	}

	if chosenMW != nil {
		if clusterData.manifestWorkCreated.IsZero() {
			clusterData.manifestWorkName = chosenMW.GetName()
			clusterData.manifestWorkCreated = chosenTime
			logger.Info("📦 ManifestWork creation timestamp recorded",
				"manifestWork", clusterData.manifestWorkName, "timestamp", clusterData.manifestWorkCreated)
		}
	} else {
		logger.Info("No matching ManifestWork found for workload", "workload", entry.name)
	}
}

func (r *GenericLatencyCollectorReconciler) lookupAppliedManifestWork(ctx context.Context, key, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues("workload", key, "cluster", clusterName, "function", "lookupAppliedManifestWork")
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

	gvr := schema.GroupVersionResource{Group: "work.open-cluster-management.io", Version: "v1", Resource: "appliedmanifestworks"}
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

func (r *GenericLatencyCollectorReconciler) lookupWorkStatus(ctx context.Context, key, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues("workload", key, "cluster", clusterName, "function", "lookupWorkStatus")
	clusterData := entry.clusterData[clusterName]
	if clusterData == nil {
		return
	}

	gvr := schema.GroupVersionResource{Group: "control.kubestellar.io", Version: "v1alpha1", Resource: "workstatuses"}
	list, err := r.ItsDynamic.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "Failed to list WorkStatuses")
		return
	}

	suffix := fmt.Sprintf(workStatusSuffixFmt,
		normalizedGroupVersion(entry.gvk),
		strings.ToLower(entry.gvk.Kind),
		entry.namespace,
		entry.name)

	for _, ws := range list.Items {
		if strings.HasSuffix(ws.GetName(), suffix) {
			ts := getStatusTime(&ws)
			if ts.IsZero() {
				continue
			}
			if clusterData.workStatusTime.IsZero() || ts.After(clusterData.workStatusTime) {
				clusterData.workStatusTime = ts
				logger.Info("📝 WorkStatus timestamp recorded",
					"workStatus", ws.GetName(), "timestamp", ts)
			}
			return
		}
	}
	logger.Info("No matching WorkStatus found", "suffix", suffix)
}

func (r *GenericLatencyCollectorReconciler) lookupWECObject(ctx context.Context, key, clusterName string, entry *PerWorkloadCache) {
	logger := log.FromContext(ctx).WithValues("workload", key, "cluster", clusterName, "function", "lookupWECObject")
	dynClient, exists := r.WecDynamics[clusterName]
	if !exists {
		logger.Error(nil, "WEC dynamic client not found for cluster")
		return
	}

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

	createdTime := obj.GetCreationTimestamp().Time
	if clusterData.wecObjectCreated.IsZero() {
		clusterData.wecObjectCreated = createdTime
		logger.Info("🏭 WEC object creation timestamp recorded", "kind", entry.gvk.Kind, "timestamp", createdTime)
	}

	statusTime := r.getGenericStatusTime(obj)
	if !statusTime.IsZero() {
		if clusterData.wecObjectStatusTime.IsZero() || statusTime.After(clusterData.wecObjectStatusTime) {
			clusterData.wecObjectStatusTime = statusTime
			logger.Info("✅ WEC object status timestamp recorded", "kind", entry.gvk.Kind, "timestamp", statusTime)
		}
	} else {
		logger.Info("⚠️ No valid status conditions found for WEC object", "kind", entry.gvk.Kind)
	}
}

func (r *GenericLatencyCollectorReconciler) processGenericObject(ctx context.Context, obj *unstructured.Unstructured) {
	logger := log.FromContext(ctx).WithValues("object", obj.GetName())
	gvk := obj.GetObjectKind().GroupVersionKind()
	if gvk.Kind == "" {
		if kind, found, _ := unstructured.NestedString(obj.Object, "kind"); found {
			if apiVersion, found, _ := unstructured.NestedString(obj.Object, "apiVersion"); found {
				if gv, err := schema.ParseGroupVersion(apiVersion); err == nil {
					gvk = schema.GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind}
					obj.SetGroupVersionKind(gvk)
				}
			}
		}
		if gvk.Kind == "" {
			logger.Error(nil, "Unable to determine GVK for object", "object", obj.Object)
			return
		}
	}

	gvr, found := r.gvkToGVR[gvk]
	if !found {
		logger.Info("Skipping object: no GVR mapping found", "gvk", gvk)
		return
	}
	logger.Info("Processing generic object", "namespace", obj.GetNamespace(), "name", obj.GetName(), "kind", gvk.Kind)

	key := fmt.Sprintf("%s/%s/%s", strings.ToLower(gvk.Kind), obj.GetNamespace(), obj.GetName())
	r.cacheMux.Lock()
	defer r.cacheMux.Unlock()
	entry, exists := r.cache[key]
	hasStatusField := r.objectHasStatusField(gvk.Kind)

	if !exists {
		entry = &PerWorkloadCache{
			gvr:         gvr,
			gvk:         gvk,
			name:        obj.GetName(),
			namespace:   obj.GetNamespace(),
			clusterData: make(map[string]*ClusterData),
		}
		r.cache[key] = entry

		entry.wdsObjectCreated = obj.GetCreationTimestamp().Time
		logger.Info("🟢 WDS object creation timestamp recorded",
			"object", obj.GetName(),
			"namespace", obj.GetNamespace(),
			"kind", gvk.Kind,
			"timestamp", entry.wdsObjectCreated)

		if hasStatusField {
			entry.wdsObjectStatusTime = r.getGenericStatusTime(obj)
			if !entry.wdsObjectStatusTime.IsZero() {
				logger.Info("📌 WDS object status timestamp recorded",
					"object", obj.GetName(),
					"namespace", obj.GetNamespace(),
					"kind", gvk.Kind,
					"timestamp", entry.wdsObjectStatusTime)
			}
		}

		shouldRemove := r.processClusters(ctx, entry, key, true)
		if shouldRemove || !hasStatusField {
			delete(r.cache, key)
			logger.Info("Creation event processed: cache entry purged", "key", key, "hasStatus", hasStatusField)
		} else {
			logger.Info("Creation event processed: cache entry retained for status updates", "key", key)
		}
		return
	}

	if hasStatusField {
		oldStatus := entry.wdsObjectStatusTime
		newStatus := r.getGenericStatusTime(obj)
		if !newStatus.IsZero() && newStatus.After(oldStatus) {
			entry.wdsObjectStatusTime = newStatus
			logger.Info("🔄 WDS object status timestamp updated",
				"object", obj.GetName(),
				"namespace", obj.GetNamespace(),
				"kind", gvk.Kind,
				"old", oldStatus,
				"new", newStatus)

			shouldRemove := r.processClusters(ctx, entry, key, false)
			if shouldRemove {
				delete(r.cache, key)
				logger.Info("Status update event processed: cache entry purged", "key", key)
			}
			return
		}
	}
	logger.Info("No significant changes detected for workload", "key", key)
}

func (r *GenericLatencyCollectorReconciler) objectHasStatusField(kind string) bool {
	noStatusObjects := map[string]bool{
		"ConfigMap":          true,
		"Secret":             true,
		"ServiceAccount":     true,
		"Role":               true,
		"ClusterRole":        true,
		"RoleBinding":        true,
		"ClusterRoleBinding": true,
		"LimitRange":         true,
		"ResourceQuota":      true,
		"NetworkPolicy":      true,
		"StorageClass":       true,
		"Endpoints":          true,
		"EndpointSlice":      true,
	}
	return !noStatusObjects[kind]
}

func (r *GenericLatencyCollectorReconciler) processClusters(ctx context.Context, entry *PerWorkloadCache, key string, isCreationEvent bool) bool {
	logger := log.FromContext(ctx)
	logger.Info("Processing clusters for workload", "name", entry.name, "kind", entry.gvk.Kind, "isCreation", isCreationEvent)
	if entry.wdsObjectCreated.IsZero() {
		logger.Info("Skipping metrics: WDS object creation time not recorded")
		return false
	}
	hasE2ELatency := false

	apiVersion := entry.gvk.GroupVersion().String()
	labels := prometheus.Labels{
		"workload":          entry.name,
		"cluster":           "",
		"kind":              entry.gvk.Kind,
		"apiVersion":        apiVersion,
		"namespace":         entry.namespace,
		"bindingPolicyname": r.BindingPolicyName,
	}

	for clusterName := range r.WecClients {
		if entry.clusterData[clusterName] == nil {
			entry.clusterData[clusterName] = &ClusterData{}
		}
		clusterData := entry.clusterData[clusterName]
		labels["cluster"] = clusterName

		if isCreationEvent {
			r.lookupManifestWorkForCluster(ctx, key, clusterName, entry)
			if clusterData.manifestWorkName != "" {
				r.lookupAppliedManifestWork(ctx, key, clusterName, entry)
				r.lookupWorkStatus(ctx, key, clusterName, entry)
				r.lookupWECObject(ctx, key, clusterName, entry)
			}
		} else {
			if clusterData.manifestWorkName != "" {
				r.lookupWorkStatus(ctx, key, clusterName, entry)
				r.lookupWECObject(ctx, key, clusterName, entry)
			}
		}

		if isCreationEvent && !clusterData.manifestWorkCreated.IsZero() {
			pkg := clusterData.manifestWorkCreated.Sub(entry.wdsObjectCreated).Seconds()
			if pkg < 0 {
				pkg = 0
			}
			r.totalPackagingHistogram.With(labels).Observe(pkg)
		}
		if isCreationEvent && !clusterData.manifestWorkCreated.IsZero() && !clusterData.appliedManifestWorkCreated.IsZero() {
			delivery := clusterData.appliedManifestWorkCreated.Sub(clusterData.manifestWorkCreated).Seconds()
			if delivery < 0 {
				delivery = 0
			}
			r.totalDeliveryHistogram.With(labels).Observe(delivery)
		}
		if isCreationEvent && !clusterData.appliedManifestWorkCreated.IsZero() && !clusterData.wecObjectCreated.IsZero() {
			activation := clusterData.wecObjectCreated.Sub(clusterData.appliedManifestWorkCreated).Seconds()
			if activation < 0 {
				activation = 0
			}
			r.totalActivationHistogram.With(labels).Observe(activation)
		}
		if isCreationEvent && !clusterData.wecObjectCreated.IsZero() {
			downsync := clusterData.wecObjectCreated.Sub(entry.wdsObjectCreated).Seconds()
			if downsync < 0 {
				downsync = 0
			}
			r.totalDownsyncHistogram.With(labels).Observe(downsync)
		}
		if !clusterData.wecObjectStatusTime.IsZero() && !clusterData.workStatusTime.IsZero() {
			report := clusterData.workStatusTime.Sub(clusterData.wecObjectStatusTime).Seconds()
			if report < 0 {
				report = 0
			}
			r.totalStatusPropagationReportHistogram.With(labels).Observe(report)
		}
		if !clusterData.workStatusTime.IsZero() && !entry.wdsObjectStatusTime.IsZero() {
			final := entry.wdsObjectStatusTime.Sub(clusterData.workStatusTime).Seconds()
			if final < 0 {
				final = 0
			}
			r.totalStatusPropagationFinalHistogram.With(labels).Observe(final)
		}
		if !clusterData.wecObjectStatusTime.IsZero() && !entry.wdsObjectStatusTime.IsZero() {
			total := entry.wdsObjectStatusTime.Sub(clusterData.wecObjectStatusTime).Seconds()
			if total < 0 {
				total = 0
			}
			r.totalStatusPropagationHistogram.With(labels).Observe(total)
		}
		if !entry.wdsObjectStatusTime.IsZero() {
			e2e := entry.wdsObjectStatusTime.Sub(entry.wdsObjectCreated).Seconds()
			if e2e < 0 {
				e2e = 0
			}
			r.totalE2EHistogram.With(labels).Observe(e2e)
			hasE2ELatency = true
		}
	}

	if isCreationEvent {
		for clusterName, dynClient := range r.WecDynamics {
			labels := []string{
				clusterName,
				entry.gvk.Kind,
				entry.gvk.GroupVersion().String(),
				entry.namespace,
				r.BindingPolicyName,
			}
			if list, err := dynClient.Resource(entry.gvr).Namespace(r.MonitoredNamespace).List(ctx, metav1.ListOptions{}); err == nil {
				r.workloadCountGauge.WithLabelValues(labels...).Set(float64(len(list.Items)))
			}
		}
	}
	return hasE2ELatency
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
	preferred := map[string]bool{
		"controller-manager":    true,
		"ocm-status-addon":      true,
		"kubelet":               true,
		"Status":                true,
		"Go-http-client":        true,
		"registration-operator": true,
		"transport-controller":  true,
	}
	var latest time.Time
	for _, mf := range obj.GetManagedFields() {
		if mf.Operation == metav1.ManagedFieldsOperationUpdate && mf.Subresource == "status" {
			if mf.Manager != "" && preferred[mf.Manager] {
				if mf.Time != nil && mf.Time.Time.After(latest) {
					latest = mf.Time.Time
					break
				}
			}
		}
	}
	return latest
}
