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

package placement

// MappingToEmptyableReceiver is a mapping receiver specialized to the case
// where the value is an collection that is VisitableEmptyable and one bit of
// history is reported with each change.
type MappingToVisitableEmptyableReceiver[Key comparable, ValElt any] interface {
	// Receive asserts the current value associated with a key and whether
	// the associated value was previously empty.
	Receive(key Key, val VisitableEmptyable[ValElt], wasEmpty bool)
}

// VisitableEmptyable is a collection is both Visitable and Emptyable.
type VisitableEmptyable[Elt any] interface {
	Emptyable
	Visitable[Elt]
}

// NewIndex123by13to2s maintains an index into a three-column table.
// The index is keyed by the first and second columns and each entry holds
// the set of associated middle column values.
func NewIndex123by13to2s[ColA comparable, ColB comparable, ColC comparable](inner MappingToVisitableEmptyableReceiver[Pair[ColA, ColC], ColB]) TripleSetChangeReceiver[ColA, ColB, ColC] {
	return &index123by13to2s[ColA, ColB, ColC]{
		inner: inner,
		by13:  map[Pair[ColA, ColC]]mapSet[ColB]{},
	}
}

type index123by13to2s[ColA comparable, ColB comparable, ColC comparable] struct {
	inner MappingToVisitableEmptyableReceiver[Pair[ColA, ColC], ColB]
	by13  map[Pair[ColA, ColC]]mapSet[ColB]
}

func (index *index123by13to2s[ColA, ColB, ColC]) Add(first ColA, second ColB, third ColC) {
	key := Pair[ColA, ColC]{first, third}
	seconds := index.by13[key]
	if seconds == nil {
		seconds = mapSet[ColB]{}
		index.by13[key] = seconds
	}
	had := seconds.Has(second)
	if !had {
		wasEmpty := seconds.IsEmpty()
		seconds.Insert(second)
		index.inner.Receive(key, seconds, wasEmpty)
	}
}

func (index *index123by13to2s[ColA, ColB, ColC]) Remove(first ColA, second ColB, third ColC) {
	key := Pair[ColA, ColC]{first, third}
	seconds := index.by13[key]
	if seconds == nil {
		return
	}
	had := seconds.Has(second)
	if had {
		wasEmpty := seconds.IsEmpty()
		seconds.Remove(second)
		index.inner.Receive(key, seconds, wasEmpty)
	}
}

// NewProjectIncremental maintains the projection from an index to its key set,
// relying on the index to report a bit of state.
// This is stated in passive terms: it transforms a receiver of the key set into
// a receiver of index updates.
func NewProjectIncremental[ColA comparable, ColB comparable](inner SetChangeReceiver[ColA]) MappingToVisitableEmptyableReceiver[ColA, ColB] {
	return projectIncremental[ColA, ColB]{inner: inner}
}

type projectIncremental[ColA comparable, ColB comparable] struct {
	inner SetChangeReceiver[ColA]
}

func (proj projectIncremental[ColA, ColB]) Receive(key ColA, vals VisitableEmptyable[ColB], wasEmpty bool) {
	if wasEmpty {
		if !vals.IsEmpty() {
			proj.inner.Add(key)
		}
	} else if vals.IsEmpty() {
		proj.inner.Remove(key)
	}
}
