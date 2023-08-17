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
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

var (
	followFlag bool
	linesFlag  string
	whichArg   string
)

func New(streams genericclioptions.IOStreams, cliOpts *genericclioptions.ConfigFlags) *cobra.Command {

	getLogCmd := &cobra.Command{
		Use:   "get-log [flags] which",
		Short: "Get the log output from a central KubeStellar or kcp process",
		Args:  cobra.ExactArgs(1),
		Run: func(getLogCmd *cobra.Command, args []string) {
			getLogRun(getLogCmd, cliOpts, args)
		},
	}

	// Create a new pflag.FlagSet for custom flags
	fs := pflag.NewFlagSet("get-log", pflag.ExitOnError)
	fs.BoolVarP(&followFlag, "follow", "f", false, "Follow the log output")
	fs.StringVar(&linesFlag, "lines", "+0", "Number of lines to display")

	// Add custom pflag.FlagSet to the Cobra command
	getLogCmd.Flags().AddFlagSet(fs)

	return getLogCmd
}

func getLogRun(getLogCmd *cobra.Command, cliOpts *genericclioptions.ConfigFlags, args []string) {
	whichArg = args[0]
	ksNamespace := "kubestellar"
	ksSelector := "app=kubestellar-server"

	ctx := context.Background()
	logger := klog.FromContext(ctx)
	// ctx = klog.NewContext(ctx, logger)

	logfile := ""
	switch whichArg {
	case "kcp":
		logfile = "kcp.log"
	case "where-resolver":
		logfile = "kubestellar-where-resolver-log.txt"
	case "mailbox-controller":
		logfile = "mailbox-controller-log.txt"
	case "placement-translator":
		logfile = "placement-translator-log.txt"
	default:
		fmt.Println("Argument 1 must be one of: kcp, where-resolver, mailbox-controller, placement-translator")
		os.Exit(1)
	}

	getLogCmd.Flags().VisitAll(func(flg *pflag.Flag) {
		logger.V(1).Info(fmt.Sprintf("Command line flag %s %s", flg.Name, flg.Value))
	})

	// Build the Kubernetes clientset
	config, err := cliOpts.ToRESTConfig()
	if err != nil {
		logger.Error(err, "Failed to build config from flags")
		os.Exit(5)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Error(err, "Error creating Kubernetes clientset:")
		os.Exit(10)
	}
	namespace, err := getNamespace(clientset, ksNamespace)
	if err != nil {
		logger.Error(err, "Error getting server namespace:")
		os.Exit(15)
	}
	podName, err := getServerPodName(clientset, namespace, ksSelector)
	if err != nil {
		logger.Error(err, "Error getting server pod name:")
		os.Exit(20)
	} else {
		podName = strings.Trim(string(podName), "'")
	}

	err = executeCommandInPod(clientset, namespace, config, podName, []string{"ls", "/home/kubestellar/ready"}, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		logger.Error(err, "Error pod is not ready to accept commands:")
		os.Exit(25)
	}

	kubectlLogArgs := []string{"tail", "kubestellar-logs/" + logfile}
	if followFlag {
		kubectlLogArgs = append(kubectlLogArgs, "-f")
	}
	kubectlLogArgs = append(kubectlLogArgs, "-n", linesFlag)

	err = executeCommandInPod(clientset, namespace, config, podName, kubectlLogArgs, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		logger.Error(err, "Error executing command in pod:")
		os.Exit(30)
	}

}

func getNamespace(clientset *kubernetes.Clientset, namespace string) (string, error) {
	ns, err := clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})

	if ns != nil {
		return namespace, err
	}

	return "", fmt.Errorf("No namespaces found with name " + namespace)
}

func getServerPodName(clientset *kubernetes.Clientset, namespace string, selector string) (string, error) {
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), listOptions)

	if err != nil {
		return "", err
	}
	if len(pods.Items) == 0 {
		return "", fmt.Errorf("No pods found with label " + selector)
	}

	return pods.Items[0].Name, nil
}

func executeCommandInPod(clientset *kubernetes.Clientset, podNamespace string, config *rest.Config, podName string, command []string,
	stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	req := clientset.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(podNamespace).SubResource("exec")
	fmt.Println(command)
	option := &v1.PodExecOptions{
		Command: command,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}

	return nil
}
