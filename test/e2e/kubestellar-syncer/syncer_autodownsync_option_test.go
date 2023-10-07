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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"

	kcpdynamic "github.com/kcp-dev/client-go/dynamic"
	"github.com/kcp-dev/kcp/test/e2e/framework"
	"github.com/kcp-dev/logicalcluster/v3"

	edgeframework "github.com/kubestellar/kubestellar/test/e2e/framework"
)

func TestKubeStellarAutoDownsyncOption(t *testing.T) {
	var syncerConfigUnst *unstructured.Unstructured
	err := edgeframework.LoadFile("testdata/autodownsync-option/syncer-config.yaml", embedded, &syncerConfigUnst)
	require.NoError(t, err)

	var cmUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/autodownsync-option/configmap.yaml", embedded, &cmUnst)
	require.NoError(t, err)

	var cmNotAutoDownSyncUnst *unstructured.Unstructured
	err = edgeframework.LoadFile("testdata/autodownsync-option/configmap-not-autodownsync.yaml", embedded, &cmNotAutoDownSyncUnst)
	require.NoError(t, err)

	framework.Suite(t, "kubestellar-syncer")

	test1KubeStellarAutoDownsyncOption(
		setupKubeStellarAutoDownsyncOptionWorkspaces(t, syncerConfigUnst, cmUnst, cmNotAutoDownSyncUnst),
	)
	test2KubeStellarAutoDownsyncOption(
		setupKubeStellarAutoDownsyncOptionWorkspaces(t, syncerConfigUnst, cmUnst, cmNotAutoDownSyncUnst),
	)
	test3KubeStellarAutoDownsyncOption(
		setupKubeStellarAutoDownsyncOptionWorkspaces(t, syncerConfigUnst, cmUnst, cmNotAutoDownSyncUnst),
	)
}

func test1KubeStellarAutoDownsyncOption(fixture *TestKubeStellarAutoDownsyncOptionFixture) {
	t := fixture.t
	ctx := fixture.ctx
	dynamicClient := fixture.dynamicClient
	cmTest1Unst := fixture.cmTest1Unst
	cmNotAutoDownSyncTest1Unst := fixture.cmNotAutoDownSyncTest1Unst
	cmTest2Unst := fixture.cmTest2Unst
	cmNotAutoDownSyncTest2Unst := fixture.cmNotAutoDownSyncTest2Unst

	updateConfigmapAtDownstreamAndCheckTheResults := func(value string) {
		t.Logf("Update configmap field at downstream")
		for _, obj := range []*unstructured.Unstructured{cmTest1Unst, cmNotAutoDownSyncTest1Unst, cmTest2Unst, cmNotAutoDownSyncTest2Unst} {
			fetched, err := dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), v1.GetOptions{})
			require.NoError(t, err)
			err = unstructured.SetNestedField(fetched.Object, value, "data", "field1")
			require.NoError(t, err)
			_, err = dynamicClient.Resource(configmapGvr).Namespace(fetched.GetNamespace()).Update(ctx, fetched, v1.UpdateOptions{})
			require.NoError(t, err)
		}

		t.Logf("Check configmaps with auto-downsync enabled are replaced by upstream while configmaps with auto-downsync disabled are not")
		fixture.checkConfigmapsValueAtDownstream([]*unstructured.Unstructured{cmTest1Unst, cmTest2Unst})
		for _, obj := range []*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst} {
			fetched, err := dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), v1.GetOptions{})
			require.NoError(t, err)
			val, _, _ := unstructured.NestedString(fetched.Object, "data", "field1")
			assert.Equal(t, val, value)
		}
	}

	updateConfigmapAtDownstreamAndCheckTheResults("bar update")

	t.Logf("Change auto-downsync from false to true at upstream")
	fixture.changeAutoDownsyncAtUpstream([]*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst}, "false", "true")
	fixture.checkConfigmapsAutoDownsyncAtDownstream([]*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst}, "true")

	t.Logf("Check configmaps with auto-downsync changed to enable are replaced by upstream")
	fixture.checkConfigmapsValueAtDownstream([]*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst})

	t.Logf("Change auto-downsync from true to false at upstream")
	fixture.changeAutoDownsyncAtUpstream([]*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst}, "true", "false")
	fixture.checkConfigmapsAutoDownsyncAtDownstream([]*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst}, "false")

	updateConfigmapAtDownstreamAndCheckTheResults("bar update2")
}

