/*
Copyright 2023 The KCP Authors.

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

package mailboxwatch

import (
	"context"
	"fmt"
	"strings"
	"sync"

	machmeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8ssets "k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/watch"
	upstreamcache "k8s.io/client-go/tools/cache"

	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyv1a1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	"github.com/kcp-dev/edge-mc/pkg/placement"
)

func newFilteredListerWatcher(
	mailboxWorkspacePreInformer tenancyv1a1informers.WorkspaceInformer,
	unfiltered upstreamcache.ListerWatcher,
) *filteredListerWatcher {
	flw := &filteredListerWatcher{
		unfiltered: unfiltered,
		clusters:   k8ssets.NewString(),
	}
	mailboxWorkspacePreInformer.Informer().AddEventHandler(flw)
	return flw
}

type filteredListerWatcher struct {
	sync.Mutex
	unfiltered upstreamcache.ListerWatcher
	clusters   k8ssets.String
}

func (flw *filteredListerWatcher) OnAdd(obj any) {
	flw.setInclusion(obj, true)
}

func (flw *filteredListerWatcher) OnUpdate(old, obj any) {
	flw.setInclusion(obj, true)
}

func (flw *filteredListerWatcher) OnDelete(obj any) {
	flw.setInclusion(obj, false)
}

func (flw *filteredListerWatcher) setInclusion(obj any, include bool) {
	ws := obj.(*tenancyv1a1.Workspace)
	mbwsName := ws.Name
	mbwsNameParts := strings.Split(mbwsName, placement.WSNameSep)
	if len(mbwsNameParts) != 2 {
		// Only accept the workspace if its name looks like a mailbox workspace name
		include = false
	}
	cluster := ws.Spec.Cluster
	flw.Lock()
	defer flw.Unlock()
	if include {
		flw.clusters.Insert(cluster)
	} else {
		flw.clusters.Delete(cluster)
	}
}

type myList struct {
	metav1.TypeMeta
	metav1.ListMeta
	items []runtime.Object
}

var _ upstreamcache.ListerWatcher = &filteredListerWatcher{}

func (flw *filteredListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	unfilteredListObj, err := flw.unfiltered.List(options)
	if err != nil {
		return nil, err
	}
	listMeta, err := machmeta.ListAccessor(unfilteredListObj)
	if err != nil {
		return nil, fmt.Errorf("inner Lister produced a %T (%+v), which machmeta.ListAccessor rejects: %w", unfilteredListObj, unfilteredListObj, err)
	}
	unfilteredList, err := machmeta.ExtractList(unfilteredListObj)
	if err != nil {
		return nil, fmt.Errorf("inner Lister produced a %T (%+v), which machmeta.ExtractList rejects: %w", unfilteredListObj, unfilteredListObj, err)
	}
	flw.Lock()
	defer flw.Unlock()
	filteredSlice := []runtime.Object{}
	for _, item := range unfilteredList {
		objm := item.(metav1.Object)
		cluster := logicalcluster.From(objm)
		if flw.clusters.Has(cluster.String()) {
			filteredSlice = append(filteredSlice, item)
		}
	}
	return &myList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: unfilteredListObj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:       unfilteredListObj.GetObjectKind().GroupVersionKind().Kind,
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion:    listMeta.GetResourceVersion(),
			Continue:           listMeta.GetContinue(),
			RemainingItemCount: listMeta.GetRemainingItemCount(),
		},
		items: filteredSlice,
	}, nil
}

func (ml *myList) DeepCopyObject() runtime.Object {
	ans := myList{
		TypeMeta: ml.TypeMeta,
		ListMeta: *ml.ListMeta.DeepCopy(),
		items:    make([]runtime.Object, len(ml.items)),
	}
	for index, obj := range ml.items {
		ans.items[index] = obj.DeepCopyObject()
	}
	return &ans
}

func (flw *filteredListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	unfilteredWatch, err := flw.unfiltered.Watch(options)
	if err != nil {
		return nil, fmt.Errorf("inner Watcher failed: %w", err)
	}
	ctx := context.Background()
	ctx, close := context.WithCancel(ctx)
	ans := &myWatch{
		unfilteredIfc:  unfilteredWatch,
		unfilteredChan: unfilteredWatch.ResultChan(),
		close:          close,
		doneChan:       ctx.Done(),
		filtered:       make(chan watch.Event),
	}
	go func() {
		for {
			select {
			case <-ans.doneChan:
				ans.unfilteredIfc.Stop()
				return
			case event, ok := <-ans.unfilteredChan:
				if !ok {
					ans.unfilteredIfc.Stop() // do we need this in this case?  https://kubernetes.slack.com/archives/C0EG7JC6T/p1679684882556809
					ans.close()
					return
				}
				objm := event.Object.(metav1.Object)
				cluster := logicalcluster.From(objm)
				pass := func() bool {
					flw.Lock()
					defer flw.Unlock()
					return flw.clusters.Has(cluster.String())
				}()
				if pass {
					ans.filtered <- event
				}
			}
		}
	}()
	return ans, nil
}

type myWatch struct {
	unfilteredIfc  watch.Interface
	unfilteredChan <-chan watch.Event
	close          func()
	doneChan       <-chan struct{}
	filtered       chan watch.Event
}

func (mw *myWatch) Stop() {
	mw.close()
}

func (mw *myWatch) ResultChan() <-chan watch.Event {
	return mw.filtered
}
