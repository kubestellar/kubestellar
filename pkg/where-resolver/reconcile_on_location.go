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
	"encoding/json"
	"errors"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

	edgev2alpha1 "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
	edgeclientset "github.com/kubestellar/kubestellar/pkg/client/clientset/versioned"
	"github.com/kubestellar/kubestellar/pkg/kbuser"
)

func (c *controller) reconcileOnLocation(ctx context.Context, locKey string) error {
	logger := klog.FromContext(ctx)
	_, lName, err := cache.SplitMetaNamespaceKey(locKey)
	if err != nil {
		logger.Error(err, "invalid Location key")
		return err
	}
	logger = logger.WithValues("location", lName)
	logger.V(2).Info("reconciling")

	/*
		On location 'loc' change:
		1) from store, find ep(s) that selected loc

		2a) from cache, find st(s) that being selected by loc
		2b) from cache, find ep(s) that selecting loc

		3) update store
		3a) update store, with ep(s) that selecting loc
		3b) update store, with st(s) that being selected by loc

		4) update apiserver
		4a) for each of its obsolete ep, remove sp(s)
		4b) for each of its ongoing ep, update sp(s)
		4c) for each of its new ep, ensure the existence of sp(s)

		Need data structure:
		- map from a location to its eps

		TODO(waltforme): Maybe I can merge 4b) and 4c)
	*/

	store.l.Lock()
	defer store.l.Unlock() // TODO(waltforme): Is it safe to shorten the critical section?

	locDeleted := false
	loc, err := c.locationLister.Get(lName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			logger.V(1).Info("Location not found")
			locDeleted = true
		} else {
			logger.Error(err, "failed to get Location")
			return err
		}
	}
	_, locOriginalName, kbSpaceID, err := kbuser.AnalyzeObjectID(loc)
	if err != nil {
		return err
	}
	locSpaceID := c.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if locSpaceID == "" {
		return errors.New("failed to obtain space ID from kube-bind reference")
	}

	// 1)
	epsSelectedLoc := store.epsBySelectedLoc[locKey]

	// 2a)
	stsFilteredByLoc := []*edgev2alpha1.SyncTarget{}
	if !locDeleted {
		allSts, err := c.synctargetLister.List(labels.Everything())
		if err != nil {
			logger.Error(err, "failed to list SyncTargets")
			return err
		}
		stsFilteredByLoc, err = filterStsByLoc(allSts, loc)
		if err != nil {
			logger.Error(err, "failed to find SyncTargets for Location")
			return err
		}
	}
	stsSelecting := packStKeys(stsFilteredByLoc)

	// 2b)
	epsFilteredByLoc := []*edgev2alpha1.EdgePlacement{}
	if !locDeleted {
		epsAll, err := c.edgePlacementLister.List(labels.Everything())
		if err != nil {
			logger.Error(err, "failed to list EdgePlacements in all workspaces")
			return err
		}
		epsFilteredByLoc, err = filterEpsByLoc(epsAll, loc)
		if err != nil {
			logger.Error(err, "failed to find EdgePlacements for Location")
		}
	}
	epsSelectingLoc := packEpKeys(epsFilteredByLoc)

	// 3)
	if !locDeleted {
		logger.V(3).Info("updating store")
		// 3a)
		store.epsBySelectedLoc[locKey] = epsSelectingLoc
		// 3b)
		for st, locs := range store.locsBySelectedSt {
			if _, ok := stsSelecting[st]; !ok {
				delete(locs, locKey)
			}
		}
		for st := range stsSelecting {
			if store.locsBySelectedSt[st] == nil {
				store.locsBySelectedSt[st] = map[string]empty{locKey: {}}
			} else {
				store.locsBySelectedSt[st][locKey] = empty{}
			}
		}
	} else {
		logger.V(3).Info("dropping Location from store")
		store.dropLoc(locKey)
	}

	// 4)
	singles := c.makeSinglePlacementsForLoc(loc, stsFilteredByLoc)

	for ep := range epsSelectedLoc {
		if _, ok := epsSelectingLoc[ep]; !ok {
			// 4a)
			// an (obsolete) ep doesn't select loc anymore
			// we need to remove all relevant sp(s) from the corresponding sps where 'relevant' means an sp has loc
			logger.V(1).Info("stop selecting Location", "edgePlacement", ep)
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
			// get consumer's space ID by currentSPS
			originalName, spaceID, err := c.getConsumerSpaceForSPS(currentSPS)
			if err != nil {
				logger.Error(err, "failed to get consumer space ID from a provider's copy", "singlePlacementSlice", name)
				return err
			}

			nextSPS := cleanSPSByLoc(currentSPS, locSpaceID, locOriginalName)
			err = c.patchSpsDestinations(nextSPS.Destinations, spaceID, originalName)
			if err != nil {
				logger.Error(err, "failed to update SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
				return err
			} else {
				logger.V(1).Info("updated SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
			}
		} else {
			// 4b)
			// an (ongoing) ep continues to select loc
			// loc selected a set of synctargets, AND is selecting another set of synctargets
			// the two sets of synctargets may or may not overlap
			// thus, we need to clean then extend the corresponding sps
			logger.V(1).Info("continue to select Location", "edgePlacement", ep)
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
			// get consumer's space ID by currentSPS
			originalName, spaceID, err := c.getConsumerSpaceForSPS(currentSPS)
			if err != nil {
				logger.Error(err, "failed to get consumer space ID from a provider's copy", "singlePlacementSlice", name)
				return err
			}
			nextSPS := cleanSPSByLoc(currentSPS, locSpaceID, locOriginalName)
			nextSPS = extendSPS(nextSPS, singles)
			err = c.patchSpsDestinations(nextSPS.Destinations, spaceID, originalName)
			if err != nil {
				logger.Error(err, "failed to update SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
				return err
			} else {
				logger.V(1).Info("updated SinglePlacementSlice", "singlePlacementSlice", nextSPS.Name)
			}
		}
	}

	for ep := range epsSelectingLoc {
		if _, ok := epsSelectedLoc[ep]; !ok {
			// 4c)
			// a (new) ep begins to select loc
			// we need to ensure the existence of sp(s) in the corresponding sps
			logger.V(1).Info("begin to select Location", "edgePlacement", ep)
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
			// get consumer's space ID by currentSPS
			originalName, spaceID, err := c.getConsumerSpaceForSPS(currentSPS)
			if err != nil {
				logger.Error(err, "failed to get consumer space ID from a provider's copy", "singlePlacementSlice", name)
				return err
			}

			nextSPS := cleanSPSByLoc(currentSPS, locSpaceID, locOriginalName)
			nextSPS = extendSPS(nextSPS, singles)
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

// filterStsByLoc returns those SyncTargets that selected by the Location
func filterStsByLoc(sts []*edgev2alpha1.SyncTarget, loc *edgev2alpha1.Location) ([]*edgev2alpha1.SyncTarget, error) {
	filtered := []*edgev2alpha1.SyncTarget{}

	_, _, locKBSpaceID, err := kbuser.AnalyzeObjectID(loc)
	if err != nil {
		return filtered, err
	}

	for _, st := range sts {
		_, _, stKBSpaceID, err := kbuser.AnalyzeObjectID(st)
		if err != nil {
			return filtered, err
		}

		if stKBSpaceID != locKBSpaceID {
			continue
		}
		s := loc.Spec.InstanceSelector
		selector, err := metav1.LabelSelectorAsSelector(s)
		if err != nil {
			return filtered, err
		}
		if selector.Matches(labels.Set(st.Labels)) {
			filtered = append(filtered, st)
		}
	}
	return filtered, nil
}

// filterEpsByLoc returns those EdgePlacements that select the Location
func filterEpsByLoc(eps []*edgev2alpha1.EdgePlacement, loc *edgev2alpha1.Location) ([]*edgev2alpha1.EdgePlacement, error) {
	filtered := []*edgev2alpha1.EdgePlacement{}
	for _, ep := range eps {
		for _, s := range ep.Spec.LocationSelectors {
			selector, err := metav1.LabelSelectorAsSelector(&s)
			if err != nil {
				return filtered, err
			}
			if selector.Matches(labels.Set(loc.Labels)) {
				filtered = append(filtered, ep)
				break
			}
		}
	}
	return filtered, nil
}

// packEpKeys extracts keys from given EdgePlacements and put the keys in a map
func packEpKeys(eps []*edgev2alpha1.EdgePlacement) map[string]empty {
	keys := map[string]empty{}
	for _, ep := range eps {
		key, _ := cache.MetaNamespaceKeyFunc(ep)
		keys[key] = empty{}
	}
	return keys
}

// packStKeys extracts keys from given SyncTargets and put the keys in a map
func packStKeys(sts []*edgev2alpha1.SyncTarget) map[string]empty {
	keys := map[string]empty{}
	for _, st := range sts {
		key, _ := cache.MetaNamespaceKeyFunc(st)
		keys[key] = empty{}
	}
	return keys
}

// cleanSPSByLoc removes all singleplacements that has the specified location, from a singleplacementslice
func cleanSPSByLoc(sps *edgev2alpha1.SinglePlacementSlice, locSpaceID, lName string) *edgev2alpha1.SinglePlacementSlice {
	nextDests := []edgev2alpha1.SinglePlacement{}
	for _, sp := range sps.Destinations {
		if sp.Cluster != locSpaceID || sp.LocationName != lName {
			nextDests = append(nextDests, sp)
		}
	}
	sps.Destinations = nextDests
	return sps
}

func (c *controller) makeSinglePlacementsForLoc(locSelectingSts *edgev2alpha1.Location, sts []*edgev2alpha1.SyncTarget) []edgev2alpha1.SinglePlacement {
	made := []edgev2alpha1.SinglePlacement{}
	if locSelectingSts == nil || len(sts) == 0 {
		return made
	}
	_, locOriginalName, kbSpaceID, err := kbuser.AnalyzeObjectID(locSelectingSts)
	if err != nil {
		return made
	}
	locSpaceID := c.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if locSpaceID == "" {
		return made
	}
	for _, st := range sts {
		_, stOriginalName, _, err := kbuser.AnalyzeObjectID(st)
		if err != nil {
			continue
		}
		sp := edgev2alpha1.SinglePlacement{
			Cluster:        locSpaceID,
			LocationName:   locOriginalName,
			SyncTargetName: stOriginalName,
			SyncTargetUID:  st.UID,
		}
		made = append(made, sp)
	}
	return made
}

func (c *controller) getConsumerSpaceForSPS(sps *edgev2alpha1.SinglePlacementSlice) (string, string, error) {
	_, name, kbSpaceID, err := kbuser.AnalyzeObjectID(sps)
	if err != nil {
		return "", "", err
	}
	spaceID := c.kbSpaceRelation.SpaceIDFromKubeBind(kbSpaceID)
	if spaceID == "" {
		return "", "", errors.New("failed to get consumer space ID from a provider's copy")
	}
	return name, spaceID, nil
}

func (c *controller) patchSpsDestinations(destinations []edgev2alpha1.SinglePlacement, spaceID string, spsName string) error {
	destBytes, err := json.Marshal(destinations)
	if err != nil {
		return err
	}
	patch := []byte(fmt.Sprintf(`{"destinations": %s}`, destBytes))

	spaceConfig, err := c.spaceClient.ConfigForSpace(spaceID, c.spaceProviderNs)
	if err != nil {
		return err
	}
	edgeClientset, err := edgeclientset.NewForConfig(spaceConfig)
	if err != nil {
		return err
	}
	_, err = edgeClientset.EdgeV2alpha1().SinglePlacementSlices().Patch(c.context, spsName, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}