func test2KubeStellarAutoDownsyncOption(fixture *TestKubeStellarAutoDownsyncOptionFixture) {
	t := fixture.t
	ctx := fixture.ctx
	wsPath := fixture.wsPath
	dynamicClient := fixture.dynamicClient
	upstreamDynamicClueterClient := fixture.upstreamDynamicClueterClient
	cmTest1Unst := fixture.cmTest1Unst
	cmNotAutoDownSyncTest1Unst := fixture.cmNotAutoDownSyncTest1Unst
	cmTest2Unst := fixture.cmTest2Unst
	cmNotAutoDownSyncTest2Unst := fixture.cmNotAutoDownSyncTest2Unst

	t.Logf("Delete configmaps from upstream")
	for _, obj := range []*unstructured.Unstructured{cmTest1Unst, cmNotAutoDownSyncTest1Unst, cmTest2Unst, cmNotAutoDownSyncTest2Unst} {
		err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(configmapGvr).Namespace(obj.GetNamespace()).Delete(ctx, obj.GetName(), v1.DeleteOptions{})
		require.NoError(t, err)
	}

	t.Logf("Check configmaps with auto-downsync enabled are deleted while configmaps with auto-downsync disabled are not")
	fixture.checkConfigmapsDeleted([]*unstructured.Unstructured{cmTest1Unst, cmTest2Unst})
	for _, obj := range []*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst} {
		_, err := dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), v1.GetOptions{})
		assert.NoError(t, err, "Should exist")
	}
}

func test3KubeStellarAutoDownsyncOption(fixture *TestKubeStellarAutoDownsyncOptionFixture) {
	t := fixture.t
	ctx := fixture.ctx
	wsPath := fixture.wsPath
	upstreamDynamicClueterClient := fixture.upstreamDynamicClueterClient
	cmTest1Unst := fixture.cmTest1Unst
	cmNotAutoDownSyncTest1Unst := fixture.cmNotAutoDownSyncTest1Unst
	cmTest2Unst := fixture.cmTest2Unst
	cmNotAutoDownSyncTest2Unst := fixture.cmNotAutoDownSyncTest2Unst

	t.Logf("Change auto-downsync from false to true at upstream")
	fixture.changeAutoDownsyncAtUpstream([]*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst}, "false", "true")
	fixture.checkConfigmapsAutoDownsyncAtDownstream([]*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst}, "true")

	t.Logf("Delete configmaps from upstream")
	for _, obj := range []*unstructured.Unstructured{cmTest1Unst, cmNotAutoDownSyncTest1Unst, cmTest2Unst, cmNotAutoDownSyncTest2Unst} {
		err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(configmapGvr).Namespace(obj.GetNamespace()).Delete(ctx, obj.GetName(), v1.DeleteOptions{})
		require.NoError(t, err)
	}

	t.Logf("Check configmaps are deleted")
	fixture.checkConfigmapsDeleted([]*unstructured.Unstructured{cmNotAutoDownSyncTest1Unst, cmNotAutoDownSyncTest2Unst})
}

