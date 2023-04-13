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

// NewDynamicFullJoin12VWith13 constructs a data structure that incrementally maintains an equijoin
// of (1) a map from pairs to values with (2) a pair set.
// The join is on the pair first members.
// Given a receiver of changes to the result of the equijoin,
// this function returns receivers of changes to the two ipnputs.
// In other words, this joins the change streams of two tables to produce the change stream
// of the join of those two tables --- in a passive stance (i.e., is in terms of the stream receivers).
func NewDynamicFullJoin12VWith13[ColX, ColY, ColZ comparable, Val any](logger klog.Logger, receiver MappingReceiver[Triple[ColX, ColY, ColZ], Val]) (MappingReceiver[Pair[ColX, ColY], Val], SetChangeReceiver[Pair[ColX, ColZ]]) {
	dj := &dynamicJoin12VWith13[ColX, ColY, ColZ, Val]{
		logger:   logger,
		receiver: receiver,
		xzReln:   NewMapRelation2[ColX, ColZ](),
	}
	dj.xyvReln = NewFactoredMapMap[Pair[ColX, ColY], ColX, ColY, Val](PairFactorer[ColX, ColY](), extrapolate2v[ColX, ColY, ColZ, Val]{dj.xzReln, receiver}, nil, nil)
	dj.xzReln = Relation2WithObservers[ColX, ColZ](dj.xzReln, extrapolate1v[ColX, ColY, ColZ, Val]{dj.xyvReln, receiver})
	return dj.xyvReln, dj.xzReln
}

// dynamicJoin12VWith13 implements NewDynamicFullJoin12VWith13.
// It buffers the two incoming tables and passes on changes.
type dynamicJoin12VWith13[ColX, ColY, ColZ comparable, Val any] struct {
	logger   klog.Logger
	receiver MappingReceiver[Triple[ColX, ColY, ColZ], Val]
	xyvReln  FactoredMap[Pair[ColX, ColY], ColX, ColY, Val]
	xzReln   MutableRelation2[ColX, ColZ]
}

type extrapolate1v[ColX, ColY, ColZ comparable, Val any] struct {
	xyvReln  FactoredMap[Pair[ColX, ColY], ColX, ColY, Val]
	receiver MappingReceiver[Triple[ColX, ColY, ColZ], Val]
}

func (er extrapolate1v[ColX, ColY, ColZ, Val]) Add(xz Pair[ColX, ColZ]) bool {
	er.xyvReln.GetIndex().Visit1to2(xz.First, func(inner Pair[ColY, Val]) error {
		er.receiver.Put(Triple[ColX, ColY, ColZ]{xz.First, inner.First, xz.Second}, inner.Second)
		return nil
	})
	return true
}

func (er extrapolate1v[ColX, ColY, ColZ, Val]) Remove(xz Pair[ColX, ColZ]) bool {
	er.xyvReln.GetIndex().Visit1to2(xz.First, func(inner Pair[ColY, Val]) error {
		er.receiver.Delete(Triple[ColX, ColY, ColZ]{xz.First, inner.First, xz.Second})
		return nil
	})
	return true
}

type extrapolate2v[ColX, ColY, ColZ comparable, Val any] struct {
	xzReln   Relation2[ColX, ColZ]
	receiver MappingReceiver[Triple[ColX, ColY, ColZ], Val]
}

func (er extrapolate2v[ColX, ColY, ColZ, Val]) Create(xy Pair[ColX, ColY], val Val) {
	er.xzReln.GetIndex1to2().Visit1to2(xy.First, func(z ColZ) error {
		er.receiver.Put(Triple[ColX, ColY, ColZ]{xy.First, xy.Second, z}, val)
		return nil
	})
}
func (er extrapolate2v[ColX, ColY, ColZ, Val]) Update(xy Pair[ColX, ColY], oldVal, newVal Val) {
	er.Create(xy, newVal) // this thing rubs off the difference
}

func (er extrapolate2v[ColX, ColY, ColZ, Val]) DeleteWithFinal(xy Pair[ColX, ColY], val Val) {
	er.xzReln.GetIndex1to2().Visit1to2(xy.First, func(z ColZ) error {
		er.receiver.Delete(Triple[ColX, ColY, ColZ]{xy.First, xy.Second, z})
		return nil
	})
}
