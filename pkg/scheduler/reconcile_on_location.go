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
	schedulingv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/scheduling/v1alpha1"
	workloadv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/workload/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func (c *controller) reconcileOnLocation(ctx context.Context, locKey string) error {
	logger := klog.FromContext(ctx)
	lws, _, lName, err := kcpcache.SplitMetaClusterNamespaceKey(locKey)
	if err != nil {
		logger.Error(err, "invalid Location key")
		return err
	}
	logger = logger.WithValues("locationWorkspace", lws, "location", lName)
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
	loc, err := c.locationLister.Cluster(lws).Get(lName)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.V(1).Info("Location not found")
			locDeleted = true
		} else {
			logger.Error(err, "failed to get Location")
			return err
		}
	}

	// 1)
	epsSelectedLoc := store.epsBySelectedLoc[locKey]

	// 2a)
	stsFilteredByLoc := []*workloadv1alpha1.SyncTarget{}
	if !locDeleted {
		stsInLws, err := c.synctargetLister.Cluster(lws).List(labels.Everything())
		if err != nil {
			logger.Error(err, "failed to list SyncTargets")
			return err
		}
		stsFilteredByLoc, err = filterStsByLoc(stsInLws, loc)
		if err != nil {
			logger.Error(err, "failed to find SyncTargets for Location")
			return err
		}
	}
	stsSelecting := packStKeys(stsFilteredByLoc)

	// 2b)
	epsFilteredByLoc := []*edgev1alpha1.EdgePlacement{}
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
	singles := makeSinglePlacementsForLoc(loc, stsFilteredByLoc)

	for ep := range epsSelectedLoc {
		if _, ok := epsSelectingLoc[ep]; !ok {
			// 4a)
			// an (obsolete) ep doesn't select loc anymore
			// we need to remove all relevant sp(s) from the corresponding sps where 'relevant' means an sp has loc
			logger.V(1).Info("stop to select Location", "edgePlacement", ep)
			ws, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, err := c.singlePlacementSliceLister.Cluster(ws).Get(name)
			if err != nil {
				logger.Error(err, "failed to get SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", name)
				return err
			}
			nextSPS := cleanSPSByLoc(currentSPS, lws.String(), lName)
			_, err = c.edgeClusterClient.EdgeV1alpha1().SinglePlacementSlices().Cluster(ws.Path()).Update(ctx, nextSPS, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "failed to update SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", nextSPS.Name)
				return err
			} else {
				logger.V(1).Info("updated SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", nextSPS.Name)
			}
		} else {
			// 4b)
			// an (ongoing) ep continues to select loc
			// loc selected a set of synctargets, AND is selecting another set of synctargets
			// the two sets of synctargets may or may not overlap
			// thus, we need to clean then extend the corresponding sps
			logger.V(1).Info("continue to select Location", "edgePlacement", ep)
			ws, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, err := c.singlePlacementSliceLister.Cluster(ws).Get(name)
			if err != nil {
				logger.Error(err, "failed to get SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", name)
				return err
			}

			nextSPS := cleanSPSByLoc(currentSPS, lws.String(), lName)
			nextSPS = extendSPS(nextSPS, singles)

			_, err = c.edgeClusterClient.EdgeV1alpha1().SinglePlacementSlices().Cluster(ws.Path()).Update(ctx, nextSPS, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "failed to update SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", nextSPS.Name)
				return err
			} else {
				logger.V(1).Info("updated SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", nextSPS.Name)
			}
		}
	}

	for ep := range epsSelectingLoc {
		if _, ok := epsSelectedLoc[ep]; !ok {
			// 4c)
			// a (new) ep begins to select loc
			// we need to ensure the existence of sp(s) in the corresponding sps
			logger.V(1).Info("begin to select Location", "edgePlacement", ep)
			ws, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, err := c.singlePlacementSliceLister.Cluster(ws).Get(name)
			if err != nil {
				logger.Error(err, "failed to get SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", name)
				return err
			}

			nextSPS := cleanSPSByLoc(currentSPS, lws.String(), lName)
			nextSPS = extendSPS(nextSPS, singles)

			_, err = c.edgeClusterClient.EdgeV1alpha1().SinglePlacementSlices().Cluster(ws.Path()).Update(ctx, nextSPS, metav1.UpdateOptions{})
			if err != nil {
				logger.Error(err, "failed to update SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", nextSPS.Name)
				return err
			} else {
				logger.V(1).Info("updated SinglePlacementSlice", "workloadWorkspace", ws, "singlePlacementSlice", nextSPS.Name)
			}
		}
	}

	return nil
}

// filterStsByLoc returns those SyncTargets that selected by the Location
func filterStsByLoc(sts []*workloadv1alpha1.SyncTarget, loc *schedulingv1alpha1.Location) ([]*workloadv1alpha1.SyncTarget, error) {
	filtered := []*workloadv1alpha1.SyncTarget{}
	for _, st := range sts {
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
func filterEpsByLoc(eps []*edgev1alpha1.EdgePlacement, loc *schedulingv1alpha1.Location) ([]*edgev1alpha1.EdgePlacement, error) {
	filtered := []*edgev1alpha1.EdgePlacement{}
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
func packEpKeys(eps []*edgev1alpha1.EdgePlacement) map[string]empty {
	keys := map[string]empty{}
	for _, ep := range eps {
		key, _ := kcpcache.MetaClusterNamespaceKeyFunc(ep)
		keys[key] = empty{}
	}
	return keys
}

// packStKeys extracts keys from given SyncTargets and put the keys in a map
func packStKeys(sts []*workloadv1alpha1.SyncTarget) map[string]empty {
	keys := map[string]empty{}
	for _, st := range sts {
		key, _ := kcpcache.MetaClusterNamespaceKeyFunc(st)
		keys[key] = empty{}
	}
	return keys
}

// cleanSPSByLoc removes all singleplacements that has the specified location, from a singleplacementslice
func cleanSPSByLoc(sps *edgev1alpha1.SinglePlacementSlice, lws, lName string) *edgev1alpha1.SinglePlacementSlice {
	nextDests := []edgev1alpha1.SinglePlacement{}
	for _, sp := range sps.Destinations {
		if sp.Cluster != lws || sp.LocationName != lName {
			nextDests = append(nextDests, sp)
		}
	}
	sps.Destinations = nextDests
	return sps
}

func makeSinglePlacementsForLoc(locSelectingSts *schedulingv1alpha1.Location, sts []*workloadv1alpha1.SyncTarget) []edgev1alpha1.SinglePlacement {
	made := []edgev1alpha1.SinglePlacement{}
	if locSelectingSts == nil || len(sts) == 0 {
		return made
	}
	ws := logicalcluster.From(locSelectingSts).String()
	for _, st := range sts {
		sp := edgev1alpha1.SinglePlacement{
			Cluster:        ws,
			LocationName:   locSelectingSts.Name,
			SyncTargetName: st.Name,
			SyncTargetUID:  st.UID,
		}
		made = append(made, sp)
	}
	return made
}
