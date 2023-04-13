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
	"context"
	"fmt"
	"math/rand"
	"testing"

	"k8s.io/klog/v2"
)

func TestDynamicJoin12VWith13(t *testing.T) {
	ctx := context.Background()
	logger := klog.FromContext(ctx)
	exerciseDynamicJoin12VWith13W[int, string, float32, complex64, Empty](
		t,
		logger,
		NewMapMapFactory[Pair[int, string], complex64](nil),
		NewMapMapFactory[Pair[int, float32], Empty](nil),
		func() Pair[Pair[int, string], complex64] {
			return Pair[Pair[int, string], complex64]{
				Pair[int, string]{rand.Intn(3), fmt.Sprintf("%d", rand.Intn(3))},
				complex(0.5, float32(rand.Intn(3)+5)/4.0)}
		},
		func() Pair[Pair[int, float32], Empty] {
			return Pair[Pair[int, float32], Empty]{
				Pair[int, float32]{rand.Intn(3), float32(rand.Intn(3)+5) / 4.0},
				Empty{}}
		},
		NewMapMapFactory[Triple[int, string, float32], Pair[complex64, Empty]](nil),
		func(wideReceiver MappingReceiver[Triple[int, string, float32], Pair[complex64, Empty]]) (
			MappingReceiver[Pair[int, string], complex64],
			MappingReceiver[Pair[int, float32], Empty],
		) {
			narrowReceiver := TransformMappingReceiver[Triple[int, string, float32], Triple[int, string, float32], complex64, Pair[complex64, Empty]]{
				Identity1[Triple[int, string, float32]],
				NewPair2Then1[complex64, Empty](Empty{}),
				wideReceiver,
			}
			leftReceiver, rightReceiver := NewDynamicFullJoin12VWith13[int, string, float32, complex64](logger, narrowReceiver)
			return leftReceiver, MapKeySetReceiverLossy[Pair[int, float32], Empty](rightReceiver)
		},
	)
}
