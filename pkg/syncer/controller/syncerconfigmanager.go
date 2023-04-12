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
	"fmt"
	"sync"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	syncerv1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/syncer/clientfactory"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/restmapper"
	"k8s.io/klog/v2"
)

const (
	DOWNSYNC_NAMESPACED_SUFFIX    string = "_downsync_namespaced"
	DOWNSYNC_CLUSTERSCOPED_SUFFIX string = "_downsync_clusterscoped"
	UPSYNC_SUFFIX                 string = "_upsync"
)

func NewSyncerConfigManager(logger klog.Logger, syncConfigManager *SyncConfigManager, upstreamClientFactory clientfactory.ClientFactory, downstreamClientFactory clientfactory.ClientFactory) *SyncerConfigManager {
	return &SyncerConfigManager{
		logger:                  logger,
		syncConfigManager:       syncConfigManager,
		syncerConfigMap:         map[string]edgev1alpha1.SyncerConfig{},
		upstreamClientFactory:   upstreamClientFactory,
		downstreamClientFactory: downstreamClientFactory,
	}
}

type SyncerConfigManager struct {
	sync.Mutex
	logger                  klog.Logger
	syncConfigManager       *SyncConfigManager
	syncerConfigMap         map[string]edgev1alpha1.SyncerConfig
	upstreamClientFactory   clientfactory.ClientFactory
	downstreamClientFactory clientfactory.ClientFactory
}

func (s *SyncerConfigManager) upsert(syncerConfig edgev1alpha1.SyncerConfig) {
	logger := s.logger.WithValues("syncerConfigName", syncerConfig.Name)
	s.Lock()
	defer s.Unlock()
	s.syncerConfigMap[syncerConfig.Name] = syncerConfig
	logger.V(3).Info("upsert syncerConfig")
}

func (s *SyncerConfigManager) Refresh() {
	s.Lock()
	defer s.Unlock()
	for _, syncerConfig := range s.syncerConfigMap {
		logger := s.logger.WithValues("syncerConfigName", syncerConfig.Name)
		logger.V(3).Info("upsert syncerConfig to syncConfigManager stores")
		upstreamGroupResourcesList, err := s.upstreamClientFactory.GetAPIGroupResources()
		if err != nil {
			logger.Error(err, "Failed to get API Group resources from upstream. Skip upsert operation")
			return
		}
		downstreamGroupResourcesList, err := s.downstreamClientFactory.GetAPIGroupResources()
		if err != nil {
			logger.Error(err, "Failed to get API Group resources from downstream. Skip upsert operation")
			return
		}
		s.upsertNamespaceScoped(syncerConfig, upstreamGroupResourcesList)
		s.upsertClusterScoped(syncerConfig, upstreamGroupResourcesList)
		s.upsertUpsync(syncerConfig, downstreamGroupResourcesList)
	}
}

func (s *SyncerConfigManager) upsertNamespaceScoped(syncerConfig edgev1alpha1.SyncerConfig, upstreamGroupResourcesList []*restmapper.APIGroupResources) {
	s.logger.V(3).Info(fmt.Sprintf("upsert namespace scoped resources as syncerConfig %s to syncConfigManager stores", syncerConfig.Name))
	edgeSyncConfigResources := []syncerv1alpha1.EdgeSyncConfigResource{}
	for _, namespace := range syncerConfig.Spec.NamespaceScope.Namespaces {
		edgeSyncConfigResourceForNamespace := syncerv1alpha1.EdgeSyncConfigResource{
			Group:   "",
			Version: "v1",
			Kind:    "Namespace",
			Name:    namespace,
		}
		edgeSyncConfigResources = append(edgeSyncConfigResources, edgeSyncConfigResourceForNamespace)
		for _, syncerConfigResource := range syncerConfig.Spec.NamespaceScope.Resources {
			group := syncerConfigResource.Group
			version := syncerConfigResource.APIVersion
			resource := syncerConfigResource.Resource
			versionedResources := findVersionedResourcesByGVR(group, version, resource, upstreamGroupResourcesList, s.logger)
			for _, versionedResource := range versionedResources {
				edgeSyncConfigResource := syncerv1alpha1.EdgeSyncConfigResource{
					Group:     group,
					Version:   version,
					Kind:      versionedResource.Kind,
					Namespace: namespace,
					Name:      "*",
				}
				edgeSyncConfigResources = append(edgeSyncConfigResources, edgeSyncConfigResource)
			}
		}
	}
	edgeSyncConfig := syncerv1alpha1.EdgeSyncConfig{
		ObjectMeta: v1.ObjectMeta{
			Name: syncerConfig.Name + DOWNSYNC_NAMESPACED_SUFFIX,
		},
		Spec: syncerv1alpha1.EdgeSyncConfigSpec{
			DownSyncedResources: edgeSyncConfigResources,
		},
	}
	s.syncConfigManager.upsert(edgeSyncConfig)
}

