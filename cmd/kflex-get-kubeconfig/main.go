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

package main

// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
// to ensure that exec-entrypoint and run can make use of them.

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	dynclient "k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"

	kfapi "github.com/kubestellar/kubeflex/api/v1alpha1"

	clientopts "github.com/kubestellar/kubestellar/options"
)

func main() {
	clientOptions := clientopts.NewClientOptions[*pflag.FlagSet]("", "reading ControlPlane")
	var cpName, cpLabelSelectorStr string
	var outputFilePath string
	inCluster := true
	clientOptions.AddFlagsSansName(pflag.CommandLine)
	pflag.StringVar(&cpName, "control-plane-name", cpName, "name of ControlPlane to read, or empty string (meaning to pick by label selector); default is empty string")
	pflag.StringVar(&cpLabelSelectorStr, "control-plane-label-selector", cpLabelSelectorStr, "label selector that identifies exactly one ControlPlane, or empty string (meaning to pick by name); default is empty string")
	pflag.StringVar(&outputFilePath, "output-file", outputFilePath, "pathname of file where the kubeconfig will be written; '-' means stdout")
	pflag.BoolVar(&inCluster, "in-cluster", inCluster, "whether to extract the kubeconfig for use in the kubeflex hosting cluster")
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	ctx := context.Background()
	logger := klog.FromContext(ctx)
	ctx = klog.NewContext(ctx, logger)
	setupLog := logger.WithName("setup")

	pflag.VisitAll(func(flg *pflag.Flag) {
		setupLog.V(1).Info("Command line flag", "name", flg.Name, "value", flg.Value)
	})

	if cpName == "" && cpLabelSelectorStr == "" {
		logger.Error(nil, "You must provide either a non-empty --control-plane-name OR a non-empty --control-plane-label-selector")
		os.Exit(1)
	}
	if cpName != "" && cpLabelSelectorStr != "" {
		logger.Error(nil, "You may not provide both a non-empty --control-plane-name AND a non-empty --control-plane-label-selector")
		os.Exit(1)
	}
	if outputFilePath == "" {
		logger.Error(nil, "The output file pathname may not be empty")
		os.Exit(1)
	}

	restConfig, err := clientOptions.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to determine REST client config")
		os.Exit(10)
	}

	kubeClient := kubernetes.NewForConfigOrDie(restConfig)
	coreClient := kubeClient.CoreV1()

	dynClient := dynclient.NewForConfigOrDie(restConfig)
	controlplanes := kfapi.GroupVersion.WithResource("controlplanes")
	cpClient := dynClient.Resource(controlplanes)

	var kubeconfigContent []byte
	backoff := wait.Backoff{
		Duration: time.Second * 5,
		Factor:   1.414,
		Jitter:   0.25,
		Steps:    24 * 60 * 3,
		Cap:      time.Second * 20,
	}
	err = wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (done bool, err error) {
		var cpU *unstructured.Unstructured
		if cpName != "" {
			cpU, err = cpClient.Get(ctx, cpName, metav1.GetOptions{})
			if err != nil {
				logger.Info("Failed to fetch ControlPlane", "name", cpName, "err", err)
				return false, nil
			}
		} else {
			cpsU, err := cpClient.List(ctx, metav1.ListOptions{LabelSelector: cpLabelSelectorStr})
			if err != nil {
				logger.Info("Failed to fetch ControlPlanes by label selector", "name", cpLabelSelectorStr, "err", err)
				return false, nil
			}
			if numGot := len(cpsU.Items); numGot != 1 {
				logger.Info("Did not get exactly 1 ControlPlane", "numGot", numGot)
				return false, nil
			}
			cpU = &cpsU.Items[0]
		}
		var cp kfapi.ControlPlane
		err = runtime.DefaultUnstructuredConverter.FromUnstructuredWithValidation(cpU.UnstructuredContent(), &cp, true)
		if err != nil {
			logger.Info("Failed to unmarshal", "err", err)
			return false, nil
		}
		if !kfapi.HasConditionAvailable(cp.Status.Conditions) {
			logger.Info("The ControlPlane is not ready")
			return false, nil
		}
		secretsClient := coreClient.Secrets(cp.Status.SecretRef.Namespace)
		secret, err := secretsClient.Get(ctx, cp.Status.SecretRef.Name, metav1.GetOptions{})
		if err != nil {
			logger.Info("Failed to read Secret", "namespace", cp.Status.SecretRef.Namespace, "name", cp.Status.SecretRef.Name, "err", err)
			return false, nil
		}
		key := cp.Status.SecretRef.Key
		if inCluster {
			key = cp.Status.SecretRef.InClusterKey
		}
		kubeconfigContent = secret.Data[key]
		if len(kubeconfigContent) == 0 {
			logger.Info("Secret lacks kubeconfig", "namespace", cp.Status.SecretRef.Namespace, "name", cp.Status.SecretRef.Name, "key", key)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		logger.Error(err, "Timed out waiting for ready ControlPlane")
		os.Exit(86)
	}
	out := os.Stdout
	if outputFilePath != "-" {
		out, err = os.Create(outputFilePath)
		if err != nil {
			logger.Error(err, "Failed to open file for writing", "path", outputFilePath)
			os.Exit(99)
		}
	}
	_, err = out.Write(kubeconfigContent)
	if err != nil {
		logger.Error(err, "Failed to write into the file")
		os.Exit(100)
	}
	if outputFilePath != "-" {
		err = out.Close()
		if err != nil {
			logger.Error(err, "Failed to close the file")
			os.Exit(101)
		}
	}
}
