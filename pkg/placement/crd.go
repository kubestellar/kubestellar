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

package placement

import (
	"context"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"

	"github.com/kubestellar/kubestellar/pkg/util"
)

const (
	waitBeforeTrackingCRDs = 5 * time.Second
	CRDKind                = "CustomResourceDefinition"
	CRDGroup               = "apiextensions.k8s.io"
	CRDVersion             = "v1"
)

type APIResource struct {
	groupVersion schema.GroupVersion
	resource     metav1.APIResource
}

// Handle CRDs should account for CRDs being added or deleted to start/stop new informers as needed
func (c *Controller) handleCRD(obj runtime.Object) error {
	toStartList, toStopList, err := c.checkAPIResourcesForUpdates()
	if err != nil {
		return err
	}

	go c.startInformersForNewAPIResources(toStartList)

	for _, key := range toStopList {
		c.logger.Info("API removed, stopping informer.", "key", key)
		stopper := c.stoppers[key]
		// close channel
		close(stopper)
		// remove entries for key
		delete(c.informers, key)
		delete(c.listers, key)
		delete(c.stoppers, key)
		c.gvkGvrMapper.DeleteByGvkKey(key)
	}

	return nil
}

// checks what APIs need starting new informers or stopping informers.
// Returns a list of APIResources for informers to start and a list of keys for infomers to stop
func (c *Controller) checkAPIResourcesForUpdates() ([]APIResource, []string, error) {
	toStart := []APIResource{}
	toStop := []string{}

	// tracking keys are used to detect what API resources have been removed
	trackingKeys := map[string]bool{}
	for k := range c.informers {
		trackingKeys[k] = true
	}

	// Get all the api resources in the cluster
	apiResources, err := c.kubernetesClient.Discovery().ServerPreferredResources()
	if err != nil {
		// ignore the error caused by a stale API service
		if !strings.Contains(err.Error(), util.UnableToRetrieveCompleteAPIListError) {
			return nil, nil, err
		}
	}

	// Loop through the api resources and create informers and listers for each of them
	for _, group := range apiResources {
		if _, excluded := excludedGroupVersions[group.GroupVersion]; excluded {
			continue
		}
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			c.logger.Error(err, "Failed to parse a GroupVersion", "groupVersion", group.GroupVersion)
			continue
		}
		for _, resource := range group.APIResources {
			if _, excluded := excludedResourceNames[resource.Name]; excluded {
				continue
			}
			informable := verbsSupportInformers(resource.Verbs)
			if informable {
				key := util.KeyForGroupVersionKind(gv.Group, gv.Version, resource.Kind)
				if _, ok := c.informers[key]; !ok {
					toStart = append(toStart, APIResource{
						groupVersion: gv,
						resource:     resource,
					})
				}
				// remove the key from tracking keys, what is left in the map at the end are
				// keys to the informers that need to be stopped.
				delete(trackingKeys, key)
			}
		}
	}

	for k := range trackingKeys {
		toStop = append(toStop, k)
	}
	return toStart, toStop, nil
}

func (c *Controller) startInformersForNewAPIResources(toStartList []APIResource) {
	for _, toStart := range toStartList {
		c.logger.Info("New API added. Starting informer for:", "group", toStart.groupVersion.Group,
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
				if shouldSkipDelete(obj) {
					return
				}
				c.handleObject(obj)
			},
		})
		key := util.KeyForGroupVersionKind(toStart.groupVersion.Group,
			toStart.groupVersion.Version, toStart.resource.Kind)
		c.informers[key] = informer

		// add the mapping between GVK and GVR
		c.gvkGvrMapper.Add(toStart.groupVersion.WithKind(toStart.resource.Kind), gvr)

		// create and index the lister
		lister := cache.NewGenericLister(informer.GetIndexer(), gvr.GroupResource())
		c.listers[key] = lister
		stopper := make(chan struct{})
		defer close(stopper)
		c.stoppers[key] = stopper

		go informer.Run(stopper)
	}
	// block
	select {}
}
