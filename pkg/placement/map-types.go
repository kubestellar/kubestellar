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
	"sync"

	"k8s.io/klog/v2"
)

// Map is a finite set of (key,value) pairs
// that has at most one value for any given key.
// The collection may or may not be mutable.
// This view of the collection may or may not have a limited scope of validity.
// This view may or may not have concurrency restrictions.
type Map[Key, Val any] interface {
	Emptyable
	Len() int
	LenIsCheap() bool
	Get(Key) (Val, bool)
	Visitable[Pair[Key, Val]]
}

// MutableMap is a Map that can be written to.
type MutableMap[Key, Val any] interface {
	Map[Key, Val]
	MappingReceiver[Key, Val]
}

// MappingReceiver is something that can be given key/value pairs.
// This is the writable aspect of a Map.
// Some DynamicMapProvider implementations require receivers to be comparable.
type MappingReceiver[Key, Val any] interface {
	Put(Key, Val)
	Delete(Key)
}

func NewMappingReceiverFuncs[Key, Val any](put func(Key, Val), delete func(Key)) MappingReceiver[Key, Val] {
	return MappingReceiverFuncs[Key, Val]{put, delete}
}

// MappingReceiverFuncs is a convenient constructor of MappingReceiver from two funcs
type MappingReceiverFuncs[Key, Val any] struct {
	OnPut    func(Key, Val)
	OnDelete func(Key)
}

var _ MappingReceiver[float32, map[string]func()] = MappingReceiverFuncs[float32, map[string]func()]{}
var _ MapChangeReceiver[float32, map[string]func()] = MappingReceiverFuncs[float32, map[string]func()]{}

func (mrf MappingReceiverFuncs[Key, Val]) Put(key Key, val Val) {
	if mrf.OnPut != nil {
		mrf.OnPut(key, val)
	}
}

func (mrf MappingReceiverFuncs[Key, Val]) Delete(key Key) {
	if mrf.OnDelete != nil {
		mrf.OnDelete(key)
	}
}

func (mrf MappingReceiverFuncs[Key, Val]) Create(key Key, val Val) {
	if mrf.OnPut != nil {
		mrf.OnPut(key, val)
	}
}

func (mrf MappingReceiverFuncs[Key, Val]) Update(key Key, oldVal, newVal Val) {
	if mrf.OnPut != nil {
		mrf.OnPut(key, newVal)
	}
}

func (mrf MappingReceiverFuncs[Key, Val]) DeleteWithFinal(key Key, oldVal Val) {
	if mrf.OnDelete != nil {
		mrf.OnDelete(key)
	}
}

type MappingReceiverFork[Key, Val any] []MappingReceiver[Key, Val]

var _ MappingReceiver[int, func()] = MappingReceiverFork[int, func()]{}

func (mrf MappingReceiverFork[Key, Val]) Put(key Key, val Val) {
	for _, mr := range mrf {
		mr.Put(key, val)
	}
}

func (mrf MappingReceiverFork[Key, Val]) Delete(key Key) {
	for _, mr := range mrf {
		mr.Delete(key)
	}
}

type MappingReceiverHolderFork[Key, Val any] []*MappingReceiverHolder[Key, Val]

var _ MappingReceiver[int, func()] = MappingReceiverHolderFork[int, func()]{}

func (mrf MappingReceiverHolderFork[Key, Val]) Put(key Key, val Val) {
	for _, mr := range mrf {
		mr.Put(key, val)
	}
}

func (mrf MappingReceiverHolderFork[Key, Val]) Delete(key Key) {
	for _, mr := range mrf {
		mr.Delete(key)
	}
}

// MappingReceiverFunc produces a MappingReceiver that defers to another MappingReceiver computed on each use
func MappingReceiverFunc[Key, Val any](fn func() MappingReceiver[Key, Val]) MappingReceiver[Key, Val] {
	return mappingReceiverFunc[Key, Val]{fn}
}

type mappingReceiverFunc[Key, Val any] struct {
	fn func() MappingReceiver[Key, Val]
}

func (mrf mappingReceiverFunc[Key, Val]) Delete(key Key) {
	mr := mrf.fn()
	mr.Delete(key)
}

func (mrf mappingReceiverFunc[Key, Val]) Put(key Key, val Val) {
	mr := mrf.fn()
	mr.Put(key, val)
}

// MapChangeReceiver is what a stateful map offers to an observer
type MapChangeReceiver[Key, Val any] interface {
	Create(Key, Val)

	// Update is given key, old value, new value
	Update(Key, Val, Val)

	// DeleteWithFinal is given key and last value
	DeleteWithFinal(Key, Val)
}

// MapChangeReceiverFuncs is a convenient constructor of MapChangeReceiver from three funcs
type MapChangeReceiverFuncs[Key, Val any] struct {
	OnCreate func(Key, Val)
	OnUpdate func(Key, Val, Val)
	OnDelete func(Key, Val)
}

var _ MapChangeReceiver[string, func()] = MapChangeReceiverFuncs[string, func()]{}

func (mrf MapChangeReceiverFuncs[Key, Val]) Create(key Key, val Val) {
	if mrf.OnCreate != nil {
		mrf.OnCreate(key, val)
	}
}

