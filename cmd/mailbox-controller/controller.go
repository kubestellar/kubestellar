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

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgev2alpha1informers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions/edge/v2alpha1"
	edgev2alpha1listers "github.com/kubestellar/kubestellar/pkg/client/listers/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
	spacev1alpha1 "github.com/kubestellar/kubestellar/space-framework/pkg/apis/space/v1alpha1"
	spaceclientset "github.com/kubestellar/kubestellar/space-framework/pkg/client/clientset/versioned"
	spacev1alpha1informers "github.com/kubestellar/kubestellar/space-framework/pkg/client/informers/externalversions/space/v1alpha1"
	spacev1a1listers "github.com/kubestellar/kubestellar/space-framework/pkg/client/listers/space/v1alpha1"
)

const mbsNameSep = "-mb-"

// This identifies an index in the SyncTarget informer
const mbsNameIndexKey = "mbsName"

// SyncTargetNameAnnotationKey identifies the annotation on a mailbox space
// that points to the corresponding SyncTarget.
const SyncTargetNameAnnotationKey = "edge.kubestellar.io/sync-target-name"

type mbCtl struct {
	context               context.Context
	syncTargetInformer    cache.SharedIndexInformer
	synctargetLister      edgev2alpha1listers.SyncTargetLister
	syncTargetIndexer     cache.Indexer
	spaceInformer         cache.SharedIndexInformer
	spaceLister           spacev1a1listers.SpaceNamespaceLister
	spaceManagementClient spaceclientset.Clientset
	spaceProvider         string
	spaceProviderNs       string
	kbSpaceRelation       kbuser.KubeBindSpaceRelation
	queue                 workqueue.RateLimitingInterface
}

type refSyncTarget string

var errNoSpaceId = errors.New("failed to retrive spaceID")

// newMailboxController constructs a new mailbox controller.
// syncTargetClusterPreInformer is a pre-informer for all the relevant
// SyncTarget objects (not limited to one cluster).
func newMailboxController(ctx context.Context,
	syncTargetPreInformer edgev2alpha1informers.SyncTargetInformer,
	spacePreInformer spacev1alpha1informers.SpaceInformer,
	spaceManagementClient *spaceclientset.Clientset,
	spaceProvider string,
	spaceProviderNs string,
	kbSpaceRelation kbuser.KubeBindSpaceRelation,
) *mbCtl {
	syncTargetInformer := syncTargetPreInformer.Informer()
	spacesInformer := spacePreInformer.Informer()

	ctl := &mbCtl{
		context:               ctx,
		syncTargetInformer:    syncTargetInformer,
		synctargetLister:      syncTargetPreInformer.Lister(),
		syncTargetIndexer:     syncTargetInformer.GetIndexer(),
		spaceInformer:         spacesInformer,
		spaceLister:           spacePreInformer.Lister().Spaces(spaceProviderNs),
		spaceManagementClient: *spaceManagementClient,
		spaceProvider:         spaceProvider,
		spaceProviderNs:       spaceProviderNs,
		kbSpaceRelation:       kbSpaceRelation,
		queue:                 workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "mailbox-controller"),
	}
	syncTargetInformer.AddIndexers(cache.Indexers{mbsNameIndexKey: ctl.mbsNameOfObj})

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
	case *spacev1alpha1.Space:
		logger.V(4).Info("Enqueuing reference due to space", "spaceName", typed.Name)
		ctl.queue.Add(typed.Name)
	case *edgev2alpha1.SyncTarget:
		logger.V(4).Info("Enqueuing SyncTarget reference", "syncTargetName", typed.Name)
		ctl.queue.Add(refSyncTarget(typed.Name))
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

