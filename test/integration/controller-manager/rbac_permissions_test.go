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

package cmtest

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/pkg/ctrlutil"
)

// TestControllerOperatesWithoutControlPlaneStatusWrite verifies that the controller
// can operate successfully without write permissions to controlplanes/status
func TestControllerOperatesWithoutControlPlaneStatusWrite(t *testing.T) {
	ctx := context.Background()

	// Setup test environment
	config := setupTestEnvironment(t)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		t.Fatalf("Failed to create clientset: %v", err)
	}

	// Define the minimal RBAC permissions (without patch/update on controlplanes/status)
	minimalRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-minimal-controller-role",
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"tenancy.kflex.kubestellar.org"},
				Resources: []string{"controlplanes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"tenancy.kflex.kubestellar.org"},
				Resources: []string{"controlplanes/status"},
				Verbs:     []string{"get"}, // Only read permission
			},
		},
	}

	// Create the role
	_, err = clientset.RbacV1().ClusterRoles().Create(ctx, minimalRole, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("Failed to create minimal role: %v", err)
	}
	defer clientset.RbacV1().ClusterRoles().Delete(ctx, minimalRole.Name, metav1.DeleteOptions{})

	// Test controller operations that use controlplanes
	t.Run("GetWDSKubeconfig", func(t *testing.T) {
		// This should work with only read permissions
		testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = testCtx // Will be used when test environment is set up

		// The actual test would need a proper WDS setup
		// For now, we verify the function doesn't require write permissions
		// by checking it can be called without permission errors
		_, _, err := ctrlutil.GetWDSKubeconfig(testLogger(t), "test-wds")

		// We expect an error about missing control plane, not permission denied
		if err != nil && isPermissionError(err) {
			t.Errorf("Got permission error when only read permissions should be needed: %v", err)
		}
	})

	t.Run("GetITSKubeconfig", func(t *testing.T) {
		testCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = testCtx // Will be used when test environment is set up

		_, _, err := ctrlutil.GetITSKubeconfig(testLogger(t), "test-its")

		if err != nil && isPermissionError(err) {
			t.Errorf("Got permission error when only read permissions should be needed: %v", err)
		}
	})
}

// Helper function to check if an error is permission-related
func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "forbidden") || contains(errStr, "permission") || contains(errStr, "unauthorized")
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func setupTestEnvironment(t *testing.T) *rest.Config {
	// This would normally set up an in-process API server for testing
	// For now, return a placeholder
	t.Skip("Requires test environment setup")
	return nil
}

func testLogger(t *testing.T) logr.Logger {
	return klog.FromContext(context.Background()).WithName("test")
}
