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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	kcpcache "github.com/kcp-dev/apimachinery/v2/pkg/cache"
	schedulingv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/scheduling/v1alpha1"
	workloadv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/workload/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v3"

	edgev1alpha1 "github.com/kcp-dev/edge-mc/pkg/apis/edge/v1alpha1"
)

func (c *controller) reconcileOnSyncTarget(ctx context.Context, stKey string) error {
	logger := klog.FromContext(ctx)
	stws, _, stName, err := kcpcache.SplitMetaClusterNamespaceKey(stKey)
	if err != nil {
		logger.Error(err, "invalid SyncTarget key")
		return err
	}
	logger = logger.WithValues("workspace", stws, "syncTarget", stName)
	logger.V(2).Info("reconciling")

	/*
		On synctarget change:
		1) find all its loc(s), and update store

		2a) find all its ep(s), and update store
		2b) for each of its obsolete ep, remove all sp(s) that has st
		2c) for each of its ongoing ep, update sp(s)
		2d) for each of its new ep, add sp(s)

		Need data structure:
		- map from a synctarget to its locations, say 'locsBySelectedSt'
		- map from a synctarget to its eps, say 'epsByUsedSt'
	*/

	st, _ := c.synctargetLister.Cluster(stws).Get(stName) // TODO: handle err

	// 1)
	locsColocating, _ := c.locationLister.Cluster(stws).List(labels.NewSelector()) // TOOD: handle err
	locsSelectingSt, _ := filterLocsBySt(locsColocating, st)                       // TODO: handle err

	locKeys := extractLocKeys(locsSelectingSt)
	// locsSelectedSt := store.locsBySelectedSt[stKey]
	store.locsBySelectedSt[stKey] = locKeys

	// 2a)
	epsUsingSt := []string{}
	for _, loc := range locsSelectingSt {
		locKey := kcpcache.ToClusterAwareKey(string(logicalcluster.From(loc)), loc.GetNamespace(), loc.GetName())
		eps := store.epsBySelectedLoc[locKey]
		epsUsingSt = append(epsUsingSt, eps...)
	}
	epsUsingSt, usingSet := dedupAndHash(epsUsingSt)

	epsUsedSt := store.epsByUsedSt[stKey]
	store.epsByUsedSt[stKey] = epsUsingSt

	epsUsedSt, usedSet := dedupAndHash(epsUsedSt)

	for _, ep := range epsUsedSt {
		if _, ok := usingSet[ep]; !ok {
			// 2b)
			// for an (obsolite) ep in epsUsedSt but not in epsUsingSt
			// remove all relevant sp(s) from that ep, so that that ep doesn't use st
			// an obsolite ep doesn't use st anymore because its locs don't select the st anymore
			logger.V(1).Info("stop to use SyncTarget", "edgePlacement", ep)
			ws, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, _ := c.singlePlacementSliceLister.Cluster(ws).Get(name) // TODO: handle err
			nextSPS := cleanSPSBySt(currentSPS, stws.String(), stName)
			c.edgeClusterClient.EdgeV1alpha1().SinglePlacementSlices().Cluster(ws.Path()).Update(ctx, nextSPS, metav1.UpdateOptions{}) // TODO: handle err
		} else {
			// 2c)
			// for an (ongoing) ep in both epsUsedSt and epsUsingSt
			// the ep continues to use this st,
			// because at least one locs selected the st, AND at least one locs are selecting the st
			// but the two sets of locs may or may not overlap
			// thus, we need to clean then extend the corresponding sps
			logger.V(1).Info("continue to use SyncTarget", "edgePlacement", ep)
			ws, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, _ := c.singlePlacementSliceLister.Cluster(ws).Get(name) // TODO: handle err
			nextSPS := cleanSPSBySt(currentSPS, stws.String(), stName)

			// TODO(waltforme): Maybe I can improve this by making epsBySelectedLoc's value be a map instead of a slice
			// But the improvements should be marginal because I have to access the apiserver (to update sps) anyway
			epObj, _ := c.edgePlacementLister.Cluster(ws).Get(name)
			locsSelectedByEp, _ := filterLocsByEp(locsSelectingSt, epObj)      // TODO: handle err
			additionalSingles, _ := makeSinglePlacements(locsSelectedByEp, st) // TODO: handle err
			nextSPS = extendSPS(nextSPS, additionalSingles)

			c.edgeClusterClient.EdgeV1alpha1().SinglePlacementSlices().Cluster(ws.Path()).Update(ctx, nextSPS, metav1.UpdateOptions{}) // TODO: handle err
		}
	}

	// 2d)
	// for a (new) ep in epsUsingSt but not in epsUsedSt
	// insert composed sp(s) into the ep's sps
	for _, ep := range epsUsingSt {
		if _, ok := usedSet[ep]; !ok {
			logger.V(1).Info("begin to use SyncTarget", "edgePlacement", ep)
			ws, _, name, err := kcpcache.SplitMetaClusterNamespaceKey(ep)
			if err != nil {
				logger.Error(err, "invalid EdgePlacement key")
				return err
			}
			currentSPS, _ := c.singlePlacementSliceLister.Cluster(ws).Get(name) // TODO: handle err

			// TODO(waltforme): Maybe I can improve this by making epsBySelectedLoc's value be a map instead of a slice
			// But the improvements should be marginal because I have to access the apiserver (to update sps) anyway
			epObj, _ := c.edgePlacementLister.Cluster(ws).Get(name)
			locsSelectedByEp, _ := filterLocsByEp(locsSelectingSt, epObj)      // TODO: handle err
			additionalSingles, _ := makeSinglePlacements(locsSelectedByEp, st) // TODO: handle err
			nextSPS := extendSPS(currentSPS, additionalSingles)

			c.edgeClusterClient.EdgeV1alpha1().SinglePlacementSlices().Cluster(ws.Path()).Update(ctx, nextSPS, metav1.UpdateOptions{}) // TODO: handle err
		}
	}

	// dev-time tests
	store.show()

	return nil
}

