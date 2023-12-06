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
	"math/rand"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	machtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"

	edgeapi "github.com/kubestellar/kubestellar/pkg/apis/edge/v2alpha1"
)

func exerciseFactorer[Whole, PartA, PartB comparable](factor Factorer[Whole, PartA, PartB], exampleWhole Whole, examplePartA PartA, examplePartB PartB) func(*testing.T) {
	rotator := Rotator[Whole, Pair[PartA, PartB]](factor)
	return exerciseRotator(rotator, exampleWhole, Pair[PartA, PartB]{examplePartA, examplePartB})
}

func exerciseFactorerParametric[Whole, PartA, PartB any](
	domainWhole HashDomain[Whole],
	domainA HashDomain[PartA],
	domainB HashDomain[PartB],
	factor Factorer[Whole, PartA, PartB],
	exampleWhole Whole, examplePartA PartA, examplePartB PartB,
) func(*testing.T) {
	rotator := Rotator[Whole, Pair[PartA, PartB]](factor)
	return exerciseRotatorParametric(domainWhole,
		PairHashDomain(domainA, domainB),
		rotator,
		exampleWhole, Pair[PartA, PartB]{examplePartA, examplePartB})
}

func exerciseRotator[Original, Rotated comparable](rotator Rotator[Original, Rotated], exampleOriginal Original, exampleRotated Rotated) func(*testing.T) {
	return func(t *testing.T) {
		exampleOriginalR := rotator.First(exampleOriginal)
		exampleOriginalRR := rotator.Second(exampleOriginalR)
		if exampleOriginal != exampleOriginalRR {
			t.Errorf("Round trip failed: expected %#v, got %#v", exampleOriginal, exampleOriginalRR)
		}
		exampleRotatedR := rotator.Second(exampleRotated)
		exampleRotatedRR := rotator.First(exampleRotatedR)
		if exampleRotated != exampleRotatedRR {
			t.Errorf("Reverse round trip failed: expected %#v, got %#v", exampleRotated, exampleRotatedRR)
		}
	}
}

func exerciseRotatorParametric[Original, Rotated any](hashOriginal HashDomain[Original],
	hashRotated HashDomain[Rotated],
	rotator Rotator[Original, Rotated],
	exampleOriginal Original, exampleRotated Rotated,
) func(*testing.T) {
	return func(t *testing.T) {
		exampleOriginalR := rotator.First(exampleOriginal)
		exampleOriginalRR := rotator.Second(exampleOriginalR)
		if !hashOriginal.Equal(exampleOriginal, exampleOriginalRR) {
			t.Errorf("Round trip failed: expected %#v, got %#v", exampleOriginal, exampleOriginalRR)
		}
		exampleRotatedR := rotator.Second(exampleRotated)
		exampleRotatedRR := rotator.First(exampleRotatedR)
		if !hashRotated.Equal(exampleRotated, exampleRotatedRR) {
			t.Errorf("Reverse round trip failed: expected %#v, got %#v", exampleRotated, exampleRotatedRR)
		}
	}
}

