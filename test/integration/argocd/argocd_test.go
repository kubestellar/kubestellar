/*
Copyright 2026 The KubeStellar Authors.

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

package argocd

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// managedClusterGVR defines the GroupVersionResource for ManagedCluster custom resource
// This is used by the dynamic client to interact with cluster.open-cluster-management.io/v1 API
var (
	managedClusterGVR = schema.GroupVersionResource{
		Group:    "cluster.open-cluster-management.io",
		Version:  "v1",
		Resource: "managedclusters",
	}
)

// TestArgoCDClusterSecretIsCreated validates that KubeStellar creates an Argo CD cluster Secret
// when a ManagedCluster is created. This tests the integration between KubeStellar and Argo CD.
func TestArgoCDClusterSecretIsCreated(t *testing.T) {
	// Load kubeconfig from default location (~/.kube/config or KUBECONFIG env var)
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	// Get the REST config from kubeconfig
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		t.Fatalf("failed to load kubeconfig: %v", err)
	}

	// Create a standard Kubernetes clientset for working with core resources (Secrets, Pods, etc.)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("failed to create kubernetes clientset: %v", err)
	}

	ctx := context.Background()
	argoCDNamespace := "argocd"

	// Pre-check: verify argocd namespace exists
	_, err = clientset.CoreV1().Namespaces().Get(ctx, argoCDNamespace, metav1.GetOptions{})
	if err != nil {
		t.Skipf("argocd namespace does not exist, skipping test: %v", err)
	}

	// Create a dynamic client for working with custom resources (ManagedCluster)
	// Dynamic client allows us to work with CRDs without having their Go types
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		t.Fatalf("failed to create dynamic client: %v", err)
	}

	// Create unique cluster name with timestamp to avoid collisions
	testClusterName := fmt.Sprintf("test-managed-cluster-%d", time.Now().Unix())
	// Argo CD expects cluster Secrets to follow the naming pattern: cluster-<name>
	secretName := "cluster-" + testClusterName

	// Create an unstructured object representing the ManagedCluster CRD
	// We use unstructured because we don't have the ManagedCluster Go types imported
	managedCluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.open-cluster-management.io/v1",
			"kind":       "ManagedCluster",
			"metadata": map[string]interface{}{
				"name": testClusterName,
			},
			"spec": map[string]interface{}{
				"hubAcceptsClient": true,
			},
		},
	}

	// Create the ManagedCluster resource
	// This should trigger KubeStellar controllers to reconcile and create the Argo CD Secret
	t.Logf("Creating ManagedCluster: %s", testClusterName)
	_, err = dynamicClient.Resource(managedClusterGVR).Create(ctx, managedCluster, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("failed to create ManagedCluster: %v", err)
	}

	// Ensure cleanup happens even if the test fails
	// Delete the ManagedCluster at the end of the test
	defer func() {
		t.Logf("Cleaning up ManagedCluster: %s", testClusterName)
		err := dynamicClient.Resource(managedClusterGVR).Delete(ctx, testClusterName, metav1.DeleteOptions{})
		if err != nil {
			t.Logf("warning: failed to delete ManagedCluster: %v", err)
		}
	}()

	// Poll every 2 seconds for up to 60 seconds waiting for the Secret to be created
	// This gives KubeStellar controllers time to reconcile and create the Argo CD cluster Secret
	t.Logf("Waiting for Argo CD cluster Secret to be created: %s/%s", argoCDNamespace, secretName)
	err = wait.PollImmediate(2*time.Second, 60*time.Second, func() (bool, error) {
		// Try to get the Secret from the argocd namespace
		secret, err := clientset.CoreV1().Secrets(argoCDNamespace).Get(ctx, secretName, metav1.GetOptions{})
		if err != nil {
			// Secret doesn't exist yet, keep polling
			return false, nil
		}

		// Check if the Secret has labels
		if secret.Labels == nil {
			return false, nil
		}

		// Verify the Secret has the required Argo CD label
		// Argo CD requires: argocd.argoproj.io/secret-type=cluster
		secretType, exists := secret.Labels["argocd.argoproj.io/secret-type"]
		if !exists || secretType != "cluster" {
			return false, nil
		}

		// Secret exists with the correct label, polling succeeded
		return true, nil
	})

	// If polling timed out or failed, the integration is broken
	if err != nil {
		t.Fatalf("Argo CD cluster Secret was not created or does not have the correct label: %v", err)
	}

	// Get the Secret one more time for final validation
	secret, err := clientset.CoreV1().Secrets(argoCDNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get Secret after successful poll: %v", err)
	}

	// Validate that the Secret is of type Opaque (standard for Argo CD cluster Secrets)
	if secret.Type != corev1.SecretTypeOpaque {
		t.Errorf("expected Secret type to be Opaque, got: %s", secret.Type)
	}

	// Double-check the Argo CD cluster label is present and correct
	secretType := secret.Labels["argocd.argoproj.io/secret-type"]
	if secretType != "cluster" {
		t.Errorf("expected label argocd.argoproj.io/secret-type=cluster, got: %s", secretType)
	}

	t.Logf("Successfully validated Argo CD cluster Secret: %s/%s", argoCDNamespace, secretName)
}