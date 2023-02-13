/*
Copyright 2022 The KCP Authors.

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

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	"github.com/kcp-dev/client-go/discovery"
	"github.com/kcp-dev/client-go/dynamic"
	"github.com/kcp-dev/client-go/dynamic/dynamicinformer"
	"github.com/kcp-dev/logicalcluster/v3"
)

/* This program is a kcp client that uses (a) either an all-cluster
   API discovery client or a cluster-focused one and (b) an
   all-cluster dynamic client to list all objects of the kinds found
   by API discovery.

   Note that the all-cluster discovery client does not discover all
   resources.

   Note that the all-cluster dynamic client does not work for some
   resources --- the ones found in cluster-focused but not all-cluster
   API discovery.
*/

type ClientOpts struct {
	loadingRules *clientcmd.ClientConfigLoadingRules
	overrides    clientcmd.ConfigOverrides
}

func NewClientOpts() *ClientOpts {
	return &ClientOpts{
		loadingRules: clientcmd.NewDefaultClientConfigLoadingRules(),
		overrides:    clientcmd.ConfigOverrides{},
	}
}

func (opts *ClientOpts) AddFlags(flags *pflag.FlagSet) {
	flags.StringVar(&opts.loadingRules.ExplicitPath, "kubeconfig", opts.loadingRules.ExplicitPath, "Path to the kubeconfig file to use for CLI requests")
	flags.StringVar(&opts.overrides.CurrentContext, "context", opts.overrides.CurrentContext, "The name of the kubeconfig context to use")
}

func (opts *ClientOpts) ToRESTConfig() (*rest.Config, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(opts.loadingRules, &opts.overrides)
	return clientConfig.ClientConfig()
}

func main() {
	resyncPeriod := time.Duration(0)
	var concurrency int = 4
	serverBindAddress := ":10203"
	var clusterPathStr string
	fs := pflag.NewFlagSet("watchall", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	cliOpts := NewClientOpts()
	cliOpts.AddFlags(fs)
	fs.Var(&utilflag.IPPortVar{Val: &serverBindAddress}, "server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")
	fs.IntVar(&concurrency, "concurrency", concurrency, "number of syncs to run in parallel")
	fs.StringVar(&clusterPathStr, "cluster", clusterPathStr, "logicalcluster.Path to focus on, empty string means no focus")
	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.FromContext(ctx)
	ctx = klog.NewContext(ctx, logger)
	fs.VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info("Command line flag", flg.Name, flg.Value)
	})

	mymux := mux.NewPathRecorderMux("watchall")
	mymux.Handle("/metrics", legacyregistry.Handler())
	routes.Profiling{}.Install(mymux)
	go func() {
		err := http.ListenAndServe(serverBindAddress, mymux)
		if err != nil {
			logger.Error(err, "Failure in web serving")
			panic(err)
		}
	}()

	config, err := cliOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to build config from flags")
		os.Exit(10)
	}

	discoveryClusterClient, err := discovery.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create discovery client")
		os.Exit(20)
	}

	var resourceListList []*metav1.APIResourceList

	if clusterPathStr != "" {
		path, ok := logicalcluster.NewValidatedPath(clusterPathStr)
		if !ok {
			logger.Error(nil, "Failed to parse cluster path", "path", clusterPathStr)
			os.Exit(30)
		}
		discoveryScopedClient := discoveryClusterClient.Cluster(path)
		resourceListList, err = discoveryScopedClient.ServerPreferredResources()
	} else {
		resourceListList, err = discoveryClusterClient.ServerPreferredResources()
	}
	if err != nil {
		logger.Error(err, "Failed to create discovery client")
		os.Exit(40)
	}

	dynamicClusterClient, err := dynamic.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create dynamic client")
		os.Exit(50)
	}

	clusterInformerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClusterClient, resyncPeriod)

	out := csv.NewWriter(os.Stdout)
	watcher := &watcherBase{logger: logger, out: out}
	notificationHandler := &myNotificationHandler{watcher}
	for _, group := range resourceListList {
		watcher.WriteCSV([]string{"GROUP", group.GroupVersion})
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			logger.Error(err, "Failed to parse a GroupVersion", "groupVersion", group.GroupVersion)
			continue
		}
		for _, resource := range group.APIResources {
			informable := verbsSupportInformers(resource.Verbs)
			watcher.WriteCSV([]string{"RESOURCE", group.GroupVersion, resource.Kind, fmt.Sprintf("%v", informable)})
			if informable {
				informer := clusterInformerFactory.ForResource(gv.WithResource(resource.Name)).Informer()
				informer.AddEventHandler(notificationHandler)
				// go informer.Run(ctx.Done())
				// time.Sleep(2 * time.Second)
			}
		}
	}
	clusterInformerFactory.Start(ctx.Done())
	clusterInformerFactory.WaitForCacheSync(ctx.Done())
	time.Sleep(5 * time.Second)
}

func verbsSupportInformers(verbs []string) bool {
	var hasList, hasWatch bool
	for _, verb := range verbs {
		switch verb {
		case "list":
			hasList = true
		case "watch":
			hasWatch = true
		}
	}
	return hasList && hasWatch
}

type watcherBase struct {
	logger klog.Logger

	mutex sync.Mutex
	out   *csv.Writer
}

func (wb *watcherBase) WriteCSV(record []string) {
	wb.mutex.Lock()
	defer wb.mutex.Unlock()
	wb.out.Write(record)
	wb.out.Flush()
}

type myNotificationHandler struct {
	*watcherBase
}

func (nh *myNotificationHandler) OnAdd(obj any) {
	mObj := obj.(metav1.Object)
	rObj := obj.(runtime.Object)
	ok := rObj.GetObjectKind()
	gvk := ok.GroupVersionKind()
	cluster := logicalcluster.From(mObj).String()
	nh.WriteCSV([]string{"OBJECT", cluster, gvk.GroupVersion().String(), gvk.Kind, mObj.GetNamespace(), mObj.GetName()})
}

func (nh *myNotificationHandler) OnUpdate(oldObj, newObj any) {
}

func (nh *myNotificationHandler) OnDelete(obj any) {
}