func TestFactorers(t *testing.T) {
	gen := generator{}
	t.Run("factorClusterWhatWhereFullKey", exerciseFactorer(factorClusterWhatWhereFullKey,
		gen.ClusterWhatWhereFullKey(),
		gen.NonNamespacedDistributionTuple(),
		gen.ObjectName()))
	t.Run("factorExternalName", exerciseFactorer(factorExternalName,
		gen.ExternalName(),
		gen.ClusterName(),
		gen.ObjectName()))
	t.Run("factorNamespacedJoinKeyLessNS", exerciseFactorer(
		factorNamespacedJoinKeyLessNS,
		gen.NamespacedJoinKeyLessnS(),
		gen.ProjectionModeKey(),
		gen.ClusterName()))
	t.Run("factorNamespacedWhatWhereFullKey", exerciseFactorer(factorNamespacedWhatWhereFullKey,
		gen.NamespacedWhatWhereFullKey(),
		gen.NamespacedDistributionTuple(),
		gen.ObjectName()))

	t.Run("factorNamespacedDistributionTupleForSync1", exerciseFactorer(factorNamespacedDistributionTupleForSync1,
		gen.NamespacedDistributionTuple(),
		gen.SinglePlacement(),
		NewPair(gen.GroupResourceNamespacedName(), gen.ClusterName())))

	t.Run("factorNamespacedDistributionTupleForProj1", exerciseFactorer(factorNamespacedDistributionTupleForProj1,
		gen.NamespacedDistributionTuple(),
		gen.ClusterName(),
		NewPair(gen.GroupResourceNamespacedName(), gen.SinglePlacement())))
	t.Run("factorNamespacedDistributionTupleForProj1and234", exerciseFactorer(factorNamespacedDistributionTupleForProj1and234,
		gen.NamespacedDistributionTuple(),
		gen.ClusterName(),
		NewTriple(gen.GroupResource(), gen.NamespacedName(), gen.SinglePlacement())))

	t.Run("factorNonNamespacedDistributionTupleForSync1", exerciseFactorer(factorNonNamespacedDistributionTupleForSync1,
		gen.NonNamespacedDistributionTuple(),
		gen.SinglePlacement(),
		NewPair(gen.GroupResourceObjectName(), gen.ClusterName())))
	t.Run("factorNonNamespacedDistributionTupleForProj1", exerciseFactorer(factorNonNamespacedDistributionTupleForProj1,
		gen.NonNamespacedDistributionTuple(),
		gen.ClusterName(),
		NewPair(gen.GroupResourceObjectName(), gen.SinglePlacement())))
	t.Run("factorNonNamespacedDistributionTupleForProj1and234", exerciseFactorer(factorNonNamespacedDistributionTupleForProj1and234,
		gen.NonNamespacedDistributionTuple(),
		gen.ClusterName(),
		NewTriple(gen.GroupResource(), gen.ObjectName(), gen.SinglePlacement())))

	t.Run("factorProjectionModeKeyForSyncer", exerciseFactorer(factorProjectionModeKeyForSyncer,
		gen.ProjectionModeKey(),
		gen.SinglePlacement(),
		gen.GroupResource(),
	))
	t.Run("factorProjectionModeKeyForProj", exerciseFactorer(factorProjectionModeKeyForProj,
		gen.ProjectionModeKey(),
		gen.GroupResource(),
		gen.SinglePlacement(),
	))

	t.Run("factorUpsyncTuple", exerciseFactorerParametric(
		TripleHashDomain[ExternalName, edgeapi.UpsyncSet, SinglePlacement](HashExternalName, HashUpsyncSet{}, HashSinglePlacement{}),
		PairHashDomain[SinglePlacement, edgeapi.UpsyncSet](HashSinglePlacement{}, HashUpsyncSet{}),
		HashExternalName,
		factorUpsyncTuple,
		NewTriple(gen.ExternalName(), gen.UpsyncSet(), gen.SinglePlacement()),
		NewPair(gen.SinglePlacement(), gen.UpsyncSet()),
		gen.ExternalName()))
	t.Run("PairFactorer", exerciseFactorer(PairFactorer[int, string](),
		Pair[int, string]{rand.Intn(100) + 301, gen.String()},
		rand.Intn(100)+3,
		gen.String()))
	t.Run("PairReverser", exerciseRotator(PairReverser[int, string](),
		Pair[int, string]{rand.Intn(100) + 301, gen.String()},
		Pair[string, int]{gen.String(), rand.Intn(100) + 301}))
	t.Run("TripleFactorerTo23and1", exerciseFactorer(
		TripleFactorerTo23and1[int, string, float32](),
		Triple[int, string, float32]{rand.Intn(100) + 200, gen.String(), rand.Float32()},
		Pair[string, float32]{gen.String(), rand.Float32()},
		rand.Intn(100)-100))
	t.Run("TripleFactorTo13and2", exerciseFactorer(
		TripleFactorerTo13and2[int, string, float32](),
		Triple[int, string, float32]{rand.Intn(100) + 200, gen.String(), rand.Float32()},
		Pair[int, float32]{rand.Intn(100) - 200, rand.Float32()},
		gen.String()))
	t.Run("TripleReverser", exerciseRotator(TripleReverser[int, string, float64](),
		Triple[int, string, float64]{rand.Intn(100) + 301, gen.String(), rand.Float64()},
		Triple[float64, string, int]{rand.Float64(), gen.String(), rand.Intn(100) + 301}))
}

