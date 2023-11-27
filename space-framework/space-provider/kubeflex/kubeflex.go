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

package kflexprovider

import (
	"context"
	"encoding/base64"
	"sync"
	"time"

	"github.com/go-logr/logr"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	clusterprovider "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager/providerclient"
)

const (
	CPGroup    = "tenancy.kflex.kubestellar.org"
	CPVersion  = "v1alpha1"
	CPKind     = "ControlPlane"
	CPResource = "controlplanes"
)

// KflexClusterProvider is a kubeflex cluster provider
type KflexClusterProvider struct {
	logger     logr.Logger
	ctx        context.Context
	pConfig    string
	dClient    dynamic.Interface
	kubeClient *kubernetes.Clientset
	watch      clusterprovider.Watcher
}

// New creates a new KflexClusterProvider
func New(pConfig string) (KflexClusterProvider, error) {

	ctx := context.Background()
	logger := klog.FromContext(ctx)

	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(pConfig))
	if err != nil {
		logger.Error(err, "Error loading kubeconfig")
		return KflexClusterProvider{}, err
	}

	dClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create dynamic clientset")
		return KflexClusterProvider{}, err
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create kube clientset")
		return KflexClusterProvider{}, err
	}

	return KflexClusterProvider{
		pConfig:    pConfig,
		dClient:    dClient,
		kubeClient: kubeClient,
		logger:     logger,
		ctx:        ctx,
	}, nil
}

func (k KflexClusterProvider) Create(name string, opts clusterprovider.Options) error {

	cpGVR := schema.GroupVersionResource{
		Group:    CPGroup,
		Version:  CPVersion,
		Resource: CPResource,
	}

	crUnstruct := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": CPGroup + "/" + CPVersion,
		"kind":       CPKind,
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": map[string]interface{}{
			"backend": "shared",
			"type":    "k8s",
		},
	}}

	_, err := k.dClient.Resource(cpGVR).Create(k.ctx, crUnstruct, v1.CreateOptions{})
	if err != nil {
		// TODO:  Need to differentiate between "already exists" and an error
		k.logger.Error(err, "Failed to create cluster", "name", name)
	}

	return err
}

func (k KflexClusterProvider) Delete(name string, opts clusterprovider.Options) error {
	logger := klog.Background()
	logger.V(2).Info("Deleting KubeFlex cluster", "name", name)

	cpGVR := schema.GroupVersionResource{
		Group:    CPGroup,
		Version:  CPVersion,
		Resource: CPResource,
	}

	err := k.dClient.Resource(cpGVR).Delete(k.ctx, name, v1.DeleteOptions{})
	if err != nil {
		// TODO:  Need to differentiate between "already exists" and an error
		k.logger.Error(err, "Failed to delete cluster", "name", name)
	}

	return err
}

// ListSpacesNames: returns a list of clusters in KubeFlex that are
// in the Ready condition.
func (k KflexClusterProvider) ListSpacesNames() ([]string, error) {
	var listClusterNames []string

	listSpaces, err := k.dClient.Resource(schema.GroupVersionResource{
		Group:    CPGroup,
		Version:  CPVersion,
		Resource: CPResource,
	}).List(context.TODO(), v1.ListOptions{})

	if err != nil {
		// TODO:  Need to differentiate between "already exists" and an error
		k.logger.Error(err, "Failed to list spaces")
		return nil, err
	}

	for _, unspace := range listSpaces.Items {

		if isSpaceReady(unspace) {
			metadata, found, err := unstructured.NestedMap(unspace.Object, "metadata")
			if err != nil || !found {
				k.logger.Error(err, "can't find the metadata field in the KF ControlPlan object")
				return nil, err
			}
			listClusterNames = append(listClusterNames, metadata["name"].(string))
		}
	}

	return listClusterNames, nil
}

