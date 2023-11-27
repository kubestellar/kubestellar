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

package spacemanager

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"

	v1 "k8s.io/api/core/v1"
	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	spacev1alpha1apis "github.com/kubestellar/kubestellar/space-framework/pkg/apis/space/v1alpha1"
	spaceprovider "github.com/kubestellar/kubestellar/space-framework/pkg/space-manager/providerclient"
	providerkcp "github.com/kubestellar/kubestellar/space-framework/space-provider/kcp"
	kindprovider "github.com/kubestellar/kubestellar/space-framework/space-provider/kind"
	kflexprovider "github.com/kubestellar/kubestellar/space-framework/space-provider/kubeflex"
)

// Each provider gets its own namespace named prefixNamespace+providerName
const (
	prefixNamespace     = "spaceprovider-"
	PROVIDER_CONFIG_KEY = "kubeconfig"
)

func ProviderNS(name string) string {
	return prefixNamespace + name
}

func spaceKeyFunc(ns string, name string) string {
	if ns != "" {
		return ns + "/" + name
	}
	return name
}

type provider struct {
	name            string
	providerClient  spaceprovider.ProviderClient
	c               *controller
	providerWatcher spaceprovider.Watcher
	nameSpace       string
	discoveryPrefix string
}

// TODO: this is termporary for stage 1. For stage 2 we expect to have a uniform interface for all informers.
func newProviderClient(pType spacev1alpha1apis.SpaceProviderType, config string) spaceprovider.ProviderClient {
	var pClient spaceprovider.ProviderClient = nil
	var err error
	switch pType {
	case spacev1alpha1apis.KindProviderType:
		pClient, err = kindprovider.New(config)
	case spacev1alpha1apis.KubeflexProviderType:
		pClient, err = kflexprovider.New(config)
	case spacev1alpha1apis.KcpProviderType:
		pClient, err = providerkcp.New(config)
	default:
		return nil
	}
	if err != nil {
		runtime.HandleError(err)
		return nil
	}

	return pClient
}

// CreateProvider returns new provider client
func CreateProvider(c *controller, providerDesc *spacev1alpha1apis.SpaceProviderDesc) (*provider, error) {
	var configStr string
	var err error
	providerName := providerDesc.Name
	c.lock.Lock()
	defer c.lock.Unlock()

	prov, exists := c.providers[providerName]
	if exists {
		return prov, fmt.Errorf("provider %s already in the list", string(providerDesc.Spec.ProviderType))
	}

	if providerDesc.Spec.ProviderType != spacev1alpha1apis.KindProviderType {
		if providerDesc.Spec.SecretRef == nil {
			return nil, fmt.Errorf("Provider description for %s is missing secret reference", string(providerDesc.Name))
		}

		configStr, err = getConfigFromSecret(*c.k8sClientset, providerDesc.Spec.SecretRef)
		if err != nil {
			return nil, err
		}
	}
	newProviderClient := newProviderClient(providerDesc.Spec.ProviderType, configStr)
	if newProviderClient == nil {
		return nil, fmt.Errorf("failed to create client for provider: %s", string(providerDesc.Spec.ProviderType))
	}

	discoveryPrefix := providerDesc.Spec.SpacePrefixForDiscovery
	if discoveryPrefix == "" {
		discoveryPrefix = "*"
	}

	p := &provider{
		name:            providerName,
		providerClient:  newProviderClient,
		c:               c,
		nameSpace:       ProviderNS(providerName),
		discoveryPrefix: discoveryPrefix,
	}

	c.providers[providerName] = p
	return p, nil
}

