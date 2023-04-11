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

import (
	"errors"
	"sync"
)

type Set[Elt comparable] interface {
	Emptyable
	Len() int
	LenIsCheap() bool
	Has(Elt) bool
	Visitable[Elt]
}

// Visitable is a collection that can do an interruptable enumeration of its members.
// The collection may or may not be mutable.
// This view of the collection may or may not have a limited scope of validity.
// This view may or may not have concurrency restrictions.
type Visitable[Elt any] interface {
	// Visit calls the given function on every member, aborting on error
	Visit(func(Elt) error) error
}

// Emptyable is something that can be tested for emptiness.
// The thing may or may not be mutable.
// This view of the thing may or may not have a limited scope of validity.
// This view may or may not have concurrency restrictions.
type Emptyable interface {
	IsEmpty() bool
}

type MutableSet[Elt comparable] interface {
	Set[Elt]
	SetChangeReceiver[Elt]
}

// SetChangeReceiver is kept appraised of changes in a set of T
type SetChangeReceiver[Elt comparable] interface {
	Add(Elt) bool    /* changed */
	Remove(Elt) bool /* changed */
}

// SetChangeReceiverFuncs puts the SetChangeReceiver stamp of approval on a pair of funcs.
// Either may be `nil“, in which case the corresponding SetChangeReceiver method returns `false`.
type SetChangeReceiverFuncs[Elt comparable] struct {
	OnAdd    func(Elt) bool
	OnRemove func(Elt) bool
}

var _ SetChangeReceiver[int] = SetChangeReceiverFuncs[int]{}

func (scrf SetChangeReceiverFuncs[Elt]) Add(elt Elt) bool {
	if scrf.OnAdd != nil {
		return scrf.OnAdd(elt)
	}
	return false
}
func (scrf SetChangeReceiverFuncs[Elt]) Remove(elt Elt) bool {
	if scrf.OnRemove != nil {
		return scrf.OnRemove(elt)
	}
	return false
}

func VisitableLen[Elt any](visitable Visitable[Elt]) int {
	var ans int
	visitable.Visit(func(_ Elt) error {
		ans++
		return nil
	})
	return ans
}

func VisitableHas[Elt comparable](set Visitable[Elt], seek Elt) bool {
	return set.Visit(func(check Elt) error {
		if check == seek {
			return errStop
		}
		return nil
	}) != nil
}

func VisitableGet[Elt any](seq Visitable[Elt], index int) (Elt, bool) {
	var ans Elt
	var currentIndex int
	res := seq.Visit(func(elt Elt) error {
		if currentIndex == index {
			ans = elt
			return errStop
		}
		currentIndex++
		return nil
	})
	return ans, res != nil
}

func SetAddAll[Elt comparable](set MutableSet[Elt], adds Visitable[Elt]) (someNew, allNew bool) {
	someNew, allNew = false, true
	adds.Visit(func(add Elt) error {
		if set.Add(add) {
			someNew = true
		} else {
			allNew = false
		}
		return nil
	})
	return
}

func SetRemoveAll[Elt comparable](set MutableSet[Elt], goners Visitable[Elt]) (someNew, allNew bool) {
	someNew, allNew = false, true
	goners.Visit(func(goner Elt) error {
		if set.Remove(goner) {
			someNew = true
		} else {
			allNew = false
		}
		return nil
	})
	return
}

// Func11Compose11 composes two 1-arg 1-result functions
func Func11Compose11[Type1, Type2, Type3 any](fn1 func(Type1) Type2, fn2 func(Type2) Type3) func(Type1) Type3 {
	return func(val1 Type1) Type3 {
		val2 := fn1(val1)
		return fn2(val2)
	}
}

func NewSetReadonly[Elt comparable](set MutableSet[Elt]) Set[Elt] {
	return SetReadonly[Elt]{set}
}

// SetReadonly is a wrapper that removes the ability to write to the set
type SetReadonly[Elt comparable] struct{ set Set[Elt] }

var _ Set[string] = SetReadonly[string]{}

func (sr SetReadonly[Elt]) IsEmpty() bool                       { return sr.set.IsEmpty() }
func (sr SetReadonly[Elt]) LenIsCheap() bool                    { return sr.set.LenIsCheap() }
func (sr SetReadonly[Elt]) Len() int                            { return sr.set.Len() }
func (sr SetReadonly[Elt]) Has(elt Elt) bool                    { return sr.set.Has(elt) }
func (sr SetReadonly[Elt]) Visit(visitor func(Elt) error) error { return sr.set.Visit(visitor) }

