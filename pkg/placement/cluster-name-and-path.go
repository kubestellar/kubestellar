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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8scache "k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"
)

type NameAndPathBinder struct {
	logger     klog.Logger
	nameToPath RelayMap[logicalcluster.Name, string]
	pathToName RelayMap[string, logicalcluster.Name]
}

const PathKey = "kcp.io/path"

func NewNameAndPath(logger klog.Logger, informer k8scache.SharedInformer, dedupReceivers bool) (
	nameToPath DynamicMapProvider[logicalcluster.Name, string],
	pathToName DynamicMapProvider[string, logicalcluster.Name]) {
	nap := &NameAndPathBinder{
		logger:     logger,
		nameToPath: NewRelayMap[logicalcluster.Name, string](dedupReceivers),
		pathToName: NewRelayMap[string, logicalcluster.Name](dedupReceivers),
	}
	informer.AddEventHandler(nap)
	return nap.nameToPath, nap.pathToName
}

func (nap *NameAndPathBinder) OnAdd(obj any) {
	mObj := obj.(metav1.Object)
	name := logicalcluster.From(mObj)
	if name == "" {
		nap.logger.Error(nil, "No logicalcluster.Name", "obj", obj)
		return
	}
	path := mObj.GetAnnotations()[PathKey]
	if path == "" {
		nap.logger.Error(nil, "No logicalcluster Path", "obj", obj)
		return
	}
	nap.nameToPath.Receive(name, path)
	nap.pathToName.Receive(path, name)
}

func (nap *NameAndPathBinder) OnUpdate(oldObj, newObj any) {
	nap.OnAdd(newObj)
}

func (nap *NameAndPathBinder) OnDelete(obj any) {
	if typed, ok := obj.(k8scache.DeletedFinalStateUnknown); ok {
		obj = typed.Obj
	}
	mObj := obj.(metav1.Object)
	name := logicalcluster.From(mObj)
	if name == "" {
		nap.logger.Error(nil, "No logicalcluster.Name", "obj", obj)
		return
	}
	path := mObj.GetAnnotations()[PathKey]
	if path == "" {
		nap.logger.Error(nil, "No logicalcluster Path", "obj", obj)
		return
	}
	nap.nameToPath.Remove(name)
	nap.pathToName.Remove(path)
}
