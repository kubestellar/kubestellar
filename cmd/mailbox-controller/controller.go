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

	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgev2alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	spacev1alpha1 "github.com/kubestellar/kubestellar/space-framework/pkg/apis/space/v1alpha1"
	spaceclientset "github.com/kubestellar/kubestellar/space-framework/pkg/client/clientset/versioned"
	spacev1alpha1informers "github.com/kubestellar/kubestellar/space-framework/pkg/client/informers/externalversions/space/v1alpha1"
	spacev1a1listers "github.com/kubestellar/kubestellar/space-framework/pkg/client/listers/space/v1alpha1"
)

const wsNameSep = "-mb-"

// This identifies an index in the SyncTarget informer
const mbwsNameIndexKey = "mbwsName"

// SyncTargetNameAnnotationKey identifies the annotation on a mailbox Workspace
// that points to the corresponding SyncTarget.
const SyncTargetNameAnnotationKey = "edge.kubestellar.io/sync-target-name"

type mbCtl struct {
	context               context.Context
	syncTargetInformer    cache.SharedIndexInformer
	syncTargetIndexer     cache.Indexer
	spaceInformer         cache.SharedIndexInformer
	spaceLister           spacev1a1listers.SpaceNamespaceLister
	spaceManagementClient spaceclientset.Clientset
	spaceProvider         string
	spaceProviderNs       string
	queue                 workqueue.RateLimitingInterface // of mailbox workspace Name
}

// newMailboxController constructs a new mailbox controller.
// syncTargetClusterPreInformer is a pre-informer for all the relevant
// SyncTarget objects (not limited to one cluster).
func newMailboxController(ctx context.Context,
	syncTargetPreInformer edgev2alpha1informers.SyncTargetInformer,
	spacePreInformer spacev1alpha1informers.SpaceInformer,
	spaceManagementClient *spaceclientset.Clientset,
	spaceProvider string,
	spaceProviderNs string,
) *mbCtl {
	syncTargetInformer := syncTargetPreInformer.Informer()
	syncTargetInformer.AddIndexers(cache.Indexers{mbwsNameIndexKey: mbwsNameOfObj})
	spacesInformer := spacePreInformer.Informer()

	ctl := &mbCtl{
		context:               ctx,
		syncTargetInformer:    syncTargetInformer,
		syncTargetIndexer:     syncTargetInformer.GetIndexer(),
		spaceInformer:         spacesInformer,
		spaceLister:           spacePreInformer.Lister().Spaces(spaceProviderNs),
		spaceManagementClient: *spaceManagementClient,
		spaceProvider:         spaceProvider,
		spaceProviderNs:       spaceProviderNs,
		queue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mailbox-controller"),
	}

	syncTargetInformer.AddEventHandler(ctl)
	spacesInformer.AddEventHandler(ctl)
	return ctl
}

// Run animates the controller, finishing and returning when the context of
// the controller is done.
// Call this after the informers have been started.
func (ctl *mbCtl) Run(concurrency int) {
	ctx := ctl.context
	logger := klog.FromContext(ctx)
	doneCh := ctx.Done()
	if !cache.WaitForNamedCacheSync("mailbox-controller", doneCh, ctl.syncTargetInformer.HasSynced, ctl.spaceInformer.HasSynced) {
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
	case *edgev2alpha1.SyncTarget:
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
	var syncTarget *edgev2alpha1.SyncTarget
	if len(byIndex) == 0 {
	} else {
		syncTarget = byIndex[0].(*edgev2alpha1.SyncTarget)
		if len(byIndex) > 1 {
			logger.Error(nil, "Impossible: more than one SyncTarget fetched from index; using the first", "mbwsName", mbwsName, "fetched", byIndex)
		}
	}
	workspace, err := ctl.spaceLister.Get(mbwsName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Unable to Get referenced space", "mbwsName", mbwsName)
		return true
	}
	if syncTarget == nil || syncTarget.DeletionTimestamp != nil {
		if workspace == nil || workspace.DeletionTimestamp != nil {
			logger.V(3).Info("Both SyncTarget and Workspace are absent or deleting, nothing to do", "mbwsName", mbwsName)
			return false
		}
		err := ctl.spaceManagementClient.SpaceV1alpha1().Spaces(ctl.spaceProviderNs).Delete(ctx, mbwsName, metav1.DeleteOptions{})
		if err == nil || k8sapierrors.IsNotFound(err) {
			logger.V(2).Info("Deleted unwanted workspace", "mbwsName", mbwsName)
			return false
		}
		logger.Error(err, "Failed to delete unwanted workspace", "mbwsName", mbwsName)
		return true
	}
	// Now we have established that the SyncTarget exists and is not being deleted
	if workspace == nil {
		// create space for mailbox cluster
		mbspace := &spacev1alpha1.Space{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{SyncTargetNameAnnotationKey: syncTarget.Name},
				Name:        mbwsName,
			},
			Spec: spacev1alpha1.SpaceSpec{
				SpaceProviderDescName: ctl.spaceProvider,
				Type:                  spacev1alpha1.SpaceTypeManaged,
			},
		}
		_, err := ctl.spaceManagementClient.SpaceV1alpha1().Spaces(ctl.spaceProviderNs).Create(ctx, mbspace, metav1.CreateOptions{FieldManager: "mailbox-controller"})
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
	if workspace.Status.Phase != spacev1alpha1.SpacePhaseReady {
		logger.V(3).Info("Wanted workspace is not ready, will retry later", "mbwsName", mbwsName)
		return true
	}
	logger.V(3).Info("Both SyncTarget and Workspace exist and are not being deleted, now check on the binding to edge", "mbwsName", mbwsName)
	return ctl.ensureBinding(ctx, workspace.Name)

}

func (ctl *mbCtl) ensureBinding(ctx context.Context, spacename string) bool {
	logger := klog.FromContext(ctx).WithValues("mbsName", spacename)

	// TODO verify binding is already in place.

	// TODO bind all resources. Note: running in pod, need to access the dex

	// TODO get kube-bind NS from the output and update the Space object(label/annotation). For phase1 just log the NS

	logger.V(2).Info("binding created")
	return false
}

func mbwsNameOfSynctarget(st *edgev2alpha1.SyncTarget) string {
	cluster := logicalcluster.From(st)
	return cluster.String() + wsNameSep + string(st.UID)
}

func mbwsNameOfObj(obj any) ([]string, error) {
	st, ok := obj.(*edgev2alpha1.SyncTarget)
	if !ok {
		return nil, fmt.Errorf("expected a SyncTarget but got %#+v, a %T", obj, obj)
	}
	return []string{mbwsNameOfSynctarget(st)}, nil
}
