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

package binding

import (
	"context"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/pkg/util"
)

type APIResource struct {
	groupVersion schema.GroupVersion
	resource     metav1.APIResource
}

// Handle CRDs should account for CRDs being added or deleted to start/stop new informers as needed
func (c *Controller) handleCRD(obj runtime.Object) error {
	uObj := obj.(*unstructured.Unstructured)
	var crdObj *apiextensionsv1.CustomResourceDefinition
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(uObj.UnstructuredContent(), &crdObj); err != nil {
		return fmt.Errorf("failed to convert Unstructured to CRD: %w", err)
	}

	// for each CustomResourceDefinitionVersion, follow a decision tree to tell whether the corresponding gvr should be watched
	toStartList, toStopList := []APIResource{}, []string{}
	for _, ver := range crdObj.Spec.Versions {
		gvr := APIResource{
			groupVersion: schema.GroupVersion{
				Group:   crdObj.Spec.Group,
				Version: ver.Name,
			},
			resource: metav1.APIResource{
				Name: crdObj.Spec.Names.Plural,
				Kind: crdObj.Spec.Names.Kind,
			},
		}
		key := util.KeyForGroupVersionKind(gvr.groupVersion.Group, gvr.groupVersion.Version, gvr.resource.Kind)
		if isBeingDeleted(obj) {
			toStopList = append(toStopList, key)
			continue
		}
		if !c.includedToWatch(gvr) {
			toStopList = append(toStopList, key)
			continue
		}
		if !ver.Served {
			toStopList = append(toStopList, key)
			continue
		}
		toStartList = append(toStartList, gvr)
	}

	go c.startInformersForNewAPIResources(toStartList)

	for _, key := range toStopList {
		c.logger.Info("API should be removed, stopping informer.", "key", key)
		if stopper, ok := c.stoppers[key]; ok {
			// close channel
			close(stopper)
		}
		// remove entries for key
		delete(c.informers, key)
		delete(c.listers, key)
		delete(c.stoppers, key)
		c.gvkGvrMapper.DeleteByGvkKey(key)
	}

	return nil
}

func (c *Controller) includedToWatch(r APIResource) bool {
	if _, excluded := excludedGroups[r.groupVersion.Group]; excluded {
		return false
	}
	if !util.IsAPIGroupAllowed(r.groupVersion.Group, c.allowedGroupsSet) {
		return false
	}
	if _, excluded := excludedResourceNames[r.resource.Name]; excluded {
		return false
	}
	return true
}

func (c *Controller) startInformersForNewAPIResources(toStartList []APIResource) {
	for _, toStart := range toStartList {
		c.logger.Info("Ensuring informer for:", "group", toStart.groupVersion.Group,
			"version", toStart.groupVersion, "kind", toStart.resource.Kind)

		gvr := schema.GroupVersionResource{
			Group:    toStart.groupVersion.Group,
			Version:  toStart.groupVersion.Version,
			Resource: toStart.resource.Name,
		}

		informer := cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					return c.dynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return c.dynamicClient.Resource(gvr).Watch(context.TODO(), metav1.ListOptions{})
				},
			},
			nil,
			0, //Skip resync
			cache.Indexers{},
		)

		// add the event handler functions (same as those used by the startup logic)
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: c.handleObject,
			UpdateFunc: func(old, new interface{}) {
				if shouldSkipUpdate(old, new) {
					return
				}
				c.handleObject(new)
			},
			DeleteFunc: func(obj interface{}) {
				c.handleObject(obj)
			},
		})
		key := util.KeyForGroupVersionKind(toStart.groupVersion.Group,
			toStart.groupVersion.Version, toStart.resource.Kind)

		// add the mapping between GVK and GVR
		c.gvkGvrMapper.Add(toStart.groupVersion.WithKind(toStart.resource.Kind), gvr)

		// ensure the lister
		lister := cache.NewGenericLister(informer.GetIndexer(), gvr.GroupResource())
		if _, ok := c.listers[key]; !ok {
			c.logger.V(3).Info("Setting lister", "GVK", key)
			c.listers[key] = lister
		} else {
			c.logger.V(3).Info("Lister already in place", "GVK", key)
		}

		// ensure the stopper
		stopper := make(chan struct{})
		if _, ok := c.stoppers[key]; !ok {
			c.logger.V(3).Info("Setting stopper", "GVK", key)
			c.stoppers[key] = stopper
		} else {
			c.logger.V(3).Info("Stopper already in place", "GVK", key)
		}

		// ensure the informer
		if _, ok := c.informers[key]; !ok {
			c.logger.V(3).Info("Setting and running informer", "GVK", key)
			c.informers[key] = informer
			go informer.Run(stopper)
		} else {
			c.logger.V(3).Info("Informer already in place", "GVK", key)
		}
	}
}
