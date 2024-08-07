/*
Copyright 2024 The KubeStellar Authors.

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

package status

import (
	"context"
	"fmt"
	"strings"

	// "k8s.io/apimachinery/pkg/runtime"
	// utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	// "k8s.io/apimachinery/pkg/util/wait"
	// "k8s.io/client-go/dynamic"
	// "k8s.io/client-go/dynamic/dynamicinformer"
	// "k8s.io/client-go/rest"
	// "k8s.io/client-go/tools/cache"
	// "k8s.io/client-go/util/workqueue"
	// "sigs.k8s.io/controller-runtime/pkg/log"
	// "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	// "github.com/kubestellar/kubestellar/pkg/binding"
	// ksclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned"
	// ksinformers "github.com/kubestellar/kubestellar/pkg/generated/informers/externalversions"
	// controllisters "github.com/kubestellar/kubestellar/pkg/generated/listers/control/v1alpha1"
	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/util"
)

// Should be wr-locked
type singletonState struct {
	wObjSync map[util.ObjectIdentifier]bool
}

func (ss *singletonState) Set(id util.ObjectIdentifier) {
	ss.wObjSync[id] = true
}

func (ss *singletonState) Unset(id util.ObjectIdentifier) {
	ss.wObjSync[id] = false
}

// Should it be a field of Contoller?
var theSingletonState singletonState = singletonState{
	wObjSync: map[util.ObjectIdentifier]bool{},
}

func (c *Controller) initSingletonState(ctx context.Context) error {
	logger := klog.FromContext(ctx)
	logger.Info("Initializing the desired state for singleton statuses")

	allBdgs, err := c.wdsKsClient.ControlV1alpha1().Bindings().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, bdg := range allBdgs.Items {
		logger.V(2).Info("Got Binding", "name", bdg.Name)
		list := bdg.Spec.Workload.NamespaceScope // Should also do the same thing for cluster scoped
		for _, item := range list {
			nsWObjRef := item.NamespaceScopeDownsyncObject
			lister, found := c.listers.Get(metav1GVR2schemaGVR(nsWObjRef.GroupVersionResource))
			if !found {
				return fmt.Errorf("could not get lister for gvr: %s", nsWObjRef.GroupVersionResource)
			}
			nsWObj, _ := lister.ByNamespace(nsWObjRef.Namespace).Get(nsWObjRef.Name)
			labels := nsWObj.(metav1.Object).GetLabels()
			if v, ok := labels[util.BindingPolicyLabelSingletonStatusKey]; ok {
				logger.V(2).Info("Found namespace scoped singleton workload object", "name", nsWObjRef.Name)
				if v == util.BindingPolicyLabelSingletonStatusValueSet {
					wObjIdentifier := util.ObjectIdentifier{
						GVK:        schema.GroupVersionKind{Group: nsWObjRef.Group, Version: nsWObjRef.Version, Kind: nsWObj.GetObjectKind().GroupVersionKind().Kind},
						Resource:   nsWObjRef.Resource,
						ObjectName: cache.NewObjectName(nsWObjRef.Namespace, nsWObjRef.Name),
					}
					theSingletonState.Set(wObjIdentifier)
					c.reconcileSingletonByWObj(ctx, nsWObjRef) // Test only
				}
			}
		}
	}

	return nil
}

func metav1GVR2schemaGVR(in metav1.GroupVersionResource) (out schema.GroupVersionResource) {
	out = schema.GroupVersionResource{
		Group:    in.Group,
		Version:  in.Version,
		Resource: in.Resource,
	}
	return
}

// Test only, should be removed soon
func (c *Controller) reconcileSingletonByWObj(ctx context.Context, wObjRef v1alpha1.NamespaceScopeDownsyncObject) error {
	logger := klog.FromContext(ctx)
	wsNameSuffix := strings.Join([]string{
		wObjRef.Namespace,
		wObjRef.Name,
	}, "-")
	// A (probably critical) question: How do I know which WEC is selected?
	list, _ := c.workStatusLister.ByNamespace("cluster1").List(labels.Everything())
	for _, obj := range list {
		wsName := obj.(metav1.Object).GetName()
		if !strings.HasSuffix(wsName, wsNameSuffix) {
			continue
		}
		status, err := util.GetWorkStatusStatus(obj)
		if err != nil {
			return err
		}
		if status == nil {
			return nil
		}
		logger.V(2).Info("Got status for workstatus", "name", wsName)
		wsRef, _ := runtimeObjectToWorkStatusRef(obj)
		return updateObjectStatus(ctx, &wsRef.sourceObjectIdentifier, status, c.listers, c.wdsDynClient)
	}
	return nil
}

// To be implemented
func (c *Controller) reconcileSingletonByBdg() error { return nil }

func (c *Controller) reconcileSingletonByWs(ctx context.Context, ref singletonWorkStatusRef) error {
	logger := klog.FromContext(ctx)
	obj, _ := c.workStatusLister.ByNamespace(ref.wecName).Get(ref.name)
	status, _ := util.GetWorkStatusStatus(obj)
	if sync, ok := theSingletonState.wObjSync[ref.sourceObjectIdentifier]; !ok {
		logger.V(2).Info("Not a singleton workload object, aborting", "sync", sync)
		return nil
	} else if !sync {
		logger.V(2).Info("Singleton workload object should not be synced, aborting", "sync", sync)
		return nil
	} else {
		logger.V(2).Info("Singleton workload object should be synced", "objectIdentifier", ref.sourceObjectIdentifier)
		return updateObjectStatus(ctx, &ref.sourceObjectIdentifier, status, c.listers, c.wdsDynClient)
	}
}