// SetChangeReceiverReverse returns a receiver that acts in the opposite way as the given receiver.
// That is, Add and Remove are swapped.
func SetChangeReceiverReverse[Elt comparable](forward SetChangeReceiver[Elt]) SetChangeReceiver[Elt] {
	return setChangeReceiverReverse[Elt]{forward}
}

type setChangeReceiverReverse[Elt comparable] struct {
	forward SetChangeReceiver[Elt]
}

// SetChangeReceiverFork constructs a SetChangeReceiver that broadcasts incoming changes to the given receivers.
// Each change is passed on to every receiver, and
// combineWithAnd says whether to combine the returned values with AND
// (the alternative is OR).
func SetChangeReceiverFork[Elt comparable](combineWithAnd bool, receivers ...SetChangeReceiver[Elt]) SetChangeReceiver[Elt] {
	return &setChangeReceiverFork[Elt]{
		combineWithAnd: combineWithAnd,
		receivers:      receivers,
	}
}

type setChangeReceiverFork[Elt comparable] struct {
	combineWithAnd bool
	receivers      []SetChangeReceiver[Elt]
}

func (crf *setChangeReceiverFork[Elt]) Add(elt Elt) bool {
	ans := crf.combineWithAnd
	for _, receiver := range crf.receivers {
		change := receiver.Add(elt)
		if crf.combineWithAnd {
			ans = ans && change
		} else {
			ans = ans || change
		}
	}
	return ans
}

func (crf *setChangeReceiverFork[Elt]) Remove(elt Elt) bool {
	ans := crf.combineWithAnd
	for _, receiver := range crf.receivers {
		change := receiver.Remove(elt)
		if crf.combineWithAnd {
			ans = ans && change
		} else {
			ans = ans || change
		}
	}
	return ans
}

func (crr setChangeReceiverReverse[Elt]) Add(elt Elt) bool {
	return crr.forward.Remove(elt)
}

func (crr setChangeReceiverReverse[Elt]) Remove(elt Elt) bool {
	return crr.forward.Add(elt)
}

func TransformSetChangeReceiver[Type1, Type2 comparable](
	transform func(Type1) Type2,
	inner SetChangeReceiver[Type2]) SetChangeReceiver[Type1] {
	return transformSetChangeReceiver[Type1, Type2]{transform, inner}
}

type transformSetChangeReceiver[Type1, Type2 comparable] struct {
	Transform func(Type1) Type2
	Inner     SetChangeReceiver[Type2]
}

var _ SetChangeReceiver[int] = &transformSetChangeReceiver[int, string]{}

func (xr transformSetChangeReceiver[Type1, Type2]) Add(v1 Type1) bool {
	v2 := xr.Transform(v1)
	return xr.Inner.Add(v2)
}

func (xr transformSetChangeReceiver[Type1, Type2]) Remove(v1 Type1) bool {
	v2 := xr.Transform(v1)
	return xr.Inner.Remove(v2)
}

func WrapSetWithMutex[Elt comparable](inner MutableSet[Elt]) MutableSet[Elt] {
	return &setMutex[Elt]{inner: inner}
}

type setMutex[Elt comparable] struct {
	sync.RWMutex
	inner MutableSet[Elt]
}

func (sm *setMutex[Elt]) IsEmpty() bool {
	sm.RLock()
	defer sm.RUnlock()
	return sm.inner.IsEmpty()
}

func (sm *setMutex[Elt]) LenIsCheap() bool {
	sm.RLock()
	defer sm.RUnlock()
	return sm.inner.LenIsCheap()
}

func (sm *setMutex[Elt]) Len() int {
	sm.RLock()
	defer sm.RUnlock()
	return sm.inner.Len()
}

func (sm *setMutex[Elt]) Has(elt Elt) bool {
	sm.RLock()
	defer sm.RUnlock()
	return sm.inner.Has(elt)
}

func (sm *setMutex[Elt]) Visit(visitor func(Elt) error) error {
	sm.RLock()
	defer sm.RUnlock()
	return sm.inner.Visit(visitor)
}

func (sm *setMutex[Elt]) Add(elt Elt) bool {
	sm.Lock()
	defer sm.Unlock()
	return sm.inner.Add(elt)
}

func (sm *setMutex[Elt]) Remove(elt Elt) bool {
	sm.Lock()
	defer sm.Unlock()
	return sm.inner.Add(elt)
}

func SetRotate[Original, Rotated comparable](originalSet Set[Original], rotator Rotator[Original, Rotated]) Set[Rotated] {
	return &setRotate[Original, Rotated]{originalSet, rotator}
}

