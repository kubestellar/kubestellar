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

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/syncer/apis/edge/v1alpha1"
	edgefakeclient "github.com/kcp-dev/edge-mc/pkg/syncer/client/clientset/versioned/fake"
	edgeinformers "github.com/kcp-dev/edge-mc/pkg/syncer/client/informers/externalversions"
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
}

func (s *FakeSyncer) ReInitializeClients(resources []edgev1alpha1.EdgeSyncConfigResource, conversions []edgev1alpha1.EdgeSynConversion) error {
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

func TestSyncConfig(t *testing.T) {
	tests := map[string]struct {
		syncConfig         *edgev1alpha1.EdgeSyncConfig
		reInitializedCount int
		upSyncer           FakeSyncer
		downSyncer         FakeSyncer
	}{
		"Syncer updates downsyncer/upsyncer following to syncConfig": {
			syncConfig:         syncConfig("test-sync-config", types.UID("uid")),
			reInitializedCount: 1,
			upSyncer:           FakeSyncer{t: t},
			downSyncer:         FakeSyncer{t: t},
		},
		"Syncer does reconsiliation loop until the process succeeds or timeout": {
			syncConfig:         syncConfig("test-sync-config-for-error-case", types.UID("uid")),
			reInitializedCount: 2,
			upSyncer:           FakeSyncer{t: t, reInitializeClientsCallback: reInitializeClientsCallback},
			downSyncer:         FakeSyncer{t: t, reInitializeClientsCallback: reInitializeClientsCallback},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			logger := klog.FromContext(ctx)

			syncConfigClientSet := edgefakeclient.NewSimpleClientset(tc.syncConfig)
			syncConfigClient := syncConfigClientSet.EdgeV1alpha1().EdgeSyncConfigs()
			syncConfigInformerFactory := edgeinformers.NewSharedScopedInformerFactoryWithOptions(syncConfigClientSet, 0)
			syncConfigInformer := syncConfigInformerFactory.Edge().V1alpha1().EdgeSyncConfigs()

			controller, err := NewSyncConfigController(logger, syncConfigClient, syncConfigInformer, tc.syncConfig.Name, &tc.upSyncer, &tc.downSyncer, 1*time.Second)
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
