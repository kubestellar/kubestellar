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

package kbuser

import (
	"context"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
)

const KubeBindLabelName string = "kubestellar.io/kube-bind-id"

const CMNamePrefix string = "kbmap-"

// KubeBindSpaceRelation is specific to one kube-bind service provider
// and reveals the 1:1 relation between the kube-bind identifier for
// a consumer (i.e., the name of the "cluster namespace" for that consumer
// in the provider) and the underlying space identifier (i.e., the
// `logicalcluster.Name`).
type KubeBindSpaceRelation interface {
	// SpaceIDToKubeBind maps an underlying space ID to a kube-bind space ID.
	// Returns empty string if there is no relationship.
	SpaceIDToKubeBind(spaceID string) string

	// SpaceIDFromKubeBind maps a kube-bind space ID to an underlying space ID.
	// Returns empty string if there is no relationship.
	SpaceIDFromKubeBind(kubeBindID string) string
}

// NewKubeBindSpaceRelation creates a KubeBindSpaceRelation given a
// Kubernetes client for the provider's space and a context bounding
// the relation's lifetime.
func NewKubeBindSpaceRelation(ctx context.Context, client kubernetes.Interface) *kubeBindSpaceRelation {
	informer := informerscorev1.NewFilteredConfigMapInformer(client, "kubestellar", 0, cache.Indexers{}, tweakListOptions)
	reln := &kubeBindSpaceRelation{
		informer:     informer,
		toKubeBind:   map[string]string{},
		fromKubeBind: map[string]string{},
	}
	informer.AddEventHandler(reln)
	go informer.Run(ctx.Done())
	return reln
}

var _ KubeBindSpaceRelation = &kubeBindSpaceRelation{}

func tweakListOptions(opts *metav1.ListOptions) {
	if opts.LabelSelector == "" {
		opts.LabelSelector = KubeBindLabelName
	} else {
		opts.LabelSelector = opts.LabelSelector + "," + KubeBindLabelName
	}
}

type kubeBindSpaceRelation struct {
	sync.Mutex
	informer     cache.SharedIndexInformer
	toKubeBind   map[string]string
	fromKubeBind map[string]string
}

func (reln *kubeBindSpaceRelation) OnAdd(obj any) {
	cm := obj.(*corev1.ConfigMap)
	if !strings.HasPrefix(cm.Name, CMNamePrefix) || len(cm.Name) <= len(CMNamePrefix)+1 {
		return
	}
	underID := cm.Name[len(CMNamePrefix):]
	kbSpaceID := cm.Labels[KubeBindLabelName]
	reln.setMapping(underID, kbSpaceID)
}

func (reln *kubeBindSpaceRelation) OnUpdate(oldObj, newObj any) {
	reln.OnAdd(newObj)
}

func (reln *kubeBindSpaceRelation) OnDelete(obj any) {
	switch o1 := obj.(type) {
	case cache.DeletedFinalStateUnknown:
		obj = o1.Obj
	default:
	}
	cm := obj.(*corev1.ConfigMap)
	if !strings.HasPrefix(cm.Name, CMNamePrefix) || len(cm.Name) <= len(CMNamePrefix)+1 {
		return
	}
	underID := cm.Name[len(CMNamePrefix):]
	reln.setMapping(underID, "")
}

// Associate newUnderID with newKBSpaceID, removing any pre-existing association
// for either of them.
// newUnderID is non-empty.
// newKBSpaceID may be empty, meaning to unassociate newUnderID with anything.
func (reln *kubeBindSpaceRelation) setMapping(newUnderID, newKBSpaceID string) {
	reln.Lock()
	defer reln.Unlock()
	oldKBSpaceID := reln.toKubeBind[newUnderID]
	if newKBSpaceID == oldKBSpaceID {
		return
	}
	if newKBSpaceID == "" {
		delete(reln.fromKubeBind, oldKBSpaceID)
		delete(reln.toKubeBind, newUnderID)
		return
	}
	oldUnderID := reln.fromKubeBind[newKBSpaceID]
	if oldUnderID != "" {
		delete(reln.toKubeBind, oldUnderID)
	}
	if oldKBSpaceID != "" {
		delete(reln.fromKubeBind, oldKBSpaceID)
	}
	reln.fromKubeBind[newKBSpaceID] = newUnderID
	reln.toKubeBind[newUnderID] = newKBSpaceID
}

func (reln *kubeBindSpaceRelation) InformerSynced() bool {
	return reln.informer.HasSynced()
}

func (reln *kubeBindSpaceRelation) SpaceIDToKubeBind(spaceID string) string {
	reln.Lock()
	defer reln.Unlock()
	return reln.toKubeBind[spaceID]
}

func (reln *kubeBindSpaceRelation) SpaceIDFromKubeBind(kubeBindID string) string {
	reln.Lock()
	defer reln.Unlock()
	return reln.fromKubeBind[kubeBindID]
}
