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
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	workloadv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/workload/v1alpha1"
	tenancyclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/typed/tenancy/v1alpha1"
	kcptenancyinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
	kcpworkloadinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/workload/v1alpha1"
	tenancylisters "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	workloadlisters "github.com/kcp-dev/kcp/pkg/client/listers/workload/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"
)

const wsNameSep = "-mb-"

// This identifies an index in the SyncTarget informer
const mbwsNameIndexKey = "mbwsName"

// SyncTargetNameAnnotationKey identifies the annotation on a mailbox Workspace
// that points to the corresponding SyncTarget.
const SyncTargetNameAnnotationKey = "edge.kcp.io/sync-target-name"

type mbCtl struct {
	context                   context.Context
	syncTargetClusterInformer kcpcache.ScopeableSharedIndexInformer
	syncTargetClusterLister   workloadlisters.SyncTargetClusterLister
	syncTargetIndexer         cache.Indexer
	workspaceScopedInformer   cache.SharedIndexInformer
	workspaceScopedLister     tenancylisters.WorkspaceLister
	workspaceScopedClient     tenancyclient.WorkspaceInterface
	queue                     workqueue.RateLimitingInterface // of mailbox workspace Name
}

// newMailboxController constructs a new mailbox controller.
// syncTargetClusterPreInformer is a pre-informer for all the relevant
// SyncTarget objects (not limited to one cluster).
func newMailboxController(ctx context.Context,
	syncTargetClusterPreInformer kcpworkloadinformers.SyncTargetClusterInformer,
	workspaceScopedPreInformer kcptenancyinformers.WorkspaceInformer,
	workspaceScopedClient tenancyclient.WorkspaceInterface,
) *mbCtl {
	syncTargetClusterInformer := syncTargetClusterPreInformer.Informer()
	syncTargetClusterInformer.AddIndexers(cache.Indexers{mbwsNameIndexKey: mbwsNameOfObj})
	workspacesInformer := workspaceScopedPreInformer.Informer()

	ctl := &mbCtl{
		context:                   ctx,
		syncTargetClusterInformer: syncTargetClusterInformer,
		syncTargetClusterLister:   syncTargetClusterPreInformer.Lister(),
		syncTargetIndexer:         syncTargetClusterInformer.GetIndexer(),
		workspaceScopedInformer:   workspacesInformer,
		workspaceScopedLister:     workspaceScopedPreInformer.Lister(),
		workspaceScopedClient:     workspaceScopedClient,
		queue:                     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mailbox-controller"),
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
	case *workloadv1alpha1.SyncTarget:
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
	var syncTarget *workloadv1alpha1.SyncTarget
	if len(byIndex) == 0 {
	} else {
		syncTarget = byIndex[0].(*workloadv1alpha1.SyncTarget)
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
		err := ctl.workspaceScopedClient.Delete(ctx, mbwsName, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &workspace.UID}})
		if err == nil || k8sapierrors.IsNotFound(err) {
			logger.V(2).Info("Deleted unwanted workspace", "mbwsName", mbwsName)
			return false
		}
		logger.Error(err, "Failed to delete unwanted workspace", "mbwsName", mbwsName)
		return true
	}
	// Now we have established that the SyncTarget exists and is not being deleted
	if workspace == nil {
		ws := &tenancyv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{SyncTargetNameAnnotationKey: syncTarget.Name},
				Name:        mbwsName,
			},
			Spec: tenancyv1alpha1.WorkspaceSpec{},
		}
		_, err := ctl.workspaceScopedClient.Create(ctx, ws, metav1.CreateOptions{FieldManager: "mailbox-controller"})
		if err == nil {
			logger.V(2).Info("Created missing workspace", "mbwsName", mbwsName)
			return false
		}
		if k8sapierrors.IsAlreadyExists(err) {
			logger.V(3).Info("Missing workspace was created concurrently", "mbwsName", mbwsName)
			return false
		}
		logger.Error(err, "Failed to create workspace", "mbwsName", mbwsName)
		return true
	}
	if workspace.DeletionTimestamp != nil {
		logger.V(3).Info("Wanted workspace is being deleted, will retry later", "mbwsName", mbwsName)
		return true
	}
	logger.V(3).Info("Both SyncTarget and Workspace exist and are not being deleted, nothing to do", "mbwsName", mbwsName)
	return false

}

func mbwsNameOfSynctarget(st *workloadv1alpha1.SyncTarget) string {
	cluster := logicalcluster.From(st)
	return cluster.String() + wsNameSep + string(st.UID)
}

func mbwsNameOfObj(obj any) ([]string, error) {
	st, ok := obj.(*workloadv1alpha1.SyncTarget)
	if !ok {
		return nil, fmt.Errorf("expected a SyncTarget but got %#+v, a %T", obj, obj)
	}
	return []string{mbwsNameOfSynctarget(st)}, nil
}
