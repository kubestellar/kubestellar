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

package placement

import (
	"context"
	"fmt"
	"strings"
	"time"

	workv1 "open-cluster-management.io/api/work/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kubestellar/kubestellar/api/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/ocm"
	"github.com/kubestellar/kubestellar/pkg/util"
)

const (
	waitBeforeTrackingPlacements = 5 * time.Second
	PlacementKind                = "Placement"
	KSFinalizer                  = "placement.kubestellar.io/kscontroller"
)

// Handle placement as follows:
//  1. requeue all objects to account for changes in placement
//  2. handle finalizers and deletion of objects associated with the placement
//  3. for updates on label selectors, re-evaluate if existing objects should be removed
//     from clusters.
func (c *Controller) handlePlacement(obj runtime.Object) error {
	placement := obj.DeepCopyObject()

	// handle requeing for changes in placement, excluding deletion
	if !isBeingDeleted(obj) {
		if err := c.requeueForPlacementChanges(); err != nil {
			return err
		}
	}

	if err := c.handlePlacementFinalizer(placement); err != nil {
		return err
	}

	if err := c.cleanUpObjectsNoLongerMatching(placement); err != nil {
		return err
	}

	return nil
}

func (c *Controller) requeueForPlacementChanges() error {
	// allow some time before checking to settle
	now := time.Now()
	if now.Sub(c.initializedTs) < waitBeforeTrackingPlacements {
		return nil
	}
	if err := c.requeueAll(); err != nil {
		return err
	}
	return nil
}

func (c *Controller) listPlacements() ([]runtime.Object, error) {
	pLister := c.listers["edge.kubestellar.io/v1alpha1/Placement"]
	if pLister == nil {
		return nil, fmt.Errorf("could not get lister for placememt")
	}
	lister := *pLister
	list, err := lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	return list, nil
}

func runtimeObjectToPlacement(obj runtime.Object) (*v1alpha1.Placement, error) {
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("failed to convert runtime.Object to unstructured.Unstructured")
	}
	var placement *v1alpha1.Placement
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredObj.UnstructuredContent(), &placement); err != nil {
		return nil, err
	}
	return placement, nil
}

// read objects from all listers and enqueue the keys
// this is useful for when a new placement is added
// or a placement is updated
func (c *Controller) requeueAll() error {
	for key, ptr := range c.listers {
		// do not requeue placement
		if key == util.GetPlacementListerKey() {
			fmt.Printf("Matched key %s\n", key)
			continue
		}
		lister := *ptr
		objs, err := lister.List(labels.Everything())
		if err != nil {
			return err
		}
		for _, obj := range objs {
			c.enqueueObject(obj, true)
		}
	}
	return nil
}

// finalizer logic
func (c *Controller) handlePlacementFinalizer(obj runtime.Object) error {
	mObj := obj.(metav1.Object)
	cObj, err := convertToClientObject(obj)
	if err != nil {
		return err
	}
	if mObj.GetDeletionTimestamp() != nil {
		if controllerutil.ContainsFinalizer(cObj, KSFinalizer) {
			if err := c.deleteExternalResources(obj); err != nil {
				return err
			}
			controllerutil.RemoveFinalizer(cObj, KSFinalizer)
			if err = updatePlacement(*c.dynamicClient, obj); err != nil {
				return err
			}
		}
		return nil
	}

	if !controllerutil.ContainsFinalizer(cObj, KSFinalizer) {
		controllerutil.AddFinalizer(cObj, KSFinalizer)
		if err = updatePlacement(*c.dynamicClient, obj); err != nil {
			return err
		}
	}
	return nil
}

func convertToClientObject(obj runtime.Object) (client.Object, error) {
	clientObj, ok := obj.(client.Object)
	if !ok {
		return nil, fmt.Errorf("object is not a client.Object: %v", obj)
	}
	return clientObj, nil
}

func updatePlacement(client dynamic.DynamicClient, obj runtime.Object) error {
	gvr := schema.GroupVersionResource{
		Group:    v1alpha1.GroupVersion.Group,
		Version:  obj.GetObjectKind().GroupVersionKind().Version,
		Resource: "placements",
	}
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unexpected type for obj, expected *unstructured.Unstructured")
	}
	client.Resource(gvr).Namespace("").Update(context.Background(), unstructuredObj, metav1.UpdateOptions{})
	return nil
}

func (c *Controller) deleteExternalResources(obj runtime.Object) error {
	list, err := listManifestsForPlacement(c.ocmClient, c.wdsName, obj)
	if err != nil {
		return err
	}
	mObj := obj.(metav1.Object)
	labelKey := util.GenerateManagedByPlacementLabelKey(c.wdsName, mObj.GetName())
	for _, manifest := range list.Items {
		c.logger.Info("Trying to delete manifest", "manifest name", manifest.Name, "namespace", manifest.Namespace, "for placement", mObj.GetName())
		if err := deleteManifestOrLabel(labelKey, manifest, c.ocmClient); err != nil {
			return err
		}
	}
	return nil
}

