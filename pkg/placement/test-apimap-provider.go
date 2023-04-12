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
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"
)

func NewTestAPIMapProvider(logger klog.Logger) *TestAPIMapProvider {
	tap := &TestAPIMapProvider{
		logger:     logger,
		perCluster: NewMapMap[logicalcluster.Name, *TestAPIPerCluster](nil),
	}
	logger.V(2).Info("NewTestAPIMapProvider", "tap", fmt.Sprintf("%p", tap))
	return tap
}

// TestAPIMapProvider is a funky APIMapProvider for testing purposes.
// It is not fully compliant regarding locking.
// It exposes internals for test functions to examine and manipulate.
// A test func can modify a cluster's GroupInfo and/or ResourceInfo
// and the registered receivers will be synchronously updated, but this
// should only be done while there is no concurrent acess going on.
type TestAPIMapProvider struct {
	logger klog.Logger
	sync.Mutex
	perCluster MutableMap[logicalcluster.Name, *TestAPIPerCluster]
}

type TestAPIPerCluster struct {
	GroupInfo         MutableMap[string /*group name*/, APIGroupInfo]
	ResourceInfo      MutableMap[metav1.GroupResource, ResourceDetails]
	groupReceivers    MappingReceiverHolderFork[string /*group name*/, APIGroupInfo]
	resourceReceivers MappingReceiverHolderFork[metav1.GroupResource, ResourceDetails]
}

var _ APIMapProvider = &TestAPIMapProvider{}

func (tap *TestAPIMapProvider) AddReceivers(cluster logicalcluster.Name,
	groupReceiver *MappingReceiverHolder[string /*group name*/, APIGroupInfo],
	resourceReceiver *MappingReceiverHolder[metav1.GroupResource, ResourceDetails]) {
	tap.Lock()
	defer tap.Unlock()
	tpc := tap.ensureClusterLocked(cluster)
	MapApply[string /*group name*/, APIGroupInfo](tpc.GroupInfo, groupReceiver)
	MapApply[metav1.GroupResource, ResourceDetails](tpc.ResourceInfo, resourceReceiver)
	tpc.groupReceivers = append(tpc.groupReceivers, groupReceiver)
	tpc.resourceReceivers = append(tpc.resourceReceivers, resourceReceiver)
	tap.logger.V(2).Info("AddReceivers", "tap", fmt.Sprintf("%p", tap), "cluster", cluster, "tpc", fmt.Sprintf("%p", tpc))
}

func (tap *TestAPIMapProvider) ensureClusterLocked(cluster logicalcluster.Name) *TestAPIPerCluster {
	return MapGetAdd(tap.perCluster, cluster, true, func(cluster logicalcluster.Name) *TestAPIPerCluster {
		tpc := &TestAPIPerCluster{}
		tpc.GroupInfo = NewMapMap(MappingReceiverDiscardsPrevious(MappingReceiverFunc(
			func() MappingReceiver[string /*group name*/, APIGroupInfo] { return tpc.groupReceivers })))
		tpc.ResourceInfo = NewMapMap(MappingReceiverDiscardsPrevious(MappingReceiverFunc(
			func() MappingReceiver[metav1.GroupResource, ResourceDetails] { return tpc.resourceReceivers })))
		return tpc
	})
}

func (tap *TestAPIMapProvider) RemoveReceivers(cluster logicalcluster.Name,
	groupReceiver *MappingReceiverHolder[string /*group name*/, APIGroupInfo],
	resourceReceiver *MappingReceiverHolder[metav1.GroupResource, ResourceDetails]) {
	tap.Lock()
	defer tap.Unlock()
	tpc, has := tap.perCluster.Get(cluster)
	if !has {
		return
	}
	tpc.groupReceivers = SliceRemoveFunctional(tpc.groupReceivers, groupReceiver)
	tpc.resourceReceivers = SliceRemoveFunctional(tpc.resourceReceivers, resourceReceiver)
}

func (tap *TestAPIMapProvider) AsResourceReceiver() MappingReceiver[Pair[logicalcluster.Name, metav1.GroupResource], ResourceDetails] {
	return MappingReceiverFuncs[Pair[logicalcluster.Name, metav1.GroupResource], ResourceDetails]{
		OnPut: func(key Pair[logicalcluster.Name, metav1.GroupResource], val ResourceDetails) {
			tap.Lock()
			defer tap.Unlock()
			tpc := tap.ensureClusterLocked(key.First)
			tap.logger.V(2).Info("AsResourceReceiver.Put", "tap", fmt.Sprintf("%p", tap), "cluster", key.First, "tpc", fmt.Sprintf("%p", tpc), "numReceivers", len(tpc.resourceReceivers))
			tpc.ResourceInfo.Put(key.Second, val)
		},
		OnDelete: func(key Pair[logicalcluster.Name, metav1.GroupResource]) {
			tap.Lock()
			defer tap.Unlock()
			tpc := tap.ensureClusterLocked(key.First)
			tpc.ResourceInfo.Delete(key.Second)
		}}
}
