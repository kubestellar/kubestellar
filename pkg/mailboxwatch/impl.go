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

package mailboxwatch

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	machmeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	machschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	tenancyv1a1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	tenancyv1a1client "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/typed/tenancy/v1alpha1"
	tenancyv1a1informers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	"github.com/kubestellar/kubestellar/pkg/placement"
)

func newCrossClusterListerWatcher[Scoped ScopedListerWatcher[ListType], ListType runtime.Object](
	ctx context.Context,
	listGVK machschema.GroupVersionKind,
	mailboxWorkspacePreInformer tenancyv1a1informers.WorkspaceInformer,
	gen ClusterListerWatcher[Scoped, ListType],
) *crossClusterListerWatcher[Scoped, ListType] {
	clw := &crossClusterListerWatcher[Scoped, ListType]{
		ctx:          ctx,
		listGVK:      listGVK,
		gen:          gen,
		perCluster:   map[logicalcluster.Name]*lwPerCluster[Scoped, ListType]{},
		reconfigChan: make(chan struct{}),
	}
	mailboxWorkspacePreInformer.Informer().AddEventHandler(clw)
	return clw
}

// crossClusterListerWatcher implements upstreamcache.ListerWatcher by scatter/gather to
// a collection of per-cluster ListerWatchers, one for each mailbox workspace.
type crossClusterListerWatcher[Scoped ScopedListerWatcher[ListType], ListType runtime.Object] struct {
	ctx     context.Context
	listGVK machschema.GroupVersionKind
	gen     ClusterListerWatcher[Scoped, ListType]

	sync.Mutex

	perCluster map[logicalcluster.Name]*lwPerCluster[Scoped, ListType]

	// clusterListFirst is the first one to scatter to when doing a List.
	clusterListFirst logicalcluster.Name

	// reconfigChan gets closed and replaced whenever perCluster changes
	reconfigChan chan struct{}
}

type lwPerCluster[Scoped ScopedListerWatcher[ListType], ListType runtime.Object] struct {
	*crossClusterListerWatcher[Scoped, ListType]
	cluster  logicalcluster.Name
	scopedLW Scoped
}

func (clw *crossClusterListerWatcher[Scoped, ListType]) OnAdd(obj any) {
	clw.setInclusion(obj, true)
}

func (clw *crossClusterListerWatcher[Scoped, ListType]) OnUpdate(old, obj any) {
	clw.setInclusion(obj, true)
}

func (clw *crossClusterListerWatcher[Scoped, ListType]) OnDelete(obj any) {
	clw.setInclusion(obj, false)
}

func (clw *crossClusterListerWatcher[Scoped, ListType]) setInclusion(obj any, include bool) {
	ws := obj.(*tenancyv1a1.Workspace)
	mbwsName := ws.Name
	mbwsNameParts := strings.Split(mbwsName, placement.WSNameSep)
	if len(mbwsNameParts) != 2 {
		// Only accept the workspace if its name looks like a mailbox workspace name
		include = false
	}
	clusterStr := ws.Spec.Cluster
	clusterName := logicalcluster.Name(clusterStr)
	clw.Lock()
	defer clw.Unlock()
	_, have := clw.perCluster[clusterName]
	if have == include {
		return
	}
	if include {
		lwForCluster := &lwPerCluster[Scoped, ListType]{clw, clusterName, clw.gen.Cluster(clusterName.Path())}
		clw.perCluster[clusterName] = lwForCluster
	} else {
		delete(clw.perCluster, clusterName)
	}
	close(clw.reconfigChan)
	clw.reconfigChan = make(chan struct{})
}

type myList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []runtime.Object
}

var _ upstreamcache.ListerWatcher = &crossClusterListerWatcher[tenancyv1a1client.WorkspaceInterface, *tenancyv1a1.WorkspaceList]{}