type generator struct{}

func (generator) String() string {
	var builder strings.Builder
	size := rand.Intn(20)
	for index := 0; index < size; index++ {
		builder.WriteRune(rune('a' + rand.Intn(26)))
	}
	return builder.String()
}

func (gen generator) ClusterName() string {
	return gen.String()
}

func (gen generator) NamespaceName() NamespaceName {
	return NamespaceName(gen.String())
}

func (gen generator) ObjectName() ObjectName {
	return ObjectName(gen.String())
}

func (gen generator) ExternalName() ExternalName {
	return ExternalName{gen.ClusterName(), gen.ObjectName()}
}

func (gen generator) ExternalNamespacedName() ExternalNamespacedName {
	return NewTriple(gen.ClusterName(), gen.NamespaceName(), gen.ObjectName())
}

func (gen generator) NamespacedName() NamespacedName {
	return NewPair(gen.NamespaceName(), gen.ObjectName())
}

func (gen generator) GroupResource() metav1.GroupResource {
	return metav1.GroupResource{Group: "g" + gen.String(), Resource: gen.String() + "s"}
}

func (gen generator) GroupResourceNamespacedName() GroupResourceNamespacedName {
	return NewPair(gen.GroupResource(), gen.NamespacedName())
}

func (gen generator) GroupResourceObjectName() GroupResourceObjectName {
	return NewPair(gen.GroupResource(), gen.ObjectName())
}

func (generator) UID() machtypes.UID {
	return uuid.NewUUID()
}

func (gen generator) SinglePlacement() SinglePlacement {
	return SinglePlacement{
		Cluster:        "clu-" + gen.String(),
		LocationName:   "loc-" + gen.String(),
		SyncTargetName: "st-" + gen.String(),
		SyncTargetUID:  gen.UID(),
	}
}

func (gen generator) NamespaceAndDestination() NamespaceAndDestination {
	return NamespaceAndDestination{
		First:  gen.NamespaceName(),
		Second: gen.SinglePlacement()}
}

func (gen generator) WorkloadPartID() WorkloadPartID {
	return NewTriple(
		gen.GroupResource(),
		gen.NamespaceName(),
		gen.ObjectName(),
	)
}
func (gen generator) NamespacedWhatWhereFullKey() NamespacedWhatWhereFullKey {
	return NamespacedWhatWhereFullKey{
		First:  gen.ExternalName(),
		Second: gen.WorkloadPartID(),
		Third:  gen.SinglePlacement()}
}

func (gen generator) ClusterWhatWhereFullKey() ClusterWhatWhereFullKey {
	return ClusterWhatWhereFullKey{
		First:  gen.ExternalName(),
		Second: Pair[metav1.GroupResource, ObjectName]{gen.GroupResource(), gen.ObjectName()},
		Third:  gen.SinglePlacement()}
}

func (gen generator) NonNamespacedDistributionTuple() NonNamespacedDistributionTuple {
	return NonNamespacedDistributionTuple{
		First:  gen.ProjectionModeKey(),
		Second: gen.ExternalName()}
}

func (gen generator) NamespacedDistributionTuple() NamespacedDistributionTuple {
	return NewPair(gen.ProjectionModeKey(), gen.ExternalNamespacedName())
}

func (gen generator) NamespacedJoinKeyLessnS() NamespacedJoinKeyLessnS {
	return NamespacedJoinKeyLessnS{
		First:  gen.ClusterName(),
		Second: gen.GroupResource(),
		Third:  gen.SinglePlacement()}
}

func (gen generator) ProjectionModeKey() ProjectionModeKey {
	return ProjectionModeKey{
		GroupResource: gen.GroupResource(),
		Destination:   gen.SinglePlacement()}
}

func (gen generator) UpsyncSet() edgeapi.UpsyncSet {
	return edgeapi.UpsyncSet{
		APIGroup:   gen.String() + "." + gen.String(),
		Resources:  []string{gen.String() + "s", gen.String() + "s"},
		Namespaces: []string{"ns" + gen.String(), "ns" + gen.String()},
		Names:      []string{"n" + gen.String(), "n" + gen.String()}}
}
