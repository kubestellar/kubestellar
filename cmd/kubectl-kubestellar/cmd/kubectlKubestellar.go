/*
Copyright 2021 The KubeStellar Authors.

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

package cmd

import (
	goflags "flag"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	getlogcmd "github.com/kubestellar/kubestellar/pkg/cliplugins/kubestellar/get_log"
	// workspacecmd "github.com/kubestellar/kubestellar/pkg/cliplugins/workspace/cmd"
	// "github.com/kubestellar/kubestellar/pkg/cmd/help"
)

func KubectlKubestellarCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "kubestellar",
		Short: "kubectl plugin for KubeStellar",
		Long: help.Doc(`
			KubeStellar is a flexible solution for challenges associated with multi-cluster 
			configuration management for edge, multi-cloud, and hybrid cloud.

			This command provides KubeStellar specific sub-command for kubectl.
		`),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// setup klog
	fs := goflags.NewFlagSet("klog", goflags.PanicOnError)
	klog.InitFlags(fs)
	root.PersistentFlags().AddGoFlagSet(fs)

	if v := version.Get().String(); len(v) == 0 {
		root.Version = "<unknown>"
	} else {
		root.Version = v
	}

	// workspaceCmd, err := workspacecmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	// 	os.Exit(1)
	// }
	// root.AddCommand(workspaceCmd)

	getlogCmd := getlogcmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	root.AddCommand(getlogCmd)

	return root
}
