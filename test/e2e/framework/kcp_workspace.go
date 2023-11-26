/*
Copyright 2022 The KubeStellar Authors.

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

/*
Portions of this code are based on or inspired by the KCP Author's work
Copyright Copyright 2022 The KCP Authors
Original code: https://github.com/kcp-dev/kcp/blob/release-0.11/test/e2e/framework/kcp.go
*/

package framework

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	kubernetesclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/kubestellar/kubestellar/test/e2e/logicalcluster"
)

type kcpServer struct {
	name           string
	lock           *sync.Mutex
	cfg            clientcmd.ClientConfig
	kubeconfigPath string
}

// Name exposes the path of the kubeconfig file of this kcp server.
func (c *kcpServer) KubeconfigPath() string {
	return c.kubeconfigPath
}

// RawConfig exposes a copy of the client config for this server.
func (c *kcpServer) RawConfig() (clientcmdapi.Config, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.cfg == nil {
		return clientcmdapi.Config{}, fmt.Errorf("programmer error: kcpServer.RawConfig() called before load succeeded. Stack: %s", string(debug.Stack()))
	}
	return c.cfg.RawConfig()
}

// BaseConfig returns a rest.Config for the "base" context. Client-side throttling is disabled (QPS=-1).
func (c *kcpServer) BaseConfig(t *testing.T) *rest.Config {
	t.Helper()

	cfg, err := c.config("base")
	require.NoError(t, err)
	cfg = rest.CopyConfig(cfg)
	return rest.AddUserAgent(cfg, t.Name())
}

// Config exposes a copy of the base client config for this server. Client-side throttling is disabled (QPS=-1).
func (c *kcpServer) config(context string) (*rest.Config, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.cfg == nil {
		return nil, fmt.Errorf("programmer error: kcpServer.Config() called before load succeeded. Stack: %s", string(debug.Stack()))
	}
	raw, err := c.cfg.RawConfig()
	if err != nil {
		return nil, err
	}

	config := clientcmd.NewNonInteractiveClientConfig(raw, context, nil, nil)

	restConfig, err := config.ClientConfig()
	if err != nil {
		return nil, err
	}

	restConfig.QPS = -1

	return restConfig, nil
}

var workspaceGVR = schema.GroupVersionResource{
	Group:    "tenancy.kcp.io",
	Version:  "v1alpha1",
	Resource: "workspaces",
}

func NewWorkspaceFixture(t *testing.T, server *kcpServer, parent logicalcluster.Path, prefix string) (path logicalcluster.Path, workspaceName string) {
	t.Helper()

	rawCfg, err := server.RawConfig()
	require.NoError(t, err, "failed to get raw kubeconfig")

	clientCfg, _ := WriteLogicalClusterConfig(t, rawCfg, "base", parent)
	cfg, err := clientCfg.ClientConfig()
	require.NoError(t, err, "failed to get rest config")

	dynamicClient, err := dynamic.NewForConfig(cfg)
	require.NoError(t, err, "failed to construct client for server")

	wsObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": workspaceGVR.GroupVersion().String(),
			"kind":       "Workspace",
			"metadata": map[string]interface{}{
				"generateName": prefix,
			},
			"spec": map[string]interface{}{
				"type": map[string]interface{}{
					"name": "universal",
					"path": "root",
				},
			},
		},
	}

	ctx, cancelFn := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
	defer cancelFn()

	_wsObj, err := dynamicClient.Resource(workspaceGVR).Create(ctx, wsObj, v1.CreateOptions{})
	require.NoError(t, err, "failed to create workspace")

	workspaceName = _wsObj.GetName()

	t.Logf("Waiting for workspace %s to be ready", workspaceName)

	require.Eventually(t, func() bool {
		_wsObj, err := dynamicClient.Resource(workspaceGVR).Get(ctx, workspaceName, v1.GetOptions{})
		if err != nil {
			t.Logf("error seen waiting for workspace %s to be ready: %v", workspaceName, err)
			return false
		}
		phase, ok := _wsObj.GetLabels()["tenancy.kcp.io/phase"]
		return ok && phase == "Ready"
	}, wait.ForeverTestTimeout, time.Millisecond*100)

	t.Cleanup(func() {
		ctx, cancelFn := context.WithDeadline(context.Background(), time.Now().Add(time.Second*30))
		defer cancelFn()

		err := dynamicClient.Resource(workspaceGVR).Delete(ctx, workspaceName, v1.DeleteOptions{})
		if apierrors.IsNotFound(err) || apierrors.IsForbidden(err) {
			return // ignore not found and forbidden because this probably means the parent has been deleted
		}
		require.NoErrorf(t, err, "failed to delete workspace %s", workspaceName)
	})

	return parent.Join(workspaceName), workspaceName
}

