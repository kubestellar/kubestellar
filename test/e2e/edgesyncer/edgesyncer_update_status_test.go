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

package edgesyncer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kcp-dev/kcp/test/e2e/framework"

	edgeframework "github.com/kcp-dev/edge-mc/test/e2e/framework"
)

func TestEdgeSyncerForUpdateStatus(t *testing.T) {
	var syncerConfigUnst *unstructured.Unstructured
	err := edgeframework.LoadFile("testdata/update-status/syncer-config.yaml", embedded, &syncerConfigUnst)
	require.NoError(t, err)

	var deploymentUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/update-status/deployment.yaml", embedded, &deploymentUnst)
	require.NoError(t, err)

	var statusObj map[string]interface{}
	err = edgeframework.LoadFile("testdata/update-status/status.yaml", embedded, &statusObj)
	require.NoError(t, err)

	var status2Obj map[string]interface{}
	err = edgeframework.LoadFile("testdata/update-status/status2.yaml", embedded, &status2Obj)
	require.NoError(t, err)

	framework.Suite(t, "edge-syncer")

	syncerFixture := setup(t)
	wsPath := syncerFixture.WorkspacePath

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	downstreamKubeClient := syncerFixture.DownstreamKubeClient
	downstreamDynamicClient := syncerFixture.DownstreamDynamicKubeClient

	upstreamDynamicClueterClient := syncerFixture.UpstreamDynamicKubeClient
	upstreamKubeClusterClient := syncerFixture.UpstreamKubeClusterClient

	t.Logf("Create a SyncerConfig for test in workspace %q.", wsPath.String())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(syncerConfigGvr).Create(ctx, syncerConfigUnst, v1.CreateOptions{})
	require.NoError(t, err)

	testNamespaceObj := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
	}
	t.Logf("Create namespace %q in workspace %q.", testNamespaceObj.Name, wsPath.String())
	_, err = upstreamKubeClusterClient.Cluster(wsPath).CoreV1().Namespaces().Create(ctx, testNamespaceObj, v1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("Create deployment %q in workspace %q.", deploymentUnst.GetName(), wsPath.String())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(deploymentGvr).Namespace(deploymentUnst.GetNamespace()).Create(ctx, deploymentUnst, v1.CreateOptions{})
	require.NoError(t, err)

	t.Log("Wait for resources to be downsynced.")
	framework.Eventually(t, func() (bool, string) {
		_, err := downstreamKubeClient.CoreV1().Namespaces().Get(ctx, testNamespaceObj.Name, v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get namespace %s: %v", testNamespaceObj.Name, err)
		}
		_, err = downstreamDynamicClient.Resource(deploymentGvr).Namespace(deploymentUnst.GetNamespace()).Get(ctx, deploymentUnst.GetName(), v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get deployment %s: %v", deploymentUnst.GetName(), err)
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")

	updateStatus := func(statusObj map[string]interface{}) map[string]interface{} {
		fetched, _ := downstreamDynamicClient.Resource(deploymentGvr).Namespace(deploymentUnst.GetNamespace()).Get(ctx, deploymentUnst.GetName(), v1.GetOptions{})
		err = unstructured.SetNestedMap(fetched.Object, statusObj, "status")
		require.NoError(t, err)
		updated, err := downstreamDynamicClient.Resource(deploymentGvr).Namespace(deploymentUnst.GetNamespace()).UpdateStatus(ctx, fetched, v1.UpdateOptions{})
		require.NoError(t, err)
		updatedStatus, ok, err := unstructured.NestedMap(updated.Object, "status")
		require.NoError(t, err)
		require.True(t, ok)
		require.NotEmpty(t, updatedStatus)
		return updatedStatus
	}

	checkStatusUpsync := func(downstreamStatus map[string]interface{}) {
		framework.Eventually(t, func() (bool, string) {
			fetched, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(deploymentGvr).Namespace(deploymentUnst.GetNamespace()).Get(ctx, deploymentUnst.GetName(), v1.GetOptions{})
			if err != nil {
				return false, fmt.Sprintf("Failed to get deployment %s: %v", deploymentUnst.GetName(), err)
			}
			fetchedStatus, ok, err := unstructured.NestedMap(fetched.Object, "status")
			if err != nil || !ok {
				return false, fmt.Sprintf("Failed to get status %s: %v", deploymentUnst.GetName(), err)
			}
			return assert.ObjectsAreEqual(downstreamStatus, fetchedStatus), ""
		}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")
	}

	t.Log("Update status manually for test purpose")
	updatedStatus := updateStatus(statusObj)
	t.Log("Wait for resource status to be upsynced.")
	checkStatusUpsync(updatedStatus)

	t.Log("Update status manually for test purpose additionally")
	updatedStatus = updateStatus(status2Obj)
	t.Log("Wait for resource status to be upsynced.")
	checkStatusUpsync(updatedStatus)
}
