/*
Copyright 2023 The KCP Authors.

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

package scheduler

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func (c *controller) reconcileOnEdgePlacement(ctx context.Context, key string) error {
	logger := klog.FromContext(ctx)
	ws, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(key)
	if err != nil {
		logger.Error(err, "invalid key")
		return err
	}
	logger = logger.WithValues("workspace", ws, "edgePlacement", name)
	ctx = klog.NewContext(ctx, logger)
	logger.V(2).Info("reconciling")

	// TODO(waltforme): should I use a client to bother the apiserver or use local store?
	ep, err := c.edgeClusterClient.EdgeV1alpha1().EdgePlacements().Cluster(ws.Path()).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	_, err = c.edgeClusterClient.EdgeV1alpha1().SinglePlacementSlices().Cluster(ws.Path()).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(1).Info("creating SinglePlacementSlice")
			sps := &edgev1alpha1.SinglePlacementSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: edgev1alpha1.SchemeGroupVersion.String(),
							Kind:       "EdgePlacement",
							Name:       name,
							UID:        ep.UID,
						},
					},
				},
				Destinations: []edgev1alpha1.SinglePlacement{},
			}
			_, err = c.edgeClusterClient.Cluster(ws.Path()).EdgeV1alpha1().SinglePlacementSlices().Create(ctx, sps, metav1.CreateOptions{})
			if err != nil {
				if !errors.IsAlreadyExists(err) {
					logger.Error(err, "failed creating singlePlacementSlice")
					return err
				}
			} else {
				logger.V(1).Info("created SinglePlacementSlice")
			}
		} else {
			logger.Error(err, "failed getting SinglePlacementSlice for EdgePlacement")
			return err
		}
	}

	return nil
}