// NewFakeWorkloadServer creates a workspace in the provided server and org
// and creates a server fixture for the logical cluster that results.
func NewFakeWorkloadServer(t *testing.T, server *kcpServer, org logicalcluster.Path, syncTargetName string) *kcpServer {
	t.Helper()

	path, wsName := NewWorkspaceFixture(t, server, org, "downstream-")
	rawConfig, err := server.RawConfig()
	require.NoError(t, err, "failed to read config for server")
	logicalConfig, kubeconfigPath := WriteLogicalClusterConfig(t, rawConfig, "base", path)
	fakeServer := &kcpServer{
		name:           wsName,
		cfg:            logicalConfig,
		kubeconfigPath: kubeconfigPath,
		lock:           &sync.Mutex{},
	}
	return fakeServer
}

// WriteLogicalClusterConfig creates a logical cluster config for the given config and
// cluster name and writes it to the test's artifact path. Useful for configuring the
// workspace plugin with --kubeconfig.
func WriteLogicalClusterConfig(t *testing.T, rawConfig clientcmdapi.Config, contextName string, clusterName logicalcluster.Path) (clientcmd.ClientConfig, string) {
	t.Helper()

	logicalRawConfig := logicalClusterRawConfig(rawConfig, clusterName, contextName)
	artifactDir, err := CreateTempDirForTest(t, "artifacts")
	require.NoError(t, err)
	pathSafeClusterName := strings.ReplaceAll(clusterName.String(), ":", "_")
	kubeconfigPath := filepath.Join(artifactDir, fmt.Sprintf("%s.kubeconfig", pathSafeClusterName))
	err = clientcmd.WriteToFile(logicalRawConfig, kubeconfigPath)
	require.NoError(t, err)
	logicalConfig := clientcmd.NewNonInteractiveClientConfig(logicalRawConfig, logicalRawConfig.CurrentContext, &clientcmd.ConfigOverrides{}, nil)
	return logicalConfig, kubeconfigPath
}

// SharedKcpServer returns a kcp server fixture intended to be shared
// between tests.
func SharedKcpServer(t *testing.T) *kcpServer {
	t.Helper()

	serverName := "shared"
	kubeconfig := TestConfig.KCPKubeconfig()
	require.Greater(t, len(kubeconfig), 0)

	t.Logf("shared kcp server will target configuration %q", kubeconfig)

	cfg, err := loadKubeConfig(kubeconfig, "base")
	require.NoError(t, err, "failed to create persistent server fixture")

	return &kcpServer{
		name:           serverName,
		kubeconfigPath: kubeconfig,
		cfg:            cfg,
		lock:           &sync.Mutex{},
	}
}

// LoadKubeConfig loads a kubeconfig from disk. This method is
// intended to be common between fixture for servers whose lifecycle
// is test-managed and fixture for servers whose lifecycle is managed
// separately from a test run.
func loadKubeConfig(kubeconfigPath, contextName string) (clientcmd.ClientConfig, error) {
	fs, err := os.Stat(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	if fs.Size() == 0 {
		return nil, fmt.Errorf("%s points to an empty file", kubeconfigPath)
	}

	rawConfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load admin kubeconfig: %w", err)
	}

	return clientcmd.NewNonInteractiveClientConfig(*rawConfig, contextName, nil, nil), nil
}

// LogicalClusterRawConfig returns the raw cluster config of the given config.
func logicalClusterRawConfig(rawConfig clientcmdapi.Config, logicalClusterName logicalcluster.Path, contextName string) clientcmdapi.Config {
	var (
		contextClusterName  = rawConfig.Contexts[contextName].Cluster
		contextAuthInfoName = rawConfig.Contexts[contextName].AuthInfo
		configCluster       = *rawConfig.Clusters[contextClusterName] // shallow copy
	)

	configCluster.Server += logicalClusterName.RequestPath()

	return clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{
			contextName: &configCluster,
		},
		Contexts: map[string]*clientcmdapi.Context{
			contextName: {
				Cluster:  contextName,
				AuthInfo: contextAuthInfoName,
			},
		},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			contextAuthInfoName: rawConfig.AuthInfos[contextAuthInfoName],
		},
		CurrentContext: contextName,
	}
}

type KcpClusterClient struct {
	client *kubernetesclient.Clientset
}

func (c *KcpClusterClient) Cluster(path logicalcluster.Path) *kubernetesclient.Clientset {
	return c.client
}

type KcpDynamicClient struct {
	client dynamic.Interface
}

func (c *KcpDynamicClient) Cluster(path logicalcluster.Path) dynamic.Interface {
	return c.client
}
