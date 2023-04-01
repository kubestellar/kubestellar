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

package placement

import (
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8ssets "k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func (pk ProjectionKey) FromPair(part WorkloadPart, where edgeapi.SinglePlacement) ProjectionKey {
	return ProjectionKey{part.GroupResource(), where}
}

func (partID WorkloadPartID) GroupResource() metav1.GroupResource {
	return metav1.GroupResource{
		Group:    partID.APIGroup,
		Resource: partID.Resource,
	}
}

func SimpleBindingOrganizer(logger klog.Logger) BindingOrganizer {
	return func(discovery APIMapProvider, resourceModes ResourceModes, eventHandler EventHandler) (SingleBinder, ProjectionMapProvider) {
		ans := &simpleBindingOrganizer{
			logger:                logger,
			discovery:             discovery,
			resourceModes:         resourceModes,
			eventHandler:          eventHandler,
			projectionMapProvider: NewRelayAndProjectMap[ProjectionKey, *projectionPerClusterImpl, *ProjectionPerCluster](false, exportProjectionPerCluster),
		}
		return ans, ans.projectionMapProvider
	}
}

// simpleBindingOrganizer is the top-level data structure of the organizer.
// In the locking order it precedes its discovery and its projectionMapProvider,
// which in turn precedes each projectionPerClusterImpl.
type simpleBindingOrganizer struct {
	logger                klog.Logger
	discovery             APIMapProvider
	resourceModes         ResourceModes
	eventHandler          EventHandler
	projectionMapProvider TransformingRelayMap[ProjectionKey, *projectionPerClusterImpl, *ProjectionPerCluster]
	sync.Mutex
}

type projectionPerClusterImpl struct {
	organizer *simpleBindingOrganizer
	ProjectionPerCluster
	perSourceCluster RelayMap[logicalcluster.Name, ProjectionDetails]
	sync.Mutex
	apiProvider ScopedAPIProvider
}

func exportProjectionPerCluster(impl *projectionPerClusterImpl) *ProjectionPerCluster {
	return &impl.ProjectionPerCluster
}

func (pc *projectionPerClusterImpl) SetProvider(apiProvider ScopedAPIProvider) {
	pc.Lock()
	defer pc.Unlock()
	pc.organizer.logger.V(2).Info("Got ScopedAPIProvider")
	pc.apiProvider = apiProvider
}

func (sbo *simpleBindingOrganizer) Transact(xn func(SingleBindingOps)) {
	sbo.Lock()
	defer sbo.Unlock()
	sbo.logger.V(3).Info("Begin transaction")
	xn(sboXnOps{sbo})
	sbo.logger.V(3).Info("End transaction")
}

// sboXnOps exposes the Add and Remove methods only in the locked context of a transaction
type sboXnOps struct{ sbo *simpleBindingOrganizer }

func (sxo sboXnOps) Add(pair Pair[WorkloadPart, edgeapi.SinglePlacement]) bool {
	sbo := sxo.sbo
	pk := ProjectionKey{}.FromPair(pair.First, pair.Second)
	if mgrIsNamespace(pk.GroupResource) && !pair.First.IncludeNamespaceObject {
		// In this case what is needed is to make sbo.projectionMapProvider say
		// to downsync all the (namespaced) objects in this namespace.
		// TODO: that
		sbo.logger.V(3).Info("Not implemented: atomic pair about namespace contents", "workloadPart", pair.First, "sps", pair.Second)
		return false
	}
	cluster := logicalcluster.Name(pair.Second.Cluster)
	pc := sbo.projectionMapProvider.OuterGet(pk)
	if pc == nil {
		perSourceCluster := NewRelayMap[logicalcluster.Name, ProjectionDetails](true)
		pc = &projectionPerClusterImpl{
			organizer: sbo,
			ProjectionPerCluster: ProjectionPerCluster{
				APIVersion:       pair.First.APIVersion,
				PerSourceCluster: perSourceCluster,
			},
			perSourceCluster: perSourceCluster,
		}
		sbo.logger.V(2).Info("Adding ProjectionPerCluster", "workloadPart", pair.First, "sps", pair.Second)
		sbo.discovery.AddClient(cluster, pc)
		sbo.projectionMapProvider.Put(pk, pc)
	}
	pd := pc.perSourceCluster.OuterGet(cluster)
	var change bool
	if pd.Names == nil {
		pd.Names = ToHeap[k8ssets.String](k8ssets.NewString(pair.First.Name))
		change = true
	} else if !pd.Names.Has(pair.First.Name) {
		change = true
		pd.Names.Insert(pair.First.Name)
	}
	if change {
		sbo.logger.V(2).Info("Passing along addition", "workloadPart", pair.First, "sps", pair.Second, "pk", pk, "cluster", cluster, "pd", pd)
		pc.perSourceCluster.Put(cluster, pd)
	} else {
		sbo.logger.V(2).Info("No news in addition", "workloadPart", pair.First, "sps", pair.Second, "pk", pk, "cluster", cluster, "pd", pd)

	}
	return change
}

func (sxo sboXnOps) Remove(pair Pair[WorkloadPart, edgeapi.SinglePlacement]) bool {
	sbo := sxo.sbo
	pk := ProjectionKey{}.FromPair(pair.First, pair.Second)
	if mgrIsNamespace(pk.GroupResource) && !pair.First.IncludeNamespaceObject {
		// In this case what is needed is to make sbo.projectionMapProvider stop saying
		// to downsync all the (namespaced) objects in this namespace.
		// TODO: that
		sbo.logger.V(3).Info("Not implemented: atomic pair about namespace contents", "workloadPart", pair.First, "sps", pair.Second)
		return false
	}
	cluster := logicalcluster.Name(pair.Second.Cluster)
	pc := sbo.projectionMapProvider.OuterGet(pk)
	if pc == nil {
		sbo.logger.V(2).Info("No cluster data", "workloadPart", pair.First, "sps", pair.Second, "pk", pk, "cluster", cluster)
		return false
	}
	pd := pc.perSourceCluster.OuterGet(cluster)
	if pd.Names == nil || !pd.Names.Has(pair.First.Name) {
		sbo.logger.V(2).Info("Norhing to remove", "workloadPart", pair.First, "sps", pair.Second, "pk", pk, "cluster", cluster, "pd", pd)
		return false
	}
	sbo.logger.V(2).Info("Removing internal", "workloadPart", pair.First, "sps", pair.Second, "pk", pk, "cluster", cluster, "pd", pd)
	pd.Names.Delete(pair.First.Name)
	pc.perSourceCluster.Put(cluster, pd)
	return true
}

func mgrIsNamespace(gr metav1.GroupResource) bool {
	return gr.Group == "" && gr.Resource == "namespaces"
}
