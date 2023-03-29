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
		byX:      map[ColX]*dynamicJoinPerEdge[ColY]{},
		byY:      map[ColY]*dynamicJoinPerCenter[ColX, ColZ]{},
		byZ:      map[ColZ]*dynamicJoinPerEdge[ColY]{},
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
// For every (x,y): byX[x].zs contains y iff byY[y].xs contains x.
// For every (z,y): byZ[z].ys contains y iff byY[y].zs contains z.
type dynamicJoin[ColX comparable, ColY comparable, ColZ comparable] struct {
	logger   klog.Logger
	receiver TripleSetChangeReceiver[ColX, ColY, ColZ]
	byX      map[ColX]*dynamicJoinPerEdge[ColY]
	byY      map[ColY]*dynamicJoinPerCenter[ColX, ColZ]
	byZ      map[ColZ]*dynamicJoinPerEdge[ColY]
}

type dynamicJoinPerCenter[ColX comparable, ColZ comparable] struct {
	xs map[ColX]Empty
	zs map[ColZ]Empty
}

type dynamicJoinPerEdge[ColY comparable] struct {
	ys map[ColY]Empty
}

func NewDynamicJoinPerEdge[ColY comparable]() *dynamicJoinPerEdge[ColY] {
	return &dynamicJoinPerEdge[ColY]{
		ys: map[ColY]Empty{},
	}
}

type dynamicJoinCenterIndex[ColA comparable, ColB comparable, ColC comparable] interface {
	GetB(ColB) dynamicJoinCenterIndexEntry[ColA, ColC]
	RemoveB(ColB) bool // whether it was present
}

type dynamicJoinCenterIndexEntry[ColX comparable, ColZ comparable] interface {
	IsEmpty() bool
	HasLeft(ColX) bool
	InsertLeft(ColX)
	RemoveLeft(ColX)
	VisitRights(func(ColZ))
}

type dynamicJoinCenterIndexForward[ColA comparable, ColB comparable, ColC comparable] map[ColB]*dynamicJoinPerCenter[ColA, ColC]

func (index dynamicJoinCenterIndexForward[ColA, ColB, ColC]) GetB(b ColB) dynamicJoinCenterIndexEntry[ColA, ColC] {
	entry, found := index[b]
	if !found {
		entry := &dynamicJoinPerCenter[ColA, ColC]{
			xs: map[ColA]Empty{},
			zs: map[ColC]Empty{},
		}
		index[b] = entry
	}
	return entry
}

func (index dynamicJoinCenterIndexForward[ColA, ColB, ColC]) RemoveB(b ColB) bool {
	_, found := index[b]
	if found {
		delete(index, b)
	}
	return found
}

func (jpc *dynamicJoinPerCenter[ColX, ColZ]) IsEmpty() bool {
	return len(jpc.xs) == 0 && len(jpc.zs) == 0
}

func (jpc *dynamicJoinPerCenter[ColX, ColZ]) HasLeft(x ColX) bool {
	_, has := jpc.xs[x]
	return has
}

func (jpc *dynamicJoinPerCenter[ColX, ColZ]) InsertLeft(x ColX) {
	jpc.xs[x] = Empty{}
}

func (jpc *dynamicJoinPerCenter[ColX, ColZ]) RemoveLeft(x ColX) {
	delete(jpc.xs, x)
}

func (jpc *dynamicJoinPerCenter[ColX, ColZ]) VisitRights(visitor func(ColZ)) {
	for z := range jpc.zs {
		visitor(z)
	}
}

type dynamicJoinCenterIndexReverse[ColA comparable, ColB comparable, ColC comparable] map[ColB]*dynamicJoinPerCenter[ColA, ColC]

func (index dynamicJoinCenterIndexReverse[ColX, ColY, ColZ]) GetB(b ColY) dynamicJoinCenterIndexEntry[ColZ, ColX] {
	forward, found := index[b]
	if !found {
		forward := &dynamicJoinPerCenter[ColX, ColZ]{
			xs: map[ColX]Empty{},
			zs: map[ColZ]Empty{},
		}
		index[b] = forward
	}
	return reverseCenterEntry[ColX, ColZ]{forward}
}

func (index dynamicJoinCenterIndexReverse[ColX, ColY, ColZ]) RemoveB(b ColY) bool {
	_, found := index[b]
	if found {
		delete(index, b)
	}
	return found
}

type reverseCenterEntry[ColX comparable, ColZ comparable] struct {
	forward *dynamicJoinPerCenter[ColX, ColZ]
}

func (rce reverseCenterEntry[ColX, ColZ]) IsEmpty() bool {
	return len(rce.forward.xs) == 0 && len(rce.forward.zs) == 0
}

func (rce reverseCenterEntry[ColX, ColZ]) HasLeft(z ColZ) bool {
	_, has := rce.forward.zs[z]
	return has
}

func (rce reverseCenterEntry[ColX, ColZ]) InsertLeft(z ColZ) {
	rce.forward.zs[z] = Empty{}
}
func (rce reverseCenterEntry[ColX, ColZ]) RemoveLeft(z ColZ) {
	delete(rce.forward.zs, z)
}