func (s *SyncerConfigManager) upsertClusterScoped(syncerConfig edgev1alpha1.SyncerConfig, upstreamGroupResourcesList []*restmapper.APIGroupResources) {
	s.logger.V(3).Info(fmt.Sprintf("upsert clusterscoped resources as syncerConfig %s to syncConfigManager stores", syncerConfig.Name))
	edgeSyncConfigResources := []syncerv1alpha1.EdgeSyncConfigResource{}
	for _, clusterScope := range syncerConfig.Spec.ClusterScope {
		group := clusterScope.Group
		version := clusterScope.APIVersion
		resource := clusterScope.Resource
		objects := clusterScope.Objects
		if objects != nil && len(objects) == 0 {
			// empty list means nothing to downsync (see ClusterScopeDownsyncResource definition)
			continue
		}
		versionedResources := findVersionedResourcesByGVR(group, version, resource, upstreamGroupResourcesList, s.logger)
		for _, versionedResource := range versionedResources {
			if objects == nil {
				edgeSyncConfigResource := syncerv1alpha1.EdgeSyncConfigResource{
					Group:   group,
					Version: version,
					Kind:    versionedResource.Kind,
					Name:    "*",
				}
				edgeSyncConfigResources = append(edgeSyncConfigResources, edgeSyncConfigResource)
			} else {
				for _, object := range objects {
					edgeSyncConfigResource := syncerv1alpha1.EdgeSyncConfigResource{
						Group:   group,
						Version: version,
						Kind:    versionedResource.Kind,
						Name:    object,
					}
					edgeSyncConfigResources = append(edgeSyncConfigResources, edgeSyncConfigResource)
				}
			}
		}
	}
	edgeSyncConfig := syncerv1alpha1.EdgeSyncConfig{
		ObjectMeta: v1.ObjectMeta{
			Name: syncerConfig.Name + DOWNSYNC_CLUSTERSCOPED_SUFFIX,
		},
		Spec: syncerv1alpha1.EdgeSyncConfigSpec{
			DownSyncedResources: edgeSyncConfigResources,
		},
	}
	s.syncConfigManager.upsert(edgeSyncConfig)
}

func (s *SyncerConfigManager) upsertUpsync(syncerConfig edgev1alpha1.SyncerConfig, downstreamGroupResourcesList []*restmapper.APIGroupResources) {
	s.logger.V(3).Info(fmt.Sprintf("upsert clusterscoped resources as syncerConfig %s to syncConfigManager stores", syncerConfig.Name))
	edgeSyncConfigResources := []syncerv1alpha1.EdgeSyncConfigResource{}
	for _, upsync := range syncerConfig.Spec.Upsync {
		group := upsync.APIGroup
		resources := upsync.Resources
		namespaces := upsync.Namespaces
		names := upsync.Names
		for _, resource := range resources {
			versionedResources := findVersionedResourcesByGV(group, resource, downstreamGroupResourcesList, s.logger)
			for _, versionedResource := range versionedResources {
				edgeSyncConfigResource := syncerv1alpha1.EdgeSyncConfigResource{
					Group:   group,
					Version: versionedResource.Version,
					Kind:    versionedResource.Kind,
				}
				if versionedResource.Namespaced {
					for _, namespace := range namespaces {
						for _, name := range names {
							edgeSyncConfigResource.Namespace = namespace
							edgeSyncConfigResource.Name = name
							edgeSyncConfigResources = append(edgeSyncConfigResources, edgeSyncConfigResource)
						}
					}
				} else {
					for _, name := range names {
						edgeSyncConfigResource.Name = name
						edgeSyncConfigResources = append(edgeSyncConfigResources, edgeSyncConfigResource)
					}
				}
			}
		}
	}
	edgeSyncConfig := syncerv1alpha1.EdgeSyncConfig{
		ObjectMeta: v1.ObjectMeta{
			Name: syncerConfig.Name + UPSYNC_SUFFIX,
		},
		Spec: syncerv1alpha1.EdgeSyncConfigSpec{
			UpSyncedResources: edgeSyncConfigResources,
		},
	}
	s.syncConfigManager.upsert(edgeSyncConfig)
}

func (s *SyncerConfigManager) delete(key string) {
	logger := s.logger.WithValues("syncerConfigName", key)
	logger.V(3).Info("delete syncConfigs for syncerConfig from syncConfigManager stores")
	s.Lock()
	defer s.Unlock()
	delete(s.syncerConfigMap, key)
	s.syncConfigManager.delete(key + DOWNSYNC_NAMESPACED_SUFFIX)
	s.syncConfigManager.delete(key + DOWNSYNC_CLUSTERSCOPED_SUFFIX)
	s.syncConfigManager.delete(key + UPSYNC_SUFFIX)
}

func findVersionedResourcesByGVR(group string, version string, resource string, apiGroupResourcesList []*restmapper.APIGroupResources, logger klog.Logger) []v1.APIResource {
	_versionedResources := []v1.APIResource{}
	var apiGroupResources *restmapper.APIGroupResources
	for _, groupResources := range apiGroupResourcesList {
		if groupResources.Group.Name == group {
			apiGroupResources = groupResources
			break
		}
	}
	if apiGroupResources == nil {
		return _versionedResources
	}
	versionedResources := apiGroupResources.VersionedResources[version]
	for _, versionedResource := range versionedResources {
		if resource == versionedResource.Name {
			_versionedResources = append(_versionedResources, versionedResource)
		} else if resource == "*" {
			_versionedResources = append(_versionedResources, versionedResource)
		}
	}
	return _versionedResources
}

func findVersionedResourcesByGV(group string, resource string, apiGroupResourcesList []*restmapper.APIGroupResources, logger klog.Logger) []v1.APIResource {
	_versionedResources := []v1.APIResource{}
	var apiGroupResources *restmapper.APIGroupResources
	for _, groupResources := range apiGroupResourcesList {
		if groupResources.Group.Name == group {
			apiGroupResources = groupResources
			break
		}
	}
	if apiGroupResources == nil {
		return _versionedResources
	}
	for version, versionedResources := range apiGroupResources.VersionedResources {
		for _, versionedResource := range versionedResources {
			if versionedResource.Version == "" {
				versionedResource.Version = version
			}
			if resource == versionedResource.Name {
				_versionedResources = append(_versionedResources, versionedResource)
			} else if resource == "*" {
				_versionedResources = append(_versionedResources, versionedResource)
			}
		}
	}
	return _versionedResources
}
