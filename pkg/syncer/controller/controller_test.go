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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
	edgefakeclient "github.com/kcp-dev/edge-mc/pkg/client/clientset/versioned/fake"
	edgeinformers "github.com/kcp-dev/edge-mc/pkg/client/informers/externalversions"
)

var scheme *runtime.Scheme

func init() {
	scheme = runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
}

type FakeSyncer struct {
	t                           *testing.T
	reInitializedCount          int
	reInitializeClientsCallback func(FakeSyncer) error
	passedSyncedResources       []edgev1alpha1.EdgeSyncConfigResource
}

func (s *FakeSyncer) ReInitializeClients(resources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	s.passedSyncedResources = resources
	s.reInitializedCount++
	if s.reInitializeClientsCallback != nil {
		return s.reInitializeClientsCallback(*s)
	} else {
		return nil
	}
}

func (s *FakeSyncer) SyncOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	return nil
}

func (s *FakeSyncer) BackStatusOne(resource edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
	return nil
}

func reInitializeClientsCallback(s FakeSyncer) error {
	if s.reInitializedCount > 2 {
		return nil
	}
	return fmt.Errorf("")
}

var upSyncedResource = edgev1alpha1.EdgeSyncConfigResource{
	Kind:      "uk1",
	Group:     "ug1",
	Version:   "uv1",
	Name:      "un1",
	Namespace: "uns1",
}
var downSyncedResource = edgev1alpha1.EdgeSyncConfigResource{
	Kind:      "dk1",
	Group:     "dg1",
	Version:   "dv1",
	Name:      "dn1",
	Namespace: "dns1",
}
var downSyncedResource2 = edgev1alpha1.EdgeSyncConfigResource{
	Kind:      "dk2",
	Group:     "dg2",
	Version:   "dv2",
	Name:      "dn2",
	Namespace: "dns2",
}

