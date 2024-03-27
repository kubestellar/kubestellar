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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/util"
)

type APIResource struct {
	groupVersion schema.GroupVersion
	resource     metav1.APIResource
}

// Handle CRDs should account for CRDs being added or deleted to start/stop new informers as needed
func (c *Controller) handleCRD(ctx context.Context, objIdentifier util.ObjectIdentifier) error {
	logger := klog.FromContext(ctx)
	var crdObj *apiextensionsv1.CustomResourceDefinition
	var specVersions []apiextensionsv1.CustomResourceDefinitionVersion

	obj, err := c.getObjectFromIdentifier(objIdentifier)
	if errors.IsNotFound(err) {
		logger.V(2).Info("Handling deleted CRD", "name", objIdentifier.ObjectName.Name)
	} else if err != nil {
		return fmt.Errorf("failed to get runtime.Object from identifier (%v): %w", objIdentifier, err)
	} else {
		uObj := obj.(*unstructured.Unstructured)
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(uObj.UnstructuredContent(), &crdObj); err != nil {
			return fmt.Errorf("failed to convert Unstructured to CRD: %w", err)
		}
		specVersions = crdObj.Spec.Versions
	}

	// for each CustomResourceDefinitionVersion, follow a decision tree to tell whether the corresponding gvr should be watched
	toStartList, toStopList := []APIResource{}, []schema.GroupVersionResource{}
	for _, ver := range specVersions {
		apiResource := APIResource{
			groupVersion: schema.GroupVersion{
				Group:   crdObj.Spec.Group,
				Version: ver.Name,
			},
			resource: metav1.APIResource{
				Name: crdObj.Spec.Names.Plural,
				Kind: crdObj.Spec.Names.Kind,
			},
		}

		gvr := apiResource.groupVersion.WithResource(apiResource.resource.Name)

		if !c.includedToWatch(apiResource) {
			continue
		}
		if isBeingDeleted(crdObj) {
			toStopList = append(toStopList, gvr)
			continue
		}
		if !ver.Served {
			toStopList = append(toStopList, gvr)
			continue
		}
		if crdEstablished(crdObj) {
			toStartList = append(toStartList, apiResource)
		}
	}

	if len(toStartList) > 0 {
		go c.startInformersForNewAPIResources(ctx, toStartList)
	}

	for _, gvr := range toStopList {
		logger.Info("API should not be watched, ensuring the informer's absence.", "gvr", gvr)
		stopper, ok := c.stoppers.Get(gvr)
		if !ok {
			logger.V(3).Info("Informer is already absent.", "gvr", gvr)
		} else {
			// close channel
			close(stopper)
		}
		// remove entries for gvr
		c.informers.Remove(gvr)
		c.listers.Remove(gvr)
		c.stoppers.Remove(gvr)
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

func crdEstablished(crd *apiextensionsv1.CustomResourceDefinition) bool {
	for _, condition := range crd.Status.Conditions {
		if condition.Type == apiextensionsv1.Established && condition.Status == apiextensionsv1.ConditionTrue {
			return true
		}
	}
	return false
}

func (c *Controller) startInformersForNewAPIResources(ctx context.Context, toStartList []APIResource) {
	logger := klog.FromContext(ctx)

	for _, toStart := range toStartList {
		gvr := toStart.groupVersion.WithResource(toStart.resource.Name)

		if _, found := c.informers.Get(gvr); found {
			logger.V(3).Info("Informer already ensured.", "gvr", gvr)
			continue
		}
		// from this point onwards, the gvr is guaranteed not to be mapped in the informers, listers and stoppers maps
		logger.Info("New API added. Starting informer for:", "group", toStart.groupVersion.Group,
			"version", toStart.groupVersion, "kind", toStart.resource.Kind, "resource", toStart.resource.Name)

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
			AddFunc: func(obj interface{}) {
				c.handleObject(obj, toStart.resource.Name, "add")
			},
			UpdateFunc: func(old, new interface{}) {
				if shouldSkipUpdate(old, new) {
					return
				}
				c.handleObject(new, toStart.resource.Name, "update")
			},
			DeleteFunc: func(obj interface{}) {
				c.handleObject(obj, toStart.resource.Name, "delete")
			},
		})
		c.informers.Set(gvr, informer)

		// add the lister since it necessarily does not exist
		logger.V(3).Info("Setting lister", "gvr", gvr)
		c.listers.Set(gvr, cache.NewGenericLister(informer.GetIndexer(), gvr.GroupResource()))

		// add the stopper since it necessarily does not exist
		stopper := make(chan struct{})
		logger.V(3).Info("Setting stopper", "gvr", gvr)
		c.stoppers.Set(gvr, stopper)

		// add the informer since it necessarily does not exist
		logger.V(3).Info("Setting and running informer", "gvr", gvr)
		c.informers.Set(gvr, informer)
		go informer.Run(stopper)
	}
}
