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

package msclientlib

// kubestellar space-aware client impl

import (
	"context"
	"errors"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	spacev1alpha1 "github.com/kubestellar/kubestellar/space-framework/pkg/apis/space/v1alpha1"
	mgtclientset "github.com/kubestellar/kubestellar/space-framework/pkg/client/clientset/versioned"
	ksinformers "github.com/kubestellar/kubestellar/space-framework/pkg/client/informers/externalversions"
)

const defaultProviderNs = "spaceprovider-default"

type KubestellarSpaceInterface interface {
	// Space returns clientset for given space.

	//Space(name string, namespace ...string) (client mcclientset.Interface)

	// ConfigForSpace returns rest config for given space.
	ConfigForSpace(name string, providerNS string) (*rest.Config, error)
}

type multiSpaceClient struct {
	ctx context.Context
	// Currently configs holds the kubeconfig as a string for each Space.
	// This is temporary as Space is bing changed to use secrets instead of holding the config as
	// text string
	// The key to the map is the spaceName+ProviderNS  (currently we use the provider NS and not directly the provider name)
	configs          map[string]*rest.Config
	managerClientset *mgtclientset.Clientset
	lock             sync.Mutex
}

/*
	 func (mcc *multiSpaceClient) Space(name string, namespace ...string) mcclientset.Interface {
		key, ns := namespaceKey(name, namespace)
		var err error
		mcc.lock.Lock()
		defer mcc.lock.Unlock()
		clientset, ok := mcc.clientsets[key]
		if !ok {
			// Try to get Space from API server.
			if _, clientset, err = mcc.getFromServer(name, ns); err != nil {
				panic(fmt.Sprintf("invalid space name: %s. error: %v", name, err))
			}
		}
		return clientset
	}
*/

// Use the default Provider
// The folowing functions are temporary, once Space become a cluster-scope resurce
// We will not use the providerNS anymore.
func (mcc *multiSpaceClient) ConfigForSpace(name string, providerNS string) (*rest.Config, error) {
	nameS := defaultProviderNs
	if len(providerNS) > 0 {
		nameS = providerNS
	}
	key, ns := namespaceKey(name, nameS)
	var err error
	mcc.lock.Lock()
	defer mcc.lock.Unlock()
	config, ok := mcc.configs[key]
	if !ok {
		// Try to get Space from API server.
		if config, err = mcc.getFromServer(name, ns); err != nil {
			return nil, err
		}
	}
	return config, nil
}

var client *multiSpaceClient
var clientLock = &sync.Mutex{}

// NewMultiSpace creates new multi-space client and starts collecting space configs
func NewMultiSpace(ctx context.Context, managerConfig *rest.Config) (KubestellarSpaceInterface, error) {
	clientLock.Lock()
	defer clientLock.Unlock()

	if client != nil {
		return client, nil
	}

	managerClientset, err := mgtclientset.NewForConfig(managerConfig)
	if err != nil {
		return client, err
	}

	client = &multiSpaceClient{
		ctx:              ctx,
		configs:          make(map[string]*rest.Config),
		managerClientset: managerClientset,
		lock:             sync.Mutex{},
	}

	client.startSpaceCollection(ctx, managerClientset)
	return client, nil
}

func (mcc *multiSpaceClient) startSpaceCollection(ctx context.Context, managerClientset *mgtclientset.Clientset) {
	numThreads := 2
	resyncPeriod := time.Duration(0)

	spaceInformerFactory := ksinformers.NewSharedScopedInformerFactory(managerClientset, resyncPeriod, metav1.NamespaceAll)
	spaceInformer := spaceInformerFactory.Space().V1alpha1().Spaces().Informer()

	spaceInformerFactory.Start(ctx.Done())

	spaceController := newController(ctx, spaceInformer, mcc)
	go spaceController.Run(numThreads)
	cache.WaitForNamedCacheSync("spaces-management", ctx.Done(), spaceInformer.HasSynced)
}

// getFromServer will query API server for specific Space and cache it if it exists and ready.
// getFromServer caller function should acquire the mcc lock.
func (mcc *multiSpaceClient) getFromServer(name string, namespace string) (*rest.Config, error) {
	space, err := mcc.managerClientset.SpaceV1alpha1().Spaces(namespace).Get(mcc.ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if space.Status.Phase != spacev1alpha1.SpacePhaseReady {
		return nil, errors.New("space is not ready")
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(space.Status.SpaceConfig))
	if err != nil {
		return nil, err
	}
	// Calling function should acquire the mcc lock
	mcc.configs[space.Name] = config

	return config, nil
}

func namespaceKey(name string, ns string) (key string, namespace string) {
	return ns + "/" + name, ns
}
