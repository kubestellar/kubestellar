/*
Copyright 2022 The KubeStellar Authors.

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

package listwatch

import (
	"context"
	"fmt"

	machmeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	upcache "k8s.io/client-go/tools/cache"
	log "k8s.io/klog/v2"
)

type CrossClusterListerWatcher struct {
	clusterListWatch map[string]*upcache.ListWatch
}

func NewCrossClusterListerWatcher(clusterListWatch map[string]*upcache.ListWatch) *CrossClusterListerWatcher {
	clw := &CrossClusterListerWatcher{
		clusterListWatch: clusterListWatch,
	}
	return clw
}

func (clw *CrossClusterListerWatcher) Watch(options metav1.ListOptions) (watch.Interface, error) {
	combinedCh := make(chan watch.Event)
	for clusterName, lwForCluster := range clw.clusterListWatch {
		log.Infof("--- Whatch for cluster: %s", clusterName)
		clusterWatch, err := lwForCluster.Watch(options)
		if err != nil {
			return nil, fmt.Errorf("Watch for cluster %s failed: %w", clusterName, err)
		}

		go func(clusterName string) {
			for {
				select {
				case <-context.Background().Done(): //TODO: pass the ctx as param?
					log.Info("-ch closed")
					clusterWatch.Stop()
				case event, ok := <-clusterWatch.ResultChan():
					if !ok { //TODO this happenes after 10min(timeout). not good
						clusterWatch.Stop()
						log.Infof("ch closed for cluster: %s", clusterName)
						return
					}
					combinedCh <- event
					log.Infof("---- combine event %v for cluster %s", event.Object, clusterName)
					// cm := event.Object.(*apiv1.ConfigMap)
					// log.Infof("event is on the ch: %s", cm.Name)
				}
			}
		}(clusterName)
	}
	return watch.NewProxyWatcher(combinedCh), nil
}

func (clw *CrossClusterListerWatcher) List(options metav1.ListOptions) (runtime.Object, error) {
	// TODO implement general logic. this is just a test
	list1, err := clw.clusterListWatch["cluster1"].List(options)
	if err != nil {
		log.Error("failed to list cluster1")
		return list1, err
	}
	list2, err := clw.clusterListWatch["cluster2"].List(options)
	if err != nil {
		log.Error("failed to list cluster2")
		return list2, err
	}
	subItems1, err := machmeta.ExtractList(list1)
	if err != nil {
		//logger.Error(err, "Failed to machmeta.ExtractList", "sublist", sublist)
		log.Error("failed to get subItems1")
		return list1, err
	}
	subItems2, err := machmeta.ExtractList(list2)
	if err != nil {
		log.Error("failed to get subItems2")
		return list2, err
	}
	allItems := append(subItems1, subItems2...)
	mergedList := list2.DeepCopyObject()
	err = machmeta.SetList(mergedList, allItems)
	if err != nil {
		log.Error("failed to set  allItems")
		return mergedList, err
	}
	return mergedList, nil
}

func ClusterListWatch(config *rest.Config, gv schema.GroupVersion, resource string, namespace string, optionsModifier func(options *metav1.ListOptions)) *upcache.ListWatch {
	config.GroupVersion = &gv
	//TODO api/apis
	config.APIPath = "/api"
	config.NegotiatedSerializer = clientscheme.Codecs.WithoutConversion()
	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}
	c, err := rest.RESTClientFor(config)
	if err != nil {
		log.Fatalf("get rest client failed: %v", err)
	}
	return upcache.NewListWatchFromClient(c, resource, namespace, fields.Everything()) //Works. timeout 10min
}
