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

package controller

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/klog/v2"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

type _syncConfigManager struct {
	sync.Mutex
	logger                     klog.Logger
	syncConfigMap              map[string]edgev1alpha1.EdgeSyncConfig
	indexedDownSyncedResources _indexedSyncedResources
	indexedUpSyncedResources   _indexedSyncedResources
	conversions                []edgev1alpha1.EdgeSynConversion
}

type _indexedSyncedResources struct {
	syncedResources []edgev1alpha1.EdgeSyncConfigResource
	index           map[edgev1alpha1.EdgeSyncConfigResource]bool
}

func newIndexedSyncedResources(resources ...edgev1alpha1.EdgeSyncConfigResource) _indexedSyncedResources {
	isr := _indexedSyncedResources{
		syncedResources: []edgev1alpha1.EdgeSyncConfigResource{},
		index:           map[edgev1alpha1.EdgeSyncConfigResource]bool{},
	}
	for _, resource := range resources {
		_, ok := isr.index[resource]
		if !ok {
			isr.index[resource] = true
		}
	}
	isr.syncedResources = []edgev1alpha1.EdgeSyncConfigResource{}
	for key := range isr.index {
		isr.syncedResources = append(isr.syncedResources, key)
	}
	return isr
}

var syncConfigManager = _syncConfigManager{
	logger:                     klog.FromContext(context.TODO()),
	syncConfigMap:              map[string]edgev1alpha1.EdgeSyncConfig{},
	indexedDownSyncedResources: newIndexedSyncedResources(),
	indexedUpSyncedResources:   newIndexedSyncedResources(),
	conversions:                []edgev1alpha1.EdgeSynConversion{},
}

func (s *_syncConfigManager) upsert(syncConfig edgev1alpha1.EdgeSyncConfig) {
	key := syncConfig.Name
	s.logger.V(3).Info(fmt.Sprintf("upsert %s to synConfigMap", key))
	s.Lock()
	defer s.Unlock()
	s.syncConfigMap[key] = syncConfig
	s.refresh()
}

func (s *_syncConfigManager) delete(key string) {
	s.logger.V(3).Info(fmt.Sprintf("delete %s from synConfigMap", key))
	s.Lock()
	defer s.Unlock()
	delete(s.syncConfigMap, key)
	s.refresh()
}

func (s *_syncConfigManager) refresh() {
	downSyncedResources := []edgev1alpha1.EdgeSyncConfigResource{}
	upSyncedResources := []edgev1alpha1.EdgeSyncConfigResource{}
	conversions := []edgev1alpha1.EdgeSynConversion{}
	for _, _syncConfig := range s.syncConfigMap {
		downSyncedResources = append(downSyncedResources, _syncConfig.Spec.DownSyncedResources...)
		upSyncedResources = append(upSyncedResources, _syncConfig.Spec.UpSyncedResources...)
		conversions = append(conversions, _syncConfig.Spec.Conversions...)
	}
	s.indexedDownSyncedResources = newIndexedSyncedResources(downSyncedResources...)
	s.indexedUpSyncedResources = newIndexedSyncedResources(upSyncedResources...)
	s.conversions = conversions
	s.logger.V(3).Info("refreshed syncConfigManager")
}

func (s *_syncConfigManager) getDownSyncedResources() []edgev1alpha1.EdgeSyncConfigResource {
	return s.indexedDownSyncedResources.syncedResources
}

func (s *_syncConfigManager) getUpSyncedResources() []edgev1alpha1.EdgeSyncConfigResource {
	return s.indexedUpSyncedResources.syncedResources
}

func (s *_syncConfigManager) getConversions() []edgev1alpha1.EdgeSynConversion {
	return s.conversions
}

func GetDownSyncedResources() []edgev1alpha1.EdgeSyncConfigResource {
	return syncConfigManager.getDownSyncedResources()
}

func GetUpSyncedResources() []edgev1alpha1.EdgeSyncConfigResource {
	return syncConfigManager.getUpSyncedResources()
}

func GetConversions() []edgev1alpha1.EdgeSynConversion {
	return syncConfigManager.getConversions()
}
