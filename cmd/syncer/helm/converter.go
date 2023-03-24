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
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/klog/v2"
	sigyaml "sigs.k8s.io/yaml"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"
)

func main() {
	var pathToHelmTemplate, outputDir, kcpUrl, kcpToken string
	var doConversion bool

	fs := pflag.NewFlagSet("helmconverter", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	fs.StringVar(&pathToHelmTemplate, "path-to-helm-template", "", "path to Helm tamplate file")
	fs.StringVar(&outputDir, "output-dir", "", "output directory")
	fs.StringVar(&kcpUrl, "kcp-url", "", "url of kcp workspace")
	fs.StringVar(&kcpToken, "kcp-token", "", "service token in kcp workspace")
	fs.BoolVar(&doConversion, "do-conversion", false, "service token in kcp workspace")
	fs.Parse(os.Args[1:])

	ctx := context.Background()
	logger := klog.FromContext(ctx)
	ctx = klog.NewContext(ctx, logger)
	_ = ctx
	fs.VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info(fmt.Sprintf("Command line flag %s %s", flg.Name, flg.Value))
	})

	unstObjs, err := loadYaml(pathToHelmTemplate)
	if err != nil {
		panic(err)
	}
	downSyncedResources := []edgev1alpha1.EdgeSyncConfigResource{}
	requiredNamespaces := map[string]bool{}
	for _, unstObj := range unstObjs {
		gvk := unstObj.GroupVersionKind()
		group := gvk.Group
		version := gvk.Version
		kind := gvk.Kind
		name := unstObj.GetName()
		namespace := unstObj.GetNamespace()
		downSyncedResource := edgev1alpha1.EdgeSyncConfigResource{
			Namespace: namespace,
			Name:      name,
			Group:     group,
			Kind:      kind,
			Version:   version,
		}
		downSyncedResources = append(downSyncedResources, downSyncedResource)
		if namespace != "" {
			_, ok := requiredNamespaces[namespace]
			if !ok {
				requiredNamespaces[namespace] = true
			}
		}
	}

	// conversions := []syncerconfig.Conversion{}
	// if doConversion {
	// 	conversions = []syncerconfig.Conversion{
	// 		{
	// 			Upstream: syncerconfig.ConversionGroupKind{
	// 				Group: "rbac.authorization.k8s.io.dummy",
	// 				Kind:  "RoleBinding",
	// 			},
	// 			Downstream: syncerconfig.ConversionGroupKind{
	// 				Group: "rbac.authorization.k8s.io",
	// 				Kind:  "RoleBinding",
	// 			},
	// 		},
	// 	}
	// }
	// syncerConfig.Conversions = conversions

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		panic(err)
	}
	// file, err := os.Create(outputDir + "/install.yaml")
	// if err != nil {
	// 	panic(err)
	// }
	// enc := yaml.NewEncoder(file)
	// for _, unstObj := range unstObjs {
	// 	gv, _ := schema.ParseGroupVersion(unstObj.GetAPIVersion())
	// 	for _, conversion := range conversions {
	// 		if gv.Group == conversion.Downstream.Group {
	// 			if unstObj.GetKind() == conversion.Downstream.Kind {
	// 				gv.Group = gv.Group + ".dummy"
	// 			}
	// 		}
	// 	}
	// 	unstObj.SetAPIVersion(gv.String())
	// 	if err := enc.Encode(unstObj.Object); err != nil {
	// 		panic(err)
	// 	}
	// }

	prependDownSyncedResources := []edgev1alpha1.EdgeSyncConfigResource{}
	for namespace, value := range requiredNamespaces {
		_ = value
		downSyncedResource := edgev1alpha1.EdgeSyncConfigResource{
			Name:    namespace,
			Kind:    "Namespace",
			Version: "v1",
		}
		prependDownSyncedResources = append(prependDownSyncedResources, downSyncedResource)
	}
	downSyncedResources = append(prependDownSyncedResources, downSyncedResources...)

	edgeSyncConfig := edgev1alpha1.EdgeSyncConfig{
		ObjectMeta: v1.ObjectMeta{
			Name: "sync-config",
		},
		Spec: edgev1alpha1.EdgeSyncConfigSpec{
			DownSyncedResources: downSyncedResources,
		},
	}
	if yamlData, err := sigyaml.Marshal(edgeSyncConfig); err != nil {
		panic(err)
	} else {
		if err := ioutil.WriteFile(outputDir+"/install.yaml", yamlData, os.ModePerm); err != nil {
			panic(err)
		}
	}
}

func loadYaml(path string) ([]*unstructured.Unstructured, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(f)
	k8sdec := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	var objects []*unstructured.Unstructured
	for {
		var objInput map[string]interface{}
		err := dec.Decode(&objInput)
		if errors.Is(err, io.EOF) {
			break
		} else if objInput == nil {
			continue
		} else if err != nil {
			return objects, err
		}
		yamlByte, err := yaml.Marshal(objInput)
		if err != nil {
			return objects, err
		}
		obj := &unstructured.Unstructured{}
		_, gvk, err := k8sdec.Decode(yamlByte, nil, obj)
		_ = gvk
		if err != nil {
			return objects, err
		}
		objects = append(objects, obj)
	}
	return objects, nil
}