func listManifestsForPlacement(ocmClient client.Client, wdsName string, obj runtime.Object) (*workv1.ManifestWorkList, error) {
	mObj := obj.(metav1.Object)
	list := &workv1.ManifestWorkList{}
	labelKey := util.GenerateManagedByPlacementLabelKey(wdsName, mObj.GetName())

	// TODO - the ocm client used this way is not using cache. Replace with informer/lister based
	// on dynamic client to make sure to use the cache
	if err := ocmClient.List(context.TODO(), list, client.InNamespace(""),
		client.MatchingLabels(map[string]string{labelKey: util.PlacementLabelValueEnabled})); err != nil {
		return nil, err
	}
	return list, nil
}

func deleteManifestOrLabel(managedByLabelKey string, manifest workv1.ManifestWork, ocmClient client.Client) error {
	labels := manifest.GetLabels()

	if isAlsoManagedByOtherPlacements(labels, managedByLabelKey) {
		delete(labels, managedByLabelKey)
		if err := ocmClient.Update(context.TODO(), &manifest, &client.UpdateOptions{}); err != nil {
			return err
		}
		return nil
	}

	// if no other labels can safely delete
	if err := ocmClient.Delete(context.TODO(), &manifest, &client.DeleteOptions{}); err != nil {
		// can ignore as it could be already deleted by another thread
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func isAlsoManagedByOtherPlacements(labels map[string]string, managedByLabelKey string) bool {
	for key := range labels {
		if key == managedByLabelKey {
			continue
		}
		if strings.HasPrefix(key, util.PlacementLabelKeyBase) {
			return true
		}
	}
	return false
}

// Handle removal of objects no longer matching cluster/selector
func (c *Controller) cleanUpObjectsNoLongerMatching(placement runtime.Object) error {
	// allow some time before checking to settle
	now := time.Now()
	if now.Sub(c.initializedTs) < waitBeforeTrackingPlacements {
		return nil
	}

	list, err := listManifestsForPlacement(c.ocmClient, c.wdsName, placement)
	if err != nil {
		return err
	}

	for _, manifest := range list.Items {
		obj, err := extractObjectFromManifest(manifest)
		if err != nil {
			return err
		}
		matches, err := c.checkObjectMatchesWhatAndWhere(placement, *obj, manifest)
		if err != nil {
			return err
		}
		if !matches {
			if err := c.ocmClient.Delete(context.TODO(), &manifest, &client.DeleteOptions{}); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractObjectFromManifest(manifest workv1.ManifestWork) (*runtime.Object, error) {
	nObjs := len(manifest.Spec.Workload.Manifests)
	if nObjs != 1 {
		return nil, fmt.Errorf("manifest should contain one object. Found %d", nObjs)
	}

	raw := manifest.Spec.Workload.Manifests[0].RawExtension
	obj := &unstructured.Unstructured{}
	err := obj.UnmarshalJSON(raw.Raw)
	if err != nil {
		fmt.Printf("Error while decoding RawExtension: %v", err)
		return nil, err
	}
	rObj := runtime.Object(obj)

	return &rObj, nil
}

func (c *Controller) checkObjectMatchesWhatAndWhere(placementObj, obj runtime.Object, manifest workv1.ManifestWork) (bool, error) {
	// default is doing nothing, that is, return match ==true
	match := true
	placement, err := runtimeObjectToPlacement(placementObj)
	if err != nil {
		return match, err
	}

	// check the What matches
	mObj := obj.(metav1.Object)
	matchAllSelectors := false
	for _, downsync := range placement.Spec.Downsync {
		for _, s := range downsync.LabelSelectors {
			selector, err := metav1.LabelSelectorAsSelector(&s)
			if err != nil {
				return match, err
			}
			if selector.Matches(labels.Set(mObj.GetLabels())) {
				matchAllSelectors = true
			} else {
				matchAllSelectors = false
				break
			}
		}
	}
	if !matchAllSelectors {
		c.logger.Info("The 'What' no longer matches. Object marked for removal.", "object", util.GenerateObjectInfoString(obj), "for placement", placement.GetName())
		return false, nil
	}

	// check the Where matches
	clusterName := getClusterNameFromManifest(manifest)
	matchedClusters, err := ocm.ListClustersBySelectors(c.ocmClient, placement.Spec.ClusterSelectors)
	if err != nil {
		return match, err
	}
	if !util.StringInSlice(clusterName, matchedClusters) {
		c.logger.Info("The 'Where' no longer matches. Object marked for removal.", "object", util.GenerateObjectInfoString(obj), "for placement", placement.GetName(), "cluster", clusterName)
		return false, nil
	}

	return match, nil
}

func getClusterNameFromManifest(manifest workv1.ManifestWork) string {
	return manifest.Namespace
}
