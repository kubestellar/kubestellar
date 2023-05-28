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

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"

	"github.com/spf13/pflag"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	machruntime "k8s.io/apimachinery/pkg/runtime"
	machserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	machjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/klog/v2"

	schedulingv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/scheduling/v1alpha1"

	edgeapi "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	"github.com/kcp-dev/edge-mc/pkg/customize"
)

func main() {
	fs := pflag.NewFlagSet("emc-customize", pflag.ExitOnError)
	klog.InitFlags(flag.CommandLine)
	fs.AddGoFlagSet(flag.CommandLine)
	var customizerFilename string = ""
	fs.StringVar(&customizerFilename, "customizer-filename", customizerFilename, "pathname of file holding Customizer to apply")

	ctx := context.Background()
	logger := klog.FromContext(ctx)

	err := fs.Parse(os.Args[1:])
	if err != nil {
		logger.Error(err, "Command line parse failed")
		os.Exit(1)
	}

	if fs.NArg() < 1 {
		logger.Error(nil, "Need at least one positional arguments")
		os.Exit(1)
	}

	scheme := machruntime.NewScheme()
	edgeapi.AddToScheme(scheme)
	schedulingv1alpha1.AddToScheme(scheme)

	codecFactory := machserializer.NewCodecFactory(scheme)
	decoder := codecFactory.UniversalDeserializer()

	neededArgs := 1

	var customizer *edgeapi.Customizer
	if customizerFilename != "" {
		obj, err := readObject(decoder, customizerFilename, &edgeapi.Customizer{})
		if err != nil {
			logger.Error(err, "Failed to read Customizer", "customizerFilename", customizerFilename)
			os.Exit(10)
		}
		customizer = obj.(*edgeapi.Customizer)
		expandCustomizer := customizer.GetAnnotations() != nil && customizer.GetAnnotations()[edgeapi.ParameterExpansionAnnotationKey] == "true"
		if expandCustomizer {
			neededArgs = 2
		}
	}
	logger.V(2).Info("Customizer", "cust", customizer)

	subj, err := readObject(decoder, fs.Arg(0), &unstructured.Unstructured{})
	if err != nil {
		logger.Error(err, "Failed to unmarshal subject", "filename", fs.Arg(0))
		os.Exit(20)
	}
	subject, ok := subj.(*unstructured.Unstructured)
	if !ok {
		logger.Error(nil, "Failed to convert read object to Unstructure", "readObj", subj)
		os.Exit(25)
	}
	if subject.GetAnnotations() != nil && subject.GetAnnotations()[edgeapi.ParameterExpansionAnnotationKey] == "true" {
		neededArgs = 2
	}

	if fs.NArg() < neededArgs || fs.NArg() > 2 {
		logger.Error(nil, "Wrong number of positional arguments", "min", neededArgs, "max", 2, "got", fs.NArg())
		os.Exit(1)
	}

	var location *schedulingv1alpha1.Location
	if fs.NArg() > 1 {
		obj, err := readObject(decoder, fs.Arg(1), &schedulingv1alpha1.Location{})
		if err != nil {
			logger.Error(err, "Failed to read Location", "locationFilename", fs.Arg(1))
			os.Exit(30)
		}
		location = obj.(*schedulingv1alpha1.Location)
	}
	logger.V(2).Info("Location", "loc", location)

	subject = customize.Customize(logger, subject, customizer, location)

	err = writeObject(codecFactory, os.Stdout, subject)
	if err != nil {
		logger.Error(err, "Failed to write output")
		os.Exit(99)
	}
}

func readObject(decoder machruntime.Decoder, filename string, dest machruntime.Object) (machruntime.Object, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	decoded, _, err := decoder.Decode(bytes, nil, dest)
	return decoded, err
}

func writeObject(cf machserializer.CodecFactory, dest io.Writer, source machruntime.Object) error {
	ser := machjson.NewYAMLSerializer(machjson.DefaultMetaFactory, nil, nil)
	ser.Encode(source, dest)
	return nil
}

func readJSON(filename string, dest any) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, dest)
}

func writeJSON(filename string, source any) error {
	bytes, err := json.MarshalIndent(source, "", "  ")
	if err == nil { // no, not really
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = file.Write(bytes)
	return err
}