func (mrf MapChangeReceiverFuncs[Key, Val]) Update(key Key, oldVal, newVal Val) {
	if mrf.OnUpdate != nil {
		mrf.OnUpdate(key, oldVal, newVal)
	}
}

func (mrf MapChangeReceiverFuncs[Key, Val]) DeleteWithFinal(key Key, val Val) {
	if mrf.OnDelete != nil {
		mrf.OnDelete(key, val)
	}
}

type MapChangeReceiverFork[Key, Val any] []MapChangeReceiver[Key, Val]

var _ MapChangeReceiver[int, func()] = MapChangeReceiverFork[int, func()]{}

func (mrf MapChangeReceiverFork[Key, Val]) Create(key Key, val Val) {
	for _, mr := range mrf {
		mr.Create(key, val)
	}
}

func (mrf MapChangeReceiverFork[Key, Val]) Update(key Key, oldVal, newVal Val) {
	for _, mr := range mrf {
		mr.Update(key, oldVal, newVal)
	}
}

func (mrf MapChangeReceiverFork[Key, Val]) DeleteWithFinal(key Key, val Val) {
	for _, mr := range mrf {
		mr.DeleteWithFinal(key, val)
	}
}

// MappingReceiverDiscardsPrevious produces a MapChangeReceiver that dumbs down its info to pass along to the given MappingReceiver
func MappingReceiverDiscardsPrevious[Key, Val any](mr MappingReceiver[Key, Val]) MapChangeReceiver[Key, Val] {
	return mappingReceiverDiscardsPrevious[Key, Val]{inner: mr}
}

type mappingReceiverDiscardsPrevious[Key, Val any] struct{ inner MappingReceiver[Key, Val] }

func (mr mappingReceiverDiscardsPrevious[Key, Val]) Create(key Key, val Val) {
	mr.inner.Put(key, val)
}

func (mr mappingReceiverDiscardsPrevious[Key, Val]) Update(key Key, oldVal, newVal Val) {
	mr.inner.Put(key, newVal)
}

func (mr mappingReceiverDiscardsPrevious[Key, Val]) DeleteWithFinal(key Key, val Val) {
	mr.inner.Delete(key)
}

// TransactionalMappingReceiver is one that takes updates in batches
type TransactionalMappingReceiver[Key, Val any] interface {
	Transact(func(MappingReceiver[Key, Val]))
}

// MapReadonly returns a version of the argument that does not support writes
func MapReadonly[Key, Val any](inner Map[Key, Val]) Map[Key, Val] {
	return mapReadonly[Key, Val]{inner}
}

type mapReadonly[Key, Val any] struct {
	Map[Key, Val]
}

func MutableMapWithKeyObserver[Key, Val any](mm MutableMap[Key, Val], observer SetWriter[Key]) MutableMap[Key, Val] {
	return &mutableMapWithKeyObserver[Key, Val]{mm, observer}
}

type mutableMapWithKeyObserver[Key, Val any] struct {
	MutableMap[Key, Val]
	observer SetWriter[Key]
}

func (mko *mutableMapWithKeyObserver[Key, Val]) Put(key Key, val Val) {
	mko.MutableMap.Put(key, val)
	mko.observer.Add(key)
}

func (mko *mutableMapWithKeyObserver[Key, Val]) Delete(key Key) {
	mko.MutableMap.Delete(key)
	mko.observer.Remove(key)
}

type TransformMappingReceiver[KeyOriginal, KeyTransformed, ValOriginal, ValTransformed any] struct {
	TransformKey func(KeyOriginal) KeyTransformed
	TransformVal func(ValOriginal) ValTransformed
	Inner        MappingReceiver[KeyTransformed, ValTransformed]
}

var _ MappingReceiver[int, func()] = &TransformMappingReceiver[int, string, func(), []int]{}

func (xr TransformMappingReceiver[KeyOriginal, KeyTransformed, ValOriginal, ValTransformed]) Put(keyOriginal KeyOriginal, valOriginal ValOriginal) {
	keyTransformed := xr.TransformKey(keyOriginal)
	valTransformed := xr.TransformVal(valOriginal)
	xr.Inner.Put(keyTransformed, valTransformed)
}

func (xr TransformMappingReceiver[KeyOriginal, KeyTransformed, ValOriginal, ValTransformed]) Delete(keyOriginal KeyOriginal) {
	keyTransformed := xr.TransformKey(keyOriginal)
	xr.Inner.Delete(keyTransformed)
}

func WrapMapWithMutex[Key comparable, Val any](theMap MutableMap[Key, Val]) MutableMap[Key, Val] {
	return &mapMutex[Key, Val]{theMap: theMap}
}

type mapMutex[Key comparable, Val any] struct {
	sync.RWMutex
	theMap MutableMap[Key, Val]
}

func (mm *mapMutex[Key, Val]) IsEmpty() bool {
	mm.RLock()
	defer mm.RUnlock()
	return mm.theMap.IsEmpty()
}