func isSpaceReady(unspace unstructured.Unstructured) bool {

	logger := klog.Background()

	// Access the "status" field as a map.
	status, found, err := unstructured.NestedMap(unspace.Object, "status")
	if err != nil || !found {
		logger.Error(err, "can't find the status field in the KF ControlPlan object")
		return false
	}

	// Access the "conditions" field as a slice of maps.
	conditions, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil || !found {
		logger.Error(err, "can't find the conditions field in the KF ControlPlan status")
		return false
	}

	// Iterate through the array of maps and access a specific entry (e.g., "key1") in each map.
	for _, item := range conditions {
		if cond, ok := item.(map[string]interface{}); ok {
			if reason, exists := cond["reason"]; exists {
				if reason == "Available" {
					if type1, exists := cond["type"]; exists {
						if type1 == "Ready" {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// Get: obtains the kubeconfig for the given lcName cluster.
// TODO: switch from cli to kube directives
func (k KflexClusterProvider) Get(lcName string) (clusterprovider.SpaceInfo, error) {

	var externalConf, internalConf []byte
	secret, err := k.kubeClient.CoreV1().Secrets(lcName+"-system").Get(k.ctx, "admin-kubeconfig", v1.GetOptions{})
	if err != nil {
		return clusterprovider.SpaceInfo{}, err
	}

	externalConf, err = base64.StdEncoding.DecodeString(string(secret.Data["kubeconfig"]))
	if err != nil {
		//We assume this happen because the data was not encoded pring message and get the string
		externalConf = secret.Data["kubeconfig"]
		k.logger.Info("Provider secret was not encoded", "Secret", secret.Name)
	}

	internalConf, err = base64.StdEncoding.DecodeString(string(secret.Data["kubeconfig-incluster"]))
	if err != nil {
		//We assume this happen because the data was not encoded pring message and get the string
		internalConf = secret.Data["kubeconfig-incluster"]
		k.logger.Info("Provider secret was not encoded", "Secret", secret.Name)
	}

	lcInfo := clusterprovider.SpaceInfo{
		Name: lcName,
		Config: map[string]string{
			clusterprovider.EXTERNAL:  string(externalConf),
			clusterprovider.INCLUSTER: string(internalConf),
		},
	}
	return lcInfo, nil
}

func (k KflexClusterProvider) ListSpaces() ([]clusterprovider.SpaceInfo, error) {
	logger := klog.Background()
	lcNames, _ := k.ListSpacesNames()

	lcInfoList := make([]clusterprovider.SpaceInfo, 0, len(lcNames))

	for _, lcName := range lcNames {

		spInfo, err := k.Get(lcName)
		if err != nil {
			logger.Error(err, "couldn't get space", "space", lcName)
			continue
		}
		lcInfoList = append(lcInfoList, spInfo)
	}

	return lcInfoList, nil
}

func (k KflexClusterProvider) Watch() (clusterprovider.Watcher, error) {
	w := &KflexWatcher{
		ch:       make(chan clusterprovider.WatchEvent),
		provider: &k,
		chClosed: false}
	k.watch = w
	return w, nil
}

type KflexWatcher struct {
	init     sync.Once
	wg       sync.WaitGroup
	ch       chan clusterprovider.WatchEvent
	cancel   context.CancelFunc
	provider *KflexClusterProvider
	chClosed bool
}

func (k *KflexWatcher) Stop() {
	if k.cancel != nil {
		k.cancel()
	}
	k.wg.Wait()
	if !k.chClosed {
		close(k.ch)
		k.chClosed = true

	}
}

func (k *KflexWatcher) ResultChan() <-chan clusterprovider.WatchEvent {
	k.init.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		logger := klog.FromContext(ctx)
		k.cancel = cancel
		setClusters := sets.NewString()

		k.wg.Add(1)
		go func() {
			defer k.wg.Done()
			for {
				select {
				// TODO replace the 2 with a param at the cluster-provider-client level
				case <-time.After(2 * time.Second):
					list, err := k.provider.ListSpacesNames()
					if err != nil {
						logger.Error(err, "Failed to list KubeFlex clusters")
						continue
					}
					newSetClusters := sets.NewString(list...)
					// Check for new clusters.
					for _, name := range newSetClusters.Difference(setClusters).UnsortedList() {
						logger.V(2).Info("Processing KubeFlex cluster", "name", name)
						spaceInfo, err := k.provider.Get(name)
						logger.Error(err, "kflex Added")
						logger.Error(err, name)
						if err != nil {
							logger.V(2).Info("KubeFlex cluster is not ready. Retrying", "cluster", name)
							// Can't get the cluster info, so let's discover it again
							newSetClusters.Delete(name)
							continue
						}
						k.ch <- clusterprovider.WatchEvent{
							Type:      clusterprovider.Added,
							Name:      name,
							SpaceInfo: spaceInfo,
						}
					}
					// Check for deleted clusters.
					for _, cl := range setClusters.Difference(newSetClusters).UnsortedList() {
						logger.V(2).Info("Processing KubeFlex cluster delete", "name", cl)
						k.ch <- clusterprovider.WatchEvent{
							Type: clusterprovider.Deleted,
							Name: cl,
						}
					}
					setClusters = newSetClusters
				case <-ctx.Done():
					return
				}
			}
		}()
	})

	return k.ch
}
