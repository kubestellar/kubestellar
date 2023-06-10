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

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kcp-dev/kcp/test/e2e/framework"

	edgeframework "github.com/kcp-dev/edge-mc/test/e2e/framework"
)

func TestKubeStellarSyncerForKyvernoWithSyncerConfig(t *testing.T) {

	var syncerConfigUnst *unstructured.Unstructured
	err := edgeframework.LoadFile("testdata/kyverno/syncer-config.yaml", embedded, &syncerConfigUnst)
	require.NoError(t, err)

	var policyReportUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/kyverno/policy-report.yaml", embedded, &policyReportUnst)
	require.NoError(t, err)

	var clusterPolicyReportUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/kyverno/cluster-policy-report.yaml", embedded, &clusterPolicyReportUnst)
	require.NoError(t, err)

	framework.Suite(t, "kubestellar-syncer")

	syncerFixture := setup(t)
	wsPath := syncerFixture.WorkspacePath

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	upstreamDynamicClueterClient := syncerFixture.UpstreamDynamicKubeClient

	t.Logf("Deploy kyverno resources 1 in workspace %q.", wsPath.String())
	framework.Kubectl(t, syncerFixture.UpstreamKubeconfigPath, "create", "-f", pathToTestdataDir()+"/kyverno/resources1.yaml")

	framework.Eventually(t, func() (bool, string) {
		_, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(policyGvr).Namespace("default").List(ctx, v1.ListOptions{})
		return err == nil, "Waiting Kyverno Policy API is ready"
	}, wait.ForeverTestTimeout, time.Second, "Kyverno Policy API has't been ready.")

	t.Logf("Deploy kyverno resources 2 in workspace %q.", wsPath.String())
	framework.Kubectl(t, syncerFixture.UpstreamKubeconfigPath, "create", "-f", pathToTestdataDir()+"/kyverno/resources2.yaml")

	t.Logf("Create a SyncerConfig for test in workspace %q.", wsPath.String())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(syncerConfigGvr).Create(ctx, syncerConfigUnst, v1.CreateOptions{})
	require.NoError(t, err)

	downstreamKubeClient := syncerFixture.DownstreamKubeClient
	downstreamDynamicClient := syncerFixture.DownstreamDynamicKubeClient

	t.Logf("Wait for resources to be downsynced.")
	framework.Eventually(t, func() (bool, string) {
		_, err := downstreamKubeClient.CoreV1().Namespaces().Get(ctx, "kyverno", v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get namespace %s: %v", "kyverno", err)
		}
		_, err = downstreamKubeClient.CoreV1().Namespaces().Get(ctx, "policy", v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get namespace %s: %v", "policy", err)
		}
		_, err = downstreamDynamicClient.Resource(policyGvr).Namespace("policy").Get(ctx, "sample-policy", v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get policy %s: %v", "sample-policy", err)
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")

	t.Logf("Create a policy report for upsync test in downstream")
	_, err = downstreamDynamicClient.Resource(policyReportGvr).Namespace(policyReportUnst.GetNamespace()).Create(ctx, policyReportUnst, v1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("Create a cluster policy report for upsync test in downstream")
	_, err = downstreamDynamicClient.Resource(clusterPolicyReportGvr).Create(ctx, clusterPolicyReportUnst, v1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("Wait for resources to be upsynced.")
	framework.Eventually(t, func() (bool, string) {
		fetched, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(policyReportGvr).Namespace(policyReportUnst.GetNamespace()).Get(ctx, policyReportUnst.GetName(), v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get policy report %q in workspace %q: %v", policyReportUnst.GetName(), wsPath, err)
		}
		t.Logf("fetched upsynced policy report: %v", fetched)
		fetched, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(clusterPolicyReportGvr).Get(ctx, clusterPolicyReportUnst.GetName(), v1.GetOptions{})
		if err != nil {
			return false, fmt.Sprintf("Failed to get cluster policy report %q in workspace %q: %v", clusterPolicyReportUnst.GetName(), wsPath, err)
		}
		t.Logf("fetched upsynced cluster policy report: %v", fetched)
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All upsynced resources haven't been propagated to upstream yet.")
}
