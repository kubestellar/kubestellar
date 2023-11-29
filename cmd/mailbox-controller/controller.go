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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"

	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/typed/tenancy/v1alpha1"
	kcptenancyinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
	tenancylisters "github.com/kcp-dev/kcp/pkg/client/listers/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgev2alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev2alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
)

const wsNameSep = "-mb-"

// This identifies an index in the SyncTarget informer
const mbwsNameIndexKey = "mbwsName"

// SyncTargetNameAnnotationKey identifies the annotation on a mailbox Workspace
// that points to the corresponding SyncTarget.
const SyncTargetNameAnnotationKey = "edge.kubestellar.io/sync-target-name"

type mbCtl struct {
	context                 context.Context
	espwPath                string
	syncTargetInformer      cache.SharedIndexInformer
	synctargetLister        edgev2alpha1listers.SyncTargetLister
	syncTargetIndexer       cache.Indexer
	workspaceScopedInformer cache.SharedIndexInformer
	workspaceScopedLister   tenancylisters.WorkspaceLister
	workspaceScopedClient   tenancyclient.WorkspaceInterface
	kbSpaceRelation         kbuser.KubeBindSpaceRelation
	queue                   workqueue.RateLimitingInterface // of mailbox workspace Name
}

type refSyncTarget string

var suffix uint64 = 142857
var errNoSpaceId = errors.New("failed to retrive spaceID")

// newMailboxController constructs a new mailbox controller.
// syncTargetClusterPreInformer is a pre-informer for all the relevant
// SyncTarget objects (not limited to one cluster).
func newMailboxController(ctx context.Context,
	espwPath string,
	syncTargetPreInformer edgev2alpha1informers.SyncTargetInformer,
	workspaceScopedPreInformer kcptenancyinformers.WorkspaceInformer,
	workspaceScopedClient tenancyclient.WorkspaceInterface,
	kbSpaceRelation kbuser.KubeBindSpaceRelation,
) *mbCtl {
	syncTargetInformer := syncTargetPreInformer.Informer()
	workspacesInformer := workspaceScopedPreInformer.Informer()

	ctl := &mbCtl{
		context:                 ctx,
		espwPath:                espwPath,
		syncTargetInformer:      syncTargetInformer,
		synctargetLister:        syncTargetPreInformer.Lister(),
		syncTargetIndexer:       syncTargetInformer.GetIndexer(),
		workspaceScopedInformer: workspacesInformer,
		workspaceScopedLister:   workspaceScopedPreInformer.Lister(),
		workspaceScopedClient:   workspaceScopedClient,
		kbSpaceRelation:         kbSpaceRelation,
		queue:                   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mailbox-controller"),
	}
	syncTargetInformer.AddIndexers(cache.Indexers{mbwsNameIndexKey: ctl.mbwsNameOfObj})

	syncTargetInformer.AddEventHandler(ctl)
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
	if !cache.WaitForNamedCacheSync("mailbox-controller", doneCh, ctl.syncTargetInformer.HasSynced, ctl.workspaceScopedInformer.HasSynced) {
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
		mbwsName, err := ctl.mbwsNameOfSynctarget(typed)
		if err != nil {
			if err.Error() == errNoSpaceId.Error() {
				logger.V(4).Info("Enqueuing SyncTarget reference for later retry", "syncTargetName", typed.Name)
				ctl.queue.Add(refSyncTarget(typed.Name))
			} else {
				logger.Error(nil, "Failed to construct mailbox workspace name from SyncTarget", "syncTargetName", typed.Name)
			}
		} else {
			logger.V(4).Info("Enqueuing reference due to SyncTarget", "wsName", mbwsName, "syncTargetName", typed.Name)
			ctl.queue.Add(mbwsName)
		}
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
	stName, ok := refany.(refSyncTarget)
	if ok {
		st, err := ctl.synctargetLister.Get(string(stName))
		if err != nil {
			logger.Error(err, "Failed to fetch SyncTarget from local cache", "stName", stName)
			return false
		}
		mbwsName, err := ctl.mbwsNameOfSynctarget(st)
		if err != nil {
			return true
		}
		ctl.queue.Add(mbwsName)
		return false
	}
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
		_, stOriginalName, _, err := kbuser.AnalyzeObjectID(syncTarget)
		if err != nil {
			logger.Error(err, "object does not appear to be a provider's copy of a consumer's object", "syncTarget", syncTarget.Name)
			return false
		}
		ws := &tenancyv1alpha1.Workspace{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{SyncTargetNameAnnotationKey: stOriginalName},
				Name:        mbwsName,
			},
			Spec: tenancyv1alpha1.WorkspaceSpec{},
		}
		_, err = ctl.workspaceScopedClient.Create(ctx, ws, metav1.CreateOptions{FieldManager: "mailbox-controller"})
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
	logger.V(3).Info("Both SyncTarget and Workspace exist and are not being deleted, now check on the binding to edge", "mbwsName", mbwsName)
	return ctl.ensureBinding(ctx, workspace)

}