func getConfigFromSecret(cs kubeclient.Clientset, sRef *v1.SecretReference) (string, error) {

	ctx := context.Background()
	logger := klog.FromContext(ctx)

	var kubeconfigBytes []byte
	secret, err := cs.CoreV1().Secrets(sRef.Namespace).Get(ctx, sRef.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	// Retrieve the kubeconfig data from the Secret
	kubeconfigData, found := secret.Data[PROVIDER_CONFIG_KEY]
	if !found {
		return "", errors.New("secret doesn't have kubeconfig data")
	}

	kubeconfigBytes, err = base64.StdEncoding.DecodeString(string(kubeconfigData))
	if err != nil {
		//We assume this happen because the data was not encoded pring message and get the string
		kubeconfigBytes = kubeconfigData
		logger.Info("Provider secret was not encoded", "Secret", sRef.Name)
	}

	return string(kubeconfigBytes), nil
}

func (p *provider) filterOut(spaceName string) bool {
	if p.discoveryPrefix == "*" {
		return false
	}
	return !strings.HasPrefix(spaceName, p.discoveryPrefix)
}

// StartDiscovery will start watching provider spaces for changes
func (p *provider) StartDiscovery() error {
	if p.providerWatcher != nil {
		return nil
	}
	watcher, err := p.providerClient.Watch()
	if err != nil {
		return err
	}
	p.providerWatcher = watcher
	go p.processProviderWatchEvents()

	return nil
}

// StopDiscovery will stop watching provider spaces for changes
func (p *provider) StopDiscovery() error {
	p.c.lock.Lock()
	defer p.c.lock.Unlock()

	if p.providerWatcher == nil {
		return errors.New("failed to stop provider discovery. provider watcher does not exist")
	}
	p.providerWatcher.Stop()
	return nil
}

func (p *provider) processProviderWatchEvents() {
	ctx := context.Background()
	logger := klog.FromContext(ctx)
	var refspace *spacev1alpha1apis.Space
	for {
		event, ok := <-p.providerWatcher.ResultChan()
		if !ok {
			logger.Info("Space provider watch was stopped", "provider", p.name)
			return
		}
		spaceName := event.Name
		refspaceObj, found, errSpace := p.c.spaceInformer.GetIndexer().GetByKey(spaceKeyFunc(p.nameSpace, spaceName))

		if found {
			refspace, ok = refspaceObj.(*spacev1alpha1apis.Space)
			if !ok {
				runtime.HandleError(errors.New("unexpected object type. expected SpaceProviderDesc"))
				continue
			}
		}

		if !found || (found && refspace.Spec.Type != spacev1alpha1apis.SpaceTypeManaged) {
			// For unmanaged spaces discover & update only spaces that match the provider prefix
			if p.filterOut(spaceName) {
				continue
			}
		}
		switch event.Type {
		case spaceprovider.Added:
			logger.Info("New space was detected", "space", event.Name, "provider", p.name)
			// A new space was detected either create it or change the status to READY
			if !found || errSpace != nil {
				// No corresponding Space, let's create it
				logger.V(2).Info("Creating new Space object", "space", event.Name)
				space := spacev1alpha1apis.Space{}
				space.Name = spaceName
				space.Spec.SpaceProviderDescName = p.name
				space.Spec.Type = spacev1alpha1apis.SpaceTypeUnmanaged
				err := p.createSpaceSecrets(&space, event.SpaceInfo)
				if err != nil {
					logger.Error(err, "Failed to create secrests for space "+p.name)
					continue
				}
				space.Status.Phase = spacev1alpha1apis.SpacePhaseReady
				_, err = p.c.clientset.SpaceV1alpha1().Spaces(p.nameSpace).Create(ctx, &space, metav1.CreateOptions{})
				chkErrAndReturn(logger, err, "Detected New space. Couldn't create the corresponding Space", "space name", spaceName)
			} else {
				logger.V(2).Info("Updating Space object", "space", event.Name)

				err := p.createSpaceSecrets(refspace, event.SpaceInfo)
				if err != nil {
					logger.Error(err, "Failed to create secrests for space "+p.name)
					continue
				}
				if refspace.Spec.Type == spacev1alpha1apis.SpaceTypeManaged && !containsFinalizer(refspace, finalizerName) {
					// When a physical space is removed we remove its finalizer
					// from the space object. when the space returns, we
					// need to restore the finalizer.
					refspace.ObjectMeta.Finalizers = append(refspace.ObjectMeta.Finalizers, finalizerName)
				}
				refspace.Status.Phase = spacev1alpha1apis.SpacePhaseReady
				_, err = p.c.clientset.SpaceV1alpha1().Spaces(p.nameSpace).Update(ctx, refspace, metav1.UpdateOptions{})
				chkErrAndReturn(logger, err, "Detected New space. Couldn't update the corresponding Space status", "space name", spaceName)
			}

		case spaceprovider.Deleted:
			logger.Info("A space was removed", "space", event.Name, "provider", p.name)
			if !found {
				// There is no space object so there is nothing we should do
				continue
			}
			if refspace.Spec.Type == spacev1alpha1apis.SpaceTypeManaged {
				_ = p.deleteSpaceSecrets(refspace)
				// If managed then we need to remove the finalizer.
				f := refspace.ObjectMeta.Finalizers
				for i := 0; i < len(refspace.ObjectMeta.Finalizers); i++ {
					if f[i] == finalizerName {
						refspace.ObjectMeta.Finalizers = append(f[:i], f[i+1:]...)
						i--
					}
				}
			}
			// If managed then we need to remove the finalizer.
			refspace.Status.Phase = spacev1alpha1apis.SpacePhaseNotReady
			_, err := p.c.clientset.SpaceV1alpha1().Spaces(p.nameSpace).Update(ctx, refspace, metav1.UpdateOptions{})
			chkErrAndReturn(logger, err, "Space was removed, Couldn't update the Space status")

		default:
			logger.Info("unknown event type", "type", event.Type)
		}
	}
}

const (
	SECRET_DATA_KEY = "kubeconfig"
)

func (p *provider) deleteSpaceSecrets(space *spacev1alpha1apis.Space) error {
	if space.Status.InClusterSecretRef != nil {
		p.c.k8sClientset.CoreV1().Secrets(space.Status.InClusterSecretRef.Namespace).Delete(p.c.ctx, space.Status.InClusterSecretRef.Name, metav1.DeleteOptions{})
	}

	if space.Status.ExternalSecretRef != nil {
		p.c.k8sClientset.CoreV1().Secrets(space.Status.ExternalSecretRef.Namespace).Delete(p.c.ctx, space.Status.ExternalSecretRef.Name, metav1.DeleteOptions{})

	}
	return nil
}

func (p *provider) createSpaceSecrets(space *spacev1alpha1apis.Space, spInfo spaceprovider.SpaceInfo) error {

	var secret *v1.Secret
	var secretName string

	needInClustr := space.Spec.AccessScope == spacev1alpha1apis.AccessScopeBoth ||
		space.Spec.AccessScope == spacev1alpha1apis.AccessScopeInCluster

	needExternal := space.Spec.AccessScope == spacev1alpha1apis.AccessScopeBoth ||
		space.Spec.AccessScope == spacev1alpha1apis.AccessScopeExternal

	if needInClustr {
		if spInfo.Config[spaceprovider.INCLUSTER] != "" {
			secretName = "incluster-" + space.Name
			secret = buildSecret(secretName, spInfo.Config[spaceprovider.INCLUSTER])
			_, err := p.c.k8sClientset.CoreV1().Secrets(p.nameSpace).Create(p.c.ctx, secret, metav1.CreateOptions{})
			if err != nil {
				if !k8sapierrors.IsAlreadyExists(err) {
					return err
				} else {
					_, err := p.c.k8sClientset.CoreV1().Secrets(p.nameSpace).Update(p.c.ctx, secret, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
			}

			space.Status.InClusterSecretRef = &v1.SecretReference{
				Name:      secretName,
				Namespace: p.nameSpace,
			}
		} else {
			return errors.New("missing needed in-cluster secret for space")
		}
	}

	if needExternal {
		if spInfo.Config[spaceprovider.EXTERNAL] != "" {
			secretName = "external-" + space.Name
			secret = buildSecret(secretName, spInfo.Config[spaceprovider.EXTERNAL])
			_, err := p.c.k8sClientset.CoreV1().Secrets(p.nameSpace).Create(p.c.ctx, secret, metav1.CreateOptions{})
			if err != nil {
				if !k8sapierrors.IsAlreadyExists(err) {
					return err
				} else {
					_, err := p.c.k8sClientset.CoreV1().Secrets(p.nameSpace).Update(p.c.ctx, secret, metav1.UpdateOptions{})
					if err != nil {
						return err
					}
				}
			}

			space.Status.ExternalSecretRef = &v1.SecretReference{
				Name:      secretName,
				Namespace: p.nameSpace,
			}
		} else {
			return errors.New("missing needed external secret for space")
		}
	}

	return nil
}

func buildSecret(secretName string, conf string) *v1.Secret {

	confData := base64.StdEncoding.EncodeToString([]byte(conf))
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
		},
		Data: map[string][]byte{
			SECRET_DATA_KEY: []byte(confData),
		},
		Type: v1.SecretTypeOpaque,
	}
	return secret
}

func chkErrAndReturn(logger logr.Logger, err error, msg string, keysAndValues ...interface{}) {
	if err != nil {
		logger.Error(err, msg)
	}
}
