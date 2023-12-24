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

package pluggable_transport_where_resolver

import (
	"context"
	"errors"

	k8sapierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
)

func (c *controller) reconcileOnSyncTarget(ctx context.Context, stKey string) error {
	logger := klog.FromContext(ctx)
	_, stName, err := cache.SplitMetaNamespaceKey(stKey)
	if err != nil {
		logger.Error(err, "invalid SyncTarget key")
		return err
	}
	logger = logger.WithValues("syncTarget", stName)
	logger.V(2).Info("reconciling")

	/*
		On synctarget 'st' change:
		1) from store, find ep(s) that used st

		2a) from cache, find loc(s) that selecting st
		2b) from store and cache, find ep(s) that using st

		3) update store, with loc(s) that selecting st

		4) update apiserver
		4a) for each of its obsolete ep, remove sp(s)
		4b) for each of its ongoing ep, update sp(s)
		4c) for each of its new ep, ensure the existence of sp(s)

		Need data structure:
		- map from a synctarget to its locations

		TODO(waltforme): Maybe I can merge 4b) and 4c)
	*/

	store.l.Lock()
	defer store.l.Unlock() // TODO(waltforme): Is it safe to shorten the critical section?

	stDeleted := false
	st, err := c.synctargetLister.Get(stName)
	if err != nil {
		if k8sapierrors.IsNotFound(err) {
			logger.V(1).Info("SyncTarget not found")
			stDeleted = true
		} else {
			logger.Error(err, "failed to get SyncTarget")
			return err
		}
	}
	stOriginalName, kbSpaceID, err := kbuser.AnalyzeClusterScopedObject(st)
	if err != nil {
		return err
	}
	stSpaceID := c.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if stSpaceID == "" {
		return errors.New("failed to obtain space ID from kube-bind reference")
	}

	// 1)
	epsUsedSt := store.findEpsUsedSt(stKey)

	// 2a)
	locsFilteredBySt := []*edgev2alpha1.Location{}
	if !stDeleted {
		locsInStws, err := c.locationLister.List(labels.Everything())
		if err != nil {
			logger.Error(err, "failed to list Locations")
			return err
		}
		locsFilteredBySt, err = filterLocsBySt(locsInStws, st)
		if err != nil {
			logger.Error(err, "failed to find Locations for SyncTarget")
			return err
		}
	}
	locsSelectingSt := packLocKeys(locsFilteredBySt)

	// 2b)
	epsUsingSt := map[string]empty{}
	if !stDeleted {
		for _, loc := range locsFilteredBySt {
			locKey, _ := cache.MetaNamespaceKeyFunc(loc)
			eps := store.epsBySelectedLoc[locKey]
			epsUsingSt = unionTwo(epsUsingSt, eps)
		}
	}

	// 3)
	if !stDeleted {
		logger.V(3).Info("updating store")
		store.locsBySelectedSt[stKey] = locsSelectingSt
	} else {
		logger.V(3).Info("dropping SyncTarget from store")
		store.dropSt(stKey)
	}

	// 4)
	for ep := range epsUsedSt {
		if _, ok := epsUsingSt[ep]; !ok {
			// 4a)
			// for an (obsolite) ep in epsUsedSt but not in epsUsingSt
			// remove all relevant sp(s) from that ep, so that that ep doesn't use st
			// an obsolite ep doesn't use st anymore because its locs don't select the st anymore
			logger.V(1).Info("stop using SyncTarget", "edgePlacement", ep)
			_, name, err := cache.SplitMetaNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, err := c.singlePlacementSliceLister.Get(name)
			if err != nil {
				logger.Error(err, "failed to get SinglePlacementSlice", "singlePlacementSlice", name)
				return err
			}
			nextSPS := cleanSPSBySt(currentSPS, stSpaceID, stOriginalName)
			originalName, spaceID, err := c.getConsumerSpaceForSPS(currentSPS)
			if err != nil {
				logger.Error(err, "failed to get consumer space ID from a provider's copy", "singlePlacementSlice", name)
				return err
			}
			err = c.patchSpsDestinations(nextSPS.Destinations, spaceID, originalName)
			if err != nil {
				logger.Error(err, "failed to update SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
				return err
			} else {
				logger.V(1).Info("updated SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
			}
		} else {
			// 4b)
			// for an (ongoing) ep in both epsUsedSt and epsUsingSt
			// the ep continues to use this st,
			// because at least one locs selected the st, AND at least one locs are selecting the st
			// but the two sets of locs may or may not overlap
			// thus, we need to clean then extend the corresponding sps
			logger.V(1).Info("continue to use SyncTarget", "edgePlacement", ep)
			_, name, err := cache.SplitMetaNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, err := c.singlePlacementSliceLister.Get(name)
			if err != nil {
				logger.Error(err, "failed to get SinglePlacementSlice", "singlePlacementSlice", name)
				return err
			}
			nextSPS := cleanSPSBySt(currentSPS, stSpaceID, stOriginalName)

			epObj, err := c.edgePlacementLister.Get(name)
			if err != nil {
				logger.Error(err, "failed to get EdgePlacement", "edgePlacement", name)
				return err
			}
			locsFilteredByStAndEp, err := filterLocsByEp(locsFilteredBySt, epObj)
			if err != nil {
				logger.Error(err, "failed to find Locations selected by EdgePlacement", "edgePlacement", epObj.Name)
				return err
			}
			additionalSingles := c.makeSinglePlacementsForSt(locsFilteredByStAndEp, st)
			nextSPS = extendSPS(nextSPS, additionalSingles)

			originalName, spaceID, err := c.getConsumerSpaceForSPS(currentSPS)
			if err != nil {
				logger.Error(err, "failed to get consumer space ID from a provider's copy", "singlePlacementSlice", name)
				return err
			}
			err = c.patchSpsDestinations(nextSPS.Destinations, spaceID, originalName)
			if err != nil {
				logger.Error(err, "failed to update SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
				return err
			} else {
				logger.V(1).Info("updated SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
			}
		}
	}

	for ep := range epsUsingSt {
		if _, ok := epsUsedSt[ep]; !ok {
			// 4c)
			// for a (new) ep in epsUsingSt but not in epsUsedSt
			// ensure the existence of sp(s) in the ep's sps
			logger.V(1).Info("begin to use SyncTarget", "edgePlacement", ep)
			_, name, err := cache.SplitMetaNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, err := c.singlePlacementSliceLister.Get(name)
			if err != nil {
				logger.Error(err, "failed to get SinglePlacementSlice", "singlePlacementSlice", name)
				return err
			}
			nextSPS := cleanSPSBySt(currentSPS, stSpaceID, stOriginalName)

			epObj, err := c.edgePlacementLister.Get(name)
			if err != nil {
				logger.Error(err, "failed to get EdgePlacement", "edgePlacement", name)
				return err
			}
			locsFilteredByStAndEp, err := filterLocsByEp(locsFilteredBySt, epObj)
			if err != nil {
				logger.Error(err, "failed to find Locations selected by EdgePlacement", "edgePlacement", epObj.Name)
				return err
			}
			additionalSingles := c.makeSinglePlacementsForSt(locsFilteredByStAndEp, st)
			nextSPS = extendSPS(nextSPS, additionalSingles)

			originalName, spaceID, err := c.getConsumerSpaceForSPS(currentSPS)
			if err != nil {
				logger.Error(err, "failed to get consumer space ID from a provider's copy", "singlePlacementSlice", name)
				return err
			}
			err = c.patchSpsDestinations(nextSPS.Destinations, spaceID, originalName)
			if err != nil {
				logger.Error(err, "failed to update SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
				return err
			} else {
				logger.V(1).Info("updated SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
			}
		}
	}

	return nil
}

// filterLocsBySt returns those Locations that select the SyncTarget
func filterLocsBySt(locs []*edgev2alpha1.Location, st *edgev2alpha1.SyncTarget) ([]*edgev2alpha1.Location, error) {
	filtered := []*edgev2alpha1.Location{}
	for _, l := range locs {
		s := l.Spec.InstanceSelector
		selector, err := metav1.LabelSelectorAsSelector(s)
		if err != nil {
			return filtered, err
		}
		if selector.Matches(labels.Set(st.Labels)) {
			filtered = append(filtered, l)
		}
	}
	return filtered, nil
}

// filterLocsByEp returns those Locations that are selected by the EdgePlacement
func filterLocsByEp(locs []*edgev2alpha1.Location, ep *edgev2alpha1.EdgePlacement) ([]*edgev2alpha1.Location, error) {
	filtered := []*edgev2alpha1.Location{}
	for _, l := range locs {
		for _, s := range ep.Spec.LocationSelectors {
			selector, err := metav1.LabelSelectorAsSelector(&s)
			if err != nil {
				return filtered, err
			}
			if selector.Matches(labels.Set(l.Labels)) {
				filtered = append(filtered, l)
				break
			}
		}
	}
	return filtered, nil
}

func (c *controller) makeSinglePlacementsForSt(locsSelectingSt []*edgev2alpha1.Location, st *edgev2alpha1.SyncTarget) []edgev2alpha1.SinglePlacement {
	made := []edgev2alpha1.SinglePlacement{}
	if len(locsSelectingSt) == 0 || st == nil {
		return made
	}
	stOriginalName, kbSpaceID, err := kbuser.AnalyzeClusterScopedObject(st)
	if err != nil {
		return made
	}
	stSpaceID := c.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if stSpaceID == "" {
		return made
	}
	for _, loc := range locsSelectingSt {
		locOriginalName, _, err := kbuser.AnalyzeClusterScopedObject(loc)
		if err != nil {
			continue
		}
		sp := edgev2alpha1.SinglePlacement{
			Cluster:        stSpaceID,
			LocationName:   locOriginalName,
			SyncTargetName: stOriginalName,
			SyncTargetUID:  st.UID,
		}
		made = append(made, sp)
	}
	return made
}

// packLocKeys extracts keys from given Locations and put the keys in a map
func packLocKeys(locs []*edgev2alpha1.Location) map[string]empty {
	keys := map[string]empty{}
	for _, l := range locs {
		key, _ := cache.MetaNamespaceKeyFunc(l)
		keys[key] = empty{}
	}
	return keys
}

// cleanSPSBySt removes all singleplacements that has the specified synctarget, from a singleplacementslice
func cleanSPSBySt(sps *edgev2alpha1.SinglePlacementSlice, stSpaceID, stName string) *edgev2alpha1.SinglePlacementSlice {
	nextDests := []edgev2alpha1.SinglePlacement{}
	for _, sp := range sps.Destinations {
		if sp.Cluster != stSpaceID || sp.SyncTargetName != stName {
			nextDests = append(nextDests, sp)
		}
	}
	sps.Destinations = nextDests
	return sps
}

func extendSPS(sps *edgev2alpha1.SinglePlacementSlice, singles []edgev2alpha1.SinglePlacement) *edgev2alpha1.SinglePlacementSlice {
	sps.Destinations = append(sps.Destinations, singles...)
	return sps
}