func (ctl *mbCtl) ensureBinding(ctx context.Context, workspace *tenancyv1alpha1.Workspace) bool {
	logger := klog.FromContext(ctx).WithValues("mbsName", workspace.Name)
	mbwsCluster := logicalcluster.Name(workspace.Spec.Cluster)
	if mbwsCluster == "" {
		logger.V(2).Info("Mailbox workspace does not have a Spec.Cluster yet")
		return true
	}

	// The script must be idempotent.
	shellScriptName := "kubestellar-kube-bind"

	resourcesToBind := []string{"syncerconfigs", "edgesyncconfigs"}
	for idx, resource := range resourcesToBind {
		logger.V(2).Info("Ensuring binding", "script", shellScriptName, "resource", resource)

		// suffix helps isolate this controller's multiple workers to prevent concurrent access of a single kubeconfig file.
		// This is a compromise due to the situation that we are using kcp which modifies kubeconfig file when changing workspaces.
		freshSuffix := atomic.AddUint64(&suffix, 1)

		makeCopyOfKubeConfig := fmt.Sprintf("cp $KUBECONFIG $KUBECONFIG.copy%d", freshSuffix)
		removeCopyWhenExits := fmt.Sprintf("trap 'rm $KUBECONFIG.copy%d' EXIT", freshSuffix)
		invokeScript := strings.Join([]string{
			fmt.Sprintf("KUBECONFIG=$KUBECONFIG.copy%d", freshSuffix),
			shellScriptName,
			workspace.Name,
			resource,
		}, " ")
		if idx == 0 {
			invokeScript = invokeScript + " --start-konnector true"
		}
		cmdLine := strings.Join([]string{
			makeCopyOfKubeConfig,
			removeCopyWhenExits,
			invokeScript,
		}, "; ")
		logger.V(2).Info("About to exec", "cmdLine", cmdLine)
		cmd := exec.Command("/bin/sh", "-c", cmdLine)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			logger.Error(err, "Failed to extract stdout pipe for exec of script", "script", shellScriptName)
			return true
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			logger.Error(err, "Failed to extract stderr pipe for exec of script", "script", shellScriptName)
			return true
		}
		err = cmd.Start()
		if err != nil {
			logger.Error(err, "Unable to start executing the script", "script", shellScriptName)
			return true
		}
		outBuf := bufio.NewReader(stdout)
		errBuf := bufio.NewReader(stderr)
		for {
			line, _, err := outBuf.ReadLine()
			if err != nil {
				if err == io.EOF {
					logger.V(4).Info("End of stdout")
					break
				} else {
					logger.Error(err, "Unable to read stdout")
					return true
				}
			}
			logger.V(2).Info("Stdout from exec", "resource", resource, "line", line)
		}
		for {
			line, _, err := errBuf.ReadLine()
			if err != nil {
				if err == io.EOF {
					logger.V(4).Info("End of stderr")
					break
				} else {
					logger.Error(err, "Unable to read stderr")
					return true
				}
			}
			logger.V(2).Info("Stderr from exec", "resource", resource, "line", line)
		}
		if err = cmd.Wait(); err != nil {
			logger.Error(err, "Unable to bind", "workspace", workspace.Name, "resrouce", resource)
			return true
		}
	}
	return false
}

func (ctl *mbCtl) mbwsNameOfSynctarget(st *edgev2alpha1.SyncTarget) (string, error) {
	logger := klog.FromContext(ctl.context)
	_, _, kbSpaceID, err := kbuser.AnalyzeObjectID(st)
	if err != nil {
		logger.Error(err, "object does not appear to be a provider's copy of a consumer's object", "syncTarget", st.Name)
		return "", err
	}
	spaceID := ctl.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if spaceID == "" {
		logger.Error(errNoSpaceId, "failed to get consumer space ID from a provider's copy", "syncTarget", st.Name)
		return "", errNoSpaceId
	}
	// Use consumer spaceID and provider st.UID
	return spaceID + wsNameSep + string(st.UID), nil
}

func (ctl *mbCtl) mbwsNameOfObj(obj any) ([]string, error) {
	st, ok := obj.(*edgev2alpha1.SyncTarget)
	if !ok {
		return nil, fmt.Errorf("expected a SyncTarget but got %#+v, a %T", obj, obj)
	}
	mbwsName, err := ctl.mbwsNameOfSynctarget(st)
	if err != nil {
		return nil, err
	}
	return []string{mbwsName}, nil
}
