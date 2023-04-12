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

package controller

import (
	"context"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamic "k8s.io/client-go/dynamic/fake"
	clientset "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog/v2"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	syncerv1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgefakeclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/fake"
	edgeinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"

	"github.com/kcp-dev/edge-mc/pkg/syncer/clientfactory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testAPIResourceList = []*metav1.APIResourceList{
	{
		GroupVersion: corev1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "configmaps", Namespaced: true, Kind: "ConfigMap"},
			{Name: "namespaces", Namespaced: false, Kind: "Namespace"},
		},
	},
	{
		GroupVersion: appsv1.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{Name: "deployments", Namespaced: true, Kind: "Deployment"},
			{Name: "deployments/scale", Namespaced: true, Kind: "Scale", Group: "apps", Version: "v1"},
		},
	},
	{
		GroupVersion: "cheese.testing.k8s.io/v1",
		APIResources: []metav1.APIResource{
			{Name: "cheddars", Namespaced: true, Kind: "Cheddar"},
			{Name: "goudas", Namespaced: false, Kind: "Gouda"},
		},
	},
	{
		GroupVersion: "cheese.testing.k8s.io/v27alpha15",
		APIResources: []metav1.APIResource{
			{Name: "cheddars", Namespaced: true, Kind: "Cheddar"},
			{Name: "goudas", Namespaced: false, Kind: "Gouda"},
		},
	},
}

