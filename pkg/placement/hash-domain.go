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
	"hash/crc64"

	"github.com/kcp-dev/logicalcluster/v3"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v1alpha1"
)

type HashDomain[Elt any] interface {
	Equal(Elt, Elt) bool
	Hash(Elt) HashValue
}

type HashValue = uint64

func NewHashDomainFuncs[Elt any](equal func(Elt, Elt) bool, hash func(Elt) HashValue) HashDomainFuncs[Elt] {
	return HashDomainFuncs[Elt]{equal, hash}
}

type HashDomainFuncs[Elt any] struct {
	DoEqual func(Elt, Elt) bool
	DoHash  func(Elt) HashValue
}

func (hdf HashDomainFuncs[Elt]) Equal(left, right Elt) bool { return hdf.DoEqual(left, right) }
func (hdf HashDomainFuncs[Elt]) Hash(elt Elt) HashValue     { return hdf.DoHash(elt) }

func PairHashDomain[First, Second any](first HashDomain[First], second HashDomain[Second]) HashDomain[Pair[First, Second]] {
	return HashDomainFuncs[Pair[First, Second]]{
		DoEqual: func(left, right Pair[First, Second]) bool {
			return first.Equal(left.First, right.First) && second.Equal(left.Second, right.Second)
		},
		DoHash: func(tup Pair[First, Second]) HashValue { return first.Hash(tup.First)*257 + second.Hash(tup.Second) },
	}
}

func TripleHashDomain[First, Second, Third any](first HashDomain[First], second HashDomain[Second], third HashDomain[Third]) HashDomain[Triple[First, Second, Third]] {
	return HashDomainFuncs[Triple[First, Second, Third]]{
		DoEqual: func(left, right Triple[First, Second, Third]) bool {
			return first.Equal(left.First, right.First) && second.Equal(left.Second, right.Second) && third.Equal(left.Third, right.Third)
		},
		DoHash: func(tup Triple[First, Second, Third]) HashValue {
			return first.Hash(tup.First)*65539 + second.Hash(tup.Second)*257 + third.Hash(tup.Third)
		},
	}
}

func NewTransformHashDomain[Original, Transformed any](tranform func(Original) Transformed, transformedDomain HashDomain[Transformed]) HashDomain[Original] {
	return TransformHashDomain[Original, Transformed]{tranform, transformedDomain}
}

type TransformHashDomain[Original, Transformed any] struct {
	Transform         func(Original) Transformed
	TransformedDomain HashDomain[Transformed]
}

func (thd TransformHashDomain[Original, Transformed]) Equal(left, right Original) bool {
	leftT := thd.Transform(left)
	rightT := thd.Transform(right)
	return thd.TransformedDomain.Equal(leftT, rightT)
}

func (thd TransformHashDomain[Original, Transformed]) Hash(arg Original) HashValue {
	argT := thd.Transform(arg)
	return thd.TransformedDomain.Hash(argT)
}

func NewSliceHashDomain[Elt any](eltDomain HashDomain[Elt]) HashDomain[[]Elt] {
	return SliceHashDomain[Elt]{eltDomain}
}

type SliceHashDomain[Elt any] struct{ EltDomain HashDomain[Elt] }

func (shd SliceHashDomain[Elt]) Equal(left, right []Elt) bool {
	if len(left) != len(right) {
		return false
	}
	for index, leftElt := range left {
		if !shd.EltDomain.Equal(leftElt, right[index]) {
			return false
		}
	}
	return true
}

func (shd SliceHashDomain[Elt]) Hash(slice []Elt) HashValue {
	var ans HashValue
	for _, elt := range slice {
		// TODO later: much better.  Or import something.
		ans = ans*263 + shd.EltDomain.Hash(elt)
	}
	return ans
}

type HashDomainString struct{}

var _ HashDomain[string] = HashDomainString{}

func (HashDomainString) Equal(left, right string) bool { return left == right }
func (HashDomainString) Hash(arg string) HashValue     { return StringHash(arg) }

var ecmaTable = crc64.MakeTable(crc64.ECMA)

func StringHash(arg string) HashValue { return crc64.Checksum([]byte(arg), ecmaTable) }

var SliceOfStringDomain = NewSliceHashDomain[string](HashDomainString{})

var HashLogicalClusterName = NewTransformHashDomain[logicalcluster.Name, string](func(name logicalcluster.Name) string { return string(name) }, HashDomainString{})

var HashClusterString = PairHashDomain[logicalcluster.Name, string](HashLogicalClusterName, HashDomainString{})

var factorExternalName = NewFactorer(
	func(whole ExternalName) Pair[logicalcluster.Name, string] { return NewPair(whole.Cluster, whole.Name) },
	func(parts Pair[logicalcluster.Name, string]) ExternalName {
		return ExternalName{parts.First, parts.Second}
	})

var HashExternalName = NewTransformHashDomain(factorExternalName.First, HashClusterString)

type HashUpsyncSet struct{}

var _ HashDomain[edgeapi.UpsyncSet] = HashUpsyncSet{}

func (HashUpsyncSet) Equal(left, right edgeapi.UpsyncSet) bool {
	return UpsyncSetEqual(left, right)
}

func (HashUpsyncSet) Hash(arg edgeapi.UpsyncSet) HashValue {
	return StringHash(arg.APIGroup) + 5*SliceOfStringDomain.Hash(arg.Resources) + 37*SliceOfStringDomain.Hash(arg.Namespaces) + 257*SliceOfStringDomain.Hash(arg.Names)
}

func UpsyncSetEqual(left, right edgeapi.UpsyncSet) bool {
	if left.APIGroup != right.APIGroup {
		return false
	}
	return SliceEqual(left.Resources, right.Resources) && SliceEqual(left.Namespaces, right.Namespaces) && SliceEqual(left.Names, right.Names)
}

type HashSinglePlacement struct{}

var _ HashDomain[SinglePlacement] = HashSinglePlacement{}

func (HashSinglePlacement) Equal(left, right SinglePlacement) bool {
	return left.Cluster == right.Cluster &&
		left.LocationName == right.LocationName &&
		left.SyncTargetName == right.SyncTargetName &&
		left.SyncTargetUID == right.SyncTargetUID
}

func (HashSinglePlacement) Hash(arg SinglePlacement) HashValue {
	return StringHash(arg.Cluster) + StringHash(arg.LocationName) + StringHash(arg.SyncTargetName) + StringHash(string(arg.SyncTargetUID))
}
