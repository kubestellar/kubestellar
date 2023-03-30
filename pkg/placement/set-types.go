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

import "errors"

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
type SetChangeReceiver[T comparable] interface {
	Add(T) bool    /* changed */
	Remove(T) bool /* changed */
}

func VisitableHas[Elt comparable](set Visitable[Elt], seek Elt) bool {
	return set.Visit(func(check Elt) error {
		if check == seek {
			return errStop
		}
		return nil
	}) != nil
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

// Identity1 is useful in reduers where the accumulator has the same type as the result
func Identity1[Val any](val Val) Val { return val }

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
