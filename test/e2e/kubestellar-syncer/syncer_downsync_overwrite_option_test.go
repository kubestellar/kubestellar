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

	"github.com/kubestellar/kubestellar/test/e2e/framework"
	"github.com/kubestellar/kubestellar/test/e2e/logicalcluster"
)

func TestKubeStellarDownsyncOverwriteOption(t *testing.T) {
	var syncerConfigUnst *unstructured.Unstructured
	err := framework.LoadFile("testdata/downsync-overwrite-option/syncer-config.yaml", embedded, &syncerConfigUnst)
	require.NoError(t, err)

	var cmUnst *unstructured.Unstructured
	err = framework.LoadFile("testdata/downsync-overwrite-option/configmap.yaml", embedded, &cmUnst)
	require.NoError(t, err)

	var cmNotDownsyncOverwriteUnst *unstructured.Unstructured
	err = framework.LoadFile("testdata/downsync-overwrite-option/configmap-not-downsync-overwrite.yaml", embedded, &cmNotDownsyncOverwriteUnst)
	require.NoError(t, err)

	framework.Suite(t, "kubestellar-syncer")

	test1KubeStellarDownsyncOverwriteOption(
		setupKubeStellarDownsyncOverwriteOptionWorkspaces(t, syncerConfigUnst, cmUnst, cmNotDownsyncOverwriteUnst),
	)
	test2KubeStellarDownsyncOverwriteOption(
		setupKubeStellarDownsyncOverwriteOptionWorkspaces(t, syncerConfigUnst, cmUnst, cmNotDownsyncOverwriteUnst),
	)
	test3KubeStellarDownsyncOverwriteOption(
		setupKubeStellarDownsyncOverwriteOptionWorkspaces(t, syncerConfigUnst, cmUnst, cmNotDownsyncOverwriteUnst),
	)
}

func test1KubeStellarDownsyncOverwriteOption(fixture *TestKubeStellarDownsyncOverwriteOptionFixture) {
	t := fixture.t
	ctx := fixture.ctx
	dynamicClient := fixture.dynamicClient
	cmTest1Unst := fixture.cmTest1Unst
	cmNotDownsyncOverwriteTest1Unst := fixture.cmNotDownsyncOverwriteTest1Unst
	cmTest2Unst := fixture.cmTest2Unst
	cmNotDownsyncOverwriteTest2Unst := fixture.cmNotDownsyncOverwriteTest2Unst

	updateConfigmapAtDownstreamAndCheckTheResults := func(value string) {
		t.Logf("Update configmap field at downstream")
		for _, obj := range []*unstructured.Unstructured{cmTest1Unst, cmNotDownsyncOverwriteTest1Unst, cmTest2Unst, cmNotDownsyncOverwriteTest2Unst} {
			fetched, err := dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), v1.GetOptions{})
			require.NoError(t, err)
			err = unstructured.SetNestedField(fetched.Object, value, "data", "field1")
			require.NoError(t, err)
			_, err = dynamicClient.Resource(configmapGvr).Namespace(fetched.GetNamespace()).Update(ctx, fetched, v1.UpdateOptions{})
			require.NoError(t, err)
		}

		t.Logf("Check configmaps with downsync-overwrite enabled are replaced by upstream while configmaps with downsync-overwrite disabled are not")
		fixture.checkConfigmapsValueAtDownstream([]*unstructured.Unstructured{cmTest1Unst, cmTest2Unst})
		for _, obj := range []*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst} {
			fetched, err := dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), v1.GetOptions{})
			require.NoError(t, err)
			val, _, _ := unstructured.NestedString(fetched.Object, "data", "field1")
			assert.Equal(t, val, value)
		}
	}

	updateConfigmapAtDownstreamAndCheckTheResults("bar update")

	t.Logf("Change downsync-overwrite from false to true at upstream")
	fixture.changeDownsyncOverwriteAtUpstream([]*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst}, "false", "true")
	fixture.checkConfigmapsDownsyncOverwriteAtDownstream([]*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst}, "true")

	t.Logf("Check configmaps with downsync-overwrite changed to enable are replaced by upstream")
	fixture.checkConfigmapsValueAtDownstream([]*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst})

	t.Logf("Change downsync-overwrite from true to false at upstream")
	fixture.changeDownsyncOverwriteAtUpstream([]*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst}, "true", "false")
	fixture.checkConfigmapsDownsyncOverwriteAtDownstream([]*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst}, "false")

	updateConfigmapAtDownstreamAndCheckTheResults("bar update2")
}

