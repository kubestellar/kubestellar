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

import (
	"container/list"
	"sync"
	"time"
)

// NewMapRecent returns a Map that limits its memory usage by forgetting old entries.
// An entry is retained at least as long as its age does not exceed the given threshold
// or the map's length does not exceed the given threshold.
func NewMapRecent[Key comparable, Value any](ageAllowed time.Duration, lenAllowed int, getNow func() time.Time) MutableMap[Key, Value] {
	return &mapRecent[Key, Value]{
		ageAllowed: ageAllowed,
		lenAllowed: lenAllowed,
		getNow:     getNow,
		asMap:      map[Key]recentElement[Value]{},
		asList:     *list.New(),
	}
}

type mapRecent[Key comparable, Value any] struct {
	ageAllowed time.Duration
	lenAllowed int
	getNow     func() time.Time
	sync.Mutex
	asMap  map[Key]recentElement[Value]
	asList list.List // of recentEntry[Value]
}

type recentElement[Value any] struct {
	val     Value
	element *list.Element
}

type recentEntry[Key comparable] struct {
	key  Key
	when time.Time
}

func (mr *mapRecent[Key, Value]) Delete(key Key) {
	mr.Lock()
	defer mr.Unlock()
	relt, has := mr.asMap[key]
	if !has {
		return
	}
	delete(mr.asMap, key)
	mr.asList.Remove(relt.element)
}

func (mr *mapRecent[Key, Value]) Get(key Key) (Value, bool) {
	mr.Lock()
	defer mr.Unlock()
	relt, has := mr.asMap[key]
	if !has {
		var zero Value
		return zero, false
	}
	return relt.val, true
}

func (mr *mapRecent[Key, Value]) IsEmpty() bool {
	mr.Lock()
	defer mr.Unlock()
	return len(mr.asMap) == 0
}

func (mr *mapRecent[Key, Value]) Len() int {
	mr.Lock()
	defer mr.Unlock()
	return len(mr.asMap)
}

func (mr *mapRecent[Key, Value]) LenIsCheap() bool {
	return true
}

func (mr *mapRecent[Key, Value]) Put(key Key, val Value) {
	mr.Lock()
	defer mr.Unlock()
	relt, has := mr.asMap[key]
	if has {
		mr.asList.Remove(relt.element)
	}
	rent := &recentEntry[Key]{key: key, when: mr.getNow()}
	relt = recentElement[Value]{val: val, element: mr.asList.PushBack(rent)}
	mr.asMap[key] = relt
	mr.trimLocked()
}

func (mr *mapRecent[Key, Value]) trimLocked() {
	now := mr.getNow()
	for mr.asList.Len() > mr.lenAllowed {
		oldest := mr.asList.Front()
		rent := oldest.Value.(*recentEntry[Key])
		if now.Sub(rent.when) <= mr.ageAllowed {
			return
		}
		mr.asList.Remove(oldest)
		delete(mr.asMap, rent.key)
	}
}

func (mr *mapRecent[Key, Value]) Visit(visitor func(Pair[Key, Value]) error) error {
	mr.Lock()
	defer mr.Unlock()
	for key, relt := range mr.asMap {
		err := visitor(NewPair(key, relt.val))
		if err != nil {
			return err
		}
	}
	return nil
}
