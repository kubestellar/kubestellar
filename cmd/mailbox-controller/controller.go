/*
Copyright 2022 The KubeStellar Authors.

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
	"context"
	"fmt"
	"strings"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	apisv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	apisclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster/typed/apis/v1alpha1"
	kcptenancyinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
	tenancylisters "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev1alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v1alpha1"
	edgev1alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v1alpha1"
	edgev1alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v1alpha1"
	spacev1alpha1 "github.com/kubestellar/kubestellar/space-framework/pkg/apis/space/v1alpha1"
	spaceclientset "github.com/kubestellar/kubestellar/space-framework/pkg/client/clientset/versioned"
	spacemanager "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager"
)

const wsNameSep = "-mb-"

// This identifies an index in the SyncTarget informer
const mbwsNameIndexKey = "mbwsName"

// SyncTargetNameAnnotationKey identifies the annotation on a mailbox Workspace
// that points to the corresponding SyncTarget.
const SyncTargetNameAnnotationKey = "edge.kubestellar.io/sync-target-name"

type mbCtl struct {
	context                    context.Context
	espwPath                   string
	spaceProvider              string
	syncTargetClusterInformer  kcpcache.ScopeableSharedIndexInformer
	syncTargetClusterLister    edgev1alpha1listers.SyncTargetClusterLister
	syncTargetIndexer          cache.Indexer
	workspaceScopedInformer    cache.SharedIndexInformer
	workspaceScopedLister      tenancylisters.WorkspaceLister
	spaceMgtClientset          *spaceclientset.Clientset
	apiBindingClusterInterface apisclient.APIBindingClusterInterface
	queue                      workqueue.RateLimitingInterface // of mailbox workspace Name
}

// newMailboxController constructs a new mailbox controller.
// syncTargetClusterPreInformer is a pre-informer for all the relevant
// SyncTarget objects (not limited to one cluster).
func newMailboxController(ctx context.Context,
	espwPath string,
	spaceProvider string,
	syncTargetClusterPreInformer edgev1alpha1informers.SyncTargetClusterInformer,
	workspaceScopedPreInformer kcptenancyinformers.WorkspaceInformer,
	managerClientset *spaceclientset.Clientset,
	apiBindingClusterInterface apisclient.APIBindingClusterInterface,
) *mbCtl {
	syncTargetClusterInformer := syncTargetClusterPreInformer.Informer()
	syncTargetClusterInformer.AddIndexers(cache.Indexers{mbwsNameIndexKey: mbwsNameOfObj})
	workspacesInformer := workspaceScopedPreInformer.Informer()

	ctl := &mbCtl{
		context:                    ctx,
		espwPath:                   espwPath,
		spaceProvider:              spaceProvider,
		syncTargetClusterInformer:  syncTargetClusterInformer,
		syncTargetClusterLister:    syncTargetClusterPreInformer.Lister(),
		syncTargetIndexer:          syncTargetClusterInformer.GetIndexer(),
		workspaceScopedInformer:    workspacesInformer,
		workspaceScopedLister:      workspaceScopedPreInformer.Lister(),
		spaceMgtClientset:          managerClientset,
		apiBindingClusterInterface: apiBindingClusterInterface,
		queue:                      workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mailbox-controller"),
	}

	syncTargetClusterInformer.AddEventHandler(ctl)
	workspacesInformer.AddEventHandler(ctl)
	return ctl
}

// Run animates the controller, finishing and returning when the context of
// the controller is done.
// Call this after the informers have been started.
func (ctl *mbCtl) Run(concurrency int) {
	ctx := ctl.context
	logger := klog.FromContext(ctx)
	doneCh := ctx.Done()
	if !cache.WaitForNamedCacheSync("mailbox-controller", doneCh, ctl.syncTargetClusterInformer.HasSynced, ctl.workspaceScopedInformer.HasSynced) {
		logger.Error(nil, "Informer syncs not achieved")
		return
	}
	logger.V(1).Info("Informers synced")
	for worker := 0; worker < concurrency; worker++ {
		go ctl.syncLoop(ctx, worker)
	}
	<-doneCh
}

func (ctl *mbCtl) OnAdd(obj any) {
	logger := klog.FromContext(ctl.context)
	logger.V(4).Info("Observed add", "obj", obj)
	ctl.enqueue(obj)
}

func (ctl *mbCtl) OnUpdate(oldObj, newObj any) {
	logger := klog.FromContext(ctl.context)
	logger.V(4).Info("Observed update", "oldObj", oldObj, "newObj", newObj)
	if newObj != nil {
		ctl.enqueue(newObj)
	} else if oldObj != nil {
		ctl.enqueue(oldObj)
	}
}
func (ctl *mbCtl) OnDelete(obj any) {
	logger := klog.FromContext(ctl.context)
	logger.V(4).Info("Observed delete", "obj", obj)
	ctl.enqueue(obj)
}

func (ctl *mbCtl) enqueue(obj any) {
	logger := klog.FromContext(ctl.context)
	switch typed := obj.(type) {
	case *tenancyv1alpha1.Workspace:
		logger.V(4).Info("Enqueuing reference due to workspace", "wsName", typed.Name)
		ctl.queue.Add(typed.Name)
	case *edgev1alpha1.SyncTarget:
		mbwsName := mbwsNameOfSynctarget(typed)
		logger.V(4).Info("Enqueuing reference due to SyncTarget", "wsName", mbwsName, "syncTargetName", typed.Name)
		ctl.queue.Add(mbwsName)
	default:
		logger.Error(nil, "Notified of object of unexpected type", "object", obj, "type", fmt.Sprintf("%T", obj))
	}
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
	mbwsName, ok := refany.(string)
	if !ok {
		logger.Error(nil, "Sync expected a string", "ref", refany, "type", fmt.Sprintf("%T", refany))
		return false
	}
	parts := strings.Split(mbwsName, wsNameSep)
	if len(parts) != 2 {
		logger.V(3).Info("Ignoring non-mailbox workspace name", "wsName", mbwsName)
		return false
	}
	byIndex, err := ctl.syncTargetIndexer.ByIndex(mbwsNameIndexKey, mbwsName)
	if err != nil {
		logger.Error(err, "Failed to lookup SyncTargets by mailbox workspace name", "mbwsName", mbwsName)
		return false
	}
	var syncTarget *edgev1alpha1.SyncTarget
	if len(byIndex) == 0 {
	} else {
		syncTarget = byIndex[0].(*edgev1alpha1.SyncTarget)
		if len(byIndex) > 1 {
			logger.Error(nil, "Impossible: more than one SyncTarget fetched from index; using the first", "mbwsName", mbwsName, "fetched", byIndex)
		}
	}
	workspace, err := ctl.workspaceScopedLister.Get(mbwsName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Unable to Get referenced Workspace", "mbwsName", mbwsName)
		return true
	}
	if syncTarget == nil || syncTarget.DeletionTimestamp != nil {
		if workspace == nil || workspace.DeletionTimestamp != nil {
			logger.V(3).Info("Both SyncTarget and Workspace are absent or deleting, nothing to do", "mbwsName", mbwsName)
			return false
		}
		err := ctl.spaceMgtClientset.SpaceV1alpha1().Spaces(spacemanager.ProviderNS(ctl.spaceProvider)).Delete(ctx, mbwsName, metav1.DeleteOptions{})
		if err == nil || k8sapierrors.IsNotFound(err) {
			logger.V(2).Info("Deleted unwanted space", "mbwsName", mbwsName)
			return false
		}
		logger.Error(err, "Failed to delete unwanted space", "mbwsName", mbwsName)
		return true
	}
	// Now we have established that the SyncTarget exists and is not being deleted
	if workspace == nil {
		//create space for mailbox cluster
		space := &spacev1alpha1.Space{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{SyncTargetNameAnnotationKey: syncTarget.Name},
				Name:        mbwsName,
			},
			Spec: spacev1alpha1.SpaceSpec{
				SpaceProviderDescName: ctl.spaceProvider,
				Type:                  spacev1alpha1.SpaceTypeManaged,
			},
		}
		_, err := ctl.spaceMgtClientset.SpaceV1alpha1().Spaces(spacemanager.ProviderNS(ctl.spaceProvider)).Create(ctx, space, metav1.CreateOptions{FieldManager: "mailbox-controller"})
		if err == nil {
			logger.V(2).Info("Created missing space", "mbwsName", mbwsName)
			return false
		}
		if k8sapierrors.IsAlreadyExists(err) {
			logger.V(3).Info("Missing space was created concurrently", "mbwsName", mbwsName)
			return false
		}
		logger.Error(err, "Failed to create space", "mbwsName", mbwsName)
		return true
	}
	if workspace.DeletionTimestamp != nil {
		logger.V(3).Info("Wanted workspace is being deleted, will retry later", "mbwsName", mbwsName)
		return true
	}
	logger.V(3).Info("Both SyncTarget and Workspace exist and are not being deleted, now check on the APIBinding to edge", "mbwsName", mbwsName)
	return ctl.ensureEdgeBinding(ctx, workspace)

}

const TheEdgeBindingName = "bind-edge"
const TheEdgeExportName = "edge.kubestellar.io"

func (ctl *mbCtl) ensureEdgeBinding(ctx context.Context, workspace *tenancyv1alpha1.Workspace) bool {
	logger := klog.FromContext(ctx).WithValues("mbwsName", workspace.Name)
	mbwsCluster := logicalcluster.Name(workspace.Spec.Cluster)
	if mbwsCluster == "" {
		logger.V(2).Info("Mailbox workspace does not have a Spec.Cluster yet")
		return true
	}
	logger = logger.WithValues("mbwsCluster", mbwsCluster, "bindingName", TheEdgeBindingName)
	//TODO discuss: we can use space-aware client here, but because APIBinding is a third-party API,
	//              we will need to call ConfigForSpace() and create clientset for each MBWS.
	scopedAPIBindingIfc := ctl.apiBindingClusterInterface.Cluster(mbwsCluster.Path())
	theBinding, err := scopedAPIBindingIfc.Get(ctx, TheEdgeBindingName, metav1.GetOptions{})
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Failed to read APIBinding")
		return true
	}
	if err == nil {
		logger.V(4).Info("Found existing APIBinding, not checking spec", "spec", theBinding.Spec)
		return false
	}
	binding := &apisv1alpha1.APIBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: TheEdgeBindingName,
		},
		Spec: apisv1alpha1.APIBindingSpec{
			Reference: apisv1alpha1.BindingReference{
				Export: &apisv1alpha1.ExportBindingReference{
					Path: ctl.espwPath,
					Name: TheEdgeExportName,
				},
			},
		},
	}
	binding2, err := scopedAPIBindingIfc.Create(ctx, binding, metav1.CreateOptions{FieldManager: "TODO"})
	if err != nil {
		logger.Error(err, "Failed to create APIBinding", "binding", binding)
		return true
	}
	logger.V(2).Info("Created APIBinding", "resourceVersion", binding2.ResourceVersion)
	return false
}

func mbwsNameOfSynctarget(st *edgev1alpha1.SyncTarget) string {
	cluster := logicalcluster.From(st)
	return cluster.String() + wsNameSep + string(st.UID)
}

func mbwsNameOfObj(obj any) ([]string, error) {
	st, ok := obj.(*edgev1alpha1.SyncTarget)
	if !ok {
		return nil, fmt.Errorf("expected a SyncTarget but got %#+v, a %T", obj, obj)
	}
	return []string{mbwsNameOfSynctarget(st)}, nil
}
