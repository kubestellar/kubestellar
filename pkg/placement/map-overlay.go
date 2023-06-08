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

// MapFunctional is a map that is immutable but has operations
// to construct a new map that differs from the receiver in a
// defined way.
type MapFunctional[Key, Value any] interface {
	Map[Key, Value]

	// Put returns a new map that has the given entry
	// and otherwise agrees with the receiver.
	Put(Key, Value) MapFunctional[Key, Value]

	// Delete returns a new map that has no entry for the given key
	// and otherwise agrees with the receiver.
	Delete(Key) MapFunctional[Key, Value]
}

// NewMapOverlay returns a MapFunctional that works by maintaining an overlay
// of limited size.  The cost of Put and Delete is O(N) but it is not as bad
// as making a complete new copy.
func NewMapOverlay[Key comparable, Value any]() MapFunctional[Key, Value] {
	return &MapOverlay[Key, Value]{base: NewMapMap[Key, Value](nil)}
}

// MapOverlay is base overridden by overlay then overridden by deletions.
// Every field holds an immutable value.
// A more efficient implementation would use red-black trees.
type MapOverlay[Key comparable, Value any] struct {
	base      Map[Key, Value]
	overlay   Map[Key, Value]
	deletions Set[Key]
	len       int
}

func (mo *MapOverlay[Key, Value]) Get(key Key) (val Value, have bool) {
	var zero Value
	if mo.deletions != nil && mo.deletions.Has(key) {
		return zero, false
	}
	if mo.overlay != nil {
		val, have = mo.overlay.Get(key)
		if have {
			return
		}
	}
	return mo.base.Get(key)
}

func (mo *MapOverlay[Key, Value]) IsEmpty() bool    { return mo.len == 0 }
func (mo *MapOverlay[Key, Value]) LenIsCheap() bool { return true }
func (mo *MapOverlay[Key, Value]) Len() int         { return mo.len }

func (mo *MapOverlay[Key, Value]) Visit(visitor func(Pair[Key, Value]) error) error {
	if mo.overlay != nil {
		err := mo.overlay.Visit(func(tup Pair[Key, Value]) error {
			if mo.deletions != nil && mo.deletions.Has(tup.First) {
				return nil
			}
			return visitor(tup)
		})
		if err != nil {
			return err
		}
	}
	return mo.base.Visit(func(tup Pair[Key, Value]) error {
		if mo.deletions != nil && mo.deletions.Has(tup.First) {
			return nil
		}
		if mo.overlay != nil {
			if _, has := mo.overlay.Get(tup.First); has {
				return nil
			}
		}
		return visitor(tup)
	})
}

func (mo *MapOverlay[Key, Value]) Put(key Key, val Value) MapFunctional[Key, Value] {
	lenThresh := mo.len / 3
	var exns int
	if mo.deletions != nil {
		exns += mo.deletions.Len()
	}
	if mo.overlay != nil {
		exns += mo.overlay.Len()
	}
	if exns+1 > lenThresh {
		newBase := MapMapCopy[Key, Value](nil, mo)
		newBase.Put(key, val)
		return &MapOverlay[Key, Value]{base: newBase, len: newBase.Len()}
	}
	ans := *mo
	if mo.deletions != nil && mo.deletions.Has(key) {
		if mo.deletions.Len() > 1 {
			newDeletions := MapSetCopy[Key](mo.deletions)
			newDeletions.Remove(key)
			ans.deletions = newDeletions
		} else {
			ans.deletions = nil
		}
	}
	var newOverlay MapMap[Key, Value]
	if mo.overlay != nil {
		newOverlay = MapMapCopy[Key, Value](nil, mo.overlay)
	} else {
		newOverlay = NewMapMap[Key, Value](nil)
	}
	newOverlay.Put(key, val)
	ans.overlay = newOverlay
	if _, had := mo.Get(key); !had {
		ans.len++
	}
	return &ans
}

func (mo *MapOverlay[Key, Value]) Delete(key Key) MapFunctional[Key, Value] {
	if _, has := mo.Get(key); !has {
		return mo
	}
	lenThresh := mo.len / 3
	var exns int
	if mo.deletions != nil {
		exns += mo.deletions.Len()
	}
	if mo.overlay != nil {
		exns += mo.overlay.Len()
	}
	if exns+1 > lenThresh {
		newBase := MapMapCopy[Key, Value](nil, mo)
		newBase.Delete(key)
		return &MapOverlay[Key, Value]{base: newBase, len: newBase.Len()}
	}
	ans := *mo
	if _, has := mo.base.Get(key); has {
		// Have to add to deletions.
		// Don't bother with overlay.
		if mo.deletions != nil {
			newDeletions := MapSetCopy[Key](mo.deletions)
			newDeletions.Add(key)
			ans.deletions = newDeletions
		} else {
			ans.deletions = NewMapSet[Key](key)
		}
	} else { // Present in overlay, remove.
		if ans.overlay.Len() == 1 {
			ans.overlay = nil
		} else {
			newOverlay := MapMapCopy[Key, Value](nil, ans.overlay)
			newOverlay.Delete(key)
			ans.overlay = newOverlay
		}
	}
	ans.len--
	return &ans
}
