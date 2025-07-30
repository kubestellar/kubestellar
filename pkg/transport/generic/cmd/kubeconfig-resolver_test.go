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

package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/klog/v2"
)

func TestResolveWDSKubeconfig(t *testing.T) {
	logger := klog.Background()

	t.Run("✅ Test Direct File Path Priority", func(t *testing.T) {
		// Create a temporary kubeconfig file
		tempDir := t.TempDir()
		kubeconfigPath := filepath.Join(tempDir, "test-kubeconfig")

		// Create a simple kubeconfig content
		kubeconfigContent := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://test-server:6443
  name: test-cluster
contexts:
- context:
    cluster: test-cluster
    user: test-user
  name: test-context
current-context: test-context
users:
- name: test-user
  user:
    token: test-token
`
		err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test kubeconfig: %v", err)
		}

		options := &TransportOptions{
			WdsKubeconfigPath: kubeconfigPath,
			WdsName:           "should-be-ignored", // This should be ignored due to priority
		}

		config, name, err := resolveWDSKubeconfig(options, logger)
		if err != nil {
			t.Fatalf("resolveWDSKubeconfig failed: %v", err)
		}

		if config == nil {
			t.Fatal("expected config to be non-nil")
		}

		if config.Host != "https://test-server:6443" {
			t.Errorf("expected host to be 'https://test-server:6443', got '%s'", config.Host)
		}

		t.Logf("✅ PASS: Direct file path test passed: %s", name)
	})

	t.Run("⏭️  Skip WDS Name Test (Requires Real Cluster)", func(t *testing.T) {
		t.Skip("Skipping cluster-based test - would hang waiting for real control plane")
	})

	t.Run("✅ Test No Configuration Error", func(t *testing.T) {
		options := &TransportOptions{
			WdsKubeconfigPath: "", // No file path
			WdsName:           "", // No WDS name
		}

		_, _, err := resolveWDSKubeconfig(options, logger)
		if err == nil {
			t.Fatal("expected error when no configuration provided")
		}

		expectedError := "no WDS configuration provided: specify either --wds-kubeconfig or --wds-name"
		if err.Error() != expectedError {
			t.Errorf("expected error '%s', got '%s'", expectedError, err.Error())
		}

		t.Logf("✅ PASS: No configuration error test: %v", err)
	})
}

func TestLoadKubeconfigFromPath(t *testing.T) {
	t.Run("✅ Test Valid Kubeconfig File", func(t *testing.T) {
		tempDir := t.TempDir()
		kubeconfigPath := filepath.Join(tempDir, "valid-kubeconfig")

		kubeconfigContent := `
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://valid-server:6443
  name: valid-cluster
contexts:
- context:
    cluster: valid-cluster
    user: valid-user
  name: valid-context
current-context: valid-context
users:
- name: valid-user
  user:
    token: valid-token
`
		err := os.WriteFile(kubeconfigPath, []byte(kubeconfigContent), 0644)
		if err != nil {
			t.Fatalf("failed to create test kubeconfig: %v", err)
		}

		config, err := loadKubeconfigFromPath(kubeconfigPath)
		if err != nil {
			t.Fatalf("loadKubeconfigFromPath failed: %v", err)
		}

		if config.Host != "https://valid-server:6443" {
			t.Errorf("expected host 'https://valid-server:6443', got '%s'", config.Host)
		}

		t.Log("✅ PASS: Valid kubeconfig file test")
	})

	t.Run("✅ Test Invalid Kubeconfig File", func(t *testing.T) {
		tempDir := t.TempDir()
		kubeconfigPath := filepath.Join(tempDir, "invalid-kubeconfig")

		// Invalid YAML content
		err := os.WriteFile(kubeconfigPath, []byte("invalid: yaml: content: ["), 0644)
		if err != nil {
			t.Fatalf("failed to create test kubeconfig: %v", err)
		}

		_, err = loadKubeconfigFromPath(kubeconfigPath)
		if err == nil {
			t.Fatal("expected error for invalid kubeconfig file")
		}

		t.Logf("✅ PASS: Invalid kubeconfig file test: %v", err)
	})

	t.Run("✅ Test Non-existent File", func(t *testing.T) {
		_, err := loadKubeconfigFromPath("/non/existent/path")
		if err == nil {
			t.Fatal("expected error for non-existent file")
		}

		t.Logf("✅ PASS: Non-existent file test: %v", err)
	})
}