func test2KubeStellarDownsyncOverwriteOption(fixture *TestKubeStellarDownsyncOverwriteOptionFixture) {
	t := fixture.t
	ctx := fixture.ctx
	wsPath := fixture.wsPath
	dynamicClient := fixture.dynamicClient
	upstreamDynamicClueterClient := fixture.upstreamDynamicClueterClient
	cmTest1Unst := fixture.cmTest1Unst
	cmNotDownsyncOverwriteTest1Unst := fixture.cmNotDownsyncOverwriteTest1Unst
	cmTest2Unst := fixture.cmTest2Unst
	cmNotDownsyncOverwriteTest2Unst := fixture.cmNotDownsyncOverwriteTest2Unst

	t.Logf("Delete configmaps from upstream")
	for _, obj := range []*unstructured.Unstructured{cmTest1Unst, cmNotDownsyncOverwriteTest1Unst, cmTest2Unst, cmNotDownsyncOverwriteTest2Unst} {
		err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(configmapGvr).Namespace(obj.GetNamespace()).Delete(ctx, obj.GetName(), v1.DeleteOptions{})
		require.NoError(t, err)
	}

	t.Logf("Check configmaps with downsync-overwrite enabled are deleted while configmaps with downsync-overwrite disabled are not")
	fixture.checkConfigmapsDeleted([]*unstructured.Unstructured{cmTest1Unst, cmTest2Unst})
	for _, obj := range []*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst} {
		_, err := dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), v1.GetOptions{})
		assert.NoError(t, err, "Should exist")
	}
}

func test3KubeStellarDownsyncOverwriteOption(fixture *TestKubeStellarDownsyncOverwriteOptionFixture) {
	t := fixture.t
	ctx := fixture.ctx
	wsPath := fixture.wsPath
	upstreamDynamicClueterClient := fixture.upstreamDynamicClueterClient
	cmTest1Unst := fixture.cmTest1Unst
	cmNotDownsyncOverwriteTest1Unst := fixture.cmNotDownsyncOverwriteTest1Unst
	cmTest2Unst := fixture.cmTest2Unst
	cmNotDownsyncOverwriteTest2Unst := fixture.cmNotDownsyncOverwriteTest2Unst

	t.Logf("Change downsync-overwrite from false to true at upstream")
	fixture.changeDownsyncOverwriteAtUpstream([]*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst}, "false", "true")
	fixture.checkConfigmapsDownsyncOverwriteAtDownstream([]*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst}, "true")

	t.Logf("Delete configmaps from upstream")
	for _, obj := range []*unstructured.Unstructured{cmTest1Unst, cmNotDownsyncOverwriteTest1Unst, cmTest2Unst, cmNotDownsyncOverwriteTest2Unst} {
		err := upstreamDynamicClueterClient.Cluster(wsPath).Resource(configmapGvr).Namespace(obj.GetNamespace()).Delete(ctx, obj.GetName(), v1.DeleteOptions{})
		require.NoError(t, err)
	}

	t.Logf("Check configmaps are deleted")
	fixture.checkConfigmapsDeleted([]*unstructured.Unstructured{cmNotDownsyncOverwriteTest1Unst, cmNotDownsyncOverwriteTest2Unst})
}

