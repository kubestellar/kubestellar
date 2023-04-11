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

// NewDynamicJoin12with13 constructs a data structure that incrementally maintains an equijoin.
// The join is on the first column of the input tables.
// Given a receiver of changes to the result of the equijoin between two two-column tables,
// this function returns receivers of changes to the two tables.
// In other words, this joins the change streams of two tables to produce the change stream
// of the join of those two tables --- in a passive stance (i.e., is in terms of the stream receivers).
// Note: the uniformity of the input and output types means that this can be chained.
func NewDynamicJoin12with13[ColX, ColY, ColZ comparable](logger klog.Logger, receiver SetWriter[Pair[ColY, ColZ]]) (SetWriter[Pair[ColX, ColY]], SetWriter[Pair[ColX, ColZ]]) {
	indexerAsReceiver := NewSetChangeProjectorByMapMap[Triple[ColX, ColY, ColZ], Pair[ColY, ColZ], ColX](
		TripleFactorerTo23and1[ColX, ColY, ColZ](),
		receiver,
	)
	return NewDynamicFullJoin12with13(logger, indexerAsReceiver)
}

// NewDynamicFullJoin12with13 is like NewDynamicJoin12with13 but passes along the set of joint values too.
func NewDynamicFullJoin12with13[ColX, ColY, ColZ comparable](logger klog.Logger, receiver SetWriter[Triple[ColX, ColY, ColZ]]) (SetWriter[Pair[ColX, ColY]], SetWriter[Pair[ColX, ColZ]]) {
	dj := &dynamicJoin12with13[ColX, ColY, ColZ]{
		logger:   logger,
		receiver: receiver,
		xyReln:   NewMapRelation2[ColX, ColY](),
		xzReln:   NewMapRelation2[ColX, ColZ](),
	}
	dj.xyReln = Relation2WithObservers[ColX, ColY](dj.xyReln, extrapolateFwd1[ColX, ColY, ColZ]{dj.xzReln, receiver})
	dj.xzReln = Relation2WithObservers[ColX, ColZ](dj.xzReln, extrapolateFwd1[ColX, ColZ, ColY]{dj.xyReln, TripleSetWriterReverse23[ColX, ColY, ColZ]{receiver}})
	return dj.xyReln, dj.xzReln
}

// NewDynamicFullJoin12with13Parametric is like NewDynamicFullJoin12with13 but is based on general hash maps
func NewDynamicFullJoin12with13Parametric[ColX, ColY, ColZ any](logger klog.Logger, hashDomainX HashDomain[ColX], hashDomainY HashDomain[ColY], hashDomainZ HashDomain[ColZ], receiver SetWriter[Triple[ColX, ColY, ColZ]]) (SetWriter[Pair[ColX, ColY]], SetWriter[Pair[ColX, ColZ]]) {
	dj := &dynamicJoin12with13[ColX, ColY, ColZ]{
		logger:   logger,
		receiver: receiver,
		xyReln:   NewHashRelation2[ColX, ColY](hashDomainX, hashDomainY),
		xzReln:   NewHashRelation2[ColX, ColZ](hashDomainX, hashDomainZ),
	}
	dj.xyReln = Relation2WithObservers[ColX, ColY](dj.xyReln, extrapolateFwd1[ColX, ColY, ColZ]{dj.xzReln, receiver})
	dj.xzReln = Relation2WithObservers[ColX, ColZ](dj.xzReln, extrapolateFwd1[ColX, ColZ, ColY]{dj.xyReln, TripleSetWriterReverse23[ColX, ColY, ColZ]{receiver}})
	return dj.xyReln, dj.xzReln
}

// dynamicJoin implements DynamicJoin.
// It buffers the two incoming relations and passes on changes.
type dynamicJoin12with13[ColX, ColY, ColZ any] struct {
	logger   klog.Logger
	receiver SetWriter[Triple[ColX, ColY, ColZ]]
	xyReln   MutableRelation2[ColX, ColY]
	xzReln   MutableRelation2[ColX, ColZ]
}

type extrapolateFwd1[ColX, ColY, ColZ any] struct {
	xzReln   Relation2[ColX, ColZ]
	receiver SetWriter[Triple[ColX, ColY, ColZ]]
}

func (er extrapolateFwd1[ColX, ColY, ColZ]) Add(xy Pair[ColX, ColY]) bool {
	er.xzReln.GetIndex1to2().Visit1to2(xy.First, func(z ColZ) error {
		er.receiver.Add(Triple[ColX, ColY, ColZ]{xy.First, xy.Second, z})
		return nil
	})
	return true
}

func (er extrapolateFwd1[ColX, ColY, ColZ]) Remove(xy Pair[ColX, ColY]) bool {
	er.xzReln.GetIndex1to2().Visit1to2(xy.First, func(z ColZ) error {
		er.receiver.Remove(Triple[ColX, ColY, ColZ]{xy.First, xy.Second, z})
		return nil
	})
	return true
}

type TripleSetWriterReverse23[Left, Middle, Right any] struct {
	forward SetWriter[Triple[Left, Middle, Right]]
}

func (prr TripleSetWriterReverse23[Left, Middle, Right]) Add(tup Triple[Left, Right, Middle]) bool {
	return prr.forward.Add(Triple[Left, Middle, Right]{tup.First, tup.Third, tup.Second})
}

func (prr TripleSetWriterReverse23[Left, Middle, Right]) Remove(tup Triple[Left, Right, Middle]) bool {
	return prr.forward.Remove(Triple[Left, Middle, Right]{tup.First, tup.Third, tup.Second})
}
