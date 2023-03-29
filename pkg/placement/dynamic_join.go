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
	"k8s.io/klog/v2"
)

// NewDynamicJoin constructs a data structure that incrementally maintains an equijoin.
// Given a receiver of changes to the result of the equijoin between two two-column tables,
// this function returns receivers of changes to the two tables.
// In other words, this joins the change streams of two tables to produce the change stream
// of the join of those two tables --- in a passive stance (i.e., is in terms of the stream receivers).
// Note: the uniformity of the input and output types means that this can be chained.
func NewDynamicJoin[ColX comparable, ColY comparable, ColZ comparable](logger klog.Logger, receiver SetChangeReceiver[Pair[ColX, ColZ]]) (SetChangeReceiver[Pair[ColX, ColY]], SetChangeReceiver[Pair[ColY, ColZ]]) {
	projector := NewProjectIncremental[Pair[ColX, ColZ], ColY](receiver)
	indexer := NewIndex123by13to2s(projector)
	return NewDynamicFullJoin(logger, indexer)
}

// NewDynamicFullJoin is like NewDynamicJoin but passes along the set of middle values too.
func NewDynamicFullJoin[ColX comparable, ColY comparable, ColZ comparable](logger klog.Logger, receiver TripleSetChangeReceiver[ColX, ColY, ColZ]) (SetChangeReceiver[Pair[ColX, ColY]], SetChangeReceiver[Pair[ColY, ColZ]]) {
	dj := &dynamicJoin[ColX, ColY, ColZ]{
		logger:   logger,
		receiver: receiver,
		xyReln:   NewMapRelation2[ColX, ColY](),
		yzReln:   NewMapRelation2[ColY, ColZ](),
	}
	return dynamicJoinXY[ColX, ColY, ColZ]{dj}, dynamicJoinYZ[ColX, ColY, ColZ]{dj}
}

// TripleSetChangeReceiver is given a series of changes to a set of triples
type TripleSetChangeReceiver[First any, Second any, Third any] interface {
	Add(First, Second, Third)
	Remove(First, Second, Third)
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

// dynamicJoin implements DynamicJoin.
// It buffers the two incoming relations and passes on changes.
type dynamicJoin[ColX comparable, ColY comparable, ColZ comparable] struct {
	logger   klog.Logger
	receiver TripleSetChangeReceiver[ColX, ColY, ColZ]
	xyReln   *MapRelation2[ColX, ColY]
	yzReln   *MapRelation2[ColY, ColZ]
}

type dynamicJoinXY[ColX, ColY, ColZ comparable] struct{ *dynamicJoin[ColX, ColY, ColZ] }
type dynamicJoinYZ[ColX, ColY, ColZ comparable] struct{ *dynamicJoin[ColX, ColY, ColZ] }

func (dj dynamicJoinXY[ColX, ColY, ColZ]) Add(xy Pair[ColX, ColY]) bool {
	return addXYZ[ColX, ColY, ColZ](dj.logger, dj.xyReln, dj.yzReln, xy, dj.receiver)
}

func (dj dynamicJoinYZ[ColX, ColY, ColZ]) Add(yz Pair[ColY, ColZ]) bool {
	return addXYZ[ColZ, ColY, ColX](dj.logger, MutableRelation2Reverse[ColY, ColZ](dj.yzReln), MutableRelation2Reverse[ColX, ColY](dj.xyReln), yz.Reverse(), TripleSetChangeReceiverReverse[ColX, ColY, ColZ]{dj.receiver})
}

func addXYZ[ColX, ColY, ColZ comparable](
	logger klog.Logger,
	xyReln MutableRelation2[ColX, ColY],
	yzReln MutableRelation2[ColY, ColZ],
	xy Pair[ColX, ColY],
	receiver TripleSetChangeReceiver[ColX, ColY, ColZ],
) bool {
	if !xyReln.Add(xy) {
		return false
	}
	yzReln.Visit1to2(xy.Second, func(z ColZ) error {
		receiver.Add(xy.First, xy.Second, z)
		return nil
	})
	return true
}

func (dj dynamicJoinXY[ColX, ColY, ColZ]) Remove(xy Pair[ColX, ColY]) bool {
	var xyReln MutableRelation2[ColX, ColY] = dj.xyReln
	var yzReln MutableRelation2[ColY, ColZ] = dj.yzReln
	return removeXYZ(dj.logger, xyReln, yzReln, xy, dj.receiver)
}

func (dj dynamicJoinYZ[ColX, ColY, ColZ]) Remove(yz Pair[ColY, ColZ]) bool {
	return removeXYZ[ColZ, ColY, ColX](dj.logger, MutableRelation2Reverse[ColY, ColZ](dj.yzReln), MutableRelation2Reverse[ColX, ColY](dj.xyReln), yz.Reverse(), TripleSetChangeReceiverReverse[ColX, ColY, ColZ]{dj.receiver})
}

func removeXYZ[ColX, ColY, ColZ comparable](
	logger klog.Logger,
	xyReln MutableRelation2[ColX, ColY],
	yzReln MutableRelation2[ColY, ColZ],
	xy Pair[ColX, ColY],
	receiver TripleSetChangeReceiver[ColX, ColY, ColZ],
) bool {
	if !xyReln.Remove(xy) {
		return false
	}
	yzReln.Visit1to2(xy.Second, func(z ColZ) error {
		receiver.Remove(xy.First, xy.Second, z)
		return nil
	})
	return true
}

type TripleSetChangeReceiverReverse[Left any, Middle any, Right any] struct {
	forward TripleSetChangeReceiver[Left, Middle, Right]
}

func (prr TripleSetChangeReceiverReverse[Left, Middle, Right]) Add(right Right, middle Middle, left Left) {
	prr.forward.Add(left, middle, right)
}

func (prr TripleSetChangeReceiverReverse[Left, Middle, Right]) Remove(right Right, middle Middle, left Left) {
	prr.forward.Remove(left, middle, right)
}
