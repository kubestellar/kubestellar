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

	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kubestellar/kubestellar/test/e2e/framework"
)

func TestKubeStellarSyncerForNamespacedObjects(t *testing.T) {
	var syncerConfigUnst *unstructured.Unstructured
	err := framework.LoadFile("testdata/namespaced-objects/syncer-config.yaml", embedded, &syncerConfigUnst)
	require.NoError(t, err)
	testKubeStellarSyncerForNamespacedObjects(t, syncerConfigUnst)
}

func testKubeStellarSyncerForNamespacedObjects(t *testing.T, syncerConfigUnst *unstructured.Unstructured) {

	var err error
	testDataDirectory := "testdata/namespaced-objects"

	type typeUnstObjWithPath struct {
		UnstObj *unstructured.Unstructured
		Path    string
	}
	var test1cm1, test1cm2, test2cm1, test2cm2, test3cm1, test1secret, test2secret, test3secret unstructured.Unstructured
	_test1cm1 := typeUnstObjWithPath{UnstObj: &test1cm1, Path: testDataDirectory + "/configmap-test1-cm1.yaml"}
	_test1cm2 := typeUnstObjWithPath{UnstObj: &test1cm2, Path: testDataDirectory + "/configmap-test1-cm2.yaml"}
	_test2cm1 := typeUnstObjWithPath{UnstObj: &test2cm1, Path: testDataDirectory + "/configmap-test2-cm1.yaml"}
	_test2cm2 := typeUnstObjWithPath{UnstObj: &test2cm2, Path: testDataDirectory + "/configmap-test2-cm2.yaml"}
	_test3cm1 := typeUnstObjWithPath{UnstObj: &test3cm1, Path: testDataDirectory + "/configmap-test3-cm1.yaml"}
	_test1secret := typeUnstObjWithPath{UnstObj: &test1secret, Path: testDataDirectory + "/secret-test1.yaml"} // pragma: allowlist secret
	_test2secret := typeUnstObjWithPath{UnstObj: &test2secret, Path: testDataDirectory + "/secret-test2.yaml"} // pragma: allowlist secret
	_test3secret := typeUnstObjWithPath{UnstObj: &test3secret, Path: testDataDirectory + "/secret-test3.yaml"} // pragma: allowlist secret
	for _, test := range []*typeUnstObjWithPath{&_test1cm1, &_test1cm2, &_test2cm1, &_test2cm2, &_test3cm1, &_test1secret, &_test2secret, &_test3secret} {
		err := framework.LoadFile(test.Path, embedded, &test.UnstObj)
		require.NoError(t, err)
	}
	framework.Suite(t, "kubestellar-syncer")

	syncerFixture := setup(t)
	wsPath := syncerFixture.WorkspacePath

	ctx, cancelFunc := context.WithCancel(context.Background())
	t.Cleanup(cancelFunc)

	upstreamDynamicClueterClient := syncerFixture.UpstreamDynamicKubeClient
	upstreamKubeClusterClient := syncerFixture.UpstreamKubeClusterClient

	t.Logf("Create a SyncerConfig for test in workspace %q.", wsPath.String())
	_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(syncerConfigGvr).Create(ctx, syncerConfigUnst, v1.CreateOptions{})
	require.NoError(t, err)

	t.Logf("Create namespaces for test in workspace %q.", wsPath.String())
	test1NamespaceObj := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "test1",
		},
	}
	test2NamespaceObj := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "test2",
		},
	}
	test3NamespaceObj := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "test3",
		},
	}
	testShouldNotSyncedNamespaceObj := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "should-not-synced",
		},
	}
	for _, testNamespaceObj := range []*corev1.Namespace{test1NamespaceObj, test2NamespaceObj, test3NamespaceObj, testShouldNotSyncedNamespaceObj} {
		t.Logf("Create namespace %q in workspace %q.", testNamespaceObj.Name, wsPath.String())
		_, err = upstreamKubeClusterClient.Cluster(wsPath).CoreV1().Namespaces().Create(ctx, testNamespaceObj, v1.CreateOptions{})
		require.NoError(t, err)
	}

	t.Logf("Create test configmaps in workspace %q.", wsPath.String())
	for _, configmap := range []unstructured.Unstructured{test1cm1, test1cm2, test2cm1, test2cm2, test3cm1} {
		_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(configmapGvr).Namespace(configmap.GetNamespace()).Create(ctx, &configmap, v1.CreateOptions{})
		require.NoError(t, err)
	}

	t.Logf("Create test secrets in workspace %q.", wsPath.String())
	for _, secret := range []unstructured.Unstructured{test1secret, test2secret, test3secret} { // pragma: allowlist secret
		_, err = upstreamDynamicClueterClient.Cluster(wsPath).Resource(secretGvr).Namespace(secret.GetNamespace()).Create(ctx, &secret, v1.CreateOptions{})
		require.NoError(t, err)
	}

	client := syncerFixture.DownstreamKubeClient
	dynamicClient := syncerFixture.DownstreamDynamicKubeClient

	t.Logf("Wait for resources to be downsynced.")
	framework.Eventually(t, func() (bool, string) {
		t.Logf("  Check namespaces to be downsynced.")
		for _, testNamespaceObj := range []*corev1.Namespace{test1NamespaceObj, test2NamespaceObj, test3NamespaceObj} {
			_, err := client.CoreV1().Namespaces().Get(ctx, testNamespaceObj.Name, v1.GetOptions{})
			if err != nil {
				return false, fmt.Sprintf("Failed to get namespace %s: %v", testNamespaceObj.Name, err)
			}
		}

		t.Logf("  Check required configmaps to be downsynced.")
		for _, testCm := range []unstructured.Unstructured{test1cm1, test1cm2, test2cm1} {
			_, err := dynamicClient.Resource(configmapGvr).Namespace(testCm.GetNamespace()).Get(ctx, testCm.GetName(), v1.GetOptions{})
			if err != nil {
				return false, fmt.Sprintf("Failed to get configmap %s: %v", testCm.GetName(), err)
			}
		}

		t.Logf("  Check required secrets to be downsynced.")
		for _, testSecret := range []unstructured.Unstructured{test1secret, test3secret} { // pragma: allowlist secret
			_, err := dynamicClient.Resource(secretGvr).Namespace(testSecret.GetNamespace()).Get(ctx, testSecret.GetName(), v1.GetOptions{})
			if err != nil {
				return false, fmt.Sprintf("Failed to get secret %s: %v", testSecret.GetName(), err)
			}
		}
		return true, ""
	}, wait.ForeverTestTimeout, time.Second*5, "All downsynced resources haven't been propagated to downstream yet.")

	t.Logf("Check should-not-synced namespace doesn't exist in downstream.")
	_, errShoultNotSyncedNamespace := client.CoreV1().Namespaces().Get(ctx, testShouldNotSyncedNamespaceObj.Name, v1.GetOptions{})
	assert.NotNil(t, errShoultNotSyncedNamespace)
	assert.Equal(t, true, k8serror.IsNotFound(errShoultNotSyncedNamespace))

	t.Logf("Check excluded configmaps not to be downsynced.")
	for _, testCm := range []unstructured.Unstructured{test2cm2, test3cm1} {
		_, err := dynamicClient.Resource(configmapGvr).Namespace(testCm.GetNamespace()).Get(ctx, testCm.GetName(), v1.GetOptions{})
		assert.NotNil(t, err)
		assert.Equal(t, true, k8serror.IsNotFound(err))
	}

	t.Logf("Check excluded secrets not to be downsynced.")
	for _, testSecret := range []unstructured.Unstructured{test2secret} { // pragma: allowlist secret
		_, err := dynamicClient.Resource(secretGvr).Namespace(testSecret.GetNamespace()).Get(ctx, testSecret.GetName(), v1.GetOptions{})
		assert.NotNil(t, err)
		assert.Equal(t, true, k8serror.IsNotFound(err))
	}
}
