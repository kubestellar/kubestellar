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

// Import of k8s.io/client-go/plugin/pkg/client/auth ensures
// that all in-tree Kubernetes client auth plugins
// (e.g. Azure, GCP, OIDC, etc.)  are available.

import (
	"context"
	"flag"
	"os"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v1alpha1"
	clientopts "github.com/kubestellar/kubestellar/pkg/client-options"
	clusterclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned/cluster"
)

func main() {
	var clusterNameStr = ""
	fs := pflag.NewFlagSet("ls-syncer-config", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	fs.StringVar(&clusterNameStr, "cluster-name", clusterNameStr, "cluster to list in, empty string for all")

	clientOpts := clientopts.NewClientOpts("the", "access to all logical clusters")
	clientOpts.SetDefaultCurrentContext("system:admin")
	clientOpts.AddFlags(fs)

	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	clientConfig, err := clientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make client config")
		os.Exit(2)
	}

	clientConfig.UserAgent = "ls-syncer-config"

	clusterClientset := clusterclientset.NewForConfigOrDie(clientConfig)

	clusterEdge := clusterClientset.EdgeV1alpha1()
	clusterEdgeSyncfgs := clusterEdge.SyncerConfigs()

	var list *edgeapi.SyncerConfigList

	if clusterNameStr != "" {
		clusterName := logicalcluster.Name(clusterNameStr)

		clusterEdgeSyncfgsScoped := clusterEdgeSyncfgs.Cluster(clusterName.Path())
		list, err = clusterEdgeSyncfgsScoped.List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Error(err, "Failed to .EdgeV1alpha1().SyncerConfigs().Cluster().List()")
		} else {
			logger.Info("api-then-cluster list succeeded", "list", list)
		}

		scopedClientset := clusterClientset.Cluster(clusterName.Path())
		scopedEdge := scopedClientset.EdgeV1alpha1()
		scopedEdgeSyncfgs := scopedEdge.SyncerConfigs()
		list, err = scopedEdgeSyncfgs.List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Error(err, "Failed to .Cluster().EdgeV1alpha1().SyncerConfigs().List()")
		} else {
			logger.Info("cluster-then-api list succeeded", "list", list)
		}
	}
	list, err = clusterEdgeSyncfgs.List(ctx, metav1.ListOptions{})
	if err != nil {
		logger.Error(err, "Failed to .EdgeV1alpha1().SyncerConfigs().List()")
	} else {
		logger.Info("all-cluster list succeeded", "list", list)
	}
}
