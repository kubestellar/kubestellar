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
	"fmt"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/openshift/client-go/security/clientset/versioned"

	tenancyv1alpha1 "github.com/kubestellar/kubeflex/api/v1alpha1"

	edgev1alpha1 "github.com/kubestellar/kubestellar/api/v1alpha1"
)

func GetClientSet(kubeconfig string) *kubernetes.Clientset {
	config := GetConfig(kubeconfig)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating clientset: %v\n", err)
		os.Exit(1)
	}
	return clientset
}

func GetClient() *client.Client {
	config := config.GetConfigOrDie()
	scheme := runtime.NewScheme()

	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating HTTPClient: %v\n", err)
		os.Exit(1)
	}
	mapper, err := apiutil.NewDiscoveryRESTMapper(config, httpClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating NewDiscoveryRESTMapper: %v\n", err)
		os.Exit(1)
	}
	if err := edgev1alpha1.AddToScheme(scheme); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding to schema: %v\n", err)
		os.Exit(1)
	}
	if err := tenancyv1alpha1.AddToScheme(scheme); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding to schema: %v\n", err)
		os.Exit(1)
	}
	c, err := client.New(config, client.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating client: %v\n", err)
		os.Exit(1)
	}
	return &c
}

func GetOpendShiftSecClient(kubeconfig string) (*versioned.Clientset, error) {
	config := GetConfig(kubeconfig)
	return versioned.NewForConfig(config)
}

func GetConfig(kubeconfig string) *rest.Config {
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, err := homedir.Dir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error finding home directory: %v\n", err)
				os.Exit(1)
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}
	return config
}
