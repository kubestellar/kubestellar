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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	machschema "k8s.io/apimachinery/pkg/runtime/schema"
	k8sdynamic "k8s.io/client-go/dynamic"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	upstreamcache "k8s.io/client-go/tools/cache"
	_ "k8s.io/component-base/metrics/prometheus/clientgo"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"

	kcpdynamic "github.com/kcp-dev/client-go/dynamic"
	kcpscopedclient "github.com/kcp-dev/kcp/pkg/client/clientset/versioned"
	kcpinformers "github.com/kcp-dev/kcp/pkg/client/informers/externalversions"

	clientopts "github.com/kubestellar/kubestellar/pkg/client-options"
	"github.com/kubestellar/kubestellar/pkg/mailboxwatch"
)

const mainName = "kubestellar-list-syncing-objects"

var watch bool
var outputJson bool

func main() {
	fs := pflag.NewFlagSet(mainName, pflag.ExitOnError)
	gvr := machschema.GroupVersionResource{Version: "v1"}
	var kind string
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)

	fs.StringVar(&gvr.Group, "api-group", gvr.Group, "API group of objects to watch (default is Kubernetes core group)")
	fs.StringVar(&gvr.Version, "api-version", gvr.Version, "API version (just version, no group) of objects to watch")
	fs.StringVar(&gvr.Resource, "api-resource", gvr.Resource, "API resource (lowercase plural) of objects to watch (defaults to lowercase(kind)+'s')")
	fs.StringVar(&kind, "api-kind", kind, "kind of objects to watch")
	fs.BoolVar(&watch, "watch", watch, "indicates whether to inform rather than just list")
	fs.BoolVar(&outputJson, "json", outputJson, "indicates whether to output as lines of JSON rather than YAML")

	parentClientOpts := clientopts.NewClientOpts("parent", "access to the parent of mailbox workspaces")
	parentClientOpts.SetDefaultCurrentContext("root")
	parentClientOpts.AddFlags(fs)

	allClientOpts := clientopts.NewClientOpts("all", "access to the chosen objects in all clusters")
	allClientOpts.SetDefaultCurrentContext("system:admin")
	allClientOpts.AddFlags(fs)

	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.Background()
	ctx = klog.NewContext(ctx, logger)

	if len(kind) == 0 {
		logger.Error(nil, "The --kind must not be the empty string")
	}
	if len(gvr.Resource) == 0 {
		gvr.Resource = strings.ToLower(kind) + "s"
	}

	parentClientConfig, err := parentClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make parent client config")
		os.Exit(2)
	}
	parentClientConfig.UserAgent = mainName

	parentClient := kcpscopedclient.NewForConfigOrDie(parentClientConfig)

	allClientConfig, err := allClientOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "failed to make all-cluster client config")
		os.Exit(2)
	}
	allClientConfig.UserAgent = mainName

	dynamicClusterClientset, err := kcpdynamic.NewForConfig(allClientConfig)
	if err != nil {
		logger.Error(err, "failed to make all-cluster dynamic client")
		os.Exit(3)
	}
	dynamicClusterResource := dynamicClusterClientset.Resource(gvr)

	listGVK := machschema.GroupVersionKind{Group: gvr.Group,
		Version: gvr.Version,
		Kind:    kind + "List"}

	exampleObj := &unstructured.Unstructured{}
	exampleObj.SetAPIVersion(gvr.GroupVersion().String())
	exampleObj.SetKind(kind)

	parentInformerFactory := kcpinformers.NewSharedScopedInformerFactory(parentClient, 0, "")
	mbPreInformer := parentInformerFactory.Tenancy().V1alpha1().Workspaces()

	informer := mailboxwatch.NewSharedInformer[k8sdynamic.NamespaceableResourceInterface, *unstructured.UnstructuredList](ctx, listGVK, mbPreInformer, dynamicClusterResource, exampleObj, 0, upstreamcache.Indexers{})

	informer.AddEventHandler(upstreamcache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj any) { log(logger, "add", obj) },
		UpdateFunc: func(oldObj, newObj any) { log(logger, "update", newObj) },
		DeleteFunc: func(obj any) { log(logger, "delete", obj) },
	})
	parentInformerFactory.Start(ctx.Done())
	upstreamcache.WaitForCacheSync(ctx.Done(), mbPreInformer.Informer().HasSynced)
	go informer.Run(ctx.Done())
	logger.V(2).Info("Running", "group", gvr.Group, "version", gvr.Version, "resource", gvr.Resource, "kind", kind)
	if watch {
		<-ctx.Done()
	} else {
		if !upstreamcache.WaitForCacheSync(ctx.Done(),
			informer.HasSynced) {
			logger.Error(nil, "Impossible")
		}
		time.Sleep(15 * time.Second)
	}
}

var logmu sync.Mutex

func log(logger klog.Logger, action string, obj any) {
	objU := obj.(*unstructured.Unstructured)
	logmu.Lock()
	defer logmu.Unlock()
	objData := objU.UnstructuredContent()
	outData := objData
	if watch {
		outData = map[string]any{"Action": action}
		for key, val := range objData {
			outData[key] = val
		}
	}
	objJ, err := json.Marshal(outData)
	if err != nil {
		logger.Error(err, "Failed to marshal as JSON", "outData", outData)
		return
	}
	if outputJson {
		objJS := string(objJ)
		fmt.Println(objJS)
		return
	}
	objY, err := yaml.JSONToYAML(objJ)
	if err != nil {
		logger.Error(err, "Failed to convert JSON to YAML", "objJ", objJ)
		return
	}
	fmt.Println("---")
	objYS := string(objY)
	fmt.Println(objYS)
}
