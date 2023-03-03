/*
Copyright 2022 The KCP Authors.

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

// Import of k8s.io/client-go/plugin/pkg/client/auth ensures
// that all in-tree Kubernetes client auth plugins
// (e.g. Azure, GCP, OIDC, etc.)  are available.
//
// Import of k8s.io/component-base/metrics/prometheus/clientgo
// makes the k8s client library produce Prometheus metrics.

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/pflag"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	"github.com/kcp-dev/kcp/pkg/apis/third_party/conditions/util/conditions"
	workloadv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/workload/v1alpha1"
	kcpclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpclusterclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	tenancyclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/typed/tenancy/v1alpha1"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	tenancylisters "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	workloadlisters "github.com/kcp-dev/kcp/pkg/client/listers/workload/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"
)

type mbCtl struct {
	context          context.Context
	synctargetLister workloadlisters.SyncTargetClusterLister
	workspaceLister  tenancylisters.WorkspaceLister
	workspaceClient  tenancyclient.WorkspaceInterface
	queue            workqueue.RateLimitingInterface // of syncTargetRef
}

type syncTargetRef struct {
	cluster logicalcluster.Name
	name    string
	uid     apimachtypes.UID
}

// SyncTargetNameAnnotationKey identifies the annotation on a mailbox Workspace
// that points to the corresponding SyncTarget.
const SyncTargetNameAnnotationKey = "edge.kcp.io/sync-target-name"

func main() {
	resyncPeriod := time.Duration(0)
	var concurrency int = 4
	serverBindAddress := ":10203"
	fs := pflag.NewFlagSet("mailbox-controller", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	fs.Var(&utilflag.IPPortVar{Val: &serverBindAddress}, "server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")

	fs.IntVar(&concurrency, "concurrency", concurrency, "number of syncs to run in parallel")

	inventoryLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	inventoryConfigOverrides := &clientcmd.ConfigOverrides{}

	fs.StringVar(&inventoryLoadingRules.ExplicitPath, "inventory-kubeconfig", inventoryLoadingRules.ExplicitPath, "pathname of kubeconfig file for inventory service provider workspace")
	fs.StringVar(&inventoryConfigOverrides.CurrentContext, "inventory-context", "root", "current-context override for inventory-kubeconfig")

	workloadLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	workloadConfigOverrides := &clientcmd.ConfigOverrides{}

	fs.StringVar(&workloadLoadingRules.ExplicitPath, "workload-kubeconfig", workloadLoadingRules.ExplicitPath, "pathname of kubeconfig file for edge workload service provider workspace")
	fs.StringVar(&workloadConfigOverrides.CurrentContext, "workload-context", workloadConfigOverrides.CurrentContext, "current-context override for workload-kubeconfig")

	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	fs.VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info("Command line flag", flg.Name, flg.Value)
	})

	mymux := mux.NewPathRecorderMux("mailbox-controller")
	mymux.Handle("/metrics", legacyregistry.Handler())
	routes.Profiling{}.Install(mymux)
	go func() {
		err := http.ListenAndServe(serverBindAddress, mymux)
		if err != nil {
			logger.Error(err, "Failure in web serving")
			panic(err)
		}
	}()

	// create config for accessing TMC service provider workspace
	inventoryClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(inventoryLoadingRules, inventoryConfigOverrides)
	inventoryConfig, err := inventoryClientConfig.ClientConfig()
	if err != nil {
		logger.Error(err, "failed to make inventory config")
		os.Exit(2)
	}

	inventoryConfig.UserAgent = "mailbox-controller"

	// Get client config for view of SyncTarget objects
	syncTargetViewConfig, err := configForViewOfExport(ctx, inventoryConfig, "workload.kcp.io")
	if err != nil {
		logger.Error(err, "Failed to create client config for view of SyncTarget exports")
		os.Exit(4)
	}

	tmcClusterClientset, err := kcpclusterclientset.NewForConfig(syncTargetViewConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for view of SyncTarget exports")
		os.Exit(6)
	}

	tmcInformerFactory := kcpinformers.NewSharedInformerFactoryWithOptions(tmcClusterClientset, resyncPeriod)
	syncTargetsPreInformer := tmcInformerFactory.Workload().V1alpha1().SyncTargets()
	syncTargetsInformer := syncTargetsPreInformer.Informer()

	// create config for accessing TMC service provider workspace
	workspacesClientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(workloadLoadingRules, workloadConfigOverrides)
	workspacesConfig, err := workspacesClientConfig.ClientConfig()
	if err != nil {
		logger.Error(err, "failed to make workspaces config")
		os.Exit(8)
	}

	workspacesConfig.UserAgent = "mailbox-controller"

	workspacesClientset, err := kcpclientset.NewForConfig(workspacesConfig)
	if err != nil {
		logger.Error(err, "Failed to create clientset for workspaces")
	}

	workspacesInformerFactory := kcpinformers.NewSharedScopedInformerFactoryWithOptions(workspacesClientset, resyncPeriod)
	workspacesPreInformer := workspacesInformerFactory.Tenancy().V1alpha1().Workspaces()
	workspacesInformer := workspacesPreInformer.Informer()

	ctl := &mbCtl{
		context:          ctx,
		synctargetLister: syncTargetsPreInformer.Lister(),
		workspaceLister:  workspacesPreInformer.Lister(),
		workspaceClient:  workspacesClientset.TenancyV1alpha1().Workspaces(),
		queue:            workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mailbox-controller"),
	}

	onAdd := func(obj any) {
		logger.V(4).Info("Observed add", "obj", obj)
		ctl.enqueue(obj)
	}
	onUpdate := func(oldObj, newObj any) {
		logger.V(4).Info("Observed update", "oldObj", oldObj, "newObj", newObj)
		if newObj != nil {
			ctl.enqueue(newObj)
		} else if oldObj != nil {
			ctl.enqueue(oldObj)
		}
	}
	onDelete := func(obj any) {
		logger.V(4).Info("Observed delete", "obj", obj)
		ctl.enqueue(obj)
	}
	syncTargetsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	workspacesInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onAdd,
		UpdateFunc: onUpdate,
		DeleteFunc: onDelete,
	})
	doneCh := ctx.Done()
	tmcInformerFactory.Start(doneCh)
	workspacesInformerFactory.Start(doneCh)
	if cache.WaitForNamedCacheSync("mailbox-controller", doneCh, syncTargetsInformer.HasSynced, workspacesInformer.HasSynced) {
		logger.V(1).Info("Informers synced")
		for worker := 0; worker < concurrency; worker++ {
			go ctl.syncLoop(ctx, worker)
		}
	} else {
		logger.Error(nil, "Informer syncs not achieved")
	}
	<-doneCh
	logger.Info("Time to stop")
}

func configForViewOfExport(ctx context.Context, providerConfig *rest.Config, exportName string) (*rest.Config, error) {
	providerClient, err := kcpclientset.NewForConfig(providerConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating client for service provider workspace: %w", err)
	}
	apiExportClient := providerClient.ApisV1alpha1().APIExports()
	logger := klog.FromContext(ctx)
	var apiExport *apisv1alpha1.APIExport
	for {
		apiExport, err = apiExportClient.Get(ctx, exportName, metav1.GetOptions{})
		if err != nil {
			if k8sapierrors.IsNotFound(err) {
				logger.V(2).Info("Pause because APIExport not found", "exportName", exportName)
				time.Sleep(time.Second * 15)
				continue
			}
			return nil, fmt.Errorf("error reading APIExport %s: %w", exportName, err)
		}
		if isAPIExportReady(logger, apiExport) {
			break
		}
		logger.V(2).Info("Pause because APIExport not ready", "exportName", exportName)
		time.Sleep(time.Second * 15)
	}
	viewConfig := rest.CopyConfig(providerConfig)
	serverURL := apiExport.Status.VirtualWorkspaces[0].URL
	logger.V(2).Info("Found APIExport view", "exportName", exportName, "serverURL", serverURL)
	viewConfig.Host = serverURL
	return viewConfig, nil
}

func isAPIExportReady(logger klog.Logger, apiExport *apisv1alpha1.APIExport) bool {
	if !conditions.IsTrue(apiExport, apisv1alpha1.APIExportVirtualWorkspaceURLsReady) {
		logger.V(2).Info("APIExport virtual workspace URLs are not ready", "APIExport", apiExport.Name)
		return false
	}
	if len(apiExport.Status.VirtualWorkspaces) == 0 {
		logger.V(2).Info("APIExport does not have any virtual workspace URLs", "APIExport", apiExport.Name)
		return false
	}
	return true
}

func (ctl *mbCtl) enqueue(obj any) {
	logger := klog.FromContext(ctl.context)
	switch typed := obj.(type) {
	case *tenancyv1alpha1.Workspace:
		nameParts := strings.Split(typed.Name, wsNameSep)
		if len(nameParts) != 2 {
			logger.V(3).Info("Ignoring workspace with malformed name", "workspace", typed.Name)
			return
		}
		syncTargetName := typed.GetAnnotations()[SyncTargetNameAnnotationKey]
		ref := syncTargetRef{
			cluster: logicalcluster.Name(nameParts[0]),
			name:    syncTargetName,
			uid:     apimachtypes.UID(nameParts[1]),
		}
		logger.V(4).Info("Enqueuing reference due to workspace", "cluster", ref.cluster, "name", ref.name, "uid", ref.uid)
		ctl.queue.Add(ref)
	case *workloadv1alpha1.SyncTarget:
		ref := syncTargetRef{
			cluster: logicalcluster.From(typed),
			name:    typed.Name,
			uid:     typed.UID,
		}
		logger.V(4).Info("Enqueuing reference due to SyncTarget", "cluster", ref.cluster, "name", ref.name, "uid", ref.uid)
		ctl.queue.Add(ref)
	default:
		logger.Error(nil, "Notified of object of unexpected type", "object", obj, "type", fmt.Sprintf("%T", obj))
	}
}

const wsNameSep = "-mb-"

func (ref syncTargetRef) mailboxWSName() string {
	return ref.cluster.String() + wsNameSep + string(ref.uid)
}

func (ctl *mbCtl) syncLoop(ctx context.Context, worker int) {
	doneCh := ctx.Done()
	logger := klog.FromContext(ctx)
	logger = logger.WithValues("worker", worker)
	ctx = klog.NewContext(ctx, logger)
	logger.V(4).Info("SyncLoop start")
	for {
		select {
		case <-doneCh:
			logger.V(2).Info("SyncLoop done")
			return
		default:
			ref, shutdown := ctl.queue.Get()
			if shutdown {
				logger.V(2).Info("Queue shutdown")
				return
			}
			ctl.sync1(ctx, ref)
		}
	}
}

func (ctl *mbCtl) sync1(ctx context.Context, ref any) {
	defer ctl.queue.Done(ref)
	logger := klog.FromContext(ctx)
	logger.V(4).Info("Dequeued reference", "ref", ref)
	retry := ctl.sync(ctx, ref)
	if retry {
		ctl.queue.AddRateLimited(ref)
	} else {
		ctl.queue.Forget(ref)
	}
}

func (ctl *mbCtl) sync(ctx context.Context, refany any) bool {
	logger := klog.FromContext(ctx)
	ref, ok := refany.(syncTargetRef)
	if !ok {
		logger.Error(nil, "Sync expected a syncTargetRef", "ref", refany, "type", fmt.Sprintf("%T", refany))
		return false
	}
	cluster := ref.cluster
	targetName := ref.name
	syncTarget, err := ctl.synctargetLister.Cluster(cluster).Get(targetName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Unable to Get referenced SyncTarget", "ref", ref)
		return true
	}
	wsName := ref.mailboxWSName()
	workspace, err := ctl.workspaceLister.Get(wsName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Unable to Get referenced Workspace", "ref", ref, "wsName", wsName)
		return true
	}
	if syncTarget == nil || syncTarget.DeletionTimestamp != nil {
		if workspace == nil || workspace.DeletionTimestamp != nil {
			logger.V(3).Info("Both absent or deleting, nothing to do", "ref", ref)
			return false
		}
		err := ctl.workspaceClient.Delete(ctx, wsName, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &workspace.UID}})
		if err == nil || k8sapierrors.IsNotFound(err) {
			logger.V(2).Info("Deleted unwanted workspace", "ref", ref)
			return false
		}
		logger.Error(err, "Failed to delete unwanted workspace", "ref", ref)
		return true
	}
	// Now we have established that the SyncTarget exists and is not being deleted
	if workspace == nil {
		ws := &tenancyv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{SyncTargetNameAnnotationKey: ref.name},
				Name:        wsName,
			},
			Spec: tenancyv1alpha1.WorkspaceSpec{},
		}
		_, err := ctl.workspaceClient.Create(ctx, ws, metav1.CreateOptions{FieldManager: "mailbox-controller"})
		if err == nil {
			logger.V(2).Info("Created missing workspace", "ref", ref)
			return false
		}
		if k8sapierrors.IsAlreadyExists(err) {
			logger.V(3).Info("Missing workspace was created concurrently", "ref", ref)
			return false
		}
		logger.Error(err, "Failed to create workspace", "ref", ref)
		return true
	}
	if workspace.DeletionTimestamp != nil {
		logger.V(3).Info("Wanted workspace is being deleted, will retry later", "ref", ref)
		return true
	}
	logger.V(3).Info("Both exist and are not being deleted, nothing to do", "ref", ref)
	return false

}
