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

package transport

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Transport interface {
	InformerSynced() bool
	WrapObjects([]*unstructured.Unstructured) *unstructured.Unstructured // wrap multiple objects into a singel wrapped object.
}

// TODO
// Transport specific should create informer for the specific wrapped object (e.g., ocm should create informer for ManifestWork)
// then, upon a reconciliation event for WrappedObject, transport specific should parse the name of the edgePlacementDecision object and
// make sure it's pushed to the generic transport controller work queue.

// explanation - let's assume transport controller works perfect and creates manifest work objects as a result of edge placement decision.
// somehow the manifest work object in one of the WECs namespaces was updated/deleted - the transport controller should get the event for the
// edge placement decision, recalculate what's needed and override the changes.

// therefore, transport specific should supply a function that it's informer was synced and it's ready.
// we will need an additional function(s) for using wrapped object (e.g, manifest work) lister, to get the concrete object and parse from it
// the edge placement decision object name. edge placement decision should be written in the wrapped objects either as annotation, or any other alternative.
