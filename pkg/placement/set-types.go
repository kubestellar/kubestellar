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

	"k8s.io/klog/v2"
)

type Set[Elt any] interface {
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

type MutableSet[Elt any] interface {
	Set[Elt]
	SetWriter[Elt]
}

// SetChangeReceiver is SetWriter shorn of returned information and refactored
// to a simpler signature.  The `bool` is `true` for additions to the set,
// `false` for removals.
type SetChangeReceiver[Elt any] func(bool, Elt)

// SetWriter is the write aspect of a Set
type SetWriter[Elt any] interface {
	Add(Elt) bool    /* changed */
	Remove(Elt) bool /* changed */
}

func NewSetWriterFuncs[Elt any](OnAdd, OnRemove func(Elt) bool) SetWriterFuncs[Elt] {
	return SetWriterFuncs[Elt]{OnAdd, OnRemove}
}

// SetWriterFuncs puts the SetWriter stamp of approval on a pair of funcs.
// Either may be `nil“, in which case the corresponding SetWriter method returns `false`.
type SetWriterFuncs[Elt any] struct {
	OnAdd    func(Elt) bool
	OnRemove func(Elt) bool
}

var _ SetWriter[int] = SetWriterFuncs[int]{}

func (scrf SetWriterFuncs[Elt]) Add(elt Elt) bool {
	if scrf.OnAdd != nil {
		return scrf.OnAdd(elt)
	}
	return false
}
func (scrf SetWriterFuncs[Elt]) Remove(elt Elt) bool {
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

func SetAddAll[Elt any](set MutableSet[Elt], adds Visitable[Elt]) (someNew, allNew bool) {
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

func SetRemoveAll[Elt any](set MutableSet[Elt], goners Visitable[Elt]) (someNew, allNew bool) {
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

func NewSetReadonly[Elt any](set MutableSet[Elt]) Set[Elt] {
	return SetReadonly[Elt]{set}
}

// SetReadonly is a wrapper that removes the ability to write to the set
type SetReadonly[Elt any] struct{ set Set[Elt] }

var _ Set[string] = SetReadonly[string]{}

func (sr SetReadonly[Elt]) IsEmpty() bool                       { return sr.set.IsEmpty() }
func (sr SetReadonly[Elt]) LenIsCheap() bool                    { return sr.set.LenIsCheap() }
func (sr SetReadonly[Elt]) Len() int                            { return sr.set.Len() }
func (sr SetReadonly[Elt]) Has(elt Elt) bool                    { return sr.set.Has(elt) }
func (sr SetReadonly[Elt]) Visit(visitor func(Elt) error) error { return sr.set.Visit(visitor) }

// SetWriterReverse returns a receiver that acts in the opposite way as the given receiver.
// That is, Add and Remove are swapped.
func SetWriterReverse[Elt any](forward SetWriter[Elt]) SetWriter[Elt] {
	return setChangeReceiverReverse[Elt]{forward}
}

type setChangeReceiverReverse[Elt any] struct {
	forward SetWriter[Elt]
}

// SetWriterFork constructs a SetWriter that broadcasts incoming changes to the given receivers.
// Each change is passed on to every receiver, and
// combineWithAnd says whether to combine the returned values with AND
// (the alternative is OR).
func SetWriterFork[Elt any](combineWithAnd bool, receivers ...SetWriter[Elt]) SetWriter[Elt] {
	return &setChangeReceiverFork[Elt]{
		combineWithAnd: combineWithAnd,
		receivers:      receivers,
	}
}

type setChangeReceiverFork[Elt any] struct {
	combineWithAnd bool
	receivers      []SetWriter[Elt]
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

func TransformSetWriter[Type1, Type2 any](
	transform func(Type1) Type2,
	inner SetWriter[Type2]) SetWriter[Type1] {
	return transformSetWriter[Type1, Type2]{transform, inner}
}

type transformSetWriter[Type1, Type2 any] struct {
	Transform func(Type1) Type2
	Inner     SetWriter[Type2]
}

var _ SetWriter[int] = &transformSetWriter[int, string]{}

func (xr transformSetWriter[Type1, Type2]) Add(v1 Type1) bool {
	v2 := xr.Transform(v1)
	return xr.Inner.Add(v2)
}

func (xr transformSetWriter[Type1, Type2]) Remove(v1 Type1) bool {
	v2 := xr.Transform(v1)
	return xr.Inner.Remove(v2)
}

func WrapSetWithMutex[Elt any](inner MutableSet[Elt]) MutableSet[Elt] {
	return &setMutex[Elt]{inner: inner}
}

type setMutex[Elt any] struct {
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

func SetRotate[Original, Rotated any](originalSet Set[Original], rotator Rotator[Original, Rotated]) Set[Rotated] {
	return &setRotate[Original, Rotated]{originalSet, rotator}
}

type setRotate[Original, Rotated any] struct {
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

func NewLoggingSetWriter[Elt any](setName string, logger klog.Logger) SetWriter[Elt] {
	logger = logger.WithValues("set", setName)
	return loggingSetWriter[Elt]{logger}
}

type loggingSetWriter[Elt any] struct {
	logger klog.Logger
}

var _ SetWriter[int] = loggingSetWriter[int]{}

func (lcr loggingSetWriter[Elt]) Add(elt Elt) bool {
	lcr.logger.Info("Add", "elt", elt)
	return true
}

func (lcr loggingSetWriter[Elt]) Remove(elt Elt) bool {
	lcr.logger.Info("Remove", "elt", elt)
	return true
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

func SetEnumerateDifferences[Elt any](left, right Set[Elt], receiver SetWriter[Elt]) {
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

func SetLessOrEqual[Elt any](set1, set2 Set[Elt]) bool {
	return set1.Visit(func(elt Elt) error {
		if !set2.Has(elt) {
			return errStop
		}
		return nil
	}) == nil
}

func SetCompare[Elt any](set1, set2 Set[Elt]) Comparison {
	return Comparison{
		LessOrEqual:    SetLessOrEqual(set1, set2),
		GreaterOrEqual: SetLessOrEqual(set2, set1),
	}
}

func SetEqual[Elt any](set1, set2 Set[Elt]) bool {
	if set1.LenIsCheap() {
		return set1.Len() == set2.Len() && SetLessOrEqual(set1, set2)
	}
	return SetCompare[Elt](set1, set2).IsEqual()
}

func SetIntersection[Elt any](set1, set2 Set[Elt]) Set[Elt] {
	return setIntersection[Elt]{set1, set2}
}

type setIntersection[Elt any] struct {
	set1 Set[Elt]
	set2 Set[Elt]
}

func (si setIntersection[Elt]) IsEmpty() bool {
	if si.set2.IsEmpty() {
		return true
	}
	return si.Visit(func(elt Elt) error {
		return errStop
	}) == nil
}

func (si setIntersection[Elt]) LenIsCheap() bool { return false }

func (si setIntersection[Elt]) Len() int {
	return VisitableLen[Elt](si)
}

func (si setIntersection[Elt]) Has(elt Elt) bool { return si.set1.Has(elt) && si.set2.Has(elt) }

func (si setIntersection[Elt]) Visit(visitor func(Elt) error) error {
	if si.set2.IsEmpty() {
		return nil
	}
	return si.set1.Visit(func(elt Elt) error {
		if si.set2.Has(elt) {
			return visitor(elt)
		}
		return nil
	})
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
