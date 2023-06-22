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

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kcp-dev/kcp/test/e2e/framework"

	edgeframework "github.com/kubestellar/kubestellar/test/e2e/framework"
)

func TestKubeStellarSyncerWithEdgeSyncConfig(t *testing.T) {

	var edgeSyncConfigUnst *unstructured.Unstructured
	err := edgeframework.LoadFile("testdata/edgesyncconfig/edge-sync-config.yaml", embedded, &edgeSyncConfigUnst)
	require.NoError(t, err)

	var sampleCRDUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/edgesyncconfig/sample-crd.yaml", embedded, &sampleCRDUnst)
	require.NoError(t, err)

	var sampleCRUpsyncUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/edgesyncconfig/sample-cr-upsync.yaml", embedded, &sampleCRUpsyncUnst)
	require.NoError(t, err)

	var sampleCRDownsyncUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/edgesyncconfig/sample-cr-downsync.yaml", embedded, &sampleCRDownsyncUnst)
	require.NoError(t, err)

	framework.Suite(t, "kubestellar-syncer")

	syncerFixture := setup(t)
	wsPath := syncerFixture.WorkspacePath

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	upstreamDynamicClueterClient := syncerFixture.UpstreamDynamicKubeClient
	upstreamKubeClusterClient := syncerFixture.UpstreamKubeClusterClient

	t.Logf("Create an EdgeSyncConfig for test in workspace %q.", wsPath.String())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(edgeSyncConfigGvr).Create(ctx, edgeSyncConfigUnst, v1.CreateOptions{})
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

	t.Logf("Create sample CR %q in workspace %q.", sampleCRDownsyncUnst.GetName(), wsPath.String())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(sampleCRGVR).Create(ctx, sampleCRDownsyncUnst, v1.CreateOptions{})
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
		_, err = dynamicClient.Resource(sampleCRGVR).Get(ctx, sampleCRDownsyncUnst.GetName(), v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get sample downsync CR %s: %v", sampleCRDownsyncUnst.GetName(), err)
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")

	t.Logf("Create sample CR %q in downstream cluster %q for upsyncing.", sampleCRUpsyncUnst.GetName(), wsPath.String())
	_, err = syncerFixture.DownstreamDynamicKubeClient.Resource(sampleCRGVR).Create(ctx, sampleCRUpsyncUnst, v1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("Wait for resources to be upsynced.")
	framework.Eventually(t, func() (bool, string) {
		_, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(sampleCRGVR).Get(ctx, sampleCRUpsyncUnst.GetName(), v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get sample CR %q in workspace %q: %v", sampleCRUpsyncUnst.GetName(), wsPath, err)
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All upsynced resources haven't been propagated to upstream yet.")
}
