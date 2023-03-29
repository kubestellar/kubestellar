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

// DynamicJoin maintains the equijoin between two tables, one with columns X and Y
// and the other with columns Y and Z.
// Not safe for concurrent access.
type DynamicJoin[ColX comparable, ColY comparable, ColZ comparable] interface {
	AddXY(ColX, ColY)
	RemoveXY(ColX, ColY)
	AddYZ(ColY, ColZ)
	RemoveYZ(ColY, ColZ)
}

// dynamicJoin implements DynamicJoin.
// For every (x,y): byX[x].zs contains y iff byY[y].xs contains x.
// For every (z,y): byZ[z].ys contains y iff byY[y].zs contains z.
type dynamicJoin[ColX comparable, ColY comparable, ColZ comparable] struct {
	logger   klog.Logger
	receiver PairSetChangeReceiver[ColX, ColZ]
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

func NewDynamicJoin[ColX comparable, ColY comparable, ColZ comparable](logger klog.Logger, receiver PairSetChangeReceiver[ColX, ColZ]) DynamicJoin[ColX, ColY, ColZ] {
	dj := &dynamicJoin[ColX, ColY, ColZ]{
		logger:   logger,
		receiver: receiver,
		byX:      map[ColX]*dynamicJoinPerEdge[ColY]{},
		byY:      map[ColY]*dynamicJoinPerCenter[ColX, ColZ]{},
		byZ:      map[ColZ]*dynamicJoinPerEdge[ColY]{},
	}
	return dj
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

func (dj *dynamicJoin[ColX, ColY, ColZ]) AddXY(x ColX, y ColY) {
	addABC[ColX, ColY, ColZ](dj.logger, dj.byX, x, dynamicJoinCenterIndexForward[ColX, ColY, ColZ](dj.byY), y, dj.byZ, dj.receiver)
}

func (dj *dynamicJoin[ColX, ColY, ColZ]) AddYZ(y ColY, z ColZ) {
	addABC[ColZ, ColY, ColX](dj.logger, dj.byZ, z, dynamicJoinCenterIndexReverse[ColX, ColY, ColZ](dj.byY), y, dj.byX, PairSetChangeReceiverReverse[ColX, ColZ]{dj.receiver})
}

func addABC[ColA comparable, ColB comparable, ColC comparable](
	logger klog.Logger,
	byA map[ColA]*dynamicJoinPerEdge[ColB],
	a ColA,
	byB dynamicJoinCenterIndex[ColA, ColB, ColC],
	b ColB,
	byC map[ColC]*dynamicJoinPerEdge[ColB],
	receiver PairSetChangeReceiver[ColA, ColC],
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
		receiver.Add(a, c)
	})
}

func (dj *dynamicJoin[ColX, ColY, ColZ]) RemoveXY(x ColX, y ColY) {
	removeABC[ColX, ColY, ColZ](dj.logger, dj.byX, x, dynamicJoinCenterIndexForward[ColX, ColY, ColZ](dj.byY), y, dj.byZ, dj.receiver)
}

func (dj *dynamicJoin[ColX, ColY, ColZ]) RemoveYZ(y ColY, z ColZ) {
	removeABC[ColZ, ColY, ColX](dj.logger, dj.byZ, z, dynamicJoinCenterIndexReverse[ColX, ColY, ColZ](dj.byY), y, dj.byX, PairSetChangeReceiverReverse[ColX, ColZ]{dj.receiver})
}

func removeABC[ColA comparable, ColB comparable, ColC comparable](
	logger klog.Logger,
	byA map[ColA]*dynamicJoinPerEdge[ColB],
	a ColA,
	byB dynamicJoinCenterIndex[ColA, ColB, ColC],
	b ColB,
	byC map[ColC]*dynamicJoinPerEdge[ColB],
	receiver PairSetChangeReceiver[ColA, ColC],
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
			receiver.Remove(a, c)
		})
	}
	if len(aData.ys) == 0 {
		delete(byA, a)
	}
	if bData.IsEmpty() {
		byB.RemoveB(b)
	}
}

type PairSetChangeReceiverReverse[First any, Second any] struct {
	forward PairSetChangeReceiver[First, Second]
}

func (prr PairSetChangeReceiverReverse[First, Second]) Add(second Second, first First) {
	prr.forward.Add(first, second)
}

func (prr PairSetChangeReceiverReverse[First, Second]) Remove(second Second, first First) {
	prr.forward.Remove(first, second)
}

// PairSetChangeReceiver is given a series of changes to a set of pairs
type PairSetChangeReceiver[First any, Second any] interface {
	Add(First, Second)
	Remove(First, Second)
}
