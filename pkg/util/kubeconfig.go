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

package util

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
	ControlPlaneTypeIMBS  = "imbs"
	ControlPlaneTypeWDS   = "wds"
	// errors
	ErrNoControlPlane        = "no control plane found. At least one control plane labeled with %s=%s must be present"
	ErrControlPlaneNotFound  = "control plane with label %s=%s and name %s was not found"
	ErrMultipleControlPlanes = "more than one control plane with same label %s=%s was found and no name was specified"
)

func GetWDSKubeconfig(logger logr.Logger, wdsName, wdsLabel string) (*rest.Config, string, error) {
	// default label for WDS
	label := Label{
		Key:   ControlPlaneTypeLabel,
		Value: ControlPlaneTypeWDS,
	}
	// use the specified label instead of default label if provided
	var err error
	if wdsLabel != "" {
		label, err = SplitLabelKeyAndValue(wdsLabel)
		if err != nil {
			return nil, "", err
		}
	}
	logger.Info("using label", "key", label.Key, "value", label.Value)
	return getRestConfig(logger, wdsName, label.Key, label.Value)
}

func GetIMBSKubeconfig(logger logr.Logger) (*rest.Config, string, error) {
	return getRestConfig(logger, "", ControlPlaneTypeLabel, ControlPlaneTypeIMBS)
}

// get the rest config for a control plane based on labels and name
func getRestConfig(logger logr.Logger, cpName, labelKey, labelValue string) (*rest.Config, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	kubeClient := *kslclient.GetClient()
	list := &kfv1aplha1.ControlPlaneList{}
	labelSelector := labels.SelectorFromSet(labels.Set(map[string]string{
		ControlPlaneTypeLabel: labelValue,
	}))

	// wait until there are control planes with supplied labels
	logger.Info("waiting for cp with label", "key", labelKey, "value", labelValue)
	err := wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		err := kubeClient.List(ctx, list, &client.ListOptions{LabelSelector: labelSelector})
		if err != nil {
			logger.Error(err, "error listing control planes")
			return false, nil
		}
		if len(list.Items) == 0 {
			return true, fmt.Errorf(ErrNoControlPlane, labelKey, labelValue)
		}
		return true, nil
	})
	if err != nil {
		return nil, "", err
	}

	if len(list.Items) == 0 {
		return nil, "", fmt.Errorf(ErrNoControlPlane, labelKey, labelValue)
	}

	// get the target control plane
	var targetCP kfv1aplha1.ControlPlane
	if cpName != "" {
		found := false
		for _, cp := range list.Items {
			if cp.Name == cpName {
				targetCP = cp
				found = true
				break
			}
		}
		if !found {
			return nil, "", fmt.Errorf(ErrControlPlaneNotFound, labelKey, labelValue, cpName)
		}
	} else {
		// TODO - we do not allow this case for a WDS as there is a 1:1 relashionship controller:cp for WDS
		// Need to revisit for IMBS where we can have multiple shards
		if len(list.Items) > 1 {
			return nil, "", fmt.Errorf(ErrMultipleControlPlanes, labelKey, labelValue)
		}
		targetCP = list.Items[0]
	}

	clientset, err := kubernetes.NewForConfig(ctrl.GetConfigOrDie())
	if err != nil {
		return nil, "", fmt.Errorf("error creating new clientset: %w", err)
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
