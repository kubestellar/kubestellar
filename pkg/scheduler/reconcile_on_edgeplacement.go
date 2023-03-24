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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func (c *controller) reconcileOnEdgePlacement(ctx context.Context, epKey string) error {
	logger := klog.FromContext(ctx)
	epws, _, epName, err := kcpcache.SplitMetaClusterNamespaceKey(epKey)
	if err != nil {
		logger.Error(err, "invalid EdgePlacement key")
		return err
	}
	logger = logger.WithValues("workspace", epws, "edgePlacement", epName)
	ctx = klog.NewContext(ctx, logger)
	logger.V(2).Info("reconciling")

	/*
		On EdgePlacement 'ep' change:
		1) from cache, find loc(s) that being selected by ep

		2) from cache, for each of the found loc, find st(s) that being selected by the loc

		3) update store, with loc(s) that being selected by ep

		4) update apiserver

		Need data structure: none.
	*/

	store.l.Lock()
	defer store.l.Unlock() // TODO(waltforme): Is it safe to shorten the critical section?

	ep, err := c.edgePlacementLister.Cluster(epws).Get(epName)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(1).Info("EdgePlacement not found")
			logger.V(3).Info("dropping EdgePlacement from store")
			store.dropEp(epKey)
			return nil
		} else {
			logger.Error(err, "failed to get EdgePlacement")
			return err
		}
	}

	// 1)
	locsAll, err := c.locationLister.List(labels.Everything())
	if err != nil {
		logger.Error(err, "failed to list Locations in all workspaces")
		return err
	}
	locsFilteredByEp, err := filterLocsByEp(locsAll, ep)
	if err != nil {
		logger.Error(err, "failed to find Locations for EdgePlacement")
	}
	locsSelecting := packLocKeys(locsFilteredByEp)

	singles := []edgev1alpha1.SinglePlacement{}
	for _, loc := range locsFilteredByEp {
		// 2)
		lws := logicalcluster.From(loc)
		stsInLws, err := c.synctargetLister.Cluster(lws).List(labels.Everything())
		if err != nil {
			logger.Error(err, "failed to list SyncTargets in Location workspace", "locationWorkspace", lws.String())
			return err
		}
		stsSelecting, err := filterStsByLoc(stsInLws, loc)
		if err != nil {
			logger.Error(err, "failed to find SyncTargets for Location", "locationWorkspace", lws.String(), "location", loc.Name)
			return err
		}
		singles = append(singles, makeSinglePlacementsForLoc(loc, stsSelecting)...)
	}

	// 3)
	for loc, eps := range store.epsBySelectedLoc {
		if _, ok := locsSelecting[loc]; !ok {
			delete(eps, epKey)
		}
	}
	for loc := range locsSelecting {
		if store.epsBySelectedLoc[loc] == nil {
			store.epsBySelectedLoc[loc] = map[string]empty{epKey: {}}
		} else {
			store.epsBySelectedLoc[loc][epKey] = empty{}
		}
	}

	// 4)
	currentSPS, err := c.singlePlacementSliceLister.Cluster(epws).Get(epName)
	if err != nil {
		if errors.IsNotFound(err) { // create
			logger.V(1).Info("creating SinglePlacementSlice")
			sps := &edgev1alpha1.SinglePlacementSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name: epName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: edgev1alpha1.SchemeGroupVersion.String(),
							Kind:       "EdgePlacement",
							Name:       epName,
							UID:        ep.UID,
						},
					},
				},
				Destinations: singles,
			}
			_, err = c.edgeClusterClient.Cluster(epws.Path()).EdgeV1alpha1().SinglePlacementSlices().Create(ctx, sps, metav1.CreateOptions{})
			if err != nil {
				if !errors.IsAlreadyExists(err) {
					logger.Error(err, "failed creating SinglePlacementSlice")
					return err
				}
			} else {
				logger.V(1).Info("created SinglePlacementSlice")
			}
		} else {
			logger.Error(err, "failed getting SinglePlacementSlice")
			return err
		}
	} else { // update
		currentSPS.Destinations = singles
		_, err = c.edgeClusterClient.Cluster(epws.Path()).EdgeV1alpha1().SinglePlacementSlices().Update(ctx, currentSPS, metav1.UpdateOptions{})
		if err != nil {
			logger.Error(err, "failed updating SinglePlacementSlice")
			return err
		} else {
			logger.V(1).Info("updated SinglePlacementSlice")
		}
	}

	return nil
}