func setupKubeStellarDownsyncOverwriteOptionWorkspaces(t *testing.T, syncerConfigUnst *unstructured.Unstructured, cmUnst *unstructured.Unstructured, cmNotDownsyncOverwriteUnst *unstructured.Unstructured) *TestKubeStellarDownsyncOverwriteOptionFixture {
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
	cmNotDownsyncOverwriteTest1Unst := createNamespacedObject(configmapGvr, testNamespace.Name, cmNotDownsyncOverwriteUnst)
	cmTest2Unst := createNamespacedObject(configmapGvr, test2Namespace.Name, cmUnst)
	cmNotDownsyncOverwriteTest2Unst := createNamespacedObject(configmapGvr, test2Namespace.Name, cmNotDownsyncOverwriteUnst)

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

		for _, obj := range []*unstructured.Unstructured{cmTest1Unst, cmNotDownsyncOverwriteTest1Unst, cmTest2Unst, cmNotDownsyncOverwriteTest2Unst} {
			_, err := dynamicClient.Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(ctx, obj.GetName(), v1.GetOptions{})
			if err != nil {
				return false, fmt.Sprintf("Failed to get configmap %s: %v", obj.GetName(), err)
			}
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")

	return &TestKubeStellarDownsyncOverwriteOptionFixture{
		t:                               t,
		ctx:                             ctx,
		wsPath:                          wsPath,
		dynamicClient:                   dynamicClient,
		upstreamDynamicClueterClient:    upstreamDynamicClueterClient,
		cmTest1Unst:                     cmTest1Unst,
		cmNotDownsyncOverwriteTest1Unst: cmNotDownsyncOverwriteTest1Unst,
		cmTest2Unst:                     cmTest2Unst,
		cmNotDownsyncOverwriteTest2Unst: cmNotDownsyncOverwriteTest2Unst,
	}
}

type TestKubeStellarDownsyncOverwriteOptionFixture struct {
	t                               *testing.T
	ctx                             context.Context
	wsPath                          logicalcluster.Path
	dynamicClient                   dynamic.Interface
	upstreamDynamicClueterClient    *framework.KcpDynamicClient
	cmTest1Unst                     *unstructured.Unstructured
	cmNotDownsyncOverwriteTest1Unst *unstructured.Unstructured
	cmTest2Unst                     *unstructured.Unstructured
	cmNotDownsyncOverwriteTest2Unst *unstructured.Unstructured
}

func (fixture *TestKubeStellarDownsyncOverwriteOptionFixture) checkConfigmapsSomething(resourceClient dynamic.NamespaceableResourceInterface, configmaps []*unstructured.Unstructured, callback func(fetched *unstructured.Unstructured) (bool, string), message string) {
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

func (fixture *TestKubeStellarDownsyncOverwriteOptionFixture) checkConfigmapsSomethingAtDownstream(configmaps []*unstructured.Unstructured, callback func(fetched *unstructured.Unstructured) (bool, string), message string) {
	fixture.checkConfigmapsSomething(fixture.dynamicClient.Resource(configmapGvr), configmaps, callback, message)
}

func (fixture *TestKubeStellarDownsyncOverwriteOptionFixture) checkConfigmapsValueAtDownstream(configmaps []*unstructured.Unstructured) {
	callback := func(fetched *unstructured.Unstructured) (bool, string) {
		val, ok, _ := unstructured.NestedString(fetched.Object, "data", "field1")
		if !(ok && val == "value1") {
			return false, fmt.Sprintf("Configmap %q in namespace %q hasn't been replaced by upstream one yet.", fetched.GetName(), fetched.GetNamespace())
		}
		return true, ""
	}
	fixture.checkConfigmapsSomethingAtDownstream(configmaps, callback, "Upsynced resources haven't been updated in upstream yet.")
}

func (fixture *TestKubeStellarDownsyncOverwriteOptionFixture) checkConfigmapsDeleted(configmaps []*unstructured.Unstructured) {
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

func (fixture *TestKubeStellarDownsyncOverwriteOptionFixture) changeDownsyncOverwriteAtUpstream(configmaps []*unstructured.Unstructured, prevVal string, newVal string) {
	for _, obj := range configmaps {
		fetched, err := fixture.upstreamDynamicClueterClient.Cluster(fixture.wsPath).Resource(configmapGvr).Namespace(obj.GetNamespace()).Get(fixture.ctx, obj.GetName(), v1.GetOptions{})
		require.NoError(fixture.t, err)
		annotations := fetched.GetAnnotations()
		val, ok := annotations["edge.kubestellar.io/downsync-overwrite"]
		require.True(fixture.t, ok)
		require.Equal(fixture.t, val, prevVal)
		annotations["edge.kubestellar.io/downsync-overwrite"] = newVal
		fetched.SetAnnotations(annotations)
		_, err = fixture.upstreamDynamicClueterClient.Cluster(fixture.wsPath).Resource(configmapGvr).Namespace(fetched.GetNamespace()).Update(fixture.ctx, fetched, v1.UpdateOptions{})
		require.NoError(fixture.t, err)
	}
}

func (fixture *TestKubeStellarDownsyncOverwriteOptionFixture) checkConfigmapsDownsyncOverwriteAtDownstream(configmaps []*unstructured.Unstructured, value string) {
	callback := func(fetched *unstructured.Unstructured) (bool, string) {
		annotations := fetched.GetAnnotations()
		val, ok := annotations["edge.kubestellar.io/downsync-overwrite"]
		if !(ok && val == value) {
			return false, fmt.Sprintf("downsync-overwrite of configmap %q in namespace %q is not %q.", fetched.GetName(), fetched.GetNamespace(), value)
		}
		return true, ""
	}
	fixture.checkConfigmapsSomethingAtDownstream(configmaps, callback, "downsync-overwrite annotation hasn't been downsynced")
}
