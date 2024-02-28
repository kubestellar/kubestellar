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
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVKToGVRMapper is a thread-safe mapping between GVKs -> GVRs.
// This mapper is only for top-level resources and there is an expectation that
// within this restricted scope, the relation between resource and Kind is 1:1.
type GVKToGVRMapper interface {
	// Add adds a mapping between the given GVK and GVR.
	// This overwrites the mapping if exists.
	Add(gvk schema.GroupVersionKind, gvr schema.GroupVersionResource)
	// Delete deletes the mapping associated with the given GVK.
	Delete(gvk schema.GroupVersionKind)
	// GetGvr returns the GVR associated with the given GVK if it exists.
	// The returned boolean indicates whether it exists or not.
	GetGvr(gvk schema.GroupVersionKind) (schema.GroupVersionResource, bool)
}

type gvkToGvrMapper struct {
	sync.RWMutex
	gvkToGvr map[schema.GroupVersionKind]schema.GroupVersionResource
}

func NewGvkGvrMapper() GVKToGVRMapper {
	return &gvkToGvrMapper{
		gvkToGvr: make(map[schema.GroupVersionKind]schema.GroupVersionResource),
	}
}

func (m *gvkToGvrMapper) Add(gvk schema.GroupVersionKind, gvr schema.GroupVersionResource) {
	m.Lock()
	defer m.Unlock()

	m.gvkToGvr[gvk] = gvr
}

func (m *gvkToGvrMapper) Delete(gvk schema.GroupVersionKind) {
	m.Lock()
	defer m.Unlock()

	delete(m.gvkToGvr, gvk)
}

func (m *gvkToGvrMapper) GetGvr(gvk schema.GroupVersionKind) (schema.GroupVersionResource, bool) {
	m.RLock()
	defer m.RUnlock()

	gvr, ok := m.gvkToGvr[gvk]
	return gvr, ok
}

// KeyForGroupVersionKind creates a string key in the form group/version/Kind
// or version/Kind if the group is empty.
func KeyForGroupVersionKind(group, version, kind string) string {
	if group == "" {
		return fmt.Sprintf("%s/%s", version, kind)
	}

	return fmt.Sprintf("%s/%s/%s", group, version, kind)
}
