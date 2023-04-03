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
	// We use a variable so that an observer passed into the factored map constructor can
	// use the map created by that constructor.
	// We need to do that because FactoredMap does not have a hook that gets called on
	// every change and given the (KerPartA, Map[KeyPartB,Val]) pair.
	var fm FactoredMap[WholeKey, KeyPartA, KeyPartB, Val]
	observeToSolve := MappingReceiverFuncs[WholeKey, Val]{
		OnPut: func(wholeKey WholeKey, val Val) {
			decomposedKey := keyDecomposer.First(wholeKey)
			problem, _ := fm.GetIndex().Get(decomposedKey.First)
			aggregation := aggregate(decomposedKey.First, problem)
			aggregationObserver.Put(decomposedKey.First, aggregation)
		},
		OnDelete: func(wholeKey WholeKey) {
			decomposedKey := keyDecomposer.First(wholeKey)
			aggregationObserver.Delete(decomposedKey.First)
		},
	}
	if unifiedObserver == nil {
		unifiedObserver = observeToSolve
	} else {
		unifiedObserver = MapChangeReceiverFork[WholeKey, Val]{unifiedObserver, observeToSolve}
	}
	fm = NewFactoredMapMap[WholeKey, KeyPartA, KeyPartB, Val](keyDecomposer, unifiedObserver, outerObserver)
	return fm
}
