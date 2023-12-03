/*
Copyright 2023 The KubeStellar Authors.

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

package where_resolver

import (
	"context"
	"errors"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
)

func (c *controller) reconcileOnEdgePlacement(ctx context.Context, epKey string) error {
	logger := klog.FromContext(ctx)
	_, epName, err := cache.SplitMetaNamespaceKey(epKey)
	if err != nil {
		logger.Error(err, "invalid EdgePlacement key")
		return err
	}
	logger = logger.WithValues("edgePlacement", epName)
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

	ep, err := c.edgePlacementLister.Get(epName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
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

	singles := []edgev2alpha1.SinglePlacement{}
	for _, loc := range locsFilteredByEp {
		// 2)
		stsInLws, err := c.synctargetLister.List(labels.Everything())
		if err != nil {
			logger.Error(err, "Failed to list SyncTargets in local cache")
			return err
		}
		stsSelecting, err := filterStsByLoc(stsInLws, loc)
		if err != nil {
			logger.Error(err, "failed to find SyncTargets for Location", "location", loc.Name)
			return err
		}
		singles = append(singles, c.makeSinglePlacementsForLoc(loc, stsSelecting)...)
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
	_, originalName, kbSpaceID, err := kbuser.AnalyzeObjectID(ep)
	if err != nil {
		logger.Error(err, "Object does not appear to be a provider's copy of a consumer's object", "edgePlacement", ep.Name)
		return err
	}
	spaceID := c.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if spaceID == "" {
		relErr := errors.New("failed to obtain space ID from kube-bind reference")
		logger.Error(relErr, "Failed to get consumer space ID from a provider's copy", "edgePlacement", originalName)
		return relErr
	}

	spaceConfig, err := c.spaceClient.ConfigForSpace(spaceID, c.spaceProviderNs)
	if err != nil {
		return err
	}
	edgeClientset, err := edgeclientset.NewForConfig(spaceConfig)
	if err != nil {
		return err
	}
	originalEP, err := edgeClientset.EdgeV2alpha1().EdgePlacements().Get(ctx, originalName, metav1.GetOptions{})
	if err != nil {
		logger.Error(err, "failed to get consumer's object", "edgePlacement", originalName)
		return err
	}
	_, err = c.singlePlacementSliceLister.Get(epName)
	if err != nil {
		if k8serrors.IsNotFound(err) { // create
			logger.V(1).Info("creating SinglePlacementSlice")
			sps := &edgev2alpha1.SinglePlacementSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name: originalName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: edgev2alpha1.SchemeGroupVersion.String(),
							Kind:       "EdgePlacement",
							Name:       originalName,
							UID:        originalEP.UID,
						},
					},
				},
				Destinations: singles,
			}
			_, err = edgeClientset.EdgeV2alpha1().SinglePlacementSlices().Create(ctx, sps, metav1.CreateOptions{})
			if err != nil {
				if !k8serrors.IsAlreadyExists(err) {
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
		err := c.patchSpsDestinations(singles, spaceID, originalName)
		if err != nil {
			logger.Error(err, "failed updating SinglePlacementSlice")
			return err
		}
	}

	return nil
}
