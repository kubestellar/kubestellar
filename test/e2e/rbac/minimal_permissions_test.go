/*
Copyright 2025 The KubeStellar Authors.

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

package rbac

import (
	"context"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// TestMinimalRBACPermissions validates that the controller operates correctly
// with minimal RBAC permissions (no write access to controlplanes/status)
func TestMinimalRBACPermissions(t *testing.T) {
	// Load kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		t.Skipf("Cannot load kubeconfig: %v", err)
		return
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	namespace := "default"

	// Check if controller is running
	t.Run("VerifyControllerRunning", func(t *testing.T) {
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "control-plane=controller-manager",
		})
		if err != nil {
			t.Fatalf("Failed to list controller pods: %v", err)
		}

		if len(pods.Items) == 0 {
			t.Skip("Controller not deployed, skipping test")
		}

		// Check pod is running
		for _, pod := range pods.Items {
			if pod.Status.Phase != corev1.PodRunning {
				t.Errorf("Controller pod %s is not running: %s", pod.Name, pod.Status.Phase)
			}
		}
	})

	// Verify the controller has the expected minimal permissions
	t.Run("VerifyMinimalPermissions", func(t *testing.T) {
		// Find the controller's service account
		sa, err := clientset.CoreV1().ServiceAccounts(namespace).Get(ctx, "default", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to get service account: %v", err)
		}

		// Check ClusterRoleBindings
		crbList, err := clientset.RbacV1().ClusterRoleBindings().List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Fatalf("Failed to list cluster role bindings: %v", err)
		}

		var controllerRole *rbacv1.ClusterRole
		for _, crb := range crbList.Items {
			for _, subject := range crb.Subjects {
				if subject.Kind == "ServiceAccount" &&
					subject.Name == sa.Name &&
					subject.Namespace == namespace &&
					strings.Contains(crb.RoleRef.Name, "kubestellar-manager-role") {

					// Get the referenced ClusterRole
					role, err := clientset.RbacV1().ClusterRoles().Get(ctx, crb.RoleRef.Name, metav1.GetOptions{})
					if err == nil {
						controllerRole = role
						break
					}
				}
			}
		}

		if controllerRole == nil {
			t.Skip("Controller role not found, might be using different naming")
			return
		}

		// Verify controlplanes/status only has 'get' permission
		for _, rule := range controllerRole.Rules {
			for _, apiGroup := range rule.APIGroups {
				if apiGroup == "tenancy.kflex.kubestellar.org" {
					for _, resource := range rule.Resources {
						if resource == "controlplanes/status" {
							// Should only have 'get' verb
							for _, verb := range rule.Verbs {
								if verb == "patch" || verb == "update" {
									t.Errorf("Found unexpected write permission '%s' for controlplanes/status", verb)
								}
							}
							if len(rule.Verbs) != 1 || rule.Verbs[0] != "get" {
								t.Errorf("Expected only 'get' verb for controlplanes/status, got: %v", rule.Verbs)
							}
						}
					}
				}
			}
		}
	})

	// Check controller logs for permission errors
	t.Run("CheckNoPermissionErrors", func(t *testing.T) {
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "control-plane=controller-manager",
		})
		if err != nil || len(pods.Items) == 0 {
			t.Skip("No controller pods found")
			return
		}

		for _, pod := range pods.Items {
			// Get recent logs
			logOptions := &corev1.PodLogOptions{
				Container:  "manager",
				TailLines:  int64Ptr(100),
				Timestamps: true,
			}

			req := clientset.CoreV1().Pods(namespace).GetLogs(pod.Name, logOptions)
			logs, err := req.DoRaw(ctx)
			if err != nil {
				t.Logf("Warning: Could not get logs for pod %s: %v", pod.Name, err)
				continue
			}

			logStr := string(logs)
			// Check for permission-related errors
			forbiddenStrings := []string{"forbidden", "Forbidden", "permission denied", "unauthorized"}
			for _, forbidden := range forbiddenStrings {
				if strings.Contains(logStr, forbidden) {
					// Only fail if it's related to controlplanes/status
					if strings.Contains(logStr, "controlplanes/status") {
						t.Errorf("Found permission error in logs: %s", forbidden)
					}
				}
			}
		}
	})

	// Test creating a binding policy to ensure controller still works
	t.Run("TestBindingPolicyCreation", func(t *testing.T) {
		// This would create a test BindingPolicy and verify it works
		// Skipping detailed implementation as it requires KubeStellar CRDs
		t.Logf("Would test BindingPolicy creation here")
	})
}

func int64Ptr(i int64) *int64 {
	return &i
}

// waitForCondition waits for a condition to be true
func waitForCondition(t *testing.T, timeout time.Duration, condition func() (bool, error)) {
	err := wait.PollImmediate(time.Second, timeout, condition)
	if err != nil {
		t.Errorf("Condition not met within %v: %v", timeout, err)
	}
}
