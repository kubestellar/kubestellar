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

package cmd

import (
	"flag"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/component-base/version"
	"k8s.io/klog/v2"

	getlogcmd "github.com/kubestellar/kubestellar/pkg/cliplugins/kubestellar/get_log/cmd"
	"github.com/kubestellar/kubestellar/pkg/cmd/help"
)

func NewKubestellarCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "kubestellar",
		Short: "kubectl plugin for KubeStellar",
		Long: `KubeStellar is a flexible solution for challenges associated with multi-cluster 
configuration management for edge, multi-cloud, and hybrid cloud.

This command provides the kubestellar sub-command for kubectl.`,
//		SilenceUsage:  false,
//		SilenceErrors: false,
	}

	// setup klog
	fs := flag.NewFlagSet("klog", flag.PanicOnError)
	klog.InitFlags(fs)
	root.PersistentFlags().AddGoFlagSet(fs)
	cliOpts := genericclioptions.NewConfigFlags(false)
	cliOpts.AddFlags(root.PersistentFlags())

/*
	if v := version.Get().String(); len(v) == 0 {
		root.Version = "<unknown>"
	} else {
		root.Version = v
	}
*/

//	getlogCmd := getlogcmd.New(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, cliOpts)
//	root.AddCommand(getlogCmd)

	return root
}





//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func init() {
  rootCmd.AddCommand(tryCmd)
}

var tryCmd = &cobra.Command{
  Use:   "try",
  Short: "Try and possibly fail at something",
  RunE: func(cmd *cobra.Command, args []string) error {
    if err := someFunc(); err != nil {
	return err
    }
    return nil
  },
}