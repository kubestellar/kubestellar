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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apiserver/pkg/server/mux"
	"k8s.io/apiserver/pkg/server/routes"
	kubedynamicinformer "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	upstreamcache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/metrics/legacyregistry"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	utilflag "k8s.io/kubernetes/pkg/util/flag"

	clusterdiscovery "github.com/kcp-dev/client-go/discovery"
	clusterdynamic "github.com/kcp-dev/client-go/dynamic"
	clusterdynamicinformer "github.com/kcp-dev/client-go/dynamic/dynamicinformer"
	kcpinformers "github.com/kcp-dev/client-go/informers"
	kcpapis "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	coreapi "github.com/kcp-dev/kcp/pkg/apis/core/v1alpha1"
	clusterclientset "github.com/kcp-dev/kcp/pkg/client/clientset/versioned/cluster"
	extkcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"
	"github.com/kcp-dev/logicalcluster/v3"

	urmetav1a1 "github.com/kcp-dev/edge-mc/pkg/apis/meta/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/apiwatch"
)

/* This program is a kcp client that monitors all LogicalClusters and,
   for each one of those, all objects in that cluster of kinds that
   existed when its monitoring began.
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
	serverBindAddress := ":10203"
	fs := pflag.NewFlagSet("watchall", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	cliOpts := NewClientOpts()
	cliOpts.AddFlags(fs)
	fs.Var(&utilflag.IPPortVar{Val: &serverBindAddress}, "server-bind-address", "The IP address with port at which to serve /metrics and /debug/pprof/")
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

	clusterClient, err := clusterclientset.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create kcp cluster client")
		os.Exit(20)
	}

	kcpClusterInformerFactory := extkcpinformers.NewSharedInformerFactory(clusterClient, 0)
	lcClusterInformer := kcpClusterInformerFactory.Core().V1alpha1().LogicalClusters().Informer()

	out := csv.NewWriter(os.Stdout)
	watcher := &watcherBase{ctx: ctx, logger: logger, out: out}

	discoveryClusterClient, err := clusterdiscovery.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create discovery client")
		os.Exit(20)
	}

	dynamicClusterClient, err := clusterdynamic.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Failed to create dynamic client")
		os.Exit(30)
	}

	dynamicClusterInformerFactory := clusterdynamicinformer.NewDynamicSharedInformerFactory(dynamicClusterClient, 0)

	crdClusterInformer := dynamicClusterInformerFactory.ForResource(apiextensionsv1.SchemeGroupVersion.WithResource("customresourcedefinitions"))
	bindingClusterInformer := dynamicClusterInformerFactory.ForResource(kcpapis.SchemeGroupVersion.WithResource("apibindings"))
	dynamicClusterInformerFactory.Start(ctx.Done())
	if !upstreamcache.WaitForCacheSync(ctx.Done(), crdClusterInformer.Informer().HasSynced, bindingClusterInformer.Informer().HasSynced) {
		logger.Error(nil, "Failed to sync all-cluster dynamic informers on CRDs, APIBindings")
		os.Exit(40)
	}
	acw := &allClustersWatcher{watcher, discoveryClusterClient,
		crdClusterInformer, bindingClusterInformer, dynamicClusterClient,
		map[logicalcluster.Name]*clusterWatcher{}}
	lcClusterInformer.AddEventHandler(acw)
	kcpClusterInformerFactory.Start(ctx.Done())
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.Tick(time.Second * 15):
			out.Flush()
		}
	}
}

type watcherBase struct {
	ctx    context.Context
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

type allClustersWatcher struct {
	*watcherBase
	discoveryClusterClient clusterdiscovery.DiscoveryClusterInterface
	crdClusterInformer     kcpinformers.GenericClusterInformer
	bindingClusterInformer kcpinformers.GenericClusterInformer
	dynamicClusterClient   clusterdynamic.ClusterInterface
	clusterWatchers        map[logicalcluster.Name]*clusterWatcher
}

func (acw *allClustersWatcher) OnAdd(obj any) {
	cluster, ok := obj.(*coreapi.LogicalCluster)
	if !ok {
		acw.logger.Error(nil, "Failed to cast notification object to LogicalCluster", "obj", obj)
		return
	}
	clusterName := logicalcluster.From(cluster)
	acw.WriteCSV([]string{"CLUSTER", clusterName.String()})
	if _, ok := acw.clusterWatchers[clusterName]; !ok {
		cw := acw.NewClusterWatcher(clusterName)
		acw.clusterWatchers[clusterName] = cw
	}
}

func (acw *allClustersWatcher) OnUpdate(oldObj, newObj any) {
}

func (acw *allClustersWatcher) OnDelete(obj any) {
}

func (acw *allClustersWatcher) NewClusterWatcher(clusterName logicalcluster.Name) *clusterWatcher {
	scopedDynamic := acw.dynamicClusterClient.Cluster(clusterName.Path())
	informerFactory := kubedynamicinformer.NewDynamicSharedInformerFactory(scopedDynamic, 0)
	discoveryScopedClient := acw.discoveryClusterClient.Cluster(clusterName.Path())
	crdInformer := acw.crdClusterInformer.Cluster(clusterName).Informer()
	bindingInformer := acw.bindingClusterInformer.Cluster(clusterName).Informer()
	resourceInformer, _, _ := apiwatch.NewAPIResourceInformer(context.Background(), clusterName.String(), discoveryScopedClient,
		crdInformer, bindingInformer)
	cw := &clusterWatcher{
		watcherBase:      acw.watcherBase,
		resourceWatchers: map[schema.GroupVersionResource]*resourceWatcher{},
	}
	resourceInformer.AddEventHandler(upstreamcache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			rsc := obj.(*urmetav1a1.APIResource)
			gvr := schema.GroupVersionResource{
				Group:    rsc.Spec.Group,
				Version:  rsc.Spec.Version,
				Resource: rsc.Spec.Name,
			}
			informable := verbsSupportInformers(rsc.Spec.Verbs)
			if _, ok := cw.resourceWatchers[gvr]; ok {
				return
			}
			acw.WriteCSV([]string{"RESOURCE", clusterName.String(), gvr.Group, gvr.Version, rsc.Spec.Kind, fmt.Sprintf("%v", informable)})
			rw := &resourceWatcher{cw}
			cw.resourceWatchers[gvr] = rw
			if informable {
				gi := informerFactory.ForResource(gvr).Informer()
				go gi.Run(acw.ctx.Done())
				gi.AddEventHandler(rw)
			}
		},
	})
	informerFactory.Start(acw.ctx.Done())
	go func() {
		acw.logger.V(3).Info("Starting resource informer", "cluster", clusterName)
		resourceInformer.Run(acw.ctx.Done())
	}()
	return cw
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

type clusterWatcher struct {
	*watcherBase
	resourceWatchers map[schema.GroupVersionResource]*resourceWatcher
}

type resourceWatcher struct {
	*clusterWatcher
}

func (nh *resourceWatcher) OnAdd(obj any) {
	mObj := obj.(metav1.Object)
	rObj := obj.(runtime.Object)
	ok := rObj.GetObjectKind()
	gvk := ok.GroupVersionKind()
	cluster := logicalcluster.From(mObj).String()
	nh.WriteCSV([]string{"OBJECT", cluster, gvk.GroupVersion().String(), gvk.Kind, mObj.GetNamespace(), mObj.GetName()})
}

func (nh *resourceWatcher) OnUpdate(oldObj, newObj any) {
}

func (nh *resourceWatcher) OnDelete(obj any) {
}