type setRotate[Original, Rotated comparable] struct {
	originalSet Set[Original]
	rotator     Rotator[Original, Rotated]
}

func (sr *setRotate[Original, Rotated]) IsEmpty() bool    { return sr.originalSet.IsEmpty() }
func (sr *setRotate[Original, Rotated]) LenIsCheap() bool { return sr.originalSet.LenIsCheap() }
func (sr *setRotate[Original, Rotated]) Len() int         { return sr.originalSet.Len() }

func (sr *setRotate[Original, Rotated]) Has(rotatedElt Rotated) bool {
	originalElt := sr.rotator.Second(rotatedElt)
	return sr.originalSet.Has(originalElt)
}

func (sr *setRotate[Original, Rotated]) Visit(visitor func(Rotated) error) error {
	return sr.originalSet.Visit(func(originalElt Original) error {
		rotatedElt := sr.rotator.First(originalElt)
		return visitor(rotatedElt)
	})
}

func TransformVisitable[Original, Transformed any](originalVisitable Visitable[Original], transform func(Original) Transformed) Visitable[Transformed] {
	return &transformVisitable[Original, Transformed]{originalVisitable, transform}
}

type transformVisitable[Original, Transformed any] struct {
	originalVisitable Visitable[Original]
	transform         func(Original) Transformed
}

func (tv *transformVisitable[Original, Transformed]) Visit(visitor func(Transformed) error) error {
	return tv.originalVisitable.Visit(func(originalElt Original) error {
		transformedElt := tv.transform(originalElt)
		return visitor(transformedElt)
	})
}

// Reducer is something that crunches a collection down into one value
type Reducer[Elt any, Ans any] func(Visitable[Elt]) Ans

// NewReducer makes Reducer that works with accumulator values
func ValueReducer[Elt any, Accum any, Ans any](initialize func() Accum, add func(Accum, Elt) Accum, finish func(Accum) Ans) Reducer[Elt, Ans] {
	return func(elts Visitable[Elt]) Ans {
		accum := initialize()
		elts.Visit(func(elt Elt) error {
			accum = add(accum, elt)
			return nil
		})
		return finish(accum)
	}
}

// StatefulReducer makes a Reducer that works with a stateful accumulator
func StatefulReducer[Elt any, Accum any, Ans any](initialize func() Accum, add func(Accum, Elt), finish func(Accum) Ans) Reducer[Elt, Ans] {
	return func(elts Visitable[Elt]) Ans {
		accum := initialize()
		elts.Visit(func(elt Elt) error {
			add(accum, elt)
			return nil
		})
		return finish(accum)
	}
}

func SetEnumerateDifferences[Elt comparable](left, right Set[Elt], receiver SetChangeReceiver[Elt]) {
	left.Visit(func(elt Elt) error {
		has := right.Has(elt)
		if !has {
			receiver.Remove(elt)
		}
		return nil
	})
	right.Visit(func(elt Elt) error {
		has := left.Has(elt)
		if !has {
			receiver.Add(elt)
		}
		return nil
	})
}

var errStop = errors.New("it is done")

func SetLessOrEqual[Elt comparable](set1, set2 Set[Elt]) bool {
	return set1.Visit(func(elt Elt) error {
		if !set2.Has(elt) {
			return errStop
		}
		return nil
	}) == nil
}

func SetCompare[Elt comparable](set1, set2 Set[Elt]) Comparison {
	return Comparison{
		LessOrEqual:    SetLessOrEqual(set1, set2),
		GreaterOrEqual: SetLessOrEqual(set2, set1),
	}
}

func SetEqual[Elt comparable](set1, set2 Set[Elt]) bool {
	if set1.LenIsCheap() {
		return set1.Len() == set2.Len() && SetLessOrEqual(set1, set2)
	}
	return SetCompare[Elt](set1, set2).IsEqual()
}

type Comparison struct{ LessOrEqual, GreaterOrEqual bool }

func (comp Comparison) Reverse() Comparison {
	return Comparison{LessOrEqual: comp.GreaterOrEqual, GreaterOrEqual: comp.LessOrEqual}
}

func (comp Comparison) IsEqual() bool           { return comp.LessOrEqual && comp.GreaterOrEqual }
func (comp Comparison) IsStrictlyLess() bool    { return comp.LessOrEqual && !comp.GreaterOrEqual }
func (comp Comparison) IsStrictlyGreater() bool { return comp.GreaterOrEqual && !comp.LessOrEqual }
func (comp Comparison) IsRelated() bool         { return comp.LessOrEqual || comp.GreaterOrEqual }

func ToHeap[Val any](val Val) *Val {
	return &val
}
