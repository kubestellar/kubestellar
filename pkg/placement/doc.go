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

// This package contains the functionality of the placement translator.
// See declarations.go for how it is factored into big pieces.

//
// This package also contains some generic operations directly on
// Go data structures.

// Generic ops on slices:
// - SliceCopy.
// - SliceContains.
// - SliceContainsParametric.
// - SliceRemoveFunctional.
// - SliceEqual.
// - SliceApply applies a 1-arg 0-result fn to elts of a slice.

//
// This package also contains some generic types and functions for
// general-purpose logic.

// Functions:
// - Func11Compose11: composes two 1-arg 1-result functions.
// - Identity1: 1-arg identity function.
// - NewThunk: creates a 0-arg 1-result function.
// - Rotator: a bijection between two types.
// - Factorer: a bijection between a Pair and something isomorphic to it.
// - Runnable: something that Runs with a given Context.

// - interface Receiver: can be given a value.
// - interface Client: is configured once with a server.
// - interface DynamicValueProvider supports a dynamic set of Receivers
//   and has a callback-based Get method.
// - interface DynamicMapProvider is like DynamicValueProvider but specialized
//   to providing a map.

// Partial relations:
// - Comparison is the four possible results of comparing
//   two values in a partial order.

//
// This package also contains generic types and functions for relational algebra.
// This consists of the following.

// The collection interfaces here are not specific about concurrency:
// some implementations are safe for concurrent access, and some are not.

// Tuple types: Pair, Triple, and Quad; in tuples.go.
// All of the generics here have a fixed number of type parameters
// because Go does not allow a variable number of type parameters.

// Iterator/enumerator/generator:
// - interface Visitable declares the ability to enumerate
//   values to a consumer that can abort the enumeration.
// - VisitableFunc is a func that implements Visitable.
// - TransformVisitable composes a Visitable with a transform.
// - VisitableMap does the same as TransformVisitable.
// - VisitableLen computes length by enumerating.
// - VisitableHas enumerates and tests.
// - VisitableGet(index) enumerates as needed.
// - VisitableToSlice converts to a Go slice.
// - VisitableTransformToSlice composes a transform with VisitableToSlice.
// - VisitableStringer wraps a Visitable to add a String() method
//   that enumerates and formats members.
// - type Slice adds Visitable behavior to a slice,
//   and adds a functional Append method.

// Set types:
// - Interface Set declares methods to read aspects of a set.
// - Interface SetWriter declares methods to change a set.
// - Interface MutableSet is Set and SetWriter.
// - WrapSetWithMutex wraps a RWMutex around a given MutableSet.
// - NewSetReadonly wraps a MutableSet in a layer that delegates
//   the methods of Set and does NOT implement SetWriter.
// - SetWriterReverse transforms a SetWriter to its inverse.
// - SetWriterFork combines isomorphic SetWriters into one.
// - TransformSetWriter composes an element transform and a SetWriter.
// - SetRotate wraps an element bijection around a Set.
// - NewLoggingSetWriter creates a SetWriter that just logs.
// - SetEnumerateDifferences enumerates to a SetWriter.
// - Comparison functions: SetLessOrEqual, SetCompare, SetEqual -
//   "less" means "subset".
// - SetIntersection produces a Set that answers its methods based on delegated calls.
// - SetAddAll(MutableSet, Visitable) (someNew, allNew bool).
// - SetRemoveAll(MutableSet, Visitable) (someNew, allNew bool).
// - MutableSetUnion1Elt(MutableSet, Elt).
// - MutableSetUnion1Set(MutableSet, Set).

// - MapSet implements MutableSet in the same way as the generic Set of Kubernetes
//   (which is not in Kubernetes release 1.24).
// - MapSetAddNoResult(MapSet, Elt).
// - MapSetSymmetricDifference computes intersection and differences into fresh MapSets.

// A "relation" is a set of tuples.
// - interface Relation2 is a Set of Pair.
// - interface MutableRelation2 extends Relation2 with mutability.
// - interface Index2 is map from column1-type to set of column2-type.
// - interface MutableIndex2 adds mutability.
// - GenericIndexedSet is a Set of tuples + access to an Index2.
// - GenericMutableIndexedSet adds mutability and the ability to cast it away.
// - Relation2WithObservers wraps a MutableRelation2 with observers of changes.
// - SingleIndexedRelation2 is a MutableRelation2 implemented by one index.
// - SingleIndexedRelation3 is a mutable 3-ary relation implemented by two layers of indexing.
// - SingleIndexedRelation4 is a mutable 4-ary relation implemented by three layers of indexing.
// - NewMapRelation2 creates a SingleIndexedRelation2 using Go map based Maps and Sets.
// - NewHashRelation2 creates a SingleIndexedRelation2 using user-level hashing.
// - NewMapRelation3 creates a SingleIndexedRelation3 using Go map based Maps and Sets.
// - NewMapRelation4 creates a SingleIndexedRelation4 using GO map based Maps and Sets.

