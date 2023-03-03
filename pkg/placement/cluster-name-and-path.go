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
	"k8s.io/klog/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8scache "k8s.io/client-go/tools/cache"

	"github.com/kcp-dev/logicalcluster/v3"
)

type NameAndPathBinder struct {
	logger     klog.Logger
	nameToPath RelayMap[logicalcluster.Name, string]
	pathToName RelayMap[string, logicalcluster.Name]
}

const PathKey = "kcp.io/path"

func NewNameAndPath(logger klog.Logger, informer k8scache.SharedInformer, dedupConsumers bool) (
	nameToPath DynamicMapProducer[logicalcluster.Name, string],
	pathToName DynamicMapProducer[string, logicalcluster.Name]) {
	nap := &NameAndPathBinder{
		logger:     logger,
		nameToPath: NewRelayMap[logicalcluster.Name, string](dedupConsumers),
		pathToName: NewRelayMap[string, logicalcluster.Name](dedupConsumers),
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
	nap.nameToPath.Set(name, path)
	nap.pathToName.Set(path, name)
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