// filterLocsBySt returns those Locations that select the SyncTarget
func filterLocsBySt(locs []*schedulingv1alpha1.Location, st *workloadv1alpha1.SyncTarget) ([]*schedulingv1alpha1.Location, error) {
	filtered := []*schedulingv1alpha1.Location{}
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
func filterLocsByEp(locs []*schedulingv1alpha1.Location, ep *edgev1alpha1.EdgePlacement) ([]*schedulingv1alpha1.Location, error) {
	filtered := []*schedulingv1alpha1.Location{}
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

func makeSinglePlacements(locsSelectingSt []*schedulingv1alpha1.Location, st *workloadv1alpha1.SyncTarget) ([]edgev1alpha1.SinglePlacement, error) {
	ws := logicalcluster.From(st).String()
	made := []edgev1alpha1.SinglePlacement{}
	for _, loc := range locsSelectingSt {
		sp := edgev1alpha1.SinglePlacement{
			Cluster:        ws,
			LocationName:   loc.Name,
			SyncTargetName: st.Name,
			SyncTargetUID:  st.UID,
		}
		made = append(made, sp)
	}
	return made, nil
}

func dedupAndHash(s []string) ([]string, map[string]struct{}) {
	t, hash := []string{}, map[string]struct{}{}
	for _, k := range s {
		if _, ok := hash[k]; !ok {
			t = append(t, k)
			hash[k] = struct{}{}
		}
	}
	return t, hash
}

func extractLocKeys(locs []*schedulingv1alpha1.Location) []string {
	keys := []string{}
	for _, l := range locs {
		key := kcpcache.ToClusterAwareKey(string(logicalcluster.From(l)), l.GetNamespace(), l.GetName())
		keys = append(keys, key)
	}
	return keys
}

// cleanSPSBySt removes all singleplacements that has the specified synctarget, from a singleplacementslice
func cleanSPSBySt(sps *edgev1alpha1.SinglePlacementSlice, stws, stName string) *edgev1alpha1.SinglePlacementSlice {
	nextDests := []edgev1alpha1.SinglePlacement{}
	for _, sp := range sps.Destinations {
		if sp.Cluster != stws || sp.SyncTargetName != stName {
			nextDests = append(nextDests, sp)
		}
	}
	sps.Destinations = nextDests
	return sps
}

func extendSPS(sps *edgev1alpha1.SinglePlacementSlice, singles []edgev1alpha1.SinglePlacement) *edgev1alpha1.SinglePlacementSlice {
	sps.Destinations = append(sps.Destinations, singles...)
	return sps
}
