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
	"embed"
	"path"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kcp-dev/kcp/test/e2e/framework"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeframework "github.com/kcp-dev/edge-mc/test/e2e/framework"
)

//go:embed testdata/*
var embedded embed.FS

var edgeSyncConfigGvr = schema.GroupVersionResource{
	Group:    "edge.kcp.io",
	Version:  "v1alpha1",
	Resource: "edgesyncconfigs",
}

var crdGVR = schema.GroupVersionResource{
	Group:    "apiextensions.k8s.io",
	Version:  "v1",
	Resource: "customresourcedefinitions",
}

var sampleCRGVR = schema.GroupVersionResource{
	Group:    "my.domain",
	Version:  "v1alpha1",
	Resource: "samples",
}

var sampleDownsyncCRGVR = schema.GroupVersionResource{
	Group:    "my.domain",
	Version:  "v1alpha1",
	Resource: "sampledownsyncs",
}

var sampleUpsyncCRGVR = schema.GroupVersionResource{
	Group:    "my.domain",
	Version:  "v1alpha1",
	Resource: "sampleupsyncs",
}

var deploymentGvr = schema.GroupVersionResource{
	Group:    "apps",
	Version:  "v1",
	Resource: "deployments",
}

var policyGvr = schema.GroupVersionResource{
	Group:    "kyverno.io",
	Version:  "v1",
	Resource: "policies",
}

var policyReportGvr = schema.GroupVersionResource{
	Group:    "wgpolicyk8s.io",
	Version:  "v1alpha2",
	Resource: "policyreports",
}

var clusterPolicyReportGvr = schema.GroupVersionResource{
	Group:    "wgpolicyk8s.io",
	Version:  "v1alpha2",
	Resource: "clusterpolicyreports",
}

var syncerConfigGvr = schema.GroupVersionResource{
	Group:    "edge.kcp.io",
	Version:  "v1alpha1",
	Resource: "syncerconfigs",
}

func setup(t *testing.T) *edgeframework.StartedKubeStellarSyncerFixture {
	framework.Suite(t, "kubestellar-syncer")

	upstreamServer := framework.SharedKcpServer(t)

	t.Log("Creating an organization")
	orgPath, _ := framework.NewOrganizationFixture(t, upstreamServer, framework.TODO_WithoutMultiShardSupport())

	t.Log("Creating a workspace")
	_, ws := framework.NewWorkspaceFixture(t, upstreamServer, orgPath, framework.WithName("upstream-sink"), framework.TODO_WithoutMultiShardSupport())
	wsPath := logicalcluster.NewPath(logicalcluster.Name(ws.Spec.Cluster).String())
	syncerFixture := edgeframework.NewKubeStellarSyncerFixture(t, upstreamServer, wsPath).CreateEdgeSyncTargetAndApplyToDownstream(t).RunSyncer(t)

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	t.Logf("Confirm that a default EdgeSyncConfig is created.")
	unstList, err := syncerFixture.UpstreamDynamicKubeClient.Cluster(wsPath).Resource(edgeSyncConfigGvr).List(ctx, v1.ListOptions{})
	require.NoError(t, err)
	require.Greater(t, len(unstList.Items), 0)
	return syncerFixture
}

func pathToTestdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "./testdata")
	return dir
}
