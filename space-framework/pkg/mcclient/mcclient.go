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

package mcclient

// kubestellar space-aware client impl

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	ksclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	ksinformers "github.com/kubestellar/kubestellar/pkg/client/informers/externalversions"
	spacev1alpha1 "github.com/kubestellar/kubestellar/space-framework/pkg/apis/space/v1alpha1"
	mcclientset "github.com/kubestellar/kubestellar/space-framework/pkg/mcclient/clientset"
)

const defaultProviderNs = "spaceprovider-default"

type KubestellarSpaceInterface interface {
	// Space returns clientset for given space.
	Space(name string, namespace ...string) (client mcclientset.Interface)
	// ConfigForSpace returns rest config for given space.
	ConfigForSpace(name string, namespace ...string) (*rest.Config, error)
}

type multiSpaceClient struct {
	ctx              context.Context
	configs          map[string]*rest.Config
	clientsets       map[string]*mcclientset.Clientset
	managerClientset *ksclientset.Clientset
	lock             sync.Mutex
}

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

func (mcc *multiSpaceClient) ConfigForSpace(name string, namespace ...string) (*rest.Config, error) {
	key, ns := namespaceKey(name, namespace)
	var err error
	mcc.lock.Lock()
	defer mcc.lock.Unlock()
	config, ok := mcc.configs[key]
	if !ok {
		// Try to get Space from API server.
		if config, _, err = mcc.getFromServer(name, ns); err != nil {
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

	managerClientset, err := ksclientset.NewForConfig(managerConfig)
	if err != nil {
		return client, err
	}

	client = &multiSpaceClient{
		ctx:              ctx,
		configs:          make(map[string]*rest.Config),
		clientsets:       make(map[string]*mcclientset.Clientset),
		managerClientset: managerClientset,
		lock:             sync.Mutex{},
	}

	client.startSpaceCollection(ctx, managerClientset)
	return client, nil
}

func (mcc *multiSpaceClient) startSpaceCollection(ctx context.Context, managerClientset *ksclientset.Clientset) {
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
func (mcc *multiSpaceClient) getFromServer(name string, namespace string) (*rest.Config, *mcclientset.Clientset, error) {
	space, err := mcc.managerClientset.SpaceV1alpha1().Spaces(namespace).Get(mcc.ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}

	if space.Status.Phase != spacev1alpha1.SpacePhaseReady {
		return nil, nil, errors.New("space is not ready")
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(space.Status.SpaceConfig))
	if err != nil {
		return nil, nil, err
	}
	cs, err := mcclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	// Calling function should acquire the mcc lock
	mcc.configs[space.Name] = config
	mcc.clientsets[space.Name] = cs

	return config, cs, nil
}

func namespaceKey(name string, namespaces []string) (key string, namespace string) {
	ns := defaultProviderNs
	if len(namespace) > 0 {
		ns = namespaces[0]
	}
	return ns + "/" + name, ns
}