func (rce reverseCenterEntry[ColX, ColZ]) VisitRights(visitor func(ColX)) {
	for x := range rce.forward.xs {
		visitor(x)
	}
}

type dynamicJoinXY[ColX, ColY, ColZ comparable] struct{ *dynamicJoin[ColX, ColY, ColZ] }
type dynamicJoinYZ[ColX, ColY, ColZ comparable] struct{ *dynamicJoin[ColX, ColY, ColZ] }

func (dj dynamicJoinXY[ColX, ColY, ColZ]) Add(xy Pair[ColX, ColY]) {
	addABC[ColX, ColY, ColZ](dj.logger, dj.byX, xy.First, dynamicJoinCenterIndexForward[ColX, ColY, ColZ](dj.byY), xy.Second, dj.byZ, dj.receiver)
}

func (dj dynamicJoinYZ[ColX, ColY, ColZ]) Add(yz Pair[ColY, ColZ]) {
	addABC[ColZ, ColY, ColX](dj.logger, dj.byZ, yz.Second, dynamicJoinCenterIndexReverse[ColX, ColY, ColZ](dj.byY), yz.First, dj.byX, TripleSetChangeReceiverReverse[ColX, ColY, ColZ]{dj.receiver})
}

func addABC[ColA comparable, ColB comparable, ColC comparable](
	logger klog.Logger,
	byA map[ColA]*dynamicJoinPerEdge[ColB],
	a ColA,
	byB dynamicJoinCenterIndex[ColA, ColB, ColC],
	b ColB,
	byC map[ColC]*dynamicJoinPerEdge[ColB],
	receiver TripleSetChangeReceiver[ColA, ColB, ColC],
) {
	aData, aFound := byA[a]
	if !aFound {
		aData = NewDynamicJoinPerEdge[ColB]()
		byA[a] = aData
	}
	bData := byB.GetB(b)
	_, bForA := aData.ys[b]
	var aForB bool = bData.HasLeft(a)
	if bForA != aForB {
		logger.Error(nil, "Impossible inconsistency", "a", a, "b", b, "aForB", aForB, "bForA", bForA)
		return
	} else if aForB {
		return // no news
	}
	bData.InsertLeft(a)
	aData.ys[b] = Empty{}
	bData.VisitRights(func(c ColC) {
		receiver.Add(a, b, c)
	})
}

func (dj dynamicJoinXY[ColX, ColY, ColZ]) Remove(xy Pair[ColX, ColY]) {
	removeABC[ColX, ColY, ColZ](dj.logger, dj.byX, xy.First, dynamicJoinCenterIndexForward[ColX, ColY, ColZ](dj.byY), xy.Second, dj.byZ, dj.receiver)
}

func (dj dynamicJoinYZ[ColX, ColY, ColZ]) Remove(yz Pair[ColY, ColZ]) {
	removeABC[ColZ, ColY, ColX](dj.logger, dj.byZ, yz.Second, dynamicJoinCenterIndexReverse[ColX, ColY, ColZ](dj.byY), yz.First, dj.byX, TripleSetChangeReceiverReverse[ColX, ColY, ColZ]{dj.receiver})
}

func removeABC[ColA comparable, ColB comparable, ColC comparable](
	logger klog.Logger,
	byA map[ColA]*dynamicJoinPerEdge[ColB],
	a ColA,
	byB dynamicJoinCenterIndex[ColA, ColB, ColC],
	b ColB,
	byC map[ColC]*dynamicJoinPerEdge[ColB],
	receiver TripleSetChangeReceiver[ColA, ColB, ColC],
) {
	aData, aFound := byA[a]
	if !aFound {
		aData = NewDynamicJoinPerEdge[ColB]()
		byA[a] = aData
	}
	bData := byB.GetB(b)
	_, bForA := aData.ys[b]
	var aForB bool = bData.HasLeft(a)
	if bForA != aForB {
		logger.Error(nil, "Impossible inconsistency", "a", a, "b", b, "aForB", aForB, "bForA", bForA)
		return
	} else if aForB { // Need to remove
		delete(aData.ys, b)
		bData.RemoveLeft(a)
		bData.VisitRights(func(c ColC) {
			receiver.Remove(a, b, c)
		})
	}
	if len(aData.ys) == 0 {
		delete(byA, a)
	}
	if bData.IsEmpty() {
		byB.RemoveB(b)
	}
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

type mapSet[Elt comparable] map[Elt]Empty

func (ms mapSet[Elt]) IsEmpty() bool { return len(ms) == 0 }

func (ms mapSet[Elt]) Has(elt Elt) bool {
	_, has := ms[elt]
	return has
}

func (ms mapSet[Elt]) Visit(visitor func(Elt) error) error {
	for element := range ms {
		if err := visitor(element); err != nil {
			return err
		}
	}
	return nil
}

func (ms mapSet[Elt]) Insert(elts ...Elt) {
	for _, elt := range elts {
		ms[elt] = Empty{}
	}
}

func (ms mapSet[Elt]) Remove(elts ...Elt) {
	for _, elt := range elts {
		delete(ms, elt)
	}
}