func (mm *mapMutex[Key, Val]) LenIsCheap() bool {
	return mm.theMap.LenIsCheap()
}

func (mm *mapMutex[Key, Val]) Len() int {
	mm.RLock()
	defer mm.RUnlock()
	return mm.theMap.Len()
}

func (mm *mapMutex[Key, Val]) Delete(key Key) {
	mm.Lock()
	defer mm.Unlock()
	mm.theMap.Delete(key)
}

func (mm *mapMutex[Key, Val]) Put(key Key, val Val) {
	mm.Lock()
	defer mm.Unlock()
	mm.theMap.Put(key, val)
}

func (mm *mapMutex[Key, Val]) Get(key Key) (Val, bool) {
	mm.RLock()
	defer mm.RUnlock()
	return mm.theMap.Get(key)
}

func (mm *mapMutex[Key, Val]) Visit(visitor func(Pair[Key, Val]) error) error {
	mm.RLock()
	defer mm.RUnlock()
	return mm.theMap.Visit(visitor)
}

func NewLoggingMappingReceiver[Key comparable, Val any](mapName string, logger klog.Logger) MappingReceiver[Key, Val] {
	return loggingMappingReceiver[Key, Val]{mapName, logger}
}

type loggingMappingReceiver[Key, Val any] struct {
	mapName string
	logger  klog.Logger
}

var _ MappingReceiver[string, []any] = loggingMappingReceiver[string, []any]{}

func (lmr loggingMappingReceiver[Key, Val]) Put(key Key, val Val) {
	lmr.logger.Info("Put", "map", lmr.mapName, "key", key, "val", val)
}

func (lmr loggingMappingReceiver[Key, Val]) Delete(key Key) {
	lmr.logger.Info("Delete", "map", lmr.mapName, "key", key)
}

func MappingReceiverAsVisitor[Key, Val any](receiver MappingReceiver[Key, Val]) func(Pair[Key, Val]) error {
	return func(tup Pair[Key, Val]) error {
		receiver.Put(tup.First, tup.Second)
		return nil
	}
}

func MappingReceiverNegativeAsVisitor[Key, Val any](receiver MappingReceiver[Key, Val]) func(Pair[Key, Val]) error {
	return func(tup Pair[Key, Val]) error {
		receiver.Delete(tup.First)
		return nil
	}
}

func MapApply[Key, Val any](theMap Map[Key, Val], receiver MappingReceiver[Key, Val]) {
	theMap.Visit(MappingReceiverAsVisitor(receiver))
}

func MapAddAll[Key, Val any](theMap MutableMap[Key, Val], adds Visitable[Pair[Key, Val]]) {
	adds.Visit(func(add Pair[Key, Val]) error {
		theMap.Put(add.First, add.Second)
		return nil
	})
}

func MapRemoveAll[Key, Val any](theMap MutableMap[Key, Val], goners Visitable[Pair[Key, Val]]) {
	goners.Visit(func(goner Pair[Key, Val]) error {
		theMap.Delete(goner.First)
		return nil
	})
}

// MapGetAdd does a Get and an add if the sought mmapping is missing and desired.
// If the sought mapping is missing and undesired then the result is the zero value of Val.
func MapGetAdd[Key, Val any](theMap MutableMap[Key, Val], key Key, want bool, valGenerator func(Key) Val) Val {
	val, have := theMap.Get(key)
	if have {
		return val
	}
	if want {
		val = valGenerator(key)
		theMap.Put(key, val)
		return val
	}
	var zero Val
	return zero
}

func MapEqual[Key, Val comparable](left, right Map[Key, Val]) bool {
	return MapEqualParametric[Key, Val](func(a, b Val) bool { return a == b })(left, right)
}

func MapEqualParametric[Key comparable, Val any](isEqual func(Val, Val) bool) func(map1, map2 Map[Key, Val]) bool {
	return func(map1, map2 Map[Key, Val]) bool {
		if map1.Len() != map2.Len() {
			return false
		}
		return map1.Visit(func(tup1 Pair[Key, Val]) error {
			val2, have := map2.Get(tup1.First)
			if !have || !isEqual(tup1.Second, val2) {
				return errStop
			}
			return nil
		}) == nil
	}
}

func MapEnumerateDifferences[Key, Val comparable](left, right Map[Key, Val], receiver MapChangeReceiver[Key, Val]) {
	MapEnumerateDifferencesParametric(func(a, b Val) bool { return a == b }, left, right, receiver)
}

func MapEnumerateDifferencesParametric[Key, Val any](isEqual func(Val, Val) bool, left, right Map[Key, Val], receiver MapChangeReceiver[Key, Val]) {
	left.Visit(func(tup Pair[Key, Val]) error {
		valRight, has := right.Get(tup.First)
		if !has {
			receiver.DeleteWithFinal(tup.First, tup.Second)
		} else if !isEqual(valRight, tup.Second) {
			receiver.Update(tup.First, tup.Second, valRight)
		}
		return nil
	})
	right.Visit(func(tup Pair[Key, Val]) error {
		_, has := left.Get(tup.First)
		if !has {
			receiver.Create(tup.First, tup.Second)
		}
		return nil
	})
}
