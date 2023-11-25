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

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kcp-dev/logicalcluster/v3"
	"github.com/kubestellar/kubestellar/test/e2e/framework"
)

//go:embed testdata/*
var embedded embed.FS

var edgeSyncConfigGvr = schema.GroupVersionResource{
	Group:    "edge.kubestellar.io",
	Version:  "v2alpha1",
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

var configmapGvr = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "configmaps",
}

var secretGvr = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "secrets",
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
	Group:    "edge.kubestellar.io",
	Version:  "v2alpha1",
	Resource: "syncerconfigs",
}

func setup(t *testing.T) *framework.StartedKubeStellarSyncerFixture {
	framework.Suite(t, "kubestellar-syncer")

	upstreamServer := framework.SharedKcpServer(t)

	t.Log("Creating a workspace")
	wsPath, _ := framework.NewWorkspaceFixture(t, upstreamServer, logicalcluster.NewPath("root"), "e2e-tests-")
	syncerFixture := framework.NewKubeStellarSyncerFixture(t, upstreamServer, wsPath).CreateEdgeSyncTargetAndApplyToDownstream(t).RunSyncer(t)

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

func createNamespace(t *testing.T, ctx context.Context, upstreamKubeClusterClient *framework.KcpClusterClient, wsPath logicalcluster.Path, namespaceName string) *corev1.Namespace {
	namespaceObj := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: namespaceName,
		},
	}
	t.Logf("Create namespace %q in workspace %q.", namespaceObj.Name, wsPath.String())
	_, err := upstreamKubeClusterClient.Cluster(wsPath).CoreV1().Namespaces().Create(ctx, namespaceObj, v1.CreateOptions{})
	require.NoError(t, err)
	return namespaceObj
}
