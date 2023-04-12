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

// NewFactoredMapMapAggregator makes a FactoredMap that also notifies a receiver
// of updates to a GROUP BY KeyPartA & aggregate.
func NewFactoredMapMapAggregator[WholeKey, KeyPartA, KeyPartB comparable, Val any, Aggregation any](
	keyDecomposer Factorer[WholeKey, KeyPartA, KeyPartB],
	unifiedObserver MapChangeReceiver[WholeKey, Val],
	outerObserver MapChangeReceiver[KeyPartA, MutableMap[KeyPartB, Val]],
	aggregate func(KeyPartA, Map[KeyPartB, Val]) Aggregation,
	aggregationObserver MappingReceiver[KeyPartA, Aggregation],
) FactoredMap[WholeKey, KeyPartA, KeyPartB, Val] {
	observeToSolve := MappingReceiverFuncs[KeyPartA, Map[KeyPartB, Val]]{
		OnPut: func(keyPartA KeyPartA, problem Map[KeyPartB, Val]) {
			aggregation := aggregate(keyPartA, problem)
			aggregationObserver.Put(keyPartA, aggregation)
		},
		OnDelete: func(keyPartA KeyPartA) {
			aggregationObserver.Delete(keyPartA)
		},
	}
	return NewFactoredMapMap[WholeKey, KeyPartA, KeyPartB, Val](keyDecomposer, unifiedObserver, outerObserver, observeToSolve)
}
