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

package util

import (
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GvkGvrMapper is a thread-safe mapping between GVKs and GVRs.
// This mapper is only for top-level resources and there is an expectation that
// within this restricted scope, the relation between resource and Kind is 1:1.
type GvkGvrMapper interface {
	// Add adds a mapping between the given GVK and GVR.
	// This overwrites the GVK <-> GVR mapping if pre-existing.
	Add(gvk schema.GroupVersionKind, gvr schema.GroupVersionResource)
	// DeleteByGvkKey deletes the mapping associated with the given GVK.
	// The key format is that outputted by utils.KeyForGroupVersionKind.
	DeleteByGvkKey(string)
	// DeleteByGvrKey deletes the mapping associated with the given GVR.
	// The key format is that outputted by utils.KeyForGroupVersionResource.
	DeleteByGvrKey(string)
	// GetGvr returns the GVR associated with the given GVK if it exists.
	// The returned boolean indicates whether it exists or not.
	GetGvr(gvk schema.GroupVersionKind) (schema.GroupVersionResource, bool)
	// GetGvk returns the GVK associated with the given GVR if it exists.
	// The returned boolean indicates whether it exists or not.
	GetGvk(gvr schema.GroupVersionResource) (schema.GroupVersionKind, bool)
}

type gvkGvrMapper struct {
	sync.RWMutex
	gvkToGvr map[string]schema.GroupVersionResource
	gvrToGvk map[string]schema.GroupVersionKind
}

func NewGvkGvrMapper() GvkGvrMapper {
	return &gvkGvrMapper{
		gvkToGvr: make(map[string]schema.GroupVersionResource),
		gvrToGvk: make(map[string]schema.GroupVersionKind),
	}
}

func (m *gvkGvrMapper) Add(gvk schema.GroupVersionKind, gvr schema.GroupVersionResource) {
	m.Lock()
	defer m.Unlock()

	m.gvkToGvr[KeyForGroupVersionKind(gvk.Group, gvk.Version, gvk.Kind)] = gvr
	m.gvrToGvk[KeyForGroupVersionResource(gvr.Group, gvr.Version, gvr.Resource)] = gvk
}

func (m *gvkGvrMapper) DeleteByGvkKey(gvkKey string) {
	m.Lock()
	defer m.Unlock()

	gvr, found := m.gvkToGvr[gvkKey]
	if !found {
		return
	}

	delete(m.gvkToGvr, gvkKey)
	delete(m.gvrToGvk, KeyForGroupVersionResource(gvr.Group, gvr.Version, gvr.Resource))
}

func (m *gvkGvrMapper) DeleteByGvrKey(gvrKey string) {
	m.Lock()
	defer m.Unlock()

	gvk, found := m.gvrToGvk[gvrKey]
	if !found {
		return
	}

	delete(m.gvrToGvk, gvrKey)
	delete(m.gvkToGvr, KeyForGroupVersionKind(gvk.Group, gvk.Version, gvk.Kind))
}

func (m *gvkGvrMapper) GetGvr(gvk schema.GroupVersionKind) (schema.GroupVersionResource, bool) {
	m.RLock()
	defer m.RUnlock()

	gvr, ok := m.gvkToGvr[KeyForGroupVersionKind(gvk.Group, gvk.Version, gvk.Kind)]
	return gvr, ok
}

func (m *gvkGvrMapper) GetGvk(gvr schema.GroupVersionResource) (schema.GroupVersionKind, bool) {
	m.RLock()
	defer m.RUnlock()

	gvk, ok := m.gvrToGvk[KeyForGroupVersionResource(gvr.Group, gvr.Version, gvr.Resource)]
	return gvk, ok
}
