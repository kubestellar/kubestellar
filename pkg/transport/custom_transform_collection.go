/*
Copyright 2024 The KubeStellar Authors.

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

package transport

import (
	"context"
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/kubestellar/kubestellar/api/control/v1alpha1"
	controlclient "github.com/kubestellar/kubestellar/pkg/generated/clientset/versioned/typed/control/v1alpha1"
	"github.com/kubestellar/kubestellar/pkg/jsonpath"
)

// customTransformCollection maintains the digested custom transformation instructions
// for each metav1.GroupResource.
type customTransformCollection struct {
	// client is here for updating status in a CustomTransform
	client controlclient.CustomTransformInterface

	// getTransformObjects is the part of the CustomTransform informer's cache.Indexer behavior
	// that is needed here, for using the index named in `customTransformDomainIndexName`.
	getTransformObjects func(indexName, indexedValue string) ([]any, error)

	// enqueue is used to add a reference to a Binding that needs to be re-processed because
	// of a change to a CustomTransform that the Binding is sensitive to.
	enqueue func(any)

	// mutex must be locked while accessing the following fields or their contents
	mutex sync.Mutex

	// grToTransformData has an entry for every GroupResource that some Binding cares about
	// (i.e., lists an object of that GroupResource), and no more entries.
	grToTransformData map[metav1.GroupResource]*groupResourceTransformData

	// ctNameToSpec holds the Spec last processed for each CustomTransform
	ctNameToSpec map[string]v1alpha1.CustomTransformSpec

	// bindingNameToGroupResources tracks the set of GroupResource that each Binding
	// references. This is so that when the set for a given Binding changes,
	// for the GroupResources that are no longer in the set, the Binding's Name can
	// be removed from groupResourceTransformData.bindingsThatCare.
	bindingNameToGroupResources map[string]sets.Set[metav1.GroupResource]
}

// groupResourceTransformData is the ingested custom transforms for a given GroupResource
type groupResourceTransformData struct {
	bindingsThatCare sets.Set[string /*Binding name*/] // not empty
	removes          []jsonpath.Query
}

func newCustomTransformCollection(client controlclient.CustomTransformInterface, getTransformObjects func(indexName, indexedValue string) ([]any, error), enqueue func(any)) *customTransformCollection {
	return &customTransformCollection{
		client:                      client,
		getTransformObjects:         getTransformObjects,
		enqueue:                     enqueue,
		grToTransformData:           make(map[metav1.GroupResource]*groupResourceTransformData),
		ctNameToSpec:                make(map[string]v1alpha1.CustomTransformSpec),
		bindingNameToGroupResources: make(map[string]sets.Set[metav1.GroupResource]),
	}
}

// Call this to process a new version, or deletion, of a CustomTransform.
func (ctc *customTransformCollection) noteCustomTransform(ctx context.Context, name string, ct *v1alpha1.CustomTransform) {
	ctc.mutex.Lock()
	defer ctc.mutex.Unlock()
	oldSpec, hadSpec := ctc.ctNameToSpec[name]
	if ct == nil && !hadSpec {
		return // was absent and is absent; no change
	}
	var oldGroupResource, newGroupResource metav1.GroupResource
	if hadSpec {
		oldGroupResource = ctSpecGroupResource(oldSpec)
	}
	if ct != nil {
		newGroupResource = ctSpecGroupResource(ct.Spec)
	}
	if ct != nil && hadSpec &&
		oldGroupResource == newGroupResource &&
		sets.New(oldSpec.Remove...).Equal(sets.New(ct.Spec.Remove...)) {
		return // unchanged
	}
	// GroupResource or Spec.Remove changed
	logger := klog.FromContext(ctx)
	if hadSpec {
		// Invalidate the cached analysis for the relevant GroupResource, if there is one
		if oldGRTD, hadGRTD := ctc.grToTransformData[oldGroupResource]; hadGRTD {
			delete(ctc.grToTransformData, oldGroupResource)
			for bindingName := range oldGRTD.bindingsThatCare {
				logger.V(5).Info("Enqueuing reference to Binding because of change to CustomTranform", "bindingName", bindingName, "customTransformName", name)
				ctc.enqueue(bindingName)
			}
		}
	}
	if ct != nil {
		ctc.ctNameToSpec[name] = ct.Spec
	} else {
		delete(ctc.ctNameToSpec, name)
	}
}

