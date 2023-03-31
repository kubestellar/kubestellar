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

type simpleBindingOrganizer struct {
	logger klog.Logger
	sync.Mutex
	discovery             APIMapProvider
	resourceModes         ResourceModes
	eventHandler          EventHandler
	projectionMapProvider TransformingRelayMap[ProjectionKey, *projectionPerClusterImpl, *ProjectionPerCluster]
}

type projectionPerClusterImpl struct {
	ProjectionPerCluster
	apiProvider      ScopedAPIProvider
	perSourceCluster RelayMap[logicalcluster.Name, ProjectionDetails]
}

func exportProjectionPerCluster(impl *projectionPerClusterImpl) *ProjectionPerCluster {
	return &impl.ProjectionPerCluster
}

func (pc *projectionPerClusterImpl) SetProvider(apiProvider ScopedAPIProvider) {
	pc.apiProvider = apiProvider
}

func (sbo *simpleBindingOrganizer) Transact(xn func(SingleBindingOps)) {
	sbo.Lock()
	defer sbo.Unlock()
	xn(sbo)
}

func (sbo *simpleBindingOrganizer) Add(pair Pair[WorkloadPart, edgeapi.SinglePlacement]) bool {
	pk := ProjectionKey{}.FromPair(pair.First, pair.Second)
	if mgrIsNamespace(pk.GroupResource) && !pair.First.IncludeNamespaceObject {
		// In this case what is needed is to make sbo.projectionMapProvider say
		// to downsync all the (namespaced) objects in this namespace.
		// TODO: that
		return false
	}
	cluster := logicalcluster.Name(pair.Second.Cluster)
	pc := sbo.projectionMapProvider.OuterGet(pk)
	if pc == nil {
		perSourceCluster := NewRelayMap[logicalcluster.Name, ProjectionDetails](true)
		pc = &projectionPerClusterImpl{
			ProjectionPerCluster: ProjectionPerCluster{
				APIVersion:       pair.First.APIVersion,
				PerSourceCluster: perSourceCluster,
			},
			perSourceCluster: perSourceCluster,
		}
		sbo.discovery.AddClient(cluster, pc)
		sbo.projectionMapProvider.Receive(pk, pc)
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
		pc.perSourceCluster.Receive(cluster, pd)
	}
	return change
}

func (sbo *simpleBindingOrganizer) Remove(pair Pair[WorkloadPart, edgeapi.SinglePlacement]) bool {
	pk := ProjectionKey{}.FromPair(pair.First, pair.Second)
	if mgrIsNamespace(pk.GroupResource) && !pair.First.IncludeNamespaceObject {
		// In this case what is needed is to make sbo.projectionMapProvider stop saying
		// to downsync all the (namespaced) objects in this namespace.
		// TODO: that
		return false
	}
	cluster := logicalcluster.Name(pair.Second.Cluster)
	pc := sbo.projectionMapProvider.OuterGet(pk)
	if pc == nil {
		return false
	}
	pd := pc.perSourceCluster.OuterGet(cluster)
	if pd.Names == nil || !pd.Names.Has(pair.First.Name) {
		return false
	}
	pd.Names.Delete(pair.First.Name)
	pc.perSourceCluster.Receive(cluster, pd)
	return true
}

func mgrIsNamespace(gr metav1.GroupResource) bool {
	return gr.Group == "" && gr.Resource == "namespaces"
}