func TestSyncConfig(t *testing.T) {
	type Expected struct {
		downSyncedResources []edgev1alpha1.EdgeSyncConfigResource
		upSyncedResources   []edgev1alpha1.EdgeSyncConfigResource
		conversions         []edgev1alpha1.EdgeSynConversion
	}
	tests := []struct {
		description        string
		op                 string
		syncConfig         *edgev1alpha1.EdgeSyncConfig
		reInitializedCount int
		syncConfigSpec     edgev1alpha1.EdgeSyncConfigSpec
		upSyncer           FakeSyncer
		downSyncer         FakeSyncer
		expected           Expected
	}{
		{
			description: "Syncer updates downsyncer/upsyncer and synced resources following to syncConfig",
			syncConfig:  syncConfig("test-sync-config", types.UID("uid")),
			syncConfigSpec: edgev1alpha1.EdgeSyncConfigSpec{
				DownSyncedResources: []edgev1alpha1.EdgeSyncConfigResource{downSyncedResource},
				UpSyncedResources:   []edgev1alpha1.EdgeSyncConfigResource{upSyncedResource},
				Conversions: []edgev1alpha1.EdgeSynConversion{{
					Upstream:   upSyncedResource,
					Downstream: downSyncedResource,
				}},
			},
			reInitializedCount: 1,
			upSyncer:           FakeSyncer{t: t},
			downSyncer:         FakeSyncer{t: t},
			expected: Expected{
				downSyncedResources: []edgev1alpha1.EdgeSyncConfigResource{downSyncedResource},
				upSyncedResources:   []edgev1alpha1.EdgeSyncConfigResource{upSyncedResource},
				conversions: []edgev1alpha1.EdgeSynConversion{{
					Upstream:   upSyncedResource,
					Downstream: downSyncedResource,
				}},
			},
		},
		{
			description: "Syncer updates downsyncer/upsyncer and synced resources following to additional syncConfig",
			syncConfig:  syncConfig("test-sync-config-2", types.UID("uid-2")),
			syncConfigSpec: edgev1alpha1.EdgeSyncConfigSpec{
				DownSyncedResources: []edgev1alpha1.EdgeSyncConfigResource{downSyncedResource2},
			},
			reInitializedCount: 1,
			upSyncer:           FakeSyncer{t: t},
			downSyncer:         FakeSyncer{t: t},
			expected: Expected{
				downSyncedResources: []edgev1alpha1.EdgeSyncConfigResource{downSyncedResource, downSyncedResource2},
				upSyncedResources:   []edgev1alpha1.EdgeSyncConfigResource{upSyncedResource},
				conversions: []edgev1alpha1.EdgeSynConversion{{
					Upstream:   upSyncedResource,
					Downstream: downSyncedResource,
				}},
			},
		},
		{
			description: "Syncer updates downsyncer/upsyncer and deletes synced resources of the deleted syncConfig",
			op:          "delete",
			syncConfig:  syncConfig("test-sync-config-2", types.UID("uid-2")),
			syncConfigSpec: edgev1alpha1.EdgeSyncConfigSpec{
				DownSyncedResources: []edgev1alpha1.EdgeSyncConfigResource{downSyncedResource2},
			},
			reInitializedCount: 1,
			upSyncer:           FakeSyncer{t: t},
			downSyncer:         FakeSyncer{t: t},
			expected: Expected{
				downSyncedResources: []edgev1alpha1.EdgeSyncConfigResource{downSyncedResource},
				upSyncedResources:   []edgev1alpha1.EdgeSyncConfigResource{upSyncedResource},
				conversions: []edgev1alpha1.EdgeSynConversion{{
					Upstream:   upSyncedResource,
					Downstream: downSyncedResource,
				}},
			},
		},
		{
			description:        "(Error case) Syncer does reconsiliation loop until the process succeeds or timeout",
			syncConfig:         syncConfig("test-sync-config-for-error-case", types.UID("uid")),
			reInitializedCount: 2,
			upSyncer:           FakeSyncer{t: t, reInitializeClientsCallback: reInitializeClientsCallback},
			downSyncer:         FakeSyncer{t: t, reInitializeClientsCallback: reInitializeClientsCallback},
			expected: Expected{ // Nothing to change the down/up synced resources and converions since no SyncConfig is added
				downSyncedResources: []edgev1alpha1.EdgeSyncConfigResource{downSyncedResource},
				upSyncedResources:   []edgev1alpha1.EdgeSyncConfigResource{upSyncedResource},
				conversions: []edgev1alpha1.EdgeSynConversion{{
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

			tc.syncConfig.Spec = tc.syncConfigSpec
			syncConfigClientSet := edgefakeclient.NewSimpleClientset(tc.syncConfig)
			syncConfigClient := syncConfigClientSet.EdgeV1alpha1().EdgeSyncConfigs()
			syncConfigInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(syncConfigClientSet, 0)
			syncConfigInformer := syncConfigInformerFactory.Edge().V1alpha1().EdgeSyncConfigs()

			controller, err := NewEdgeSyncConfigController(logger, syncConfigClient, syncConfigInformer, &tc.upSyncer, &tc.downSyncer, 1*time.Second)
			require.NoError(t, err)

			syncConfigInformerFactory.Start(ctx.Done())

			err = nil
			require.Eventually(t, func() bool {
				syncConfig, _err := syncConfigInformer.Lister().Get(tc.syncConfig.Name)
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
				return tc.upSyncer.reInitializedCount == tc.reInitializedCount && tc.downSyncer.reInitializedCount == tc.reInitializedCount
			}, wait.ForeverTestTimeout, 1*time.Second)

			if tc.op == "delete" {
				err = syncConfigClient.Delete(ctx, tc.syncConfig.Name, metav1.DeleteOptions{})
				assert.NoError(t, err)
				require.Eventually(t, func() bool {
					return tc.upSyncer.reInitializedCount == tc.reInitializedCount+1 && tc.downSyncer.reInitializedCount == tc.reInitializedCount+1
				}, wait.ForeverTestTimeout, 1*time.Second)
			}

			assert.Equal(t, len(tc.expected.downSyncedResources), len(tc.downSyncer.passedSyncedResources))
			if len(tc.expected.downSyncedResources) > 0 {
				assert.Equal(t, tc.expected.downSyncedResources, tc.downSyncer.passedSyncedResources)
			}
			assert.Equal(t, len(tc.expected.upSyncedResources), len(tc.upSyncer.passedSyncedResources))
			if len(tc.expected.upSyncedResources) > 0 {
				assert.Equal(t, tc.expected.upSyncedResources, tc.upSyncer.passedSyncedResources)
			}
		})
	}
}

func syncConfig(name string, uid types.UID) *edgev1alpha1.EdgeSyncConfig {
	return &edgev1alpha1.EdgeSyncConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			UID:  uid,
		},
	}
}
