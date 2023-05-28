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

package controller

import (
	"fmt"
	"sync"

	"k8s.io/klog/v2"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

type SyncConfigManager struct {
	sync.Mutex
	logger                       klog.Logger
	syncConfigMap                map[string]edgev1alpha1.EdgeSyncConfig
	indexedDownSyncedResources   _indexedSyncedResources
	indexedUpSyncedResources     _indexedSyncedResources
	indexedDownUnsyncedResources _indexedSyncedResources
	indexedUpUnsyncedResources   _indexedSyncedResources
	conversions                  []edgev1alpha1.EdgeSynConversion
}

type _indexedSyncedResources struct {
	syncedResources []edgev1alpha1.EdgeSyncConfigResource
	index           map[edgev1alpha1.EdgeSyncConfigResource]bool
}

func createIndexedSyncedResources(resources ...edgev1alpha1.EdgeSyncConfigResource) _indexedSyncedResources {
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

func NewSyncConfigManager(logger klog.Logger) *SyncConfigManager {
	return &SyncConfigManager{
		logger:                     logger,
		syncConfigMap:              map[string]edgev1alpha1.EdgeSyncConfig{},
		indexedDownSyncedResources: createIndexedSyncedResources(),
		indexedUpSyncedResources:   createIndexedSyncedResources(),
		conversions:                []edgev1alpha1.EdgeSynConversion{},
	}
}

func (s *SyncConfigManager) upsert(syncConfig edgev1alpha1.EdgeSyncConfig) {
	key := syncConfig.Name
	s.logger.V(3).Info(fmt.Sprintf("upsert %s to synConfigMap", key))
	s.Lock()
	defer s.Unlock()
	newSyncConfig := copySyncConfigMap(s.syncConfigMap)
	newSyncConfig[key] = syncConfig
	s.refresh(newSyncConfig)
	s.syncConfigMap = newSyncConfig
}

func (s *SyncConfigManager) delete(key string) {
	s.logger.V(3).Info(fmt.Sprintf("delete %s from synConfigMap", key))
	s.Lock()
	defer s.Unlock()
	newSyncConfig := copySyncConfigMap(s.syncConfigMap)
	delete(newSyncConfig, key)
	s.refresh(newSyncConfig)
	s.syncConfigMap = newSyncConfig
}

func (s *SyncConfigManager) refresh(newSyncConfig map[string]edgev1alpha1.EdgeSyncConfig) {
	conversions := []edgev1alpha1.EdgeSynConversion{}
	for _, _syncConfig := range s.syncConfigMap {
		conversions = append(conversions, _syncConfig.Spec.Conversions...)
	}
	s.conversions = conversions

	currentIndexedDownSyncedResources, currentIndexedUpSyncedResources := createIndexedDownAndUpSyncedResources(s.syncConfigMap)
	newIndexedDownSyncedResources, newIndexedUpSyncedResources := createIndexedDownAndUpSyncedResources(newSyncConfig)

	s.indexedDownSyncedResources = newIndexedDownSyncedResources
	s.indexedDownUnsyncedResources = updateUnsyncedResources(s.indexedDownUnsyncedResources, currentIndexedDownSyncedResources, newIndexedDownSyncedResources)

	s.indexedUpSyncedResources = newIndexedUpSyncedResources
	s.indexedUpUnsyncedResources = updateUnsyncedResources(s.indexedUpUnsyncedResources, currentIndexedUpSyncedResources, newIndexedUpSyncedResources)

	s.logger.V(3).Info("refreshed syncConfigManager")
}

func createIndexedDownAndUpSyncedResources(syncConfigMap map[string]edgev1alpha1.EdgeSyncConfig) (_indexedSyncedResources, _indexedSyncedResources) {
	downSyncedResources := []edgev1alpha1.EdgeSyncConfigResource{}
	upSyncedResources := []edgev1alpha1.EdgeSyncConfigResource{}
	for _, _syncConfig := range syncConfigMap {
		downSyncedResources = append(downSyncedResources, _syncConfig.Spec.DownSyncedResources...)
		upSyncedResources = append(upSyncedResources, _syncConfig.Spec.UpSyncedResources...)
	}
	return createIndexedSyncedResources(downSyncedResources...), createIndexedSyncedResources(upSyncedResources...)
}

func updateUnsyncedResources(currentIndexedUnsyncedResources _indexedSyncedResources, currentIndexedSyncedResources _indexedSyncedResources, newIndexedSyncedResources _indexedSyncedResources) _indexedSyncedResources {
	unsyncedResources := []edgev1alpha1.EdgeSyncConfigResource{}
	// Current unsynced resources (exclude synced resources listed in new SyncConfig)
	for key := range currentIndexedUnsyncedResources.index {
		_, ok := newIndexedSyncedResources.index[key]
		if !ok {
			unsyncedResources = append(unsyncedResources, key)
		}
	}
	// If the resource exists in current SyncConfig but doesn't exist in new one, then add it to the unsyncedResources
	for key := range currentIndexedSyncedResources.index {
		_, ok := newIndexedSyncedResources.index[key]
		if !ok {
			unsyncedResources = append(unsyncedResources, key)
		}
	}

	return createIndexedSyncedResources(unsyncedResources...)
}

func copySyncConfigMap(syncConfigMap map[string]edgev1alpha1.EdgeSyncConfig) map[string]edgev1alpha1.EdgeSyncConfig {
	_syncConfigMap := map[string]edgev1alpha1.EdgeSyncConfig{}
	for key, value := range syncConfigMap {
		_syncConfigMap[key] = value
	}
	return _syncConfigMap
}

func (s *SyncConfigManager) GetDownSyncedResources() []edgev1alpha1.EdgeSyncConfigResource {
	return s.indexedDownSyncedResources.syncedResources
}

func (s *SyncConfigManager) GetUpSyncedResources() []edgev1alpha1.EdgeSyncConfigResource {
	return s.indexedUpSyncedResources.syncedResources
}

func (s *SyncConfigManager) GetDownUnsyncedResources() []edgev1alpha1.EdgeSyncConfigResource {
	return s.indexedDownUnsyncedResources.syncedResources
}

func (s *SyncConfigManager) GetUpUnsyncedResources() []edgev1alpha1.EdgeSyncConfigResource {
	return s.indexedUpUnsyncedResources.syncedResources
}

func (s *SyncConfigManager) GetConversions() []edgev1alpha1.EdgeSynConversion {
	return s.conversions
}