// sync returns true to retry, false on success or unrecoverable error
func (ctl *mbCtl) sync(ctx context.Context, refany any) bool {
	logger := klog.FromContext(ctx)
	if stName, ok := refany.(refSyncTarget); ok {
		st, err := ctl.synctargetLister.Get(string(stName))
		if err != nil {
			logger.Error(err, "Failed to fetch SyncTarget from local cache", "stName", stName)
			return false
		}
		mbsName, err := ctl.mbsNameOfSynctarget(st)
		if err != nil {
			if err.Error() == errNoSpaceId.Error() {
				logger.V(4).Info("Failed to retrieve Space ID. Retrying", "syncTargetName", st.Name)
				return true
			} else {
				logger.Error(err, "Failed to construct mailbox space name from SyncTarget", "syncTargetName", st.Name)
				return false
			}
		}
		ctl.queue.Add(mbsName)
		return false
	}
	mbsName, ok := refany.(string)
	if !ok {
		logger.Error(nil, "Sync expected a string", "ref", refany, "type", fmt.Sprintf("%T", refany))
		return false
	}
	parts := strings.Split(mbsName, mbsNameSep)
	if len(parts) != 2 {
		logger.V(3).Info("Ignoring non-mailbox space name", "spaceName", mbsName)
		return false
	}
	byIndex, err := ctl.syncTargetIndexer.ByIndex(mbsNameIndexKey, mbsName)
	if err != nil {
		logger.Error(err, "Failed to lookup SyncTargets by mailbox space name", "mbsName", mbsName)
		return false
	}
	var syncTarget *edgev2alpha1.SyncTarget
	if len(byIndex) == 0 {
	} else {
		syncTarget = byIndex[0].(*edgev2alpha1.SyncTarget)
		if len(byIndex) > 1 {
			logger.Error(nil, "Impossible: more than one SyncTarget fetched from index; using the first", "mbsName", mbsName, "fetched", byIndex)
		}
	}
	space, err := ctl.spaceLister.Get(mbsName)
	if err != nil && !k8sapierrors.IsNotFound(err) {
		logger.Error(err, "Unable to Get referenced space", "mbsName", mbsName)
		return true
	}
	if syncTarget == nil || syncTarget.DeletionTimestamp != nil {
		if space == nil || space.DeletionTimestamp != nil {
			logger.V(3).Info("Both SyncTarget and Mailbox space are absent or deleting, nothing to do", "mbsName", mbsName)
			return false
		}
		err := ctl.spaceManagementClient.SpaceV1alpha1().Spaces(ctl.spaceProviderNs).Delete(ctx, mbsName, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &space.UID}})
		if err == nil || k8sapierrors.IsNotFound(err) {
			logger.V(2).Info("Deleted unwanted space", "mbsName", mbsName)
			return false
		}
		logger.Error(err, "Failed to delete unwanted space", "mbsName", mbsName)
		return true
	}
	// Now we have established that the SyncTarget exists and is not being deleted
	if space == nil {
		_, stOriginalName, _, err := kbuser.AnalyzeObjectID(syncTarget)
		if err != nil {
			logger.Error(err, "Object does not appear to be a provider's copy of a consumer's object", "syncTarget", syncTarget.Name)
			return false
		}
		// create the mailbox space
		mbspace := &spacev1alpha1.Space{
			ObjectMeta: metav1.ObjectMeta{
				Annotations: map[string]string{SyncTargetNameAnnotationKey: stOriginalName},
				Name:        mbsName,
			},
			Spec: spacev1alpha1.SpaceSpec{
				SpaceProviderDescName: ctl.spaceProvider,
				Type:                  spacev1alpha1.SpaceTypeManaged,
			},
		}
		_, err = ctl.spaceManagementClient.SpaceV1alpha1().Spaces(ctl.spaceProviderNs).Create(ctx, mbspace, metav1.CreateOptions{FieldManager: "mailbox-controller"})
		if err == nil {
			logger.V(2).Info("Created missing space", "mbsName", mbsName)
			return false
		}
		if k8sapierrors.IsAlreadyExists(err) {
			logger.V(3).Info("Missing space was created concurrently", "mbsName", mbsName)
			return false
		}
		logger.Error(err, "Failed to create space", "mbsName", mbsName)
		return true
	}
	if space.DeletionTimestamp != nil {
		logger.V(3).Info("Wanted space is being deleted, will retry later", "mbsName", mbsName)
		return true
	}
	if space.Status.Phase != spacev1alpha1.SpacePhaseReady {
		logger.V(3).Info("Wanted space is not ready, will retry later", "mbsName", mbsName)
		return true
	}
	logger.V(3).Info("Both SyncTarget and Mailbox space exist and are not being deleted, now check on the binding to edge", "mbsName", mbsName)
	return ctl.ensureBinding(ctx, space.Name)

}

func (ctl *mbCtl) ensureBinding(ctx context.Context, spacename string) bool {
	logger := klog.FromContext(ctx).WithValues("mbsName", spacename)

	// The script must be idempotent.
	shellScriptName := "kubestellar-kube-bind"

	resourcesToBind := []string{"syncerconfigs", "edgesyncconfigs"}
	for _, resource := range resourcesToBind {
		logger.V(2).Info("Ensuring binding", "script", shellScriptName, "resource", resource)

		// We assume that SM_CONFIG was set on the script that created the MB controller
		invokeScript := strings.Join([]string{
			"KUBECONFIG=$SM_CONFIG",
			shellScriptName,
			spacename,
			resource,
		}, " ")
		cmdLine := strings.Join([]string{
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
			logger.Error(err, "Unable to bind", "space", spacename, "resrouce", resource)
			return true
		}
	}
	return false
}

func (ctl *mbCtl) mbsNameOfSynctarget(st *edgev2alpha1.SyncTarget) (string, error) {
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
	return spaceID + mbsNameSep + string(st.UID), nil
}

func (ctl *mbCtl) mbsNameOfObj(obj any) ([]string, error) {
	st, ok := obj.(*edgev2alpha1.SyncTarget)
	if !ok {
		return nil, fmt.Errorf("expected a SyncTarget but got %#+v, a %T", obj, obj)
	}
	mbsName, err := ctl.mbsNameOfSynctarget(st)
	if err != nil {
		return nil, err
	}
	return []string{mbsName}, nil
}
