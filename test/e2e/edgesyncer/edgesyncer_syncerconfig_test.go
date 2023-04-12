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

	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kcp-dev/kcp/test/e2e/framework"

	edgeframework "github.com/kcp-dev/edge-mc/test/e2e/framework"
)

var syncerConfigGvr = schema.GroupVersionResource{
	Group:    "edge.kcp.io",
	Version:  "v1alpha1",
	Resource: "syncerconfigs",
}

func TestEdgeSyncerWithSyncerConfig(t *testing.T) {

	var syncerConfigUnst *unstructured.Unstructured
	err := edgeframework.LoadFile("testdata/syncer-config.yaml", embedded, &syncerConfigUnst)
	require.NoError(t, err)

	var sampleCRDUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/sample-crd.yaml", embedded, &sampleCRDUnst)
	require.NoError(t, err)

	var sampleUpsyncCRUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/sample-upsync-cr.yaml", embedded, &sampleUpsyncCRUnst)
	require.NoError(t, err)

	var sampleDownsyncCRUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/sample-downsync-cr.yaml", embedded, &sampleDownsyncCRUnst)
	require.NoError(t, err)

	framework.Suite(t, "edge-syncer")

	syncerFixture := setup(t)
	wsPath := syncerFixture.WorkspacePath

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

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
	t.Logf("Create sample CRD %q in workspace %q.", sampleCRDUnst.GetName(), wsPath.String())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(crdGVR).Create(ctx, sampleCRDUnst, v1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("Wait for API %q to be available.", sampleCRDUnst.GetName())
	framework.Eventually(t, func() (bool, string) {
		_, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(sampleCRGVR).List(ctx, v1.ListOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to list sample CR: %v", err)
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*1, "API %q hasn't been available yet.", sampleCRDUnst.GetName())

	t.Logf("Create sample CR %q in workspace %q.", sampleDownsyncCRUnst.GetName(), wsPath.String())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(sampleCRGVR).Create(ctx, sampleDownsyncCRUnst, v1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("Wait for resources to be downsynced.")
	framework.Eventually(t, func() (bool, string) {
		client := syncerFixture.DownstreamKubeClient
		dynamicClient := syncerFixture.DownstreamDynamicKubeClient
		_, err := client.CoreV1().Namespaces().Get(ctx, testNamespaceObj.Name, v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get namespace %s: %v", testNamespaceObj.Name, err)
		}
		_, err = dynamicClient.Resource(crdGVR).Get(ctx, sampleCRDUnst.GetName(), v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get CRD %s: %v", sampleCRDUnst.GetName(), err)
		}
		_, err = dynamicClient.Resource(sampleCRGVR).Get(ctx, sampleDownsyncCRUnst.GetName(), v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get sample downsync CR %s: %v", sampleDownsyncCRUnst.GetName(), err)
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")

	// t.Logf("Create sample CR %q in downstream cluster %q for upsyncing.", sampleUpsyncCRUnst.GetName(), wsPath.String())
	// _, err = syncerFixture.DownstreamDynamicKubeClient.Resource(sampleCRGVR).Create(ctx, sampleUpsyncCRUnst, v1.CreateOptions{})
	// require.NoError(t, err)

	// t.Logf("Wait for resources to be upsynced.")
	// framework.Eventually(t, func() (bool, string) {
	// 	_, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(sampleCRGVR).Get(ctx, sampleUpsyncCRUnst.GetName(), v1.GetOptions{})
	// 	if err != nil {
	// 		return false, fmt.Sprintf("Failed to get sample CR %q in workspace %q: %v", sampleUpsyncCRUnst.GetName(), wsPath, err)
	// 	}
	// 	return true, ""
	// }, wait.ForeverTestTimeout, time.Second*5, "All upsynced resources haven't been propagated to upstream yet.")
}
