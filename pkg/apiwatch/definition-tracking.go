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

package apiwatch

import (
	"encoding/json"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (rlw *resourcesListWatcher) setDefinerLocked(oid objectID, enumr ResourceDefinitionEnumerator) {
	oldRscs := ensureMap(rlw.definerToRscs[oid])
	newRscs := GoSet[metav1.GroupVersionResource]{}
	rlw.logger.V(4).Info("Start setDefinerLocked", "oid", oid, "oldRscs", oldRscs)
	enumr(func(gvr metav1.GroupVersionResource) {
		newRscs[gvr] = Empty{}
		if _, had := oldRscs[gvr]; !had {
			definers := ensureMap(rlw.rscToDefiners[gvr])
			definers[oid] = Empty{}
			rlw.rscToDefiners[gvr] = definers
			rlw.logger.V(4).Info("Adding definition", "gvr", gvr, "oid", oid)
		}
	})
	for oldRsc := range oldRscs {
		if _, have := newRscs[oldRsc]; !have {
			definers := rlw.rscToDefiners[oldRsc]
			rlw.logger.V(4).Info("Removing definition", "oldRsc", oldRsc, "oid", oid)
			delete(definers, oid)
			if len(definers) == 0 {
				delete(rlw.rscToDefiners, oldRsc)
				rlw.logger.V(4).Info("No more definers", "oldRsc", oldRsc)
			} else {
				rlw.rscToDefiners[oldRsc] = definers
			}
		}
	}
	rlw.definerToRscs[oid] = newRscs
	rlw.logger.V(4).Info("Finish setDefinerLocked", "oid", oid, "newRscs", newRscs)
}

func ensureMap[Key comparable](in GoSet[Key]) GoSet[Key] {
	if in != nil {
		return in
	}
	return GoSet[Key]{}
}

func MarshalMap[Key comparable, Val any](it map[Key]Val) ([]byte, error) {
	if it == nil {
		return []byte("null"), nil
	}
	var builder strings.Builder
	enc := json.NewEncoder(&builder)
	builder.WriteRune('[')
	first := true
	for key, val := range it {
		item := struct{ Key, Val any }{key, val}
		if first {
			first = false
		} else {
			builder.WriteString(", ")
		}
		err := enc.Encode(item)
		if err != nil {
			errS := err.Error()
			enc.Encode(errS)
		}
	}
	builder.WriteRune(']')
	return []byte(builder.String()), nil
}

func MarshalSet[Key comparable](it map[Key]Empty) ([]byte, error) {
	if it == nil {
		return []byte("null"), nil
	}
	var builder strings.Builder
	enc := json.NewEncoder(&builder)
	builder.WriteRune('[')
	first := true
	for key := range it {
		if first {
			first = false
		} else {
			builder.WriteString(", ")
		}
		err := enc.Encode(key)
		if err != nil {
			errS := err.Error()
			enc.Encode(errS)
		}
	}
	builder.WriteRune(']')
	return []byte(builder.String()), nil
}
