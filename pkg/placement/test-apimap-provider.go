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

	"github.com/kcp-dev/logicalcluster/v3"
)

func NewTestAPIMapProvider() *TestAPIMapProvider {
	return &TestAPIMapProvider{
		perCluster: NewMapMap[logicalcluster.Name, *TestAPIPerCluster](nil),
	}
}

// TestAPIMapProvider is a funky APIMapProvider for testing purposes.
// It is not fully compliant regarding locking.
// It exposes internals for test functions to examine and manipulate.
// A test func can modify a cluster's GroupInfo and/or ResourceInfo
// and the registered receivers will be synchronously updated, but this
// should only be done while there is no concurrent acess going on.
// RemoveReceivers will not be implemented until go 1.20 or later is
// required for this module.
type TestAPIMapProvider struct {
	sync.Mutex
	perCluster MutableMap[logicalcluster.Name, *TestAPIPerCluster]
}

type TestAPIPerCluster struct {
	GroupInfo         MutableMap[string /*group name*/, APIGroupInfo]
	ResourceInfo      MutableMap[metav1.GroupResource, ResourceDetails]
	groupReceivers    MappingReceiverFork[string /*group name*/, APIGroupInfo]
	resourceReceivers MappingReceiverFork[metav1.GroupResource, ResourceDetails]
}

var _ APIMapProvider = &TestAPIMapProvider{}

func (tap *TestAPIMapProvider) AddReceivers(cluster logicalcluster.Name,
	groupReceiver MappingReceiver[string /*group name*/, APIGroupInfo],
	resourceReceiver MappingReceiver[metav1.GroupResource, ResourceDetails]) {
	tap.Lock()
	defer tap.Unlock()
	tpc := MapGetAdd(tap.perCluster, cluster, true, func(cluster logicalcluster.Name) *TestAPIPerCluster {
		groupReceivers := MappingReceiverFork[string /*group name*/, APIGroupInfo]{}
		resourceReceivers := MappingReceiverFork[metav1.GroupResource, ResourceDetails]{}
		tpc := &TestAPIPerCluster{
			GroupInfo:         NewMapMap[string /*group name*/, APIGroupInfo](MappingReceiverDiscardsPrevious[string /*group name*/, APIGroupInfo](groupReceivers)),
			ResourceInfo:      NewMapMap[metav1.GroupResource, ResourceDetails](MappingReceiverDiscardsPrevious[metav1.GroupResource, ResourceDetails](resourceReceivers)),
			groupReceivers:    groupReceivers,
			resourceReceivers: resourceReceivers,
		}
		return tpc
	})
	MapApply[string /*group name*/, APIGroupInfo](tpc.GroupInfo, groupReceiver)
	MapApply[metav1.GroupResource, ResourceDetails](tpc.ResourceInfo, resourceReceiver)
	tpc.groupReceivers = append(tpc.groupReceivers, groupReceiver)
	tpc.resourceReceivers = append(tpc.resourceReceivers, resourceReceiver)
}

func (tap *TestAPIMapProvider) RemoveReceivers(cluster logicalcluster.Name,
	groupReceiver MappingReceiver[string /*group name*/, APIGroupInfo],
	resourceReceiver MappingReceiver[metav1.GroupResource, ResourceDetails]) {
	tap.Lock()
	defer tap.Unlock()
	tpc, has := tap.perCluster.Get(cluster)
	if !has {
		return
	} else {
		// The following statement is only here to stop linters from complaining that tpc is unused.
		// Remove this statement once the panic is removed.
		tpc.groupReceivers = MappingReceiverFork[string /*group name*/, APIGroupInfo]{}
	}
	panic("not implemented until go 1.20 is required for this module")
	// The following requires go 1.20:
	// tpc.groupReceivers = SliceRemoveFunctional(tpc.groupReceivers, groupReceiver)
	// tpc.resourceReceivers = SliceRemoveFunctional(tpc.resourceReceivers, resourceReceiver)
}