// - NewDynamicJoin12with13 does a passive join of (X,Y) with (X,Z) to a receiver of (Y,Z).
// - NewDynamicFullJoin12VWith13 does a passive join of ((X,Y)->V) with (X,Z) to a receiver of ((X,Y,Z)->V).
// - Relation2Equijoin12with13 joins (X,Y) with (X,Z) to compute (Y,Z) once.
// - Map12VEquijoinRelation13 joins ((X,Y)->V) with (X,Z) to compute ((X,Y,Z)->V) once.
// - MapEquijoin12With13 joins (X->Y) with (X->Z) to compute (X->(Y,Z)) once.
// - JoinByVisitSquared takes two Visitables and a matching function to produce a Visitable over a general join.

// Map types:
// A "map" is a computable and enumerable function from some domain to some range.
// - interface Map is the readable aspect of a map.
// - interface MappingReceiver has methods for writing to a map.
// - interface MutableMap extends Map with MappingReceiver.
// - MapTransformToSlice composes a transform with to-slice.
// - NewMappingReceiverFuncs is a pair of functions that together implement MappingReceiver.
// - MappingReceiverFork combines isomorphic MappingReceivers into one.
// - MappingReceiverFunc delegates to a MappingReceiver chosen independently on every call.
// - MapChangeReceiver is an observer of changes to a map.
// - MapChangeReceiverFuncs is three funcs that together implement MapChangeReceiver.
// - MapChangeReceiverFork combines isomorphic MapChangeReceivers into one.
// - MappingReceiverDiscardsPrevious wraps a MappingReceiver to produce
//   a MapChangeReceiver that discards information.
// - TransactionalMappingReceiver adds batching.
// - MapReadonly wraps a Map to hide mutability.
// - MutableMapWithKeyObserver wraps a MutableMap with an observer of changes to the set of keys.
// - TransformMappingReceiver composes two transforms with a MappingReceiver.
// - WrapMapWithMutex wraps a MutableMap with an RWMutex.
// - NewLoggingMappingReceiver constructs a MappingReceiver that just logs.
// - MapApply adds all entries in a given Map to a given MappingReceiver.
// - MapAddAll adds all entries in a Visitable to a MutableMap.
// - fn MapGetAdd is Get + generate&put-if-missing.
// - MapEqual requires comparable range type.
// - MapEqualParametric takes range comparison function.
// - MapEnumerateDifferences requires comparable range type and enumerates to a MapChangeReceiver.
// - MapEnumerateDifferencesParametric takes range comparison function and enumerates to a MapChangeReceiver.
// - MapKeySet exposes the Set of key values.
// - MapKeySetReceiver wraps a SetWriter for keys to make a MapChangeReceiver that only passes along the key set changes.
// - MapKeySetReceiverLossy wraps a SetWriter for keys to make a MappingReceiver that only passes along key information.
// - NewSetByMapToEmpty takesa MutableMap[Domain,Empty] and casts it to a MutableSet[Domain].
// - NewMapToConstant takes a domain set and a single range value and makes a Map that associates
//   each of those domain values to that one range value.
// - RotateKeyMap combines a Map with a bijection on its keys.
// - RotatedKeyMutableMap combines a MutableMap with a bijeciton on its keys.
// - MappingReceiver composes a bijection on keys with a MappingReceiver.

// - interface MapMap implements MutableMap by a Go `map[Domain]Range` and holds an optional observer.
// - MapMapCopy produces a fresh MapMap based on a given observer and Visitable of initial mappings.

// - FactoredMap is a map (a) whose Domain can be factored into two parts
//   and (b) has a two-layer index based on those two parts.
// - NewFactoredMapMapAggregator makes a FactoredMap that also notifies a receiver
//   of updates to a GROUP BY KeyPartA & aggregate.

// - NewHashMap creates a crummy hashtable given the hash and equality functions.  There has to be a better way!
// - NewHashSet uses NewHashMap to make a MutableSet.
// - HashSetCopy calls NewHashSet and populates it from a Visitable.

// Map/reduce:
// - a Reducer is a fn that takes a Visitable and reduces its elements.
// - ValueReducer produces a Reducer from a functional accumulator.
// - StatefulReducer produces a Reducer from a state-based accumulator.
// - VisitableMapReduce takes a Visitable, a map fn, and a functional reduce fn.
// - VisitableMapFnReduceOr is special case for reducer is OR.