// getCustomTransformData returns the groupResourceTransformData to use
// for the given GroupResource on behalf of the named Binding.
func (ctc *customTransformCollection) getCustomTransformData(ctx context.Context, groupResource metav1.GroupResource, bindingName string) *groupResourceTransformData {
	logger := klog.FromContext(ctx)
	ctc.mutex.Lock()
	defer ctc.mutex.Unlock()
	grtd, ok := ctc.grToTransformData[groupResource]
	if !ok {
		grtd = &groupResourceTransformData{
			bindingsThatCare: sets.New(bindingName),
		}
		ctKey := customTransformDomainKey(groupResource.Group, groupResource.Resource)
		ctAnys, err := ctc.getTransformObjects(customTransformDomainIndexName, ctKey)
		if err != nil {
			// This only happens if the index is not defined;
			// that is, it never happens.
			// If it does, retry will not help.
			logger.Error(err, "Failed to get objects from CustomTransform domain index", "key", ctKey)
		}
		ctNames := []string{}
		for _, ctAny := range ctAnys {
			ct := ctAny.(*v1alpha1.CustomTransform)
			removes := ctc.parseRemovesAndUpdateStatus(ctx, ct)
			grtd.removes = append(grtd.removes, removes...)
			oldSpec, had := ctc.ctNameToSpec[ct.Name]
			if had {
				oldGroupResource := ctSpecGroupResource(oldSpec)
				if oldGroupResource != groupResource {
					oldGRTD := ctc.grToTransformData[oldGroupResource]
					if oldGRTD != nil {
						for _, oldBindingName := range oldGRTD.bindingsThatCare {
							logger.V(5).Info("Enqueuing reference to Binding because CustomTransform changed its GroupResource", "customTransformName", ct.Name, "bindingName", oldBindingName, "oldGroupResource", oldGroupResource, "newGroupResource", groupResource)
							ctc.enqueue(oldBindingName)
						}
					}
				}
			}
			ctc.ctNameToSpec[ct.Name] = ct.Spec
			ctNames = append(ctNames, ct.Name)
		}
		if len(ctAnys) > 1 {
			logger.Error(nil, "Multiple CustomTransform objects apply to one GroupResource", "groupResource", groupResource, "names", ctNames)
		}
		ctc.grToTransformData[groupResource] = grtd
	} else {
		grtd.bindingsThatCare.Insert(bindingName)
	}
	return grtd
}

func (ctc *customTransformCollection) parseRemovesAndUpdateStatus(ctx context.Context, ct *v1alpha1.CustomTransform) (removes []jsonpath.Query) {
	logger := klog.FromContext(ctx)
	ctCopy := ct.DeepCopy()
	ctCopy.Status = v1alpha1.CustomTransformStatus{ObservedGeneration: ct.Generation}
	for idx, queryS := range ct.Spec.Remove {
		query, err := jsonpath.ParseQuery(queryS)
		if err != nil {
			ctCopy.Status.Errors = append(ctCopy.Status.Errors, fmt.Sprintf("Error in spec.remove[%d]: %s", idx, err.Error()))
		} else if len(query) == 0 {
			ctCopy.Status.Errors = append(ctCopy.Status.Errors, fmt.Sprintf("Invalid spec.remove[%d]: it identifies the whole object", idx))
		} else {
			removes = append(removes, query)
		}
	}
	ctEcho, err := ctc.client.UpdateStatus(ctx, ctCopy, metav1.UpdateOptions{FieldManager: ControllerName})
	if err != nil {
		logger.Error(err, "Failed to write status of CustomTransform", "name", ct.Name, "resourceVersion", ct.ResourceVersion, "status", ctCopy.Status)
	} else {
		logger.V(4).Info("Wrote status of CustomTransform", "name", ct.Name, "resourceVersion", ctEcho.ResourceVersion, "observedGeneration", ctCopy.Status.ObservedGeneration)
	}
	return
}

// Call this when the new full set is known, so that obsolete relationships can be deleted
func (ctc *customTransformCollection) setBindingGroupResources(bindingName string, newGroupResources sets.Set[metav1.GroupResource]) {
	ctc.mutex.Lock()
	defer ctc.mutex.Unlock()
	oldGroupResources := ctc.bindingNameToGroupResources[bindingName]
	for groupResource := range oldGroupResources {
		if newGroupResources.Has(groupResource) {
			continue
		}
		// This one is being removed
		if grtd, ok := ctc.grToTransformData[groupResource]; ok {
			grtd.bindingsThatCare.Delete(bindingName)
			// When the set goes empty, time to delete this data
			if grtd.bindingsThatCare.Len() == 0 {
				delete(ctc.grToTransformData, groupResource)
			}
		}
	}
	if len(newGroupResources) == 0 {
		delete(ctc.bindingNameToGroupResources, bindingName)
	} else {
		ctc.bindingNameToGroupResources[bindingName] = newGroupResources
	}
}
