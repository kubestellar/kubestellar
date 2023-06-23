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

package syncer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kcp-dev/kcp/test/e2e/framework"

	edgeframework "github.com/kubestellar/kubestellar/test/e2e/framework"
)

func TestKubeStellarSyncerForTurboForDeletionScenario(t *testing.T) {
	var syncerConfigUnst *unstructured.Unstructured
	err := edgeframework.LoadFile("testdata/turbo-for-deletion-scenario/syncer-config.yaml", embedded, &syncerConfigUnst)
	require.NoError(t, err)

	var syncerConfig2Unst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/turbo-for-deletion-scenario/syncer-config2.yaml", embedded, &syncerConfig2Unst)
	require.NoError(t, err)

	framework.Suite(t, "kubestellar-syncer")

	syncerFixture := setup(t)
	wsPath := syncerFixture.WorkspacePath

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	upstreamKubeClueterClient := syncerFixture.UpstreamKubeClusterClient
	upstreamDynamicClueterClient := syncerFixture.UpstreamDynamicKubeClient

	t.Logf("Deploy workloads in workspace %q.", wsPath.String())
	framework.Kubectl(t, syncerFixture.UpstreamKubeconfigPath, "create", "-f", pathToTestdataDir()+"/turbo-for-deletion-scenario/resources.yaml")

	t.Logf("Create a SyncerConfig for test in workspace %q.", wsPath.String())
	createdSyncerConfigUnst, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(syncerConfigGvr).Create(ctx, syncerConfigUnst, v1.CreateOptions{})
	require.NoError(t, err)

	downstreamKubeClient := syncerFixture.DownstreamKubeClient

	t.Logf("Wait for resources to be downsynced.")
	framework.Eventually(t, func() (bool, string) {
		_, err := downstreamKubeClient.CoreV1().Namespaces().Get(ctx, "optimized", v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get namespace %s: %v", "optimized", err)
		}
		_, err = downstreamKubeClient.AppsV1().Deployments("optimized").Get(ctx, "cpu-usage", v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get deployment %s: %v", "cpu-usage", err)
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")

	t.Logf("Update SyncerConfig for test in workspace %q.", wsPath.String())
	syncerConfig2Unst.SetResourceVersion(createdSyncerConfigUnst.GetResourceVersion())
	syncerConfig2Unst.SetUID(createdSyncerConfigUnst.GetUID())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(syncerConfigGvr).Update(ctx, syncerConfig2Unst, v1.UpdateOptions{})
	require.NoError(t, err)

	t.Logf("Delete workloads from upstream")
	err = upstreamKubeClueterClient.AppsV1().Deployments().Cluster(wsPath).Namespace("optimized").Delete(ctx, "cpu-usage", v1.DeleteOptions{})
	require.NoError(t, err)

	t.Logf("Delete APIBinding from upstream")
	syncerFixture.DeleteRootComputeAPIBinding(t)

	t.Logf("Wait for deployments to be deleted from downstream too.")
	framework.Eventually(t, func() (bool, string) {
		_, err := downstreamKubeClient.CoreV1().Namespaces().Get(ctx, "optimized", v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get namespace %s: %v", "optimized", err)
		}
		_, err = downstreamKubeClient.AppsV1().Deployments("optimized").Get(ctx, "cpu-usage", v1.GetOptions{})
		if err == nil {
			return false, fmt.Sprintf("Still found deployment %s: %v", "cpu-usage", err)
		} else if !k8serrors.IsNotFound(err) {
			return false, fmt.Sprintf("Failed to get deployment %s: %v", "cpu-usage", err)
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "Deployments haven't been deleted from downstream yet.")
}
