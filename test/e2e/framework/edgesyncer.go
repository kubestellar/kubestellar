/*
Copyright 2022 The KCP Authors.

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

package framework

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kcp-dev/logicalcluster/v3"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	kubernetesclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/yaml"

	workloadcliplugin "github.com/kcp-dev/kcp/pkg/cliplugins/workload/plugin"
	"github.com/kcp-dev/kcp/test/e2e/framework"

	edgesyncer "github.com/kcp-dev/edge-mc/pkg/syncer"
)

type SyncerOption func(t *testing.T, fs *edgeSyncerFixture)

func NewEdgeSyncerFixture(t *testing.T, server framework.RunningServer, path logicalcluster.Path) *edgeSyncerFixture {
	t.Helper()

	if !sets.NewString(framework.TestConfig.Suites()...).HasAny("edge-syncer") {
		t.Fatalf("invalid to use an edge syncer fixture when only the following suites were requested: %v", framework.TestConfig.Suites())
	}
	sf := &edgeSyncerFixture{
		upstreamServer:     server,
		edgeSyncTargetPath: path,
		edgeSyncTargetName: "psyncer-01",
	}
	return sf
}

// edgeSyncerFixture configures a syncer fixture. Its `Start` method does the work of starting a syncer.
type edgeSyncerFixture struct {
	upstreamServer     framework.RunningServer
	edgeSyncTargetPath logicalcluster.Path
	edgeSyncTargetName string
}

// CreateEdgeSyncTargetAndApplyToDownstream creates a default EdgeSyncConfig resource through the `workload edge-sync` CLI command,
// applies the edge-syncer-related resources in the physical cluster.
func (sf *edgeSyncerFixture) CreateEdgeSyncTargetAndApplyToDownstream(t *testing.T) *appliedEdgeSyncerFixture {
	t.Helper()

	useDeployedSyncer := len(framework.TestConfig.PClusterKubeconfig()) > 0

	// Write the upstream logical cluster config to disk for the workspace plugin
	upstreamRawConfig, err := sf.upstreamServer.RawConfig()
	require.NoError(t, err)
	_, kubeconfigPath := framework.WriteLogicalClusterConfig(t, upstreamRawConfig, "base", sf.edgeSyncTargetPath)

	syncerImage := framework.TestConfig.SyncerImage()
	if useDeployedSyncer {
		require.NotZero(t, len(syncerImage), "--syncer-image must be specified if testing with a deployed syncer")
	} else {
		// The image needs to be a non-empty string for the plugin command but the value doesn't matter if not deploying a syncer.
		syncerImage = "not-a-valid-image"
	}

	// Run the plugin command to enable the edge syncer and collect the resulting yaml
	t.Logf("Configuring workspace %s for syncing", sf.edgeSyncTargetPath)
	pluginArgs := []string{
		"workload",
		"edge-sync",
		sf.edgeSyncTargetName,
		"--syncer-image=" + syncerImage,
		"--output-file=-",
	}

	syncerYAML := framework.RunKcpCliPlugin(t, kubeconfigPath, pluginArgs)

	var downstreamConfig *rest.Config
	var downstreamKubeconfigPath string
	if useDeployedSyncer {
		// Test code is not implemented yet
	} else {
		// The syncer will target a logical cluster that is a child of the current workspace. A
		// logical server provides as a lightweight approximation of a pcluster for tests that
		// don't need to validate running workloads or interaction with kube controllers.
		downstreamServer := framework.NewFakeWorkloadServer(t, sf.upstreamServer, sf.edgeSyncTargetPath, sf.edgeSyncTargetName)
		downstreamConfig = downstreamServer.BaseConfig(t)
		downstreamKubeconfigPath = downstreamServer.KubeconfigPath()
	}

	// Apply the yaml output from the plugin to the downstream server
	framework.KubectlApply(t, downstreamKubeconfigPath, syncerYAML)

	// Extract the configuration for an in-process syncer from the resources that were
	// applied to the downstream server. This maximizes the parity between the
	// configuration of a deployed and in-process syncer.
	var syncerID string
	for _, doc := range strings.Split(string(syncerYAML), "\n---\n") {
		var manifest struct {
			metav1.ObjectMeta `json:"metadata"`
		}
		err := yaml.Unmarshal([]byte(doc), &manifest)
		require.NoError(t, err)
		if manifest.Namespace != "" {
			syncerID = manifest.Namespace
			break
		}
	}
	require.NotEmpty(t, syncerID, "failed to extract syncer namespace from yaml produced by plugin:\n%s", string(syncerYAML))

	syncerConfig := syncerConfigFromCluster(t, downstreamConfig, syncerID, syncerID)
	downstreamKubeClient, err := kubernetesclient.NewForConfig(downstreamConfig)
	require.NoError(t, err)

	downstreamDynamicKubeClient, err := dynamic.NewForConfig(downstreamConfig)
	require.NoError(t, err)

	return &appliedEdgeSyncerFixture{
		edgeSyncerFixture: *sf,

		SyncerConfig:                syncerConfig,
		SyncerID:                    syncerID,
		DownstreamConfig:            downstreamConfig,
		DownstreamKubeClient:        downstreamKubeClient,
		DownstreamDynamicKubeClient: downstreamDynamicKubeClient,
		DownstreamKubeconfigPath:    downstreamKubeconfigPath,
	}
}

// RunSyncer runs a new Syncer against the upstream kcp workspaces
// Whether the syncer runs in-process or deployed on a pcluster will depend
// whether --pcluster-kubeconfig and --syncer-image are supplied to the test invocation.
func (sf *appliedEdgeSyncerFixture) RunSyncer(t *testing.T) *StartedEdgeSyncerFixture {
	t.Helper()

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		err := edgesyncer.RunSyncer(ctx, sf.SyncerConfig, 1)
		require.NoError(t, err, "syncer failed to start")
	}()

	t.Cleanup(cancelFunc)

	return &StartedEdgeSyncerFixture{
		sf,
	}
}

// appliedEdgeSyncerFixture contains the configuration required to start an edge syncer and interact with its
// downstream cluster.
type appliedEdgeSyncerFixture struct {
	edgeSyncerFixture

	SyncerConfig *edgesyncer.SyncerConfig
	SyncerID     string

	// Provide cluster-admin config and client for test purposes. The downstream config in
	// SyncerConfig will be less privileged.
	DownstreamConfig            *rest.Config
	DownstreamKubeClient        kubernetesclient.Interface
	DownstreamDynamicKubeClient dynamic.Interface
	DownstreamKubeconfigPath    string
}

// StartedEdgeSyncerFixture contains the configuration used to start a syncer and interact with its
// downstream cluster.
type StartedEdgeSyncerFixture struct {
	*appliedEdgeSyncerFixture
}

// syncerConfigFromCluster reads the configuration needed to start an in-process
// syncer from the resources applied to a cluster for a deployed syncer.
func syncerConfigFromCluster(t *testing.T, downstreamConfig *rest.Config, namespace, syncerID string) *edgesyncer.SyncerConfig {
	t.Helper()

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	downstreamKubeClient, err := kubernetesclient.NewForConfig(downstreamConfig)
	require.NoError(t, err)

	// Read the upstream kubeconfig from the syncer secret
	secret, err := downstreamKubeClient.CoreV1().Secrets(namespace).Get(ctx, syncerID, metav1.GetOptions{})
	require.NoError(t, err)
	upstreamConfigBytes := secret.Data[workloadcliplugin.SyncerSecretConfigKey]
	require.NotEmpty(t, upstreamConfigBytes, "upstream config is required")
	upstreamConfig, err := clientcmd.RESTConfigFromKubeConfig(upstreamConfigBytes)
	require.NoError(t, err, "failed to load upstream config")

	// Read the downstream token from the deployment's service account secret
	var tokenSecret corev1.Secret
	framework.Eventually(t, func() (bool, string) {
		secrets, err := downstreamKubeClient.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			t.Errorf("failed to list secrets: %v", err)
			return false, fmt.Sprintf("failed to list secrets downstream: %v", err)
		}
		for _, secret := range secrets.Items {
			t.Logf("checking secret %s/%s for annotation %s=%s", secret.Namespace, secret.Name, corev1.ServiceAccountNameKey, syncerID)
			if secret.Annotations[corev1.ServiceAccountNameKey] == syncerID {
				tokenSecret = secret
				return len(secret.Data["token"]) > 0, fmt.Sprintf("token secret %s/%s for service account %s found", namespace, secret.Name, syncerID)
			}
		}
		return false, fmt.Sprintf("token secret for service account %s/%s not found", namespace, syncerID)
	}, wait.ForeverTestTimeout, time.Millisecond*100, "token secret in namespace %q for syncer service account %q not found", namespace, syncerID)
	token := tokenSecret.Data["token"]
	require.NotEmpty(t, token, "token is required")

	// Compose a new downstream config that uses the token
	downstreamConfigWithToken := framework.ConfigWithToken(string(token), rest.CopyConfig(downstreamConfig))
	return &edgesyncer.SyncerConfig{
		UpstreamConfig:   upstreamConfig,
		DownstreamConfig: downstreamConfigWithToken,
		SyncTargetPath:   logicalcluster.NewPath(""),
		SyncTargetName:   "",
		SyncTargetUID:    "",
		Interval:         time.Second * 3,
	}
}
