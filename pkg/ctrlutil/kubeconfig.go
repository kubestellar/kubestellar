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

package ctrlutil

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kfv1aplha1 "github.com/kubestellar/kubeflex/api/v1alpha1"

	kslclient "github.com/kubestellar/kubestellar/pkg/client"
)

const (
	ControlPlaneTypeLabel = "kflex.kubestellar.io/cptype"
	ControlPlaneTypeITS   = "its"
	ControlPlaneTypeWDS   = "wds"
	// errors
	ErrNoControlPlane        = "no control plane found. At least one control plane labeled with %s=%s must be present"
	ErrControlPlaneNotFound  = "control plane with type %s and name %s was not found"
	ErrMultipleControlPlanes = "more than one control plane of type %s was found and no name was specified"
)

func GetWDSKubeconfig(logger logr.Logger, wdsName string) (*rest.Config, string, error) {
	return getRestConfig(logger, wdsName, ControlPlaneTypeWDS)
}

func GetITSKubeconfig(logger logr.Logger, itsName string) (*rest.Config, string, error) {
	return getRestConfig(logger, itsName, ControlPlaneTypeITS)
}

// get the rest config for a control plane based on labels and name
func getRestConfig(logger logr.Logger, cpName, labelValue string) (*rest.Config, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	kubeClient := *kslclient.GetClient()
	labelSelector := labels.SelectorFromSet(labels.Set(map[string]string{
		ControlPlaneTypeLabel: labelValue,
	}))
	var targetCP *kfv1aplha1.ControlPlane

	// wait until there are control planes with supplied labels
	logger.Info("Waiting for cp", "type", labelValue, "name", cpName)
	err := wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		if cpName != "" {
			cp := &kfv1aplha1.ControlPlane{}
			err := kubeClient.Get(ctx, client.ObjectKey{Name: cpName}, cp, &client.GetOptions{})
			if err == nil {
				targetCP = cp
				return true, nil
			}
			return false, nil
		}
		list := &kfv1aplha1.ControlPlaneList{}
		err := kubeClient.List(ctx, list, &client.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			logger.Info("Failed to list control planes, will retry", "err", err)
			return false, nil
		}
		if len(list.Items) == 0 {
			logger.Info("No control planes yet, will retry")
			return false, nil
		}
		if len(list.Items) == 1 {
			targetCP = &list.Items[0]
			return true, nil
		}
		// TODO - we do not allow this case for a WDS as there is a 1:1 relashionship controller:cp for WDS.
		// Need to revisit for ITS where we can have multiple shards.
		// Assume we are not waiting for control planes to go away.
		return true, fmt.Errorf(ErrMultipleControlPlanes, labelValue)
	})
	if err != nil {
		return nil, "", err
	}
	if targetCP == nil {
		if cpName != "" {
			return nil, "", fmt.Errorf(ErrControlPlaneNotFound, labelValue, cpName)
		}
		return nil, "", fmt.Errorf(ErrNoControlPlane, ControlPlaneTypeLabel, labelValue)
	}

	clientset, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		return nil, "", fmt.Errorf("error creating new clientset: %w", err)
	}

	if targetCP.Status.SecretRef == nil {
		return nil, "", fmt.Errorf("access secret reference doesn't exist for %s", targetCP.Name)
	}
	namespace := targetCP.Status.SecretRef.Namespace
	name := targetCP.Status.SecretRef.Name
	key := targetCP.Status.SecretRef.InClusterKey

	// determine if the configuration is in-cluster or off-cluster and use related key
	_, err = rest.InClusterConfig()
	if err != nil { // off-cluster
		key = targetCP.Status.SecretRef.Key
	}

	secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("error getting secrets: %w", err)
	}

	restConf, err := restConfigFromBytes(secret.Data[key])
	if err != nil {
		return nil, "", fmt.Errorf("error getting rest config from bytes: %w", err)
	}

	return restConf, targetCP.Name, nil
}

func restConfigFromBytes(kubeconfig []byte) (*rest.Config, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}

	return clientConfig.ClientConfig()
}