func setupKubeStellarAutoDownsyncOptionWorkspaces(t *testing.T, syncerConfigUnst *unstructured.Unstructured, cmUnst *unstructured.Unstructured, cmNotAutoDownSyncUnst *unstructured.Unstructured) *TestKubeStellarAutoDownsyncOptionFixture {
	syncerFixture := setup(t)
	wsPath := syncerFixture.WorkspacePath

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	upstreamDynamicClueterClient := syncerFixture.UpstreamDynamicKubeClient
	upstreamKubeClusterClient := syncerFixture.UpstreamKubeClusterClient

	t.Logf("Create a SyncerConfig for test in workspace %q.", wsPath.String())
	_, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(syncerConfigGvr).Create(ctx, syncerConfigUnst, v1.CreateOptions{})
	require.NoError(t, err)

	testNamespace := createNamespace(t, ctx, upstreamKubeClusterClient, wsPath, "test")
	test2Namespace := createNamespace(t, ctx, upstreamKubeClusterClient, wsPath, "test2")

	createNamespacedObject := func(gvr schema.GroupVersionResource, namespace string, unstObj *unstructured.Unstructured) *unstructured.Unstructured {
		t.Logf("Create %q %q in namespace %q in workspace %q.", gvr.GroupResource().String(), unstObj.GetName(), namespace, wsPath.String())
		ret, err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(gvr).Namespace(namespace).Create(ctx, unstObj, v1.CreateOptions{})
		require.NoError(t, err)
		return ret
	}

	cmTest1Unst := createNamespacedObject(configmapGvr, testNamespace.Name, cmUnst)
	cmNotAutoDownSyncTest1Unst := createNamespacedObject(configmapGvr, testNamespace.Name, cmNotAutoDownSyncUnst)
	cmTest2Unst := createNamespacedObject(configmapGvr, test2Namespace.Name, cmUnst)
	cmNotAutoDownSyncTest2Unst := createNamespacedObject(configmapGvr, test2Namespace.Name, cmNotAutoDownSyncUnst)

	client := syncerFixture.DownstreamKubeClient
	dynamicClient := syncerFixture.DownstreamDynamicKubeClient

	t.Logf("Wait for resources to be downsynced.")
	framework.Eventually(t, func() (bool, string) {
		for _, name := range []string{testNamespace.Name, test2Namespace.Name} {
			_, err := client.CoreV1().Namespaces().Get(ctx, name, v1.GetOptions{})
			if err != nil {
				return false, fmt.Sprintf("Failed to get namespace %s: %v", name, err)
			}
		}

		for _, obj := range []*unstructured.Unstructured{cmTest1Unst, cmNotAutoDownSyncTest1Unst, cmTest2Unst, cmNotAutoDownSyncTest2Unst} {
			_, err := dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), v1.GetOptions{})
			if err != nil {
				return false, fmt.Sprintf("Failed to get configmap %s: %v", obj.GetName(), err)
			}
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")

	return &TestKubeStellarAutoDownsyncOptionFixture{
		t:                            t,
		ctx:                          ctx,
		wsPath:                       wsPath,
		dynamicClient:                dynamicClient,
		upstreamDynamicClueterClient: upstreamDynamicClueterClient,
		cmTest1Unst:                  cmTest1Unst,
		cmNotAutoDownSyncTest1Unst:   cmNotAutoDownSyncTest1Unst,
		cmTest2Unst:                  cmTest2Unst,
		cmNotAutoDownSyncTest2Unst:   cmNotAutoDownSyncTest2Unst,
	}
}

type TestKubeStellarAutoDownsyncOptionFixture struct {
	t                            *testing.T
	ctx                          context.Context
	wsPath                       logicalcluster.Path
	dynamicClient                dynamic.Interface
	upstreamDynamicClueterClient *kcpdynamic.ClusterClientset
	cmTest1Unst                  *unstructured.Unstructured
	cmNotAutoDownSyncTest1Unst   *unstructured.Unstructured
	cmTest2Unst                  *unstructured.Unstructured
	cmNotAutoDownSyncTest2Unst   *unstructured.Unstructured
}

func (fixture *TestKubeStellarAutoDownsyncOptionFixture) checkConfigmapsSomething(resourceClient dynamic.NamespaceableResourceInterface, configmaps []*unstructured.Unstructured, callback func(fetched *unstructured.Unstructured) (bool, string), message string) {
	framework.Eventually(fixture.t, func() (bool, string) {
		for _, obj := range configmaps {
			fetched, err := resourceClient.Namespace(obj.GetNamespace()).Get(fixture.ctx, obj.GetName(), v1.GetOptions{})
			if err != nil {
				return false, fmt.Sprintf("Failed to get configmap %q in namespace %q: %v", fetched.GetName(), fetched.GetNamespace(), err)
			}
			ok, msg := callback(fetched)
			if !ok {
				return false, msg
			}
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, message)
}

func (fixture *TestKubeStellarAutoDownsyncOptionFixture) checkConfigmapsSomethingAtDownstream(configmaps []*unstructured.Unstructured, callback func(fetched *unstructured.Unstructured) (bool, string), message string) {
	fixture.checkConfigmapsSomething(fixture.dynamicClient.Resource(configmapGvr), configmaps, callback, message)
}

func (fixture *TestKubeStellarAutoDownsyncOptionFixture) checkConfigmapsValueAtDownstream(configmaps []*unstructured.Unstructured) {
	callback := func(fetched *unstructured.Unstructured) (bool, string) {
		val, ok, _ := unstructured.NestedString(fetched.Object, "data", "field1")
		if !(ok && val == "value1") {
			return false, fmt.Sprintf("Configmap %q in namespace %q hasn't been replaced by upstream one yet.", fetched.GetName(), fetched.GetNamespace())
		}
		return true, ""
	}
	fixture.checkConfigmapsSomethingAtDownstream(configmaps, callback, "Upsynced resources haven't been updated in upstream yet.")
}

func (fixture *TestKubeStellarAutoDownsyncOptionFixture) checkConfigmapsDeleted(configmaps []*unstructured.Unstructured) {
	framework.Eventually(fixture.t, func() (bool, string) {
		for _, obj := range configmaps {
			_, err := fixture.dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(fixture.ctx, obj.GetName(), v1.GetOptions{})
			if err == nil || !k8serrors.IsNotFound(err) {
				return false, fmt.Sprintf("Still found configmap %q in namespace %q: %v", obj.GetName(), obj.GetNamespace(), err)
			}
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "Deletion hasn't been propageted to downstream")
}

func (fixture *TestKubeStellarAutoDownsyncOptionFixture) changeAutoDownsyncAtUpstream(configmaps []*unstructured.Unstructured, prevVal string, newVal string) {
	for _, obj := range configmaps {
		fetched, err := fixture.upstreamDynamicClueterClient.Cluster(fixture.wsPath).Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(fixture.ctx, obj.GetName(), v1.GetOptions{})
		require.NoError(fixture.t, err)
		annotations := fetched.GetAnnotations()
		val, ok := annotations["edge.kubestellar.io/auto-downsync"]
		require.True(fixture.t, ok)
		require.Equal(fixture.t, val, prevVal)
		annotations["edge.kubestellar.io/auto-downsync"] = newVal
		fetched.SetAnnotations(annotations)
		_, err = fixture.upstreamDynamicClueterClient.Cluster(fixture.wsPath).Resource(configmapGvr).Namespace(fetched.GetNamespace()).Update(fixture.ctx, fetched, v1.UpdateOptions{})
		require.NoError(fixture.t, err)
	}
}

func (fixture *TestKubeStellarAutoDownsyncOptionFixture) checkConfigmapsAutoDownsyncAtDownstream(configmaps []*unstructured.Unstructured, value string) {
	callback := func(fetched *unstructured.Unstructured) (bool, string) {
		annotations := fetched.GetAnnotations()
		val, ok := annotations["edge.kubestellar.io/auto-downsync"]
		if !(ok && val == value) {
			return false, fmt.Sprintf("Auto-downsync of configmap %q in namespace %q is not %q.", fetched.GetName(), fetched.GetNamespace(), value)
		}
		return true, ""
	}
	fixture.checkConfigmapsSomethingAtDownstream(configmaps, callback, "AudoDownsync option hasn't been downsynced")
}
