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

package client

import (
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	tenancyv1alpha1 "github.com/kubestellar/kubeflex/api/v1alpha1"

	controlv1alpha1 "github.com/kubestellar/kubestellar/api/control/v1alpha1"
)

func GetClient() *client.Client {
	logger := klog.Background().WithName("client")

	config := config.GetConfigOrDie()
	scheme := runtime.NewScheme()

	if err := controlv1alpha1.AddToScheme(scheme); err != nil {
		logger.Error(err, "Error adding to schema - controlv1alpha1")
		os.Exit(1)
	}
	if err := tenancyv1alpha1.AddToScheme(scheme); err != nil {
		logger.Error(err, "Error adding to schema - tenancyv1alpha1")
		os.Exit(1)
	}
	c, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		logger.Error(err, "Error creating client")
		os.Exit(1)
	}
	return &c
}