// List queries each mailbox workspace, one at a time, and combines the replies.
// The ResourceVersion is a pain point here.
// Background: all the underlying clusters share in the same progression of ResourceVersion.
// Problem: what ResourceVersiom to put on the whole result?
// The highest safe ResourceVersion for the whole list is the RV returned from the first cluster queried.
// Sadly, we have to parse ResourceVersions and discard items from later queries that are too new.
// As a hueristic to minimze that lossage, List keeps track of which cluster returned the highest RV
// and starts with that one next time.
func (clw *crossClusterListerWatcher[Scoped, ListType]) List(options metav1.ListOptions) (runtime.Object, error) {
	logger := klog.FromContext(clw.ctx)
	allItems := []runtime.Object{}
	var firstListRV *int64
	var maxRV int64 = -1
	var clusterOfMaxRV logicalcluster.Name
	listCluster := func(clusterName logicalcluster.Name, lwForCluster *lwPerCluster[Scoped, ListType]) {
		logger := logger.WithValues("cluster", clusterName)
		sublist, err := lwForCluster.scopedLW.List(clw.ctx, options)
		if err != nil {
			if k8sapierrors.IsNotFound(err) {
				logger.V(4).Info("Resourece not (yet) known")
			} else {
				logger.Error(err, "Failed to list")
			}
			return
		}
		sublistGVK := sublist.GetObjectKind().GroupVersionKind()
		if sublistGVK != clw.listGVK && sublistGVK != (machschema.GroupVersionKind{}) {
			logger.Error(nil, "List returned unexpected GroupVersionKind", "expected", clw.listGVK, "got", sublistGVK)
			return
		}
		subItems, err := machmeta.ExtractList(sublist)
		if err != nil {
			logger.Error(err, "Failed to machmeta.ExtractList", "sublist", sublist)
			return
		}
		listMeta, err := machmeta.ListAccessor(sublist)
		if err != nil {
			logger.Error(err, "Failed to machmeta.ListAccessor", "sublist", sublist)
			return
		}
		rv, err := strconv.ParseInt(listMeta.GetResourceVersion(), 10, 64)
		if err != nil {
			logger.Error(err, "Failed to parse ResourceVersion of a List result")
			return
		}
		if rv > maxRV {
			maxRV = rv
			clusterOfMaxRV = clusterName
		}
		if firstListRV == nil {
			firstListRV = &rv
			allItems = append(allItems, subItems...)
			return
		}
		for _, item := range subItems {
			itemM := item.(metav1.Object)
			itemRV, err := strconv.ParseInt(itemM.GetResourceVersion(), 10, 64)
			if err != nil {
				logger.Error(err, "Failed to parse ResourceVersion of item", "item", item)
				continue
			}
			if itemRV <= *firstListRV {
				allItems = append(allItems, item)
			}
		}
		if listMeta.GetContinue() != "" {
			logger.Info("Warning: skipping list CONTINUE due to incomplete implementation", "gvk", clw.listGVK) // TODO: better
		}
	}
	clw.Lock()
	defer clw.Unlock()
	if clw.clusterListFirst != "" {
		if lwFirst, have := clw.perCluster[clw.clusterListFirst]; have {
			listCluster(clw.clusterListFirst, lwFirst)
		}
	}
	for clusterName, lwForCluster := range clw.perCluster {
		if clusterName == clw.clusterListFirst {
			continue
		}
		listCluster(clusterName, lwForCluster)
	}
	clw.clusterListFirst = clusterOfMaxRV // hueristic to lose minimize lossage next time
	if firstListRV == nil {
		var one int64 = 1
		firstListRV = &one
	}
	return &myList{
		TypeMeta: metav1.TypeMeta{
			APIVersion: clw.listGVK.GroupVersion().String(),
			Kind:       clw.listGVK.Kind,
		},
		ListMeta: metav1.ListMeta{
			ResourceVersion: strconv.FormatInt(*firstListRV, 10),
		},
		Items: allItems,
	}, nil
}

func (ml *myList) DeepCopyObject() runtime.Object {
	ans := myList{
		TypeMeta: ml.TypeMeta,
		ListMeta: *ml.ListMeta.DeepCopy(),
		Items:    make([]runtime.Object, len(ml.Items)),
	}
	for index, obj := range ml.Items {
		ans.Items[index] = obj.DeepCopyObject()
	}
	return &ans
}

func (clw *crossClusterListerWatcher[Scoped, ListType]) Watch(options metav1.ListOptions) (watch.Interface, error) {
	ctx := clw.ctx
	ctx, close := context.WithCancel(ctx)
	clw.Lock()
	defer clw.Unlock()
	ans := &myWatch[Scoped, ListType]{
		clw:          clw,
		close:        close,
		doneChan:     ctx.Done(),
		reconfigChan: clw.reconfigChan,
		filtered:     make(chan watch.Event),
	}
	for clusterName, lwForCluster := range clw.perCluster {
		clusterWatch, err := lwForCluster.scopedLW.Watch(clw.ctx, options)
		if err != nil {
			return nil, fmt.Errorf("Watch for cluster %s failed: %w", clusterName, err)
		}
		wpc := &watchPerCluster[Scoped, ListType]{
			myWatch:     ans,
			scopedWatch: clusterWatch,
			scopedChan:  clusterWatch.ResultChan(),
		}
		go wpc.Run()
	}
	return ans, nil
}

func (wpc *watchPerCluster[Scoped, ListType]) Run() {
	for {
		select {
		case <-wpc.doneChan:
			wpc.scopedWatch.Stop()
			return
		case <-wpc.reconfigChan:
			wpc.scopedWatch.Stop()
			return
		case event, ok := <-wpc.scopedChan:
			if !ok {
				wpc.close()
				return
			}
			wpc.filtered <- event
		}
	}
}

type myWatch[Scoped ScopedListerWatcher[ListType], ListType runtime.Object] struct {
	clw          *crossClusterListerWatcher[Scoped, ListType]
	close        func()
	doneChan     <-chan struct{}
	reconfigChan <-chan struct{} // the chan that was current when this watch started
	filtered     chan watch.Event
}

type watchPerCluster[Scoped ScopedListerWatcher[ListType], ListType runtime.Object] struct {
	*myWatch[Scoped, ListType]
	scopedWatch watch.Interface
	scopedChan  <-chan watch.Event
}

func (mw *myWatch[Scoped, ListType]) Stop() {
	mw.close()
}

func (mw *myWatch[Scoped, ListType]) ResultChan() <-chan watch.Event {
	return mw.filtered
}