func TestSyncerConfig(t *testing.T) {
	tests := []struct {
		description      string
		op               string
		syncerConfig     *edgev1alpha1.SyncerConfig
		syncerConfigSpec edgev1alpha1.SyncerConfigSpec
		expected         Expected
	}{
		{
			description:  "Syncer updates downsyncer/upsyncer and synced resources following to syncConfig",
			syncerConfig: syncerConfig("test-sync-config", types.UID("uid")),
			syncerConfigSpec: edgev1alpha1.SyncerConfigSpec{
				NamespaceScope: edgev1alpha1.NamespaceScopeDownsyncs{
					Namespaces: []string{"default"},
					Resources: []edgev1alpha1.NamespaceScopeDownsyncResource{
						{
							GroupResource: metav1.GroupResource{Group: "", Resource: "configmaps"},
							APIVersion:    "v1",
						},
					},
				},
				ClusterScope: []edgev1alpha1.ClusterScopeDownsyncResource{
					{
						GroupResource: metav1.GroupResource{Group: "cheese.testing.k8s.io", Resource: "cheddars"},
						APIVersion:    "v1",
						Objects:       nil,
					},
				},
				Upsync: []edgev1alpha1.UpsyncSet{
					{
						APIGroup:   "cheese.testing.k8s.io",
						Resources:  []string{"*"},
						Namespaces: []string{"*"},
						Names:      []string{"*"},
					},
				},
			},
			expected: Expected{
				downSyncedResources: []syncerv1alpha1.EdgeSyncConfigResource{
					{Group: "", Version: "v1", Kind: "ConfigMap", Namespace: "default", Name: "*"},
					{Group: "", Version: "v1", Kind: "Namespace", Name: "default"},
					{Group: "cheese.testing.k8s.io", Version: "v1", Kind: "Cheddar", Name: "*"},
				},
				upSyncedResources: []syncerv1alpha1.EdgeSyncConfigResource{
					{Group: "cheese.testing.k8s.io", Version: "v1", Kind: "Cheddar", Namespace: "*", Name: "*"},
					{Group: "cheese.testing.k8s.io", Version: "v1", Kind: "Gouda", Name: "*"},
					{Group: "cheese.testing.k8s.io", Version: "v27alpha15", Kind: "Cheddar", Namespace: "*", Name: "*"},
					{Group: "cheese.testing.k8s.io", Version: "v27alpha15", Kind: "Gouda", Name: "*"},
				},
				conversions: []syncerv1alpha1.EdgeSynConversion{{
					Upstream:   upSyncedResource,
					Downstream: downSyncedResource,
				}},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			logger := klog.FromContext(ctx)

			tc.syncerConfig.Spec = tc.syncerConfigSpec
			syncerConfigClientSet := edgefakeclient.NewSimpleClientset(tc.syncerConfig)
			syncerConfigClient := syncerConfigClientSet.EdgeV1alpha1().SyncerConfigs()
			syncerConfigInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(syncerConfigClientSet, 0)
			syncerConfigInformer := syncerConfigInformerFactory.Edge().V1alpha1().SyncerConfigs()

			upstreamDynamicClient := dynamic.NewSimpleDynamicClient(scheme)
			upstreamClientSet := clientset.NewSimpleClientset()
			upstreamDiscoveryClient := upstreamClientSet.Discovery().(*fakediscovery.FakeDiscovery)

			upstreamDiscoveryClient.Resources = testAPIResourceList
			upstreamClientFactory, err := clientfactory.NewClientFactory(logger, upstreamDynamicClient, upstreamDiscoveryClient)
			require.NoError(t, err)

			downstreamDynamicClient := dynamic.NewSimpleDynamicClient(scheme)
			downstreamClientSet := clientset.NewSimpleClientset()
			downstreamDiscoveryClient := downstreamClientSet.Discovery().(*fakediscovery.FakeDiscovery)

			downstreamDiscoveryClient.Resources = testAPIResourceList
			downstreamClientFactory, err := clientfactory.NewClientFactory(logger, downstreamDynamicClient, downstreamDiscoveryClient)
			require.NoError(t, err)

			syncConfigManager := NewSyncConfigManager(klog.FromContext(context.TODO()))
			syncerConfigManager := NewSyncerConfigManager(logger, syncConfigManager, upstreamClientFactory, downstreamClientFactory)
			controller, err := NewSyncerConfigController(logger, syncerConfigClient, syncerConfigInformer, syncerConfigManager, 1*time.Second)
			require.NoError(t, err)

			syncerConfigInformerFactory.Start(ctx.Done())

			err = nil
			require.Eventually(t, func() bool {
				syncConfig, _err := syncerConfigInformer.Lister().Get(tc.syncerConfig.Name)
				if _err != nil {
					if errors.IsNotFound(_err) {
						return false
					} else {
						err = _err
						return true
					}
				}
				return syncConfig != nil
			}, wait.ForeverTestTimeout, 1*time.Second)
			assert.NoError(t, err)

			go controller.Run(ctx, 1)
			require.Eventually(t, func() bool {
				_, ok := syncerConfigManager.syncerConfigMap[tc.syncerConfig.Name]
				return ok
			}, wait.ForeverTestTimeout, 1*time.Second)

			syncerConfigManager.Refresh()
			require.Eventually(t, func() bool {
				count := 0
				for key := range syncConfigManager.syncConfigMap {
					if strings.Contains(key, tc.syncerConfig.Name) {
						count++
					}
				}
				return count == 3
			}, wait.ForeverTestTimeout, 1*time.Second)

			downsyncedResources := syncConfigManager.GetDownSyncedResources()
			assertEqualArrayWithouOrder(t, tc.expected.downSyncedResources, downsyncedResources)
			upsyncedResources := syncConfigManager.GetUpSyncedResources()
			assertEqualArrayWithouOrder(t, tc.expected.upSyncedResources, upsyncedResources)
		})
	}
}

func syncerConfig(name string, uid types.UID) *edgev1alpha1.SyncerConfig {
	return &edgev1alpha1.SyncerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			UID:  uid,
		},
	}
}

func assertEqualArrayWithouOrder[T any](t *testing.T, expectedArray []T, actualArray []T) {
	assert.Equal(t, len(expectedArray), len(actualArray))
	if len(expectedArray) > 0 {
		for _, expected := range expectedArray {
			equal := false
			for _, actual := range actualArray {
				equal = assert.ObjectsAreEqual(expected, actual) || equal
			}
			if !equal {
				assert.Equal(t, expectedArray, actualArray)
			}
		}
	}
}
