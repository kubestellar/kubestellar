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

func NewSliceDifferencerParametric[Elt any](isEqual func(Elt, Elt) bool, changeReceiver SetChangeReceiver[Elt], initial []Elt) Receiver[ /*immutable*/ []Elt] {
	return &sliceDifferencer[Elt]{isEqual, changeReceiver, SliceCopy(initial)}
}

type sliceDifferencer[Elt any] struct {
	isEqual        func(Elt, Elt) bool
	changeReceiver SetChangeReceiver[Elt]
	current        []Elt
}

func (sd *sliceDifferencer[Elt]) Receive(newWhole /*immutable*/ []Elt) {
	for _, oldElt := range sd.current {
		if !SliceContainsParametric(sd.isEqual, newWhole, oldElt) {
			sd.changeReceiver(false, oldElt)
		}
	}
	for _, newElt := range newWhole {
		if !SliceContainsParametric(sd.isEqual, sd.current, newElt) {
			sd.changeReceiver(true, newElt)
		}
	}
	sd.current = newWhole // only legitimate because newWhole is immutable
}
