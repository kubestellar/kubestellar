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
	"time"

	"github.com/spf13/pflag"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"
)

/* This program is a plain Kubernetes client that uses API discovery
   and the dynamic client library to list all objects in the given
   cluster, of kinds that existed when the monitoring began. */

func main() {
	resyncPeriod := time.Duration(0)
	var concurrency int = 4
	serverBindAddress := ":10203"
	fs := pflag.NewFlagSet("watchall", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	cliOpts := genericclioptions.NewConfigFlags(true)
	cliOpts.AddFlags(fs)
	fs.Var(&utilflag.IPPortVar{Val: &serverBindAddress}, "server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")
	fs.IntVar(&concurrency, "concurrency", concurrency, "number of syncs to run in parallel")
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
		os.Exit(5)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		klog.Error(err, "Failed to create dynamic client")
		os.Exit(10)
	}

	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, resyncPeriod)

	discoveryClient, err := cliOpts.ToDiscoveryClient()
	if err != nil {
		klog.Error(err, "Failed to create discovery client")
		os.Exit(15)
	}

	out := csv.NewWriter(os.Stdout)
	resourceListList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		klog.Error(err, "Failed to create discovery client")
		os.Exit(20)
	}

	notificationHandler := &myNotificationHandler{logger, out}
	for _, group := range resourceListList {
		out.Write([]string{"GROUP", group.GroupVersion})
		gv, err := schema.ParseGroupVersion(group.GroupVersion)
		if err != nil {
			logger.Error(err, "Failed to parse a GroupVersion", "groupVersion", group.GroupVersion)
			continue
		}
		for _, resource := range group.APIResources {
			informable := verbsSupportInformers(resource.Verbs)
			out.Write([]string{"RESOURCE", group.GroupVersion, resource.Kind, fmt.Sprintf("%v", informable)})
			if informable {
				informer := informerFactory.ForResource(gv.WithResource(resource.Name)).Informer()
				informer.AddEventHandler(notificationHandler)
			}
		}
	}
	out.Flush()
	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())
	out.Flush()
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

type myNotificationHandler struct {
	logger klog.Logger
	out    *csv.Writer
}

func (nh *myNotificationHandler) OnAdd(obj any, isInitialList bool) {
	mObj := obj.(metav1.Object)
	rObj := obj.(runtime.Object)
	ok := rObj.GetObjectKind()
	gvk := ok.GroupVersionKind()
	nh.out.Write([]string{"OBJECT", gvk.GroupVersion().String(), gvk.Kind, mObj.GetNamespace(), mObj.GetName()})
}

func (nh *myNotificationHandler) OnUpdate(oldObj, newObj any) {
}

func (nh *myNotificationHandler) OnDelete(obj any) {
}
