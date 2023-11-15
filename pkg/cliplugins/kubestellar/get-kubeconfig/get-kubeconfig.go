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

package plugin

import (
    "context"
    "fmt"
	"io"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// Get name of pod running KubeStellar server
func GetServerPodName(client *kubernetes.Clientset, ctx context.Context, namespace, selector string) (string, error) {
	// Get list of pods matching selector in given namespace
	podNames, err := GetPodNames(client, ctx, namespace, selector)
	if err != nil {
		return "", err
	}

	// Make sure we get one matching pod
	if len(podNames) == 0 {
		err = fmt.Errorf("No pod in namespace %s with selector %s", namespace, selector)
		return "", err
	} else if len(podNames) > 1 {
		err = fmt.Errorf("More than one pod (%d) in namespace %s with selector %s", len(podNames), namespace, selector)
		return "", err
	}
	// Return pod name
	serverPodName := podNames[0]
	return serverPodName, nil
}

// Get a list (slice) of pod names, within a given namespace matching selector
func GetPodNames(client *kubernetes.Clientset, ctx context.Context, namespace, selector string) ([]string, error) {
	// slide for holding pod names
	var podNames []string
	// Get list of pods matching selector in given namespace
	podList, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return podNames, err
	}

	// Go through each pod, pull out its name, and append to podNames
	for _, podItems := range podList.Items {
		podNames = append(podNames, podItems.Name)
	}

	// Return pod names
	return podNames, nil
}

// Check if a KubeStellar container inside a pod is ready (nil error indicates ready)
func KsPodIsReady(client *kubernetes.Clientset, config *rest.Config, namespace, podName, container string) error {
	// Check if '/home/kubestellar/ready' exists in container
	err := ExecuteCommandInPod(client, config, namespace, podName, container, []string{"ls", "/home/kubestellar/ready"}, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	return nil
}

// Execute a command within a specified container inside a pod
func ExecuteCommandInPod(client *kubernetes.Clientset, config *rest.Config, namespace, podName, container string, command []string,
	stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	// Get REST request for executing in pod
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).Namespace(namespace).SubResource("exec")

	// Query options to add to exec call
	option := &corev1.PodExecOptions{

		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
		Container: container,
		Command:   command,
	}
	if stdin == nil {
		option.Stdin = false
	}

	// Add query options to req
	req.VersionedParams(option, scheme.ParameterCodec)

	// Set up bi-directional stream
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	// POST the request
	err = exec.Stream(remotecommand.StreamOptions{Stdin: stdin, Stdout: stdout, Stderr: stderr})
	if err != nil {
		return err
	}

	return nil
}

// Get KubeStellar kubeconfig
func GetKSKubeconfig(client *kubernetes.Clientset, ctx context.Context, namespace string, isInternal bool) ([]byte, error) {

	// Get KubeStellar secrets
	secret, err := client.CoreV1().Secrets(namespace).Get(ctx, "kubestellar", metav1.GetOptions{})
	if err != nil {
		return []byte{}, err
	}

	// Return internal or external kubeconfig
	if isInternal {
		internalKubeconfig := secret.Data["cluster.kubeconfig"]
		return internalKubeconfig, nil
	}
	externalKubeconfig := secret.Data["external.kubeconfig"]
	return externalKubeconfig, nil
}