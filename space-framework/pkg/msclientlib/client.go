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
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
	kubeClientset    *kubernetes.Clientset
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
	kubeClientset, err := kubernetes.NewForConfig(managerConfig)
	if err != nil {
		return client, err
	}

	client = &multiSpaceClient{
		ctx:              ctx,
		configs:          make(map[string]*rest.Config),
		managerClientset: managerClientset,
		kubeClientset:    kubeClientset,
		lock:             sync.Mutex{},
	}

	client.startSpaceCollection(ctx, managerClientset)
	return client, nil
}

func (mcc *multiSpaceClient) startSpaceCollection(ctx context.Context, managerClientset *mgtclientset.Clientset) {
	numThreads := 2
	resyncPeriod := time.Duration(0)

	spaceInformerFactory := ksinformers.NewSharedInformerFactory(managerClientset, resyncPeriod)
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

	restConfig, err := mcc.getRestConfigFromSecret(true, space)
	if err != nil {
		return nil, err
	}
	// Calling function should acquire the mcc lock
	mcc.configs[space.Name] = restConfig

	return restConfig, nil
}

func namespaceKey(name string, ns string) (key string, namespace string) {
	return ns + "/" + name, ns
}

const CUBEKONFIG_KEY = "kubeconfig"

func (mcc *multiSpaceClient) getRestConfigFromSecret(internalAccess bool, space *spacev1alpha1.Space) (*rest.Config, error) {
	var secretRef *corev1.SecretReference
	if internalAccess {
		secretRef = space.Status.InClusterSecretRef
		if secretRef == nil {
			return nil, errors.New("missing InClusterSecretRef spec fileld for space: " + space.Name)
		}
	} else {
		secretRef = space.Status.ExternalSecretRef
		if secretRef == nil {
			return nil, errors.New("missing ExternalSecretRef spec fileld for space: " + space.Name)
		}
	}
	secret, err := mcc.kubeClientset.CoreV1().Secrets(secretRef.Namespace).Get(mcc.ctx, secretRef.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Retrieve the kubeconfig data from the Secret
	kubeconfigData, found := secret.Data[CUBEKONFIG_KEY]
	if !found {
		return nil, errors.New("secret doesn't have kubeconfig data")
	}

	kubeconfigBytes, err := base64.StdEncoding.DecodeString(string(kubeconfigData))
	if err != nil {
		return nil, err
	}

	// Restore the *rest.Config from the decoded kubeconfig data
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeconfigBytes)
	if err != nil {
		fmt.Printf("Error restoring RESTConfig: %v\n", err)
		os.Exit(1)
	}
	return restConfig, nil
}
